package afspec

import "fmt"

// ---------------------------------------------------------------------------
// Requirements-level collection mutation functions (05-REQ-1)
// ---------------------------------------------------------------------------

// deepCopyRequirements returns a deep copy of a RequirementsV1Json, ensuring
// that all slice and map fields are independently allocated. Callers can
// safely mutate the returned struct without affecting the original.
func deepCopyRequirements(req RequirementsV1Json) RequirementsV1Json {
	out := req // shallow copy of scalar/struct fields

	// Deep copy Requirements slice (and nested slices within each Requirement).
	out.Requirements = make([]Requirement, len(req.Requirements))
	for i, r := range req.Requirements {
		out.Requirements[i] = r
		out.Requirements[i].AcceptanceCriteria = make([]Criterion, len(r.AcceptanceCriteria))
		copy(out.Requirements[i].AcceptanceCriteria, r.AcceptanceCriteria)
		out.Requirements[i].EdgeCases = make([]Criterion, len(r.EdgeCases))
		copy(out.Requirements[i].EdgeCases, r.EdgeCases)
	}

	// Deep copy Glossary map.
	out.Glossary = make(RequirementsV1JsonGlossary, len(req.Glossary))
	for k, v := range req.Glossary {
		out.Glossary[k] = v
	}

	// Deep copy CorrectnessProperties slice (and nested Validates slices).
	out.CorrectnessProperties = make([]CorrectnessProperty, len(req.CorrectnessProperties))
	for i, cp := range req.CorrectnessProperties {
		out.CorrectnessProperties[i] = cp
		out.CorrectnessProperties[i].Validates = make([]string, len(cp.Validates))
		copy(out.CorrectnessProperties[i].Validates, cp.Validates)
	}

	// Deep copy ExecutionPaths slice (and nested Steps slices).
	out.ExecutionPaths = make([]ExecutionPath, len(req.ExecutionPaths))
	for i, ep := range req.ExecutionPaths {
		out.ExecutionPaths[i] = ep
		out.ExecutionPaths[i].Steps = make([]PathStep, len(ep.Steps))
		copy(out.ExecutionPaths[i].Steps, ep.Steps)
	}

	// Deep copy ErrorHandling slice.
	out.ErrorHandling = make([]ErrorHandlingEntry, len(req.ErrorHandling))
	copy(out.ErrorHandling, req.ErrorHandling)

	// Deep copy ExternalApis slice (and nested Symbols slices).
	if req.ExternalApis != nil {
		out.ExternalApis = make([]ExternalApi, len(req.ExternalApis))
		for i, api := range req.ExternalApis {
			out.ExternalApis[i] = api
			out.ExternalApis[i].Symbols = make([]ExternalApiSymbol, len(api.Symbols))
			copy(out.ExternalApis[i].Symbols, api.Symbols)
		}
	}

	return out
}

// AddRequirement returns a new RequirementsV1Json with the given Requirement
// appended. Returns an error if a Requirement with the same ID already exists.
// The original is not modified.
func AddRequirement(req RequirementsV1Json, r Requirement) (RequirementsV1Json, error) {
	for _, existing := range req.Requirements {
		if existing.Id == r.Id {
			return req, fmt.Errorf("duplicate requirement ID %q", r.Id)
		}
	}
	out := deepCopyRequirements(req)
	out.Requirements = append(out.Requirements, r)
	return out, nil
}

// GetRequirement looks up a Requirement by ID. Returns a pointer to the
// matching Requirement and true if found, or nil and false otherwise.
func GetRequirement(req RequirementsV1Json, id string) (*Requirement, bool) {
	for i := range req.Requirements {
		if req.Requirements[i].Id == id {
			return &req.Requirements[i], true
		}
	}
	return nil, false
}

// RemoveRequirement returns a new RequirementsV1Json with the Requirement
// having the given ID removed. Returns the new copy and true if found, or
// the original and false if the ID does not exist.
func RemoveRequirement(req RequirementsV1Json, id string) (RequirementsV1Json, bool) {
	idx := -1
	for i, r := range req.Requirements {
		if r.Id == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return req, false
	}
	out := deepCopyRequirements(req)
	out.Requirements = append(out.Requirements[:idx], out.Requirements[idx+1:]...)
	return out, true
}

// SetGlossaryEntry returns a new RequirementsV1Json with the glossary term
// set to the given definition. Inserts a new entry or overwrites an existing
// one. The original is not modified.
func SetGlossaryEntry(req RequirementsV1Json, term, definition string) RequirementsV1Json {
	out := deepCopyRequirements(req)
	if out.Glossary == nil {
		out.Glossary = make(RequirementsV1JsonGlossary)
	}
	out.Glossary[term] = definition
	return out
}

// RemoveGlossaryEntry returns a new RequirementsV1Json with the given term
// removed from the glossary. Returns the new copy and true if the term
// existed, or the original and false otherwise.
func RemoveGlossaryEntry(req RequirementsV1Json, term string) (RequirementsV1Json, bool) {
	if _, exists := req.Glossary[term]; !exists {
		return req, false
	}
	out := deepCopyRequirements(req)
	delete(out.Glossary, term)
	return out, true
}

// AddCorrectnessProperty returns a new RequirementsV1Json with the given
// CorrectnessProperty appended. Returns an error if a property with the same
// ID already exists. The original is not modified.
func AddCorrectnessProperty(req RequirementsV1Json, p CorrectnessProperty) (RequirementsV1Json, error) {
	for _, existing := range req.CorrectnessProperties {
		if existing.Id == p.Id {
			return req, fmt.Errorf("duplicate correctness property ID %q", p.Id)
		}
	}
	out := deepCopyRequirements(req)
	out.CorrectnessProperties = append(out.CorrectnessProperties, p)
	return out, nil
}

