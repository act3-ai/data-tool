package mirror

import (
	"github.com/spf13/cobra"

	mirroractions "github.com/act3-ai/data-tool/internal/actions/mirror"
)

func newViewCmd(action *mirroractions.Action) *cobra.Command {
	view := &mirroractions.View{Action: action}

	cmd := &cobra.Command{
		Use:   "view <input>",
		Short: "List references in an OCI layout, OCI-layout archive, or registry",
		Long: `List the root references in <input>, one per line with no header
(the sources.list format).

<input> is one of:
  - an OCI-layout directory (oci-layout, index.json, blobs/)
  - an OCI-layout tarball (plain, gzip, or zstd)
  - a tagged or digest-qualified registry reference

Output:
  default      ref
  --platforms  ref,os/arch[/variant];...   (CSV)
  --stats      per-reference cost, unique, shared, and missing data, sorted by
               cost, then totals and space saved by deduplication.
               Add --platforms to include a platforms column.`,
		Example: `  ace-dt mirror view ./sync
  ace-dt mirror view ./images.tar.zst --platforms
  ace-dt mirror view reg.example.com/library/registry:2.8.3 --platforms
  ace-dt mirror view ./sync --stats`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags and args are valid at this point; a failure should not print usage.
			cmd.SilenceUsage = true
			return view.Run(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0])
		},
	}
	cmd.Flags().BoolVar(&view.Platforms, "platforms", false, "Include image platforms as a semicolon-separated CSV field")
	cmd.Flags().BoolVar(&view.Stats, "stats", false, "Show per-reference size, sharing, and missing-blob stats")

	return cmd
}
