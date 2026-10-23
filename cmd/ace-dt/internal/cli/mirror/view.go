package mirror

import (
	"github.com/spf13/cobra"

	mirroractions "github.com/act3-ai/data-tool/internal/actions/mirror"
	"github.com/act3-ai/data-tool/internal/mirror"
)

func newViewCmd(action *mirroractions.Action) *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "view <path-to-index.json>",
		Short: "View images and manifests contained in a gather index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			indexPath := args[0]
			return mirror.ViewGatherIndex(indexPath, verbose)
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Display detailed manifest information")

	return cmd
}