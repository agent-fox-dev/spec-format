package afspec

import (
	"fmt"
	"os"
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
// and 'warnings' as empty slice.
// Test Spec: TS-01-15, Requirement: 01-REQ-8.1, 04-REQ-20.1
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

	// 'warnings' key must always be present per 04-REQ-20.1
	warnings, ok := result["warnings"].([]map[string]any)
	if !ok {
		t.Fatalf("result['warnings'] is not []map[string]any, got %T", result["warnings"])
	}
	if len(warnings) != 0 {
		t.Errorf("result['warnings'] has %d entries, want 0 for fully valid spec", len(warnings))
	}
}

// TestValidateStructured_SchemaErrors verifies that ValidateStructured on
// a spec with schema errors returns a map with 'valid' false and 'errors'
// entries with 'category', 'rule', 'message', 'file', and 'path' fields.
// Test Spec: TS-01-16, Requirement: 01-REQ-8.2, 04-REQ-20.2
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
	if file, _ := e["file"].(string); file == "" {
		t.Error("error file is empty, want non-empty")
	}
	if _, hasRule := e["rule"]; !hasRule {
		t.Error("error map missing 'rule' key")
	}
	if _, hasPath := e["path"]; !hasPath {
		t.Error("error map missing 'path' key")
	}
}

// TestValidateStructured_IntegrityErrors verifies that ValidateStructured
// on a spec with integrity errors returns entries with 'category'=='integrity',
// 'rule', 'message', 'file', and 'path' fields.
// Test Spec: TS-01-17, Requirement: 01-REQ-8.3, 04-REQ-20.2
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
			if rule, _ := e["rule"].(string); rule == "" {
				t.Error("integrity error rule is empty, want non-empty")
			}
			if msg, _ := e["message"].(string); msg == "" {
				t.Error("integrity error message is empty, want non-empty")
			}
			if _, hasFile := e["file"]; !hasFile {
				t.Error("integrity error map missing 'file' key")
			}
			if _, hasPath := e["path"]; !hasPath {
				t.Error("integrity error map missing 'path' key")
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
// 'message' and 'entity_id' fields.
// Test Spec: TS-01-18, Requirement: 01-REQ-8.4, 04-REQ-20.3
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
	if msg, _ := w["message"].(string); msg == "" {
		t.Error("warning message is empty, want non-empty")
	}
	if eid, _ := w["entity_id"].(string); eid == "" {
		t.Error("warning entity_id is empty, want non-empty")
	}
}

