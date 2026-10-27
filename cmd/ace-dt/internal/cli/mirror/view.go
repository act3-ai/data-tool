package mirror

import (
	"github.com/spf13/cobra"

	mirroractions "github.com/act3-ai/data-tool/internal/actions/mirror"
)

func newViewCmd(action *mirroractions.Action) *cobra.Command {
	var detailed bool

	cmd := &cobra.Command{
		Use:   "view <path>",
		Short: "View images and manifests contained in a gather index, OCI store, or tarball",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			
			// Call it using the existing mirroractions alias
			return mirroractions.ViewGatherIndex(cmd.Context(), path, detailed)
		},
	}

	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Display detailed manifest information")

	return cmd
}