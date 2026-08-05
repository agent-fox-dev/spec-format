package afspec

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// 4.5: Schemas, ComputeIntentHash, ComputeCoverage, CreateSpec
// ---------------------------------------------------------------------------

// TestSchemas verifies that Schemas returns a non-nil map of schema file
// names to their raw bytes.
// Test Spec: TS-01-46, Requirement: 01-REQ-22.1
func TestSchemas(t *testing.T) {
	defer requireImplemented(t)

	schemas := Schemas()

	if schemas == nil {
		t.Fatal("Schemas returned nil map")
	}
	if len(schemas) == 0 {
		t.Fatal("Schemas returned empty map")
	}

	// Verify expected schema files are present
	expectedSchemas := []string{
		"requirements.v1.json",
		"test_spec.v1.json",
		"tasks.v1.json",
		"prd-frontmatter.v1.json",
	}
	for _, name := range expectedSchemas {
		val, ok := schemas[name]
		if !ok {
			t.Errorf("expected schema %q in map, but it was missing", name)
			continue
		}
		if len(val) == 0 {
			t.Errorf("expected non-empty bytes for schema %q", name)
		}
	}
}

// TestSchemas_NonNil verifies that Schemas always returns a non-nil, non-empty
// map because schemas are embedded at compile time.
// Requirement: 01-REQ-22.E1
func TestSchemas_NonNil(t *testing.T) {
	defer requireImplemented(t)

	schemas := Schemas()
	if schemas == nil {
		t.Error("Schemas must never return nil")
	}
	if len(schemas) == 0 {
		t.Error("Schemas must never return empty map")
	}

	// Verify all values are non-empty
	for key, val := range schemas {
		if key == "" {
			t.Error("schema map contains empty key")
		}
		if len(val) == 0 {
			t.Errorf("schema %q has empty bytes", key)
		}
	}
}

// TestComputeIntentHash verifies that ComputeIntentHash extracts the ## Intent
// section, computes its SHA-256 hash, and returns a 64-char hex string.
// Test Spec: TS-01-47, Requirement: 01-REQ-23.1
func TestComputeIntentHash(t *testing.T) {
	defer requireImplemented(t)

	prdBody := "# Title\n\n## Intent\n\nThis is the intent.\n\n## Goals\n\nSome goals.\n"

	hash, err := ComputeIntentHash(prdBody)
	if err != nil {
		t.Fatalf("ComputeIntentHash returned error: %v", err)
	}
	if len(hash) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars: %q", len(hash), hash)
	}

	// Verify it's a valid hex string
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("expected lowercase hex characters, got %q in hash %q", string(c), hash)
			break
		}
	}

	// Verify determinism: same input produces same hash
	hash2, err := ComputeIntentHash(prdBody)
	if err != nil {
		t.Fatalf("second ComputeIntentHash returned error: %v", err)
	}
	if hash != hash2 {
		t.Errorf("hash is not deterministic: %q != %q", hash, hash2)
	}
}

// TestComputeIntentHash_MissingSection verifies that ComputeIntentHash
// returns an IntentError when the body has no ## Intent section.
// Requirement: 01-REQ-23.E1
func TestComputeIntentHash_MissingSection(t *testing.T) {
	defer requireImplemented(t)

	prdBody := "# Title\n\n## Goals\n\nSome goals.\n"

	hash, err := ComputeIntentHash(prdBody)
	if err == nil {
		t.Fatal("expected error when ## Intent section is missing, got nil")
	}
	if hash != "" {
		t.Errorf("expected empty hash on error, got %q", hash)
	}

	var intentErr *IntentError
	if !errors.As(err, &intentErr) {
		t.Errorf("expected IntentError, got %T", err)
	}

	var specErr *SpecError
	if !errors.As(err, &specErr) {
		t.Errorf("expected IntentError to wrap SpecError, got %T", err)
	}
}

// TestComputeIntentHash_EmptyBody verifies that ComputeIntentHash returns
// an IntentError for an empty body string.
// Requirement: 01-REQ-23.E2
func TestComputeIntentHash_EmptyBody(t *testing.T) {
	defer requireImplemented(t)

	hash, err := ComputeIntentHash("")
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
	if hash != "" {
		t.Errorf("expected empty hash on error, got %q", hash)
	}

	var intentErr *IntentError
	if !errors.As(err, &intentErr) {
		t.Errorf("expected IntentError, got %T", err)
	}
}

