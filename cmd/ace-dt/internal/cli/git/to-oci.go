package git

import (
	"context"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions/git"
)

// newToOCICmd creates a new cobra.Command for the to-oci subcommand.
func newToOCICmd(base *git.Action) *cobra.Command {
	action := &git.ToOCI{Action: base}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "to-oci GIT_REMOTE OCI_REFERENCE GIT_REFERENCES...",
		Short: "Store a git repository as an OCI artifact.",
		Long: `Sync an OCI artifact with all refs in the git reference list, or all existing refs if none are provided. Typically the git reference list contains tag and/or head references, but commits are acceptable as well.
		
For optimal efficiency, it is recommended to always use the same OCI reference for subsequent syncs. However, it is important to consider that a to-oci sync strictly updates the refs provided during its execution (or all if no refs are provided). The resulting OCI artifact may contain additional references only if a prior sync has occurred for the same artifact (OCI reference). To view the state of references in the OCI artifact please see ace-dt git list-refs. All git references included in the OCI artifact will be accessible in the git repository updated with ace-dt git from-oci, see ace-dt git from-oci for more details.

Syncing git-lfs tracked files is supported by default. The --lfs flag is available to change this behavior, however it is not recommended as missing LFS files often results in a subsequent call to ace-dt git from-oci to fail; unless the destination repository explicitly enables this. `,

		Example: `Create a new OCI artifact with all tag and head references:
		$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync
		
Create a new OCI artifact with the v1.6.1 tag reference:
    $ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync v1.6.1

Overwrite an existing OCI artifact with a new base layer with the 
  main head reference (does not include the v1.6.1 tag ref):
    $ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync main --clean

Exclude git LFS tracked files:
    $ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync main --lfs=false

Cache the git repository at GIT_REMOTE (caches lfs files as well):
    $ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync main --cache-path CACHE_PATH
		`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				action.GitRemote = args[0]
				action.Repo = args[1]
				action.RevList = args[2:]
				return action.Run(ctx)
			})
		},
	}

	cmd.Flags().BoolVar(&action.Clean, "clean", false, "Start a clean OCI artifact regardless whether or not a tag exists in the target repository. This will overwrite the existing tag.")

	return cmd
}
