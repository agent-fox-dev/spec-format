package afspec

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
	panic("not implemented")
}

// Save atomically writes all spec artifacts to the given directory.
// Each artifact is written to a temporary file first, then renamed
// to its final name. Returns a LifecycleError if the spec is sealed.
func (s *Spec) Save(dir string) error {
	panic("not implemented")
}

// MarshalJSON serializes a spec artifact struct to deterministic JSON.
// Struct fields are serialized in declaration order (matching JSON
// Schema property order) and all map[string]T keys are sorted
// alphabetically. The output is byte-for-byte identical to the Python
// library's JSON output.
func MarshalJSON(v interface{}) ([]byte, error) {
	panic("not implemented")
}
