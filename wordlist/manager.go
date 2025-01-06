package wordlist

import (
	"os"

	"fuzzer/types"
	"fuzzer/utils"
)

type Manager struct {
	baseDir string
	lists   map[string]*Wordlist
}

type Wordlist struct {
	ID    string
	Name  string
	Words []string
}

type WordlistManager interface {
	Get(id string) *types.Wordlist
	Add(name string, words []string) string
	List() []*types.Wordlist
}

// Get retrieves a wordlist by its ID from the Manager.
func (m *Manager) Get(id string) *types.Wordlist {
	wordlist, exists := m.lists[id]
	if !exists {
		return nil
	}

	return &types.Wordlist{
		ID:    wordlist.ID,
		Name:  wordlist.Name,
		Words: wordlist.Words,
	}
}

// ADD ... TODO ...
func (m *Manager) Add(name string, words []string) string {
	id := utils.GenerateID()
	m.lists[id] = &Wordlist{
		ID:    id,
		Name:  name,
		Words: words,
	}
	return id
}

// List returns all wordlists stored in the Manager as a slice of *types.Wordlist.
func (m *Manager) List() []*types.Wordlist {
	wordlists := make([]*types.Wordlist, 0, len(m.lists))
	for _, wl := range m.lists {
		wordlists = append(wordlists, &types.Wordlist{
			ID:    wl.ID,
			Name:  wl.Name,
			Words: wl.Words,
		})
	}
	return wordlists
}

// NewManager initializes and returns a Manager instance.
// It creates the base directory if it does not exist.
func NewManager(baseDir string) (*Manager, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	// Initialize and return the Manager
	return &Manager{
		baseDir: baseDir,
		lists:   make(map[string]*Wordlist),
	}, nil
}