// TestValidateStructured_WarningsKeyAlwaysPresent verifies that the
// 'warnings' key is always present, even when there are no warnings.
// Per 04-REQ-20.E1, warnings is an empty slice when no warnings exist.
// Requirement: 01-REQ-8.E1, 04-REQ-20.E1
func TestValidateStructured_WarningsKeyAlwaysPresent(t *testing.T) {
	defer requireImplemented(t)

	spec, err := LoadSpec("./../testdata/valid_spec")
	if err != nil {
		t.Fatalf("LoadSpec returned unexpected error: %v", err)
	}

	result := spec.ValidateStructured()

	warnings, ok := result["warnings"].([]map[string]any)
	if !ok {
		t.Fatalf("result['warnings'] is not []map[string]any, got %T", result["warnings"])
	}
	if len(warnings) != 0 {
		t.Errorf("result['warnings'] has %d entries, want 0 for spec with no warnings", len(warnings))
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

// ---------------------------------------------------------------------------
// Task Group 1: Cross-file rules 3, 4, 5, 6, 8, 9
// Test helpers
// ---------------------------------------------------------------------------

// filterByCheck returns all ValidationEntry values whose Check field
// matches the given check name.
func filterByCheck(entries []ValidationEntry, check string) []ValidationEntry {
	var filtered []ValidationEntry
	for _, e := range entries {
		if e.Check == check {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// buildCrossFileBaseSpec returns a minimal *Spec with consistent IDs and
// all required fields populated. Callers mutate the returned spec to add
// specific test data for cross-file rules.
func buildCrossFileBaseSpec() *Spec {
	return &Spec{
		SpecID:        "04",
		SpecName:      "test_crossfile",
		Title:         "Cross-File Test Spec",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:                "04",
			SpecName:              "test_crossfile",
			SchemaVersion:         1,
			Introduction:          "Test spec for cross-file validation.",
			Glossary:              RequirementsV1JsonGlossary{},
			Requirements:          []Requirement{},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "04",
			SpecName:      "test_crossfile",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "04",
			SpecName:      "test_crossfile",
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

// ---------------------------------------------------------------------------
// Subtask 1.1: Property test coverage rule (04-REQ-1)
// Test Spec: TS-04-1, TS-04-2
// ---------------------------------------------------------------------------

// TestValidateCrossFile_PropertyTestCoverage verifies that ValidateCrossFile
// produces one integrity error per correctness_property that lacks a matching
// property_test (matched by property_id), with the Path field set to the
// JSON path of the uncovered property.
// Test Spec: TS-04-1, TS-04-2. Requirements: 04-REQ-1.1, 04-REQ-1.2,
// 04-REQ-1.E1, 04-REQ-1.E2, 04-REQ-1.E3.
func TestValidateCrossFile_PropertyTestCoverage(t *testing.T) {
	tests := []struct {
		name             string
		spec             *Spec
		wantErrorCount   int      // expected count of cross_file_3 errors
		wantUncoveredIDs []string // property IDs expected in RequirementID fields
		wantHasPath      bool     // if true, assert Path is non-empty on each error
	}{
		{
			name: "one_uncovered_property",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.CorrectnessProperties = []CorrectnessProperty{
					{Id: "04-PROP-1", Title: "Prop A", ForAny: "any input", Invariant: "holds", Validates: []string{"04-REQ-1.1"}},
					{Id: "04-PROP-2", Title: "Prop B", ForAny: "any input", Invariant: "holds", Validates: []string{"04-REQ-1.1"}},
				}
				s.TestSpec.PropertyTests = []PropertyTest{
					{Id: "TS-04-P1", PropertyId: "04-PROP-1", Description: "test A", ForAnyStrategy: "random", InvariantCheck: "assert true", Validates: []string{"04-REQ-1.1"}},
				}
				return s
			}(),
			wantErrorCount:   1,
			wantUncoveredIDs: []string{"04-PROP-2"},
			wantHasPath:      true,
		},
		{
			name: "all_properties_covered",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.CorrectnessProperties = []CorrectnessProperty{
					{Id: "04-PROP-1", Title: "Prop A", ForAny: "any input", Invariant: "holds", Validates: []string{"04-REQ-1.1"}},
				}
				s.TestSpec.PropertyTests = []PropertyTest{
					{Id: "TS-04-P1", PropertyId: "04-PROP-1", Description: "test A", ForAnyStrategy: "random", InvariantCheck: "assert true", Validates: []string{"04-REQ-1.1"}},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_correctness_properties",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				// No correctness properties — rule should be skipped
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_test_spec_all_uncovered",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.CorrectnessProperties = []CorrectnessProperty{
					{Id: "04-PROP-1", Title: "Prop A", ForAny: "any input", Invariant: "holds", Validates: []string{"04-REQ-1.1"}},
					{Id: "04-PROP-2", Title: "Prop B", ForAny: "any input", Invariant: "holds", Validates: []string{"04-REQ-1.1"}},
				}
				s.TestSpec = nil // No test_spec — all properties uncovered
				return s
			}(),
			wantErrorCount:   2,
			wantUncoveredIDs: []string{"04-PROP-1", "04-PROP-2"},
		},
		{
			name: "duplicate_property_tests_still_covered",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.CorrectnessProperties = []CorrectnessProperty{
					{Id: "04-PROP-1", Title: "Prop A", ForAny: "any input", Invariant: "holds", Validates: []string{"04-REQ-1.1"}},
				}
				s.TestSpec.PropertyTests = []PropertyTest{
					{Id: "TS-04-P1", PropertyId: "04-PROP-1", Description: "test A", ForAnyStrategy: "random", InvariantCheck: "assert true", Validates: []string{"04-REQ-1.1"}},
					{Id: "TS-04-P2", PropertyId: "04-PROP-1", Description: "test A dup", ForAnyStrategy: "random", InvariantCheck: "assert true", Validates: []string{"04-REQ-1.1"}},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "cross_file_3")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_file_3 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_file_3 error category = %q, want %q", e.Category, "integrity")
				}
				if e.Message == "" {
					t.Error("cross_file_3 error has empty Message")
				}
			}

			// Check that uncovered property IDs appear in RequirementID fields
			for _, wantID := range tt.wantUncoveredIDs {
				found := false
				for _, e := range errors {
					if e.RequirementID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected uncovered property %q in RequirementID fields, not found", wantID)
				}
			}

			// TS-04-2: Path field should reference the correctness_property location
			if tt.wantHasPath {
				for _, e := range errors {
					if e.Path == "" {
						t.Errorf("cross_file_3 error for %q has empty Path, want non-empty JSON path", e.RequirementID)
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 1.2: Execution path smoke test coverage rule (04-REQ-2)
// Test Spec: TS-04-3
// ---------------------------------------------------------------------------

// TestValidateCrossFile_ExecutionPathSmokeCoverage verifies that
// ValidateCrossFile produces one integrity error per execution_path that
// lacks a matching smoke_test (matched by execution_path_id).
// Test Spec: TS-04-3. Requirements: 04-REQ-2.1, 04-REQ-2.E1, 04-REQ-2.E2.
func TestValidateCrossFile_ExecutionPathSmokeCoverage(t *testing.T) {
	tests := []struct {
		name             string
		spec             *Spec
		wantErrorCount   int
		wantUncoveredIDs []string // execution_path IDs expected in error Messages
	}{
		{
			name: "one_uncovered_execution_path",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.ExecutionPaths = []ExecutionPath{
					{Id: "04-PATH-1", Title: "Path A", Steps: []PathStep{{Actor: "CLI", Action: "step 1"}, {Actor: "System", Action: "step 2"}}},
					{Id: "04-PATH-2", Title: "Path B", Steps: []PathStep{{Actor: "CLI", Action: "step 1"}, {Actor: "System", Action: "step 2"}}},
				}
				s.TestSpec.SmokeTests = []SmokeTest{
					{Id: "TS-04-SMOKE-1", ExecutionPathId: "04-PATH-1", Description: "smoke A", Trigger: "run", ExpectedEffects: []string{"pass"}, Mockable: []string{}, RealComponents: []string{"all"}},
				}
				return s
			}(),
			wantErrorCount:   1,
			wantUncoveredIDs: []string{"04-PATH-2"},
		},
		{
			name: "all_paths_covered",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.ExecutionPaths = []ExecutionPath{
					{Id: "04-PATH-1", Title: "Path A", Steps: []PathStep{{Actor: "CLI", Action: "step 1"}, {Actor: "System", Action: "step 2"}}},
				}
				s.TestSpec.SmokeTests = []SmokeTest{
					{Id: "TS-04-SMOKE-1", ExecutionPathId: "04-PATH-1", Description: "smoke A", Trigger: "run", ExpectedEffects: []string{"pass"}, Mockable: []string{}, RealComponents: []string{"all"}},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_execution_paths",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				// No execution paths — rule should be skipped
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_test_spec_all_uncovered",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.ExecutionPaths = []ExecutionPath{
					{Id: "04-PATH-1", Title: "Path A", Steps: []PathStep{{Actor: "CLI", Action: "step 1"}, {Actor: "System", Action: "step 2"}}},
					{Id: "04-PATH-2", Title: "Path B", Steps: []PathStep{{Actor: "CLI", Action: "step 1"}, {Actor: "System", Action: "step 2"}}},
				}
				s.TestSpec = nil // No test_spec — all paths uncovered
				return s
			}(),
			wantErrorCount:   2,
			wantUncoveredIDs: []string{"04-PATH-1", "04-PATH-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "cross_file_4")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_file_4 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_file_4 error category = %q, want %q", e.Category, "integrity")
				}
				if e.Message == "" {
					t.Error("cross_file_4 error has empty Message")
				}
			}

			// Verify uncovered path IDs appear in error messages
			for _, wantID := range tt.wantUncoveredIDs {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, wantID) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected uncovered path %q in error Message, not found", wantID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 1.3: test_spec_id resolution rule (04-REQ-3)
// Test Spec: TS-04-4
// ---------------------------------------------------------------------------

// TestValidateCrossFile_TestSpecIDResolution verifies that ValidateCrossFile
// produces one integrity error per unresolvable test_spec_id found in
// traceability entries or subtask test_spec_refs.
// Test Spec: TS-04-4. Requirements: 04-REQ-3.1, 04-REQ-3.E1, 04-REQ-3.E2.
func TestValidateCrossFile_TestSpecIDResolution(t *testing.T) {
	tests := []struct {
		name           string
		spec           *Spec
		wantErrorCount int
		wantMissing    []string // unresolvable IDs expected in Messages
	}{
		{
			name: "unresolvable_in_traceability_and_subtask_refs",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.Traceability = []TraceabilityEntry{
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-NONEXISTENT", TaskId: "4.1"},
				}
				s.Tasks.TaskGroups = []TaskGroup{
					{
						Id:    1,
						Title: "Group 1",
						Kind:  TaskGroupKindStandard,
						Subtasks: []Subtask{
							{
								Id: "1.1", Title: "sub1", State: SubtaskStatePending,
								Details:         []string{"detail"},
								RequirementRefs: []string{},
								TestSpecRefs:    []string{"TS-04-ALSO-MISSING"},
							},
						},
						Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
					},
				}
				return s
			}(),
			wantErrorCount: 2,
			wantMissing:    []string{"TS-04-NONEXISTENT", "TS-04-ALSO-MISSING"},
		},
		{
			name: "same_unresolvable_id_in_both_locations",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.Traceability = []TraceabilityEntry{
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-GHOST", TaskId: "4.1"},
				}
				s.Tasks.TaskGroups = []TaskGroup{
					{
						Id:    1,
						Title: "Group 1",
						Kind:  TaskGroupKindStandard,
						Subtasks: []Subtask{
							{
								Id: "1.1", Title: "sub1", State: SubtaskStatePending,
								Details:         []string{"detail"},
								RequirementRefs: []string{},
								TestSpecRefs:    []string{"TS-04-GHOST"},
							},
						},
						Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
					},
				}
				return s
			}(),
			wantErrorCount: 2, // one per occurrence, not deduplicated
			wantMissing:    []string{"TS-04-GHOST"},
		},
		{
			name: "all_ids_resolve",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.TestSpec.TestCases = []TestCase{
					{Id: "TS-04-1", RequirementId: "04-REQ-1.1", Kind: "unit", Description: "test", Preconditions: []string{}, Expected: "pass", AssertionPseudocode: "assert true"},
				}
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "coverage",
						},
						AcceptanceCriteria: []Criterion{
							{Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous, System: "the system", Action: "do something", ReturnContract: nil},
						},
						EdgeCases: []Criterion{},
					},
				}
				s.Tasks.Traceability = []TraceabilityEntry{
					{RequirementId: "04-REQ-1.1", TestSpecId: "TS-04-1", TaskId: "4.1"},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_traceability_no_refs",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				// No traceability entries, no subtask test_spec_refs
				return s
			}(),
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "cross_file_5")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_file_5 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_file_5 error category = %q, want %q", e.Category, "integrity")
				}
			}

			// Verify unresolvable IDs appear in error messages
			for _, wantID := range tt.wantMissing {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, wantID) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected unresolvable ID %q in error Message, not found", wantID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 1.4: Glossary backtick term check (04-REQ-4)
// Test Spec: TS-04-5, TS-04-6
// ---------------------------------------------------------------------------

// TestValidateCrossFile_GlossaryBacktickTerms verifies that ValidateCrossFile
// extracts backtick-wrapped terms from criterion and property fields and
// produces one integrity error per term not present in the glossary, after
// applying exclusion rules (numeric, single-char, quoted, >80 chars).
// Test Spec: TS-04-5. Requirements: 04-REQ-4.1, 04-REQ-4.E1 through 04-REQ-4.E5.
func TestValidateCrossFile_GlossaryBacktickTerms(t *testing.T) {
	tests := []struct {
		name           string
		spec           *Spec
		wantErrorCount int
		wantTerms      []string // terms expected in error Messages
		wantAbsent     []string // terms that should NOT appear in errors
	}{
		{
			name: "missing_glossary_term_with_numeric_excluded",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "call `MyTerm` with value `-1`",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				// Glossary does not contain 'MyTerm'
				return s
			}(),
			wantErrorCount: 1,
			wantTerms:      []string{"MyTerm"},
			wantAbsent:     []string{"-1"}, // numeric exclusion
		},
		{
			name: "term_in_glossary_no_error",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Glossary = RequirementsV1JsonGlossary{
					"MyTerm": "A defined term.",
				}
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "call `MyTerm` method",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "single_char_excluded",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "set `x` to zero",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
			wantAbsent:     []string{"x"},
		},
		{
			name: "longer_than_80_chars_excluded",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				longTerm := strings.Repeat("a", 81) // 81 chars
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "use `" + longTerm + "` here",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "decimal_numeric_excluded",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "return `3.14` as the result",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
			wantAbsent:     []string{"3.14"},
		},
		{
			name: "quoted_string_excluded",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "return `\"hello world\"` as string",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "multiple_occurrences_one_error_each",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "call `UndefinedTerm` method",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
					{
						Id:    "04-REQ-2",
						Title: "Req 2",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-2.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "use `UndefinedTerm` again",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 2, // one per occurrence, not deduplicated
			wantTerms:      []string{"UndefinedTerm"},
		},
		{
			name: "correctness_property_fields_checked",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.CorrectnessProperties = []CorrectnessProperty{
					{
						Id:        "04-PROP-1",
						Title:     "Prop 1",
						ForAny:    "any `PropTerm` input",
						Invariant: "the `InvTerm` holds",
						Validates: []string{"04-REQ-1.1"},
					},
				}
				return s
			}(),
			wantErrorCount: 2,
			wantTerms:      []string{"PropTerm", "InvTerm"},
		},
		{
			name: "empty_glossary_all_terms_flagged",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Glossary = RequirementsV1JsonGlossary{} // empty
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{
								Id:             "04-REQ-1.1",
								EarsPattern:    CriterionEarsPatternUbiquitous,
								System:         "the system",
								Action:         "call `TermA` and `TermB`",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 2,
			wantTerms:      []string{"TermA", "TermB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "cross_file_6")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_file_6 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_file_6 error category = %q, want %q", e.Category, "integrity")
				}
			}

			// Verify expected terms appear in error messages
			for _, wantTerm := range tt.wantTerms {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, wantTerm) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected term %q in error Message, not found", wantTerm)
				}
			}

			// Verify excluded terms do NOT appear in error messages
			for _, absentTerm := range tt.wantAbsent {
				for _, e := range errors {
					if strings.Contains(e.Message, absentTerm) {
						t.Errorf("excluded term %q should not appear in error Message, but found in: %q",
							absentTerm, e.Message)
					}
				}
			}
		})
	}
}

