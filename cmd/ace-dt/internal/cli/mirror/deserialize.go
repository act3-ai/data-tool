package mirror

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "github.com/act3-ai/data-tool/internal/actions/mirror"
)

// newDeserializeCmd represents the mirror deserialize command.
func newDeserializeCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Deserialize{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "deserialize SOURCE-FILE IMAGE",
		Short: "Deserializes OCI images from SOURCE-FILE and writes them to IMAGE.",
		Long: `SOURCE-FILE is a tar file or a tape archive to read serialized data.
IMAGE is an OCI image reference to write the data to.  It is expected to have a tag specified.

If you see a "Cannot Allocate Memory error" when using a tape as the input, you probably forgot to set the block size with "--block-size" to the value that was used to write the blocks.  In the case of a tape configured wi) use in the case of large block sizes where Cannot Allocate Memory error is present)`,
		Example: `ace-dt mirror deserialize /dev/nst0 reg.other.com/project/proj:sync-45`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1])
			})
		},
	}

	cmd.Flags().BoolVar(&action.Strict, "strict", false, "Enable strict checking mode.  This will often only work if the tar stream was generated by \"ace-dt mirror serialize\".")

	cmd.Flags().BoolVar(&action.DryRun, "dry-run", false, "Enable dry run mode. This will consume the tar file without sending data to a registry.")
	cmd.Flags().IntVar(&action.BufferSize, "block-size", 0, "Size of read buffer.  If 0 then no buffer is used.")
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}

/* Can test this with...

# When doing tmp.tar is produced by serialize
go run ./cmd/ace-dt mirror deserialize log/tmp.tar zot.lion.act3-ace.ai/$PROJECT/dest:sync-1 --strict

# otherwise we can do it manually
ace-dt oci pull reg.git.act3-ace.com/ace/data/tool/ace-dt:v1.0.7 reg.git.act3-ace.com/ace/data/tool/ace-dt@sha256:a88cb2d6e96cd261b4718adb9497bb3cd6ad70edc4db6a927a1e20e299eed2ac reg.git.act3-ace.com/ace/hub/api log/oci-dir
tar cvf log/oci-dir.tar -C log/oci-dir .
# or create a tar with the worst ordering
tar cvf log/oci-dir.tar -C log/oci-dir blobs index.json oci-layout
# or a random ordering
( cd log/oci-dir && find * | sort -R | tar cvf ../oci-dir2.tar -T - )
# Run deserialize without strict mode (strict mode should fail because the files are in the wrong order)
go run ./cmd/ace-dt mirror deserialize log/oci-dir.tar zot.lion.act3-ace.ai/$PROJECT/dest2:sync-1
*/
