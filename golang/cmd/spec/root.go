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

	// Register subcommands.
	cmd.AddCommand(newNewCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newRefineCmd())
	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newRenderCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newLintCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newCampaignCmd())

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
