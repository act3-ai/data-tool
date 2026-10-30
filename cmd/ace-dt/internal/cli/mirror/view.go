package mirror

import (
	"github.com/spf13/cobra"

	mirroractions "github.com/act3-ai/data-tool/internal/actions/mirror"
)

func newViewCmd(action *mirroractions.Action) *cobra.Command {
	view := &mirroractions.View{Action: action}

	cmd := &cobra.Command{
		Use:   "view <layout>",
		Short: "View or compare images and manifests in extracted OCI layouts",
		Long: `View the contents of an extracted OCI image layout (a gather or serialize
output after untarring). <layout> is the layout directory or its index.json.
Archives are not read directly; extract them first.

Output formats (tab-separated, one record per line):
  --refs     digest  ref  index|image|artifact
  --images   digest  index     ref
             digest  image     platform      ref
             digest  artifact  artifactType  ref
             Index children follow their index and have ref <unnamed>.
  --diff     -|+  digest  ref  kind
             "-" = only in the --diff layout, "+" = only in <layout>.
             Exits non-zero when the indexes differ.

--refs and --diff read index.json only and accept a bare gather index.
Other modes need the full layout (oci-layout, index.json, blobs/).`,
		Example: `  ace-dt mirror view ./sync --refs
  ace-dt mirror view ./sync --images
  ace-dt mirror view ./sync --tree
  ace-dt mirror view ./extracted --diff ./sync/index.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags and args are valid at this point; a failure (including
			// "indexes differ") should not print usage.
			cmd.SilenceUsage = true
			return view.Run(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0])
		},
	}

	cmd.Flags().BoolVarP(&view.Detailed, "detailed", "d", false, "Display detailed information for each reference")
	cmd.Flags().BoolVar(&view.Refs, "refs", false, "List top-level digest, reference name, and kind")
	cmd.Flags().BoolVar(&view.Images, "images", false, "List indexes, images with platform, and artifacts with type")
	cmd.Flags().BoolVar(&view.Tree, "tree", false, "Display an OCI tree with sizes and referrers; with --images show manifests only, with --refs show roots only")
	cmd.Flags().StringVar(&view.Diff, "diff", "", "Compare top-level entries with this extracted OCI layout or index.json")

	cmd.MarkFlagsMutuallyExclusive("refs", "images", "detailed")
	cmd.MarkFlagsMutuallyExclusive("tree", "detailed")
	for _, f := range []string{"refs", "images", "detailed", "tree"} {
		cmd.MarkFlagsMutuallyExclusive("diff", f)
	}

	return cmd
}