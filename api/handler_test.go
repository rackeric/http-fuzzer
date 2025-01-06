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

func (m *MockFuzzerManager) StartJob(target, wordlistID, jobType string) error {
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

func TestHandleStartJob(t *testing.T) {
	mockFuzzer := new(MockFuzzerManager)
	handler := NewHandler(mockFuzzer, nil, nil)

	body := map[string]string{
		"target":     "http://example.com",
		"wordlistId": "test-wordlist",
		"type":       "fuzzing",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/jobs/start", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mockFuzzer.On("StartJob", "http://example.com", "test-wordlist", "fuzzing").Return(nil)

	handler.handleStartJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockFuzzer.AssertExpectations(t)
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
			Type:       "fuzzing",
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