// TestComputeCoverage verifies that ComputeCoverage scans test entries
// against requirement IDs, property IDs, and path IDs and returns a
// correct coverage report.
// Test Spec: TS-01-48, Requirement: 01-REQ-24.1
func TestComputeCoverage(t *testing.T) {
	defer requireImplemented(t)

	req := &RequirementsV1Json{
		SchemaVersion: 1,
		SpecId:        "01",
		SpecName:      "test",
		Introduction:  "Test",
		Glossary:      RequirementsV1JsonGlossary{},
		Requirements: []Requirement{
			{
				Id:    "01-REQ-1",
				Title: "Req 1",
				UserStory: UserStory{
					Role: "dev", Goal: "goal", Benefit: "benefit",
				},
				AcceptanceCriteria: []Criterion{
					{
						Id:          "01-REQ-1.1",
						EarsPattern: CriterionEarsPatternUbiquitous,
						System:      "sys",
						Action:      "act",
					},
				},
				EdgeCases: []Criterion{},
			},
			{
				Id:    "01-REQ-2",
				Title: "Req 2",
				UserStory: UserStory{
					Role: "dev", Goal: "goal2", Benefit: "benefit2",
				},
				AcceptanceCriteria: []Criterion{
					{
						Id:          "01-REQ-2.1",
						EarsPattern: CriterionEarsPatternUbiquitous,
						System:      "sys",
						Action:      "act2",
					},
				},
				EdgeCases: []Criterion{},
			},
		},
		CorrectnessProperties: []CorrectnessProperty{
			{
				Id:        "01-PROP-1",
				Title:     "Property 1",
				ForAny:    "any value",
				Invariant: "invariant holds",
				Validates: []string{"01-REQ-1.1"},
			},
		},
		ExecutionPaths: []ExecutionPath{
			{
				Id:    "01-PATH-1",
				Title: "Path 1",
				Steps: []PathStep{
					{Actor: "caller", Action: "calls"},
					{Actor: "system", Action: "responds"},
				},
			},
		},
		ErrorHandling: []ErrorHandlingEntry{},
	}

	// TestSpec covers 01-REQ-1 and 01-PROP-1 but NOT 01-REQ-2 or 01-PATH-1
	ts := &TestSpecV1Json{
		SchemaVersion: 1,
		SpecId:        "01",
		SpecName:      "test",
		TestCases: []TestCase{
			{
				Id:                  "TS-01-1",
				RequirementId:       "01-REQ-1.1",
				Kind:                TestCaseKindUnit,
				Description:         "Covers REQ-1",
				Preconditions:       []string{},
				Expected:            "result",
				AssertionPseudocode: "assert true",
			},
		},
		PropertyTests: []PropertyTest{
			{
				Id:             "TS-01-P1",
				PropertyId:     "01-PROP-1",
				Validates:      []string{"01-REQ-1.1"},
				Description:    "Property test",
				ForAnyStrategy: "strategy",
				InvariantCheck: "check",
			},
		},
		EdgeCaseTests: []EdgeCaseTest{},
		SmokeTests:    []SmokeTest{},
		Coverage: Coverage{
			RequirementsCovered: []string{"01-REQ-1.1"},
			PropertiesCovered:   []string{"01-PROP-1"},
			PathsCovered:        []string{},
		},
	}

	report := ts.ComputeCoverage(req)

	// 01-REQ-1 and 01-PROP-1 should be covered
	assertInSlice(t, report.Covered, "01-REQ-1", "covered requirements")
	assertInSlice(t, report.Covered, "01-PROP-1", "covered properties")

	// 01-REQ-2 and 01-PATH-1 should be uncovered
	assertInSlice(t, report.Uncovered, "01-REQ-2", "uncovered requirements")
	assertInSlice(t, report.Uncovered, "01-PATH-1", "uncovered paths")
}

// TestComputeCoverage_EmptyRequirements verifies that ComputeCoverage with
// an empty Requirements struct returns 100% coverage.
// Requirement: 01-REQ-24.E1
func TestComputeCoverage_EmptyRequirements(t *testing.T) {
	defer requireImplemented(t)

	req := &RequirementsV1Json{
		SchemaVersion:         1,
		SpecId:                "01",
		SpecName:              "test",
		Introduction:          "Test",
		Glossary:              RequirementsV1JsonGlossary{},
		Requirements:          []Requirement{},
		CorrectnessProperties: []CorrectnessProperty{},
		ExecutionPaths:        []ExecutionPath{},
		ErrorHandling:         []ErrorHandlingEntry{},
	}

	ts := &TestSpecV1Json{
		SchemaVersion: 1,
		SpecId:        "01",
		SpecName:      "test",
		TestCases:     []TestCase{},
		PropertyTests: []PropertyTest{},
		EdgeCaseTests: []EdgeCaseTest{},
		SmokeTests:    []SmokeTest{},
		Coverage:      Coverage{},
	}

	report := ts.ComputeCoverage(req)

	if len(report.Uncovered) != 0 {
		t.Errorf("expected empty uncovered list for empty requirements, got %v", report.Uncovered)
	}
}

