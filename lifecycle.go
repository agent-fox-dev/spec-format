package afspec

import (
	"fmt"
	"os"
	"path/filepath"
)

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

	// Create a shallow copy with the new status.
	// The receiver is not modified — Transition returns a new Spec.
	newSpec := *s
	newSpec.Status = target

	// Persist via saveToDisk which bypasses the public Save lifecycle guard,
	// since Transition legitimately writes sealed/superseded/archived specs.
	if err := newSpec.saveToDisk(dir); err != nil {
		return nil, err
	}

	return &newSpec, nil
}

// Supersede transitions a sealed spec to the superseded state, prepends
// a deprecation banner referencing the superseding spec ID to the PRD
// body, and persists the updated spec to disk. Returns a new Spec copy.
//
// Returns a LifecycleError if the spec is not in the sealed state.
func (s *Spec) Supersede(supersedingSpecID, dir string) (*Spec, error) {
	if s.Status != "sealed" {
		return nil, &LifecycleError{
			Msg: fmt.Sprintf("cannot supersede spec in %q state; must be sealed", s.Status),
			Err: &SpecError{Msg: fmt.Sprintf("cannot supersede spec in %q state; must be sealed", s.Status)},
		}
	}

	// Create a shallow copy with superseded status and deprecation banner.
	newSpec := *s
	newSpec.Status = "superseded"

	// Prepend deprecation banner to the PRD body.
	banner := fmt.Sprintf("> **DEPRECATED:** This specification has been superseded by spec %s. "+
		"Refer to the replacement specification for current requirements.\n\n", supersedingSpecID)
	newSpec.PRDBody = banner + s.PRDBody

	if err := newSpec.saveToDisk(dir); err != nil {
		return nil, err
	}

	return &newSpec, nil
}

// MoveToArchive transitions a spec to the archived state and moves the
// spec directory to {root}/archive/ using an atomic rename or
// copy-then-delete.
//
// Creates the archive directory if it does not exist.
// Returns a SaveError if the archive directory already contains a
// directory with the same name.
func MoveToArchive(specDir, root string) error {
	// Load the spec from the source directory.
	spec, err := LoadSpec(specDir)
	if err != nil {
		return &SaveError{
			Msg: fmt.Sprintf("cannot load spec for archiving: %s", err),
			Err: &SpecError{Msg: fmt.Sprintf("cannot load spec for archiving: %s", err)},
		}
	}

	// Transition to archived state if not already archived.
	if spec.Status != "archived" {
		_, err = spec.Transition("archived", specDir)
		if err != nil {
			return err
		}
	}

	// Create the archive directory if it does not exist.
	archiveDir := filepath.Join(root, "archive")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return &SaveError{
			Msg: fmt.Sprintf("cannot create archive directory: %s", err),
			Err: &SpecError{Msg: fmt.Sprintf("cannot create archive directory: %s", err)},
		}
	}

	// Check for conflict — refuse to overwrite existing archive entry.
	destDir := filepath.Join(archiveDir, filepath.Base(specDir))
	if _, err := os.Stat(destDir); err == nil {
		return &SaveError{
			Msg: fmt.Sprintf("archive conflict: %s already exists", destDir),
			Err: &SpecError{Msg: fmt.Sprintf("archive conflict: %s already exists", destDir)},
		}
	}

	// Move the spec directory to the archive.
	if err := os.Rename(specDir, destDir); err != nil {
		return &SaveError{
			Msg: fmt.Sprintf("cannot move spec to archive: %s", err),
			Err: &SpecError{Msg: fmt.Sprintf("cannot move spec to archive: %s", err)},
		}
	}

	return nil
}
