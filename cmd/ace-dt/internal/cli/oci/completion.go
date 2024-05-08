package oci

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/internal/actions"
)

// RefCompletion adds a ValidArgsFunction to a command. Optionally add unix-style argument indices to specify
// which arguments OCI ref completion should apply to, falls back to default shell behavior for unspecified indices.
// Applies OCI ref completion only to the first argument if no argument indices are provided.
func RefCompletion(tool *actions.DataTool, argIndices ...int) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := cmd.Context()

		s := fmt.Sprintf("args: %s toComplete: %s\n", args, toComplete)
		cobra.CompDebug(s, false)

		validPos := -1 // is valid and index of args
		switch len(argIndices) {
		case 0:
			validPos = 0
		default:
			for _, argIndex := range argIndices {
				i := argIndex - 1          // adjust unix-style arg indicies to cobra, which excludes $0
				currentArgPos := len(args) // args contains completed args, so now our index is len(args)-1+1
				if currentArgPos == i {
					validPos = i
					break
				}
			}
		}
		if validPos < 0 {
			s := fmt.Sprintf("attempting oci ref completion in an invalid arg position: current position = %d; valid positions = %v\n", len(args), argIndices)
			cobra.CompDebug(s, false)
			return nil, cobra.ShellCompDirectiveDefault
		}

		repo, err := tool.Config.ConfigureRepository(ctx, strings.Join(args[validPos:], "")+toComplete)
		if err != nil {
			// if we can't connect to the repo don't supply completion,
			// we'll let the calling command handle the connection error
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// support completion for partially provided tag
		ref := repo.Reference.String()
		idx := strings.LastIndex(ref, ":")
		if idx > -1 {
			ref = ref[:idx]
		}

		var tagList []string
		err = repo.Tags(ctx, "", func(tags []string) error {
			tl := make([]string, 0, len(tags))
			for _, tag := range tags {
				tl = append(tl, ref+":"+tag)
			}
			tagList = tl
			return nil
		})
		cobra.CheckErr(err)
		return tagList, cobra.ShellCompDirectiveNoFileComp
	}
}
