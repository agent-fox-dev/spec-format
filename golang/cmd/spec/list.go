package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newListCmd creates the "spec list" subcommand which lists all specs
// and their current states.
func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all specs and their current states",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("spec list: not implemented")
		},
	}

	return cmd
}
