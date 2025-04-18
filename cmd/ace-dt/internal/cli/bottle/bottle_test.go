package bottle

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
)

func rootTestCmd() *cobra.Command {
	action := actions.NewTool("0.0.0")

	cmd := NewBottleCmd(action)

	// TODO remove this and pass the config file directly to the action for testing
	cmd.PersistentFlags().StringArrayVar(&action.Config.ConfigFiles, "config", nil,
		`configuration file location for testing`)

	return cmd
}
