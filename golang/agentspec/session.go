package agentspec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	afspec "github.com/agent-fox-dev/spec-format"
)

// validSessionStates enumerates all recognized SessionState values.
var validSessionStates = map[SessionState]bool{
	StateInit:        true,
	StateAssessing:   true,
	StateRefining:    true,
	StatePRDAccepted: true,
	StateGenerating:  true,
	StateGenerated:   true,
}

// SessionState enumerates the valid states of a SpecSession.
type SessionState string

const (
	StateInit        SessionState = "init"
	StateAssessing   SessionState = "assessing"
	StateRefining    SessionState = "refining"
	StatePRDAccepted SessionState = "prd_accepted"
	StateGenerating  SessionState = "generating"
	StateGenerated   SessionState = "generated"
)

// Assessment captures PRD quality assessment output: Quality, Summary,
// Gaps, and Questions.
type Assessment struct {
	Quality   string           `json:"quality"`
	Summary   string           `json:"summary"`
	Gaps      []string         `json:"gaps"`
	Questions []map[string]any `json:"questions"`
}

// QAExchange records a set of answers to assessment questions at a
// point in time, including the assessment index the answers correspond
// to and a UTC timestamp.
type QAExchange struct {
	AssessmentIndex int               `json:"assessment_index"`
	Answers         map[string]string `json:"answers"`
	Timestamp       time.Time         `json:"timestamp"`
}

// LastError records the most recent error encountered during a session
// operation.
type LastError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// assessor abstracts the SpecAgent AI pipeline for testability.
// SpecAgent satisfies this interface; tests provide a mock.
type assessor interface {
	AssessPRD(ctx context.Context, prdText, specName string, opts ...AgentOption) (Assessment, error)
	RefinePRD(ctx context.Context, prdText string, answers map[string]string, prevAssessment Assessment, opts ...AgentOption) (string, Assessment, error)
	GenerateArtifacts(ctx context.Context, prdText, specID, specName string, opts ...AgentOption) (map[string]any, error)
}

// SpecSession represents the stateful session for authoring a single spec,
// persisted to _session.json inside a spec directory.
type SpecSession struct {
	specDir            string
	agent              assessor
	Current            SessionState `json:"state"`
	Mode               string       `json:"mode"`
	PRDPath            string       `json:"prd_path"`
	AssessmentHistory  []Assessment `json:"assessment_history"`
	QAExchanges        []QAExchange `json:"qa_exchanges"`
	GeneratedArtifacts []string     `json:"generated_artifacts"`
	LastErr            *LastError   `json:"last_error"`
}

// State returns the current SessionState of the session.
func (s *SpecSession) State() SessionState {
	return s.Current
}

// SpecDir returns the spec directory path associated with this session.
func (s *SpecSession) SpecDir() string {
	return s.specDir
}

// Assessment returns a pointer to the most recent Assessment in the
// session's assessment history, or nil if the history is empty.
func (s *SpecSession) Assessment() *Assessment {
	if len(s.AssessmentHistory) == 0 {
		return nil
	}
	return &s.AssessmentHistory[len(s.AssessmentHistory)-1]
}

// AcceptPRD transitions the session state to StatePRDAccepted and
// persists the updated state to _session.json atomically. Returns a
// SessionError if the current state is not StateAssessing or StateRefining.
// If the atomic persistence fails, the in-memory state is reverted.
func (s *SpecSession) AcceptPRD() error {
	if s.Current != StateAssessing && s.Current != StateRefining {
		return &SessionError{
			Msg: fmt.Sprintf("cannot accept PRD from state %q; must be %q or %q",
				s.Current, StateAssessing, StateRefining),
		}
	}

	prevState := s.Current
	s.Current = StatePRDAccepted

	if err := s.persistSession(); err != nil {
		// Revert in-memory state on persistence failure.
		s.Current = prevState
		return err
	}
	return nil
}

// PendingQuestions returns the questions from the most recent Assessment
// that have not yet been answered via QA exchanges. Returns an empty
// slice if there are no pending questions or the assessment history is empty.
func (s *SpecSession) PendingQuestions() []map[string]any {
	a := s.Assessment()
	if a == nil || len(a.Questions) == 0 {
		return []map[string]any{}
	}
	return a.Questions
}

