package cli

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
)

// loginCmd represents the login command.
func newLoginCmd(tool *actions.DataTool) *cobra.Command {
	action := &actions.Login{DataTool: tool}
	cmd := &cobra.Command{
		GroupID: "setup",
		Use:     "login REGISTRY",
		Short:   "Provide authentication credentials for OCI push and pull operations",
		Long: `Provide authentication credentials for OCI push and pull operations
This will prompt for a user name and password, and will authenticate to the provided registry. If successful, the credentials will be used for future interactions with that registry by adding an entry to your ~/.docker/config.json.

This supports storing credentials in credential helpers for increased security.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVarP(&action.Username, "username", "u", "", "username credential for login")
	cmd.Flags().StringVarP(&action.Password, "password", "p", "", "password credential for login")
	cmd.Flags().BoolVar(&action.DisableAuthCheck, "no-auth-check", false, "Skips checking the credentials against the registry")
	return cmd
}
