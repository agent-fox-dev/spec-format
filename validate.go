package afspec

// ValidationResult holds the results of spec validation, including errors
// and warnings. Valid is true if and only if Errors is empty.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationEntry
	Warnings []ValidationEntry
}

// ValidationEntry represents a single validation error or warning.
type ValidationEntry struct {
	Category      string // "schema" or "integrity" or "warning"
	Message       string
	Artifact      string // which artifact file had the issue
	Path          string // JSON path (InstanceLocation) for schema errors
	Keyword       string // KeywordLocation for schema errors
	Check         string // check name for integrity errors (e.g. "dangling_reference")
	RequirementID string // for integrity errors referencing a requirement
	EntityID      string // for warnings referencing an entity
	Value         string // optional value context
}

// Validate runs schema validation, EARS constraint checks, task group
// structure rules, and cross-file consistency checks, then returns a
// ValidationResult. Valid is true if and only if Errors is empty.
func (s *Spec) Validate() ValidationResult {
	panic("not implemented")
}

// ValidateSchema compiles each bundled schema using jsonschema compiler
// with a SchemeURLLoader backed by embedded bytes, validates each
// artifact's JSON representation against its schema, and returns a
// ValidationResult.
func (s *Spec) ValidateSchema() ValidationResult {
	panic("not implemented")
}

// ValidateCrossFile checks dangling references, coverage gaps, glossary
// completeness, and ID format validity across all artifacts and returns
// a ValidationResult.
func (s *Spec) ValidateCrossFile() ValidationResult {
	panic("not implemented")
}

// ValidateCrossSpec checks API symbol consistency, glossary conflicts,
// and contract mismatches across all provided specs using the dependency
// graph to determine which specs interact, and returns a ValidationResult.
func ValidateCrossSpec(specs []*Spec, graph *DependencyGraph) ValidationResult {
	panic("not implemented")
}

// ValidateStructured runs full validation and returns a structured
// map[string]any for CLI consumption, with 'valid', 'errors', and
// optionally 'warnings' keys.
func (s *Spec) ValidateStructured() map[string]any {
	panic("not implemented")
}

// DependencyGraph represents the inter-spec dependency graph built from
// tasks.json declarations. It is used by ValidateCrossSpec to determine
// which specs interact.
type DependencyGraph struct {
	Edges []DependencyEdge
}

// DependencyEdge represents a directed dependency edge between two specs.
type DependencyEdge struct {
	FromSpec     string
	ToSpec       string
	FromGroup    int
	ToGroup      int
	Relationship string
}
