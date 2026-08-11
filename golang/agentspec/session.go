package agentspec

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
	// TODO: implement
	return nil
}

// AcceptPRD transitions the session state to StatePRDAccepted and
// persists the updated state to _session.json atomically. Returns a
// SessionError if the current state is not StateAssessing or StateRefining.
func (s *SpecSession) AcceptPRD() error {
	// TODO: implement
	return nil
}

// PendingQuestions returns the questions from the most recent Assessment
// that have not yet been answered via QA exchanges. Returns an empty
// slice if there are no pending questions or the assessment history is empty.
func (s *SpecSession) PendingQuestions() []map[string]any {
	// TODO: implement
	return nil
}

// CreateSession initializes a new SpecSession in StateInit and persists
// it to _session.json inside specDir using the atomic temp-file-and-rename
// pattern. Returns a SessionError on failure.
func CreateSession(specDir, mode, source string) (*SpecSession, error) {
	// TODO: implement
	return nil, nil
}

// ResumeSession reads _session.json from specDir and reconstructs a
// SpecSession. Returns a SessionError if the file is missing or invalid.
func ResumeSession(specDir string) (*SpecSession, error) {
	// TODO: implement
	return nil, nil
}
