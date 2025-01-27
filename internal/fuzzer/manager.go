package fuzzer

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"fuzzer/internal/logging"
	"fuzzer/internal/storage"
	"fuzzer/internal/wordlist"
	"fuzzer/types"

	"golang.org/x/time/rate"
)

type Finding struct {
	URL   string
	Type  string // "directory" or "subdomain"
	Found time.Time
}

type Manager struct {
	ctx         context.Context
	cancel      context.CancelFunc
	store       storage.JobStorer
	wordlistMgr wordlist.WordlistStorer
	rateLimit   float64
	limiter     *rate.Limiter
	jobs        map[string]*types.Job
	mu          sync.RWMutex
}

func NewManager(ctx context.Context, store storage.JobStorer, wordlistMgr wordlist.WordlistStorer, rateLimit float64) *Manager {
	ctx, cancel := context.WithCancel(ctx)
	logging.Info("Created new fuzzer manager with rate limit: %f", rateLimit)
	return &Manager{
		ctx:         ctx,
		cancel:      cancel,
		store:       store,
		wordlistMgr: wordlistMgr,
		rateLimit:   rateLimit,
		limiter:     rate.NewLimiter(rate.Limit(rateLimit), 1),
		jobs:        make(map[string]*types.Job),
	}
}

func (m *Manager) StartJob(target, wordlistID string, jobType types.JobType) error {
	m.mu.Lock()
	job := &types.Job{
		ID:         fmt.Sprintf("job-%d", len(m.jobs)+1),
		Target:     target,
		Status:     "running",
		WordlistID: wordlistID,
		Type:       jobType,
		StartTime:  time.Now(),
		Findings:   make([]types.Finding, 0),
	}

	logging.Info("Starting new job: ID=%s Target=%s Type=%s", job.ID, target, jobType)

	// Save to both memory and persistent storage
	m.jobs[job.ID] = job
	if err := m.store.SaveJob(job); err != nil {
		m.mu.Unlock()
		logging.Error("Failed to save job: %v", err)
		return fmt.Errorf("failed to save job: %w", err)
	}
	m.mu.Unlock()

	// Verify wordlist exists before starting goroutine
	wordlist := m.wordlistMgr.Get(wordlistID)
	if wordlist == nil {
		logging.Error("Wordlist not found: %s", wordlistID)
		return fmt.Errorf("wordlist not found: %s", wordlistID)
	}

	// Start actual fuzzing in a goroutine
	go m.runJob(job)

	return nil
}

func (m *Manager) GetJobs() ([]*types.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// First, get jobs from persistent storage
	storedJobs, _ := m.store.ListJobs()

	// Update our in-memory map with any jobs from storage we don't have
	for _, job := range storedJobs {
		if _, exists := m.jobs[job.ID]; !exists {
			m.jobs[job.ID] = job
		}
	}

	logging.Info("Retrieved %d jobs", len(m.jobs))

	// Return all jobs from our in-memory map
	jobs := make([]*types.Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (m *Manager) StopJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		logging.Error("Attempted to stop non-existent job: %s", jobID)
		return fmt.Errorf("job not found: %s", jobID)
	}

	logging.Info("Stopping job: %s", jobID)
	job.Status = "stopped"
	if err := m.store.SaveJob(job); err != nil {
		logging.Error("Failed to save stopped job status: %v", err)
		return fmt.Errorf("failed to save job status: %w", err)
	}

	return nil
}

func (m *Manager) DeleteJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.jobs[jobID]
	if !exists {
		logging.Error("Attempted to delete non-existent job: %s", jobID)
		return fmt.Errorf("job not found: %s", jobID)
	}

	logging.Info("Deleting job: %s", jobID)
	delete(m.jobs, jobID)
	if err := m.store.DeleteJob(jobID); err != nil {
		logging.Error("Failed to delete job: %v", err)
		return fmt.Errorf("failed to delete job: %w", err)
	}
	m.store.Save()
	return nil
}

