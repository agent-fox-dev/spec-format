package afspec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Spec represents a complete specification package with all artifacts.
// The PRD frontmatter fields are stored directly on the struct, while
// the PRD Markdown body is stored in PRDBody. The three JSON artifacts
// are stored as pointers to the auto-generated types.
type Spec struct {
	// PRD frontmatter fields
	SpecID        string
	SpecName      string
	Title         string
	Status        string
	CreatedAt     string
	UpdatedAt     string
	Owner         string
	Source        string
	Supersedes    []string
	Tags          []string
	IntentHash    *string
	SchemaVersion int

	// PRD Markdown body (everything after the closing --- delimiter)
	PRDBody string

	// JSON artifacts
	Requirements *RequirementsV1Json
	TestSpec     *TestSpecV1Json
	Tasks        *TasksV1Json

	// Optional architecture document
	Architecture string
}

// LoadSpec reads all spec artifacts from a directory and returns a
// populated Spec struct. It reads prd.md, requirements.json,
// test_spec.json, tasks.json, and optionally architecture.md.
func LoadSpec(dir string) (*Spec, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, &LoadError{
			Msg:  fmt.Sprintf("cannot access spec directory: %s", err),
			File: dir,
			Err:  &SpecError{Msg: fmt.Sprintf("cannot access spec directory: %s", err)},
		}
	}
	if !info.IsDir() {
		return nil, &LoadError{
			Msg:  fmt.Sprintf("not a directory: %s", dir),
			File: dir,
			Err:  &SpecError{Msg: fmt.Sprintf("not a directory: %s", dir)},
		}
	}

	// Read and parse prd.md
	prdPath := filepath.Join(dir, "prd.md")
	prdData, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, &LoadError{
			Msg:  fmt.Sprintf("cannot read prd.md: %s", err),
			File: prdPath,
			Err:  &SpecError{Msg: fmt.Sprintf("cannot read prd.md: %s", err)},
		}
	}

	fm, body, err := parsePRD(prdData, prdPath)
	if err != nil {
		return nil, err // parsePRD already returns LoadError
	}

	// Read and parse requirements.json
	reqPath := filepath.Join(dir, "requirements.json")
	var req RequirementsV1Json
	if err := loadJSONArtifact(reqPath, &req); err != nil {
		return nil, err
	}

	// Read and parse test_spec.json
	tsPath := filepath.Join(dir, "test_spec.json")
	var ts TestSpecV1Json
	if err := loadJSONArtifact(tsPath, &ts); err != nil {
		return nil, err
	}

	// Read and parse tasks.json
	tasksPath := filepath.Join(dir, "tasks.json")
	var tasks TasksV1Json
	if err := loadJSONArtifact(tasksPath, &tasks); err != nil {
		return nil, err
	}

	// Read architecture.md (optional)
	var architecture string
	archPath := filepath.Join(dir, "architecture.md")
	archData, err := os.ReadFile(archPath)
	if err == nil {
		architecture = string(archData)
	}
	// If architecture.md doesn't exist, that's fine — leave empty

	spec := &Spec{
		SpecID:        fm.SpecID,
		SpecName:      fm.SpecName,
		Title:         fm.Title,
		Status:        fm.Status,
		CreatedAt:     fm.CreatedAt,
		UpdatedAt:     fm.UpdatedAt,
		Owner:         fm.Owner,
		Source:        fm.Source,
		Supersedes:    fm.Supersedes,
		Tags:          fm.Tags,
		IntentHash:    fm.IntentHash,
		SchemaVersion: fm.SchemaVersion,
		PRDBody:       body,
		Requirements:  &req,
		TestSpec:      &ts,
		Tasks:         &tasks,
		Architecture:  architecture,
	}

	return spec, nil
}

// loadJSONArtifact reads and unmarshals a JSON artifact file.
// Returns a LoadError wrapping SpecError on failure.
func loadJSONArtifact(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &LoadError{
			Msg:  fmt.Sprintf("cannot read %s: %s", filepath.Base(path), err),
			File: path,
			Err:  &SpecError{Msg: fmt.Sprintf("cannot read %s: %s", filepath.Base(path), err)},
		}
	}

	if err := json.Unmarshal(data, target); err != nil {
		return &LoadError{
			Msg:  fmt.Sprintf("cannot parse %s: %s", filepath.Base(path), err),
			File: path,
			Err:  &SpecError{Msg: fmt.Sprintf("cannot parse %s: %s", filepath.Base(path), err)},
		}
	}

	return nil
}

// Save atomically writes all spec artifacts to the given directory.
// Each artifact is written to a temporary file first, then renamed
// to its final name. Returns a LifecycleError if the spec is sealed,
// superseded, or archived.
func (s *Spec) Save(dir string) error {
	if s.Status == "sealed" || s.Status == "superseded" || s.Status == "archived" {
		return &LifecycleError{
			Msg: fmt.Sprintf("cannot save spec in %s state", s.Status),
			Err: &SpecError{Msg: fmt.Sprintf("cannot save spec in %s state", s.Status)},
		}
	}

	// Verify directory exists
	if _, err := os.Stat(dir); err != nil {
		return &SaveError{
			Msg: fmt.Sprintf("cannot access directory: %s", err),
			Err: &SpecError{Msg: fmt.Sprintf("cannot access directory: %s", err)},
		}
	}

	// Render PRD
	prdContent := renderPRD(s)

	// Marshal JSON artifacts
	reqData, err := MarshalJSON(s.Requirements)
	if err != nil {
		return &SaveError{
			Msg: fmt.Sprintf("cannot marshal requirements.json: %s", err),
			Err: &SpecError{Msg: fmt.Sprintf("cannot marshal requirements.json: %s", err)},
		}
	}

	tsData, err := MarshalJSON(s.TestSpec)
	if err != nil {
		return &SaveError{
			Msg: fmt.Sprintf("cannot marshal test_spec.json: %s", err),
			Err: &SpecError{Msg: fmt.Sprintf("cannot marshal test_spec.json: %s", err)},
		}
	}

	tasksData, err := MarshalJSON(s.Tasks)
	if err != nil {
		return &SaveError{
			Msg: fmt.Sprintf("cannot marshal tasks.json: %s", err),
			Err: &SpecError{Msg: fmt.Sprintf("cannot marshal tasks.json: %s", err)},
		}
	}

	// Write all artifacts atomically (write to temp, then rename)
	artifacts := []struct {
		name string
		data []byte
	}{
		{"prd.md", prdContent},
		{"requirements.json", reqData},
		{"test_spec.json", tsData},
		{"tasks.json", tasksData},
	}

	// Also write architecture.md if present
	if s.Architecture != "" {
		artifacts = append(artifacts, struct {
			name string
			data []byte
		}{"architecture.md", []byte(s.Architecture)})
	}

	for _, artifact := range artifacts {
		if err := writeAtomic(dir, artifact.name, artifact.data); err != nil {
			return &SaveError{
				Msg: fmt.Sprintf("cannot write %s: %s", artifact.name, err),
				Err: &SpecError{Msg: fmt.Sprintf("cannot write %s: %s", artifact.name, err)},
			}
		}
	}

	return nil
}

// writeAtomic writes data to a temporary file in dir, then renames it to name.
func writeAtomic(dir, name string, data []byte) error {
	finalPath := filepath.Join(dir, name)

	// Create temp file in the same directory for atomic rename
	tmpFile, err := os.CreateTemp(dir, name+".tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return err
	}

	success = true
	return nil
}
