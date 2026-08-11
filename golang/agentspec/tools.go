package agentspec

// AssessmentTools returns a slice containing the Anthropic tool definition
// for the assessment phase. The returned slice contains exactly one entry
// for the submit_assessment tool with input schema fields: quality (enum),
// summary (string), gaps (array of strings), and questions (array of
// objects with id, text, context, options, required).
//
// Each call returns a new slice to prevent callers from mutating the
// shared definition.
func AssessmentTools() []map[string]any {
	// TODO: implement
	return nil
}

// RefinementTools returns a slice containing the Anthropic tool definitions
// for the refinement phase. The returned slice contains exactly two entries:
// submit_prd_update (with updated_prd string field) at index 0 and
// submit_assessment (same schema as AssessmentTools) at index 1.
//
// Each call returns a new slice to prevent callers from mutating the
// shared definition.
func RefinementTools() []map[string]any {
	// TODO: implement
	return nil
}

// ArtifactTool returns a slice containing the Anthropic tool definition for
// submitting a specific artifact type. It retrieves the JSON Schema for the
// artifact from afspec.Schemas(), resolves all $ref references via
// InlineRefs, strips metadata via CleanSchema, and embeds the cleaned
// schema as the content property of the submit_{artifactName} tool.
//
// Returns an empty slice without panicking if artifactName is not one of
// requirements, test_spec, or tasks, or if Schemas() does not contain an
// entry for the requested artifact.
func ArtifactTool(artifactName string) []map[string]any {
	// TODO: implement
	return nil
}
