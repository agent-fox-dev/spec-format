package afspec

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Subtask 2.1: Spec.Validate, ValidateSchema, ValidateCrossFile
// ---------------------------------------------------------------------------

// TestValidate_FullyValidSpec verifies that Spec.Validate on a fully valid
// spec returns a ValidationResult with Valid true and empty Errors slice.
// Test Spec: TS-01-8, Requirement: 01-REQ-4.1
func TestValidate_FullyValidSpec(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec returned unexpected error: %v", err)
	}

	result := spec.Validate()
	if !result.Valid {
		t.Errorf("Validate().Valid = false, want true; errors = %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Validate().Errors has %d entries, want 0: %v", len(result.Errors), result.Errors)
	}
}

// TestValidate_SpecWithErrors verifies that Spec.Validate on a spec with
// one or more errors returns a ValidationResult with Valid false and a
// non-empty Errors slice.
// Test Spec: TS-01-9, Requirement: 01-REQ-4.2
func TestValidate_SpecWithErrors(t *testing.T) {
	defer requireImplemented(t)

	// Build a spec with a schema violation: a requirement missing required fields
	spec := buildInvalidSpec()

	result := spec.Validate()
	if result.Valid {
		t.Error("Validate().Valid = true, want false for invalid spec")
	}
	if len(result.Errors) == 0 {
		t.Error("Validate().Errors is empty, want at least one error for invalid spec")
	}
}

// TestValidate_WarningsButNoErrors verifies that Validate on a spec with
// warnings but no errors returns Valid true with non-empty Warnings.
// Requirement: 01-REQ-4.E1
func TestValidate_WarningsButNoErrors(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSpecWithWarnings()

	result := spec.Validate()
	if !result.Valid {
		t.Errorf("Validate().Valid = false, want true (warnings only); errors = %v", result.Errors)
	}
	if len(result.Warnings) == 0 {
		t.Error("Validate().Warnings is empty, want at least one warning")
	}
}

// TestValidate_MinimalSpec verifies that Validate on a minimally populated
// spec does not panic.
// Requirement: 01-REQ-4.E2
func TestValidate_MinimalSpec(t *testing.T) {
	defer requireImplemented(t)

	spec := &Spec{
		SpecID:        "01",
		SpecName:      "minimal",
		Title:         "Minimal Spec",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Minimal\n",
		Requirements: &RequirementsV1Json{
			SpecId:                "01",
			SpecName:              "minimal",
			SchemaVersion:         1,
			Introduction:          "Minimal spec.",
			Glossary:              RequirementsV1JsonGlossary{},
			Requirements:          []Requirement{},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "01",
			SpecName:      "minimal",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "01",
			SpecName:      "minimal",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}

	// Should not panic — just produce a result
	result := spec.Validate()
	_ = result // Don't assert valid/invalid — just ensure no panic
}

// TestValidate_ValidFieldConsistentWithErrors verifies the correctness
// property: Valid is true iff Errors is empty.
// Property: 01-PROP-4
func TestValidate_ValidFieldConsistentWithErrors(t *testing.T) {
	defer requireImplemented(t)

	// Valid spec: Valid should be true, Errors empty
	validSpec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec returned unexpected error: %v", err)
	}
	result := validSpec.Validate()
	if result.Valid && len(result.Errors) != 0 {
		t.Error("Valid is true but Errors is non-empty — violates PROP-4")
	}
	if !result.Valid && len(result.Errors) == 0 {
		t.Error("Valid is false but Errors is empty — violates PROP-4")
	}

	// Invalid spec: Valid should be false, Errors non-empty
	invalidSpec := buildInvalidSpec()
	result2 := invalidSpec.Validate()
	if result2.Valid && len(result2.Errors) != 0 {
		t.Error("Valid is true but Errors is non-empty — violates PROP-4")
	}
	if !result2.Valid && len(result2.Errors) == 0 {
		t.Error("Valid is false but Errors is empty — violates PROP-4")
	}
}

// TestValidateSchema_ValidSpec verifies that ValidateSchema on a structurally
// valid spec returns ValidationResult with Valid true and no errors.
// Test Spec: TS-01-10, Requirement: 01-REQ-5.1
func TestValidateSchema_ValidSpec(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec returned unexpected error: %v", err)
	}

	result := spec.ValidateSchema()
	if !result.Valid {
		t.Errorf("ValidateSchema().Valid = false, want true; errors = %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("ValidateSchema().Errors has %d entries, want 0", len(result.Errors))
	}
}

