package main

import "testing"

func hasRule(fs []Finding, rule string) bool {
	for _, f := range fs {
		if f.Rule == rule {
			return true
		}
	}
	return false
}

func TestFmtAlloc(t *testing.T) {
	src := `package main
import "fmt"
func foo() string { return fmt.Sprintf("hello %s", "world") }`
	findings, err := AnalyzeSource("test.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, "fmt-alloc") {
		t.Error("expected fmt-alloc finding")
	}
}

func TestLoopMakeAlloc(t *testing.T) {
	src := `package main
func bar() {
	for i := 0; i < 100; i++ {
		s := make([]byte, 1024)
		_ = s
	}
}`
	findings, err := AnalyzeSource("test.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, "loop-alloc") {
		t.Error("expected loop-alloc finding")
	}
}

func TestPointerEscape(t *testing.T) {
	src := `package main
type S struct{ X int }
func baz() *S {
	s := S{X: 1}
	return &s
}`
	findings, err := AnalyzeSource("test.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, "ptr-escape") {
		t.Error("expected ptr-escape finding")
	}
}

func TestCleanCode(t *testing.T) {
	src := `package main
func add(a, b int) int { return a + b }`
	findings, err := AnalyzeSource("test.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestLoopAppend(t *testing.T) {
	src := `package main
func collect() []int {
	var out []int
	for i := 0; i < 10; i++ { out = append(out, i) }
	return out
}`
	findings, err := AnalyzeSource("test.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, "loop-append") {
		t.Error("expected loop-append finding")
	}
}

func TestLoopFmt(t *testing.T) {
	src := "package main\nimport \"fmt\"\nfunc f() {\n\tfor i := 0; i < 10; i++ { fmt.Sprintf(\"%d\", i) }\n}"
	findings, err := AnalyzeSource("test.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, "loop-fmt") {
		t.Error("expected loop-fmt finding")
	}
}

func TestLoopAddressOf(t *testing.T) {
	src := `package main
type Item struct{ V int }
func ptrs() []*Item {
	var out []*Item
	for i := 0; i < 5; i++ {
		v := Item{V: i}
		out = append(out, &v)
	}
	return out
}`
	findings, err := AnalyzeSource("test.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, "loop-escape") {
		t.Error("expected loop-escape finding")
	}
}

func TestInvalidSource(t *testing.T) {
	_, err := AnalyzeSource("bad.go", "not valid go")
	if err == nil {
		t.Error("expected parse error")
	}
}
