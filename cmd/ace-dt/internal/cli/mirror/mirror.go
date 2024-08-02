// Package mirror provides commands for mirroring OCI and Bottle objects.
/*
Copyright © 2020 ACT3 DevSecOps

*/
package mirror

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
	mirroractions "gitlab.com/act3-ai/asce/data/tool/internal/actions/mirror"
)

// NewMirrorCmd represents the base mirror command.
func NewMirrorCmd(tool *actions.DataTool) *cobra.Command {
	action := &mirroractions.Action{DataTool: tool}
	cmd := &cobra.Command{
		GroupID: "core",
		Use:     "mirror",
		Short:   "OCI mirroring operations",
	}
	cmd.AddCommand(
		newGatherCmd(action),
		newSerializeCmd(action),
		newDeserializeCmd(action),
		newScatterCmd(action),
		newCloneCmd(action),
		newArchiveCmd(action),
		newUnarchiveCmd(action),
		newBatchSerializeCmd(action),
	)

	cmd.PersistentFlags().BoolVarP(&action.Recursive, "recursive", "r", false, "recursively copy the referrers")

	return cmd
}
