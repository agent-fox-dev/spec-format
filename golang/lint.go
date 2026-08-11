package afspec

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
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// SortFindings and ComputeExitCode (05-REQ-8)
// ---------------------------------------------------------------------------

// SortFindings returns a new slice sorted primarily by SpecName ascending,
// then by File ascending, then by Severity in the order error < warning < hint.
// The input slice is not modified.
func SortFindings(findings []LintFinding) []LintFinding {
	panic("not implemented")
}

// ComputeExitCode returns 1 if any finding has Severity "error", 0 otherwise.
func ComputeExitCode(findings []LintFinding) int {
	panic("not implemented")
}
