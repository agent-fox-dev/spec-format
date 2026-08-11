package spec

import (
	"io"
	"os"
)

// isAgentMode returns true when AF_AGENT=1 is set in the environment.
func isAgentMode() bool {
	return os.Getenv("AF_AGENT") == "1"
}

// shouldShowBanner determines whether the ASCII art banner should be
// displayed. The banner is suppressed when:
//   - --quiet is set
//   - AF_AGENT=1 is active
//   - the subcommand is validate, status, or list
//   - --json appears in the arguments
func shouldShowBanner(quiet bool, subcmd string, args []string) bool {
	if quiet || isAgentMode() {
		return false
	}
	switch subcmd {
	case "validate", "status", "list":
		return false
	}
	for _, a := range args {
		if a == "--json" {
			return false
		}
	}
	return true
}

// printBanner writes the ASCII art banner to the given writer.
func printBanner(w io.Writer, ver string) {
	// TODO: implement ASCII art banner
}
