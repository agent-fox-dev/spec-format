package agentspec

import (
	"context"
	"fmt"
)

// agentOptions holds the resolved configuration from AgentOption functional options.
type agentOptions struct {
	specLandscape       []map[string]any
	dependentInterfaces []map[string]any
	projectDir          string
	onArtifact          func(name string, content any)
}

// AgentOption is a functional option type for SpecAgent methods.
type AgentOption func(*agentOptions)

// WithSpecLandscape returns an AgentOption that sets the spec landscape slice.
func WithSpecLandscape(landscape []map[string]any) AgentOption {
	return func(o *agentOptions) {
		o.specLandscape = landscape
	}
}

// WithDependentInterfaces returns an AgentOption that sets the dependent
// spec interfaces slice.
func WithDependentInterfaces(interfaces []map[string]any) AgentOption {
	return func(o *agentOptions) {
		o.dependentInterfaces = interfaces
	}
}

// WithProjectDir returns an AgentOption that sets the project directory.
func WithProjectDir(dir string) AgentOption {
	return func(o *agentOptions) {
		o.projectDir = dir
	}
}

// WithOnArtifact returns an AgentOption that sets a callback invoked after
// each artifact is successfully generated. A nil callback is stored and
// skipped during invocation without panicking.
func WithOnArtifact(fn func(name string, content any)) AgentOption {
	return func(o *agentOptions) {
		o.onArtifact = fn
	}
}

// applyOptions resolves a slice of AgentOption values into an agentOptions struct.
func applyOptions(opts []AgentOption) agentOptions {
	var o agentOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// SpecAgent holds the model tier and implements the AI agent pipeline
// methods: AssessPRD, RefinePRD, and GenerateArtifacts.
type SpecAgent struct {
	modelTier string

	// aiCallFunc is an internal hook for testing. When non-nil, it replaces
	// the real AICall function. Exported tests set this via the unexported
	// field (package-level tests have access).
	aiCallFunc func(ctx context.Context, opts AICallOptions) (string, any, error)
}

// NewSpecAgent creates a SpecAgent with the given model tier string.
func NewSpecAgent(modelTier string) *SpecAgent {
	return &SpecAgent{modelTier: modelTier}
}

// AssessPRD sends a PRD to the LLM with the assessment system prompt and
// submit_assessment tool, returning a validated Assessment.
func (sa *SpecAgent) AssessPRD(ctx context.Context, prdText, specName string, opts ...AgentOption) (Assessment, error) {
	// TODO: implement
	return Assessment{}, fmt.Errorf("AssessPRD: not implemented")
}

// RefinePRD sends a PRD with user answers and prior assessment to the LLM,
// returning an updated PRD text and new Assessment.
func (sa *SpecAgent) RefinePRD(ctx context.Context, prdText string, answers map[string]string, prevAssessment Assessment, opts ...AgentOption) (string, Assessment, error) {
	// TODO: implement
	return "", Assessment{}, fmt.Errorf("RefinePRD: not implemented")
}

// GenerateArtifacts sequentially generates requirements, test_spec, and tasks
// artifacts with validation and repair loops.
func (sa *SpecAgent) GenerateArtifacts(ctx context.Context, prdText, specID, specName string, opts ...AgentOption) (map[string]any, error) {
	// TODO: implement
	return nil, fmt.Errorf("GenerateArtifacts: not implemented")
}
