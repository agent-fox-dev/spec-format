package agentspec

import (
	"path/filepath"
	"testing"
)

// TestTS_NS_5_ResolveAgent_UsesPerPhaseConfig verifies that resolveAgent()
// reads from LoadConfig() and selects the per-phase model tier, not a
// hardcoded "STANDARD" value.
// Test Spec: TS-NS-5, Requirement: NS-REQ-5
func TestTS_NS_5_ResolveAgent_UsesPerPhaseConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `[model]
model = "ADVANCED"
assess_model = "SIMPLE"
generate_model = "STANDARD"
`
	writeTestFile(t, filepath.Join(tmpDir, ".specs", "config.toml"), configContent)
	chdirTemp(t, tmpDir)

	t.Setenv("HOME", t.TempDir())
	t.Setenv("AF_SPEC_MODEL", "")

	// Create a session with no injected agent (agent == nil) so resolveAgent
	// falls through to LoadConfig().
	session := &SpecSession{
		specDir:            tmpDir,
		agent:              nil,
		Current:            StateInit,
		PRDPath:            "prd.md",
		AssessmentHistory:  []Assessment{},
		QAExchanges:        []QAExchange{},
		GeneratedArtifacts: []string{},
	}

	cases := []struct {
		phase    string
		wantTier string
	}{
		{"assess", "SIMPLE"},
		{"refine", "ADVANCED"}, // inherits top-level model
		{"generate", "STANDARD"},
	}

	for _, tc := range cases {
		a := session.resolveAgent(tc.phase)
		sa, ok := a.(*SpecAgent)
		if !ok {
			t.Fatalf("resolveAgent(%q) returned %T; want *SpecAgent", tc.phase, a)
		}
		if sa.modelTier != tc.wantTier {
			t.Errorf("resolveAgent(%q).modelTier = %q; want %q", tc.phase, sa.modelTier, tc.wantTier)
		}
	}
}

// TestTS_NS_3_ResolveAgent_BackwardCompat verifies that when no per-phase
// fields are set, resolveAgent returns the top-level model for all phases.
// Test Spec: TS-NS-3, Requirement: NS-REQ-3
func TestTS_NS_3_ResolveAgent_BackwardCompat(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `[model]
model = "ADVANCED"
`
	writeTestFile(t, filepath.Join(tmpDir, ".specs", "config.toml"), configContent)
	chdirTemp(t, tmpDir)

	t.Setenv("HOME", t.TempDir())
	t.Setenv("AF_SPEC_MODEL", "")

	session := &SpecSession{
		specDir:            tmpDir,
		agent:              nil,
		Current:            StateInit,
		PRDPath:            "prd.md",
		AssessmentHistory:  []Assessment{},
		QAExchanges:        []QAExchange{},
		GeneratedArtifacts: []string{},
	}

	for _, phase := range []string{"assess", "refine", "generate"} {
		a := session.resolveAgent(phase)
		sa, ok := a.(*SpecAgent)
		if !ok {
			t.Fatalf("resolveAgent(%q) returned %T; want *SpecAgent", phase, a)
		}
		if sa.modelTier != "ADVANCED" {
			t.Errorf("resolveAgent(%q).modelTier = %q; want %q", phase, sa.modelTier, "ADVANCED")
		}
	}
}

// TestResolveAgent_InjectMockBypassesConfig verifies that when an agent is
// injected, resolveAgent() returns it without calling LoadConfig().
func TestResolveAgent_InjectMockBypassesConfig(t *testing.T) {
	mock := &mockAssessor{}
	session := &SpecSession{
		specDir: t.TempDir(),
		agent:   mock,
		Current: StateInit,
	}

	result := session.resolveAgent("assess")
	if result != mock {
		t.Errorf("resolveAgent() = %v; want injected mock %v", result, mock)
	}
}