// persistSession writes the session state to _session.json inside specDir
// using the atomic temp-file-and-rename pattern. Returns a SessionError
// on failure and cleans up the temporary file.
func (s *SpecSession) persistSession() error {
	data, err := json.Marshal(s)
	if err != nil {
		return &SessionError{
			Msg:   fmt.Sprintf("failed to marshal session: %v", err),
			Cause: err,
		}
	}

	sessionPath := filepath.Join(s.specDir, "_session.json")
	tmpFile, err := os.CreateTemp(s.specDir, "_session.json.tmp.*")
	if err != nil {
		return &SessionError{
			Msg:   fmt.Sprintf("failed to create temp file for _session.json: %v", err),
			Cause: err,
		}
	}
	tmpPath := tmpFile.Name()

	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return &SessionError{
			Msg:   fmt.Sprintf("failed to write _session.json temp file: %v", err),
			Cause: err,
		}
	}
	if err := tmpFile.Close(); err != nil {
		return &SessionError{
			Msg:   fmt.Sprintf("failed to close _session.json temp file: %v", err),
			Cause: err,
		}
	}

	if err := os.Rename(tmpPath, sessionPath); err != nil {
		return &SessionError{
			Msg:   fmt.Sprintf("failed to rename temp file to _session.json: %v", err),
			Cause: err,
		}
	}

	success = true
	return nil
}

// CreateSession initializes a new SpecSession in StateInit and persists
// it to _session.json inside specDir using the atomic temp-file-and-rename
// pattern. Returns a SessionError on failure.
func CreateSession(specDir, mode, source string) (*SpecSession, error) {
	if specDir == "" {
		return nil, &SessionError{Msg: "specDir must not be empty"}
	}

	// Validate that specDir exists.
	if _, err := os.Stat(specDir); err != nil {
		return nil, &SessionError{
			Msg:   fmt.Sprintf("specDir does not exist: %s", specDir),
			Cause: err,
		}
	}

	session := &SpecSession{
		specDir:            specDir,
		Current:            StateInit,
		Mode:               mode,
		PRDPath:            source,
		AssessmentHistory:  []Assessment{},
		QAExchanges:        []QAExchange{},
		GeneratedArtifacts: []string{},
	}

	if err := session.persistSession(); err != nil {
		return nil, err
	}

	return session, nil
}

// ResumeSession reads _session.json from specDir and reconstructs a
// SpecSession. Returns a SessionError if the file is missing or invalid.
func ResumeSession(specDir string) (*SpecSession, error) {
	sessionPath := filepath.Join(specDir, "_session.json")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, &SessionError{
			Msg:   fmt.Sprintf("failed to read _session.json from %s: %v", specDir, err),
			Cause: err,
		}
	}

	var session SpecSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, &SessionError{
			Msg:   fmt.Sprintf("failed to parse _session.json from %s: %v", specDir, err),
			Cause: err,
		}
	}

	// Validate that the state is a recognized SessionState.
	if !validSessionStates[session.Current] {
		return nil, &SessionError{
			Msg: fmt.Sprintf("unrecognized session state %q in _session.json", session.Current),
		}
	}

	session.specDir = specDir
	return &session, nil
}

// SessionValidationResult holds the result of validating artifacts in a
// spec session directory. Valid is true when no schema or integrity
// errors are found.
type SessionValidationResult struct {
	Valid             bool
	SchemaErrors      []string
	IntegrityErrors   []string
	RepairSuggestions []RepairSuggestion
}

// RepairSuggestion describes a suggested fix for a validation error,
// including whether it can be applied automatically.
type RepairSuggestion struct {
	Description string
	AutoFixable bool
}

// knownArtifacts lists the JSON artifact filenames that LoadSpec expects.
var knownArtifacts = []string{
	"requirements.json",
	"test_spec.json",
	"tasks.json",
}

// GenerateResult holds the result of artifact generation. It contains
// the list of generated artifact names, the validation result, and
// any warnings encountered during generation.
type GenerateResult struct {
	Artifacts  []string                `json:"artifacts"`
	Validation SessionValidationResult `json:"validation"`
	Warnings   []string                `json:"warnings"`
}

// Validate loads the spec from the session's spec directory and runs
// validation, categorizing errors as schema or integrity errors. If
// LoadSpec fails (e.g., missing artifacts), it falls back to loading
// individual JSON artifact files and validating those.
func (s *SpecSession) Validate() (SessionValidationResult, error) {
	// Try the happy path: load the full spec and validate.
	spec, loadErr := afspec.LoadSpec(s.specDir)
	if loadErr == nil {
		vr := spec.Validate()
		return categorizeValidationResult(vr), nil
	}

	// LoadSpec failed — fall back to individual artifact validation.
	return s.validateFallback()
}