// TestValidateSchema_SchemaViolation verifies that ValidateSchema maps
// each *jsonschema.ValidationError to a ValidationResult error entry
// with path, keyword/message fields populated.
// Test Spec: TS-01-11, Requirement: 01-REQ-5.2
func TestValidateSchema_SchemaViolation(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSchemaViolatingSpec()

	result := spec.ValidateSchema()
	if result.Valid {
		t.Error("ValidateSchema().Valid = true, want false for schema-violating spec")
	}
	if len(result.Errors) == 0 {
		t.Fatal("ValidateSchema().Errors is empty, want at least one error")
	}

	// First error should have path and message populated
	e := result.Errors[0]
	if e.Path == "" && e.Message == "" {
		t.Error("expected error entry to have non-empty Path or Message")
	}
	if e.Message == "" {
		t.Error("expected error entry to have non-empty Message")
	}
}

// TestValidateSchema_MultipleArtifactErrors verifies that ValidateSchema
// collects errors from all artifacts rather than stopping at the first failure.
// Requirement: 01-REQ-5.E2
func TestValidateSchema_MultipleArtifactErrors(t *testing.T) {
	defer requireImplemented(t)

	// Build a spec with schema violations in multiple artifacts
	spec := buildMultiArtifactSchemaViolatingSpec()

	result := spec.ValidateSchema()
	if result.Valid {
		t.Error("ValidateSchema().Valid = true, want false")
	}
	// Should have errors from more than one artifact
	artifacts := map[string]bool{}
	for _, e := range result.Errors {
		if e.Artifact != "" {
			artifacts[e.Artifact] = true
		}
	}
	if len(artifacts) < 2 {
		t.Errorf("expected errors from at least 2 artifacts, got errors from %d: %v",
			len(artifacts), artifacts)
	}
}

// TestValidateCrossFile_ConsistentReferences verifies that ValidateCrossFile
// on a spec with consistent cross-file references returns Valid true.
// Test Spec: TS-01-12, Requirement: 01-REQ-6.1
func TestValidateCrossFile_ConsistentReferences(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec returned unexpected error: %v", err)
	}

	result := spec.ValidateCrossFile()
	if !result.Valid {
		t.Errorf("ValidateCrossFile().Valid = false, want true; errors = %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("ValidateCrossFile().Errors has %d entries, want 0", len(result.Errors))
	}
}

// TestValidateCrossFile_DanglingReference verifies that ValidateCrossFile
// records a dangling reference error when a test entry references a
// requirement ID that does not exist.
// Test Spec: TS-01-13, Requirement: 01-REQ-6.2
func TestValidateCrossFile_DanglingReference(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSpecWithDanglingRef("XX-REQ-999")

	result := spec.ValidateCrossFile()
	if result.Valid {
		t.Error("ValidateCrossFile().Valid = true, want false for dangling reference")
	}

	// Find the dangling_reference error
	var found bool
	for _, e := range result.Errors {
		if e.Check == "dangling_reference" {
			found = true
			if e.Category != "integrity" {
				t.Errorf("dangling_reference error category = %q, want %q", e.Category, "integrity")
			}
			if !strings.Contains(e.Message, "XX-REQ-999") {
				t.Errorf("dangling_reference error message %q does not contain %q", e.Message, "XX-REQ-999")
			}
			break
		}
	}
	if !found {
		t.Error("expected a dangling_reference error in ValidationResult.Errors")
	}
}

// TestValidateCrossFile_NoTestEntries verifies that ValidateCrossFile reports
// coverage gaps for requirements with no test entries.
// Requirement: 01-REQ-6.E1
func TestValidateCrossFile_NoTestEntries(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSpecWithNoTestEntries()

	result := spec.ValidateCrossFile()

	// Should report coverage gaps (as warnings or errors) for uncovered requirements
	hasGap := false
	for _, e := range result.Errors {
		if strings.Contains(e.Check, "coverage") || strings.Contains(e.Message, "coverage") {
			hasGap = true
			break
		}
	}
	for _, w := range result.Warnings {
		if strings.Contains(w.Check, "coverage") || strings.Contains(w.Message, "coverage") {
			hasGap = true
			break
		}
	}
	if !hasGap {
		t.Error("expected coverage gap warning or error for requirements with no test entries")
	}
}