// TestValidateCrossFile_BacktickRegexPackageLevel verifies that the backtick
// extraction regex is compiled at package initialization time (package-level
// var) rather than inside the ValidateCrossFile function body.
// Test Spec: TS-04-6. Requirement: 04-REQ-4.2.
func TestValidateCrossFile_BacktickRegexPackageLevel(t *testing.T) {
	// Read all non-test Go source files in the package directory
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read package directory: %v", err)
	}

	var foundPackageLevel bool
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("failed to read %s: %v", name, err)
		}
		content := string(data)

		// Check for a package-level backtick regex variable
		// (declared outside any function body as a var or with MustCompile)
		if strings.Contains(content, "backtick") && strings.Contains(content, "regexp.MustCompile") {
			foundPackageLevel = true
		}
	}

	if !foundPackageLevel {
		t.Error("no package-level backtick regex variable found; " +
			"expected a var declaration with regexp.MustCompile for backtick term extraction")
	}
}

// ---------------------------------------------------------------------------
// Subtask 1.5: Traceability deduplication rule (04-REQ-5)
// Test Spec: TS-04-7
// ---------------------------------------------------------------------------

// TestValidateCrossFile_TraceabilityDeduplication verifies that
// ValidateCrossFile produces one integrity error per duplicate
// (requirement_id, test_spec_id) pair beyond the first occurrence.
// Test Spec: TS-04-7. Requirements: 04-REQ-5.1, 04-REQ-5.E1, 04-REQ-5.E2.
func TestValidateCrossFile_TraceabilityDeduplication(t *testing.T) {
	tests := []struct {
		name           string
		spec           *Spec
		wantErrorCount int
		wantPairInMsg  []string // strings expected in error Messages
	}{
		{
			name: "three_duplicates_two_errors",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.Traceability = []TraceabilityEntry{
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-1", TaskId: "4.1"},
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-1", TaskId: "4.1"},
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-1", TaskId: "4.1"},
				}
				return s
			}(),
			wantErrorCount: 2, // N-1 = 3-1 = 2
			wantPairInMsg:  []string{"04-REQ-1", "TS-04-1"},
		},
		{
			name: "two_duplicates_one_error",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.Traceability = []TraceabilityEntry{
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-1", TaskId: "4.1"},
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-1", TaskId: "4.2"},
				}
				return s
			}(),
			wantErrorCount: 1,
			wantPairInMsg:  []string{"04-REQ-1", "TS-04-1"},
		},
		{
			name: "all_unique_no_error",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.Traceability = []TraceabilityEntry{
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-1", TaskId: "4.1"},
					{RequirementId: "04-REQ-1", TestSpecId: "TS-04-2", TaskId: "4.2"},
					{RequirementId: "04-REQ-2", TestSpecId: "TS-04-1", TaskId: "4.3"},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_traceability_entries",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				// No traceability entries
				return s
			}(),
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "cross_file_8")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_file_8 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_file_8 error category = %q, want %q", e.Category, "integrity")
				}
				if e.Message == "" {
					t.Error("cross_file_8 error has empty Message")
				}
			}

			// Verify pair identifiers appear in error messages
			if tt.wantErrorCount > 0 {
				for _, wantStr := range tt.wantPairInMsg {
					found := false
					for _, e := range errors {
						if strings.Contains(e.Message, wantStr) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected %q in cross_file_8 error Message, not found", wantStr)
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 1.6: Subtask requirement_refs resolution rule (04-REQ-6)
// Test Spec: TS-04-8
// ---------------------------------------------------------------------------

// TestValidateCrossFile_SubtaskRequirementRefs verifies that
// ValidateCrossFile produces one integrity error per subtask
// requirement_refs entry that does not resolve to a known requirement ID,
// criterion ID, or edge case ID.
// Test Spec: TS-04-8. Requirements: 04-REQ-6.1, 04-REQ-6.E1, 04-REQ-6.E2.
func TestValidateCrossFile_SubtaskRequirementRefs(t *testing.T) {
	tests := []struct {
		name           string
		spec           *Spec
		wantErrorCount int
		wantMissing    []string // unresolvable refs expected in Messages
		wantAbsent     []string // valid refs that should NOT appear as errors
	}{
		{
			name: "one_unresolvable_ref",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous, System: "the system", Action: "do something", ReturnContract: nil},
						},
						EdgeCases: []Criterion{
							{Id: "04-REQ-1.E1", EarsPattern: CriterionEarsPatternUbiquitous, System: "the system", Action: "handle edge", ReturnContract: nil},
						},
					},
				}
				s.Tasks.TaskGroups = []TaskGroup{
					{
						Id:    1,
						Title: "Group 1",
						Kind:  TaskGroupKindStandard,
						Subtasks: []Subtask{
							{
								Id: "1.1", Title: "sub1", State: SubtaskStatePending,
								Details:         []string{"detail"},
								RequirementRefs: []string{"04-REQ-1", "04-REQ-UNKNOWN", "04-REQ-1.E1"},
								TestSpecRefs:    []string{},
							},
						},
						Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
					},
				}
				return s
			}(),
			wantErrorCount: 1,
			wantMissing:    []string{"04-REQ-UNKNOWN"},
			wantAbsent:     []string{"04-REQ-1.E1"}, // edge case ID is valid
		},
		{
			name: "all_refs_resolve_including_edge_cases",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-1",
						Title: "Req 1",
						UserStory: UserStory{
							Role: "dev", Goal: "test", Benefit: "validation",
						},
						AcceptanceCriteria: []Criterion{
							{Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous, System: "the system", Action: "do something", ReturnContract: nil},
						},
						EdgeCases: []Criterion{
							{Id: "04-REQ-1.E1", EarsPattern: CriterionEarsPatternUbiquitous, System: "the system", Action: "handle edge", ReturnContract: nil},
						},
					},
				}
				s.Tasks.TaskGroups = []TaskGroup{
					{
						Id:    1,
						Title: "Group 1",
						Kind:  TaskGroupKindStandard,
						Subtasks: []Subtask{
							{
								Id: "1.1", Title: "sub1", State: SubtaskStatePending,
								Details:         []string{"detail"},
								RequirementRefs: []string{"04-REQ-1", "04-REQ-1.1", "04-REQ-1.E1"},
								TestSpecRefs:    []string{},
							},
						},
						Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_subtasks_no_requirement_refs",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				// No task groups or subtasks
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "subtask_with_empty_requirement_refs",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.TaskGroups = []TaskGroup{
					{
						Id:    1,
						Title: "Group 1",
						Kind:  TaskGroupKindStandard,
						Subtasks: []Subtask{
							{
								Id: "1.1", Title: "sub1", State: SubtaskStatePending,
								Details:         []string{"detail"},
								RequirementRefs: []string{}, // empty — nothing to resolve
								TestSpecRefs:    []string{},
							},
						},
						Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "cross_file_9")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_file_9 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_file_9 error category = %q, want %q", e.Category, "integrity")
				}
			}

			// Verify unresolvable refs appear in error messages
			for _, wantRef := range tt.wantMissing {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, wantRef) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected unresolvable ref %q in error Message, not found", wantRef)
				}
			}

			// Verify valid refs do NOT appear as errors
			for _, absentRef := range tt.wantAbsent {
				for _, e := range errors {
					if strings.Contains(e.Message, absentRef) && !strings.Contains(e.Message, "04-REQ-UNKNOWN") {
						t.Errorf("valid ref %q should not appear as error, but found in: %q",
							absentRef, e.Message)
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Task Group 2: Cross-file rules 7, 8, 9, 10 and task group structure
// Cross-spec rules 1, 3, 4, 5
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Subtask 2.1: Unwanted pattern return_contract rule (04-REQ-7)
// Test Spec: TS-04-9
// ---------------------------------------------------------------------------

// TestValidateCrossFile_UnwantedReturnContract verifies that ValidateCrossFile
// produces one integrity error for each criterion with ears_pattern='unwanted'
// that has a null or empty return_contract, and no error for unwanted criteria
// with a non-empty return_contract.
// Test Spec: TS-04-9. Requirements: 04-REQ-7.1, 04-REQ-7.E1, 04-REQ-7.E2.
func TestValidateCrossFile_UnwantedReturnContract(t *testing.T) {
	tests := []struct {
		name             string
		spec             *Spec
		wantErrorCount   int
		wantCriterionIDs []string // criterion IDs expected in error Messages
	}{
		{
			name: "null_return_contract_produces_error",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-99",
						Title: "Unwanted Test",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "04-REQ-99.1", EarsPattern: CriterionEarsPatternUnwanted,
								System: "the system", Action: "detect invalid input",
								ReturnContract: nil, // null — should trigger error
							},
							{
								Id: "04-REQ-99.2", EarsPattern: CriterionEarsPatternUnwanted,
								System: "the system", Action: "detect another error",
								ReturnContract: strPtr("returns error"), // non-null — no error
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount:   1,
			wantCriterionIDs: []string{"04-REQ-99.1"},
		},
		{
			name: "empty_string_return_contract_treated_as_null",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-99",
						Title: "Unwanted Test",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "04-REQ-99.1", EarsPattern: CriterionEarsPatternUnwanted,
								System: "the system", Action: "detect invalid input",
								ReturnContract: strPtr(""), // empty string — treated same as null
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount:   1,
			wantCriterionIDs: []string{"04-REQ-99.1"},
		},
		{
			name: "non_unwanted_pattern_null_contract_no_error",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-99",
						Title: "Non-Unwanted Test",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "04-REQ-99.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "do something",
								ReturnContract: nil, // null but not 'unwanted' — no error from this rule
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name:           "no_unwanted_criteria_skips_rule",
			spec:           buildCrossFileBaseSpec(),
			wantErrorCount: 0,
		},
		{
			name: "edge_case_criteria_also_checked",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Requirements.Requirements = []Requirement{
					{
						Id:    "04-REQ-99",
						Title: "Unwanted Edge Case Test",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "04-REQ-99.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "do something",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{
							{
								Id: "04-REQ-99.E1", EarsPattern: CriterionEarsPatternUnwanted,
								System: "the system", Action: "handle edge error",
								ReturnContract: nil, // null unwanted edge case — should trigger error
							},
						},
					},
				}
				return s
			}(),
			wantErrorCount:   1,
			wantCriterionIDs: []string{"04-REQ-99.E1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "cross_file_10")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_file_10 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_file_10 error category = %q, want %q", e.Category, "integrity")
				}
				if e.Message == "" {
					t.Error("cross_file_10 error has empty Message")
				}
			}

			// Check that expected criterion IDs appear in error messages
			for _, wantID := range tt.wantCriterionIDs {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, wantID) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected criterion %q in error Message, not found", wantID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 2.2: Task group structure validation (04-REQ-8)
// Test Spec: TS-04-10, TS-04-11
// ---------------------------------------------------------------------------

// TestValidateCrossFile_TaskGroupStructure verifies that ValidateCrossFile
// produces a schema-category error when the first task group does not have
// kind='tests' or the last task group does not have kind='wiring_verification'.
// Test Spec: TS-04-10, TS-04-11. Requirements: 04-REQ-8.1, 04-REQ-8.2,
// 04-REQ-8.E1, 04-REQ-8.E2.
func TestValidateCrossFile_TaskGroupStructure(t *testing.T) {
	tests := []struct {
		name              string
		spec              *Spec
		wantErrorCount    int
		wantFirstGroupErr bool // error about first group not being 'tests'
		wantLastGroupErr  bool // error about last group not being 'wiring_verification'
	}{
		{
			name: "first_group_not_tests",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.TaskGroups = []TaskGroup{
					{Id: 1, Title: "Group 1", Kind: TaskGroupKindStandard, Subtasks: []Subtask{}, Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}}},
					{Id: 2, Title: "Group 2", Kind: TaskGroupKindWiringVerification, Subtasks: []Subtask{}, Verification: VerificationSubtask{Id: "2.V", Checks: []string{"check"}}},
				}
				return s
			}(),
			wantErrorCount:    1,
			wantFirstGroupErr: true,
		},
		{
			name: "last_group_not_wiring_verification",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.TaskGroups = []TaskGroup{
					{Id: 1, Title: "Group 1", Kind: TaskGroupKindTests, Subtasks: []Subtask{}, Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}}},
					{Id: 2, Title: "Group 2", Kind: TaskGroupKindStandard, Subtasks: []Subtask{}, Verification: VerificationSubtask{Id: "2.V", Checks: []string{"check"}}},
				}
				return s
			}(),
			wantErrorCount:   1,
			wantLastGroupErr: true,
		},
		{
			name: "correct_structure_no_errors",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.TaskGroups = []TaskGroup{
					{Id: 1, Title: "Group 1", Kind: TaskGroupKindTests, Subtasks: []Subtask{}, Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}}},
					{Id: 2, Title: "Group 2", Kind: TaskGroupKindWiringVerification, Subtasks: []Subtask{}, Verification: VerificationSubtask{Id: "2.V", Checks: []string{"check"}}},
				}
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "no_task_groups_skips_check",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.TaskGroups = []TaskGroup{} // empty
				return s
			}(),
			wantErrorCount: 0,
		},
		{
			name: "single_group_violates_both_rules",
			spec: func() *Spec {
				s := buildCrossFileBaseSpec()
				s.Tasks.TaskGroups = []TaskGroup{
					{Id: 1, Title: "Group 1", Kind: TaskGroupKindStandard, Subtasks: []Subtask{}, Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}}},
				}
				return s
			}(),
			wantErrorCount:    2,
			wantFirstGroupErr: true,
			wantLastGroupErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.ValidateCrossFile()
			errors := filterByCheck(result.Errors, "task_group_structure")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("task_group_structure error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "schema" {
					t.Errorf("task_group_structure error category = %q, want %q", e.Category, "schema")
				}
			}

			if tt.wantFirstGroupErr {
				found := false
				for _, e := range errors {
					lower := strings.ToLower(e.Message)
					if strings.Contains(lower, "tests") || strings.Contains(lower, "first") {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected error mentioning 'tests' or 'first' for first group rule, not found")
				}
			}

			if tt.wantLastGroupErr {
				found := false
				for _, e := range errors {
					lower := strings.ToLower(e.Message)
					if strings.Contains(lower, "wiring_verification") || strings.Contains(lower, "last") {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected error mentioning 'wiring_verification' or 'last' for last group rule, not found")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 2.3: Wiring verification group semantics (04-REQ-9)
// Test Spec: TS-04-12, TS-04-13, TS-04-14
// ---------------------------------------------------------------------------

// TestValidateCrossFile_WiringVerificationSemantics verifies that
// ValidateCrossFile checks the wiring_verification group for meaningful
// smoke test refs and a stub audit subtask. It produces integrity errors
// when test_spec_refs are empty, no SMOKE reference is found, and no
// subtask mentions 'stub' or 'dead'.
// Test Spec: TS-04-12, TS-04-13, TS-04-14. Requirements: 04-REQ-9.1,
// 04-REQ-9.2, 04-REQ-9.3, 04-REQ-9.E1, 04-REQ-9.E2.
func TestValidateCrossFile_WiringVerificationSemantics(t *testing.T) {
	// buildWiringSpec creates a spec with a proper first group (tests) and a
	// wiring_verification group containing the given subtasks.
	buildWiringSpec := func(subtasks []Subtask) *Spec {
		s := buildCrossFileBaseSpec()
		s.Tasks.TaskGroups = []TaskGroup{
			{
				Id: 1, Title: "Tests", Kind: TaskGroupKindTests,
				Subtasks:     []Subtask{},
				Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
			},
			{
				Id: 2, Title: "Wiring", Kind: TaskGroupKindWiringVerification,
				Subtasks:     subtasks,
				Verification: VerificationSubtask{Id: "2.V", Checks: []string{"check"}},
			},
		}
		return s
	}

	t.Run("all_subtasks_empty_test_spec_refs", func(t *testing.T) {
		spec := buildWiringSpec([]Subtask{
			{Id: "2.1", Title: "subtask A", State: SubtaskStatePending, Details: []string{"detail"}, RequirementRefs: []string{}, TestSpecRefs: []string{}},
			{Id: "2.2", Title: "subtask B", State: SubtaskStatePending, Details: []string{"detail"}, RequirementRefs: []string{}, TestSpecRefs: []string{}},
		})
		result := spec.ValidateCrossFile()

		var refErrors []ValidationEntry
		for _, e := range result.Errors {
			if e.Category == "integrity" && strings.Contains(strings.ToLower(e.Message), "test_spec_refs") {
				refErrors = append(refErrors, e)
			}
		}
		if len(refErrors) < 1 {
			t.Errorf("expected at least 1 integrity error about empty test_spec_refs in wiring group, got %d; all errors: %v",
				len(refErrors), result.Errors)
		}
	})

	t.Run("no_smoke_ref_in_test_spec_refs", func(t *testing.T) {
		spec := buildWiringSpec([]Subtask{
			{
				Id: "2.1", Title: "verify", State: SubtaskStatePending,
				Details: []string{"detail"}, RequirementRefs: []string{},
				TestSpecRefs: []string{"TS-04-1", "TS-04-2"}, // no SMOKE entries
			},
		})
		result := spec.ValidateCrossFile()

		var smokeErrors []ValidationEntry
		for _, e := range result.Errors {
			if e.Category == "integrity" && strings.Contains(strings.ToLower(e.Message), "smoke") {
				smokeErrors = append(smokeErrors, e)
			}
		}
		if len(smokeErrors) < 1 {
			t.Errorf("expected at least 1 integrity error about missing smoke reference, got %d; all errors: %v",
				len(smokeErrors), result.Errors)
		}
	})

	t.Run("no_stub_or_dead_mention", func(t *testing.T) {
		spec := buildWiringSpec([]Subtask{
			{
				Id: "2.1", Title: "Run smoke tests", State: SubtaskStatePending,
				Details: []string{"Execute all integration tests"}, RequirementRefs: []string{},
				TestSpecRefs: []string{"TS-04-SMOKE-1"},
			},
		})
		result := spec.ValidateCrossFile()

		var stubErrors []ValidationEntry
		for _, e := range result.Errors {
			lower := strings.ToLower(e.Message)
			if e.Category == "integrity" && (strings.Contains(lower, "stub") || strings.Contains(lower, "dead")) {
				stubErrors = append(stubErrors, e)
			}
		}
		if len(stubErrors) < 1 {
			t.Errorf("expected at least 1 integrity error about missing stub/dead-code audit, got %d; all errors: %v",
				len(stubErrors), result.Errors)
		}
	})

	t.Run("fully_valid_wiring_group", func(t *testing.T) {
		spec := buildWiringSpec([]Subtask{
			{
				Id: "2.1", Title: "Run smoke tests", State: SubtaskStatePending,
				Details: []string{"detail"}, RequirementRefs: []string{},
				TestSpecRefs: []string{"TS-04-SMOKE-1"},
			},
			{
				Id: "2.2", Title: "Audit stub removal", State: SubtaskStatePending,
				Details: []string{"Verify all stubs are replaced"}, RequirementRefs: []string{},
				TestSpecRefs: []string{"TS-04-SMOKE-2"},
			},
		})
		result := spec.ValidateCrossFile()

		// Should have no wiring-specific integrity errors
		for _, e := range result.Errors {
			lower := strings.ToLower(e.Message)
			isWiringErr := (strings.Contains(lower, "wiring") &&
				(strings.Contains(lower, "test_spec_refs") ||
					strings.Contains(lower, "smoke") ||
					strings.Contains(lower, "stub") ||
					strings.Contains(lower, "dead")))
			if e.Category == "integrity" && isWiringErr {
				t.Errorf("unexpected wiring verification error: %v", e)
			}
		}
	})

	t.Run("no_wiring_group_skips_checks", func(t *testing.T) {
		s := buildCrossFileBaseSpec()
		s.Tasks.TaskGroups = []TaskGroup{
			{
				Id: 1, Title: "Tests", Kind: TaskGroupKindTests,
				Subtasks:     []Subtask{},
				Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
			},
		}
		result := s.ValidateCrossFile()

		// No wiring verification errors should appear when no wiring group exists
		for _, e := range result.Errors {
			lower := strings.ToLower(e.Message)
			isWiringErr := (strings.Contains(lower, "wiring") &&
				(strings.Contains(lower, "test_spec_refs") ||
					strings.Contains(lower, "smoke") ||
					strings.Contains(lower, "stub") ||
					strings.Contains(lower, "dead")))
			if e.Category == "integrity" && isWiringErr {
				t.Errorf("unexpected wiring verification error when no wiring group: %v", e)
			}
		}
	})

	t.Run("wiring_group_no_subtasks_three_errors", func(t *testing.T) {
		spec := buildWiringSpec([]Subtask{}) // no subtasks
		result := spec.ValidateCrossFile()

		var wiringErrors int
		for _, e := range result.Errors {
			lower := strings.ToLower(e.Message)
			if e.Category == "integrity" && (strings.Contains(lower, "test_spec_refs") ||
				strings.Contains(lower, "smoke") ||
				strings.Contains(lower, "stub") || strings.Contains(lower, "dead")) {
				wiringErrors++
			}
		}
		if wiringErrors != 3 {
			t.Errorf("expected 3 wiring verification errors for empty subtask group, got %d; all errors: %v",
				wiringErrors, result.Errors)
		}
	})
}

// ---------------------------------------------------------------------------
// Subtask 2.4: Cross-spec duplicate API symbol rule (04-REQ-10)
// Test Spec: TS-04-15
// ---------------------------------------------------------------------------

// TestValidateCrossSpec_DuplicateAPISymbol verifies that ValidateCrossSpec
// produces one integrity error per mismatched API symbol signature between
// two specs connected by a DependencyEdge.
// Test Spec: TS-04-15. Requirements: 04-REQ-10.1, 04-REQ-10.E1, 04-REQ-10.E2.
func TestValidateCrossSpec_DuplicateAPISymbol(t *testing.T) {
	tests := []struct {
		name           string
		specs          []*Spec
		graph          *DependencyGraph
		wantErrorCount int
		wantSymbol     string // symbol name expected in error Messages
	}{
		{
			name: "mismatched_signature_produces_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Foo", ImportPath: "pkg/", Signature: "func Foo() int"},
					}},
				}
				b := buildSpecB()
				b.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Foo", ImportPath: "pkg/", Signature: "func Foo() string"},
					}},
				}
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 1,
			wantSymbol:     "Foo",
		},
		{
			name: "matching_signature_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Foo", ImportPath: "pkg/", Signature: "func Foo() int"},
					}},
				}
				b := buildSpecB()
				b.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Foo", ImportPath: "pkg/", Signature: "func Foo() int"},
					}},
				}
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 0,
		},
		{
			name: "same_name_not_connected_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Foo", ImportPath: "pkg/", Signature: "func Foo() int"},
					}},
				}
				b := buildSpecB()
				b.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Foo", ImportPath: "pkg/", Signature: "func Foo() string"},
					}},
				}
				return []*Spec{a, b}
			}(),
			graph:          &DependencyGraph{Edges: []DependencyEdge{}}, // no connection
			wantErrorCount: 0,
		},
		{
			name: "no_shared_symbols_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Foo", ImportPath: "pkg/", Signature: "func Foo() int"},
					}},
				}
				b := buildSpecB()
				b.Requirements.ExternalApis = []ExternalApi{
					{Package: "pkg", Version: "v1.0", Symbols: []ExternalApiSymbol{
						{Name: "Bar", ImportPath: "pkg/", Signature: "func Bar() int"},
					}},
				}
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 0,
		},
		{
			name:           "empty_dependency_graph_no_error",
			specs:          []*Spec{buildSpecA(), buildSpecB()},
			graph:          &DependencyGraph{Edges: []DependencyEdge{}},
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCrossSpec(tt.specs, tt.graph)
			errors := filterByCheck(result.Errors, "cross_spec_1")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_spec_1 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_spec_1 error category = %q, want %q", e.Category, "integrity")
				}
			}

			if tt.wantSymbol != "" {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantSymbol) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected symbol %q in error Message, not found", tt.wantSymbol)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 2.5: Cross-spec unknown dependency and interface contract rules
