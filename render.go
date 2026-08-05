package afspec

// RenderCombined renders all artifacts (PRD body, requirements, test spec,
// tasks, architecture if present) as a single concatenated Markdown document.
func (s *Spec) RenderCombined() string {
	panic("not implemented")
}

// RenderIndividual renders each artifact separately and returns a map keyed
// by artifact name (e.g., "prd", "requirements", "test_spec", "tasks",
// "architecture") to its Markdown string. If architecture is absent, the
// "architecture" key is omitted from the returned map.
func (s *Spec) RenderIndividual() map[string]string {
	panic("not implemented")
}

// RenderIndividualScoped renders each artifact filtered to the refs of a
// target task group. It collects all requirement_refs and test_spec_refs
// from every subtask in targetGroup, renders only the referenced
// requirements and test entries, renders the target group with full subtask
// detail and all other groups as one-line summaries, and includes PRD body
// and architecture unfiltered. If the target group has no refs or does not
// exist, it falls back to full unscoped rendering.
func (s *Spec) RenderIndividualScoped(targetGroup int) map[string]string {
	panic("not implemented")
}

// Render renders the requirements artifact as a Markdown string.
func (r *RequirementsV1Json) Render() string {
	panic("not implemented")
}

// Render renders the test spec artifact as a Markdown string.
func (ts *TestSpecV1Json) Render() string {
	panic("not implemented")
}

// Render renders the tasks artifact as a Markdown string with
// checkbox-formatted subtasks.
func (t *TasksV1Json) Render() string {
	panic("not implemented")
}
