package bottle

import (
	"github.com/spf13/cobra"

	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// newLabelCmd represents the label command.
func newLabelCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Label{Action: tool}

	cmd := &cobra.Command{
		GroupID: "metadata",
		Use:     "label <key>=<value>",
		Short:   "add key-value pair as a label to specified bottle",
		Long: `Add key-value pair label to the bottle

A label key and value must begin with a letter or number, and may contain 
  letters, numbers, hyphens, dots, and underscores, up to  63 characters each.

Do not confuse bottle labels with part labels.  See "ace-dt bottle part label -h" for more information about how to add labels to parts.
`,
		Example: `
Add label <foo=bar> to bottle at path <my/bottle/path>:
	ace-dt bottle label foo=bar -d my/bottle/path

Add multiple labels to a bottle in current working directory:
	ace-dt bottle label foo1=bar1 foo2=bar2 foo3=bar3

Remove label <foo> from bottle <bar> at path <my/bottle/path>:
	ace-dt bottle label foo- -d my/bottle/path
`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args, cmd.OutOrStdout())
		},
	}

	cmd.AddCommand(
		listBtlLabelCmd(tool),
	)
	return cmd
}

// listBtlLabelCmd represents the label command.
func listBtlLabelCmd(tool *actions.Action) *cobra.Command {
	action := &actions.LabelList{Action: tool}

	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list labels applied on specified bottle",
		Example: `
List label on bottle at path <my/bottle/path>:
	ace-dt bottle label list -d my/bottle/path

List the label of the bottle of the current directory:
	ace-dt bottle label list	
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return listCmd
}
