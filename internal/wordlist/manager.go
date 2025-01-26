package wordlist

import (
	"os"

	"fuzzer/internal/logging"
	"fuzzer/types"
	"fuzzer/utils"
)

type Manager struct {
	baseDir string
	lists   map[string]*types.Wordlist
}

func (m *Manager) Get(id string) *types.Wordlist {
	wordlist, exists := m.lists[id]
	if !exists {
		logging.Debug("Wordlist not found with ID: %s", id)
		return nil
	}

	logging.Debug("Retrieved wordlist: ID=%s Name=%s", wordlist.ID, wordlist.Name)
	return &types.Wordlist{
		ID:    wordlist.ID,
		Name:  wordlist.Name,
		Words: wordlist.Words,
	}
}

func (m *Manager) GetByName(name string) *types.Wordlist {
	for _, v := range m.lists {
		if v.Name == name {
			logging.Debug("Retrieved wordlist by name: Name=%s ID=%s", name, v.ID)
			return &types.Wordlist{
				ID:    v.ID,
				Name:  v.Name,
				Words: v.Words,
			}
		}
	}
	logging.Debug("No wordlist found with name: %s", name)
	return nil
}

func (m *Manager) Add(name string, words []string) string {
	id := utils.GenerateID()
	m.lists[id] = &types.Wordlist{
		ID:    id,
		Name:  name,
		Words: words,
	}
	logging.Info("Added new wordlist: ID=%s Name=%s Words=%d", id, name, len(words))
	return id
}

func (m *Manager) List() []*types.Wordlist {
	wordlists := make([]*types.Wordlist, 0, len(m.lists))
	for _, wl := range m.lists {
		wordlists = append(wordlists, &types.Wordlist{
			ID:    wl.ID,
			Name:  wl.Name,
			Words: wl.Words,
		})
	}

	logging.Info("Listed wordlists: Total=%d", len(wordlists))
	return wordlists
}

func NewManager(baseDir string) (*Manager, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		logging.Error("Failed to create wordlist base directory: %v", err)
		return nil, err
	}

	logging.Info("Initializing wordlist manager with base directory: %s", baseDir)

	// Initialize and return the Manager
	return &Manager{
		baseDir: baseDir,
		lists:   make(map[string]*types.Wordlist),
	}, nil
}
