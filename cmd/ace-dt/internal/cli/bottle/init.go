package bottle

import (
	"github.com/spf13/cobra"

	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions/bottle"
)

// newInitCmd returns a command that represents the init command.
func newInitCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Init{Action: tool}

	cmd := &cobra.Command{
		GroupID: "basic",
		Use:     "init",
		Short:   "Initialize metadata and tracking for a data bottle",
		Long: `With a specified root directory, initialize metadata relevant for the bottle.  
The current working directory is used by default, and the data bottle directory is created if it currently does not exist.  An ".dt/entry.yaml" file is created in the data bottle directory that can be edited to fill in relevant data.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVarP(&action.Force, "force", "f", false, "Perform initialization even if the data bottle appears to already be initialized")

	cmd.Example = `
To initialize the current working directory:
	ace-dt bottle init

Given a directory TESTSET:
	ace-dt bottle init --bottle-dir TESTSET

Next steps:
  - add metadata, ace-dt bottle [annotate, artifact, author, label, metric, part, source]
  - edit metadata directly, ace-dt bottle edit
  - add to or edit bottle files, then commit changes, ace-dt bottle commit
  - Push bottle, ace-dt bottle push
`

	return cmd
}
