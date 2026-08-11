package agentspec

// InlineRefs recursively resolves all $ref references in a JSON Schema
// map against the $defs section, replacing each $ref with the definition
// it points to. The $defs key is removed from the top-level result.
//
// If a $ref points to a definition not present in $defs, the $ref is
// left in place. If the input is nil or empty, an empty map is returned.
// Circular $ref chains are detected and broken at the point of recurrence.
func InlineRefs(schema map[string]any) map[string]any {
	// TODO: implement
	return nil
}

// CleanSchema recursively strips metadata noise from a JSON Schema map
// for the Anthropic API. It removes all title fields at every nesting
// level, all default fields at every nesting level, and the top-level
// $schema field. All description fields are preserved.
//
// If the input is nil or empty, an empty map is returned.
func CleanSchema(schema map[string]any) map[string]any {
	// TODO: implement
	return nil
}
