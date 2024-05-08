/* Command delete
 */

package bottle

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/oci"
	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions/bottle"
)

// deleteCmd represents the delete command.
func newDeleteCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Delete{Action: tool}

	cmd := &cobra.Command{
		GroupID: "remote",
		Use:     "delete BOTTLE_REFERENCE",
		Short:   "Remove a bottle from remote oci storage",
		Long: `Remove a data bottle from remote oci storage, based on
the data bottle name and tag.
This will remove the tag's manifest, which will orphan
the data bottle's blobs, or parts, which will be deleted
upon a routine garbage collection cycle.
	
A bottle reference uses one of the forms
  by tag                <registry>/<repository>/<name>:<tag>
  by name (latest tag)  <registry>/<repository>/<name>
  by digest             <registry>/<repository>/<name>@sha256:<sha>
`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: oci.RefCompletion(action.DataTool),
		RunE: func(cmd *cobra.Command, args []string) error {
			action.Ref = args[0]
			return action.Run(cmd.Context())
		},
	}

	return cmd
}
