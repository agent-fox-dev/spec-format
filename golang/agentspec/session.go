package agentspec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
// point in time.
type QAExchange struct {
	Answers []map[string]any `json:"answers"`
}

// LastError records the most recent error encountered during a session
// operation.
type LastError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// SpecSession represents the stateful session for authoring a single spec,
// persisted to _session.json inside a spec directory.
type SpecSession struct {
	specDir            string
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

// Validate loads the spec from the session's spec directory and runs
// validation, categorizing errors as schema or integrity errors. If
// LoadSpec fails (e.g., missing artifacts), it falls back to loading
// individual JSON artifact files and validating those.
func (s *SpecSession) Validate() (SessionValidationResult, error) {
	// TODO: implement
	return SessionValidationResult{}, nil
}

// Render loads the spec from the session's spec directory and renders
// the artifacts. When combined is true, it delegates to
// afspec.RenderCombined and returns the combined string. When combined
// is false, it delegates to afspec.RenderIndividual and returns a
// map[string]string keyed by artifact name. If LoadSpec fails, it
// falls back to rendering only available artifact files.
func (s *SpecSession) Render(combined bool) (any, error) {
	// TODO: implement
	return nil, nil
}
