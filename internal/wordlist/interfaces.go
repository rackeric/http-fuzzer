package wordlist

import "fuzzer/types"

type WordlistStorer interface {
	Get(id string) *types.Wordlist
	GetByName(name string) *types.Wordlist
	Add(name string, words []string) string
	List() []*types.Wordlist
}