// TestComputeCoverage_EmptyTestSpec verifies that ComputeCoverage with no
// test entries returns 0% coverage.
// Requirement: 01-REQ-24.E2
func TestComputeCoverage_EmptyTestSpec(t *testing.T) {
	defer requireImplemented(t)

	req := &RequirementsV1Json{
		SchemaVersion: 1,
		SpecId:        "01",
		SpecName:      "test",
		Introduction:  "Test",
		Glossary:      RequirementsV1JsonGlossary{},
		Requirements: []Requirement{
			{
				Id:    "01-REQ-1",
				Title: "Req 1",
				UserStory: UserStory{
					Role: "dev", Goal: "goal", Benefit: "benefit",
				},
				AcceptanceCriteria: []Criterion{
					{
						Id:          "01-REQ-1.1",
						EarsPattern: CriterionEarsPatternUbiquitous,
						System:      "sys",
						Action:      "act",
					},
				},
				EdgeCases: []Criterion{},
			},
		},
		CorrectnessProperties: []CorrectnessProperty{},
		ExecutionPaths:        []ExecutionPath{},
		ErrorHandling:         []ErrorHandlingEntry{},
	}

	ts := &TestSpecV1Json{
		SchemaVersion: 1,
		SpecId:        "01",
		SpecName:      "test",
		TestCases:     []TestCase{},
		PropertyTests: []PropertyTest{},
		EdgeCaseTests: []EdgeCaseTest{},
		SmokeTests:    []SmokeTest{},
		Coverage:      Coverage{},
	}

	report := ts.ComputeCoverage(req)

	if len(report.Uncovered) == 0 {
		t.Error("expected non-empty uncovered list when TestSpec has no entries")
	}
	assertInSlice(t, report.Uncovered, "01-REQ-1", "uncovered requirement")
}

// TestCreateSpec verifies that CreateSpec returns a Spec with status "draft",
// all sub-artifacts initialized to non-nil values, and specID/specName set.
// Test Spec: TS-01-49, Requirement: 01-REQ-25.1
func TestCreateSpec(t *testing.T) {
	defer requireImplemented(t)

	spec := CreateSpec("01", "my_spec")

	if spec == nil {
		t.Fatal("CreateSpec returned nil")
	}
	if spec.Status != "draft" {
		t.Errorf("Status = %q, want %q", spec.Status, "draft")
	}
	if spec.SpecID != "01" {
		t.Errorf("SpecID = %q, want %q", spec.SpecID, "01")
	}
	if spec.SpecName != "my_spec" {
		t.Errorf("SpecName = %q, want %q", spec.SpecName, "my_spec")
	}
	if spec.Requirements == nil {
		t.Error("expected Requirements to be non-nil (initialized)")
	}
	if spec.TestSpec == nil {
		t.Error("expected TestSpec to be non-nil (initialized)")
	}
	if spec.Tasks == nil {
		t.Error("expected Tasks to be non-nil (initialized)")
	}
}

// TestCreateSpec_EmptyArguments verifies that CreateSpec with empty specID
// or specName returns a valid Spec; validation is deferred.
// Requirement: 01-REQ-25.E1
func TestCreateSpec_EmptyArguments(t *testing.T) {
	defer requireImplemented(t)

	spec := CreateSpec("", "")

	if spec == nil {
		t.Fatal("CreateSpec returned nil for empty arguments")
	}
	if spec.Status != "draft" {
		t.Errorf("Status = %q, want %q", spec.Status, "draft")
	}
	if spec.SpecID != "" {
		t.Errorf("SpecID = %q, want empty string", spec.SpecID)
	}
	if spec.SpecName != "" {
		t.Errorf("SpecName = %q, want empty string", spec.SpecName)
	}
}

// assertInSlice is a test helper that checks if a string is present in a slice.
func assertInSlice(t *testing.T, slice []string, item, label string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			return
		}
	}
	t.Errorf("%s: expected %q in slice %v", label, item, slice)
}
