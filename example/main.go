// This is a demo for loading a correction manual list into a trie structure.
// We load a file where each line contains an index query with its item (which contains a candidate query and its attributes in this example).

package main

import (
	"fmt"
	"time"

	"github.com/mozillazg/go-pinyin"
)

func main() {
	now := time.Now()
	MyPy = pinyin.NewArgs()
	MyPy.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	MyPyJp = pinyin.NewArgs()
	MyPyJp.Style = pinyin.FIRST_LETTER

	fmt.Printf("[INFO] Start loading dictionary and building Trie...\n")
	tree := LoadCorrectionManualList("example/data/example.data")
	fmt.Printf("[INFO] Dictionary loaded and Trie built, elapsed: %v\n", time.Since(now))

	{
		now1 := time.Now()
		fmt.Printf("[INFO] Query key='wz'...\n")
		vec := tree.Get("wz")
		fmt.Printf("[INFO] Query key='wz' elapsed: %v\n", time.Since(now1))
		fmt.Printf("[INFO] Query key='wz' result count: %d\n", len(vec))
		fmt.Printf("[INFO] Query key='wz' result: %+v\n", vec)
	}
	{
		now2 := time.Now()
		fmt.Printf("[INFO] Query key='yan'...\n")
		vec := tree.Get("yan")
		fmt.Printf("[INFO] Query key='yan' elapsed: %v\n", time.Since(now2))
		fmt.Printf("[INFO] Query key='yan' result count: %d\n", len(vec))
		fmt.Printf("[INFO] Query key='yan' result: %+v\n", vec)
	}
	{
		now3 := time.Now()
		fmt.Printf("[INFO] Query key='心'...\n")
		vec := tree.Get("心")
		fmt.Printf("[INFO] Query key='心' elapsed: %v\n", time.Since(now3))
		fmt.Printf("[INFO] Query key='心' result count: %d\n", len(vec))
		fmt.Printf("[INFO] Query key='心' result: %+v\n", vec)
	}

	fmt.Printf("[INFO] Program will exit automatically after 100 seconds. You can check the output above.\n")
	time.Sleep(100 * time.Second)
}
