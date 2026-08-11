package spec

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCampaignCmd creates the "spec campaign" subcommand which creates
// a named campaign to group related specs.
func newCampaignCmd() *cobra.Command {
	var path string
	var name string
	var description string

	cmd := &cobra.Command{
		Use:   "campaign",
		Short: "Create a named campaign to group related specs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = path
			_ = name
			_ = description
			return fmt.Errorf("spec campaign: not implemented")
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "path for the campaign directory")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name for the campaign")
	cmd.Flags().StringVar(&description, "description", "", "description for the campaign")

	// --path and --name are required.
	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("name")

	return cmd
}
