package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"
)

// emit writes data as pretty-printed JSON (2-space indent) to stdout.
// It silently handles broken pipe errors.
func emit(data any) error {
	_ = data
	return fmt.Errorf("emit: not implemented")
}

// emitOK merges the provided key-value pairs with {"ok": true} and
// calls emit to write the combined object to stdout.
// It accepts pairs as alternating key, value arguments.
func emitOK(pairs ...any) error {
	_ = pairs
	return fmt.Errorf("emitOK: not implemented")
}

// isBrokenPipe returns true if the error is a broken pipe error.
func isBrokenPipe(err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, syscall.EPIPE)
	}
	return errors.Is(err, syscall.EPIPE)
}

// Ensure json import is used (will be needed by implementation).
var _ = json.Marshal