// (04-REQ-11, 04-REQ-12)
// Test Spec: TS-04-16, TS-04-17
// ---------------------------------------------------------------------------

// TestValidateCrossSpec_UnknownDependency verifies that ValidateCrossSpec
// produces one integrity error per depends_on_spec value that does not
// match a key in the provided spec map.
// Test Spec: TS-04-16. Requirements: 04-REQ-11.1, 04-REQ-11.E1.
func TestValidateCrossSpec_UnknownDependency(t *testing.T) {
	tests := []struct {
		name           string
		specs          []*Spec
		graph          *DependencyGraph
		wantErrorCount int
		wantUnknown    string // unknown spec_id expected in error Messages
	}{
		{
			name: "unknown_depends_on_spec_produces_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Tasks.Dependencies = []TaskDependency{
					{DependsOnSpec: "specZ", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
				}
				return []*Spec{a, buildSpecB()}
			}(),
			graph:          &DependencyGraph{Edges: []DependencyEdge{}},
			wantErrorCount: 1,
			wantUnknown:    "specZ",
		},
		{
			name: "all_dependencies_resolve_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Tasks.Dependencies = []TaskDependency{
					{DependsOnSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
				}
				return []*Spec{a, buildSpecB()}
			}(),
			graph:          &DependencyGraph{Edges: []DependencyEdge{}},
			wantErrorCount: 0,
		},
		{
			name:           "no_task_dependencies_no_error",
			specs:          []*Spec{buildSpecA(), buildSpecB()},
			graph:          &DependencyGraph{Edges: []DependencyEdge{}},
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCrossSpec(tt.specs, tt.graph)
			errors := filterByCheck(result.Errors, "cross_spec_3")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_spec_3 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_spec_3 error category = %q, want %q", e.Category, "integrity")
				}
			}

			if tt.wantUnknown != "" {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantUnknown) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected unknown spec %q in error Message, not found", tt.wantUnknown)
				}
			}
		})
	}
}

