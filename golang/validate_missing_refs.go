package afspec

// checkMissingSubtaskRefs inspects all subtasks in the given TaskGroup and
// returns a ValidationEntry warning for each subtask that has empty
// RequirementRefs or TestSpecRefs. Groups with Kind == TaskGroupKindWiringVerification
// are skipped entirely. The warning message follows the format:
//
//	"Subtask {id} has empty {field_names} — scoped rendering will fall back to full spec dump"
//
// where {field_names} is "requirement_refs", "test_spec_refs", or
// "requirement_refs and test_spec_refs" depending on which fields are empty.
func checkMissingSubtaskRefs(group TaskGroup) []ValidationEntry {
	panic("not implemented: checkMissingSubtaskRefs")
}
