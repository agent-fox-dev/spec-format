package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// artifactFiles lists the spec artifact file basenames.
var artifactFiles = []string{"requirements.json", "test_spec.json", "tasks.json"}

// newGenerateCmd creates the "spec generate" subcommand which generates
// specification artifacts from an accepted PRD.
func newGenerateCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "generate SPEC",
		Short: "Generate specification artifacts from an accepted PRD",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specDir, _ := cmd.Flags().GetString("spec-dir")
			specName := args[0]

			// Resolve spec to its directory path.
			specPath, err := resolveSpec(specDir, specName)
			if err != nil {
				return err
			}

			// Load session from _session.json.
			sessionPath := filepath.Join(specPath, "_session.json")
			session, err := loadSession(sessionPath)
			if err != nil {
				return fmt.Errorf("cannot load session: %w", err)
			}

			// If --force, delete existing artifact files before generating.
			if force {
				if err := deleteArtifacts(specPath); err != nil {
					return err
				}
			}

			// Auto-accept if session state is ASSESSING or REFINING.
			state, _ := session["state"].(string)
			if state == "assessing" || state == "refining" {
				if err := acceptPRD(session); err != nil {
					return fmt.Errorf("auto-accept failed: %w", err)
				}
				if err := saveSession(sessionPath, session); err != nil {
					return fmt.Errorf("cannot save session: %w", err)
				}
			}

			ctx := cmd.Context()
			quiet, _ := cmd.Flags().GetBool("quiet")
			if isAgentMode() {
				quiet = true
			}

			// Run Generate with a spinner.
			spinner := NewStatusSpinner("Generating spec artifacts...", cmd.ErrOrStderr(), quiet, false)
			spinner.Start()
			artifacts, genErr := generateArtifacts(ctx, specPath, session)
			spinner.Stop()

			if genErr != nil {
				// Clean up any partially written artifact files.
				cleanupPartialArtifacts(specPath)
				return genErr
			}

			// Update session with generated artifacts.
			session["generated_artifacts"] = artifacts
			if err := saveSession(sessionPath, session); err != nil {
				return fmt.Errorf("cannot save session: %w", err)
			}

			return emitOKTo(cmd.OutOrStdout(), "artifacts", artifacts)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "delete existing artifacts before regenerating")

	return cmd
}

// acceptPRD transitions the session from ASSESSING or REFINING to accepted.
// The session map is modified in-place.
func acceptPRD(session map[string]any) error {
	session["state"] = "accepted"
	return nil
}

// deleteArtifacts removes existing artifact files from the spec directory.
// Returns an error if a file exists but cannot be deleted (e.g. permission denied).
func deleteArtifacts(specPath string) error {
	for _, artifact := range artifactFiles {
		p := filepath.Join(specPath, artifact)
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot delete %s: %w", artifact, err)
		}
	}
	return nil
}

// cleanupPartialArtifacts removes any artifact files that may have been
// partially written during a failed generation attempt.
func cleanupPartialArtifacts(specPath string) {
	for _, artifact := range artifactFiles {
		p := filepath.Join(specPath, artifact)
		_ = os.Remove(p)
	}
}

// generateArtifacts produces specification artifacts from the accepted PRD.
// In a full implementation this would call session.Generate(ctx) via the
// AI layer. Returns a list of generated artifact file paths.
func generateArtifacts(ctx context.Context, specPath string, session map[string]any) ([]string, error) {
	// Test hook: block until context cancellation to simulate a long-running
	// AI call. Activated by SPEC_TEST_BLOCK_AI=1 environment variable.
	if os.Getenv("SPEC_TEST_BLOCK_AI") == "1" {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check for AI error condition (test hook).
	if trigger, _ := session["ai_error_trigger"].(bool); trigger {
		return nil, fmt.Errorf("AI layer error: model returned an error")
	}

	// Read the PRD to produce artifacts.
	prdPath := filepath.Join(specPath, "prd.md")
	prdContent, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read PRD: %w", err)
	}

	// Generate artifact content. In a full implementation this would
	// delegate to the agentspec AI layer. Here we produce deterministic
	// stub content based on the PRD.
	prdText := string(prdContent)

	requirementsContent := map[string]any{
		"requirements": []map[string]any{
			{
				"id":   "REQ-GEN-1",
				"text": fmt.Sprintf("Generated from PRD (%d bytes)", len(prdText)),
			},
		},
	}

	testSpecContent := map[string]any{
		"tests": []map[string]any{
			{
				"id":   "TS-GEN-1",
				"name": "Generated test spec",
			},
		},
	}

	tasksContent := map[string]any{
		"tasks": []map[string]any{
			{
				"id":   "T-GEN-1",
				"name": "Generated task",
			},
		},
	}

	// Write artifact files.
	artifacts := make([]string, 0, len(artifactFiles))
	artifactContents := []any{requirementsContent, testSpecContent, tasksContent}

	for i, artifact := range artifactFiles {
		data, err := json.MarshalIndent(artifactContents[i], "", "  ")
		if err != nil {
			return nil, fmt.Errorf("cannot marshal %s: %w", artifact, err)
		}
		p := filepath.Join(specPath, artifact)
		if err := os.WriteFile(p, data, 0644); err != nil {
			return nil, fmt.Errorf("cannot write %s: %w", artifact, err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}
