package mirror

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/mirror"
)

// newSerializeCmd represents the mirror serialize command.
func newBatchSerializeCmd(tool *actions.Action) *cobra.Command {
	action := &actions.BatchSerialize{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "batch-serialize TRACKERFILE SYNCREPO",
		Short: "", //TODO
		Long: `IMAGE is a reference to an OCI image index to use as the source.  All the images in the image index will be sent to DEST.
DEST is a tar file or a tape archive.  If it is a tape archive better performance can be had by setting --buffer-size=1Gi or larger.  The tar file can also be written to the tape after serialization is completed (see "ace-dt util mbuffer").
EXISTING-IMAGE(s) are images that we use to extract blob references from to determine if we need to serialize the blob.
`,
		Example: `ace-dt mirror batch-serialize sync/data/record_keeping.csv sync/data`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1])
			})
		},
	}
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
