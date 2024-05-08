package git

import (
	"context"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions/git"
)

// newFromOCICmd creates a new cobra.Command for the rebuild subcommand.
func newFromOCICmd(base *git.Action) *cobra.Command {
	action := &git.FromOCI{Action: base}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "from-oci OCI_REFERENCE GIT_REMOTE",
		Short: "Pull updates from an OCI artifact into a git repository.",
		Long: `Sync a git repository with all git references contained in an OCI artifact. Outputs all reference updates and their new commits. To view the state of references in the OCI artifact please see ace-dt git list-refs.

Syncing prevents overwriting existing references by default to prevent undesired results. This behavior may be overidden with the --force flag, which may be interpreted as it is used in git push.

Syncing git-lfs tracked files is supported by default. The --lfs flag is available to change this behavior, however it is not recommended as missing LFS files often results in a subsequent call to ace-dt git from-oci to fail; unless the destination repository explicitly enables this.`,

		Example: `Push updates in a sync manifest to a target repository:
    $ ace-dt git from-oci reg.example.com/my/libgit2:sync https://github.com/mypersonal/libgit2
		
Force push updates in a sync manifest to a target repository:
    $ ace-dt git from-oci reg.example.com/my/libgit2:sync https://github.com/mypersonal/libgit2 --force

Cache the git repository at GIT_REMOTE (caches lfs files as well):
    $ace-dt git from-oci reg.example.com/my/libgit2:sync --cache-path CACHE_PATH`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: oci.RefCompletion(action.DataTool, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				action.Repo = args[0]
				action.GitRemote = args[1]
				return action.Run(ctx)
			})
		},
	}

	cmd.Flags().BoolVar(&action.Force, "force", false, "Force updates to repository references, analogous to git push --force.")

	return cmd
}
