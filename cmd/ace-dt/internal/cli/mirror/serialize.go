package mirror

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/flag"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
)

// newSerializeCmd represents the mirror serialize command.
func newSerializeCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Serialize{Action: tool}
	uiOptions := ui.Options{}
	mbufOpts := flag.MemoryBufferOptions{
		BlockSize: 1024 * 1024, // 1MiB is the default block size
	}
	var existingCheckpoints []string

	cmd := &cobra.Command{
		Use:   "serialize IMAGE DEST [EXISTING-IMAGE...]",
		Short: "Serialize image data from IMAGE to DEST assuming that all blobs in the EXISTING-IMAGE(s) do not need to be sent.",
		Long: `IMAGE is a reference to an OCI image index to use as the source.  All the images in the image index will be sent to DEST.
DEST is a tar file or a tape archive.  If it is a tape archive better performance can be had by setting --buffer-size=1Gi or larger.  The tar file can also be written to the tape after serialization is completed (see "ace-dt util mbuffer").
EXISTING-IMAGE(s) are images that we use to extract blob references from to determine if we need to serialize the blob.

Checkpointing can be accomplished by added the --checkpoint flag.
If serialize fails for any reason, provide the --resume-from-checkpoint flag with the checkpoint file from the previous run.  Also inspect the media (file size or tape archive position, to determine a conservative (lower value is more conservative) for the number of bytes that were properly written to the media and provide that to --resume-from-offset.`,
		Example: `ace-dt mirror serialize reg.example.com/project/repo:sync-45 /dev/nst0 reg.example.com/project/repo:complete`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, bs, hwm := mbufOpts.Options()

			// TODO it would be nice if this could easily be hoisted out of here and into a pflag.Valuer and pflag.SliceValue compatible interface
			// iterate through the checkpoint files
			for _, cp := range existingCheckpoints {
				// parse the slice elements, splicing at the ":" character
				v := strings.Split(cp, ":")
				if len(v) != 2 {
					return fmt.Errorf(`expected a colon delimited checkpoint (e.g., "path/to/file.txt:1234") but it has %d parts instead of 2`, len(v))
				}

				// parse the offset string into a uint
				ofs, err := strconv.ParseInt(v[1], 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing the given offset %s: %w", v[1], err)
				}

				action.ExistingCheckpoints = append(action.ExistingCheckpoints, mirror.ResumeFromLedger{Path: v[0], Offset: ofs})
			}

			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1], args[2:], n, bs, hwm)
			})
		},
	}

	// checkpoint flags
	cmd.Flags().StringVar(&action.Checkpoint, "checkpoint", "", "Save checkpoint file to file.  Can be provided to --resume-from and --resume-from-checkpoint to continue an incomplete serialize operation from where it left off.")
	cmd.Flags().StringSliceVar(&existingCheckpoints, "existing-from-checkpoint", []string{}, "List of checkpoint files and their offsets. e.g, checkpoint.txt:12345, checkpoint2.txt:23456")

	cmd.Flags().StringVar(&action.Compression, "compression", "", "Supports zstd and gzip compression methods. (Default behavior is no compression.)")
	flag.AddMemoryBufferFlags(cmd.Flags(), &mbufOpts)
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
