package afspec

// ValidationError represents a single validation error from BootstrapSpec.Finalize.
// Rule identifies the type of check that failed (e.g. "bootstrap" for missing
// artifacts, or a schema/cross-file rule name). Message describes the failure.
type ValidationError struct {
	Rule    string
	Message string
}

// BootstrapSpec supports incremental population of spec artifacts with
// deferred cross-file validation. Use NewBootstrapSpec to create one, set
// artifacts on it, then call Finalize to assemble and validate the Spec.
type BootstrapSpec struct {
	SpecID       string
	SpecName     string
	Requirements *RequirementsV1Json
	TestSpec     *TestSpecV1Json
	Tasks        *TasksV1Json
	PRDBody      string
	Architecture string
}

// NewBootstrapSpec creates a BootstrapSpec with the given specID and specName.
// All artifact fields start as nil/zero values.
func NewBootstrapSpec(specID, specName string) *BootstrapSpec {
	panic("not implemented")
}

// Finalize checks that all required artifacts (Requirements, TestSpec, Tasks,
// PRDBody) are set, assembles a Spec from them, and runs full schema and
// cross-file validation. Returns (*Spec, nil) on success.
//
// If any required artifacts are missing, returns (nil, []ValidationError)
// where each missing artifact produces a ValidationError with Rule="bootstrap"
// and Message="Missing artifact: {name}".
//
// If all artifacts are present but validation fails, returns
// (nil, []ValidationError) with validation error details.
//
// The BootstrapSpec is not mutated by Finalize; calling it multiple times
// produces independent results.
func (b *BootstrapSpec) Finalize() (*Spec, []ValidationError) {
	panic("not implemented")
}
