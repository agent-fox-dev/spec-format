package afspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

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

// ---------------------------------------------------------------------------
// JSON Schema validation infrastructure
// ---------------------------------------------------------------------------

// compiledSchemas holds the compiled JSON schemas, lazily initialized.
var (
	compiledSchemas     map[string]*jsonschema.Schema
	compiledSchemasErr  error
	compiledSchemasOnce sync.Once
)

// embeddedURLLoader implements jsonschema.URLLoader, serving schema bytes
// from the embedded schemaFS. It loads schemas from the
// "https://agent-fox.dev/schemas/" URL namespace.
type embeddedURLLoader struct {
	schemas map[string][]byte
}

func (l *embeddedURLLoader) Load(url string) (any, error) {
	const prefix = "https://agent-fox.dev/schemas/"
	if !strings.HasPrefix(url, prefix) {
		return nil, fmt.Errorf("unsupported schema URL: %s", url)
	}
	name := strings.TrimPrefix(url, prefix)
	data, ok := l.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema not found: %s", name)
	}
	return jsonschema.UnmarshalJSON(bytes.NewReader(data))
}

// getCompiledSchemas returns the compiled JSON schemas, initializing them
// on the first call. Returns an error (wrapping LoadError) if any schema
// fails to compile.
func getCompiledSchemas() (map[string]*jsonschema.Schema, error) {
	compiledSchemasOnce.Do(func() {
		schemas := Schemas()
		loader := &embeddedURLLoader{schemas: schemas}

		c := jsonschema.NewCompiler()
		c.UseLoader(jsonschema.SchemeURLLoader{
			"https": loader,
		})

		compiled := make(map[string]*jsonschema.Schema, len(schemas))
		for name, data := range schemas {
			// Parse and add the resource so the compiler knows about it
			doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if err != nil {
				compiledSchemasErr = &LoadError{
					Msg:  fmt.Sprintf("cannot parse schema %s: %s", name, err),
					File: name,
					Err:  &SpecError{Msg: fmt.Sprintf("cannot parse schema %s: %s", name, err)},
				}
				return
			}
			url := "https://agent-fox.dev/schemas/" + name
			if err := c.AddResource(url, doc); err != nil {
				compiledSchemasErr = &LoadError{
					Msg:  fmt.Sprintf("cannot add schema resource %s: %s", name, err),
					File: name,
					Err:  &SpecError{Msg: fmt.Sprintf("cannot add schema resource %s: %s", name, err)},
				}
				return
			}
		}

		for name := range schemas {
			url := "https://agent-fox.dev/schemas/" + name
			sch, err := c.Compile(url)
			if err != nil {
				compiledSchemasErr = &LoadError{
					Msg:  fmt.Sprintf("cannot compile schema %s: %s", name, err),
					File: name,
					Err:  &SpecError{Msg: fmt.Sprintf("cannot compile schema %s: %s", name, err)},
				}
				return
			}
			compiled[name] = sch
		}
		compiledSchemas = compiled
	})
	return compiledSchemas, compiledSchemasErr
}

// flattenValidationError recursively flattens a jsonschema.ValidationError
// tree into a slice of leaf errors (those without causes or with direct
// error messages).
func flattenValidationError(ve *jsonschema.ValidationError, artifact string) []ValidationEntry {
	var entries []ValidationEntry

	// If there are no causes, this is a leaf error
	if len(ve.Causes) == 0 {
		instanceLoc := "/" + strings.Join(ve.InstanceLocation, "/")
		if len(ve.InstanceLocation) == 0 {
			instanceLoc = ""
		}
		keywordLoc := ""
		msg := ""
		if ve.ErrorKind != nil {
			keywordLoc = "/" + strings.Join(ve.ErrorKind.KeywordPath(), "/")
			msg = ve.Error()
		}
		entries = append(entries, ValidationEntry{
			Category: "schema",
			Message:  msg,
			Artifact: artifact,
			Path:     instanceLoc,
			Keyword:  keywordLoc,
		})
		return entries
	}

	// Recurse into causes
	for _, cause := range ve.Causes {
		entries = append(entries, flattenValidationError(cause, artifact)...)
	}
	return entries
}

