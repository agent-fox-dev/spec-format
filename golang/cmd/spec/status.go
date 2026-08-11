package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newStatusCmd creates the "spec status" subcommand which checks the
// current state of a spec session.
func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status SPEC",
		Short: "Check the current state of a spec session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("spec status: not implemented")
		},
	}

	return cmd
}
