package api

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"

	"fuzzer/internal/storage"
	"fuzzer/internal/wordlist"
	"fuzzer/types"
)

type Handler struct {
	fuzzerMgr   types.FuzzerManager // Changed from *fuzzer.Manager
	wordlistMgr *wordlist.Manager
	store       *storage.JobStore
}

func NewHandler(f types.FuzzerManager, w *wordlist.Manager, s *storage.JobStore) *Handler {
	return &Handler{
		fuzzerMgr:   f,
		wordlistMgr: w,
		store:       s,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		http.NotFound(w, r)
	}
}

func (h *Handler) handleJobs(w http.ResponseWriter, r *http.Request) {
	jobs, _ := h.fuzzerMgr.GetJobs()

	// Debug logging
	log.Printf("Retrieved %d jobs\n", len(jobs))
	for _, job := range jobs {
		log.Printf("Job: ID=%s Target=%s Status=%s\n", job.ID, job.Target, job.Status)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		log.Printf("Error encoding jobs: %v\n", err)
		http.Error(w, "Failed to encode jobs", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleStartJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target     string `json:"target"`
		WordlistID string `json:"wordlistId"`
		Type       string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "fuzzing"
	}

	if err := h.fuzzerMgr.StartJob(req.Target, req.WordlistID, req.Type); err != nil {
		log.Printf("Error starting job: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the created job and return it in the response
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.fuzzerMgr.StopJob(req.JobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleAddWordlist(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("wordlist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	id := h.wordlistMgr.Add(r.FormValue("name"), words)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) handleUpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RateLimit float64 `json:"rateLimit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// h.fuzzerMgr.UpdateRateLimit(req.RateLimit)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleWordlists(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(h.wordlistMgr.List())
}
