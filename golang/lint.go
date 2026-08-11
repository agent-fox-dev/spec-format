package afspec

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// ---------------------------------------------------------------------------
// Lint module data structures (05-REQ-6)
// ---------------------------------------------------------------------------

// LintFinding represents a single lint diagnostic result with severity, rule,
// message, and location metadata.
type LintFinding struct {
	SpecName string
	File     string
	Rule     string
	Severity string // one of "error", "warning", "hint"
	Message  string
	Line     int // zero means absent
}

// LintSpecInfo contains metadata about a discovered spec folder eligible for
// linting.
type LintSpecInfo struct {
	Name     string
	Prefix   int
	Path     string
	HasTasks bool
	HasPRD   bool
}

// LintResult is the aggregate output of a lint run, containing all findings
// and an exit code.
type LintResult struct {
	Findings []LintFinding
	ExitCode int
}

// severityRank maps severity strings to sort ranks. Unknown severities
// are treated as "hint" (rank 2) per 05-REQ-6.E1.
func severityRank(s string) int {
	switch s {
	case "error":
		return 0
	case "warning":
		return 1
	case "hint":
		return 2
	default:
		return 2 // unknown severity treated as hint
	}
}

// ---------------------------------------------------------------------------
// DiscoverLintSpecs — spec folder discovery (05-REQ-7)
// ---------------------------------------------------------------------------

// DiscoverLintSpecs scans specsDir for subdirectories whose names are valid
// spec directory names, checks each for the presence of requirements.json,
// and returns a LintSpecInfo slice for all matching directories.
//
// When filterSpec is non-empty, only the entry whose Name matches exactly is
// returned; an error is returned if no match is found.
//
// HasTasks and HasPRD are set based on the presence of tasks.json and prd.md
// respectively.
func DiscoverLintSpecs(specsDir, filterSpec string) ([]LintSpecInfo, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read specs directory %q: %w", specsDir, err)
	}

	var infos []LintSpecInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()

		// 05-REQ-7.1: use IsSpecDirName to check validity.
		if !IsSpecDirName(name) {
			continue
		}

		dir := filepath.Join(specsDir, name)

		// Only include directories that contain requirements.json.
		if _, err := os.Stat(filepath.Join(dir, "requirements.json")); err != nil {
			continue
		}

		// 05-REQ-7.E3: skip if ParseSpecDirName fails.
		prefixStr, specName, parseErr := ParseSpecDirName(name)
		if parseErr != nil {
			continue
		}

		prefix, convErr := strconv.Atoi(prefixStr)
		if convErr != nil {
			continue
		}

		hasTasks := fileExists(filepath.Join(dir, "tasks.json"))
		hasPRD := fileExists(filepath.Join(dir, "prd.md"))

		infos = append(infos, LintSpecInfo{
			Name:     specName,
			Prefix:   prefix,
			Path:     dir,
			HasTasks: hasTasks,
			HasPRD:   hasPRD,
		})
	}

	// 05-REQ-7.E2: error if no specs discovered at all.
	if len(infos) == 0 {
		return nil, fmt.Errorf("no specs found in %q", specsDir)
	}

	// 05-REQ-7.2: filter by name when filterSpec is non-empty.
	if filterSpec != "" {
		for _, info := range infos {
			if info.Name == filterSpec {
				return []LintSpecInfo{info}, nil
			}
		}
		return nil, fmt.Errorf("no spec with name %q found in %q", filterSpec, specsDir)
	}

	return infos, nil
}

// fileExists returns true if the path exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ---------------------------------------------------------------------------
// SortFindings and ComputeExitCode (05-REQ-8)
// ---------------------------------------------------------------------------

// SortFindings returns a new slice sorted primarily by SpecName ascending,
// then by File ascending, then by Severity in the order error < warning < hint.
// The input slice is not modified.
func SortFindings(findings []LintFinding) []LintFinding {
	sorted := make([]LintFinding, len(findings))
	copy(sorted, findings)

	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if a.SpecName != b.SpecName {
			return a.SpecName < b.SpecName
		}
		if a.File != b.File {
			return a.File < b.File
		}
		return severityRank(a.Severity) < severityRank(b.Severity)
	})

	return sorted
}

// ComputeExitCode returns 1 if any finding has Severity "error", 0 otherwise.
func ComputeExitCode(findings []LintFinding) int {
	for _, f := range findings {
		if f.Severity == "error" {
			return 1
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// RunLintSpecs — full lint execution (05-REQ-9)
// ---------------------------------------------------------------------------

// RunLintSpecs scans specsDir for spec folders, loads and validates each,
// and returns a LintResult with sorted findings and computed exit code.
// When lintAll is false, fully-implemented specs (all subtasks in state
// 'done' or 'dropped') are skipped.
func RunLintSpecs(specsDir string, lintAll bool) (LintResult, error) {
	panic("not implemented")
}
