package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newRefineCmd creates the "spec refine" subcommand which iteratively
// refines a PRD through AI-driven Q&A.
func newRefineCmd() *cobra.Command {
	var answers string
	var force bool

	cmd := &cobra.Command{
		Use:   "refine SPEC",
		Short: "Iteratively refine a PRD through AI-driven Q&A",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = answers
			_ = force
			return fmt.Errorf("spec refine: not implemented")
		},
	}

	cmd.Flags().StringVar(&answers, "answers", "", "path to answers JSON file (use '-' for stdin)")
	cmd.Flags().BoolVar(&force, "force", false, "reset session state before refining")

	return cmd
}
