/*
Copyright © 2020 ACT3 DevSecOps

*/

package bottle

import (
	"github.com/spf13/cobra"

	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// statusCmd represents the status command.
func newStatusCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Status{Action: tool}

	cmd := &cobra.Command{
		// GroupID: "advanced",
		Use:   "status",
		Short: "Show status of items in the data bottle",
		Long:  `Show status of items in the data bottle, including whether items are cached, changed, new, or deleted. The current working directory is used as the source of the bottle, but a path can be provided to inspect an alternate location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVarP(&action.Details, "details", "D", false, "Show file paths within sub directories for changed files")

	return cmd
}
