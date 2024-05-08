/*
Copyright © 2021 ACT3 DevSecOps

*/

package cli

import (
	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/internal/actions"
)

// repoCmd represents the repo command.
func newUtilCmd(tool *actions.DataTool) *cobra.Command {
	var utilCmd = &cobra.Command{
		Use:   "util",
		Short: "Utility commands",
		Long:  `Includes a variety of less-used utility commands that do not fit elsewhere in the command tree.`,
	}
	utilCmd.AddCommand(
		newPruneCmd(tool),
		newFilterCmd(),
		newMuxCmd(),
		newDemuxCmd(),
		newMbufferCmd(),
		newSimulateCmd(tool),
		newGenKeyCmd(tool),
	)
	return utilCmd
}
