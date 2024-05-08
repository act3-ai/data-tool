package bottle

import (
	"strings"

	"github.com/spf13/cobra"

	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// newBtlPartCmd is the top level command that aggregates subcommands for interacting with bottles parts fields.
func newBtlPartCmd(tool *actions.Action) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "metadata",
		Use:     "part",
		Aliases: []string{"parts"},
		Short:   "Bottle part operations",
		Long: `This command group provides subcommands for interacting with metadata of bottle parts. You can enumerate the parts in this bottle using the list command, and modify them using the label command. 

To add or remove parts, see the the 'bottle commit' command
`,
	}

	cmd.AddCommand(
		newPartLabelCmd(tool),
		newPartListCmd(tool),
	)
	return cmd
}

func newPartListCmd(tool *actions.Action) *cobra.Command {
	action := &actions.PartList{Action: tool}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "show information about parts that are in this bottle",
		Long: `This commands shows information about parts that are in this bottle.
By default, the parts information shown are name, size, and labels.
User has the options of showing digest of a part, in lieu of labels,
by passing in the flag --digest, -D`,
		Example: `
List parts that are in bottle at current working directory:
	ace-dt bottle part list

List parts that are in the bottle at path <my/bottle/path>:
	ace-dt bottle part ls -d my/bottle/path
 
List parts that are in the bottle with digest information:
	ace-dt bottle part list -D 
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVarP(&action.WithDigest, "digest", "D", false, "show parts information with digest")

	return cmd
}

func newPartLabelCmd(tool *actions.Action) *cobra.Command {
	action := &actions.PartLabel{Action: tool}

	cmd := &cobra.Command{
		Use:   "label <key>=<value>... PATH...",
		Short: "add key-value pair as a label to specified bottle part",
		Long: `Add key-value pair as a label to specified bottle part

A label key and value must begin with a letter or number, and may contain 
letters, numbers, hyphens, dots, and underscores, up to  63 characters each.`,
		Example: `
Add label <foo=bar> to part <myPart.txt> bottle in current working directory:
	ace-dt bottle part label foo=bar myPart.txt

Add label <foo=bar> to many parts in current working directory:
	ace-dt bottle part label foo=bar myPart.txt myPicture.jpg myModel.model

Remove label "foo" from part "myPart.txt" in bottle at path "my/bottle/path":
	ace-dt bottle part label foo- my/bottle/path/myPart.txt -d my/bottle/path
`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nLabels := findNumLabels(args)
			// extract part path from expanded command line args
			labelList, pathList := args[:nLabels], args[nLabels:]

			return action.Run(cmd.Context(), pathList, labelList, cmd.OutOrStdout())
		},
	}

	return cmd
}

// findNumLabels finds the number of label arguments.  The rest are paths.
func findNumLabels(args []string) int {
	for i, item := range args {

		if strings.Contains(item, "=") || strings.HasSuffix(item, "-") {
			// found a label
			continue
		}

		// found the first part (bail out)
		return i
	}

	return -1
}
