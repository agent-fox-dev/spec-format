package spec

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setupValidSpec creates a spec directory with all required files
// (prd.md, requirements.json, test_spec.json, tasks.json) containing
// valid content. Returns the specDir path.
func setupValidSpec(t *testing.T, tmpDir, specName string) string {
	t.Helper()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, specName)
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# Test PRD\n\nContent here."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "requirements.json"),
		[]byte(`{"requirements": [{"id": "REQ-1", "text": "Do something"}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "test_spec.json"),
		[]byte(`{"tests": [{"id": "TS-1", "name": "Test something"}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "tasks.json"),
		[]byte(`{"tasks": [{"id": "T-1", "name": "Implement something"}]}`), 0644); err != nil {
		t.Fatal(err)
	}

	return specDir
}

// --- TS-08-32: Verify that spec validate with a SPEC argument checks
//     required files, runs ValidateStructured, and emits ValidationResult
//     JSON; exits 1 on errors ---

// TestTS08_32_ValidateSingleSpecAllPresent verifies that running
// spec validate with a SPEC argument checks that all required files
// (prd.md, requirements.json, test_spec.json, tasks.json) exist,
// verifies JSON readability, runs ValidateStructured, and emits the
// ValidationResult as JSON.
// Covers: TS-08-32, Requirement: 08-REQ-11.1
func TestTS08_32_ValidateSingleSpecAllPresent(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupValidSpec(t, tmpDir, "08_my_spec")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "08_my_spec"})

	err := cmd.Execute()

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// The result should contain validation result fields.
	// Exit code depends on whether there are errors.
	if errorCount, ok := parsed["error_count"].(float64); ok && errorCount > 0 {
		if err == nil {
			t.Error("exit code should be 1 when error_count > 0")
		}
	} else {
		if err != nil {
			t.Fatalf("Execute() returned error: %v; want exit 0 for valid spec", err)
		}
	}
}

// TestTS08_32_ValidateSingleSpecExitCodeReflectsErrors verifies the
// correctness property 08-PROP-4: exit code is 1 if and only if the
// ValidationResult contains at least one error.
// Covers: TS-08-32, 08-PROP-4, Requirement: 08-REQ-11.1
func TestTS08_32_ValidateSingleSpecExitCodeReflectsErrors(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupValidSpec(t, tmpDir, "08_my_spec")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "08_my_spec"})

	err := cmd.Execute()

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	errorCount, _ := parsed["error_count"].(float64)
	if errorCount > 0 && err == nil {
		t.Error("exit code 0 with error_count > 0; want exit 1 (08-PROP-4)")
	}
	if errorCount == 0 && err != nil {
		t.Errorf("exit code 1 with error_count 0; want exit 0 (08-PROP-4); err: %v", err)
	}
}

// --- TS-08-33: Verify that spec validate without a SPEC argument discovers
//     all specs, validates each, aggregates results, and exits 1 if any
//     spec has errors ---

