package agentspec

import (
	"embed"
	"fmt"
)

// templateFS embeds all prompt template files at compile time.
//
//go:embed templates/*.md
var templateFS embed.FS

// PromptTemplateNames lists the 10 embedded prompt template names.
var PromptTemplateNames = []string{
	"assessment_system",
	"assessment_user",
	"refinement_system",
	"refinement_user",
	"generation_system",
	"generation_user_base",
	"generation_user_requirements",
	"generation_user_test_spec",
	"generation_user_tasks",
	"repair_user",
}

// LoadPrompt loads a named prompt template from a project-local override path
// or the embedded defaults, stripping YAML frontmatter. The name must match
// [a-zA-Z0-9_-]+ to prevent path traversal.
//
// When a non-symlink override file exists at <projectDir>/.spec/prompts/<name>.md,
// it is used. Otherwise, the embedded default is used.
//
// Returns (content string, nil) on success, or ("", error) on failure.
func LoadPrompt(name, projectDir string) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("LoadPrompt: not implemented")
}

// LoadPromptTemplate loads a prompt template via LoadPrompt and performs safe
// $variable substitution. Unmatched $variable references pass through unchanged.
//
// Returns (substitutedContent string, nil) on success, or ("", error) on failure.
func LoadPromptTemplate(name, projectDir string, vars map[string]string) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("LoadPromptTemplate: not implemented")
}

// DetectProjectLanguage scans a project directory for manifest files and returns
// the detected language name and tooling hints string. Returns ("", "") when no
// recognized manifest files are present.
func DetectProjectLanguage(projectDir string) (language string, toolHints string) {
	// TODO: implement
	return "", ""
}

// FormatSpecLandscape formats spec landscape entries into markdown tables
// split into active and archived sections. Returns an empty string for an
// empty input slice.
func FormatSpecLandscape(landscape []map[string]any) string {
	// TODO: implement
	return ""
}

// AssessmentSystemPrompt loads the assessment_system template via LoadPrompt
// and returns its content.
func AssessmentSystemPrompt(projectDir string) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("AssessmentSystemPrompt: not implemented")
}

// AssessmentUserPrompt loads and renders the assessment_user template via
// LoadPromptTemplate, substituting $spec_name, $spec_landscape_block, and $prd_text.
func AssessmentUserPrompt(prdText, specName, projectDir string, specLandscape []map[string]any) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("AssessmentUserPrompt: not implemented")
}

// RefinementSystemPrompt loads the refinement_system template via LoadPrompt
// and returns its content.
func RefinementSystemPrompt(projectDir string) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("RefinementSystemPrompt: not implemented")
}

// RefinementUserPrompt loads and renders the refinement_user template via
// LoadPromptTemplate, substituting $prd_text, $assessment_block, $qa_block,
// and $spec_landscape_block.
func RefinementUserPrompt(prdText string, answers map[string]string, prevAssessment Assessment, projectDir string, specLandscape []map[string]any) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("RefinementUserPrompt: not implemented")
}

// GenerationSystemPrompt loads the generation_system template via LoadPrompt
// and returns its content.
func GenerationSystemPrompt(projectDir string) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("GenerationSystemPrompt: not implemented")
}

// GenerationUserPrompt loads generation_user_base and the artifact-specific
// template (generation_user_requirements, generation_user_test_spec, or
// generation_user_tasks) via LoadPromptTemplate, composes them with the
// appropriate variables, and returns the combined prompt.
//
// Returns an error if artifactName is not one of requirements, test_spec, or tasks.
func GenerationUserPrompt(prdText, artifactName, specID, projectDir string, priorArtifacts map[string]any, dependentInterfaces []map[string]any, specLandscape []map[string]any) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("GenerationUserPrompt: not implemented")
}

// RepairUserPrompt loads the repair_user template via LoadPromptTemplate,
// substituting the artifact name, original content, and validation errors.
func RepairUserPrompt(artifactName string, originalContent any, validationErrors []string, projectDir string) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("RepairUserPrompt: not implemented")
}
