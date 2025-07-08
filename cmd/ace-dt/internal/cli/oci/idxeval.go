package oci

import (
	"context"

	"github.com/act3-ai/data-tool/cmd/ace-dt/internal/cli/internal/ui"
	"github.com/act3-ai/data-tool/internal/actions/oci"
	"github.com/spf13/cobra"
)

// newIdxOfIdxCmd creates a new cobra.Command for the pushdir subcommand.
func newIdxOfIdxCmd(base *oci.Action) *cobra.Command {
	action := &oci.IdxOfIdx{Action: base}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "idx-of-idx FILE",
		Short: "Discover if any images in FILE have children OCI indexes, i.e. index-of-index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				action.File = args[0]
				return action.Run(ctx, cmd.OutOrStdout())
			})
		},
	}

	return cmd
}
