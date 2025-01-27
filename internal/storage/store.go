package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"fuzzer/internal/logging"
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

	logging.Debug("Saving job with ID: %s", job.ID)

	s.jobs[job.ID] = job

	// Convert to a serializable format
	type SerializableJob struct {
		ID         string        `json:"id"`
		Target     string        `json:"target"`
		Status     string        `json:"status"`
		WordlistID string        `json:"wordlistId"`
		Type       types.JobType `json:"type"`
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
		logging.Error("Failed to create file %s: %v", s.filename, err)
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(serializableJobs)
	if err != nil {
		logging.Error("Failed to encode jobs: %v", err)
		return err
	}

	logging.Info("Successfully saved %d jobs to %s", len(serializableJobs), s.filename)
	return nil
}

func (s *JobStore) Save() error {
	data, err := json.Marshal(s.jobs)
	if err != nil {
		logging.Error("Failed to marshal jobs: %v", err)
		return err
	}

	err = os.WriteFile(s.filename, data, 0644)
	if err != nil {
		logging.Error("Failed to write jobs to file %s: %v", s.filename, err)
		return err
	}

	logging.Info("Successfully saved %d jobs to %s", len(s.jobs), s.filename)
	return nil
}

func (s *JobStore) GetJob(id string) (*types.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		logging.Debug("Job not found with ID: %s", id)
		return nil, errors.New("job not found")
	}

	logging.Debug("Retrieved job with ID: %s", id)
	return job, nil
}

func (s *JobStore) DeleteJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[id]; !exists {
		logging.Debug("Attempted to delete non-existent job with ID: %s", id)
		return errors.New("job not found")
	}

	delete(s.jobs, id)
	s.Save()
	logging.Debug("Deleted job with ID: %s", id)
	return nil
}

func (s *JobStore) ListJobs() ([]*types.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*types.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	logging.Info("Listed %d jobs", len(jobs))
	return jobs, nil
}

func NewJobStore(filepath string) (*JobStore, error) {
	store := &JobStore{
		filename: filepath,
		jobs:     make(map[string]*types.Job),
	}

	// Check if the file exists
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		logging.Info("Creating new job store file: %s", filepath)
		return store, nil
	}

	// Read the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		logging.Error("Failed to read job store file %s: %v", filepath, err)
		return nil, err
	}

	// Unmarshal the data into the jobs map
	if len(data) > 0 {
		if err := json.Unmarshal(data, &store.jobs); err != nil {
			logging.Error("Failed to unmarshal job store data: %v", err)
			return nil, err
		}
		logging.Info("Loaded %d jobs from %s", len(store.jobs), filepath)
	}

	return store, nil
}
