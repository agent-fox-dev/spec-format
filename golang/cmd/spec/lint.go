package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newLintCmd creates the "spec lint" subcommand which runs lint checks
// across all specs and reports findings as JSON.
func newLintCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint all specs for quality and consistency issues",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = all
			return fmt.Errorf("spec lint: not implemented")
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "include fully-implemented specs in the lint run")

	return cmd
}
