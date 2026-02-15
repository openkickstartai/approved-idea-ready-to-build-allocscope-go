package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ── Unit tests for GenerateSummary ──

func TestGenerateSummaryBasic(t *testing.T) {
	findings := []Finding{
		{"a.go", 10, "loop-alloc", "high", "make in loop"},
		{"a.go", 20, "fmt-alloc", "medium", "fmt.Sprintf"},
		{"b.go", 5, "ptr-escape", "high", "pointer escape"},
	}

	s := GenerateSummary(findings)

	if s.TotalFindings != 3 {
		t.Errorf("TotalFindings = %d, want 3", s.TotalFindings)
	}
	if s.BySeverity["high"] != 2 {
		t.Errorf("BySeverity[high] = %d, want 2", s.BySeverity["high"])
	}
	if s.BySeverity["medium"] != 1 {
		t.Errorf("BySeverity[medium] = %d, want 1", s.BySeverity["medium"])
	}
	if s.HottestFile != "a.go" {
		t.Errorf("HottestFile = %q, want %q", s.HottestFile, "a.go")
	}
	// FilesScanned = unique files with findings
	if s.FilesScanned != 2 {
		t.Errorf("FilesScanned = %d, want 2", s.FilesScanned)
	}
}

func TestGenerateSummaryEmpty(t *testing.T) {
	s := GenerateSummary(nil)
	if s.TotalFindings != 0 {
		t.Errorf("TotalFindings = %d, want 0", s.TotalFindings)
	}
	if s.HottestFile != "" {
		t.Errorf("HottestFile = %q, want empty", s.HottestFile)
	}
	if s.FilesScanned != 0 {
		t.Errorf("FilesScanned = %d, want 0", s.FilesScanned)
	}
}

func TestGenerateSummarySingleFile(t *testing.T) {
	findings := []Finding{
		{"x.go", 1, "r1", "low", "msg1"},
		{"x.go", 2, "r2", "low", "msg2"},
	}
	s := GenerateSummary(findings)
	if s.HottestFile != "x.go" {
		t.Errorf("HottestFile = %q, want x.go", s.HottestFile)
	}
	if s.BySeverity["low"] != 2 {
		t.Errorf("BySeverity[low] = %d, want 2", s.BySeverity["low"])
	}
}

// ── Unit tests for ShouldFail at various thresholds ──

func TestShouldFailHighThresholdWithHighFindings(t *testing.T) {
	s := Summary{
		TotalFindings: 2,
		BySeverity:    map[string]int{"high": 1, "medium": 1},
	}
	if !ShouldFail(s, SeverityHigh) {
		t.Error("expected ShouldFail=true for threshold=high with high findings")
	}
}

func TestShouldFailHighThresholdWithOnlyMedium(t *testing.T) {
	s := Summary{
		TotalFindings: 1,
		BySeverity:    map[string]int{"medium": 1},
	}
	if ShouldFail(s, SeverityHigh) {
		t.Error("expected ShouldFail=false for threshold=high with only medium findings")
	}
}

func TestShouldFailMediumThreshold(t *testing.T) {
	s := Summary{
		TotalFindings: 1,
		BySeverity:    map[string]int{"medium": 1},
	}
	if !ShouldFail(s, SeverityMedium) {
		t.Error("expected ShouldFail=true for threshold=medium with medium findings")
	}
}

func TestShouldFailCriticalThresholdWithHighFindings(t *testing.T) {
	s := Summary{
		TotalFindings: 1,
		BySeverity:    map[string]int{"high": 1},
	}
	if ShouldFail(s, SeverityCritical) {
		t.Error("expected ShouldFail=false for threshold=critical with only high findings")
	}
}

func TestShouldFailCriticalThresholdWithCriticalFindings(t *testing.T) {
	s := Summary{
		TotalFindings: 1,
		BySeverity:    map[string]int{"critical": 1},
	}
	if !ShouldFail(s, SeverityCritical) {
		t.Error("expected ShouldFail=true for threshold=critical with critical findings")
	}
}

func TestShouldFailNeverThreshold(t *testing.T) {
	s := Summary{
		TotalFindings: 5,
		BySeverity:    map[string]int{"critical": 2, "high": 2, "medium": 1},
	}
	if ShouldFail(s, SeverityNever) {
		t.Error("expected ShouldFail=false for threshold=never regardless of findings")
	}
}

func TestShouldFailLowThreshold(t *testing.T) {
	s := Summary{
		TotalFindings: 1,
		BySeverity:    map[string]int{"low": 1},
	}
	if !ShouldFail(s, SeverityLow) {
		t.Error("expected ShouldFail=true for threshold=low with low findings")
	}
}

