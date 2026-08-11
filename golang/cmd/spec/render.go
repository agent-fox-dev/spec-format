package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newRenderCmd creates the "spec render" subcommand which renders spec
// artifacts as markdown or JSON.
func newRenderCmd() *cobra.Command {
	var jsonOutput bool
	var combined bool

	cmd := &cobra.Command{
		Use:   "render SPEC",
		Short: "Render spec artifacts as markdown or JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = jsonOutput
			_ = combined
			return fmt.Errorf("spec render: not implemented")
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON envelope")
	cmd.Flags().BoolVar(&combined, "combined", false, "combine all artifacts into a single document")

	return cmd
}
