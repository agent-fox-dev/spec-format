package afspec

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestSave_RoundTrip verifies that Spec.Save writes all artifacts atomically
// and produces byte-for-byte identical output to the original fixture files.
// Test Spec: TS-01-4, Requirement: 01-REQ-2.1
func TestSave_RoundTrip(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
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
			expected, err := os.ReadFile(filepath.Join("./../testdata/valid_spec", name))
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

	// Verify no temp files remain (temp files are named "<artifact>.tmp.<random>")
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read tmpDir: %v", err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp.") {
			t.Errorf("temp file %q was not cleaned up after successful save", entry.Name())
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

// TestSave_TwoPhase_WriteFailure verifies that when temp-file creation fails
// (write phase), no on-disk artifact is modified and no temp files remain
// (NS-REQ-2, NS-REQ-3).
//
// The test makes the target directory read-only so that os.CreateTemp fails
// immediately, exercising the "write phase fails → no renames performed"
// invariant.
func TestSave_TwoPhase_WriteFailure(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test read-only directory restrictions as root")
	}

	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec failed: %v", err)
	}

	fixtureDir := "./../testdata/valid_spec"
	artifactNames := []string{"prd.md", "requirements.json", "test_spec.json", "tasks.json"}

	// Populate a writable temp dir with the fixture files and record originals.
	tmpDir := t.TempDir()
	origContents := make(map[string][]byte, len(artifactNames))
	for _, name := range artifactNames {
		data, err := os.ReadFile(filepath.Join(fixtureDir, name))
		if err != nil {
			t.Fatalf("failed to read fixture %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), data, 0644); err != nil {
			t.Fatalf("failed to copy fixture %s: %v", name, err)
		}
		origContents[name] = data
	}

	// Make the directory read-only: os.CreateTemp will fail, so the write
	// phase cannot start and no rename will occur.
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Fatalf("chmod 0555 failed: %v", err)
	}
	// Restore write permission so t.TempDir cleanup can remove the directory.
	t.Cleanup(func() { os.Chmod(tmpDir, 0755) }) //nolint:errcheck

	saveErr := spec.saveToDisk(tmpDir)
	if saveErr == nil {
		t.Fatal("expected saveToDisk to return an error for a read-only directory, got nil")
	}

	var se *SaveError
	if !errors.As(saveErr, &se) {
		t.Errorf("expected *SaveError, got %T: %v", saveErr, saveErr)
	}

	// Restore permissions for inspection.
	if err := os.Chmod(tmpDir, 0755); err != nil {
		t.Fatalf("failed to restore dir permissions: %v", err)
	}

	// NS-REQ-3: no orphaned temp files.
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp.") {
			t.Errorf("orphaned temp file %q found after write failure", entry.Name())
		}
	}

	// NS-REQ-2: every original artifact file must be byte-for-byte unchanged.
	for name, original := range origContents {
		actual, err := os.ReadFile(filepath.Join(tmpDir, name))
		if err != nil {
			t.Fatalf("cannot read %s after failed save: %v", name, err)
		}
		if !bytes.Equal(original, actual) {
			t.Errorf("artifact %s was modified by a failed save", name)
		}
	}
}

// TestSave_TwoPhase_NoOrphanedTempsOnSuccess confirms that no *.tmp.* files
// remain after a fully successful two-phase save (NS-REQ-3 success path).
// This supplements TestSave_RoundTrip with an explicit naming-pattern check.
func TestSave_TwoPhase_NoOrphanedTempsOnSuccess(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec failed: %v", err)
	}

	tmpDir := t.TempDir()
	if err := spec.Save(tmpDir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp.") {
			t.Errorf("orphaned temp file %q found after successful save", entry.Name())
		}
	}
}