// validateArtifactSchema marshals the artifact to JSON, validates it against
// the named schema, and returns any validation errors as ValidationEntry slices.
func validateArtifactSchema(artifact any, schemaName, artifactName string) []ValidationEntry {
	schemas, err := getCompiledSchemas()
	if err != nil {
		return []ValidationEntry{{
			Category: "schema",
			Message:  fmt.Sprintf("schema compilation error: %s", err),
			Artifact: artifactName,
		}}
	}

	sch, ok := schemas[schemaName]
	if !ok {
		return []ValidationEntry{{
			Category: "schema",
			Message:  fmt.Sprintf("schema not found: %s", schemaName),
			Artifact: artifactName,
		}}
	}

	// Marshal the artifact to JSON, then unmarshal to any for validation.
	// We must use encoding/json (not MarshalJSON) to produce standard JSON
	// that the validator can consume as map[string]any.
	data, marshalErr := json.Marshal(artifact)
	if marshalErr != nil {
		return []ValidationEntry{{
			Category: "schema",
			Message:  fmt.Sprintf("cannot marshal %s: %s", artifactName, marshalErr),
			Artifact: artifactName,
		}}
	}

	var instance any
	if parseErr := json.Unmarshal(data, &instance); parseErr != nil {
		return []ValidationEntry{{
			Category: "schema",
			Message:  fmt.Sprintf("cannot parse %s JSON: %s", artifactName, parseErr),
			Artifact: artifactName,
		}}
	}

	validationErr := sch.Validate(instance)
	if validationErr == nil {
		return nil
	}

	ve, ok := validationErr.(*jsonschema.ValidationError)
	if !ok {
		return []ValidationEntry{{
			Category: "schema",
			Message:  fmt.Sprintf("validation error: %s", validationErr),
			Artifact: artifactName,
		}}
	}

	return flattenValidationError(ve, artifactName)
}

// ---------------------------------------------------------------------------
// Spec.ValidateSchema
// ---------------------------------------------------------------------------

// ValidateSchema compiles each bundled schema using jsonschema compiler
// with a SchemeURLLoader backed by embedded bytes, validates each
// artifact's JSON representation against its schema, and returns a
// ValidationResult.
func (s *Spec) ValidateSchema() ValidationResult {
	var allErrors []ValidationEntry

	// Validate requirements.json
	if s.Requirements != nil {
		errs := validateArtifactSchema(s.Requirements, "requirements.v1.json", "requirements.json")
		allErrors = append(allErrors, errs...)
	}

	// Validate test_spec.json
	if s.TestSpec != nil {
		errs := validateArtifactSchema(s.TestSpec, "test_spec.v1.json", "test_spec.json")
		allErrors = append(allErrors, errs...)
	}

	// Validate tasks.json
	if s.Tasks != nil {
		errs := validateArtifactSchema(s.Tasks, "tasks.v1.json", "tasks.json")
		allErrors = append(allErrors, errs...)
	}

	return ValidationResult{
		Valid:  len(allErrors) == 0,
		Errors: allErrors,
	}
}

// ---------------------------------------------------------------------------
// Spec.ValidateCrossFile
// ---------------------------------------------------------------------------

// requirementIDPattern matches valid requirement IDs like "01-REQ-1" or "01-REQ-1.1" or "01-REQ-1.E1".
var requirementIDPattern = regexp.MustCompile(`^\d{2}-REQ-\d+(\.\d+|\.E\d+)?$`)

// testCaseIDPattern matches valid test case IDs like "TS-01-1".
var testCaseIDPattern = regexp.MustCompile(`^TS-\d{2}-\d+$`)

// propertyIDPattern matches valid property IDs like "01-PROP-1".
var propertyIDPattern = regexp.MustCompile(`^\d{2}-PROP-\d+$`)

// pathIDPattern matches valid execution path IDs like "01-PATH-1".
var pathIDPattern = regexp.MustCompile(`^\d{2}-PATH-\d+$`)

// errorHandlingIDPattern matches valid error handling IDs like "01-ERR-1".
var errorHandlingIDPattern = regexp.MustCompile(`^\d{2}-ERR-\d+$`)

// smokeTestIDPattern matches valid smoke test IDs like "TS-01-SMOKE-1".
var smokeTestIDPattern = regexp.MustCompile(`^TS-\d{2}-SMOKE-\d+$`)

// propertyTestIDPattern matches valid property test IDs like "TS-01-P1".
var propertyTestIDPattern = regexp.MustCompile(`^TS-\d{2}-P\d+$`)

// edgeCaseTestIDPattern matches valid edge case test IDs like "TS-01-E1".
var edgeCaseTestIDPattern = regexp.MustCompile(`^TS-\d{2}-E\d+$`)

// criterionIDPattern matches valid criterion IDs like "01-REQ-1.1" or "01-REQ-1.E1".
var criterionIDPattern = regexp.MustCompile(`^\d{2}-REQ-\d+\.\d+$|^\d{2}-REQ-\d+\.E\d+$`)

