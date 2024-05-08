package bottle

import (
	"github.com/spf13/cobra"

	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// newBtlPartCmd is the top level command that aggregates subcommands for interacting with bottles parts fields.
func newBtlAuthorCmd(tool *actions.Action) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "metadata",
		Use:     "author",
		Aliases: []string{"authors"},
		Short:   "Bottle author operations",
		Long:    `This command group provides subcommands for interacting with author information of bottle parts.`,
	}

	cmd.AddCommand(
		listAuthorCmd(tool),
		addAuthorCmd(tool),
		removeAuthorCmd(tool),
	)
	return cmd
}

func listAuthorCmd(tool *actions.Action) *cobra.Command {
	action := &actions.AuthorList{Action: tool}

	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists author information of a bottle",
		Example: `
List authors of bottle in current working directory:
	ace-dt bottle author list
  
List authors of bottle at path <my/bottle/path>:
	ace-dt bottle author list -d my/bottle/path
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return listCmd
}

func addAuthorCmd(tool *actions.Action) *cobra.Command {
	action := &actions.AuthorAdd{Action: tool}

	var authorURL string

	var addCmd = &cobra.Command{
		Use:   "add [NAME] [EMAIL]",
		Short: "Adds author information to specified bottle",
		Example: `
Add author <John Doe> to bottle in current working directory:
	ace-dt bottle author add "John Doe" "jdoe@example.com"

Add author <Alice Wonders> to bottle at path <my/bottle/path>:
	ace-dt bottle author add "Alice Wonders" "alicew@example.com" --url="university.example.com/~awonders" -d my/bottle/path
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], args[1], authorURL, cmd.OutOrStdout())
		},
	}

	addCmd.Flags().StringVar(&authorURL, "url", "", "specify author's webpage, or preferred public url")

	return addCmd
}

func removeAuthorCmd(tool *actions.Action) *cobra.Command {
	action := &actions.AuthorRemove{Action: tool}

	var removeCmd = &cobra.Command{
		Use:     "remove [NAME]",
		Aliases: []string{"rm"},
		Short:   "Removes author's information from a bottle",
		Example: `
Remove author <John Doe> from bottle in current working directory:
	ace-dt bottle author remove "John Doe" 
  
Remove author <Alice Wonders> from bottle at path <my/bottle/path>:
	ace-dt bottle author rm "Alice Wonders" -d my/bottle/path
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], cmd.OutOrStdout())
		},
	}

	return removeCmd
}
