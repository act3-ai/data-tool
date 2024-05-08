package mirror

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/flag"
	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/mirror"
	"git.act3-ace.com/ace/data/tool/internal/mirror"
)

// newArchiveCmd represents the mirror archive command.
func newArchiveCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Archive{Action: tool}
	uiOptions := ui.Options{}
	mbufOpts := flag.MemoryBufferOptions{
		BlockSize: 1024 * 1024, // 1MiB is the default block size
	}
	var existingCheckpoints []string

	cmd := &cobra.Command{
		Use:   "archive SOURCES-FILE DEST-FILE REFERENCE",
		Short: "Efficiently copies images listed in SOURCES-FILE to the DEST-FILE in TAR format",
		Long: `Efficiently copies images listed in SOURCES-FILE to the DEST-FILE in TAR format. If it is a tape archive better performance can be had by setting --buffer-size=1Gi or larger.  The tar file can also be written to the tape after serialization is completed (see "ace-dt util mbuffer").
		Because this is a combination of mirror gather and mirror serialize, it inherits all of the flags and options defined in those commands.
		EXISTING-IMAGE(s) are images that we use to extract blob references from to determine if we need to serialize the blob.
		
SOURCES-FILE is a text file with one OCI image reference per line.  Lines that begin with # are ignored. 
Labels can be added to each source in the SOURCES-FILE by separating with a comma and following a key=value format. These will be added as annotations to that manifest:
reg.example.com/library/source1,component=core,module=test

DEST-FILE is the name of the TAR file to be created on the local system.

REFERENCE is a sync tag to assign to archive. E.g., "sync-1"
`,
		Example: `
To archive all the images in a file named sources.list to a local file.tar, tagging it as sync-3, you can use
ace-dt mirror archive sources.list file.tar sync-3

To specify a directory ./test to cache the oci images locally before archiving them, you can use
ace-dt mirror archive sources.list file.tar sync-3 --cache ./test
`,

		Args: cobra.ExactArgs(3),
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
				return action.Run(ctx, args[0], args[1], args[2], args[3:], n, bs, hwm)
			})
		},
	}
	// checkpoint flags
	cmd.Flags().StringVar(&action.Checkpoint, "checkpoint", "", "Save checkpoint file to file.  Can be provided to --resume-from and --resume-from-checkpoint to continue an incomplete serialize operation from where it left off.")
	cmd.Flags().StringSliceVar(&existingCheckpoints, "existing-from-checkpoint", []string{}, "List of checkpoint files and their offsets. e.g, checkpoint.txt:12345, checkpoint2.txt:23456")
	cmd.Flags().BoolVar(&action.IndexFallback, "index-fallback", false, "Tells ace-dt to add indexes in annotations for registries that do not support nested indexes (i.e., not OCI 1.1 compliant).  This makes the references to the sub-indexes not real references therefore a garbage collection process might incorrectly delete the sub-indexes.  Therefore, this should only be used when necessary (e.g., when targeting Artifactory).")
	cmd.Flags().StringToStringVarP(&action.ExtraAnnotations, "annotations", "a", map[string]string{}, "Define any additional annotations to add to the index of the gather repository.")
	cmd.Flags().StringSliceVarP(&action.Platforms, "platforms", "p", []string{}, "Only gather images that match the specified platform(s). Warning: This will modify the manifest digest/reference.")
	flag.AddMemoryBufferFlags(cmd.Flags(), &mbufOpts)
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
