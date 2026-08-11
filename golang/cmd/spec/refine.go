package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newRefineCmd creates the "spec refine" subcommand which iteratively
// refines a PRD through AI-driven Q&A.
func newRefineCmd() *cobra.Command {
	var answers string
	var force bool

	cmd := &cobra.Command{
		Use:   "refine SPEC",
		Short: "Iteratively refine a PRD through AI-driven Q&A",
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

			// Handle --force: reset session before proceeding.
			if force {
				if err := forceResetSession(specPath, session); err != nil {
					return fmt.Errorf("force reset failed: %w", err)
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

			// Handle --answers path: read answers and run Refine.
			if answers != "" {
				answersData, err := readAnswersInput(cmd, answers)
				if err != nil {
					return err
				}

				spinner := NewStatusSpinner("Refining PRD...", cmd.ErrOrStderr(), quiet, false)
				spinner.Start()
				result, refineErr := refineWithAnswers(ctx, answersData)
				spinner.Stop()
				if refineErr != nil {
					return refineErr
				}

				return emitOKTo(cmd.OutOrStdout(), "result", result)
			}

			// No --answers: assess or return pending questions.
			state, _ := session["state"].(string)
			if sessionNeedsAssessment(state) {
				spinner := NewStatusSpinner("Assessing PRD quality...", cmd.ErrOrStderr(), quiet, false)
				spinner.Start()
				assessment, assessErr := assessPRD(ctx, specPath)
				spinner.Stop()
				if assessErr != nil {
					return assessErr
				}

				return emitOKTo(cmd.OutOrStdout(), "assessment", assessment)
			}

			// Session does not need assessment -- return pending questions.
			questions := getPendingQuestions(session)
			return emitOKTo(cmd.OutOrStdout(), "questions", questions)
		},
	}

	cmd.Flags().StringVar(&answers, "answers", "", "path to answers JSON file (use '-' for stdin)")
	cmd.Flags().BoolVar(&force, "force", false, "reset session state before refining")

	return cmd
}

// loadSession reads and parses _session.json from the given path.
func loadSession(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var session map[string]any
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return session, nil
}

// saveSession writes the session map as pretty-printed JSON to the
// given path.
func saveSession(path string, session map[string]any) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// forceResetSession deletes artifact files and resets the session
// to its initial state. The session map is modified in-place.
func forceResetSession(specPath string, session map[string]any) error {
	// Delete artifact files.
	for _, artifact := range []string{"requirements.json", "test_spec.json", "tasks.json"} {
		p := filepath.Join(specPath, artifact)
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot delete %s: %w", artifact, err)
		}
	}

	// Reset session state to INIT and clear history.
	session["state"] = "init"
	session["assessment_history"] = []any{}
	session["qa_exchanges"] = []any{}
	session["generated_artifacts"] = []any{}

	return nil
}

// sessionNeedsAssessment returns true when the session state indicates
// that an AI assessment should be run (state is INIT or ASSESSING).
func sessionNeedsAssessment(state string) bool {
	return state == "init" || state == "assessing"
}

// assessPRD performs a PRD quality assessment. In a full implementation
// this would call the AI model via session.Assess(ctx). Returns the
// assessment result without mutating session state.
func assessPRD(ctx context.Context, specPath string) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read the PRD content.
	prdPath := filepath.Join(specPath, "prd.md")
	prdContent, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read PRD: %w", err)
	}

	// Produce a basic assessment. In a full implementation this would
	// delegate to the agentspec AI layer.
	prdText := string(prdContent)
	quality := "fair"
	if len(strings.Split(prdText, "\n")) > 20 {
		quality = "good"
	}

	assessment := map[string]any{
		"quality":   quality,
		"summary":   "PRD assessment complete",
		"gaps":      []string{},
		"questions": []any{},
	}

	return assessment, nil
}

// getPendingQuestions extracts unanswered questions from the session's
// assessment history, filtering out any that have been answered in
// qa_exchanges.
func getPendingQuestions(session map[string]any) []any {
	history, _ := session["assessment_history"].([]any)
	if len(history) == 0 {
		return []any{}
	}

	// Get questions from the latest assessment.
	latest, ok := history[len(history)-1].(map[string]any)
	if !ok {
		return []any{}
	}

	questions, _ := latest["questions"].([]any)
	if questions == nil {
		return []any{}
	}

	// Build a set of answered question IDs from qa_exchanges.
	answered := make(map[string]bool)
	exchanges, _ := session["qa_exchanges"].([]any)
	for _, ex := range exchanges {
		exMap, ok := ex.(map[string]any)
		if !ok {
			continue
		}
		ans, _ := exMap["answers"].(map[string]any)
		for k := range ans {
			answered[k] = true
		}
	}

	var pending []any
	for _, q := range questions {
		qMap, ok := q.(map[string]any)
		if !ok {
			pending = append(pending, q)
			continue
		}
		id, _ := qMap["id"].(string)
		if !answered[id] {
			pending = append(pending, q)
		}
	}

	if pending == nil {
		pending = []any{}
	}
	return pending
}

// readAnswersInput reads and parses answers JSON from a file or stdin.
// If the parsed JSON has an "answers" key containing a map, that inner
// map is returned (unwrap behavior per 08-REQ-8.3/8.4).
func readAnswersInput(cmd *cobra.Command, answersArg string) (map[string]any, error) {
	var data []byte
	var err error

	if answersArg == "-" {
		data, err = io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("cannot read answers from stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(answersArg)
		if err != nil {
			return nil, fmt.Errorf("cannot read answers file %q: %w", answersArg, err)
		}
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("cannot parse answers JSON: %w", err)
	}

	// Unwrap 'answers' key if present and it's a map.
	if inner, ok := parsed["answers"]; ok {
		if innerMap, ok := inner.(map[string]any); ok {
			return innerMap, nil
		}
	}

	return parsed, nil
}

// refineWithAnswers runs a refinement pass using the provided answers.
// In a full implementation this would call session.Refine(ctx, answers)
// via the AI layer. Returns the refinement result.
func refineWithAnswers(ctx context.Context, answers map[string]any) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := map[string]any{
		"status":  "refined",
		"answers": answers,
	}

	return result, nil
}
