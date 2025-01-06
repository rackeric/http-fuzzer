package fuzzer

import (
	"context"
	"fuzzer/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type MockWordlistManager struct {
	mock.Mock
}

func (m *MockWordlistManager) Get(id string) *types.Wordlist {
	args := m.Called(id)
	return args.Get(0).(*types.Wordlist)
}

func (m *MockWordlistManager) Add(name string, words []string) string {
	args := m.Called(name, words)
	return args.String(0)
}

func (m *MockWordlistManager) List() []*types.Wordlist {
	args := m.Called()
	return args.Get(0).([]*types.Wordlist)
}

type MockJobStore struct {
	mock.Mock
}

func (m *MockJobStore) Save() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockJobStore) GetJob(id string) (*types.Job, error) {
	args := m.Called(id)
	return args.Get(0).(*types.Job), args.Error(1)
}

func (m *MockJobStore) ListJobs() ([]*types.Job, error) {
	args := m.Called()
	return args.Get(0).([]*types.Job), args.Error(1)
}

func TestStartJob(t *testing.T) {
	ctx := context.Background()
	mockStore := &MockJobStore{}
	mockWordlistMgr := &MockWordlistManager{}

	manager := &Manager{
		ctx:         ctx,
		store:       mockStore,
		wordlistMgr: mockWordlistMgr,
		jobs:        make(map[string]*types.Job),
	}

	mockWordlistMgr.On("Get", "test-wordlist").Return(&types.Wordlist{
		ID:    "test-wordlist",
		Words: []string{"test1", "test2"},
	})

	mockStore.On("Save").Return(nil)

	// Test starting a job
	err := manager.StartJob("http://example.com", "test-wordlist", "fuzzing")
	assert.NoError(t, err)

	// Verify job was created
	assert.Len(t, manager.jobs, 1)

	// Get the created job
	var job *types.Job
	for _, j := range manager.jobs {
		job = j
		break
	}

	assert.Equal(t, "running", job.Status)
	assert.Equal(t, "http://example.com", job.Target)
	assert.Equal(t, "test-wordlist", job.WordlistID)
}
