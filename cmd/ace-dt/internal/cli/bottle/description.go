package bottle

import (
	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"

	"github.com/spf13/cobra"
)

func newBtlDescribeCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Describe{Action: tool}

	cmd := &cobra.Command{
		GroupID: "metadata",
		Aliases: []string{"description"},
		Use:     "describe [DESCRIPTION TEXT]",
		Short:   "Adds a description to specified bottle",
		Long:    `Add description and useful details to a bottle. Short description can be added directly from the command line, and long description can be added from a specified text file.`,
		Example: `
Add a short description to bottle at path <my/bottle/path>:
  ace-dt bottle describe "The context of this bottle is foobar." -d my/bottle/path

Add description text from a file to a bottle at path <my/bottle/path>:
  ace-dt bottle describe --from-file ./my-description.txt -d my/bottle/path
`,

		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			description := ""
			if len(args) > 0 {
				description = args[0]
			}
			return action.Run(cmd.Context(), description, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&action.File, "from-file", "f",
		"", "add description text from a file to a bottle")

	return cmd
}