// TestTS08_33_ValidateMultiSpecAggregated verifies that running spec
// validate without a SPEC argument discovers all specs in the spec
// directory, validates each with ValidateStructured, aggregates the
// results, and exits 1 if any spec has errors.
// Covers: TS-08-33, Requirement: 08-REQ-11.2
func TestTS08_33_ValidateMultiSpecAggregated(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")

	// Create a valid spec.
	spec1 := filepath.Join(specDir, "08_spec_a")
	if err := os.MkdirAll(spec1, 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"prd.md", "requirements.json", "test_spec.json", "tasks.json"} {
		content := "# PRD"
		if f != "prd.md" {
			content = `{"valid": true}`
		}
		if err := os.WriteFile(filepath.Join(spec1, f), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a spec missing tasks.json.
	spec2 := filepath.Join(specDir, "09_spec_b")
	if err := os.MkdirAll(spec2, 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"prd.md", "requirements.json", "test_spec.json"} {
		content := "# PRD"
		if f != "prd.md" {
			content = `{"valid": true}`
		}
		if err := os.WriteFile(filepath.Join(spec2, f), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// tasks.json intentionally omitted for 09_spec_b.

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate"})

	err := cmd.Execute()

	// Should exit 1 because 09_spec_b is missing tasks.json.
	if err == nil {
		t.Error("Execute() returned nil; want exit 1 because spec_b is missing tasks.json")
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// Result should contain aggregated validation across both specs.
	// Verify at least error_count is present and > 0.
	if errorCount, ok := parsed["error_count"].(float64); !ok || errorCount == 0 {
		// Check for a nested results structure as an alternative.
		if _, exists := parsed["results"]; !exists {
			t.Error("expected error_count > 0 or nested results for multi-spec validation")
		}
	}
}

// --- TS-08-34: Verify that spec validate --cross discovers specs, builds
//     dependency graph, runs ValidateCrossSpec, and emits merged
//     ValidationResult JSON ---

// TestTS08_34_ValidateCrossSpec verifies that running spec validate
// with --cross discovers all specs, builds a dependency graph via
// BuildDependencyGraph, runs ValidateCrossSpec, and emits the merged
// ValidationResult as JSON.
// Covers: TS-08-34, Requirement: 08-REQ-11.3
func TestTS08_34_ValidateCrossSpec(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")

	// Create two valid specs with all required files.
	for _, specName := range []string{"08_spec_a", "09_spec_b"} {
		specPath := filepath.Join(specDir, specName)
		if err := os.MkdirAll(specPath, 0755); err != nil {
			t.Fatal(err)
		}
		for _, f := range []string{"prd.md", "requirements.json", "test_spec.json", "tasks.json"} {
			content := "# PRD for " + specName
			if f != "prd.md" {
				content = `{"valid": true}`
			}
			if err := os.WriteFile(filepath.Join(specPath, f), []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "--cross"})

	err := cmd.Execute()

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// Verify the result contains cross-spec validation data.
	// Exit code depends on whether cross-spec errors exist.
	if errorCount, ok := parsed["error_count"].(float64); ok && errorCount > 0 {
		if err == nil {
			t.Error("exit code should be 1 when cross-spec error_count > 0")
		}
	} else if err != nil {
		t.Fatalf("Execute() returned error: %v; want exit 0 when no cross-spec errors", err)
	}
}

// --- TS-08-35: Verify that spec validate --short emits condensed output
//     with only valid, error_count, and warning_count fields ---

// TestTS08_35_ValidateShort verifies that running spec validate with
// --short emits condensed output containing exactly ok, valid,
// error_count, and warning_count fields.
// Covers: TS-08-35, Requirement: 08-REQ-11.4
func TestTS08_35_ValidateShort(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupValidSpec(t, tmpDir, "08_my_spec")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "08_my_spec", "--short"})

	err := cmd.Execute()
	_ = err // exit code depends on validation result

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// Verify required fields are present.
	if _, exists := parsed["ok"]; !exists {
		t.Error("parsed missing 'ok' field")
	}
	if _, exists := parsed["valid"]; !exists {
		t.Error("parsed missing 'valid' field")
	}
	if _, exists := parsed["error_count"]; !exists {
		t.Error("parsed missing 'error_count' field")
	}
	if _, exists := parsed["warning_count"]; !exists {
		t.Error("parsed missing 'warning_count' field")
	}

	// Verify types.
	if _, ok := parsed["error_count"].(float64); !ok {
		t.Errorf("error_count is %T; want number", parsed["error_count"])
	}
	if _, ok := parsed["warning_count"].(float64); !ok {
		t.Errorf("warning_count is %T; want number", parsed["warning_count"])
	}
}

// TestTS08_35_ValidateShortFieldsOnly verifies that --short output
// contains ONLY the condensed fields (ok, valid, error_count,
// warning_count) and no extra verbose fields like findings or results.
// Covers: TS-08-35, Requirement: 08-REQ-11.4
func TestTS08_35_ValidateShortFieldsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupValidSpec(t, tmpDir, "08_my_spec")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "08_my_spec", "--short"})

	_ = cmd.Execute()

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// The only allowed keys are: ok, valid, error_count, warning_count.
	allowedKeys := map[string]bool{
		"ok":            true,
		"valid":         true,
		"error_count":   true,
		"warning_count": true,
	}
	for key := range parsed {
		if !allowedKeys[key] {
			t.Errorf("--short output contains unexpected key %q; want only ok/valid/error_count/warning_count", key)
		}
	}
}

// --- 08-REQ-11.E1: Missing required file reported as validation error ---

// TestTS08_32_ValidateMissingRequiredFile verifies that when a required
// file (prd.md, requirements.json, test_spec.json, or tasks.json) is
// missing in single-spec mode, the missing file is reported as a
// validation error in the result (not a command failure) and exit code is 1.
// Covers: 08-REQ-11.E1
func TestTS08_32_ValidateMissingRequiredFile(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Only create prd.md and requirements.json — tasks.json and test_spec.json missing.
	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "requirements.json"),
		[]byte(`{"requirements": []}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "08_my_spec"})

	err := cmd.Execute()

	// Should exit 1 — missing required files are validation errors.
	if err == nil {
		t.Error("Execute() returned nil; want exit 1 for missing required files")
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	// The result should mention the missing files as validation errors.
	if errorCount, ok := parsed["error_count"].(float64); !ok || errorCount == 0 {
		t.Error("error_count is 0 or missing; want > 0 for missing required files")
	}
}

// --- 08-REQ-11.E2: Malformed JSON reported as validation error ---

// TestTS08_32_ValidateMalformedJSON verifies that when a JSON artifact
// file contains malformed JSON, the parse error is reported as a
// validation error in the result.
// Covers: 08-REQ-11.E2
func TestTS08_32_ValidateMalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}
	// Malformed JSON in requirements.json.
	if err := os.WriteFile(filepath.Join(specPath, "requirements.json"),
		[]byte(`{not valid json!!!`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "test_spec.json"),
		[]byte(`{"tests": []}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specPath, "tasks.json"),
		[]byte(`{"tasks": []}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "08_my_spec"})

	err := cmd.Execute()

	// Should exit 1 — parse error is a validation error.
	if err == nil {
		t.Error("Execute() returned nil; want exit 1 for malformed JSON")
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	if errorCount, ok := parsed["error_count"].(float64); !ok || errorCount == 0 {
		t.Error("error_count is 0 or missing; want > 0 for malformed JSON")
	}
}

// --- 08-REQ-11.E3: DiscoverSpecs error propagation ---

// TestTS08_33_ValidateDiscoverError verifies that when the spec
// directory does not exist or cannot be read in multi-spec mode,
// the error is propagated and the command exits without emitting
// a partial result.
// Covers: 08-REQ-11.E3
func TestTS08_33_ValidateDiscoverError(t *testing.T) {
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
	stderrBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() with unreadable spec directory returned nil; want error")
	}
}

// --- Validate: non-existent spec in single-spec mode ---

// TestTS08_32_ValidateNonexistentSpec verifies that spec validate
// returns an error when the referenced spec does not exist.
// Covers: 08-REQ-11.1
func TestTS08_32_ValidateNonexistentSpec(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "nonexistent_spec"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with non-existent spec returned nil; want error")
	}
}

// --- 08-PROP-4: Exit code reflects error presence ---

// TestTS08_32_ValidateExitCodeProperty verifies the correctness
// property 08-PROP-4 across multiple scenarios: exit code 1 iff
// validation errors exist.
// Covers: 08-PROP-4
func TestTS08_32_ValidateExitCodeProperty(t *testing.T) {
	scenarios := []struct {
		name      string
		wantError bool
		setup     func(t *testing.T, specDir string)
	}{
		{
			name:      "all_files_valid",
			wantError: false,
			setup: func(t *testing.T, specDir string) {
				specPath := filepath.Join(specDir, "08_my_spec")
				os.MkdirAll(specPath, 0755)
				os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# PRD"), 0644)
				os.WriteFile(filepath.Join(specPath, "requirements.json"), []byte(`{"requirements":[]}`), 0644)
				os.WriteFile(filepath.Join(specPath, "test_spec.json"), []byte(`{"tests":[]}`), 0644)
				os.WriteFile(filepath.Join(specPath, "tasks.json"), []byte(`{"tasks":[]}`), 0644)
			},
		},
		{
			name:      "missing_tasks_json",
			wantError: true,
			setup: func(t *testing.T, specDir string) {
				specPath := filepath.Join(specDir, "08_my_spec")
				os.MkdirAll(specPath, 0755)
				os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# PRD"), 0644)
				os.WriteFile(filepath.Join(specPath, "requirements.json"), []byte(`{"requirements":[]}`), 0644)
				os.WriteFile(filepath.Join(specPath, "test_spec.json"), []byte(`{"tests":[]}`), 0644)
				// tasks.json intentionally missing
			},
		},
		{
			name:      "malformed_json",
			wantError: true,
			setup: func(t *testing.T, specDir string) {
				specPath := filepath.Join(specDir, "08_my_spec")
				os.MkdirAll(specPath, 0755)
				os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# PRD"), 0644)
				os.WriteFile(filepath.Join(specPath, "requirements.json"), []byte(`INVALID`), 0644)
				os.WriteFile(filepath.Join(specPath, "test_spec.json"), []byte(`{"tests":[]}`), 0644)
				os.WriteFile(filepath.Join(specPath, "tasks.json"), []byte(`{"tasks":[]}`), 0644)
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, ".specs")
			sc.setup(t, specDir)

			cmd := newRootCmd()
			stdoutBuf := new(bytes.Buffer)
			cmd.SetOut(stdoutBuf)
			cmd.SetErr(new(bytes.Buffer))
			cmd.SetArgs([]string{"--spec-dir", specDir, "validate", "08_my_spec"})

			err := cmd.Execute()
			if sc.wantError && err == nil {
				t.Error("Execute() returned nil; want exit 1 for validation errors")
			}
			if !sc.wantError && err != nil {
				t.Errorf("Execute() returned error: %v; want exit 0 for no validation errors", err)
			}
		})
	}
}
