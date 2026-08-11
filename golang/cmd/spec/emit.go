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
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("emit: %w", err)
	}
	b = append(b, '\n')

	_, writeErr := os.Stdout.Write(b)
	if writeErr != nil {
		if isBrokenPipe(writeErr) {
			return nil
		}
		return writeErr
	}
	return nil
}

// emitOK merges the provided key-value pairs with {"ok": true} and
// calls emit to write the combined object to stdout.
// It accepts pairs as alternating key, value arguments.
func emitOK(pairs ...any) error {
	m := map[string]any{"ok": true}
	for i := 0; i+1 < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return fmt.Errorf("emitOK: key at position %d is not a string", i)
		}
		m[key] = pairs[i+1]
	}
	return emit(m)
}

// emitError writes {"ok": false, "error": msg} to stdout.
// Used in agent mode to wrap errors as JSON.
func emitError(msg string) error {
	return emit(map[string]any{
		"ok":    false,
		"error": msg,
	})
}

// isBrokenPipe returns true if the error is a broken pipe error.
func isBrokenPipe(err error) bool {
	if errors.Is(err, syscall.EPIPE) {
		return true
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, syscall.EPIPE)
	}
	return false
}
