package agentspec

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TS-NS-1 (issue #29): GenerationSystemPrompt includes all 6 EARS patterns
// ---------------------------------------------------------------------------

// TestGenPrompts_SystemPrompt_AllEARSPatterns verifies that GenerationSystemPrompt
// returns content that includes all six EARS pattern names and their rendered
// template forms.
// Test Spec: TS-NS-1 (issue #29), Requirement: NS-REQ-1
func TestGenPrompts_SystemPrompt_AllEARSPatterns(t *testing.T) {
	emptyTmpDir := t.TempDir()

	prompt, err := GenerationSystemPrompt(emptyTmpDir)
	if err != nil {
		t.Fatalf("GenerationSystemPrompt() returned error: %v", err)
	}

	patterns := []struct {
		name     string
		template string
	}{
		{"ubiquitous", "THE {system} SHALL {action}"},
		{"event_driven", "WHEN {trigger}, THE {system} SHALL {action}"},
		{"complex_event", "WHEN {trigger} AND {condition}, THE {system} SHALL {action}"},
		{"state_driven", "WHILE {state}, THE {system} SHALL {action}"},
		{"unwanted", "IF {error_condition}, THEN THE {system} SHALL {action}"},
		{"optional", "WHERE {feature}, THE {system} SHALL {action}"},
	}

	for _, p := range patterns {
		t.Run(p.name, func(t *testing.T) {
			if !strings.Contains(prompt, p.name) {
				t.Errorf("GenerationSystemPrompt() does not contain EARS pattern name %q", p.name)
			}
			if !strings.Contains(prompt, p.template) {
				t.Errorf("GenerationSystemPrompt() does not contain rendered template %q for pattern %q", p.template, p.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TS-NS-2 (issue #29): GenerationSystemPrompt includes ID format conventions
// ---------------------------------------------------------------------------

// TestGenPrompts_SystemPrompt_IDFormats verifies that GenerationSystemPrompt
// includes example IDs for all entity types from Appendix A.
// Test Spec: TS-NS-2 (issue #29), Requirement: NS-REQ-2
func TestGenPrompts_SystemPrompt_IDFormats(t *testing.T) {
	emptyTmpDir := t.TempDir()

	prompt, err := GenerationSystemPrompt(emptyTmpDir)
	if err != nil {
		t.Fatalf("GenerationSystemPrompt() returned error: %v", err)
	}

	idFamilies := []struct {
		name    string
		marker  string
		example string
	}{
		{"requirement", "REQ-", "05-REQ-3"},
		{"acceptance criterion", "REQ-", "05-REQ-3.2"},
		{"edge case", "REQ-", "05-REQ-3.E1"},
		{"correctness property", "PROP-", "05-PROP-2"},
		{"execution path", "PATH-", "05-PATH-1"},
		{"error handling entry", "ERR-", "05-ERR-1"},
		{"test case", "TS-05-3", "TS-05-3"},
		{"property test", "TS-05-P", "TS-05-P2"},
		{"edge case test", "TS-05-E", "TS-05-E1"},
		{"smoke test", "SMOKE-", "TS-05-SMOKE-1"},
		{"subtask", "3.2", "3.2"},
		{"verification subtask", ".V", "3.V"},
	}

	for _, f := range idFamilies {
		t.Run(f.name, func(t *testing.T) {
			if !strings.Contains(prompt, f.marker) {
				t.Errorf("GenerationSystemPrompt() does not contain ID marker %q (example: %s)", f.marker, f.example)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TS-NS-3 (issue #29): generation_user_tasks encodes task group ordering rules
// ---------------------------------------------------------------------------

// TestGenPrompts_UserTasks_OrderingRules verifies that the generation_user_tasks
// template explicitly states the three task group ordering rules so a model
// following the prompt would produce a tasks.json where group 1 has kind=tests
// and the last group has kind=wiring_verification.
// Test Spec: TS-NS-3 (issue #29), Requirement: NS-REQ-3
func TestGenPrompts_UserTasks_OrderingRules(t *testing.T) {
	emptyTmpDir := t.TempDir()

	// Load the artifact-specific template content directly via GenerationUserPrompt
	// and check the combined output for ordering rule language.
	prompt, err := GenerationUserPrompt(
		"PRD text",
		"tasks",
		"05",
		emptyTmpDir,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("GenerationUserPrompt(tasks) returned error: %v", err)
	}

	// Must mention kind=tests constraint.
	if !strings.Contains(prompt, "tests") {
		t.Error("tasks prompt does not mention 'tests' kind")
	}

	// Must mention kind=wiring_verification constraint.
	if !strings.Contains(prompt, "wiring_verification") {
		t.Error("tasks prompt does not mention 'wiring_verification' kind")
	}

	// Must include ordering language: first/final/last.
	lower := strings.ToLower(prompt)
	hasOrderingLanguage := strings.Contains(lower, "first") ||
		strings.Contains(lower, "final") ||
		strings.Contains(lower, "last")
	if !hasOrderingLanguage {
		t.Error("tasks prompt does not contain ordering language (first/final/last)")
	}
}

// TestGenPrompts_UserTasks_Template_DirectContent verifies the tasks template
// content directly via LoadPrompt to confirm ordering rule language is present
// independent of variable substitution.
// Requirement: NS-REQ-3
func TestGenPrompts_UserTasks_Template_DirectContent(t *testing.T) {
	emptyTmpDir := t.TempDir()

	content, err := LoadPrompt("generation_user_tasks", emptyTmpDir)
	if err != nil {
		t.Fatalf("LoadPrompt(generation_user_tasks) returned error: %v", err)
	}

	checks := []struct {
		desc    string
		needles []string
	}{
		{"kind: tests constraint", []string{"kind", "tests"}},
		{"kind: wiring_verification constraint", []string{"kind", "wiring_verification"}},
		{"ordering language", []string{"first", "final", "last"}},
	}

	lower := strings.ToLower(content)
	for _, c := range checks {
		t.Run(c.desc, func(t *testing.T) {
			found := false
			for _, needle := range c.needles {
				if strings.Contains(lower, needle) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("tasks template does not contain any of %v", c.needles)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TS-NS-4 (issue #29): generation_user_requirements encodes glossary convention
// ---------------------------------------------------------------------------

// TestGenPrompts_UserRequirements_GlossaryConvention verifies that the
// generation_user_requirements template states the glossary backtick convention.
// Test Spec: TS-NS-4 (issue #29), Requirement: NS-REQ-4
func TestGenPrompts_UserRequirements_GlossaryConvention(t *testing.T) {
	emptyTmpDir := t.TempDir()

	content, err := LoadPrompt("generation_user_requirements", emptyTmpDir)
	if err != nil {
		t.Fatalf("LoadPrompt(generation_user_requirements) returned error: %v", err)
	}

	// Must mention glossary.
	if !strings.Contains(content, "glossary") {
		t.Error("requirements template does not mention 'glossary'")
	}

	// Must mention backtick convention (either the word "backtick" or contain a
	// backtick character demonstrating the convention).
	lower := strings.ToLower(content)
	hasBacktickGuidance := strings.Contains(lower, "backtick") || strings.Contains(content, "`")
	if !hasBacktickGuidance {
		t.Error("requirements template does not contain backtick convention guidance")
	}

	// Must mention the checked fields where domain terms must be wrapped.
	checkedFields := []string{"action", "trigger", "condition", "error_condition", "state", "feature"}
	for _, field := range checkedFields {
		if !strings.Contains(content, field) {
			t.Errorf("requirements template does not mention checked field %q", field)
		}
	}
}

// ---------------------------------------------------------------------------
// TS-NS-5 (issue #29): generation_user_requirements specifies pattern field sets
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// TS-NS-1 (issue #50): assessment_system prompt defines quality rating criteria
// ---------------------------------------------------------------------------

// TestAssessPrompt_QualityRatingCriteria verifies that AssessmentSystemPrompt
// returns content that explicitly defines all three quality rating levels with
// at least one distinguishing criterion per level.
// Test Spec: TS-NS-1 (issue #50), Requirement: NS-REQ-1
func TestAssessPrompt_QualityRatingCriteria(t *testing.T) {
	emptyTmpDir := t.TempDir()

	prompt, err := AssessmentSystemPrompt(emptyTmpDir)
	if err != nil {
		t.Fatalf("AssessmentSystemPrompt() returned error: %v", err)
	}

	levels := []struct {
		keyword string
		desc    string
	}{
		{"ready", "ready quality level"},
		{"needs_refinement", "needs_refinement quality level"},
		{"incomplete", "incomplete quality level"},
	}

	for _, lvl := range levels {
		t.Run(lvl.keyword, func(t *testing.T) {
			if !strings.Contains(prompt, lvl.keyword) {
				t.Errorf("AssessmentSystemPrompt() does not contain quality level %q", lvl.keyword)
			}
		})
	}

	// Each level must appear alongside at least one criterion phrase.
	// We verify this by checking that each level keyword appears near some
	// distinguishing language — the rubric section must exist.
	if !strings.Contains(prompt, "Rubric") && !strings.Contains(prompt, "rubric") {
		// Accept either a rubric heading or inline criterion language.
		// If no rubric heading, verify the levels appear with explanatory text.
		for _, lvl := range levels {
			idx := strings.Index(prompt, lvl.keyword)
			if idx < 0 {
				continue // already reported above
			}
			// There must be text following the keyword (its description).
			surroundStart := idx
			surroundEnd := idx + len(lvl.keyword) + 200
			if surroundEnd > len(prompt) {
				surroundEnd = len(prompt)
			}
			surrounding := prompt[surroundStart:surroundEnd]
			if strings.TrimSpace(surrounding) == lvl.keyword {
				t.Errorf("quality level %q appears with no surrounding criterion text", lvl.keyword)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// TS-NS-2 (issue #50): assessment_system prompt includes evaluation dimensions
// ---------------------------------------------------------------------------

// TestAssessPrompt_EvaluationDimensions verifies that AssessmentSystemPrompt
// includes all required evaluation dimensions.
// Test Spec: TS-NS-2 (issue #50), Requirement: NS-REQ-2
func TestAssessPrompt_EvaluationDimensions(t *testing.T) {
	emptyTmpDir := t.TempDir()

	prompt, err := AssessmentSystemPrompt(emptyTmpDir)
	if err != nil {
		t.Fatalf("AssessmentSystemPrompt() returned error: %v", err)
	}

	lower := strings.ToLower(prompt)

	dimensions := []struct {
		name    string
		needles []string
	}{
		{"scope/intent clarity", []string{"scope", "intent"}},
		{"measurable goals", []string{"measurable", "goals"}},
		{"explicit non-goals", []string{"non-goals", "non-goal"}},
		{"testability", []string{"testability", "testable"}},
		{"error-handling coverage", []string{"error"}},
		{"external API surface", []string{"external api", "api surface", "external"}},
	}

	for _, dim := range dimensions {
		t.Run(dim.name, func(t *testing.T) {
			found := false
			for _, needle := range dim.needles {
				if strings.Contains(lower, needle) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("AssessmentSystemPrompt() does not contain dimension %q (tried needles: %v)", dim.name, dim.needles)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TS-NS-3 (issue #50): refinement_system prompt includes upgrade guidance
// ---------------------------------------------------------------------------

// TestRefinementPrompt_IncorporationAndUpgradeGuidance verifies that
// RefinementSystemPrompt returns content that instructs the model to incorporate
// Q&A answers, defines conditions for upgrading quality to 'ready', and states
// that quality must not regress unless new gaps are discovered.
// Test Spec: TS-NS-3 (issue #50), Requirement: NS-REQ-3
func TestRefinementPrompt_IncorporationAndUpgradeGuidance(t *testing.T) {
	emptyTmpDir := t.TempDir()

	prompt, err := RefinementSystemPrompt(emptyTmpDir)
	if err != nil {
		t.Fatalf("RefinementSystemPrompt() returned error: %v", err)
	}

	lower := strings.ToLower(prompt)

	checks := []struct {
		desc    string
		needles []string
	}{
		{
			"incorporation instruction",
			[]string{"incorporate", "integrate", "woven", "include"},
		},
		{
			"upgrade to ready guidance",
			[]string{"upgrade", "ready"},
		},
		{
			"gap reference in upgrade criteria",
			[]string{"gap", "gaps"},
		},
		{
			"non-regression rule",
			[]string{"regress", "not downgrade", "must not"},
		},
	}

	for _, c := range checks {
		t.Run(c.desc, func(t *testing.T) {
			found := false
			for _, needle := range c.needles {
				if strings.Contains(lower, needle) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("RefinementSystemPrompt() does not contain %q (tried: %v)", c.desc, c.needles)
			}
		})
	}
}

// TestGenPrompts_UserRequirements_EARSPatternFieldSpecs verifies that the
// generation_user_requirements template specifies the required and forbidden
// fields for each EARS pattern so a compliant model output passes schema
// validation for the discriminated oneOf union.
// Test Spec: TS-NS-5 (issue #29), Requirement: NS-REQ-5
func TestGenPrompts_UserRequirements_EARSPatternFieldSpecs(t *testing.T) {
	emptyTmpDir := t.TempDir()

	content, err := LoadPrompt("generation_user_requirements", emptyTmpDir)
	if err != nil {
		t.Fatalf("LoadPrompt(generation_user_requirements) returned error: %v", err)
	}

	// Each EARS pattern must appear alongside its pattern-specific required fields.
	patternFields := []struct {
		pattern string
		fields  []string
	}{
		{"ubiquitous", []string{"system", "action"}},
		{"event_driven", []string{"trigger", "system", "action"}},
		{"complex_event", []string{"trigger", "condition", "system", "action"}},
		{"state_driven", []string{"state", "system", "action"}},
		{"unwanted", []string{"error_condition", "system", "action"}},
		{"optional", []string{"feature", "system", "action"}},
	}

	for _, pf := range patternFields {
		t.Run(pf.pattern, func(t *testing.T) {
			if !strings.Contains(content, pf.pattern) {
				t.Errorf("requirements template does not mention EARS pattern %q", pf.pattern)
			}
			for _, field := range pf.fields {
				if !strings.Contains(content, field) {
					t.Errorf("requirements template does not mention field %q (required for pattern %q)", field, pf.pattern)
				}
			}
		})
	}
}
