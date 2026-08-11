package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newValidateCmd creates the "spec validate" subcommand which validates
// one or all specs for structural correctness.
func newValidateCmd() *cobra.Command {
	var cross bool
	var short bool

	cmd := &cobra.Command{
		Use:   "validate [SPEC]",
		Short: "Validate one or all specs for structural correctness",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cross
			_ = short
			return fmt.Errorf("spec validate: not implemented")
		},
	}

	cmd.Flags().BoolVar(&cross, "cross", false, "run cross-spec interface consistency checks")
	cmd.Flags().BoolVar(&short, "short", false, "emit condensed output with only valid/error_count/warning_count")

	return cmd
}