// categorizeValidationResult converts an afspec.ValidationResult into a
// SessionValidationResult, splitting entries by category.
func categorizeValidationResult(vr afspec.ValidationResult) SessionValidationResult {
	result := SessionValidationResult{
		Valid:           vr.Valid,
		SchemaErrors:    []string{},
		IntegrityErrors: []string{},
	}

	for _, entry := range vr.Errors {
		msg := formatValidationEntry(entry)
		switch entry.Category {
		case "schema":
			result.SchemaErrors = append(result.SchemaErrors, msg)
		case "integrity":
			result.IntegrityErrors = append(result.IntegrityErrors, msg)
			if entry.Check != "" {
				result.RepairSuggestions = append(result.RepairSuggestions, RepairSuggestion{
					Description: fmt.Sprintf("Fix %s: %s", entry.Check, entry.Message),
					AutoFixable: false,
				})
			}
		default:
			result.IntegrityErrors = append(result.IntegrityErrors, msg)
		}
	}

	return result
}

// formatValidationEntry formats a ValidationEntry into a human-readable string.
func formatValidationEntry(entry afspec.ValidationEntry) string {
	parts := []string{}
	if entry.Artifact != "" {
		parts = append(parts, entry.Artifact)
	}
	if entry.Path != "" {
		parts = append(parts, entry.Path)
	}
	if entry.Check != "" {
		parts = append(parts, entry.Check)
	}
	if len(parts) > 0 {
		return fmt.Sprintf("[%s] %s", strings.Join(parts, ": "), entry.Message)
	}
	return entry.Message
}

// validateFallback loads individual JSON artifact files that are present
// in the spec directory, validates each one, and returns a combined
// SessionValidationResult with Valid=false.
func (s *SpecSession) validateFallback() (SessionValidationResult, error) {
	result := SessionValidationResult{
		Valid:           false,
		SchemaErrors:    []string{},
		IntegrityErrors: []string{},
	}

	foundAny := false

	for _, artifactName := range knownArtifacts {
		artifactPath := filepath.Join(s.specDir, artifactName)
		data, err := os.ReadFile(artifactPath)
		if err != nil {
			// File doesn't exist or is unreadable — record as integrity error.
			if os.IsNotExist(err) {
				result.IntegrityErrors = append(result.IntegrityErrors,
					fmt.Sprintf("missing artifact: %s", artifactName))
			} else {
				result.IntegrityErrors = append(result.IntegrityErrors,
					fmt.Sprintf("cannot read %s: %v", artifactName, err))
			}
			continue
		}

		foundAny = true

		// Validate that the file contains well-formed JSON.
		var raw json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			result.SchemaErrors = append(result.SchemaErrors,
				fmt.Sprintf("invalid JSON in %s: %v", artifactName, err))
			continue
		}

		// Try to unmarshal into the specific type and build a partial spec
		// for schema validation only (cross-file validation is not meaningful
		// for partial specs).
		partialSpec, parseErr := parseArtifactIntoSpec(artifactName, data)
		if parseErr != nil {
			result.SchemaErrors = append(result.SchemaErrors,
				fmt.Sprintf("schema error in %s: %v", artifactName, parseErr))
			continue
		}

		// Run schema-only validation on the partial spec (skip cross-file
		// checks which produce false positives on incomplete specs).
		if partialSpec != nil {
			vr := partialSpec.ValidateSchema()
			partial := categorizeValidationResult(vr)
			result.SchemaErrors = append(result.SchemaErrors, partial.SchemaErrors...)
			result.IntegrityErrors = append(result.IntegrityErrors, partial.IntegrityErrors...)
			result.RepairSuggestions = append(result.RepairSuggestions, partial.RepairSuggestions...)
		}
	}

	if !foundAny {
		result.IntegrityErrors = append(result.IntegrityErrors,
			"no artifact files found in spec directory")
	}

	return result, nil
}

// parseArtifactIntoSpec parses a single artifact's data into a partial Spec
// containing only that artifact, suitable for validation.
func parseArtifactIntoSpec(artifactName string, data []byte) (*afspec.Spec, error) {
	spec := &afspec.Spec{}
	switch artifactName {
	case "requirements.json":
		var req afspec.RequirementsV1Json
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		spec.Requirements = &req
	case "test_spec.json":
		var ts afspec.TestSpecV1Json
		if err := json.Unmarshal(data, &ts); err != nil {
			return nil, err
		}
		spec.TestSpec = &ts
	case "tasks.json":
		var tasks afspec.TasksV1Json
		if err := json.Unmarshal(data, &tasks); err != nil {
			return nil, err
		}
		spec.Tasks = &tasks
	default:
		return nil, nil
	}
	return spec, nil
}

// Render loads the spec from the session's spec directory and renders
// the artifacts. When combined is true, it delegates to
// afspec.RenderCombined and returns the combined string. When combined
// is false, it delegates to afspec.RenderIndividual and returns a
// map[string]string keyed by artifact name. If LoadSpec fails, it
// falls back to rendering only available artifact files.
func (s *SpecSession) Render(combined bool) (any, error) {
	// Try the happy path: load the full spec and render.
	spec, loadErr := afspec.LoadSpec(s.specDir)
	if loadErr == nil {
		if combined {
			return spec.RenderCombined(), nil
		}
		return spec.RenderIndividual(), nil
	}

	// LoadSpec failed — fall back to rendering available artifacts.
	return s.renderFallback(combined)
}

