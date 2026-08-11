package spec

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupRenderSpec creates a spec directory with artifact files for render
// tests. Returns the specDir path.
func setupRenderSpec(t *testing.T, tmpDir string) string {
	t.Helper()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create prd.md.
	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# Test PRD\n\nThis is a test PRD."), 0644); err != nil {
		t.Fatal(err)
	}

	// Create artifact files with markdown-compatible content.
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

// --- TS-08-26: Verify that spec render without --json and without --combined
//     prints raw markdown for each artifact separated by '---' ---

// TestTS08_26_RenderRawMarkdownSeparated verifies that running spec
// render without --json and without --combined prints raw markdown
// for each artifact to stdout, separated by '---'. The output should
// NOT be a JSON envelope.
// Covers: TS-08-26, Requirement: 08-REQ-10.1
func TestTS08_26_RenderRawMarkdownSeparated(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupRenderSpec(t, tmpDir)

	t.Setenv("AF_AGENT", "")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdoutBuf.String()

	// Output should NOT start with '{' (not JSON).
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Error("output starts with '{'; want raw markdown, not JSON")
	}

	// Output should contain '---' separator between artifacts.
	if !strings.Contains(output, "---") {
		t.Error("output does not contain '---' separator; want artifacts separated by '---'")
	}

	// There should be at least 2 sections when split by '---'.
	sections := strings.Split(output, "---")
	if len(sections) < 2 {
		t.Errorf("len(sections split by '---') = %d; want >= 2", len(sections))
	}
}

// --- TS-08-27: Verify that spec render --combined without --json prints
//     a single concatenated markdown document to stdout ---

// TestTS08_27_RenderCombinedMarkdown verifies that running spec render
// with --combined but without --json prints a single concatenated
// markdown document to stdout. The output should NOT be a JSON envelope.
// Covers: TS-08-27, Requirement: 08-REQ-10.2
func TestTS08_27_RenderCombinedMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupRenderSpec(t, tmpDir)

	t.Setenv("AF_AGENT", "")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec", "--combined"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdoutBuf.String()

	// Output should NOT start with '{' (not JSON).
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Error("output starts with '{'; want raw markdown, not JSON")
	}

	// Output should contain content (not empty).
	if len(strings.TrimSpace(output)) == 0 {
		t.Error("output is empty; want combined markdown document")
	}
}

// --- TS-08-28: Verify that spec render --json --combined emits
//     {"ok": true, "format": "combined", "content": "<markdown>"} ---

// TestTS08_28_RenderJSONCombined verifies that running spec render with
// --json and --combined emits a JSON envelope with ok: true, format:
// "combined", and content containing the combined markdown.
// Covers: TS-08-28, Requirement: 08-REQ-10.3
func TestTS08_28_RenderJSONCombined(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupRenderSpec(t, tmpDir)

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec", "--json", "--combined"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", err, output)
	}

	if ok, _ := parsed["ok"].(bool); !ok {
		t.Errorf("parsed.ok = %v; want true", parsed["ok"])
	}
	if format, _ := parsed["format"].(string); format != "combined" {
		t.Errorf("parsed.format = %q; want %q", format, "combined")
	}
	content, ok := parsed["content"].(string)
	if !ok || len(content) == 0 {
		t.Errorf("parsed.content = %v; want non-empty string", parsed["content"])
	}
}

// --- TS-08-29: Verify that spec render --json without --combined emits
//     {"ok": true, "format": "individual", "artifacts": {...}} ---

// TestTS08_29_RenderJSONIndividual verifies that running spec render
// with --json but without --combined emits a JSON envelope with ok:
// true, format: "individual", and an artifacts object keyed by artifact name.
// Covers: TS-08-29, Requirement: 08-REQ-10.4
func TestTS08_29_RenderJSONIndividual(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupRenderSpec(t, tmpDir)

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", err, output)
	}

	if ok, _ := parsed["ok"].(bool); !ok {
		t.Errorf("parsed.ok = %v; want true", parsed["ok"])
	}
	if format, _ := parsed["format"].(string); format != "individual" {
		t.Errorf("parsed.format = %q; want %q", format, "individual")
	}

	artifacts, ok := parsed["artifacts"].(map[string]any)
	if !ok {
		t.Fatal("parsed.artifacts is not an object")
	}
	if _, exists := artifacts["requirements"]; !exists {
		t.Error("parsed.artifacts missing 'requirements' key")
	}
}

