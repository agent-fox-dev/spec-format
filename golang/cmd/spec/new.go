package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newNewCmd creates the "spec new" subcommand which creates a new spec
// from a PRD file.
func newNewCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "new SPEC_PATH",
		Short: "Create a new spec from a PRD file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = name
			return fmt.Errorf("spec new: not implemented")
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "spec name (must match [a-z][a-z0-9_]*)")

	return cmd
}