// renderFallback loads individual artifact files that are present in the
// spec directory, builds a partial Spec, and renders it.
func (s *SpecSession) renderFallback(combined bool) (any, error) {
	spec := s.loadPartialSpec()

	hasArtifacts := spec.Requirements != nil || spec.TestSpec != nil ||
		spec.Tasks != nil || spec.Architecture != ""

	if !hasArtifacts {
		if combined {
			return "", nil
		}
		return map[string]string{}, nil
	}

	if combined {
		return spec.RenderCombined(), nil
	}

	return spec.RenderIndividual(), nil
}

// loadPartialSpec attempts to load whatever artifact files are present in
// the spec directory into a partial Spec.
func (s *SpecSession) loadPartialSpec() *afspec.Spec {
	spec := &afspec.Spec{}

	// Try to load PRD body.
	prdPath := filepath.Join(s.specDir, "prd.md")
	if prdData, err := os.ReadFile(prdPath); err == nil {
		// Extract body after frontmatter (everything after second "---").
		body := extractPRDBody(string(prdData))
		spec.PRDBody = body
	}

	// Try to load each JSON artifact.
	reqPath := filepath.Join(s.specDir, "requirements.json")
	if data, err := os.ReadFile(reqPath); err == nil {
		var req afspec.RequirementsV1Json
		if json.Unmarshal(data, &req) == nil {
			spec.Requirements = &req
		}
	}

	tsPath := filepath.Join(s.specDir, "test_spec.json")
	if data, err := os.ReadFile(tsPath); err == nil {
		var ts afspec.TestSpecV1Json
		if json.Unmarshal(data, &ts) == nil {
			spec.TestSpec = &ts
		}
	}

	tasksPath := filepath.Join(s.specDir, "tasks.json")
	if data, err := os.ReadFile(tasksPath); err == nil {
		var tasks afspec.TasksV1Json
		if json.Unmarshal(data, &tasks) == nil {
			spec.Tasks = &tasks
		}
	}

	// Try to load architecture.
	archPath := filepath.Join(s.specDir, "architecture.md")
	if data, err := os.ReadFile(archPath); err == nil {
		spec.Architecture = string(data)
	}

	return spec
}

// extractPRDBody extracts the Markdown body from a PRD file, stripping
// the YAML frontmatter delimited by "---" lines.
func extractPRDBody(content string) string {
	// Find the opening "---"
	_, afterOpen, found := strings.Cut(content, "---")
	if !found {
		return content
	}
	// Find the closing "---"
	_, afterClose, found := strings.Cut(afterOpen, "---")
	if !found {
		return content
	}
	return strings.TrimSpace(afterClose)
}

// Assess transitions to StateAssessing, creates a SpecAgent (or uses an
// injected one), loads the spec landscape from sibling specs, calls
// AssessPRD, appends the result to the assessment history, and persists
// session state. Returns (Assessment, nil) on success.
//
// On error, persists the error as lastError in session state without
// appending to assessment history, then returns (Assessment{}, error).
func (s *SpecSession) Assess(ctx context.Context) (Assessment, error) {
	// TODO: implement
	return Assessment{}, fmt.Errorf("Assess: not implemented")
}

// Refine transitions to StateRefining, calls RefinePRD with the current
// PRD text, user answers, and latest assessment; updates the PRD file on
// disk with the refined text; appends the new assessment to history;
// records a QA exchange with assessment_index, answers, and UTC
// timestamp; and persists session state.
//
// On error, persists the error as lastError without updating the PRD
// file, without appending to assessment history, and without recording
// a QA exchange.
func (s *SpecSession) Refine(ctx context.Context, answers map[string]string) (Assessment, error) {
	// TODO: implement
	return Assessment{}, fmt.Errorf("Refine: not implemented")
}

// Generate transitions to StateGenerating immediately, checks for
// existing artifact files to skip already-generated artifacts (partial
// failure recovery), calls GenerateArtifacts with an OnArtifact callback
// that writes each artifact to disk via afspec MarshalJSON and records
// it in generatedArtifacts, transitions to StateGenerated after all
// artifacts are generated, runs Validate, and returns a GenerateResult.
//
// On error, does not transition to StateGenerated, persists the error
// as lastError, and returns (GenerateResult{}, error).
func (s *SpecSession) Generate(ctx context.Context) (GenerateResult, error) {
	// TODO: implement
	return GenerateResult{}, fmt.Errorf("Generate: not implemented")
}

