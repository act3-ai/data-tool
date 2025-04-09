package cli

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/go-common/pkg/secret"
)

// loginCmd represents the login command.
func newLoginCmd(tool *actions.DataTool) *cobra.Command {
	action := &actions.Login{DataTool: tool, Password: &secret.Value{}}
	var cmd = &cobra.Command{
		GroupID: "setup",
		Use:     "login REGISTRY",
		Short:   "Provide authentication credentials for OCI push and pull operations",
		Long: `Provide authentication credentials for OCI push and pull operations
This will prompt for a user name and password, and will authenticate to the provided registry. If successful, the credentials will be used for future interactions with that registry by adding an entry to your ~/.docker/config.json.

This supports storing credentials in credential helpers for increased security. Furthermore, it does not support directly providing passwords, whether plaintext or through an expanded environment variable.

Example - Password from stdin:
  ace-dt login -u username --password-stdin reg.example.com

Example - Password from the environment variable $PASSWORD:
  ace-dt login -u username -p=env:PASSWORD
or the shorthand version:
  ace-dt login -u username -p=PASSWORD

Example - Password from command:
  ace-dt login -u username -p="cmd:secret-tool lookup username exampleuser server reg.example.com"

Example - Password from file:
  ace-dt login -u username -p=file:/absolute/path/pass.txt
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVarP(&action.Username, "username", "u", "", "username credential for login")
	cmd.Flags().VarP(action.Password, "password", "p", "source of password credential for login, not the plaintext password itself")
	cmd.Flags().BoolVar(&action.PassStdin, "password-stdin", false, "read password credential from stdin")
	cmd.MarkFlagsMutuallyExclusive("password", "password-stdin")
	cmd.Flags().BoolVar(&action.DisableAuthCheck, "no-auth-check", false, "Skips checking the credentials against the registry")
	return cmd
}