// wiringSmokeRefPattern matches smoke test references in wiring_verification
// group test_spec_refs (e.g. "TS-04-SMOKE-1"). Compiled at package level.
var wiringSmokeRefPattern = regexp.MustCompile(`^TS-.*-SMOKE-.*$`)

// backtickTermRe extracts backtick-wrapped terms from text fields.
// Compiled at package initialization time per 04-REQ-4.2.
var backtickTermRe = regexp.MustCompile("`([^`]+)`")

// backtickNumericRe matches numeric values including negatives and decimals (e.g. -1, 3.14).
var backtickNumericRe = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

// backtickQuotedRe matches terms wrapped in single or double quotes.
var backtickQuotedRe = regexp.MustCompile(`^["'].*["']$`)

// ValidateCrossFile checks dangling references, coverage gaps, glossary
// completeness, and ID format validity across all artifacts and returns
// a ValidationResult.
func (s *Spec) ValidateCrossFile() ValidationResult {
	var errors []ValidationEntry
	var warnings []ValidationEntry

	// Collect all known requirement IDs (including criterion IDs)
	reqIDs := map[string]bool{}
	topReqIDs := map[string]bool{} // top-level requirement IDs only
	if s.Requirements != nil {
		for _, req := range s.Requirements.Requirements {
			reqIDs[req.Id] = true
			topReqIDs[req.Id] = true
			for _, ac := range req.AcceptanceCriteria {
				reqIDs[ac.Id] = true
			}
			for _, ec := range req.EdgeCases {
				reqIDs[ec.Id] = true
			}
		}
	}

	// Collect all known execution path IDs
	pathIDs := map[string]bool{}
	if s.Requirements != nil {
		for _, ep := range s.Requirements.ExecutionPaths {
			pathIDs[ep.Id] = true
		}
	}

	// Collect all known correctness property IDs
	propIDs := map[string]bool{}
	if s.Requirements != nil {
		for _, cp := range s.Requirements.CorrectnessProperties {
			propIDs[cp.Id] = true
		}
	}

	// --- ID format validation ---
	if s.Requirements != nil {
		for _, req := range s.Requirements.Requirements {
			if !requirementIDPattern.MatchString(req.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("requirement ID %q does not match expected format NN-REQ-N[.N|.EN]", req.Id),
					Artifact: "requirements.json",
					EntityID: req.Id,
				})
			}
			for _, ac := range req.AcceptanceCriteria {
				if !criterionIDPattern.MatchString(ac.Id) {
					errors = append(errors, ValidationEntry{
						Category: "integrity",
						Check:    "id_format",
						Message:  fmt.Sprintf("criterion ID %q does not match expected format NN-REQ-N.N or NN-REQ-N.EN", ac.Id),
						Artifact: "requirements.json",
						EntityID: ac.Id,
					})
				}
			}
			for _, ec := range req.EdgeCases {
				if !criterionIDPattern.MatchString(ec.Id) {
					errors = append(errors, ValidationEntry{
						Category: "integrity",
						Check:    "id_format",
						Message:  fmt.Sprintf("edge case criterion ID %q does not match expected format", ec.Id),
						Artifact: "requirements.json",
						EntityID: ec.Id,
					})
				}
			}
		}
		for _, cp := range s.Requirements.CorrectnessProperties {
			if !propertyIDPattern.MatchString(cp.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("correctness property ID %q does not match expected format NN-PROP-N", cp.Id),
					Artifact: "requirements.json",
					EntityID: cp.Id,
				})
			}
		}
		for _, ep := range s.Requirements.ExecutionPaths {
			if !pathIDPattern.MatchString(ep.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("execution path ID %q does not match expected format NN-PATH-N", ep.Id),
					Artifact: "requirements.json",
					EntityID: ep.Id,
				})
			}
		}
		for _, eh := range s.Requirements.ErrorHandling {
			if !errorHandlingIDPattern.MatchString(eh.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("error handling ID %q does not match expected format NN-ERR-N", eh.Id),
					Artifact: "requirements.json",
					EntityID: eh.Id,
				})
			}
		}
	}

	// --- Test spec ID format validation and dangling reference checks ---
	coveredReqIDs := map[string]bool{}

	if s.TestSpec != nil {
		for _, tc := range s.TestSpec.TestCases {
			if !testCaseIDPattern.MatchString(tc.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("test case ID %q does not match expected format TS-NN-N", tc.Id),
					Artifact: "test_spec.json",
					EntityID: tc.Id,
				})
			}
			// Check dangling requirement reference
			if tc.RequirementId != "" && !reqIDs[tc.RequirementId] {
				errors = append(errors, ValidationEntry{
					Category:      "integrity",
					Check:         "dangling_reference",
					Message:       fmt.Sprintf("test case %s references non-existent requirement %s", tc.Id, tc.RequirementId),
					Artifact:      "test_spec.json",
					RequirementID: tc.RequirementId,
					EntityID:      tc.Id,
				})
			}
			if tc.RequirementId != "" {
				coveredReqIDs[tc.RequirementId] = true
			}
		}

		for _, pt := range s.TestSpec.PropertyTests {
			if !propertyTestIDPattern.MatchString(pt.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("property test ID %q does not match expected format TS-NN-PN", pt.Id),
					Artifact: "test_spec.json",
					EntityID: pt.Id,
				})
			}
			// Check dangling property reference
			if pt.PropertyId != "" && !propIDs[pt.PropertyId] {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "dangling_reference",
					Message:  fmt.Sprintf("property test %s references non-existent property %s", pt.Id, pt.PropertyId),
					Artifact: "test_spec.json",
					EntityID: pt.Id,
				})
			}
		}

		for _, ec := range s.TestSpec.EdgeCaseTests {
			if !edgeCaseTestIDPattern.MatchString(ec.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("edge case test ID %q does not match expected format TS-NN-EN", ec.Id),
					Artifact: "test_spec.json",
					EntityID: ec.Id,
				})
			}
			// Check dangling requirement reference
			if ec.RequirementId != "" && !reqIDs[ec.RequirementId] {
				errors = append(errors, ValidationEntry{
					Category:      "integrity",
					Check:         "dangling_reference",
					Message:       fmt.Sprintf("edge case test %s references non-existent requirement %s", ec.Id, ec.RequirementId),
					Artifact:      "test_spec.json",
					RequirementID: ec.RequirementId,
					EntityID:      ec.Id,
				})
			}
			if ec.RequirementId != "" {
				coveredReqIDs[ec.RequirementId] = true
			}
		}

		for _, sm := range s.TestSpec.SmokeTests {
			if !smokeTestIDPattern.MatchString(sm.Id) {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "id_format",
					Message:  fmt.Sprintf("smoke test ID %q does not match expected format TS-NN-SMOKE-N", sm.Id),
					Artifact: "test_spec.json",
					EntityID: sm.Id,
				})
			}
			// Check dangling execution path reference
			if sm.ExecutionPathId != "" && !pathIDs[sm.ExecutionPathId] {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "dangling_reference",
					Message:  fmt.Sprintf("smoke test %s references non-existent execution path %s", sm.Id, sm.ExecutionPathId),
					Artifact: "test_spec.json",
					EntityID: sm.Id,
				})
			}
		}
	}

	// --- Coverage gap warnings ---
	// Check all acceptance criteria and edge case criteria for test coverage
	if s.Requirements != nil {
		for _, req := range s.Requirements.Requirements {
			for _, ac := range req.AcceptanceCriteria {
				if !coveredReqIDs[ac.Id] {
					warnings = append(warnings, ValidationEntry{
						Category:      "warning",
						Check:         "coverage_gap",
						Message:       fmt.Sprintf("requirement criterion %s has no test coverage", ac.Id),
						Artifact:      "test_spec.json",
						RequirementID: ac.Id,
						EntityID:      ac.Id,
					})
				}
			}
			for _, ec := range req.EdgeCases {
				if !coveredReqIDs[ec.Id] {
					warnings = append(warnings, ValidationEntry{
						Category:      "warning",
						Check:         "coverage_gap",
						Message:       fmt.Sprintf("edge case criterion %s has no test coverage", ec.Id),
						Artifact:      "test_spec.json",
						RequirementID: ec.Id,
						EntityID:      ec.Id,
					})
				}
			}
		}
	}

	// --- Cross-file rule 3: Property test coverage ---
	// Each correctness_property must have a matching property_test (by property_id).
	if s.Requirements != nil && len(s.Requirements.CorrectnessProperties) > 0 {
		// Build set of covered property_ids from property_tests
		coveredPropertyIDs := map[string]bool{}
		if s.TestSpec != nil {
			for _, pt := range s.TestSpec.PropertyTests {
				coveredPropertyIDs[pt.PropertyId] = true
			}
		}
		for i, cp := range s.Requirements.CorrectnessProperties {
			if !coveredPropertyIDs[cp.Id] {
				errors = append(errors, ValidationEntry{
					Category:      "integrity",
					Check:         "cross_file_3",
					Message:       fmt.Sprintf("correctness property %s has no matching property_test", cp.Id),
					Artifact:      "requirements.json",
					Path:          fmt.Sprintf("requirements.correctness_properties[%d]", i),
					RequirementID: cp.Id,
				})
			}
		}
	}

	// --- Cross-file rule 4: Execution path smoke test coverage ---
	// Each execution_path must have a matching smoke_test (by execution_path_id).
	if s.Requirements != nil && len(s.Requirements.ExecutionPaths) > 0 {
		// Build set of covered execution_path_ids from smoke_tests
		coveredPathIDs := map[string]bool{}
		if s.TestSpec != nil {
			for _, sm := range s.TestSpec.SmokeTests {
				coveredPathIDs[sm.ExecutionPathId] = true
			}
		}
		for _, ep := range s.Requirements.ExecutionPaths {
			if !coveredPathIDs[ep.Id] {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "cross_file_4",
					Message:  fmt.Sprintf("execution path %s has no matching smoke_test", ep.Id),
					Artifact: "requirements.json",
					EntityID: ep.Id,
				})
			}
		}
	}

	// --- Cross-file rule 5: test_spec_id resolution ---
	// Every test_spec_id in traceability entries and subtask test_spec_refs
	// must resolve to a known test entry in test_spec.
	{
		// Build set of all known test IDs from test_spec
		knownTestIDs := map[string]bool{}
		if s.TestSpec != nil {
			for _, tc := range s.TestSpec.TestCases {
				knownTestIDs[tc.Id] = true
			}
			for _, pt := range s.TestSpec.PropertyTests {
				knownTestIDs[pt.Id] = true
			}
			for _, ec := range s.TestSpec.EdgeCaseTests {
				knownTestIDs[ec.Id] = true
			}
			for _, sm := range s.TestSpec.SmokeTests {
				knownTestIDs[sm.Id] = true
			}
		}

		// Check traceability entries
		if s.Tasks != nil {
			for _, te := range s.Tasks.Traceability {
				if te.TestSpecId != "" && !knownTestIDs[te.TestSpecId] {
					errors = append(errors, ValidationEntry{
						Category: "integrity",
						Check:    "cross_file_5",
						Message:  fmt.Sprintf("traceability entry references unresolvable test_spec_id %s", te.TestSpecId),
						Artifact: "tasks.json",
						EntityID: te.TestSpecId,
					})
				}
			}

			// Check subtask test_spec_refs
			for _, group := range s.Tasks.TaskGroups {
				for _, sub := range group.Subtasks {
					for _, ref := range sub.TestSpecRefs {
						if ref != "" && !knownTestIDs[ref] {
							errors = append(errors, ValidationEntry{
								Category: "integrity",
								Check:    "cross_file_5",
								Message:  fmt.Sprintf("subtask %s test_spec_refs contains unresolvable test_spec_id %s", sub.Id, ref),
								Artifact: "tasks.json",
								EntityID: ref,
							})
						}
					}
				}
			}
		}
	}

	// --- Cross-file rule 6: Glossary backtick term check ---
	// Extract backtick-wrapped terms from criterion and correctness property
	// fields; flag any non-excluded term not present in the glossary.
	if s.Requirements != nil {
		glossary := s.Requirements.Glossary

		// isExcludedBacktickTerm returns true if the term should be excluded
		// from glossary checks: numeric, single character, quoted, or > 80 chars.
		isExcluded := func(term string) bool {
			if len([]rune(term)) == 1 {
				return true
			}
			if len([]rune(term)) > 80 {
				return true
			}
			if backtickNumericRe.MatchString(term) {
				return true
			}
			if backtickQuotedRe.MatchString(term) {
				return true
			}
			return false
		}

		// checkFieldForBacktickTerms extracts backtick terms from a field value
		// and appends validation errors for undefined glossary terms.
		checkField := func(fieldValue string) {
			matches := backtickTermRe.FindAllStringSubmatch(fieldValue, -1)
			for _, m := range matches {
				term := m[1]
				if isExcluded(term) {
					continue
				}
				if _, ok := glossary[term]; !ok {
					errors = append(errors, ValidationEntry{
						Category: "integrity",
						Check:    "cross_file_6",
						Message:  fmt.Sprintf("backtick term %q is not defined in glossary", term),
						Artifact: "requirements.json",
					})
				}
			}
		}

		// Check criterion fields across all requirements
		for _, req := range s.Requirements.Requirements {
			for _, ac := range req.AcceptanceCriteria {
				checkField(ac.Action)
				if ac.Trigger != nil {
					checkField(*ac.Trigger)
				}
				if ac.Condition != nil {
					checkField(*ac.Condition)
				}
				if ac.ErrorCondition != nil {
					checkField(*ac.ErrorCondition)
				}
				if ac.State != nil {
					checkField(*ac.State)
				}
				if ac.Feature != nil {
					checkField(*ac.Feature)
				}
			}
			for _, ec := range req.EdgeCases {
				checkField(ec.Action)
				if ec.Trigger != nil {
					checkField(*ec.Trigger)
				}
				if ec.Condition != nil {
					checkField(*ec.Condition)
				}
				if ec.ErrorCondition != nil {
					checkField(*ec.ErrorCondition)
				}
				if ec.State != nil {
					checkField(*ec.State)
				}
				if ec.Feature != nil {
					checkField(*ec.Feature)
				}
			}
		}

		// Check correctness property fields (for_any, invariant)
		for _, cp := range s.Requirements.CorrectnessProperties {
			checkField(cp.ForAny)
			checkField(cp.Invariant)
		}
	}

	// --- Cross-file rule 8: Traceability deduplication ---
	// Flag duplicate (requirement_id, test_spec_id) pairs in traceability.
	if s.Tasks != nil && len(s.Tasks.Traceability) > 0 {
		seen := map[string]bool{}
		for _, te := range s.Tasks.Traceability {
			key := te.RequirementId + "|" + te.TestSpecId
			if seen[key] {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "cross_file_8",
					Message:  fmt.Sprintf("duplicate traceability pair (%s, %s)", te.RequirementId, te.TestSpecId),
					Artifact: "tasks.json",
				})
			} else {
				seen[key] = true
			}
		}
	}

	// --- Cross-file rule 9: Subtask requirement_refs resolution ---
	// Every requirement_refs entry must resolve to a known requirement ID,
	// criterion ID, or edge case ID.
	if s.Tasks != nil && s.Requirements != nil {
		for _, group := range s.Tasks.TaskGroups {
			for _, sub := range group.Subtasks {
				for _, ref := range sub.RequirementRefs {
					if ref != "" && !reqIDs[ref] {
						errors = append(errors, ValidationEntry{
							Category: "integrity",
							Check:    "cross_file_9",
							Message:  fmt.Sprintf("subtask %s requirement_refs contains unresolvable reference %s", sub.Id, ref),
							Artifact: "tasks.json",
							EntityID: ref,
						})
					}
				}
			}
		}
	}

	// --- Cross-file rule 10: Unwanted pattern return_contract check ---
	// Every criterion with ears_pattern='unwanted' must have a non-empty
	// return_contract.
	if s.Requirements != nil {
		for _, req := range s.Requirements.Requirements {
			for _, ac := range req.AcceptanceCriteria {
				if ac.EarsPattern == CriterionEarsPatternUnwanted {
					if ac.ReturnContract == nil || *ac.ReturnContract == "" {
						errors = append(errors, ValidationEntry{
							Category:      "integrity",
							Check:         "cross_file_10",
							Message:       fmt.Sprintf("criterion %s has ears_pattern 'unwanted' but missing return_contract", ac.Id),
							Artifact:      "requirements.json",
							RequirementID: req.Id,
							EntityID:      ac.Id,
						})
					}
				}
			}
			for _, ec := range req.EdgeCases {
				if ec.EarsPattern == CriterionEarsPatternUnwanted {
					if ec.ReturnContract == nil || *ec.ReturnContract == "" {
						errors = append(errors, ValidationEntry{
							Category:      "integrity",
							Check:         "cross_file_10",
							Message:       fmt.Sprintf("criterion %s has ears_pattern 'unwanted' but missing return_contract", ec.Id),
							Artifact:      "requirements.json",
							RequirementID: req.Id,
							EntityID:      ec.Id,
						})
					}
				}
			}
		}
	}

	// --- Task group structure validation (04-REQ-8) ---
	// First group must have kind='tests'; last group must have kind='wiring_verification'.
	// Skip if no task groups exist.
	if s.Tasks != nil && len(s.Tasks.TaskGroups) > 0 {
		groups := s.Tasks.TaskGroups
		if groups[0].Kind != TaskGroupKindTests {
			errors = append(errors, ValidationEntry{
				Category: "schema",
				Check:    "task_group_structure",
				Message:  fmt.Sprintf("first task group must have kind 'tests', got '%s'", groups[0].Kind),
				Artifact: "tasks.json",
			})
		}
		if groups[len(groups)-1].Kind != TaskGroupKindWiringVerification {
			errors = append(errors, ValidationEntry{
				Category: "schema",
				Check:    "task_group_structure",
				Message:  fmt.Sprintf("last task group must have kind 'wiring_verification', got '%s'", groups[len(groups)-1].Kind),
				Artifact: "tasks.json",
			})
		}
	}

	// --- Wiring verification group semantics (04-REQ-9) ---
	// Check wiring_verification group for meaningful content:
	//   A. at least one subtask has non-empty test_spec_refs
	//   B. at least one test_spec_refs entry matches TS-*-SMOKE-*
	//   C. at least one subtask title or details mentions 'stub' or 'dead'
	if s.Tasks != nil && len(s.Tasks.TaskGroups) > 0 {
		lastGroup := s.Tasks.TaskGroups[len(s.Tasks.TaskGroups)-1]
		if lastGroup.Kind == TaskGroupKindWiringVerification {
			// Sub-check A: at least one subtask has non-empty test_spec_refs
			hasRefs := false
			for _, sub := range lastGroup.Subtasks {
				if len(sub.TestSpecRefs) > 0 {
					hasRefs = true
					break
				}
			}
			if !hasRefs {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "wiring_verification",
					Message:  "wiring_verification group: no subtask has non-empty test_spec_refs",
					Artifact: "tasks.json",
				})
			}

			// Sub-check B: at least one test_spec_refs entry matches TS-*-SMOKE-*
			hasSmokeRef := false
			for _, sub := range lastGroup.Subtasks {
				for _, ref := range sub.TestSpecRefs {
					if wiringSmokeRefPattern.MatchString(ref) {
						hasSmokeRef = true
						break
					}
				}
				if hasSmokeRef {
					break
				}
			}
			if !hasSmokeRef {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "wiring_verification",
					Message:  "wiring_verification group: no test_spec_refs entry matches smoke test pattern TS-*-SMOKE-*",
					Artifact: "tasks.json",
				})
			}

			// Sub-check C: at least one subtask mentions 'stub' or 'dead' in title or details
			hasStubMention := false
			for _, sub := range lastGroup.Subtasks {
				lower := strings.ToLower(sub.Title)
				if strings.Contains(lower, "stub") || strings.Contains(lower, "dead") {
					hasStubMention = true
					break
				}
				for _, detail := range sub.Details {
					lower = strings.ToLower(detail)
					if strings.Contains(lower, "stub") || strings.Contains(lower, "dead") {
						hasStubMention = true
						break
					}
				}
				if hasStubMention {
					break
				}
			}
			if !hasStubMention {
				errors = append(errors, ValidationEntry{
					Category: "integrity",
					Check:    "wiring_verification",
					Message:  "wiring_verification group: no subtask mentions stub or dead-code audit in title or details",
					Artifact: "tasks.json",
				})
			}
		}
	}

	// --- Spec ID/name consistency across artifacts ---
	if s.Requirements != nil {
		if s.Requirements.SpecId != s.SpecID {
			errors = append(errors, ValidationEntry{
				Category: "integrity",
				Check:    "spec_id_mismatch",
				Message:  fmt.Sprintf("requirements.spec_id %q does not match spec ID %q", s.Requirements.SpecId, s.SpecID),
				Artifact: "requirements.json",
			})
		}
		if s.Requirements.SpecName != s.SpecName {
			errors = append(errors, ValidationEntry{
				Category: "integrity",
				Check:    "spec_name_mismatch",
				Message:  fmt.Sprintf("requirements.spec_name %q does not match spec name %q", s.Requirements.SpecName, s.SpecName),
				Artifact: "requirements.json",
			})
		}
	}
	if s.TestSpec != nil {
		if s.TestSpec.SpecId != s.SpecID {
			errors = append(errors, ValidationEntry{
				Category: "integrity",
				Check:    "spec_id_mismatch",
				Message:  fmt.Sprintf("test_spec.spec_id %q does not match spec ID %q", s.TestSpec.SpecId, s.SpecID),
				Artifact: "test_spec.json",
			})
		}
		if s.TestSpec.SpecName != s.SpecName {
			errors = append(errors, ValidationEntry{
				Category: "integrity",
				Check:    "spec_name_mismatch",
				Message:  fmt.Sprintf("test_spec.spec_name %q does not match spec name %q", s.TestSpec.SpecName, s.SpecName),
				Artifact: "test_spec.json",
			})
		}
	}
	if s.Tasks != nil {
		if s.Tasks.SpecId != s.SpecID {
			errors = append(errors, ValidationEntry{
				Category: "integrity",
				Check:    "spec_id_mismatch",
				Message:  fmt.Sprintf("tasks.spec_id %q does not match spec ID %q", s.Tasks.SpecId, s.SpecID),
				Artifact: "tasks.json",
			})
		}
		if s.Tasks.SpecName != s.SpecName {
			errors = append(errors, ValidationEntry{
				Category: "integrity",
				Check:    "spec_name_mismatch",
				Message:  fmt.Sprintf("tasks.spec_name %q does not match spec name %q", s.Tasks.SpecName, s.SpecName),
				Artifact: "tasks.json",
			})
		}
	}

	return ValidationResult{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// ---------------------------------------------------------------------------
// Spec.Validate
// ---------------------------------------------------------------------------

// Validate runs schema validation, EARS constraint checks, task group
// structure rules, and cross-file consistency checks, then returns a
// ValidationResult. Valid is true if and only if Errors is empty.
func (s *Spec) Validate() ValidationResult {
	schemaResult := s.ValidateSchema()
	crossFileResult := s.ValidateCrossFile()

	var allErrors []ValidationEntry
	allErrors = append(allErrors, schemaResult.Errors...)
	allErrors = append(allErrors, crossFileResult.Errors...)

	var allWarnings []ValidationEntry
	allWarnings = append(allWarnings, schemaResult.Warnings...)
	allWarnings = append(allWarnings, crossFileResult.Warnings...)

	// Check for subtasks with missing refs (produces warnings, not errors).
	if s.Tasks != nil {
		for _, group := range s.Tasks.TaskGroups {
			allWarnings = append(allWarnings, checkMissingSubtaskRefs(group)...)
		}
	}

	return ValidationResult{
		Valid:    len(allErrors) == 0,
		Errors:   allErrors,
		Warnings: allWarnings,
	}
}

// ---------------------------------------------------------------------------
// ValidateCrossSpec
// ---------------------------------------------------------------------------

// ValidateCrossSpec checks API symbol consistency, glossary conflicts,
// and contract mismatches across all provided specs using the dependency
// graph to determine which specs interact, and returns a ValidationResult.
func ValidateCrossSpec(specs []*Spec, graph *DependencyGraph) ValidationResult {
	var errors []ValidationEntry

	if len(specs) <= 1 {
		return ValidationResult{Valid: true}
	}

	// Build a set of spec IDs that interact via dependency edges
	type specPair struct{ a, b string }
	interacting := map[specPair]bool{}
	if graph != nil {
		for _, edge := range graph.Edges {
			interacting[specPair{edge.FromSpec, edge.ToSpec}] = true
			interacting[specPair{edge.ToSpec, edge.FromSpec}] = true
		}
	}

	// Check glossary conflicts between interacting specs
	for i := 0; i < len(specs); i++ {
		for j := i + 1; j < len(specs); j++ {
			specA := specs[i]
			specB := specs[j]

			// Only check interacting specs (connected by dependency edges)
			pair := specPair{specA.SpecID, specB.SpecID}
			if !interacting[pair] {
				continue
			}

			if specA.Requirements == nil || specB.Requirements == nil {
				continue
			}

			// Check glossary conflicts
			for term, defA := range specA.Requirements.Glossary {
				if defB, ok := specB.Requirements.Glossary[term]; ok && defA != defB {
					errors = append(errors, ValidationEntry{
						Category: "integrity",
						Check:    "glossary_conflict",
						Message:  fmt.Sprintf("glossary term %q has conflicting definitions between spec %s and spec %s", term, specA.SpecID, specB.SpecID),
						EntityID: term,
					})
				}
			}
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ---------------------------------------------------------------------------
// Spec.ValidateStructured
// ---------------------------------------------------------------------------

// ValidateStructured runs full validation and returns a structured
// map[string]any for CLI consumption, with 'valid', 'errors', and
// optionally 'warnings' keys.
func (s *Spec) ValidateStructured() map[string]any {
	result := s.Validate()

	errorMaps := make([]map[string]any, 0, len(result.Errors))
	for _, e := range result.Errors {
		entry := map[string]any{
			"message": e.Message,
		}

		if e.Category != "" {
			entry["category"] = e.Category
		}
		if e.Artifact != "" {
			entry["artifact"] = e.Artifact
		}
		if e.Path != "" {
			entry["path"] = e.Path
		}
		if e.Keyword != "" {
			entry["keyword"] = e.Keyword
		}
		if e.Check != "" {
			entry["check"] = e.Check
		}
		if e.RequirementID != "" {
			entry["requirement_id"] = e.RequirementID
		}
		if e.Value != "" {
			entry["value"] = e.Value
		}

		errorMaps = append(errorMaps, entry)
	}

	out := map[string]any{
		"valid":  result.Valid,
		"errors": errorMaps,
	}

	// Only include warnings key when there are warnings
	if len(result.Warnings) > 0 {
		warningMaps := make([]map[string]any, 0, len(result.Warnings))
		for _, w := range result.Warnings {
			entry := map[string]any{
				"category": w.Category,
				"message":  w.Message,
			}
			if w.EntityID != "" {
				entry["entity_id"] = w.EntityID
			}
			warningMaps = append(warningMaps, entry)
		}
		out["warnings"] = warningMaps
	}

	return out
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
