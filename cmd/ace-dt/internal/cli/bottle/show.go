/*
Copyright © 2020 ACT3 DevSecOps

*/

package bottle

import (
	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/oci"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// showCmd represents the show command.
func newShowCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Show{Action: tool}

	cmd := &cobra.Command{
		GroupID: "remote",
		Use:     "show [BOTTLE_REFERENCE]",
		Short:   "Display information about a remote or local data bottle",
		Long: `This command connects to a registry and queries information about a specified bottle and tag when specified. It can also show bottle information about a local bottle if the path is specified.
	
The information provided is a list of files contained within the bottle, their digests (sha256), sizes, and a list of labels associated with each one.
	
This list can be filtered with selectors, similar to the pull command. Only files matching the provided selector will be returned, revealing the expected result of pulling a data bottle with the selector set.
 
A bottle reference uses one of the forms
  by tag                <registry>/<repository>/<name>:<tag>
  by name (latest tag)  <registry>/<repository>/<name>
  by digest             <registry>/<repository>/<name>@sha256:<sha>
`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: oci.RefCompletion(action.DataTool),
		RunE: func(cmd *cobra.Command, args []string) error {
			// check if input is a reference, else fall back to bottle directory
			// it's possible to run show without any arguments, or a user can specify current directory with a dot
			// out of habit, so we address both cases here
			if len(args) > 0 && args[0] != "." {
				action.Ref = args[0]
			}

			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	PartSelectorFlags(cmd.Flags(), &action.PartSelector)

	return cmd
}
