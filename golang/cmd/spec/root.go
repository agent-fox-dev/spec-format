package spec

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags.
var version = "dev"

// newRootCmd creates the root cobra command with all global flags.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Spec CLI — manage agentspec specifications",
		// TODO: implement root command logic
	}
	return cmd
}

// Execute runs the root command. This is the binary entry point.
func Execute() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
