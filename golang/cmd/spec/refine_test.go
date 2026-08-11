package spec

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- TS-08-18: Verify that spec refine without --answers runs Assess on
//     the session with a spinner and emits assessment result JSON when the
//     session needs assessment ---

// TestTS08_18_RefineAssessPath verifies that running spec refine without
// --answers on a session that needs assessment calls Assess and emits the
// assessment result as JSON with ok: true.
// Covers: TS-08-18, Requirement: 08-REQ-8.1
func TestTS08_18_RefineAssessPath(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a session in "init" state (needs assessment).
	sessionData := map[string]any{
		"state":              "init",
		"mode":               "standard",
		"prd_path":           "test.md",
		"assessment_history": []any{},
		"qa_exchanges":       []any{},
		"generated_artifacts": []any{},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a minimal prd.md.
	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD\nContent"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v\nstderr: %s", err, stderrBuf.String())
	}

	output := stdoutBuf.String()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", err, output)
	}

	if ok, _ := parsed["ok"].(bool); !ok {
		t.Errorf("parsed.ok = %v; want true", parsed["ok"])
	}
}

// --- TS-08-19: Verify that spec refine without --answers calls
//     PendingQuestions and emits the result when the session does not
//     need assessment ---

// TestTS08_19_RefinePendingQuestionsPath verifies that spec refine without
// --answers on a session already past assessment (e.g. REFINING state)
// calls PendingQuestions and emits the result as JSON.
// Covers: TS-08-19, Requirement: 08-REQ-8.2
func TestTS08_19_RefinePendingQuestionsPath(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Session in "refining" state with questions in assessment.
	sessionData := map[string]any{
		"state":    "refining",
		"mode":     "standard",
		"prd_path": "test.md",
		"assessment_history": []any{
			map[string]any{
				"quality": "fair",
				"summary": "Needs improvement",
				"gaps":    []string{"gap1"},
				"questions": []any{
					map[string]any{"id": "q1", "text": "What about X?"},
					map[string]any{"id": "q2", "text": "What about Y?"},
				},
			},
		},
		"qa_exchanges":       []any{},
		"generated_artifacts": []any{},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec"})

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
}

// --- TS-08-20: Verify that spec refine with --answers pointing to a JSON
//     file parses answers, runs Refine, and emits refinement result JSON ---

// TestTS08_20_RefineWithAnswersFile verifies that spec refine --answers
// reads answers from a file, passes them to Refine, and emits the result.
// Covers: TS-08-20, Requirement: 08-REQ-8.3
func TestTS08_20_RefineWithAnswersFile(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Session in "refining" state.
	sessionData := map[string]any{
		"state":    "refining",
		"mode":     "standard",
		"prd_path": "test.md",
		"assessment_history": []any{
			map[string]any{
				"quality":   "fair",
				"summary":   "Needs work",
				"gaps":      []string{},
				"questions": []any{},
			},
		},
		"qa_exchanges":       []any{},
		"generated_artifacts": []any{},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create answers file.
	answersPath := filepath.Join(tmpDir, "answers.json")
	if err := os.WriteFile(answersPath, []byte(`{"q1": "answer1"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec", "--answers", answersPath})

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
}

// TestTS08_20_RefineAnswersUnwrapKey verifies that when the answers JSON
// contains an 'answers' key, it is unwrapped before passing to Refine.
// Covers: 08-REQ-8.3
func TestTS08_20_RefineAnswersUnwrapKey(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	sessionData := map[string]any{
		"state":    "refining",
		"mode":     "standard",
		"prd_path": "test.md",
		"assessment_history": []any{
			map[string]any{
				"quality":   "fair",
				"summary":   "Needs work",
				"gaps":      []string{},
				"questions": []any{},
			},
		},
		"qa_exchanges":       []any{},
		"generated_artifacts": []any{},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	// Answers file with wrapped 'answers' key.
	answersPath := filepath.Join(tmpDir, "answers.json")
	if err := os.WriteFile(answersPath, []byte(`{"answers": {"q1": "answer1", "q2": "answer2"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec", "--answers", answersPath})

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
}

// --- TS-08-21: Verify that spec refine with --answers - reads from stdin ---

// TestTS08_21_RefineWithAnswersStdin verifies that spec refine --answers -
// reads answers from stdin, runs Refine, and emits the result.
// Covers: TS-08-21, Requirement: 08-REQ-8.4
func TestTS08_21_RefineWithAnswersStdin(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	sessionData := map[string]any{
		"state":    "refining",
		"mode":     "standard",
		"prd_path": "test.md",
		"assessment_history": []any{
			map[string]any{
				"quality":   "fair",
				"summary":   "Needs work",
				"gaps":      []string{},
				"questions": []any{},
			},
		},
		"qa_exchanges":       []any{},
		"generated_artifacts": []any{},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate stdin with answers JSON.
	stdinContent := `{"q1": "answer1"}`
	stdinReader := bytes.NewReader([]byte(stdinContent))

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetIn(stdinReader)
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec", "--answers", "-"})

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
}

// --- TS-08-22: Verify that spec refine --force resets session state ---

// TestTS08_22_RefineForceResetsSession verifies that --force deletes
// artifact files, resets session state to INIT, and clears assessment
// history and QA exchanges.
// Covers: TS-08-22, Requirement: 08-REQ-8.5
func TestTS08_22_RefineForceResetsSession(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create session in ASSESSING state with assessment history.
	sessionData := map[string]any{
		"state":    "assessing",
		"mode":     "standard",
		"prd_path": "test.md",
		"assessment_history": []any{
			map[string]any{
				"quality":   "good",
				"summary":   "Looks good",
				"gaps":      []string{},
				"questions": []any{},
			},
		},
		"qa_exchanges": []any{
			map[string]any{
				"assessment_index": 0,
				"answers":          map[string]string{"q1": "a1"},
			},
		},
		"generated_artifacts": []string{"requirements.json"},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create artifact files that should be deleted.
	for _, artifact := range []string{"requirements.json", "test_spec.json", "tasks.json"} {
		if err := os.WriteFile(filepath.Join(specPath, artifact), []byte(`{}`), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newRootCmd()
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec", "--force"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Verify artifact files are deleted.
	for _, artifact := range []string{"requirements.json", "test_spec.json", "tasks.json"} {
		artifactPath := filepath.Join(specPath, artifact)
		if _, statErr := os.Stat(artifactPath); !os.IsNotExist(statErr) {
			t.Errorf("artifact %s still exists after --force; want deleted", artifact)
		}
	}

	// Verify session state is reset to INIT.
	sessionPath := filepath.Join(specPath, "_session.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("failed to read _session.json after --force: %v", err)
	}

	var session map[string]any
	if err := json.Unmarshal(data, &session); err != nil {
		t.Fatalf("_session.json is not valid JSON after --force: %v", err)
	}

	if state, _ := session["state"].(string); state != "init" {
		t.Errorf("session.state = %q after --force; want %q", state, "init")
	}

	// Verify assessment history is cleared.
	if ah, ok := session["assessment_history"].([]any); ok && len(ah) != 0 {
		t.Errorf("assessment_history has %d entries after --force; want 0", len(ah))
	}

	// Verify QA exchanges are cleared.
	if qa, ok := session["qa_exchanges"].([]any); ok && len(qa) != 0 {
		t.Errorf("qa_exchanges has %d entries after --force; want 0", len(qa))
	}

	// Verify generated_artifacts is cleared.
	if ga, ok := session["generated_artifacts"].([]any); ok && len(ga) != 0 {
		t.Errorf("generated_artifacts has %d entries after --force; want 0", len(ga))
	}
}

// --- 08-REQ-8.E2: --answers file does not exist ---

// TestTS08_20_RefineAnswersFileNotFound verifies that spec refine returns
// an error when --answers points to a non-existent file.
// Covers: 08-REQ-8.E2
func TestTS08_20_RefineAnswersFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	sessionData := map[string]any{
		"state":              "refining",
		"mode":               "standard",
		"prd_path":           "test.md",
		"assessment_history": []any{},
		"qa_exchanges":       []any{},
		"generated_artifacts": []any{},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec", "--answers", "/nonexistent/answers.json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with non-existent --answers file returned nil; want error")
	}
}

// --- 08-REQ-8.E3: --answers content is not valid JSON ---

// TestTS08_20_RefineAnswersInvalidJSON verifies that spec refine returns
// a parse error when --answers file contains invalid JSON.
// Covers: 08-REQ-8.E3
func TestTS08_20_RefineAnswersInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	specPath := filepath.Join(specDir, "08_my_spec")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		t.Fatal(err)
	}

	sessionData := map[string]any{
		"state":              "refining",
		"mode":               "standard",
		"prd_path":           "test.md",
		"assessment_history": []any{},
		"qa_exchanges":       []any{},
		"generated_artifacts": []any{},
	}
	sessionJSON, _ := json.Marshal(sessionData)
	if err := os.WriteFile(filepath.Join(specPath, "_session.json"), sessionJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(specPath, "prd.md"), []byte("# Test PRD"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create answers file with invalid JSON.
	answersPath := filepath.Join(tmpDir, "bad_answers.json")
	if err := os.WriteFile(answersPath, []byte(`{not valid json!!!`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "08_my_spec", "--answers", answersPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with invalid JSON answers returned nil; want parse error")
	}
}

// --- 08-REQ-8: spec refine requires a spec argument ---

// TestTS08_18_RefineMissingSpecArg verifies that spec refine requires
// a positional spec argument.
// Covers: 08-REQ-8
func TestTS08_18_RefineMissingSpecArg(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"refine"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with no spec argument returned nil; want error")
	}
}

// TestTS08_18_RefineNonexistentSpec verifies that spec refine returns
// an error when the referenced spec does not exist.
// Covers: 08-REQ-8
func TestTS08_18_RefineNonexistentSpec(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--spec-dir", specDir, "refine", "nonexistent_spec"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() with non-existent spec returned nil; want error")
	}
}
