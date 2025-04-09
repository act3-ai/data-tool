package security

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
	securityactions "github.com/act3-ai/data-tool/internal/actions/security"
)

// NewSecurityCmd handles command operations within the security package.
func NewSecurityCmd(tool *actions.DataTool) *cobra.Command {
	action := &securityactions.Action{DataTool: tool}
	cmd := &cobra.Command{
		GroupID: "core",
		Use:     "security",
		Short:   "ace-dt security operations",
	}
	cmd.AddCommand(newScanCommand(action))

	return cmd
}
