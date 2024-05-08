package bottle

import (
	"github.com/spf13/cobra"

	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// newBtlPartCmd is the top level command that aggregates subcommands for interacting with bottles parts fields.
func newBtlSourceCmd(tool *actions.Action) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "metadata",
		Use:     "source",
		Aliases: []string{"sources", "src"},
		Short:   "Bottle source operations",
	}

	cmd.AddCommand(
		addSourceCmd(tool),
		removeSourceCmd(tool),
		listSourceCmd(tool),
	)
	return cmd
}

func addSourceCmd(tool *actions.Action) *cobra.Command {
	action := &actions.SourceAdd{Action: tool}

	cmd := &cobra.Command{
		Use:   "add [NAME] [URI]",
		Short: "add source information to bottle",
		Long: `Adds an source information specified bottle.
Source information is comprised of two fields: 'name' and 'uri'
When adding a source, both fields are required. The URI field accepts any valid URI (which include URLs) as well as the bottle ID format,which is 'bottle:sha256:badfacade...'`,
		Example: `
Add source 'mnist catalog' with uri https://mnist-catalog.example.com to bottle in current directory:
	ace-dt bottle source add "mnist catalog" "https://mnist-catalog.example.com"

Add source 'kaggle data' with uri https://kaggle.example.com/id256 to bottle at path my/bottle/path:
	ace-dt bottle source add "kaggle data" "https://kaggle.example.com/id256" -d my/bottle/path

Add source 'validation subset' with bottle ID sha256:989697e1f07d9454dad83b2171491ef55de9aa9ed9bf1e91814c464999517b77:
	ace-dt bottle source add 'validation subset' 'bottle://sha256:989697e1f07d9454dad83b2171491ef55de9aa9ed9bf1e91814c464999517b77'

Add source "training data" with bottle ID from another local bottle at "my/training/data"
	ace-dt bottle source "training data" "my/training/data" --path

Add source "training data" with bottle reference from a remote bottle at "reg.git.com/my/training/data"
	ace-dt bottle source "training data" "reg.git.com/my/training/data" --ref
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], args[1], cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&action.BottleForURI, "path", false, "Interpret the URI as a path to a bottle directory that will be used to get the bottle ID")
	cmd.Flags().BoolVar(&action.ReferenceForURI, "ref", false, "Interpret the URI as a bottle reference that will be used to get the bottle ID")
	cmd.Flags().Lookup("path").NoOptDefVal = "true"

	return cmd
}

func listSourceCmd(tool *actions.Action) *cobra.Command {
	action := &actions.SourceList{Action: tool}

	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list source information from a bottle",
		Example: `
List source from bottle in current directory:
	ace-dt bottle source list
  
List source from bottle at path my/bottle/path:
	ace-dt bottle source list -d my/bottle/path
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return listCmd
}

func removeSourceCmd(tool *actions.Action) *cobra.Command {
	action := &actions.SourceRemove{Action: tool}
	var removeCmd = &cobra.Command{
		Use:     "remove [NAME]",
		Aliases: []string{"rm"},
		Short:   "remove source information from a bottle",
		Long:    `Removes a source from a bottle. The source to be removed is identified by using the name field as a key.`,
		Example: `
Remove source <mnist catalog> from bottle in current working directory:
	ace-dt bottle source remove "mnist catalog" 

Remove source <kaggle data> from bottle at path <my/bottle/path>:
	ace-dt bottle source rm "kaggle data" -d my/bottle/path
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], cmd.OutOrStdout())
		},
	}

	return removeCmd
}