// AddExecutionPath returns a new RequirementsV1Json with the given
// ExecutionPath appended. Returns an error if a path with the same ID already
// exists. The original is not modified.
func AddExecutionPath(req RequirementsV1Json, p ExecutionPath) (RequirementsV1Json, error) {
	for _, existing := range req.ExecutionPaths {
		if existing.Id == p.Id {
			return req, fmt.Errorf("duplicate execution path ID %q", p.Id)
		}
	}
	out := deepCopyRequirements(req)
	out.ExecutionPaths = append(out.ExecutionPaths, p)
	return out, nil
}

// AddErrorHandling returns a new RequirementsV1Json with the given
// ErrorHandlingEntry appended. Returns an error if an entry with the same ID
// already exists. The original is not modified.
func AddErrorHandling(req RequirementsV1Json, e ErrorHandlingEntry) (RequirementsV1Json, error) {
	for _, existing := range req.ErrorHandling {
		if existing.Id == e.Id {
			return req, fmt.Errorf("duplicate error handling ID %q", e.Id)
		}
	}
	out := deepCopyRequirements(req)
	out.ErrorHandling = append(out.ErrorHandling, e)
	return out, nil
}

// ---------------------------------------------------------------------------
// Criterion-level mutation functions (05-REQ-2)
// ---------------------------------------------------------------------------

// AddCriterion returns a new Requirement with the given Criterion appended to
// acceptance_criteria. Returns an error if a Criterion with the same ID
// already exists. The original is not modified.
func AddCriterion(r Requirement, c Criterion) (Requirement, error) {
	panic("not implemented")
}

// AddEdgeCase returns a new Requirement with the given Criterion appended to
// edge_cases. Returns an error if a Criterion with the same ID already exists.
// The original is not modified.
func AddEdgeCase(r Requirement, c Criterion) (Requirement, error) {
	panic("not implemented")
}

// GetCriterion searches both acceptance_criteria and edge_cases for a Criterion
// with the given ID. Returns a pointer to the first match and true, or nil and
// false if not found in either slice.
func GetCriterion(r Requirement, id string) (*Criterion, bool) {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// TestSpec collection mutation functions (05-REQ-3)
// ---------------------------------------------------------------------------

// AddTestCase returns a new TestSpecV1Json with the given TestCase appended.
// Returns an error if a TestCase with the same ID already exists.
// The original is not modified.
func AddTestCase(ts TestSpecV1Json, tc TestCase) (TestSpecV1Json, error) {
	panic("not implemented")
}

// AddPropertyTest returns a new TestSpecV1Json with the given PropertyTest
// appended. Returns an error if a PropertyTest with the same ID already exists.
// The original is not modified.
func AddPropertyTest(ts TestSpecV1Json, pt PropertyTest) (TestSpecV1Json, error) {
	panic("not implemented")
}

// AddEdgeCaseTest returns a new TestSpecV1Json with the given EdgeCaseTest
// appended. Returns an error if an EdgeCaseTest with the same ID already
// exists. The original is not modified.
func AddEdgeCaseTest(ts TestSpecV1Json, et EdgeCaseTest) (TestSpecV1Json, error) {
	panic("not implemented")
}

// AddSmokeTest returns a new TestSpecV1Json with the given SmokeTest appended.
// Returns an error if a SmokeTest with the same ID already exists.
// The original is not modified.
func AddSmokeTest(ts TestSpecV1Json, st SmokeTest) (TestSpecV1Json, error) {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// Tasks collection mutation functions (05-REQ-4)
// ---------------------------------------------------------------------------

// AddTaskGroup returns a new TasksV1Json with the given TaskGroup appended.
// Returns an error if a TaskGroup with the same ID already exists.
// The original is not modified.
func AddTaskGroup(t TasksV1Json, g TaskGroup) (TasksV1Json, error) {
	panic("not implemented")
}

// AddSubtask returns a new TaskGroup with the given Subtask appended.
// Returns an error if a Subtask with the same ID already exists.
// The original is not modified.
func AddSubtask(g TaskGroup, s Subtask) (TaskGroup, error) {
	panic("not implemented")
}

// AddTraceabilityEntry returns a new TasksV1Json with the given
// TraceabilityEntry appended. Returns an error if an entry with the same
// (requirement_id, test_spec_id) pair already exists.
// The original is not modified.
func AddTraceabilityEntry(t TasksV1Json, e TraceabilityEntry) (TasksV1Json, error) {
	panic("not implemented")
}

// AddDependency returns a new TasksV1Json with the given TaskDependency
// appended unconditionally. The original is not modified.
func AddDependency(t TasksV1Json, d TaskDependency) TasksV1Json {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// LoadDependentInterfaces — cross-spec context loading (05-REQ-10)
// ---------------------------------------------------------------------------

// LoadDependentInterfaces loads interface summaries from upstream dependency
// specs for the given specID. It calls DiscoverSpecs, BuildDependencyGraph,
// and LoadSpec for each upstream dependency, extracting glossary entries,
// external API symbols, and criterion return contracts into a map[string]any
// per upstream spec.
//
// Returns an empty slice on any error (never returns an error value, never
// panics in production).
func LoadDependentInterfaces(specID string, specRoot string) []map[string]any {
	panic("not implemented")
}
