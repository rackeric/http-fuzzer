package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"fuzzer/types"
)

type JobStore struct {
	filename string
	jobs     map[string]*types.Job
	mu       sync.RWMutex
}

func (s *JobStore) SaveJob(job *types.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job

	// Convert to a serializable format
	type SerializableJob struct {
		ID         string        `json:"id"`
		Target     string        `json:"target"`
		Status     string        `json:"status"`
		WordlistID string        `json:"wordlistId"`
		Type       types.JobType `json:"type"`
		// Add other fields that need to be serialized
	}

	serializableJobs := make(map[string]*SerializableJob)
	for id, j := range s.jobs {
		serializableJobs[id] = &SerializableJob{
			ID:         j.ID,
			Target:     j.Target,
			Status:     j.Status,
			WordlistID: j.WordlistID,
			Type:       j.Type,
		}
	}

	file, err := os.Create(s.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(serializableJobs)
}

func (s *JobStore) Save() error {
	data, err := json.Marshal(s.jobs)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, 0644)
}

// GetJob retrieves a job by its ID from the JobStore.
func (s *JobStore) GetJob(id string) (*types.Job, error) {
	job, exists := s.jobs[id]
	if !exists {
		return nil, errors.New("job not found")
	}
	return job, nil
}

// ListJobs returns all jobs stored in the JobStore as a slice.
func (s *JobStore) ListJobs() ([]*types.Job, error) {
	jobs := make([]*types.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// NewJobStore initializes and returns a JobStore instance.
// If the file at filepath exists, it attempts to load existing jobs from it.
func NewJobStore(filepath string) (*JobStore, error) {
	store := &JobStore{
		filename: filepath,
		jobs:     make(map[string]*types.Job),
	}

	// Check if the file exists
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		// File doesn't exist, return an empty JobStore
		return store, nil
	}

	// Read the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the data into the jobs map
	if len(data) > 0 { // Only attempt to unmarshal if the file is not empty
		if err := json.Unmarshal(data, &store.jobs); err != nil {
			return nil, err
		}
	}

	return store, nil
}