// TestValidateCrossSpec_InterfaceContractMismatch verifies that
// ValidateCrossSpec produces one integrity error when a downstream spec's
// criterion references a backtick-wrapped symbol with a different
// return_contract than the upstream spec's definition.
// Test Spec: TS-04-17. Requirements: 04-REQ-12.1, 04-REQ-12.E1.
func TestValidateCrossSpec_InterfaceContractMismatch(t *testing.T) {
	tests := []struct {
		name           string
		specs          []*Spec
		graph          *DependencyGraph
		wantErrorCount int
		wantSymbol     string // symbol expected in error Messages
	}{
		{
			name: "different_return_contracts_produces_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.Requirements = []Requirement{
					{
						Id: "01-REQ-1", Title: "Req 1",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "01-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "call `Foo`",
								ReturnContract: strPtr("returns int"),
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				b := buildSpecB()
				b.Requirements.Requirements = []Requirement{
					{
						Id: "02-REQ-1", Title: "Req 1",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "02-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "call `Foo`",
								ReturnContract: strPtr("returns string"),
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 1,
			wantSymbol:     "Foo",
		},
		{
			name: "matching_contracts_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.Requirements = []Requirement{
					{
						Id: "01-REQ-1", Title: "Req 1",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "01-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "call `Foo`",
								ReturnContract: strPtr("returns int"),
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				b := buildSpecB()
				b.Requirements.Requirements = []Requirement{
					{
						Id: "02-REQ-1", Title: "Req 1",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "02-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "call `Foo`",
								ReturnContract: strPtr("returns int"),
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 0,
		},
		{
			name: "both_null_contracts_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.Requirements = []Requirement{
					{
						Id: "01-REQ-1", Title: "Req 1",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "01-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "call `Foo`",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				b := buildSpecB()
				b.Requirements.Requirements = []Requirement{
					{
						Id: "02-REQ-1", Title: "Req 1",
						UserStory: UserStory{Role: "dev", Goal: "test", Benefit: "validation"},
						AcceptanceCriteria: []Criterion{
							{
								Id: "02-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous,
								System: "the system", Action: "call `Foo`",
								ReturnContract: nil,
							},
						},
						EdgeCases: []Criterion{},
					},
				}
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCrossSpec(tt.specs, tt.graph)
			errors := filterByCheck(result.Errors, "cross_spec_4")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_spec_4 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_spec_4 error category = %q, want %q", e.Category, "integrity")
				}
			}

			if tt.wantSymbol != "" {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantSymbol) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected symbol %q in error Message, not found", tt.wantSymbol)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Subtask 2.6: Cross-spec missing boundary coverage rule (04-REQ-13)
// Test Spec: TS-04-18
// ---------------------------------------------------------------------------

// TestValidateCrossSpec_MissingBoundaryCoverage verifies that
// ValidateCrossSpec produces one integrity error per execution_path whose
// actor matches a spec_id in the DependencyGraph but lacks a corresponding
// smoke_test.
// Test Spec: TS-04-18. Requirements: 04-REQ-13.1, 04-REQ-13.E1, 04-REQ-13.E2.
func TestValidateCrossSpec_MissingBoundaryCoverage(t *testing.T) {
	tests := []struct {
		name           string
		specs          []*Spec
		graph          *DependencyGraph
		wantErrorCount int
		wantInMessage  []string // strings expected in error Messages
	}{
		{
			name: "actor_matches_spec_no_smoke_test",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.ExecutionPaths = []ExecutionPath{
					{
						Id: "01-PATH-1", Title: "Path X",
						Steps: []PathStep{
							{Actor: "CLI", Action: "do something"},
							{Actor: "02", Action: "call spec B"},
						},
					},
				}
				a.TestSpec.SmokeTests = []SmokeTest{} // no smoke tests
				b := buildSpecB()
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 1,
			wantInMessage:  []string{"01-PATH-1"},
		},
		{
			name: "boundary_covered_by_smoke_test_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.ExecutionPaths = []ExecutionPath{
					{
						Id: "01-PATH-1", Title: "Path X",
						Steps: []PathStep{
							{Actor: "CLI", Action: "do something"},
							{Actor: "02", Action: "call spec B"},
						},
					},
				}
				a.TestSpec.SmokeTests = []SmokeTest{
					{
						Id: "TS-01-SMOKE-1", ExecutionPathId: "01-PATH-1",
						Description: "smoke boundary", Trigger: "run",
						ExpectedEffects: []string{"pass"},
						Mockable: []string{}, RealComponents: []string{"all"},
					},
				}
				b := buildSpecB()
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 0,
		},
		{
			name: "actor_not_in_graph_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				a.Requirements.ExecutionPaths = []ExecutionPath{
					{
						Id: "01-PATH-1", Title: "Path X",
						Steps: []PathStep{
							{Actor: "CLI", Action: "do something"},
							{Actor: "unknown_spec", Action: "call unknown"},
						},
					},
				}
				a.TestSpec.SmokeTests = []SmokeTest{}
				b := buildSpecB()
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 0,
		},
		{
			name: "no_execution_paths_no_error",
			specs: func() []*Spec {
				a := buildSpecA()
				// No execution paths
				b := buildSpecB()
				return []*Spec{a, b}
			}(),
			graph: &DependencyGraph{Edges: []DependencyEdge{
				{FromSpec: "01", ToSpec: "02", FromGroup: 1, ToGroup: 1, Relationship: "depends_on"},
			}},
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCrossSpec(tt.specs, tt.graph)
			errors := filterByCheck(result.Errors, "cross_spec_5")

			if len(errors) != tt.wantErrorCount {
				t.Errorf("cross_spec_5 error count = %d, want %d; errors: %v",
					len(errors), tt.wantErrorCount, errors)
			}

			for _, e := range errors {
				if e.Category != "integrity" {
					t.Errorf("cross_spec_5 error category = %q, want %q", e.Category, "integrity")
				}
			}

			for _, want := range tt.wantInMessage {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in error Message, not found", want)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Warning rule tests (task group 8 / test spec TS-04-19 through TS-04-25)
// ---------------------------------------------------------------------------

// makeMinimalSpec creates a minimal Spec with matching IDs across all artifacts.
func makeMinimalSpec() *Spec {
	return &Spec{
		SpecID:        "04",
		SpecName:      "test_warnings",
		Title:         "Test Warnings",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:                "04",
			SpecName:              "test_warnings",
			SchemaVersion:         1,
			Introduction:          "Test.",
			Glossary:              RequirementsV1JsonGlossary{},
			Requirements:          []Requirement{},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "04",
			SpecName:      "test_warnings",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "04",
			SpecName:      "test_warnings",
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

// TS-04-19: Validate appends a warning when a task group's total
// test_spec_refs count across all subtasks exceeds 15.
// Requirement: 04-REQ-14.1
func TestValidate_Warning_GroupTestSpecRefsCeiling(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Tasks.TaskGroups = []TaskGroup{
		{
			Id: 1, Title: "Group 1", Kind: TaskGroupKindTests,
			Subtasks: []Subtask{
				{Id: "1.1", Title: "s1", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"a", "b", "c", "d", "e", "f"}, RequirementRefs: []string{}},
				{Id: "1.2", Title: "s2", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"g", "h", "i", "j", "k", "l"}, RequirementRefs: []string{}},
				{Id: "1.3", Title: "s3", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"m", "n", "o", "p", "q"}, RequirementRefs: []string{}},
			},
			Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
		},
	}
	result := spec.Validate()
	var found bool
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(w.Message, "17") || strings.Contains(lower, "test_spec_refs") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about group test_spec_refs count 17, got warnings: %v", result.Warnings)
	}
}

// 04-REQ-14.E1: Exactly 15 total test_spec_refs produces no warning.
func TestValidate_Warning_GroupTestSpecRefsCeiling_ExactlyAtThreshold(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Tasks.TaskGroups = []TaskGroup{
		{
			Id: 1, Title: "Group 1", Kind: TaskGroupKindTests,
			Subtasks: []Subtask{
				{Id: "1.1", Title: "s1", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"a", "b", "c", "d", "e"}, RequirementRefs: []string{}},
				{Id: "1.2", Title: "s2", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"f", "g", "h", "i", "j"}, RequirementRefs: []string{}},
				{Id: "1.3", Title: "s3", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"k", "l", "m", "n", "o"}, RequirementRefs: []string{}},
			},
			Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
		},
	}
	result := spec.Validate()
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if (strings.Contains(w.Message, "15") || strings.Contains(lower, "test_spec_refs")) &&
			strings.Contains(lower, "exceed") {
			t.Errorf("expected no group test_spec_refs ceiling warning at exactly 15, got: %s", w.Message)
		}
	}
}

// TS-04-20: Validate appends a warning when a task group has more than 6
// non-verification subtasks.
// Requirement: 04-REQ-15.1
func TestValidate_Warning_GroupSubtaskCount(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	subtasks := make([]Subtask, 7)
	for i := range subtasks {
		subtasks[i] = Subtask{
			Id: fmt.Sprintf("1.%d", i+1), Title: fmt.Sprintf("s%d", i+1),
			State: SubtaskStatePending, Details: []string{},
			TestSpecRefs: []string{}, RequirementRefs: []string{},
		}
	}
	spec.Tasks.TaskGroups = []TaskGroup{
		{
			Id: 1, Title: "Group 1", Kind: TaskGroupKindTests,
			Subtasks:     subtasks,
			Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
		},
	}
	result := spec.Validate()
	var found bool
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(w.Message, "7") || strings.Contains(lower, "subtask") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about group subtask count exceeding 6, got warnings: %v", result.Warnings)
	}
}

// 04-REQ-15.E1: Exactly 6 non-verification subtasks produces no warning.
func TestValidate_Warning_GroupSubtaskCount_ExactlyAtThreshold(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	subtasks := make([]Subtask, 6)
	for i := range subtasks {
		subtasks[i] = Subtask{
			Id: fmt.Sprintf("1.%d", i+1), Title: fmt.Sprintf("s%d", i+1),
			State: SubtaskStatePending, Details: []string{},
			TestSpecRefs: []string{}, RequirementRefs: []string{},
		}
	}
	spec.Tasks.TaskGroups = []TaskGroup{
		{
			Id: 1, Title: "Group 1", Kind: TaskGroupKindTests,
			Subtasks:     subtasks,
			Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
		},
	}
	result := spec.Validate()
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(lower, "subtask") && strings.Contains(lower, "exceed") {
			t.Errorf("expected no subtask count warning at exactly 6, got: %s", w.Message)
		}
	}
}

// TS-04-21: Validate appends a warning when a single subtask has more than 8
// test_spec_refs entries.
// Requirement: 04-REQ-16.1
func TestValidate_Warning_SubtaskTestSpecRefsCeiling(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Tasks.TaskGroups = []TaskGroup{
		{
			Id: 1, Title: "Group 1", Kind: TaskGroupKindTests,
			Subtasks: []Subtask{
				{Id: "1.1", Title: "big subtask", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, RequirementRefs: []string{}},
			},
			Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
		},
	}
	result := spec.Validate()
	var found bool
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(w.Message, "9") || strings.Contains(lower, "test_spec_refs") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about subtask test_spec_refs count 9, got warnings: %v", result.Warnings)
	}
}

// 04-REQ-16.E1: Exactly 8 test_spec_refs produces no warning.
func TestValidate_Warning_SubtaskTestSpecRefsCeiling_ExactlyAtThreshold(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Tasks.TaskGroups = []TaskGroup{
		{
			Id: 1, Title: "Group 1", Kind: TaskGroupKindTests,
			Subtasks: []Subtask{
				{Id: "1.1", Title: "subtask", State: SubtaskStatePending, Details: []string{},
					TestSpecRefs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, RequirementRefs: []string{}},
			},
			Verification: VerificationSubtask{Id: "1.V", Checks: []string{"check"}},
		},
	}
	result := spec.Validate()
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(lower, "test_spec_refs") && strings.Contains(lower, "exceed") && w.EntityID == "1.1" {
			t.Errorf("expected no subtask test_spec_refs warning at exactly 8, got: %s", w.Message)
		}
	}
}

