// Package sbom contains the ace-dt sbom commands
package sbom

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
	sbomactions "github.com/act3-ai/data-tool/internal/actions/sbom"
)

// NewSBOMCmd handles command operations within the sbom package.
func NewSBOMCmd(tool *actions.DataTool) *cobra.Command {
	action := &sbomactions.Action{DataTool: tool}
	cmd := &cobra.Command{
		GroupID: "core",
		Use:     "sbom",
		Short:   "ace-dt sbom operations",
	}
	cmd.AddCommand(NewSBOMListCommand(action))
	cmd.AddCommand(NewSBOMFetchCommand(action))
	return cmd
}
