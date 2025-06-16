// PathValueTries is a collection of PathValueTrie instances, allowing for efficient retrieval of values associated with a key across multiple tries.

package trie

import (
	"sort"
)

type PathValueTries[T Value] struct {
	pvTries       []*PathValueTrie[T]
	maxResultsLen int
}

func NewPathValueTries[T Value](tries []*PathValueTrie[T], maxResultsLen int) *PathValueTries[T] {
	if len(tries) == 0 {
		return nil
	}
	return &PathValueTries[T]{
		pvTries:       tries,
		maxResultsLen: maxResultsLen,
	}
}

func (tries *PathValueTries[T]) Get(key string) []T {
	if tries == nil || len(tries.pvTries) == 0 {
		return nil
	}
	resultsMap := make(map[string]T)
	for _, trie := range tries.pvTries {
		values := trie.Get(key) // values为[]T
		if values == nil {
			continue
		}
		for _, value := range values {
			q := value.GetQ()
			if existing, ok := resultsMap[q]; !ok || valueCompare(&value, &existing) {
				resultsMap[q] = value
			}
		}
	}
	var results []T
	for _, value := range resultsMap {
		results = append(results, value)
	}
	sort.Slice(results, func(i, j int) bool {
		return valueCompare(&results[i], &results[j])
	})
	if len(results) > tries.maxResultsLen {
		results = results[:tries.maxResultsLen]
	}
	return results
}
