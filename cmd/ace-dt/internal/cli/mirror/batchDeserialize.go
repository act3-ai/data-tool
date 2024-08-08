package mirror

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/mirror"
)

// newSerializeCmd represents the mirror serialize command.
func newBatchDeserializeCmd(tool *actions.Action) *cobra.Command {
	action := &actions.BatchDeserialize{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "batch-deserialize SYNC-REPOSITORY DESTINATION",
		Short: "A command that deserializes all of the blobs in tar files located in the SYNC-REPOSITORY to the DESTINATION.",
		Long: `SYNC-REPOSITORY is the repository holding the tar files to be deserialized. All tar files will be deserialized to DESTINATION.
DESTINATION is a remote repository WITHOUT a tag. Tags will be automatically generated based off of the image name within the tar file name.
For example, given a file "SYNC-REPOSITORY/0-image1.tar", the blobs will be deserilaized to DESTINATION and tagged as "image1".
`,
		Example: `ace-dt mirror batch-deserialize sync/data/ registry.example.com/image`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1])
			})
		},
	}
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)
	cmd.Flags().StringVar(&action.SuccessfulSyncFile, "sync-filename", "successful-syncs.csv", "used to override the default sync-file name.")

	return cmd
}
