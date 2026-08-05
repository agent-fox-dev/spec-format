package afspec

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestSave_RoundTrip verifies that Spec.Save writes all artifacts atomically
// and produces byte-for-byte identical output to the original fixture files.
// Test Spec: TS-01-4, Requirement: 01-REQ-2.1
func TestSave_RoundTrip(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec failed: %v", err)
	}

	tmpDir := t.TempDir()

	if err := spec.Save(tmpDir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check all artifact files are byte-for-byte identical
	artifactFiles := []string{
		"prd.md",
		"requirements.json",
		"test_spec.json",
		"tasks.json",
	}

	for _, name := range artifactFiles {
		t.Run(name, func(t *testing.T) {
			expected, err := os.ReadFile(filepath.Join("testdata/valid_spec", name))
			if err != nil {
				t.Fatalf("failed to read expected file: %v", err)
			}

			actual, err := os.ReadFile(filepath.Join(tmpDir, name))
			if err != nil {
				t.Fatalf("failed to read actual file: %v", err)
			}

			if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
				t.Errorf("byte mismatch for %s (-want +got):\n%s", name, diff)
			}
		})
	}

	// Verify no temp files remain
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read tmpDir: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if filepath.Ext(name) == ".tmp" {
			t.Errorf("temp file %s was not cleaned up", name)
		}
	}
}

// TestSave_SealedSpec verifies that Spec.Save on a sealed Spec returns
// a LifecycleError without writing any files.
// Test Spec: TS-01-5, Requirement: 01-REQ-2.2
func TestSave_SealedSpec(t *testing.T) {
	defer requireImplemented(t)

	spec := &Spec{
		SpecID:        "01",
		SpecName:      "test_feature",
		Title:         "Test Feature",
		Status:        "sealed",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://github.com/test/repo/issues/1",
		SchemaVersion: 1,
		PRDBody:       "# Test Feature\n",
	}

	tmpDir := t.TempDir()

	err := spec.Save(tmpDir)
	if err == nil {
		t.Fatal("expected error when saving sealed spec, got nil")
	}

	var lifecycleErr *LifecycleError
	if !errors.As(err, &lifecycleErr) {
		t.Errorf("expected errors.As(err, &LifecycleError{}) to return true, got false; err type = %T", err)
	}

	var specErr *SpecError
	if !errors.As(err, &specErr) {
		t.Errorf("expected errors.As(err, &SpecError{}) to return true, got false; err type = %T", err)
	}

	// Verify no files were created
	entries, readErr := os.ReadDir(tmpDir)
	if readErr != nil {
		t.Fatalf("failed to read tmpDir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("expected no files in tmpDir for sealed spec, got %d files", len(entries))
	}
}

// TestSave_NonexistentDir verifies that Spec.Save returns a SaveError
// when the target directory does not exist.
// Requirement: 01-REQ-2.E3
func TestSave_NonexistentDir(t *testing.T) {
	defer requireImplemented(t)

	spec := &Spec{
		SpecID:        "01",
		SpecName:      "test_feature",
		Title:         "Test Feature",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://github.com/test/repo/issues/1",
		SchemaVersion: 1,
		PRDBody:       "# Test Feature\n",
	}

	err := spec.Save("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}

	var saveErr *SaveError
	if !errors.As(err, &saveErr) {
		t.Errorf("expected errors.As(err, &SaveError{}) to return true, got false; err type = %T", err)
	}

	var specErr *SpecError
	if !errors.As(err, &specErr) {
		t.Errorf("expected errors.As(err, &SpecError{}) to return true, got false; err type = %T", err)
	}
}

// TestSave_ErrorTypes verifies that save failures produce the correct
// error types wrapping SpecError.
// Requirement: 01-REQ-26.1, 01-REQ-26.2
func TestSave_ErrorTypes(t *testing.T) {
	tests := []struct {
		name   string
		spec   *Spec
		dir    string
		errTyp string // "LifecycleError" or "SaveError"
	}{
		{
			name: "sealed spec",
			spec: &Spec{
				Status:        "sealed",
				SpecID:        "01",
				SpecName:      "test",
				SchemaVersion: 1,
			},
			dir:    t.TempDir(),
			errTyp: "LifecycleError",
		},
		{
			name: "nonexistent directory",
			spec: &Spec{
				Status:        "draft",
				SpecID:        "01",
				SpecName:      "test",
				SchemaVersion: 1,
			},
			dir:    "/nonexistent/dir",
			errTyp: "SaveError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer requireImplemented(t)

			err := tt.spec.Save(tt.dir)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// All save errors should be unwrappable to SpecError
			var specErr *SpecError
			if !errors.As(err, &specErr) {
				t.Errorf("errors.As(err, &SpecError{}) = false, want true")
			}

			switch tt.errTyp {
			case "LifecycleError":
				var le *LifecycleError
				if !errors.As(err, &le) {
					t.Errorf("errors.As(err, &LifecycleError{}) = false, want true")
				}
			case "SaveError":
				var se *SaveError
				if !errors.As(err, &se) {
					t.Errorf("errors.As(err, &SaveError{}) = false, want true")
				}
			}
		})
	}
}