// --- TS-08-30: Verify that spec render auto-enables --json mode when
//     AF_AGENT=1 is active ---

// TestTS08_30_RenderAgentModeAutoJSON verifies that when AF_AGENT=1 is
// set, spec render without --json flag auto-enables JSON mode, producing
// valid JSON output following the --json contract.
// Covers: TS-08-30, Requirement: 08-REQ-10.5
func TestTS08_30_RenderAgentModeAutoJSON(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := setupRenderSpec(t, tmpDir)

	t.Setenv("AF_AGENT", "1")

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	// No --json flag — agent mode should auto-enable it.
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON (agent mode should auto-enable --json): %v\noutput: %s", err, output)
	}

	if ok, _ := parsed["ok"].(bool); !ok {
		t.Errorf("parsed.ok = %v; want true", parsed["ok"])
	}

	format, _ := parsed["format"].(string)
	if format != "individual" && format != "combined" {
		t.Errorf("parsed.format = %q; want 'individual' or 'combined'", format)
	}
}

// --- TS-08-31: Verify that spec render falls back to rendering only
//     available artifacts when some artifact files are missing ---

// TestTS08_31_RenderFallbackMissingArtifacts verifies that when some
// artifact files are missing, spec render falls back to rendering only
// the available ones. The output includes only existing artifacts.
// Covers: TS-08-31, Requirement: 08-REQ-10.6
func TestTS08_31_RenderFallbackMissingArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create prd.md.
	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	// Only create requirements.json — leave test_spec.json and tasks.json missing.
	if err := os.WriteFile(filepath.Join(specPath, "requirements.json"),
		[]byte(`{"requirements": [{"id": "REQ-1", "text": "Test"}]}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", err, output)
	}

	if ok, _ := parsed["ok"].(bool); !ok {
		t.Errorf("parsed.ok = %v; want true", parsed["ok"])
	}

	artifacts, ok := parsed["artifacts"].(map[string]any)
	if !ok {
		t.Fatal("parsed.artifacts is not an object")
	}

	// requirements should be present.
	if _, exists := artifacts["requirements"]; !exists {
		t.Error("parsed.artifacts missing 'requirements' key; want it present since file exists")
	}

	// test_spec should not be present (or empty) since the file is missing.
	if ts, exists := artifacts["test_spec"]; exists {
		if s, ok := ts.(string); ok && s != "" {
			t.Errorf("parsed.artifacts.test_spec = %q; want absent or empty for missing file", s)
		}
	}
}

// --- 08-REQ-10.E1: All artifact files are missing ---

// TestTS08_26_RenderAllArtifactsMissing verifies that when ALL artifact
// files are missing (no requirements.json, test_spec.json, or tasks.json),
// the render command returns an error indicating no renderable artifacts.
// Covers: 08-REQ-10.E1
func TestTS08_26_RenderAllArtifactsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Only prd.md — no artifact files.
	if err := os.WriteFile(filepath.Join(specPath, "prd.md"),
		[]byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with all artifacts missing returned nil; want error")
	}
}

// --- 08-REQ-10.E2: Artifact file permission error ---

// TestTS08_26_RenderPermissionError verifies that when an artifact file
// exists but cannot be read due to permissions, the render command
// returns an OS-level error identifying the unreadable file.
// Covers: 08-REQ-10.E2
func TestTS08_26_RenderPermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

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

	// Create an unreadable artifact file.
	reqsPath := filepath.Join(specPath, "requirements.json")
	if err := os.WriteFile(reqsPath,
		[]byte(`{"requirements": []}`), 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(reqsPath, 0644)

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "08_my_spec"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() with unreadable artifact returned nil; want error")
	}
}

// --- Render: missing spec argument ---

// TestTS08_26_RenderMissingSpecArg verifies that spec render requires
// a positional spec argument.
// Covers: 08-REQ-10
func TestTS08_26_RenderMissingSpecArg(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"render"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with no spec argument returned nil; want error")
	}
}

// --- Render: non-existent spec ---

// TestTS08_26_RenderNonexistentSpec verifies that spec render returns
// an error when the referenced spec does not exist.
// Covers: 08-REQ-10
func TestTS08_26_RenderNonexistentSpec(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "render", "nonexistent_spec"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with non-existent spec returned nil; want error")
	}
}
