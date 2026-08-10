package afspec

import "strings"

// renderConfig holds resolved options for rendering methods.
// It is intentionally unexported — external callers configure rendering
// via the RenderOption functional-option pattern.
type renderConfig struct {
	maxTokens int
}

// RenderOption is a functional option for rendering methods.
type RenderOption func(*renderConfig)

// WithMaxTokens returns a RenderOption that sets the token budget cap.
// When N is 0 or negative, it is treated as no budget constraint.
//
// TODO(spec-03): implement budget evaluation in task group 10.
func WithMaxTokens(n int) RenderOption {
	return func(cfg *renderConfig) {
		cfg.maxTokens = n
	}
}

// renderTestSpecSlim renders a test spec in slim mode, omitting verbose
// fields like assertion_pseudocode, input, expected, expected_effects,
// for_any_strategy, and invariant_check. Only id, description/kind,
// and requirement_id are preserved.
//
// TODO(spec-03): implement in task group 10.
func renderTestSpecSlim(ts *TestSpecV1Json) string {
	panic("renderTestSpecSlim not yet implemented")
}

// renderTestSpecScopedSlim renders a test spec in slim mode, filtered
// to only the test entries whose IDs are in the ids set.
//
// TODO(spec-03): implement in task group 10.
func renderTestSpecScopedSlim(ts *TestSpecV1Json, ids map[string]bool) string {
	panic("renderTestSpecScopedSlim not yet implemented")
}

// renderCombinedSlim renders the combined output with slim test spec.
//
// TODO(spec-03): implement in task group 10.
func (s *Spec) renderCombinedSlim() string {
	var sb strings.Builder

	sb.WriteString("# PRD\n\n")
	sb.WriteString(s.PRDBody)
	sb.WriteString("\n")

	sb.WriteString("# Requirements\n\n")
	if s.Requirements != nil {
		sb.WriteString(s.Requirements.Render())
	}
	sb.WriteString("\n")

	sb.WriteString("# Test Specification\n\n")
	if s.TestSpec != nil {
		sb.WriteString(renderTestSpecSlim(s.TestSpec))
	}
	sb.WriteString("\n")

	sb.WriteString("# Tasks\n\n")
	if s.Tasks != nil {
		sb.WriteString(s.Tasks.Render())
	}
	sb.WriteString("\n")

	// Architecture is omitted at Level 2 (already removed at Level 1)

	return sb.String()
}
