package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	trie "github.com/jinziping/go-path-value-trie"
	"github.com/mozillazg/go-pinyin"
)

type MyValue struct {
	Q      string  `json:"-"`
	Score  float32 `json:"score"`
	GameId int64   `json:"game_id"`
}

func (v MyValue) GetQ() string      { return v.Q }
func (v MyValue) GetScore() float32 { return v.Score }

// LoadCorrectionManualList loads a correction manual list from the specified file and builds generic Trie trees.
func LoadCorrectionManualList(file string) *trie.PathValueTries[MyValue] {
	const maxPerCandidate = 200 // Max number of indexes per candidate
	const maxInputLen = 500000  // Max number of lines in the file
	const maxResultsLen = 200   // Max number of results returned by trie
	const maxNodeLen = 200      // Max number of candidates per trie node
	const floatEpsilon = 1e-6
	type CandidateInfo struct {
		indexes []string
		attr    MyValue
	}
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("LoadCorrectionManualList failed, err: %v", err)
		return nil
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	candidateMap := make(map[string]CandidateInfo)
	inputLen := 0
	overIndexCandidate, ovrLmtIdx := 0, 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		inputLen++
		if inputLen > maxInputLen {
			break
		}
		lineArr := strings.Split(strings.ReplaceAll(line, "\n", ""), "\u0001")
		if len(lineArr) < 3 {
			continue
		}
		index_query := lineArr[0]
		candidate_query := lineArr[1]
		attr_str := lineArr[2]
		if len(index_query) == 0 || len(candidate_query) == 0 {
			continue
		}
		var attr MyValue
		jErr := json.Unmarshal([]byte(attr_str), &attr)
		if jErr != nil {
			continue
		}
		attr.Q = candidate_query
		if old, ok := candidateMap[candidate_query]; !ok {
			item := CandidateInfo{indexes: []string{index_query}, attr: attr}
			candidateMap[candidate_query] = item
		} else {
			if floatCompare(attr.Score, old.attr.Score, floatEpsilon) == 1 {
				old.attr = attr
			}
			if len(old.indexes) >= maxPerCandidate {
				if len(old.indexes) > maxPerCandidate {
					ovrLmtIdx++
					continue
				}
				overIndexCandidate++
			}
			old.indexes = append(old.indexes, index_query)
			candidateMap[candidate_query] = old
		}
	}
	if overIndexCandidate > 0 {
		fmt.Printf("[WARNING] %d candidates have more than the max allowed indexes (%d), ignored.\n", overIndexCandidate, maxPerCandidate)
	}
	candidateSlice := make([]CandidateInfo, 0, len(candidateMap))
	for _, v := range candidateMap {
		candidateSlice = append(candidateSlice, v)
	}
	candidateMap = nil // Help GC
	sort.Slice(candidateSlice, func(i, j int) bool {
		scoreA := candidateSlice[i].attr.Score
		scoreB := candidateSlice[j].attr.Score
		switch floatCompare(scoreA, scoreB, floatEpsilon) {
		case 1:
			return true
		case -1:
			return false
		default:
			return candidateSlice[i].attr.Q > candidateSlice[j].attr.Q
		}
	})
	qTrie := trie.NewPathValueTrie[MyValue](maxNodeLen)
	pyTrie := trie.NewPathValueTrie[MyValue](maxNodeLen)
	jpTrie := trie.NewPathValueTrie[MyValue](maxNodeLen)
	tries := trie.NewPathValueTries([]*trie.PathValueTrie[MyValue]{qTrie, pyTrie, jpTrie}, maxResultsLen)
	qCount, pyCount, jpCount := 0, 0, 0
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for idx := range candidateSlice {
			for _, key := range candidateSlice[idx].indexes {
				if len(key) == 0 {
					continue
				}
				if !qTrie.Put(key, &candidateSlice[idx].attr) {
					fmt.Printf("Put failed: mode=0, key=%v, score=%v\n", key, candidateSlice[idx].attr.Score)
				}
				qCount++
			}
		}
	}()
	go func() {
		defer wg.Done()
		for idx := range candidateSlice {
			for _, key0 := range candidateSlice[idx].indexes {
				key := Pinyin(key0)
				if len(key) == 0 {
					continue
				}
				if !pyTrie.Put(key, &candidateSlice[idx].attr) {
					fmt.Printf("Put failed: mode=1, key=%v, score=%v\n", key, candidateSlice[idx].attr.Score)
				}
				pyCount++
			}
		}
	}()
	go func() {
		defer wg.Done()
		for idx := range candidateSlice {
			for _, key0 := range candidateSlice[idx].indexes {
				key := PinyinJp(key0)
				if len(key) == 0 {
					continue
				}
				if !jpTrie.Put(key, &candidateSlice[idx].attr) {
					fmt.Printf("Put failed: mode=2, key=%v, score=%v\n", key, candidateSlice[idx].attr.Score)
				}
				jpCount++
			}
		}
	}()
	wg.Wait()
	fmt.Printf("Trie loaded from %s, Q: %d, Pinyin: %d, PinyinJp: %d\n", file, qCount, pyCount, jpCount)
	return tries
}

// floatEqual and floatCompare helpers
func floatEqual(a, b, epsilon float32) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

func floatCompare(a, b, epsilon float32) int {
	if floatEqual(a, b, epsilon) {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}

var MyPy pinyin.Args

func Pinyin(text string) string {
	arrPy := pinyin.Pinyin(text, MyPy)
	if len(arrPy) == 0 {
		return ""
	}
	build := strings.Builder{}
	for i := 0; i < len(arrPy); i++ {
		build.WriteString(arrPy[i][0])
	}
	return build.String()
}

var MyPyJp pinyin.Args

func PinyinJp(text string) string {
	arrPy := pinyin.LazyPinyin(text, MyPyJp)
	if len(arrPy) == 0 {
		return ""
	}
	return strings.Join(arrPy, "")
}
