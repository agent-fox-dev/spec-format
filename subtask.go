package afspec

// TransitionSubtask looks up a subtask by ID, verifies that the
// transition from the subtask's current state to the target state is
// in the allowed transition table, updates the subtask state, and
// returns a new Tasks copy. The receiver is not modified.
//
// Allowed subtask transitions:
//   - pending            -> queued
//   - pending            -> dropped
//   - queued             -> in_progress
//   - queued             -> pending
//   - queued             -> dropped
//   - in_progress        -> done
//   - in_progress        -> pending_reevaluation
//   - done               -> pending_reevaluation
//   - pending_reevaluation -> pending
//   - pending_reevaluation -> dropped
//
// Returns a LifecycleError if the transition is not allowed or the
// subtask ID does not exist.
func (t *TasksV1Json) TransitionSubtask(subtaskID string, target string) (*TasksV1Json, error) {
	panic("not implemented")
}

// CompleteSubtaskStates sets every subtask in each specified group to
// the done state, bypassing the state machine transition rules.
// Returns a new Tasks copy. The receiver is not modified.
//
// Missing group IDs are silently skipped.
func (t *TasksV1Json) CompleteSubtaskStates(groupIDs []int) (*TasksV1Json, error) {
	panic("not implemented")
}

// ResetSubtaskStates sets every subtask in each specified group to the
// pending state, bypassing the state machine transition rules.
// Returns a new Tasks copy. The receiver is not modified.
//
// Missing group IDs are silently skipped.
func (t *TasksV1Json) ResetSubtaskStates(groupIDs []int) (*TasksV1Json, error) {
	panic("not implemented")
}
