package fuzzer

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

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

	// Save to both memory and persistent storage
	m.jobs[job.ID] = job
	if err := m.store.SaveJob(job); err != nil {
		m.mu.Unlock()
		return fmt.Errorf("failed to save job: %w", err)
	}
	m.mu.Unlock()

	// Verify wordlist exists before starting goroutine
	wordlist := m.wordlistMgr.Get(wordlistID)
	if wordlist == nil {
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

	// Return all jobs from our in-memory map
	jobs := make([]*types.Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// func (m *Manager) PauseJob(jobID string) error {
// 	m.mu.Lock()
// 	job := m.jobs[jobID]
// 	m.mu.Unlock()

// 	if job != nil && job.Status == "running" {
// 		job.Status = "paused"
// 		job.PauseChan <- true
// 	}
// 	return nil
// }

// func (m *Manager) ResumeJob(jobID string) error {
// 	m.mu.Lock()
// 	job := m.jobs[jobID]
// 	m.mu.Unlock()

// 	if job != nil && job.Status == "paused" {
// 		job.Status = "running"
// 		job.PauseChan <- false
// 	}
// 	return nil
// }

func (m *Manager) StopJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = "stopped"
	if err := m.store.SaveJob(job); err != nil {
		return fmt.Errorf("failed to save job status: %w", err)
	}

	return nil
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
	baseHost := strings.Split(target, "://")[1]
	subdomain := fmt.Sprintf("%s.%s", word, baseHost)

	_, err := net.LookupHost(subdomain)
	if err == nil {
		return fmt.Sprintf("http://%s", subdomain)
	}
	return ""
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

func (m *Manager) runJob(job *types.Job) {
	jobCtx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	wordlist := m.wordlistMgr.Get(job.WordlistID) //TODO: should return error
	// if err != nil {
	// 	m.updateJobStatus(job, "failed")
	// 	return
	// }
	totalWords := len(wordlist.Words)

	for i, word := range wordlist.Words {
		select {
		case <-jobCtx.Done():
			m.updateJobStatus(job, "stopped")
			return
		default:
			err := m.limiter.Wait(jobCtx)
			if err != nil {
				continue
			}

			job.Progress = int(float64(i+1) / float64(totalWords) * 100)

			switch job.Type {
			case types.DirectoryType:
				if url := m.checkDirectory(job.Target, word); url != "" {
					m.addFinding(job, url, string(types.DirectoryType))
				}

			case types.SubdomainType:
				if url := m.checkSubdomain(job.Target, word); url != "" {
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
				m.store.Save()
			}

			// _ = word
			// time.Sleep(100 * time.Millisecond)
		}
	}

	m.updateJobStatus(job, "completed")
}

func (m *Manager) updateJobStatus(job *types.Job, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job.Status = status
	m.store.SaveJob(job)
}

func (m *Manager) UpdateRateLimit(newLimit float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rateLimit = newLimit
	m.limiter = rate.NewLimiter(rate.Limit(newLimit), 1)
}
