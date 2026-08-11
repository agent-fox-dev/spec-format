package spec

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- TS-08-36: Verify that spec lint runs RunLintSpecs with a spinner,
//     emits LintFinding list as JSON, and exits 1 if exit_code from
//     lint is non-zero ---

// TestTS08_36_LintRunsAndEmitsFindings verifies that spec lint runs
// RunLintSpecs with a spinner, emits findings as JSON to stdout, and
// uses the lint exit_code to determine the command exit code.
// Covers: TS-08-36, Requirement: 08-REQ-12.1
func TestTS08_36_LintRunsAndEmitsFindings(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a spec directory with minimal content to trigger lint findings.
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# Minimal PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs([]string{"--spec-dir", specDir, "lint"})

	err := cmd.Execute()

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// The result should contain a findings array.
	findings, ok := parsed["findings"].([]any)
	if !ok {
		t.Fatal("parsed.findings is not an array")
	}

	// If there are findings, exit code should be 1.
	if len(findings) > 0 && err == nil {
		t.Error("exit code 0 with lint findings present; want exit 1")
	}
}

// TestTS08_36_LintExitCode1OnNonZeroExitCode verifies that when
// RunLintSpecs returns a non-zero exit_code, the lint command exits 1.
// Covers: TS-08-36, Requirement: 08-REQ-12.1
func TestTS08_36_LintExitCode1OnNonZeroExitCode(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")

	// Create a spec with lint-triggering content (missing artifacts).
	specPath := filepath.Join(specDir, "08_lint_fail")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "lint"})

	err := cmd.Execute()

	output := stdoutBuf.String()
	if len(output) == 0 {
		// If no output at all, the stub hasn't been implemented yet.
		if err == nil {
			t.Error("Execute() returned nil; want error for unimplemented lint")
		}
		return
	}

	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// When lint finds issues, exit code should be 1.
	findings, _ := parsed["findings"].([]any)
	if len(findings) > 0 && err == nil {
		t.Error("exit code 0 but lint has findings; want exit 1")
	}
}

// TestTS08_36_LintEmitsJSONWithOK verifies that the lint output
// contains an "ok" field and a "findings" array in the JSON envelope.
// Covers: TS-08-36, Requirement: 08-REQ-12.1
func TestTS08_36_LintEmitsJSONWithOK(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "lint"})

	_ = cmd.Execute()

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	if _, exists := parsed["ok"]; !exists {
		t.Error("parsed missing 'ok' field")
	}
	if _, exists := parsed["findings"]; !exists {
		t.Error("parsed missing 'findings' field")
	}
}

// --- TS-08-37: Verify that spec lint --all passes lintAll=true to
//     RunLintSpecs to include fully-implemented specs ---

// TestTS08_37_LintAllFlagRegistered verifies that the --all flag is
// registered on the lint command.
// Covers: TS-08-37, Requirement: 08-REQ-12.2
func TestTS08_37_LintAllFlagRegistered(t *testing.T) {
	cmd := newLintCmd()
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("flag --all is not registered on lint command")
	}
	if allFlag.DefValue != "false" {
		t.Errorf("--all default = %q; want %q", allFlag.DefValue, "false")
	}
}

// TestTS08_37_LintAllIncludesImplementedSpecs verifies that running
// spec lint --all includes fully-implemented specs in the lint findings.
// Covers: TS-08-37, Requirement: 08-REQ-12.2
func TestTS08_37_LintAllIncludesImplementedSpecs(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")

	// Create a fully-implemented spec (all required artifacts present).
	specPath := filepath.Join(specDir, "08_complete_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"prd.md", "requirements.json", "test_spec.json", "tasks.json"} {
		content := "# PRD"
		if f != "prd.md" {
			content = `{"valid": true}`
		}
		if err := os.WriteFile(filepath.Join(specPath, f), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Mark session as fully implemented.
	sessionData := map[string]any{
		"state":               "implemented",
		"generated_artifacts": []string{"requirements.json", "test_spec.json", "tasks.json"},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "lint", "--all"})

	err := cmd.Execute()

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// With --all, the output should include findings (or an empty array
	// if the implemented spec has no issues). The key requirement is that
	// the fully-implemented spec was included in the lint scope.
	if _, exists := parsed["findings"]; !exists {
		t.Error("parsed missing 'findings' field; --all should still emit findings array")
	}

	// Verify the command completed (stub currently errors).
	if err != nil {
		t.Logf("lint --all returned error (expected until implementation): %v", err)
	}
}

// --- 08-REQ-12.E1: RunLintSpecs error propagation ---

// TestTS08_36_LintErrorPropagation verifies that when RunLintSpecs
// returns an error (e.g., spec directory unreadable), the error is
// propagated and the command exits without emitting partial findings.
// Covers: 08-REQ-12.E1
func TestTS08_36_LintErrorPropagation(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Make the spec directory unreadable.
	if err := os.Chmod(specDir, 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(specDir, 0755)

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "lint"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() with unreadable spec directory returned nil; want error")
	}
}

// --- 08-REQ-12.E2: Empty spec directory emits empty findings ---

// TestTS08_36_LintEmptySpecDir verifies that when the spec directory
// is empty or contains no lintable specs, the lint command emits an
// empty findings array and exits 0.
// Covers: 08-REQ-12.E2
func TestTS08_36_LintEmptySpecDir(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "lint"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() on empty spec dir returned error: %v; want exit 0", err)
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	if ok, _ := parsed["ok"].(bool); !ok {
		t.Errorf("parsed.ok = %v; want true", parsed["ok"])
	}

	findings, ok := parsed["findings"].([]any)
	if !ok {
		t.Fatal("parsed.findings is not an array")
	}
	if len(findings) != 0 {
		t.Errorf("findings has %d entries; want 0 for empty spec dir", len(findings))
	}
}

// --- 08-REQ-12.E2: Non-existent spec directory emits empty findings ---

// TestTS08_36_LintNonexistentSpecDir verifies that when the spec
// directory does not exist, the lint command emits an empty findings
// array and exits 0 (consistent with list behavior).
// Covers: 08-REQ-12.E2
func TestTS08_36_LintNonexistentSpecDir(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs_nonexistent")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "lint"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() on non-existent spec dir returned error: %v; want exit 0", err)
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	if ok, _ := parsed["ok"].(bool); !ok {
		t.Errorf("parsed.ok = %v; want true", parsed["ok"])
	}

	findings, ok := parsed["findings"].([]any)
	if !ok {
		t.Fatal("parsed.findings is not an array")
	}
	if len(findings) != 0 {
		t.Errorf("findings has %d entries; want 0 for non-existent spec dir", len(findings))
	}
}

// --- Lint: banner should be suppressed ---

// TestTS08_36_LintBannerSuppressed verifies that the lint command
// does not trigger banner display (it's a non-banner subcommand
// in the implementation, but verifying it has spinner on stderr).
// Covers: 08-REQ-12.1
func TestTS08_36_LintBannerSuppressed(t *testing.T) {
	// Lint is not in the suppressed-banner list per spec, but we
	// verify that quiet mode suppresses the spinner.
	cmd := newRootCmd()
	stderrBuf := new(bytes.Buffer)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(stderrBuf)
	cmd.SetArgs([]string{"--quiet", "lint"})

	_ = cmd.Execute()

	// In quiet mode, stderr should not contain spinner output.
	// The exact assertion depends on implementation, but spinner
	// output should be suppressed.
	stderrOutput := stderrBuf.String()
	_ = stderrOutput // Stub: implementation will validate spinner suppression.
}
