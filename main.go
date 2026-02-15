package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Finding represents a single detected heap allocation issue.
type Finding struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func main() {
	jsonOut := flag.Bool("json", false, "output as JSON")
	maxAllocs := flag.Int("max-allocs", 0, "fail if findings exceed threshold (CI mode)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: allocscope [flags] [directory]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	dir := "."
	if args := flag.Args(); len(args) > 0 {
		dir = args[0]
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %q is not a valid directory\n", dir)
		os.Exit(2)
	}

	var findings []Finding
	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if strings.Contains(path, "vendor"+string(os.PathSeparator)) {
			return nil
		}
		results, aerr := AnalyzeFile(path)
		if aerr != nil {
			fmt.Fprintf(os.Stderr, "warn: %s: %v\n", path, aerr)
			return nil
		}
		findings = append(findings, results...)
		return nil
	})

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(findings)
	} else {
		for _, f := range findings {
			fmt.Printf("%s:%d [%s] %s (%s)\n", f.File, f.Line, f.Severity, f.Message, f.Rule)
		}
		fmt.Printf("\n\U0001F50D AllocScope found %d potential heap allocations\n", len(findings))
	}

	if *maxAllocs > 0 && len(findings) > *maxAllocs {
		fmt.Fprintf(os.Stderr, "Exceeded max-allocs=%d (found %d)\n", *maxAllocs, len(findings))
		os.Exit(1)
	}
}
