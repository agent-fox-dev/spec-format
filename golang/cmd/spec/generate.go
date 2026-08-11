package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newGenerateCmd creates the "spec generate" subcommand which generates
// specification artifacts from an accepted PRD.
func newGenerateCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "generate SPEC",
		Short: "Generate specification artifacts from an accepted PRD",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = force
			return fmt.Errorf("spec generate: not implemented")
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "delete existing artifacts before regenerating")

	return cmd
}
