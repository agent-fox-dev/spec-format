package afspec

import "fmt"

// validTransitions maps each spec status to the set of statuses it can transition to.
var validTransitions = map[string][]string{
	"draft":  {"active"},
	"active": {"sealed", "superseded"},
	"sealed": {"superseded"},
}

// ValidTransition checks the transition table and returns true if the
// transition from current to target state is allowed, false otherwise.
// This is a pure function with no side effects.
//
// Allowed transitions:
//   - draft    -> active
//   - active   -> sealed
//   - active   -> superseded
//   - sealed   -> superseded
//   - any      -> archived
func ValidTransition(current, target string) bool {
	if target == "archived" {
		return true
	}
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

// Transition updates the spec's status field to the target state,
// persists the updated spec to disk, and returns a new Spec copy with
// the updated status. The receiver is not modified.
//
// Returns a LifecycleError if the transition is not allowed.
// Returns a SaveError if disk persistence fails.
func (s *Spec) Transition(target, dir string) (*Spec, error) {
	if !ValidTransition(s.Status, target) {
		return nil, &LifecycleError{
			Msg: fmt.Sprintf("invalid transition from %q to %q", s.Status, target),
			Err: &SpecError{Msg: fmt.Sprintf("invalid transition from %q to %q", s.Status, target)},
		}
	}

	panic("not implemented")
}

// Supersede transitions a sealed spec to the superseded state, prepends
// a deprecation banner referencing the superseding spec ID to the PRD
// body, and persists the updated spec to disk. Returns a new Spec copy.
//
// Returns a LifecycleError if the spec is not in the sealed state.
func (s *Spec) Supersede(supersedingSpecID, dir string) (*Spec, error) {
	panic("not implemented")
}

// MoveToArchive transitions a spec to the archived state and moves the
// spec directory to {root}/archive/ using an atomic rename or
// copy-then-delete.
//
// Creates the archive directory if it does not exist.
// Returns a SaveError if the archive directory already contains a
// directory with the same name.
func MoveToArchive(specDir, root string) error {
	panic("not implemented")
}