func (m *Manager) runJob(job *types.Job) {
	jobCtx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	logging.Info("Running job: ID=%s Target=%s Type=%s", job.ID, job.Target, job.Type)

	wordlist := m.wordlistMgr.Get(job.WordlistID)
	if wordlist == nil {
		logging.Error("Failed to get wordlist for job: %s", job.ID)
		m.updateJobStatus(job, "failed")
		return
	}
	totalWords := len(wordlist.Words)

	for i, word := range wordlist.Words {
		select {
		case <-jobCtx.Done():
			logging.Info("Job stopped: %s", job.ID)
			m.updateJobStatus(job, "stopped")
			return
		default:
			err := m.limiter.Wait(jobCtx)
			if err != nil {
				logging.Debug("Rate limiter error: %v", err)
				continue
			}

			job.Progress = int(float64(i+1) / float64(totalWords) * 100)

			switch job.Type {
			case types.DirectoryType:
				if url := m.checkDirectory(job.Target, word); url != "" {
					logging.Info("Directory found: %s", url)
					m.addFinding(job, url, string(types.DirectoryType))
				}

			case types.SubdomainType:
				if url := m.checkSubdomain(job.Target, word); url != "" {
					logging.Info("Subdomain found: %s", url)
					m.addFinding(job, url, string(types.SubdomainType))
					// Pass context to recursive call
					go m.runJob(&types.Job{
						Target:     url,
						WordlistID: job.WordlistID,
						Type:       types.SubdomainType,
					})
				}
			}

			if i%100 == 0 {
				logging.Debug("Saving job progress: %d%%", job.Progress)
				m.store.Save()
			}
		}
	}

	logging.Info("Job completed: %s", job.ID)
	m.updateJobStatus(job, "completed")
}

func (m *Manager) updateJobStatus(job *types.Job, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logging.Info("Updating job status: ID=%s Status=%s", job.ID, status)
	job.Status = status
	m.store.SaveJob(job)
}

func (m *Manager) UpdateRateLimit(newLimit float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logging.Info("Updating rate limit from %f to %f", m.rateLimit, newLimit)
	m.rateLimit = newLimit
	m.limiter = rate.NewLimiter(rate.Limit(newLimit), 1)
}

func (m *Manager) checkDirectory(target, word string) string {
	url := fmt.Sprintf("%s/%s", target, word)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden {
		return url
	}
	return ""
}

func (m *Manager) checkSubdomain(target, word string) string {
	// Parse the original target URL to get the base host and protocol
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return ""
	}

	// Form the subdomain
	subdomain := fmt.Sprintf("%s.%s", word, parsedTarget.Host)

	// Create HTTP client with custom settings
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		// Don't follow redirects as they might reveal information
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Try both HTTP and HTTPS on the original host
	protocols := []string{"http://", "https://"}
	for _, protocol := range protocols {
		// Create the request to the original host
		baseURL := fmt.Sprintf("%s%s", protocol, parsedTarget.Host)
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			continue
		}

		// Set the Host header to the subdomain we're testing
		req.Host = subdomain

		// Add common headers that might help with virtual host detection
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		// Read a small portion of the body to help with virtual host detection
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body.Close()

		// Check if this might be a valid virtual host
		// You might want to adjust these detection methods based on your needs
		if isLikelyValidVHost(resp, bodyBytes) {
			return fmt.Sprintf("%s%s", protocol, subdomain)
		}
	}

	return ""
}

// isLikelyValidVHost checks if the response indicates a valid virtual host
func isLikelyValidVHost(resp *http.Response, bodyBytes []byte) bool {
	// Consider it valid if we get a successful response
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true
	}

	// Check for specific status codes that might indicate a valid host
	if resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusForbidden {
		return true
	}

	// If we get a 404 or 500, check if it's a custom error page
	// This might indicate a valid host with an error
	if (resp.StatusCode == http.StatusNotFound ||
		resp.StatusCode == http.StatusInternalServerError) &&
		len(bodyBytes) > 512 {
		return true
	}

	return false
}

func (m *Manager) addFinding(job *types.Job, url, findingType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job.Findings = append(job.Findings, types.Finding{
		URL:   url,
		Type:  findingType,
		Found: time.Now(),
	})
}
