package main

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzAnalyze(f *testing.F) {
	// Seed corpus from testdata/ directory
	testdataDir := "testdata"
	entries, err := os.ReadDir(testdataDir)
	if err == nil {
		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == ".go" {
				data, rerr := os.ReadFile(filepath.Join(testdataDir, entry.Name()))
				if rerr == nil {
					f.Add(string(data))
				}
			}
		}
	}

	// Inline seed corpus covering various patterns
	f.Add(`package main
func foo() {
	for i := 0; i < 10; i++ {
		_ = make([]byte, 1024)
	}
}`)

	f.Add(`package main
import "fmt"
func bar() string { return fmt.Sprintf("hello %s", "world") }`)

	f.Add(`package main
func baz() *int { x := 1; return &x }`)

	f.Add(`package main
func add(a, b int) int { return a + b }`)

	f.Add(`package main`)

	f.Add(`package main
func collect() []int {
	var out []int
	for i := 0; i < 10; i++ { out = append(out, i) }
	return out
}`)

	f.Add(`package main
func ptrs(vals []int) []*int {
	var out []*int
	for _, v := range vals {
		out = append(out, &v)
	}
	return out
}`)

	f.Add(`package main
import "fmt"
func dump(items []string) {
	for _, item := range items {
		fmt.Println(item)
	}
}`)

	f.Add(`package main
func spawn() {
	go func() {
		buf := make([]byte, 512)
		_ = buf
	}()
}`)

	f.Add(`package main
func cleanup() {
	defer func() {
		buf := make([]byte, 1024)
		_ = buf
	}()
}`)

	f.Add(`not valid go code at all`)

	f.Add(``)

	f.Fuzz(func(t *testing.T, src string) {
		// AnalyzeSource must never panic regardless of input.
		// Errors are expected for invalid Go source — that's fine.
		AnalyzeSource("fuzz.go", src)
	})
}
