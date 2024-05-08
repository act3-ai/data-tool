package bottle

import (
	"github.com/spf13/cobra"

	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions/bottle"
)

// newBtlAnnotateCmd command allows addition or removal of annotations to bottle schema.
func newBtlAnnotateCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Annotate{Action: tool}

	cmd := &cobra.Command{
		GroupID: "metadata",
		Aliases: []string{"annotation"},
		Use:     "annotate <key>=<value>",
		Short:   "(advanced) Adds or removes an annotation as key-value pair to specified bottle",
		Long: `Annotations are typically used to encode arbitrary metadata into the bottle.

An annotation key may be up to 63 characters and must begin with a letter or number. The key may contain
letters, numbers, punctuation characters. The value can be any string of arbitrary length.
`,
		Example: `
Add annotation <foo=bar> to bottle in current working directory:
	ace-dt bottle annotate foo=bar

Remove annotation <foo> from bottle <bar> at path <my/bottle/path>:
	ace-dt bottle annotate foo- -d my/bottle/path

List all annotations for bottle in current working directory:
	ace-dt bottle annotate list
`,

		Args: cobra.MaximumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args, cmd.OutOrStdout())
		},
	}

	cmd.AddCommand(
		newBtlAnnotListCmd(tool),
	)
	return cmd
}

func newBtlAnnotListCmd(tool *actions.Action) *cobra.Command {
	action := &actions.AnnotateList{Action: tool}

	listAnnotCmd := &cobra.Command{
		Aliases: []string{"ls"},
		Use:     "list",
		Short:   "Lists annotation associated with a bottle",
		Example: ` 
List annotation on bottle at path <my/bottle/path>:
	ace-dt bottle annotate list foo=bar -d my/bottle/path

List annotation on bottle in current working directory:
	ace-dt bottle annotate list
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return listAnnotCmd
}
