package afspec

import "regexp"

// reqIDRe matches requirement and acceptance-criterion ID patterns in free text
// (e.g. "02-REQ-1", "02-REQ-1.1", "02-REQ-1.E1"). Compiled once at package
// init time via regexp.MustCompile — not per function call.
var reqIDRe = regexp.MustCompile(`\d{2}-REQ-\d+(?:\.\d+|\.E\d+)?`)

// tsIDRe matches test spec ID patterns in free text (e.g. "TS-02-1").
// Compiled once at package init time via regexp.MustCompile.
var tsIDRe = regexp.MustCompile(`TS-\d{2}-\d+`)

// inferRefsFromTraceability scans spec.Tasks.Traceability for entries whose
// TaskId starts with "{targetGroup}." and returns the collected RequirementId
// and TestSpecId values as separate slices. Empty string fields are skipped.
func inferRefsFromTraceability(spec *Spec, targetGroup int) (reqRefs []string, tsRefs []string) {
	panic("not implemented: inferRefsFromTraceability")
}

// inferRefsFromSubtaskText scans the Title and Details fields of all subtasks
// in the target group using reqIDRe and tsIDRe, validates each match against
// the set of IDs actually present in the spec, and returns two collections
// of validated ID strings (inferred requirement refs, inferred test spec refs).
func inferRefsFromSubtaskText(spec *Spec, targetGroup int) (reqRefs []string, tsRefs []string) {
	panic("not implemented: inferRefsFromSubtaskText")
}
