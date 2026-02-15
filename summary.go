package main

import (
	"fmt"
	"io"
)

// Severity represents the severity level of a finding.
type Severity int

const (
	SeverityLow      Severity = 0
	SeverityMedium   Severity = 1
	SeverityHigh     Severity = 2
	SeverityCritical Severity = 3
	SeverityNever    Severity = 100 // sentinel: never fail
)

// Summary holds aggregated statistics about scan findings.
type Summary struct {
	TotalFindings int
	BySeverity    map[string]int
	FilesScanned  int
	HottestFile   string
}

// ParseSeverity converts a string severity name to a Severity value.
func ParseSeverity(s string) Severity {
	switch s {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	case "never":
		return SeverityNever
	default:
		return SeverityHigh
	}
}

// GenerateSummary computes aggregate statistics from a list of findings.
func GenerateSummary(findings []Finding) Summary {
	summary := Summary{
		TotalFindings: len(findings),
		BySeverity:    make(map[string]int),
	}

	fileCounts := make(map[string]int)

	for _, f := range findings {
		summary.BySeverity[f.Severity]++
		fileCounts[f.File]++
	}

	// FilesScanned defaults to unique files with findings;
	// callers may override with the actual walk count.
	summary.FilesScanned = len(fileCounts)

	maxCount := 0
	for file, count := range fileCounts {
		if count > maxCount {
			maxCount = count
			summary.HottestFile = file
		}
	}

	return summary
}

// ShouldFail returns true when findings exist at or above the given severity threshold.
func ShouldFail(summary Summary, threshold Severity) bool {
	for sev, count := range summary.BySeverity {
		if count > 0 && ParseSeverity(sev) >= threshold {
			return true
		}
	}
	return false
}

// PrintSummary writes a human-readable summary to the given writer.
func PrintSummary(summary Summary, w io.Writer) {
	fmt.Fprintf(w, "\n── AllocScope Summary ──\n")
	fmt.Fprintf(w, "Total findings:  %d\n", summary.TotalFindings)
	for _, sev := range []string{"critical", "high", "medium", "low"} {
		count := summary.BySeverity[sev]
		fmt.Fprintf(w, "  %-10s %d\n", sev+":", count)
	}
	fmt.Fprintf(w, "Files scanned:   %d\n", summary.FilesScanned)
	if summary.HottestFile != "" {
		fmt.Fprintf(w, "Hottest file:    %s\n", summary.HottestFile)
	}
}
