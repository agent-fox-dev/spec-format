package afspec

// CoverageReport represents the result of computing test coverage against
// requirements. Covered contains IDs that are referenced by at least one
// test entry; Uncovered contains IDs with no test coverage.
type CoverageReport struct {
	Covered   []string
	Uncovered []string
}

// ComputeCoverage scans all test entries in the TestSpec against requirement IDs,
// correctness property IDs, and execution path IDs in the Requirements struct.
// Returns a CoverageReport indicating which entities are covered and which
// are not.
//
// Coverage is computed at three levels:
//   - Requirements: a requirement is covered if any of its acceptance criteria
//     or edge case criteria is referenced by a test case or edge case test.
//   - Properties: a property is covered if it's referenced by a property test.
//   - Paths: an execution path is covered if it's referenced by a smoke test.
//
// If the Requirements struct has no entities to cover, the report indicates
// 100% coverage (empty Uncovered list). If the TestSpec has no test entries,
// all requirement entities appear in the Uncovered list.
func (ts *TestSpecV1Json) ComputeCoverage(req *RequirementsV1Json) CoverageReport {
	// Build mapping from criterion IDs to their parent requirement ID.
	criterionToReq := make(map[string]string)
	reqIDs := make(map[string]bool)
	for _, r := range req.Requirements {
		reqIDs[r.Id] = true
		for _, ac := range r.AcceptanceCriteria {
			criterionToReq[ac.Id] = r.Id
		}
		for _, ec := range r.EdgeCases {
			criterionToReq[ec.Id] = r.Id
		}
	}

	propIDs := make(map[string]bool)
	for _, p := range req.CorrectnessProperties {
		propIDs[p.Id] = true
	}

	pathIDs := make(map[string]bool)
	for _, p := range req.ExecutionPaths {
		pathIDs[p.Id] = true
	}

	// Scan test entries to determine coverage.
	coveredReqs := make(map[string]bool)
	coveredProps := make(map[string]bool)
	coveredPaths := make(map[string]bool)

	// Test cases cover requirements (via criterion references).
	for _, tc := range ts.TestCases {
		if tc.RequirementId != "" {
			if parentReq, ok := criterionToReq[tc.RequirementId]; ok {
				coveredReqs[parentReq] = true
			} else if reqIDs[tc.RequirementId] {
				coveredReqs[tc.RequirementId] = true
			}
		}
	}

	// Edge case tests also cover requirements.
	for _, ec := range ts.EdgeCaseTests {
		if ec.RequirementId != "" {
			if parentReq, ok := criterionToReq[ec.RequirementId]; ok {
				coveredReqs[parentReq] = true
			} else if reqIDs[ec.RequirementId] {
				coveredReqs[ec.RequirementId] = true
			}
		}
	}

	// Property tests cover correctness properties.
	for _, pt := range ts.PropertyTests {
		if pt.PropertyId != "" {
			coveredProps[pt.PropertyId] = true
		}
	}

	// Smoke tests cover execution paths.
	for _, sm := range ts.SmokeTests {
		if sm.ExecutionPathId != "" {
			coveredPaths[sm.ExecutionPathId] = true
		}
	}

	// Build the coverage report preserving the order from Requirements.
	var covered, uncovered []string

	for _, r := range req.Requirements {
		if coveredReqs[r.Id] {
			covered = append(covered, r.Id)
		} else {
			uncovered = append(uncovered, r.Id)
		}
	}

	for _, p := range req.CorrectnessProperties {
		if coveredProps[p.Id] {
			covered = append(covered, p.Id)
		} else {
			uncovered = append(uncovered, p.Id)
		}
	}

	for _, p := range req.ExecutionPaths {
		if coveredPaths[p.Id] {
			covered = append(covered, p.Id)
		} else {
			uncovered = append(uncovered, p.Id)
		}
	}

	if covered == nil {
		covered = []string{}
	}
	if uncovered == nil {
		uncovered = []string{}
	}

	return CoverageReport{
		Covered:   covered,
		Uncovered: uncovered,
	}
}
