package oci

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"git.act3-ace.com/ace/data/tool/internal/actions/oci"
)

// newPushDirCmd creates a new cobra.Command for the pushdir subcommand.
func newPushDirCmd(base *oci.Action) *cobra.Command {
	action := &oci.PushDir{Action: base}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "pushdir PATH IMAGE",
		Short: "Push local directory as an OCI image to a remote registry",
		Long:  `PATH is a directory that will be used as the contents of the image.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1])
			})
		},
	}

	cmd.Flags().BoolVar(&action.Legacy, "legacy", false, "use legacy (a.k.a., docker) media types instead of the newer OCI media types.  This is useful for backwards compatibility with old registries (e.g., gitlab registry).")
	cmd.Flags().Var(&action.Platform, "platform", "platform to use for uploading image.  The format is os/arch (e.g., linux/amd64)")
	cobra.CheckErr(cmd.MarkFlagRequired("platform"))

	cmd.Flags().BoolVar(&action.Reproducible, "reproducible", false, "Makes the uploaded artifact have a consistent (reproducible) digest.  This removes timestamps.  The benefit is that this produces layers that are possibly better suited for mirroring.")

	return cmd
}
