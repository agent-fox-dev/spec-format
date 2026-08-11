package agentspec

import "fmt"

// ModelTier classifies LLM models by capability level.
type ModelTier string

const (
	// TierSimple represents lightweight, cost-effective models.
	TierSimple ModelTier = "SIMPLE"
	// TierStandard represents balanced capability models.
	TierStandard ModelTier = "STANDARD"
	// TierAdvanced represents the most capable models.
	TierAdvanced ModelTier = "ADVANCED"
)

// ModelEntry holds metadata for a registered model.
type ModelEntry struct {
	ModelID string
	Tier    ModelTier
	Variant string
}

// ModelRegistry maps model ID strings to their ModelEntry metadata.
// Populated at init time with known Anthropic models.
var ModelRegistry = map[string]ModelEntry{}

// TierDefaults maps each ModelTier to its default model ID string.
var TierDefaults = map[ModelTier]string{}

// ResolveModel resolves a model name or tier constant string to a canonical
// Anthropic model ID. It performs case-insensitive matching for tier constants
// and direct lookup for model ID keys in ModelRegistry.
//
// Returns the resolved model ID and nil on success, or an empty string and
// a descriptive error if the name is not recognized.
func ResolveModel(name string) (string, error) {
	// TODO: implement model resolution logic
	return "", fmt.Errorf("ResolveModel: not implemented")
}