// TS-04-22: Validate appends a warning when a criterion has error_condition
// or error-indicating action keyword and null return_contract.
// Requirement: 04-REQ-17.1
func TestValidate_Warning_ErrorPathReturnContract(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Requirements.Requirements = []Requirement{
		{
			Id: "04-REQ-1", Title: "Error Handling",
			UserStory:      UserStory{Role: "developer", Goal: "handle errors", Benefit: "reliability"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternEventDriven, System: "the system",
				Action: "handle the error", ErrorCondition: strPtr("network timeout"), ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		},
	}
	result := spec.Validate()
	var found bool
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(w.Message, "04-REQ-1.1") || strings.Contains(lower, "return_contract") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about missing return_contract on error path, got warnings: %v", result.Warnings)
	}
}

// TS-04-22 variant: error-indicating keyword in action triggers warning.
func TestValidate_Warning_ErrorPathReturnContract_ActionKeyword(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Requirements.Requirements = []Requirement{
		{
			Id: "04-REQ-1", Title: "Error keyword",
			UserStory:      UserStory{Role: "developer", Goal: "handle errors", Benefit: "reliability"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternEventDriven, System: "the system",
				Action: "reject the invalid input", ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		},
	}
	result := spec.Validate()
	var found bool
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(w.Message, "04-REQ-1.1") || strings.Contains(lower, "return_contract") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning for error keyword in action, got warnings: %v", result.Warnings)
	}
}

