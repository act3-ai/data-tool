package mirror

import (
	"github.com/spf13/cobra"

	mirroractions "github.com/act3-ai/data-tool/internal/actions/mirror"
)

func newViewCmd(action *mirroractions.Action) *cobra.Command {
	var detailed bool
	var showRefs bool
	var showImages bool
	var diffPath string

	cmd := &cobra.Command{
		Use:   "view <path>",
		Short: "View or compare images and manifests in extracted OCI layouts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return mirroractions.ViewGatherIndex(cmd.Context(), args[0], detailed, showRefs, showImages, diffPath)
		},
	}

	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Display detailed manifest information")
	cmd.Flags().BoolVar(&showRefs, "refs", false, "List top-level digest, reference name, and type (tab-separated)")
	cmd.Flags().BoolVar(&showImages, "images", false, "List index digest/type/ref and image digest/type/platform (tab-separated)")
	cmd.Flags().StringVar(&diffPath, "diff", "", "Compare top-level entries with this extracted OCI layout or index.json")

	return cmd
}
