/*
Copyright © 2020 ACT3 DevSecOps
*/

package cli

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
)

// pruneCmd represents the prune command.
func newPruneCmd(tool *actions.DataTool) *cobra.Command {
	action := &actions.Prune{DataTool: tool}

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune bottle cache, removing least recently used files",
		Long: `Prune bottle cache, removing least recently used files.  Use
  maxsize option to choose a maximum size of data to keep, in MiB.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().Int64VarP(&action.Max, "maxsize", "s", -1, "A maximum size to keep in the cache, in MiB")

	return cmd
}