// 04-REQ-17: No warning when return_contract is non-null.
func TestValidate_Warning_ErrorPathReturnContract_NonNull(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Requirements.Requirements = []Requirement{
		{
			Id: "04-REQ-1", Title: "Error With Contract",
			UserStory:      UserStory{Role: "developer", Goal: "handle errors", Benefit: "reliability"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternEventDriven, System: "the system",
				Action: "handle the error", ErrorCondition: strPtr("network timeout"),
				ReturnContract: CriterionReturnContract(strPtr("returns HTTP 503")),
			}},
			EdgeCases: []Criterion{},
		},
	}
	result := spec.Validate()
	for _, w := range result.Warnings {
		if w.EntityID == "04-REQ-1.1" && strings.Contains(strings.ToLower(w.Message), "return_contract") {
			t.Errorf("expected no error path warning when contract is set, got: %s", w.Message)
		}
	}
}

// 04-REQ-17.E1: Unwanted pattern with null return_contract produces
// both an integrity error (cross_file_10) and a warning.
func TestValidate_Warning_ErrorPathReturnContract_UnwantedDualEntry(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Requirements.Requirements = []Requirement{
		{
			Id: "04-REQ-1", Title: "Unwanted Error Path",
			UserStory:      UserStory{Role: "developer", Goal: "handle errors", Benefit: "reliability"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternUnwanted, System: "the system",
				Action: "fail with error", ErrorCondition: strPtr("unexpected input"), ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		},
	}
	result := spec.Validate()
	var hasError bool
	for _, e := range result.Errors {
		if e.Check == "cross_file_10" && e.EntityID == "04-REQ-1.1" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected integrity error from cross_file_10 for unwanted criterion")
	}
	var hasWarning bool
	for _, w := range result.Warnings {
		if w.EntityID == "04-REQ-1.1" && strings.Contains(strings.ToLower(w.Message), "return_contract") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning about missing return_contract on error path")
	}
}

// TS-04-23: Validate appends one warning per vague word occurrence.
// Requirement: 04-REQ-18.1
func TestValidate_Warning_VagueLanguageDetection(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Requirements.Requirements = []Requirement{
		{
			Id: "04-REQ-1", Title: "Vague Criterion",
			UserStory: UserStory{Role: "developer", Goal: "detect vagueness", Benefit: "clarity"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous, System: "the system",
				Action: "handle appropriately", Trigger: strPtr("properly configured"), ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		},
	}
	result := spec.Validate()
	var vagueWarnings []ValidationEntry
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(lower, "appropriately") || strings.Contains(lower, "properly") {
			vagueWarnings = append(vagueWarnings, w)
		}
	}
	if len(vagueWarnings) != 2 {
		t.Errorf("expected 2 vague warnings (appropriately + properly), got %d: %v", len(vagueWarnings), vagueWarnings)
	}
}

// 04-REQ-18.E1: One warning per occurrence across multiple fields/requirements.
func TestValidate_Warning_VagueLanguageDetection_MultipleOccurrences(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Requirements.Requirements = []Requirement{
		{
			Id: "04-REQ-1", Title: "First",
			UserStory: UserStory{Role: "dev", Goal: "detect", Benefit: "clarity"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous, System: "sys",
				Action: "handle correctly", ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		},
		{
			Id: "04-REQ-2", Title: "Second",
			UserStory: UserStory{Role: "dev", Goal: "detect", Benefit: "clarity"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-2.1", EarsPattern: CriterionEarsPatternUbiquitous, System: "sys",
				Action: "provide adequate response with sufficient detail", ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		},
	}
	result := spec.Validate()
	var vagueWarnings []ValidationEntry
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(lower, "correctly") || strings.Contains(lower, "adequate") || strings.Contains(lower, "sufficient") {
			vagueWarnings = append(vagueWarnings, w)
		}
	}
	if len(vagueWarnings) != 3 {
		t.Errorf("expected 3 vague warnings, got %d: %v", len(vagueWarnings), vagueWarnings)
	}
}

// No vague warnings for clean language.
func TestValidate_Warning_VagueLanguageDetection_NoVagueTerms(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	spec.Requirements.Requirements = []Requirement{
		{
			Id: "04-REQ-1", Title: "Clean",
			UserStory: UserStory{Role: "dev", Goal: "clear specs", Benefit: "clarity"},
			AcceptanceCriteria: []Criterion{{
				Id: "04-REQ-1.1", EarsPattern: CriterionEarsPatternUbiquitous, System: "sys",
				Action: "return a 200 status code", ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		},
	}
	result := spec.Validate()
	for _, w := range result.Warnings {
		if strings.Contains(strings.ToLower(w.Message), "vague") {
			t.Errorf("expected no vague warnings, got: %s", w.Message)
		}
	}
}

// TS-04-24: Vague language regex is compiled at package level.
// Requirement: 04-REQ-18.2
func TestValidate_Warning_VagueLanguageRegexPackageLevel(t *testing.T) {
	if vagueLanguageRe == nil {
		t.Fatal("vagueLanguageRe is nil; expected package-level compiled regex")
	}
	for _, term := range []string{"appropriate", "properly", "correctly", "adequate", "sufficient"} {
		if !vagueLanguageRe.MatchString(term) {
			t.Errorf("vagueLanguageRe should match %q", term)
		}
	}
	for _, term := range []string{"return", "validate", "check"} {
		if vagueLanguageRe.MatchString(term) {
			t.Errorf("vagueLanguageRe should not match %q", term)
		}
	}
}

// TS-04-25: Validate appends scope limit warning when > 10 requirements.
// Requirement: 04-REQ-19.1
func TestValidate_Warning_ScopeLimit(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	for i := 1; i <= 11; i++ {
		spec.Requirements.Requirements = append(spec.Requirements.Requirements, Requirement{
			Id: fmt.Sprintf("04-REQ-%d", i), Title: fmt.Sprintf("Req %d", i),
			UserStory: UserStory{Role: "dev", Goal: "scope", Benefit: "manageable"},
			AcceptanceCriteria: []Criterion{{
				Id: fmt.Sprintf("04-REQ-%d.1", i), EarsPattern: CriterionEarsPatternUbiquitous,
				System: "sys", Action: fmt.Sprintf("do %d", i), ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		})
	}
	result := spec.Validate()
	var found bool
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(lower, "split") || strings.Contains(lower, "large") || strings.Contains(w.Message, "11") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected scope limit warning for 11 reqs, got warnings: %v", result.Warnings)
	}
}

// 04-REQ-19.E1: Exactly 10 requirements produces no warning.
func TestValidate_Warning_ScopeLimit_ExactlyAtThreshold(t *testing.T) {
	defer requireImplemented(t)
	spec := makeMinimalSpec()
	for i := 1; i <= 10; i++ {
		spec.Requirements.Requirements = append(spec.Requirements.Requirements, Requirement{
			Id: fmt.Sprintf("04-REQ-%d", i), Title: fmt.Sprintf("Req %d", i),
			UserStory: UserStory{Role: "dev", Goal: "scope", Benefit: "manageable"},
			AcceptanceCriteria: []Criterion{{
				Id: fmt.Sprintf("04-REQ-%d.1", i), EarsPattern: CriterionEarsPatternUbiquitous,
				System: "sys", Action: fmt.Sprintf("do %d", i), ReturnContract: nil,
			}},
			EdgeCases: []Criterion{},
		})
	}
	result := spec.Validate()
	for _, w := range result.Warnings {
		lower := strings.ToLower(w.Message)
		if strings.Contains(lower, "split") || strings.Contains(lower, "large") || strings.Contains(lower, "too") {
			t.Errorf("expected no scope limit warning at exactly 10 reqs, got: %s", w.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// Task Group 9: ValidateStructured output shape alignment (04-REQ-20)
// Test Spec: TS-04-26, TS-04-27, TS-04-28
// Property: 04-PROP-5
// ---------------------------------------------------------------------------

// TestValidateStructured_OutputShape_ValidKeys verifies that
// ValidateStructured returns a map with keys 'errors', 'warnings', and
// 'valid', where 'valid' is true when errors is empty.
// Test Spec: TS-04-26. Requirements: 04-REQ-20.1, 04-REQ-20.E1, 04-REQ-20.E2.
func TestValidateStructured_OutputShape_ValidKeys(t *testing.T) {
	defer requireImplemented(t)

	t.Run("valid_spec_all_keys_present", func(t *testing.T) {
		// Use a programmatically-built minimal valid spec to avoid
		// testdata dependencies on cross-file rule implementation state.
		spec := makeMinimalSpec()
		output := spec.ValidateStructured()

		// Assert all three keys exist
		if _, ok := output["valid"]; !ok {
			t.Fatal("output missing 'valid' key")
		}
		if _, ok := output["errors"]; !ok {
			t.Fatal("output missing 'errors' key")
		}
		if _, ok := output["warnings"]; !ok {
			t.Fatal("output missing 'warnings' key")
		}

		// Assert types
		valid, ok := output["valid"].(bool)
		if !ok {
			t.Fatalf("output['valid'] is not bool, got %T", output["valid"])
		}
		if !valid {
			t.Error("output['valid'] = false, want true for valid spec")
		}

		errs, ok := output["errors"].([]map[string]any)
		if !ok {
			t.Fatalf("output['errors'] is not []map[string]any, got %T", output["errors"])
		}
		if len(errs) != 0 {
			t.Errorf("output['errors'] has %d entries, want 0", len(errs))
		}

		warnings, ok := output["warnings"].([]map[string]any)
		if !ok {
			t.Fatalf("output['warnings'] is not []map[string]any, got %T", output["warnings"])
		}
		if len(warnings) != 0 {
			t.Errorf("output['warnings'] has %d entries, want 0", len(warnings))
		}
	})

	t.Run("spec_with_errors_valid_is_false", func(t *testing.T) {
		spec := buildInvalidSpec()
		output := spec.ValidateStructured()

		valid, ok := output["valid"].(bool)
		if !ok {
			t.Fatalf("output['valid'] is not bool, got %T", output["valid"])
		}
		if valid {
			t.Error("output['valid'] = true, want false for spec with errors")
		}

		errs, ok := output["errors"].([]map[string]any)
		if !ok {
			t.Fatalf("output['errors'] is not []map[string]any, got %T", output["errors"])
		}
		if len(errs) == 0 {
			t.Error("output['errors'] is empty, want non-empty for spec with errors")
		}
	})
}

// TestValidateStructured_OutputShape_ErrorFields verifies that each error
// map in ValidateStructured output includes 'category', 'rule', 'message',
// 'file', and 'path' fields, with 'category' set to 'schema' or 'integrity'
// as appropriate.
// Test Spec: TS-04-27. Requirements: 04-REQ-20.2.
func TestValidateStructured_OutputShape_ErrorFields(t *testing.T) {
	defer requireImplemented(t)

	// Build a spec that triggers both a schema error and an integrity error.
	// Schema error: requirement with empty ID violates JSON schema.
	// Integrity error: correctness property with no matching property test.
	spec := &Spec{
		SpecID:        "04",
		SpecName:      "test_output_shape",
		Title:         "Test Output Shape",
		Status:        "draft",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		Owner:         "test-author",
		Source:        "https://example.com",
		SchemaVersion: 1,
		PRDBody:       "# Test\n",
		Requirements: &RequirementsV1Json{
			SpecId:        "04",
			SpecName:      "test_output_shape",
			SchemaVersion: 1,
			Introduction:  "Test spec for output shape validation.",
			Glossary:      RequirementsV1JsonGlossary{},
			Requirements: []Requirement{
				{
					Id:    "",    // empty ID — schema violation
					Title: "",   // empty title — schema violation
					UserStory: UserStory{
						Role: "dev", Goal: "test", Benefit: "validation",
					},
					AcceptanceCriteria: []Criterion{},
					EdgeCases:          []Criterion{},
				},
			},
			CorrectnessProperties: []CorrectnessProperty{
				{
					Id: "04-PROP-1", Title: "Uncovered Prop",
					ForAny: "any input", Invariant: "holds",
					Validates: []string{"04-REQ-1.1"},
				},
			},
			ExecutionPaths: []ExecutionPath{},
			ErrorHandling:  []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			SpecId:        "04",
			SpecName:      "test_output_shape",
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{}, // no property tests — triggers cross_file_3
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
			Coverage:      Coverage{},
		},
		Tasks: &TasksV1Json{
			SpecId:        "04",
			SpecName:      "test_output_shape",
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

	output := spec.ValidateStructured()

	errs, ok := output["errors"].([]map[string]any)
	if !ok {
		t.Fatalf("output['errors'] is not []map[string]any, got %T", output["errors"])
	}
	if len(errs) == 0 {
		t.Fatal("output['errors'] is empty, want at least one entry")
	}

	// Every error map must contain all 5 keys
	for i, e := range errs {
		for _, key := range []string{"category", "rule", "message", "file", "path"} {
			if _, exists := e[key]; !exists {
				t.Errorf("errors[%d] missing key %q", i, key)
			}
		}
	}

	// Verify at least one schema error
	var hasSchema bool
	for _, e := range errs {
		if cat, _ := e["category"].(string); cat == "schema" {
			hasSchema = true
			break
		}
	}
	if !hasSchema {
		t.Error("expected at least one error with category='schema'")
	}

	// Verify at least one integrity error
	var hasIntegrity bool
	for _, e := range errs {
		if cat, _ := e["category"].(string); cat == "integrity" {
			hasIntegrity = true
			break
		}
	}
	if !hasIntegrity {
		t.Error("expected at least one error with category='integrity'")
	}
}

// TestValidateStructured_OutputShape_WarningFields verifies that each
// warning map in ValidateStructured output includes 'message' and
// 'entity_id' fields with non-empty values.
// Test Spec: TS-04-28. Requirements: 04-REQ-20.3.
func TestValidateStructured_OutputShape_WarningFields(t *testing.T) {
	defer requireImplemented(t)

	// Build a spec with 11 requirements to trigger scope limit warning.
	spec := makeMinimalSpec()
	for i := 1; i <= 11; i++ {
		spec.Requirements.Requirements = append(spec.Requirements.Requirements, Requirement{
			Id:    fmt.Sprintf("04-REQ-%d", i),
			Title: fmt.Sprintf("Requirement %d", i),
			UserStory: UserStory{
				Role: "developer", Goal: "scope test", Benefit: "manageable specs",
			},
			AcceptanceCriteria: []Criterion{
				{
					Id: fmt.Sprintf("04-REQ-%d.1", i), EarsPattern: CriterionEarsPatternUbiquitous,
					System: "the system", Action: fmt.Sprintf("do action %d", i),
					ReturnContract: nil,
				},
			},
			EdgeCases: []Criterion{},
		})
	}

	output := spec.ValidateStructured()

	warnings, ok := output["warnings"].([]map[string]any)
	if !ok {
		t.Fatalf("output['warnings'] is not []map[string]any, got %T", output["warnings"])
	}
	if len(warnings) == 0 {
		t.Fatal("output['warnings'] is empty, want at least one entry")
	}

	for i, w := range warnings {
		msg, _ := w["message"].(string)
		if msg == "" {
			t.Errorf("warnings[%d] 'message' is empty, want non-empty", i)
		}
		if _, hasEntityID := w["entity_id"]; !hasEntityID {
			t.Errorf("warnings[%d] missing 'entity_id' key", i)
		}
	}
}

// TestValidateStructured_ValidFlagConsistency is the property test for
// 04-PROP-5: the 'valid' field in ValidateStructured output is true if
// and only if the 'errors' slice is empty.
// Property: 04-PROP-5. Validates: 04-REQ-20.1, 04-REQ-20.E1, 04-REQ-20.E2.
func TestValidateStructured_ValidFlagConsistency(t *testing.T) {
	defer requireImplemented(t)

	cases := []struct {
		name string
		spec *Spec
	}{
		{"valid_spec", func() *Spec {
			s, _ := LoadSpec("./../testdata/valid_spec")
			return s
		}()},
		{"invalid_spec", buildInvalidSpec()},
		{"warnings_only_spec", buildSpecWithWarnings()},
		{"minimal_spec", makeMinimalSpec()},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.spec == nil {
				t.Skip("spec not loaded")
			}
			output := tt.spec.ValidateStructured()

			valid, ok := output["valid"].(bool)
			if !ok {
				t.Fatalf("output['valid'] is not bool, got %T", output["valid"])
			}

			errs, ok := output["errors"].([]map[string]any)
			if !ok {
				t.Fatalf("output['errors'] is not []map[string]any, got %T", output["errors"])
			}

			// PROP-5 invariant: valid == (len(errors) == 0)
			if valid && len(errs) != 0 {
				t.Errorf("valid is true but errors has %d entries — violates PROP-5", len(errs))
			}
			if !valid && len(errs) == 0 {
				t.Error("valid is false but errors is empty — violates PROP-5")
			}
		})
	}
}
