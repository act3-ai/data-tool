// Package git defines general git operations backed by oras.
package git

import (
	"os"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/internal/actions"
	"git.act3-ace.com/ace/data/tool/internal/actions/git"
)

// NewGitCmd represents the base git command.
func NewGitCmd(tool *actions.DataTool) *cobra.Command {
	action := &git.Action{DataTool: tool}
	var cmd = &cobra.Command{
		GroupID: "core",
		Use:     "git",
		Short:   "Git to/from OCI",
		Long:    `This command group provides the ability to store git repositories within OCI registries, as well as updating git repositories from an OCI registry. By default this command group supports Git LFS.`,
		Example: `A typical workflow begins with a git to OCI sync:
$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync
	
You may view all references included in the sync:
$ ace-dt git list-refs reg.example.com/my/libgit2:sync
	
Finally, sync with a secondary repository:
$ ace-dt git from-oci reg.example.com/my/libgit2:sync https://github.com/my_copies/libgit2_copy`,
	}

	cmd.PersistentFlags().BoolVar(&action.LFS, "lfs", true, "Include git LFS files, if they exist, default is true.")
	cmd.PersistentFlags().StringVar(&action.LFSServerURL, "lfs-server", "", "Directly specify an LFS server URL.")
	cmd.PersistentFlags().StringVar(&action.CacheDir, "cache-path", os.Getenv("ACE_DT_GIT_CACHE_PATH"), "Use the specified directory as a git repository cache (settable with env 'ACE_DT_GIT_CACHE_PATH').")
	cmd.PersistentFlags().StringVar(&action.AltGitExec, "git-executable", os.Getenv("ACE_DT_GIT_EXECUTABLE"), "Provide a path to an alternative git executable (settable with env 'ACE_DT_GIT_EXECUTABLE').")
	cmd.PersistentFlags().StringVar(&action.AltGitLFSExec, "git-lfs-executable", os.Getenv("ACE_DT_GIT_LFS_EXECUTABLE"), "Provide a path to an alternative git-lfs executable (settable with env 'ACE_DT_GIT_LFS_EXECUTABLE').")

	cmd.AddCommand(
		newToOCICmd(action),
		newFromOCICmd(action),
		newListRefsCmd(action),
	)

	return cmd
}
