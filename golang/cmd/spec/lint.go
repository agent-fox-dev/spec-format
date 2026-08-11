package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// lintFinding represents a single lint diagnostic result.
type lintFinding struct {
	Spec     string `json:"spec"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	File     string `json:"file,omitempty"`
}

// newLintCmd creates the "spec lint" subcommand which runs lint checks
// across all specs and reports findings as JSON.
func newLintCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint all specs for quality and consistency issues",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			specDir, _ := cmd.Flags().GetString("spec-dir")
			quiet, _ := cmd.Flags().GetBool("quiet")
			if isAgentMode() {
				quiet = true
			}
			w := cmd.OutOrStdout()

			return runLint(specDir, all, quiet, w, cmd)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "include fully-implemented specs in the lint run")

	return cmd
}

// runLint runs lint checks across all specs in the spec directory.
func runLint(specDir string, lintAll, quiet bool, w interface{ Write([]byte) (int, error) }, cmd *cobra.Command) error {
	// Handle non-existent spec directory gracefully.
	if _, err := os.Stat(specDir); os.IsNotExist(err) {
		return emitOKTo(w, "findings", []lintFinding{})
	}

	// Check if directory is readable.
	entries, err := os.ReadDir(specDir)
	if err != nil {
		return fmt.Errorf("cannot read spec directory %q: %w", specDir, err)
	}

	spinner := NewStatusSpinner("Linting specs...", cmd.ErrOrStderr(), quiet, false)
	spinner.Start()
	defer spinner.Stop()

	findings, exitCode := runLintSpecs(entries, specDir, lintAll)

	// Emit findings as JSON.
	result := map[string]any{
		"ok":       exitCode == 0,
		"findings": findings,
	}
	if err := emitTo(w, result); err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("lint found %d issue(s)", len(findings))
	}
	return nil
}

// runLintSpecs scans spec directories and produces lint findings.
// Returns findings and an exit code (0 = clean, 1 = issues found).
// When lintAll is false, specs with state "implemented" are skipped.
func runLintSpecs(entries []os.DirEntry, specDir string, lintAll bool) ([]lintFinding, int) {
	var findings []lintFinding

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if !specDirPattern.MatchString(e.Name()) {
			continue
		}

		specPath := filepath.Join(specDir, e.Name())

		// Skip fully-implemented specs unless --all is set.
		if !lintAll {
			state := readSpecState(specPath)
			if state == "implemented" {
				continue
			}
		}

		// Lint this spec.
		specFindings := lintSingleSpec(specPath, e.Name())
		findings = append(findings, specFindings...)
	}

	// Ensure findings is never nil (always an array in JSON).
	if findings == nil {
		findings = []lintFinding{}
	}

	exitCode := 0
	if len(findings) > 0 {
		exitCode = 1
	}

	return findings, exitCode
}

// lintSingleSpec runs lint checks on a single spec directory and returns findings.
func lintSingleSpec(specPath, specName string) []lintFinding {
	var findings []lintFinding

	// Check for required files.
	for _, f := range requiredFiles {
		p := filepath.Join(specPath, f)
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				findings = append(findings, lintFinding{
					Spec:     specName,
					Rule:     "missing-file",
					Severity: "error",
					Message:  fmt.Sprintf("required file %s is missing", f),
					File:     f,
				})
				continue
			}
			findings = append(findings, lintFinding{
				Spec:     specName,
				Rule:     "unreadable-file",
				Severity: "error",
				Message:  fmt.Sprintf("cannot read %s: %v", f, err),
				File:     f,
			})
			continue
		}

		// prd.md is not JSON.
		if f == "prd.md" {
			if len(data) == 0 {
				findings = append(findings, lintFinding{
					Spec:     specName,
					Rule:     "empty-prd",
					Severity: "warning",
					Message:  "prd.md is empty",
					File:     f,
				})
			}
			continue
		}

		// Verify JSON is parseable.
		if !json.Valid(data) {
			findings = append(findings, lintFinding{
				Spec:     specName,
				Rule:     "invalid-json",
				Severity: "error",
				Message:  fmt.Sprintf("%s contains malformed JSON", f),
				File:     f,
			})
			continue
		}

		// Check for empty content in JSON files.
		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err == nil {
			if len(parsed) == 0 {
				findings = append(findings, lintFinding{
					Spec:     specName,
					Rule:     "empty-artifact",
					Severity: "warning",
					Message:  fmt.Sprintf("%s has no content", f),
					File:     f,
				})
			}
		}
	}

	// Check for _session.json.
	sessionPath := filepath.Join(specPath, "_session.json")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		findings = append(findings, lintFinding{
			Spec:     specName,
			Rule:     "missing-session",
			Severity: "warning",
			Message:  "_session.json is missing",
			File:     "_session.json",
		})
	}

	return findings
}
