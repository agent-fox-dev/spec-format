package agentspec

// AgentSpecConfig holds model and provider configuration loaded from
// config.toml with optional environment variable overrides.
type AgentSpecConfig struct {
	Model         string
	AuthMethod    string
	VertexProject string
	VertexRegion  string
}

// LoadConfig searches for config.toml in .specs/ relative to the current
// working directory, then in ~/.specs/. It parses the first file found,
// reading [model] and [provider] sections. The AF_SPEC_MODEL environment
// variable overrides the Model field after all file-based loading.
func LoadConfig() (AgentSpecConfig, error) {
	// TODO: implement
	return AgentSpecConfig{}, nil
}
