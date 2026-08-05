package afspec

// ComputeIntentHash extracts the ## Intent section from a PRD body string,
// normalizes it, computes its SHA-256 hash, and returns the hex-encoded
// hash string. The hash is stable: identical intent text always produces
// the same hash.
//
// Returns an IntentError if the body does not contain a ## Intent section
// or if the body is empty.
func ComputeIntentHash(body string) (string, error) {
	panic("not implemented")
}
