// Package oci defines commands for general purpose oci manipulation back by oras.
package oci

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/data-tool/internal/actions/oci"
)

// NewOciCmd represents the base oci command.
func NewOciCmd(tool *actions.DataTool) *cobra.Command {
	action := &oci.Action{DataTool: tool}
	cmd := &cobra.Command{
		GroupID: "core",
		Use:     "oci",
		Short:   "Raw OCI operations",
	}

	cmd.AddCommand(
		newPushDirCmd(action),
		newTreeCmd(action),
		newIdxOfIdxCmd(action),
	)
	return cmd
}
