package spec

import (
	"io"
)

// StatusSpinner shows an animated spinner with a message on stderr
// during long-running operations. When quiet mode is active or
// AF_AGENT=1 is set, all methods are no-ops.
type StatusSpinner struct {
	message string
	writer  io.Writer
	quiet   bool
	isTTY   bool
}

// NewStatusSpinner creates a new StatusSpinner. If quiet is true,
// all methods are no-ops. The writer parameter specifies where
// spinner output goes (typically os.Stderr).
func NewStatusSpinner(message string, writer io.Writer, quiet bool, isTTY bool) *StatusSpinner {
	return &StatusSpinner{
		message: message,
		writer:  writer,
		quiet:   quiet,
		isTTY:   isTTY,
	}
}

// Start begins the spinner animation.
func (s *StatusSpinner) Start() {
	// TODO: implement
}

// Stop halts the spinner animation and cleans up.
func (s *StatusSpinner) Stop() {
	// TODO: implement
}

// Update changes the spinner message text.
func (s *StatusSpinner) Update(message string) {
	// TODO: implement
}

// Log prints a permanent line above the spinner.
func (s *StatusSpinner) Log(message string) {
	// TODO: implement
}
