package git

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/oci"
	"git.act3-ace.com/ace/data/tool/internal/actions/git"
)

// newListRefsCmd creates a new cobra.Command for the list-refs subcommand.
func newListRefsCmd(base *git.Action) *cobra.Command {
	action := &git.ListRefs{Action: base}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "list-refs OCI_REFERENCE",
		Short: "List all references in an OCI sync artifact.",
		Long:  `Lists all head and tag references in an OCI sync artifact along with the commits they reference.`,

		Example: `List all references at reg.example.com/my/libgit2:sync:
		$ ace-dt git list-refs reg.example.com/my/libgit2:sync`,

		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: oci.RefCompletion(action.DataTool),
		RunE: func(cmd *cobra.Command, args []string) error {

			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				action.Repo = args[0]
				return action.Run(ctx)
			})
		},
	}

	return cmd
}