func TestShouldFailEmptyFindings(t *testing.T) {
	s := Summary{
		TotalFindings: 0,
		BySeverity:    map[string]int{},
	}
	if ShouldFail(s, SeverityLow) {
		t.Error("expected ShouldFail=false with no findings")
	}
}

// ── Unit tests for ParseSeverity ──

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  Severity
	}{
		{"low", SeverityLow},
		{"medium", SeverityMedium},
		{"high", SeverityHigh},
		{"critical", SeverityCritical},
		{"never", SeverityNever},
		{"unknown", SeverityHigh},
	}
	for _, tt := range tests {
		got := ParseSeverity(tt.input)
		if got != tt.want {
			t.Errorf("ParseSeverity(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// ── Unit test for PrintSummary ──

func TestPrintSummaryContainsAllFields(t *testing.T) {
	s := Summary{
		TotalFindings: 5,
		BySeverity:    map[string]int{"high": 3, "medium": 2},
		FilesScanned:  10,
		HottestFile:   "server.go",
	}
	var buf bytes.Buffer
	PrintSummary(s, &buf)
	out := buf.String()

	for _, want := range []string{"Total findings:", "5", "high:", "3", "medium:", "2", "Files scanned:", "10", "Hottest file:", "server.go"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary output missing %q\ngot: %s", want, out)
		}
	}
}

// ── Integration tests using exec.Command to verify exit codes ──

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "allocscope")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = wd
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func writeTestFile(t *testing.T, dir, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, "example.go"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExitCodeFailOnHighWithHighFindings(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	// loop-fmt produces a "high" severity finding
	writeTestFile(t, dir, `package example
import "fmt"
func foo() {
	for i := 0; i < 10; i++ {
		_ = fmt.Sprintf("x")
	}
}
`)
	cmd := exec.Command(bin, "-fail-on", "high", dir)
	err := cmd.Run()
	if err == nil {
		t.Error("expected non-zero exit code with high findings and -fail-on=high")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
}

func TestExitCodeFailOnNeverAlwaysZero(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	writeTestFile(t, dir, `package example
import "fmt"
func foo() {
	for i := 0; i < 10; i++ {
		_ = fmt.Sprintf("x")
	}
}
`)
	cmd := exec.Command(bin, "-fail-on", "never", dir)
	err := cmd.Run()
	if err != nil {
		t.Errorf("expected exit code 0 with -fail-on=never, got error: %v", err)
	}
}

func TestExitCodeCleanCodeExitsZero(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	writeTestFile(t, dir, `package example
func add(a, b int) int { return a + b }
`)
	cmd := exec.Command(bin, "-fail-on", "high", dir)
	err := cmd.Run()
	if err != nil {
		t.Errorf("expected exit code 0 for clean code, got error: %v", err)
	}
}

func TestExitCodeFailOnCriticalWithOnlyHigh(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	// ptr-escape is "high" severity, not "critical"
	writeTestFile(t, dir, `package example
type S struct{ X int }
func baz() *S {
	s := S{X: 1}
	return &s
}
`)
	cmd := exec.Command(bin, "-fail-on", "critical", dir)
	err := cmd.Run()
	if err != nil {
		t.Errorf("expected exit code 0 with only high findings and -fail-on=critical, got: %v", err)
	}
}

func TestExitCodeFailOnMediumWithMediumFindings(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	// fmt-alloc outside loop is "medium"
	writeTestFile(t, dir, `package example
import "fmt"
func foo() string { return fmt.Sprintf("hello") }
`)
	cmd := exec.Command(bin, "-fail-on", "medium", dir)
	err := cmd.Run()
	if err == nil {
		t.Error("expected non-zero exit code with medium findings and -fail-on=medium")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
}

func TestExitCodeDefaultThresholdIsHigh(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	// ptr-escape is "high" severity
	writeTestFile(t, dir, `package example
type S struct{ X int }
func baz() *S {
	s := S{X: 1}
	return &s
}
`)
	// No -fail-on flag: default is "high"
	cmd := exec.Command(bin, dir)
	err := cmd.Run()
	if err == nil {
		t.Error("expected non-zero exit with high findings and default threshold")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
}

func TestSummaryOutputOnStderr(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	writeTestFile(t, dir, `package example
import "fmt"
func foo() string { return fmt.Sprintf("hello") }
`)
	cmd := exec.Command(bin, "-fail-on", "never", dir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stderr.String()
	for _, want := range []string{"Total findings:", "Files scanned:", "high:", "medium:"} {
		if !strings.Contains(out, want) {
			t.Errorf("stderr summary missing %q\ngot: %s", want, out)
		}
	}
}
