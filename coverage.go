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
// If the Requirements struct has no entities to cover, the report indicates
// 100% coverage (empty Uncovered list). If the TestSpec has no test entries,
// all requirement entities appear in the Uncovered list.
func (ts *TestSpecV1Json) ComputeCoverage(req *RequirementsV1Json) CoverageReport {
	panic("not implemented")
}
