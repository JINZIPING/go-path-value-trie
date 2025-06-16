# NAME
prefix matching trie developed by golang

---

# DESCRIPTION
A high-performance, generic prefix-matching trie implemented in Go. This library is designed for extremely fast prefix string searching, making it ideal for search engine prefix recall, autocomplete, and similar scenarios. It supports flexible value types via Go generics, allowing you to store and rank your custom struct at each node.

## Features
- Generic design: supports any user-defined struct as value, as long as it implements `GetQ()` and `GetScore()` methods.
- Stores values at every node along the path for efficient prefix recall
- Supports combining multiple tries for advanced scenarios
- Simple API: `Put` and `Get`

## Installation

```sh
go get github.com/yourname/go-path-value-trie
```

## Generic Design
This library uses Go generics (Go 1.20+) for maximum flexibility and type safety. You can define your own value struct, as long as it implements the following interface:

```go
type Value interface {
    GetQ() string
    GetScore() float32
}
```

For example:

```go
type MyValue struct {
    Q      string
    Score  float32
    GameId int64
}

func (v MyValue) GetQ() string      { return v.Q }
func (v MyValue) GetScore() float32 { return v.Score }
```

## Usage

### 1. Single Trie Usage (Generic)

```go
import "github.com/yourname/go-path-value-trie/trie"

func main() {
    t := trie.NewPathValueTrie[MyValue](10) // max 10 values per node
    t.Put("pubg", &MyValue{Q: "pubg", Score: 1.0, GameId: 123})
    results := t.Get("pu")
    for _, v := range results {
        fmt.Println(v.Q, v.Score, v.GameId)
    }
}
```

### 2. Multiple Tries Usage (Generic)

You can combine multiple tries (e.g., for original, pinyin, and abbreviation forms) and perform a unified search:

```go
import "github.com/yourname/go-path-value-trie/trie"

func main() {
    qTrie := trie.NewPathValueTrie[MyValue](10)
    pyTrie := trie.NewPathValueTrie[MyValue](10)
    // ... initialize and fill each trie as needed
    tries := trie.NewPathValueTries[MyValue]([]*trie.PathValueTrie[MyValue]{qTrie, pyTrie}, 20) // max 20 results
    results := tries.Get("和平")
    for _, v := range results {
        fmt.Println(v.Q, v.Score, v.GameId)
    }
}
```

## Example & How to Run

A full demo is provided in the `example/` directory. To run the example:

```sh
cd example
# Generate test data (optional)
python data/gen_data.py
# Run the example (requires Go 1.18+)
go run *.go
```