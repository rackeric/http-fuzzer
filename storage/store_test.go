package storage

import (
	"os"
	"testing"

	"fuzzer/types"

	"github.com/stretchr/testify/assert"
)

func TestJobStore(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "jobstore_test*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	store, err := NewJobStore(tmpFile.Name())
	assert.NoError(t, err)

	job := &types.Job{
		ID:         "test-job",
		Target:     "http://example.com",
		Status:     "running",
		WordlistID: "test-wordlist",
		Type:       "fuzzing",
	}

	err = store.SaveJob(job)
	assert.NoError(t, err)

	// Test retrieving the job
	retrieved, err := store.GetJob("test-job")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, job.ID, retrieved.ID)
	assert.Equal(t, job.Target, retrieved.Target)
}
