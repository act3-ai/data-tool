package bottle

import (
	"github.com/spf13/cobra"

	actions "github.com/act3-ai/data-tool/internal/actions/bottle"
)

// editCmd represents the edit command.
func newEditCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Edit{Action: tool}

	cmd := &cobra.Command{
		GroupID: "metadata",
		Use:     "edit",
		Short:   "Open a data bottle configuration in the system editor",
		Long: `Open a data bottle configuration in the system editor
	
The editor opened is either the $EDITOR environment variable, or 
  vim if no editor is specified there.
	
By default, the current directory is searched for a data bottle configuration,
  but a path to a data bottle may be specified for an alternate location..
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return cmd
}
