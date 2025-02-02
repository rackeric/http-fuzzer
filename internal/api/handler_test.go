package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fuzzer/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFuzzerManager struct {
	mock.Mock
}

func (m *MockFuzzerManager) StartJob(target, wordlistID string, jobType types.JobType) error {
	args := m.Called(target, wordlistID, jobType)
	return args.Error(0)
}

func (m *MockFuzzerManager) StopJob(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}

func (m *MockFuzzerManager) GetJobs() ([]*types.Job, error) {
	args := m.Called()
	return args.Get(0).([]*types.Job), nil
}

func (m *MockFuzzerManager) DeleteJob(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}

func TestHandleStartJob(t *testing.T) {
	mockFuzzer := new(MockFuzzerManager)
	handler := NewHandler(mockFuzzer, nil, nil)

	// Create test job data
	testJob := &types.Job{
		ID:         "test-job",
		Target:     "http://example.com",
		WordlistID: "test-wordlist",
		Type:       types.DirectoryType,
		Status:     "running",
	}

	// Mock both StartJob and GetJobs calls
	mockFuzzer.On("StartJob", "http://example.com", "test-wordlist", types.DirectoryType).Return(nil)
	mockFuzzer.On("GetJobs").Return([]*types.Job{testJob}, nil)

	body := map[string]string{
		"target":     "http://example.com",
		"wordlistId": "test-wordlist",
		"type":       string(types.DirectoryType),
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/jobs/start", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.handleStartJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockFuzzer.AssertExpectations(t)

	// Verify response
	var response *types.Job
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, testJob.ID, response.ID)
	assert.Equal(t, testJob.Target, response.Target)
}

func TestHandleGetJobs(t *testing.T) {
	mockFuzzer := new(MockFuzzerManager)
	handler := NewHandler(mockFuzzer, nil, nil)

	jobs := []*types.Job{
		{
			ID:         "test-job",
			Target:     "http://example.com",
			Status:     "running",
			WordlistID: "test-wordlist",
			Type:       types.DirectoryType,
		},
	}

	mockFuzzer.On("GetJobs").Return(jobs)

	req := httptest.NewRequest("GET", "/api/jobs", nil)
	w := httptest.NewRecorder()

	handler.handleJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*types.Job
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "test-job", response[0].ID)
}
