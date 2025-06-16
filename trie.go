// PathValueTrie会在走过的所有路径都存储数据
// eg: 存储“王者荣耀”
//     trie: 只在“耀”节点存数据
//     PathValueTrie: 在“王” “者” “荣” “耀”四个节点存数据以实现前缀召回

// PathValueTrie will store data at all nodes along the path
// eg: when storing "pubg":
//     trie: stores data only at the "g" node
//     PathValueTrie: stores data at the "p", "pu", "pub", and "pubg" nodes to achieve prefix recall

package trie

import (
	"sort"
	"unicode/utf8"
)

// Value is the interface that user-defined value types must implement for use in PathValueTrie.
type Value interface {
	GetQ() string
	GetScore() float32
}

// PathValueTrieValue represents the value stored at each node.
type PathValueTrieValue struct {
	Q      string  // Query string
	Score  float32 // Score for ranking
	GameId int64   // Associated game ID
}

type pathValueTrieNode[T Value] struct {
	r        rune
	values   []*T
	children []*pathValueTrieNode[T]
}

type PathValueTrie[T Value] struct {
	root   *pathValueTrieNode[T]
	maxLen int
}

func NewPathValueTrie[T Value](maxLen int) *PathValueTrie[T] {
	return &PathValueTrie[T]{
		maxLen: maxLen,
		root: &pathValueTrieNode[T]{
			r:        0,
			values:   make([]*T, 0, min(4, maxLen)),
			children: make([]*pathValueTrieNode[T], 0, 8),
		},
	}
}

func (trie *PathValueTrie[T]) Put(key string, value *T) bool {
	if !utf8.ValidString(key) || value == nil || len((*value).GetQ()) == 0 {
		return false
	}

	node := trie.root
	for _, r := range key {
		children := node.children
		idx := sort.Search(len(children), func(i int) bool {
			return children[i].r >= r
		})
		if idx < len(children) && children[idx].r == r {
			node = children[idx]
		} else {
			newNode := &pathValueTrieNode[T]{
				r:        r,
				values:   make([]*T, 0, min(4, trie.maxLen)),
				children: make([]*pathValueTrieNode[T], 0, 8),
			}
			newChildren := make([]*pathValueTrieNode[T], len(children)+1)
			copy(newChildren, children[:idx])
			newChildren[idx] = newNode
			copy(newChildren[idx+1:], children[idx:])
			node.children = newChildren
			node = newNode
		}
		trie.updateValue(node, value)
	}
	return true
}

func (trie *PathValueTrie[T]) updateValue(node *pathValueTrieNode[T], value *T) {
	values := node.values
	// pre-filter: if the value already exists, return
	for i := range values {
		if (*values[i]).GetQ() == (*value).GetQ() {
			return
		}
	}

	// pre-filter: if values already full, return
	if len(values) >= trie.maxLen && valueCompare(values[len(values)-1], value) {
		return
	}

	// expand values slice
	if cap(values) == len(values) {
		newCap := cap(values) * 2
		if newCap > trie.maxLen+1 {
			newCap = trie.maxLen + 1
		}
		newValues := make([]*T, len(values), newCap)
		copy(newValues, values)
		values = newValues
	}

	// main logic: insert the value in sorted order
	if len(values) == 0 || valueCompare(values[len(values)-1], value) {
		values = append(values, value)
		node.values = values
		return
	}

	// bottom logic: new value is not smaller, insert and sort
	values = append(values, value)
	sort.Slice(values, func(i, j int) bool {
		return valueCompare(values[i], values[j])
	})
	if len(values) > trie.maxLen {
		values = values[:trie.maxLen]
	}
	node.values = values
}

func (trie *PathValueTrie[T]) Get(key string) []T {
	node := trie.root
	for _, r := range key {
		children := node.children
		idx := sort.Search(len(children), func(i int) bool {
			return children[i].r >= r
		})
		if idx >= len(children) || children[idx].r != r {
			return nil
		}
		node = children[idx]
	}
	result := make([]T, 0, len(node.values))
	for _, v := range node.values {
		if v != nil {
			result = append(result, *v)
		}
	}
	return result
}

func valueCompare[T Value](a, b *T) bool {
	switch floatCompare((*a).GetScore(), (*b).GetScore()) {
	case 1:
		return true
	case -1:
		return false
	case 0:
		return (*a).GetQ() > (*b).GetQ()
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func floatEqual(a, b float32) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 1e-6
}

func floatCompare(a, b float32) int {
	if floatEqual(a, b) {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}
