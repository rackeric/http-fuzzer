package api

import (
	"bufio"
	"encoding/json"
	"net/http"

	"fuzzer/internal/logging"
	"fuzzer/internal/storage"
	"fuzzer/internal/wordlist"
	"fuzzer/types"
)

type Handler struct {
	fuzzerMgr   types.FuzzerManager
	wordlistMgr *wordlist.Manager
	store       *storage.JobStore
}

func NewHandler(f types.FuzzerManager, w *wordlist.Manager, s *storage.JobStore) *Handler {
	logging.Info("Creating new API handler")
	return &Handler{
		fuzzerMgr:   f,
		wordlistMgr: w,
		store:       s,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logging.Info("Received request: %s %s", r.Method, r.URL.Path)
	switch r.URL.Path {
	case "/api/jobs":
		h.handleJobs(w, r)
	case "/api/jobs/start":
		h.handleStartJob(w, r)
	case "/api/jobs/stop":
		h.handleStopJob(w, r)
	case "/api/wordlists":
		h.handleWordlists(w, r)
	case "/api/wordlists/add":
		h.handleAddWordlist(w, r)
	case "/api/rate-limit":
		h.handleUpdateRateLimit(w, r)
	default:
		logging.Error("Not found: %s", r.URL.Path)
		http.NotFound(w, r)
	}
}

func (h *Handler) handleJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.fuzzerMgr.GetJobs()
	if err != nil {
		logging.Error("Failed to retrieve jobs: %v", err)
		http.Error(w, "Failed to retrieve jobs", http.StatusInternalServerError)
		return
	}

	logging.Info("Retrieved %d jobs", len(jobs))
	for _, job := range jobs {
		logging.Debug("Job details: ID=%s Target=%s Status=%s", job.ID, job.Target, job.Status)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		logging.Error("Error encoding jobs: %v", err)
		http.Error(w, "Failed to encode jobs", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleStartJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target     string        `json:"target"`
		WordlistID string        `json:"wordlistId"`
		Type       types.JobType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Error("Invalid start job request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "fuzzing"
	}

	logging.Info("Starting job: Target=%s WordlistID=%s Type=%s", req.Target, req.WordlistID, req.Type)

	if err := h.fuzzerMgr.StartJob(req.Target, req.WordlistID, req.Type); err != nil {
		logging.Error("Failed to start job: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobs, _ := h.fuzzerMgr.GetJobs()
	var createdJob *types.Job
	for _, job := range jobs {
		if job.Target == req.Target && job.WordlistID == req.WordlistID {
			createdJob = job
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdJob)
}

func (h *Handler) handleStopJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JobID string `json:"jobId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Error("Invalid stop job request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logging.Info("Stopping job: JobID=%s", req.JobID)

	if err := h.fuzzerMgr.StopJob(req.JobID); err != nil {
		logging.Error("Failed to stop job: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleAddWordlist(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("wordlist")
	if err != nil {
		logging.Error("Failed to get wordlist file: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	logging.Info("Adding wordlist: %s", header.Filename)

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logging.Error("Error reading wordlist file: %v", err)
		http.Error(w, "Error reading wordlist", http.StatusInternalServerError)
		return
	}

	id := h.wordlistMgr.Add(r.FormValue("name"), words)
	logging.Info("Added wordlist with ID: %s", id)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) handleUpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RateLimit float64 `json:"rateLimit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Error("Invalid rate limit request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logging.Info("Updating rate limit to: %f", req.RateLimit)
	// Uncomment when implementation is ready
	// h.fuzzerMgr.UpdateRateLimit(req.RateLimit)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleWordlists(w http.ResponseWriter, r *http.Request) {
	wordlists := h.wordlistMgr.List()
	logging.Info("Retrieved %d wordlists", len(wordlists))
	json.NewEncoder(w).Encode(wordlists)
}
