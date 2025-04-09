package cli

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
)

// logoutCmd represents the logout command.
func newLogoutCmd(tool *actions.DataTool) *cobra.Command {
	action := &actions.Logout{DataTool: tool}

	logoutCmd := &cobra.Command{
		GroupID: "setup",
		Use:     "logout REGISTRY",
		Short:   "Logout from a remote registry",
		Long:    `Logout from a remote registry. This will remove an entry from your ~/.docker/config.json.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], cmd.OutOrStdout())
		},
	}

	return logoutCmd
}
