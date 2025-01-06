package wordlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordlistManager(t *testing.T) {
	manager, err := NewManager("./test_wordlists")
	assert.NoError(t, err)

	// Test adding a wordlist
	words := []string{"test1", "test2", "test3"}
	id := manager.Add("test-list", words)

	assert.NotEmpty(t, id)

	// Test getting the wordlist
	list := manager.Get(id)
	assert.NotNil(t, list)
	assert.Equal(t, "test-list", list.Name)
	assert.Equal(t, words, list.Words)

	// Test listing wordlists
	lists := manager.List()
	assert.Len(t, lists, 1)
	assert.Equal(t, id, lists[0].ID)
}
