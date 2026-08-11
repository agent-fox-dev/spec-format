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

// SpecSession represents the stateful session for authoring a single spec,
// persisted to _session.json inside a spec directory.
type SpecSession struct {
	SpecDir  string       `json:"spec_dir"`
	SpecName string       `json:"spec_name"`
	SpecID   string       `json:"spec_id"`
	Current  SessionState `json:"state"`
	Mode     string       `json:"mode"`
}

// State returns the current SessionState of the session.
func (s *SpecSession) State() SessionState {
	return s.Current
}

// CreateSession initializes a new SpecSession in StateInit and persists
// it to _session.json inside specDir. Returns a SessionError on failure.
func CreateSession(specDir, specName, specID, mode string) (*SpecSession, error) {
	// TODO: implement
	return nil, nil
}

// ResumeSession reads _session.json from specDir and reconstructs a
// SpecSession. Returns a SessionError if the file is missing or invalid.
func ResumeSession(specDir string) (*SpecSession, error) {
	// TODO: implement
	return nil, nil
}
