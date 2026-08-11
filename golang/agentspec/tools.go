package agentspec

import (
	"encoding/json"

	afspec "github.com/agent-fox-dev/spec-format"
)

// assessmentToolDef builds a fresh submit_assessment tool definition.
func assessmentToolDef() map[string]any {
	return map[string]any{
		"name":        "submit_assessment",
		"description": "Submit a PRD quality assessment with quality rating, summary, gaps, and clarifying questions.",
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"quality": map[string]any{
					"type": "string",
					"enum": []any{"ready", "needs_refinement", "incomplete"},
					"description": "Overall quality rating of the PRD.",
				},
				"summary": map[string]any{
					"type":        "string",
					"description": "Brief summary of the assessment.",
				},
				"gaps": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
					"description": "List of identified gaps in the PRD.",
				},
				"questions": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{
								"type":        "string",
								"description": "Unique identifier for the question.",
							},
							"text": map[string]any{
								"type":        "string",
								"description": "The question text.",
							},
							"context": map[string]any{
								"type":        "string",
								"description": "Context explaining why this question matters.",
							},
							"options": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "string",
								},
								"description": "Suggested answer options, if applicable.",
							},
							"required": map[string]any{
								"type":        "boolean",
								"description": "Whether an answer to this question is required.",
							},
						},
						"required": []any{"id", "text"},
					},
					"description": "Clarifying questions for the PRD author.",
				},
			},
			"required": []any{"quality", "summary", "gaps", "questions"},
		},
	}
}

// AssessmentTools returns a slice containing the Anthropic tool definition
// for the assessment phase. The returned slice contains exactly one entry
// for the submit_assessment tool with input schema fields: quality (enum),
// summary (string), gaps (array of strings), and questions (array of
// objects with id, text, context, options, required).
//
// Each call returns a new slice to prevent callers from mutating the
// shared definition.
func AssessmentTools() []map[string]any {
	return []map[string]any{assessmentToolDef()}
}

// RefinementTools returns a slice containing the Anthropic tool definitions
// for the refinement phase. The returned slice contains exactly two entries:
// submit_prd_update (with updated_prd string field) at index 0 and
// submit_assessment (same schema as AssessmentTools) at index 1.
//
// Each call returns a new slice to prevent callers from mutating the
// shared definition.
func RefinementTools() []map[string]any {
	return []map[string]any{
		{
			"name":        "submit_prd_update",
			"description": "Submit an updated PRD document.",
			"input_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"updated_prd": map[string]any{
						"type":        "string",
						"description": "The full updated PRD content.",
					},
				},
				"required": []any{"updated_prd"},
			},
		},
		assessmentToolDef(),
	}
}

// validArtifactNames is the set of artifact names accepted by ArtifactTool.
var validArtifactNames = map[string]bool{
	"requirements": true,
	"test_spec":    true,
	"tasks":        true,
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
	if !validArtifactNames[artifactName] {
		return []map[string]any{}
	}

	// Look up the schema bytes from afspec.Schemas().
	schemaKey := artifactName + ".v1.json"
	schemas := afspec.Schemas()
	schemaBytes, ok := schemas[schemaKey]
	if !ok {
		return []map[string]any{}
	}

	// Unmarshal the JSON Schema into a map.
	var schemaMap map[string]any
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		return []map[string]any{}
	}

	// Resolve all $ref references inline.
	inlined := InlineRefs(schemaMap)

	// Strip title, default, and $schema metadata.
	cleaned := CleanSchema(inlined)

	// Build the tool definition with the cleaned schema as the content property.
	toolName := "submit_" + artifactName
	return []map[string]any{
		{
			"name":        toolName,
			"description": "Submit the " + artifactName + " artifact.",
			"input_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content": cleaned,
				},
				"required": []any{"content"},
			},
		},
	}
}
