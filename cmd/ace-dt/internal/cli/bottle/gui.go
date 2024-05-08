package bottle

import (
	"github.com/spf13/cobra"

	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// editCmd represents the edit command.
func newGuiCmd(tool *actions.Action) *cobra.Command {
	action := &actions.GUI{Action: tool}

	guiCmd := &cobra.Command{
		Use:   "gui",
		Short: "Open browser to a local web GUI for editing a bottle",
		Long: `Open your default web browser to a page to edit the given bottle.
This command will run a local webserver to show the GUI in your browser.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	guiCmd.Flags().StringVar(&action.Listen, "listen", "localhost:0", "Address and port for the server to listen for new connections")
	guiCmd.Flags().BoolVar(&action.DisableBrowser, "no-browser", false, "Automatically open browser to the GUI")

	return guiCmd
}