// TestValidateCrossFile_IDFormatViolation verifies that ValidateCrossFile
// records an ID format validation error for malformed IDs.
// Requirement: 01-REQ-6.E2
func TestValidateCrossFile_IDFormatViolation(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSpecWithBadID()

	result := spec.ValidateCrossFile()
	if result.Valid {
		t.Error("ValidateCrossFile().Valid = true, want false for ID format violation")
	}

	var found bool
	for _, e := range result.Errors {
		if e.Check == "id_format" {
			found = true
			if e.Category != "integrity" {
				t.Errorf("id_format error category = %q, want %q", e.Category, "integrity")
			}
			break
		}
	}
	if !found {
		t.Error("expected an id_format error in ValidationResult.Errors")
	}
}

// ---------------------------------------------------------------------------
// Subtask 2.2: ValidateCrossSpec and ValidateStructured
// ---------------------------------------------------------------------------

// TestValidateCrossSpec_CompatibleSpecs verifies that ValidateCrossSpec
// with compatible specs returns Valid true.
// Test Spec: TS-01-14, Requirement: 01-REQ-7.1
func TestValidateCrossSpec_CompatibleSpecs(t *testing.T) {
	defer requireImplemented(t)

	specA := buildSpecA()
	specB := buildSpecB()
	graph := &DependencyGraph{
		Edges: []DependencyEdge{
			{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
		},
	}

	result := ValidateCrossSpec([]*Spec{specA, specB}, graph)
	if !result.Valid {
		t.Errorf("ValidateCrossSpec().Valid = false, want true; errors = %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("ValidateCrossSpec().Errors has %d entries, want 0", len(result.Errors))
	}
}

// TestValidateCrossSpec_SingleSpec verifies that ValidateCrossSpec with a
// single spec (no dependencies) returns Valid true.
// Requirement: 01-REQ-7.E1
func TestValidateCrossSpec_SingleSpec(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSpecA()
	graph := &DependencyGraph{Edges: []DependencyEdge{}}

	result := ValidateCrossSpec([]*Spec{spec}, graph)
	if !result.Valid {
		t.Errorf("ValidateCrossSpec().Valid = false, want true for single spec; errors = %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("ValidateCrossSpec().Errors has %d entries, want 0", len(result.Errors))
	}
}

// TestValidateCrossSpec_GlossaryConflict verifies that ValidateCrossSpec
// records a glossary conflict error when two specs define the same term
// with different meanings.
// Requirement: 01-REQ-7.E2
func TestValidateCrossSpec_GlossaryConflict(t *testing.T) {
	defer requireImplemented(t)

	specA := buildSpecA()
	specB := buildSpecWithConflictingGlossary()
	graph := &DependencyGraph{
		Edges: []DependencyEdge{
			{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
		},
	}

	result := ValidateCrossSpec([]*Spec{specA, specB}, graph)
	if result.Valid {
		t.Error("ValidateCrossSpec().Valid = true, want false for glossary conflict")
	}

	var found bool
	for _, e := range result.Errors {
		if e.Check == "glossary_conflict" {
			found = true
			if e.Category != "integrity" {
				t.Errorf("glossary_conflict error category = %q, want %q", e.Category, "integrity")
			}
			break
		}
	}
	if !found {
		t.Error("expected a glossary_conflict error in ValidationResult.Errors")
	}
}

// TestValidateStructured_ValidSpec verifies that ValidateStructured on a
// valid spec returns a map with 'valid' true, 'errors' as empty slice,
// and no 'warnings' key.
// Test Spec: TS-01-15, Requirement: 01-REQ-8.1
func TestValidateStructured_ValidSpec(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec returned unexpected error: %v", err)
	}

	result := spec.ValidateStructured()

	valid, ok := result["valid"].(bool)
	if !ok {
		t.Fatal("result['valid'] is not a bool")
	}
	if !valid {
		t.Error("result['valid'] = false, want true")
	}

	errs, ok := result["errors"].([]map[string]any)
	if !ok {
		t.Fatalf("result['errors'] is not []map[string]any, got %T", result["errors"])
	}
	if len(errs) != 0 {
		t.Errorf("result['errors'] has %d entries, want 0", len(errs))
	}

	// No warnings key should be present for a fully valid spec with no warnings
	if _, hasWarnings := result["warnings"]; hasWarnings {
		t.Error("result has 'warnings' key, want it absent for fully valid spec")
	}
}

// TestValidateStructured_SchemaErrors verifies that ValidateStructured on
// a spec with schema errors returns a map with 'valid' false and 'errors'
// entries with 'category', 'message', and 'artifact' fields.
// Test Spec: TS-01-16, Requirement: 01-REQ-8.2
func TestValidateStructured_SchemaErrors(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSchemaViolatingSpec()

	result := spec.ValidateStructured()

	valid, ok := result["valid"].(bool)
	if !ok {
		t.Fatal("result['valid'] is not a bool")
	}
	if valid {
		t.Error("result['valid'] = true, want false for schema-violating spec")
	}

	errs, ok := result["errors"].([]map[string]any)
	if !ok {
		t.Fatalf("result['errors'] is not []map[string]any, got %T", result["errors"])
	}
	if len(errs) == 0 {
		t.Fatal("result['errors'] is empty, want at least one entry")
	}

	e := errs[0]
	if cat, _ := e["category"].(string); cat != "schema" {
		t.Errorf("error category = %q, want %q", cat, "schema")
	}
	if msg, _ := e["message"].(string); msg == "" {
		t.Error("error message is empty, want non-empty")
	}
	if art, _ := e["artifact"].(string); art == "" {
		t.Error("error artifact is empty, want non-empty")
	}
}

// TestValidateStructured_IntegrityErrors verifies that ValidateStructured
// on a spec with integrity errors returns entries with 'category'=='integrity',
// 'message', and 'check' fields.
// Test Spec: TS-01-17, Requirement: 01-REQ-8.3
func TestValidateStructured_IntegrityErrors(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSpecWithDanglingRef("XX-REQ-999")

	result := spec.ValidateStructured()

	valid, ok := result["valid"].(bool)
	if !ok {
		t.Fatal("result['valid'] is not a bool")
	}
	if valid {
		t.Error("result['valid'] = true, want false for integrity error")
	}

	errs, ok := result["errors"].([]map[string]any)
	if !ok {
		t.Fatalf("result['errors'] is not []map[string]any, got %T", result["errors"])
	}

	var found bool
	for _, e := range errs {
		if cat, _ := e["category"].(string); cat == "integrity" {
			found = true
			if chk, _ := e["check"].(string); chk == "" {
				t.Error("integrity error check is empty, want non-empty")
			}
			if msg, _ := e["message"].(string); msg == "" {
				t.Error("integrity error message is empty, want non-empty")
			}
			break
		}
	}
	if !found {
		t.Error("expected at least one integrity error in result['errors']")
	}
}

// TestValidateStructured_WithWarnings verifies that ValidateStructured on
// a spec with warnings includes a 'warnings' key with entries containing
// 'category'=='warning', 'message', and 'entity_id' fields.
// Test Spec: TS-01-18, Requirement: 01-REQ-8.4
func TestValidateStructured_WithWarnings(t *testing.T) {
	defer requireImplemented(t)

	spec := buildSpecWithWarnings()

	result := spec.ValidateStructured()

	valid, ok := result["valid"].(bool)
	if !ok {
		t.Fatal("result['valid'] is not a bool")
	}
	if !valid {
		t.Errorf("result['valid'] = false, want true (warnings only); errors = %v", result["errors"])
	}

	warnings, ok := result["warnings"].([]map[string]any)
	if !ok {
		t.Fatalf("result['warnings'] is not []map[string]any, got %T", result["warnings"])
	}
	if len(warnings) == 0 {
		t.Fatal("result['warnings'] is empty, want at least one entry")
	}

	w := warnings[0]
	if cat, _ := w["category"].(string); cat != "warning" {
		t.Errorf("warning category = %q, want %q", cat, "warning")
	}
	if msg, _ := w["message"].(string); msg == "" {
		t.Error("warning message is empty, want non-empty")
	}
	if eid, _ := w["entity_id"].(string); eid == "" {
		t.Error("warning entity_id is empty, want non-empty")
	}
}

// TestValidateStructured_NoWarningsKey verifies that the 'warnings' key
// is omitted when there are no warnings.
// Requirement: 01-REQ-8.E1
func TestValidateStructured_NoWarningsKey(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec returned unexpected error: %v", err)
	}

	result := spec.ValidateStructured()

	if _, hasWarnings := result["warnings"]; hasWarnings {
		t.Error("result has 'warnings' key, want it absent when there are no warnings")
	}
}

// ---------------------------------------------------------------------------
// Test helpers for building test specs
// ---------------------------------------------------------------------------

// buildInvalidSpec creates a Spec with a schema violation: requirement
// with missing required fields.
func buildInvalidSpec() *Spec {
	return &Spec{
		SpecID:        "01",
		SpecName:      "test_feature",
		Title:         "Test Feature",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			Introduction:  "Test.",
			Glossary:      RequirementsV1JsonGlossary{},
			Requirements: []Requirement{
				{
					// Missing ID and Title — schema violation
					Id:    "", // empty ID violates minLength
					Title: "", // empty title violates minLength
					UserStory: UserStory{
						Role:    "dev",
						Goal:    "test",
						Benefit: "testing",
					},
					AcceptanceCriteria: []Criterion{},
					EdgeCases:          []Criterion{},
				},
			},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}
}

// buildSchemaViolatingSpec creates a Spec where requirements.json has a
// schema violation (requirement with empty ID).
func buildSchemaViolatingSpec() *Spec {
	return buildInvalidSpec() // Same as buildInvalidSpec — empty ID violates schema
}

// buildMultiArtifactSchemaViolatingSpec creates a Spec with schema
// violations in multiple artifacts.
func buildMultiArtifactSchemaViolatingSpec() *Spec {
	spec := buildInvalidSpec()
	// Also make test_spec have violations
	spec.TestSpec = &TestSpecV1Json{
		SpecId:        "01",
		SpecName:      "test_feature",
		SchemaVersion: 1,
		TestCases: []TestCase{
			{
				Id:                  "", // empty — schema violation
				RequirementId:       "",
				Kind:                "unit",
				Description:         "Bad test case",
				Preconditions:       []string{},
				Expected:            "something",
				AssertionPseudocode: "assert true",
			},
		},
		PropertyTests: []PropertyTest{},
		EdgeCaseTests: []EdgeCaseTest{},
		SmokeTests:    []SmokeTest{},
		Coverage:      Coverage{},
	}
	return spec
}

// buildSpecWithDanglingRef creates a Spec where a test case references
// a requirement ID that does not exist in requirements.json.
func buildSpecWithDanglingRef(danglingID string) *Spec {
	return &Spec{
		SpecID:        "01",
		SpecName:      "test_feature",
		Title:         "Test Feature",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			Introduction:  "Test.",
			Glossary:      RequirementsV1JsonGlossary{"spec": "A specification."},
			Requirements: []Requirement{
				{
					Id:    "01-REQ-1",
					Title: "Data Model",
					UserStory: UserStory{
						Role:    "developer",
						Goal:    "have typed models",
						Benefit: "type safety",
					},
					AcceptanceCriteria: []Criterion{
						{
							Id:             "01-REQ-1.1",
							EarsPattern:    CriterionEarsPatternUbiquitous,
							System:         "the system",
							Action:         "return a populated Spec",
							ReturnContract: nil,
						},
					},
					EdgeCases: []Criterion{},
				},
			},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCases: []TestCase{
				{
					Id:                  "TS-01-1",
					RequirementId:       danglingID, // references non-existent requirement
					Kind:                "unit",
					Description:         "Test with dangling ref",
					Preconditions:       []string{},
					Expected:            "error",
					AssertionPseudocode: "assert false",
				},
			},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}
}

// buildSpecWithWarnings creates a Spec that passes validation but triggers
// at least one warning (a requirement with no test coverage).
func buildSpecWithWarnings() *Spec {
	return &Spec{
		SpecID:        "01",
		SpecName:      "test_feature",
		Title:         "Test Feature",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			Introduction:  "Test.",
			Glossary:      RequirementsV1JsonGlossary{"spec": "A specification."},
			Requirements: []Requirement{
				{
					Id:    "01-REQ-1",
					Title: "Data Model",
					UserStory: UserStory{
						Role:    "developer",
						Goal:    "have typed models",
						Benefit: "type safety",
					},
					AcceptanceCriteria: []Criterion{
						{
							Id:             "01-REQ-1.1",
							EarsPattern:    CriterionEarsPatternUbiquitous,
							System:         "the system",
							Action:         "return a populated Spec",
							ReturnContract: nil,
						},
					},
					EdgeCases: []Criterion{},
				},
				{
					Id:    "01-REQ-2",
					Title: "Uncovered Requirement",
					UserStory: UserStory{
						Role:    "developer",
						Goal:    "have tests",
						Benefit: "confidence",
					},
					AcceptanceCriteria: []Criterion{
						{
							Id:             "01-REQ-2.1",
							EarsPattern:    CriterionEarsPatternUbiquitous,
							System:         "the system",
							Action:         "do something",
							ReturnContract: nil,
						},
					},
					EdgeCases: []Criterion{},
				},
			},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCases: []TestCase{
				{
					Id:                  "TS-01-1",
					RequirementId:       "01-REQ-1.1",
					Kind:                "unit",
					Description:         "Test requirement 1",
					Preconditions:       []string{},
					Expected:            "pass",
					AssertionPseudocode: "assert true",
				},
			},
			// No test for 01-REQ-2 — triggers a coverage warning
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage: Coverage{
				RequirementsCovered: []string{"01-REQ-1.1"},
			},
		},
		Tasks: &TasksV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}
}

// buildSpecWithNoTestEntries creates a Spec with requirements but no test entries.
func buildSpecWithNoTestEntries() *Spec {
	return &Spec{
		SpecID:        "01",
		SpecName:      "test_feature",
		Title:         "Test Feature",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			Introduction:  "Test.",
			Glossary:      RequirementsV1JsonGlossary{"spec": "A specification."},
			Requirements: []Requirement{
				{
					Id:    "01-REQ-1",
					Title: "Data Model",
					UserStory: UserStory{
						Role:    "developer",
						Goal:    "have typed models",
						Benefit: "type safety",
					},
					AcceptanceCriteria: []Criterion{
						{
							Id:             "01-REQ-1.1",
							EarsPattern:    CriterionEarsPatternUbiquitous,
							System:         "the system",
							Action:         "return a populated Spec",
							ReturnContract: nil,
						},
					},
					EdgeCases: []Criterion{},
				},
			},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCases:     []TestCase{}, // no test entries
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}
}

// buildSpecWithBadID creates a Spec with a malformed ID in a requirement.
func buildSpecWithBadID() *Spec {
	return &Spec{
		SpecID:        "01",
		SpecName:      "test_feature",
		Title:         "Test Feature",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			Introduction:  "Test.",
			Glossary:      RequirementsV1JsonGlossary{"spec": "A specification."},
			Requirements: []Requirement{
				{
					Id:    "bad!!id!!format",
					Title: "Bad ID Requirement",
					UserStory: UserStory{
						Role:    "developer",
						Goal:    "test ID format",
						Benefit: "validation",
					},
					AcceptanceCriteria: []Criterion{
						{
							Id:             "bad!!criterion!!id",
							EarsPattern:    CriterionEarsPatternEventDriven,
							System:         "the system",
							Action:         "return error",
							ReturnContract: nil,
						},
					},
					EdgeCases: []Criterion{},
				},
			},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "01",
			SpecName:      "test_feature",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}
}

// buildSpecA creates a test Spec with ID "01" for cross-spec tests.
func buildSpecA() *Spec {
	return &Spec{
		SpecID:        "01",
		SpecName:      "spec_a",
		Title:         "Spec A",
		Status:        "active",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Spec A\n",
		Requirements: &RequirementsV1Json{
			SpecId:                "01",
			SpecName:              "spec_a",
			SchemaVersion:         1,
			Introduction:          "Spec A.",
			Glossary:              RequirementsV1JsonGlossary{"widget": "A UI component."},
			Requirements:          []Requirement{},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "01",
			SpecName:      "spec_a",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "01",
			SpecName:      "spec_a",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}
}

// buildSpecB creates a test Spec with ID "02" for cross-spec tests,
// with a compatible glossary.
func buildSpecB() *Spec {
	return &Spec{
		SpecID:        "02",
		SpecName:      "spec_b",
		Title:         "Spec B",
		Status:        "active",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Spec B\n",
		Requirements: &RequirementsV1Json{
			SpecId:                "02",
			SpecName:              "spec_b",
			SchemaVersion:         1,
			Introduction:          "Spec B.",
			Glossary:              RequirementsV1JsonGlossary{"gadget": "A device."},
			Requirements:          []Requirement{},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "02",
			SpecName:      "spec_b",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "02",
			SpecName:      "spec_b",
			SchemaVersion: 1,
			TestCommands: TestCommands{
				SpecTests: "go test ./...",
				AllTests:  "go test ./...",
				Linter:    "go vet ./...",
			},
			Dependencies: []TaskDependency{},
			TaskGroups:   []TaskGroup{},
			Traceability: []TraceabilityEntry{},
		},
	}
}

// buildSpecWithConflictingGlossary creates a Spec whose glossary defines
// the same term as specA but with a different meaning.
func buildSpecWithConflictingGlossary() *Spec {
	spec := buildSpecB()
	// Add "widget" with a different definition than spec A
	spec.Requirements.Glossary["widget"] = "A mechanical gear component."
	return spec
}
