// Package cli exports the ACE Data Tool command.
package cli

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli"
)

// NewToolCmd creates the ACE Data Tool command.
func NewToolCmd(version string) *cobra.Command {
	return cli.NewToolCmd(version)
}
