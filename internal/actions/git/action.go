// Package git facilitates executing actions for git repositories stored in an OCI registry.
package git

import (
	"git.act3-ace.com/ace/data/tool/internal/actions"
)

// Action represents a general git-oci action.
type Action struct {
	*actions.DataTool

	AltGitExec    string
	AltGitLFSExec string
	LFS           bool
	LFSServerURL  string

	CacheDir string
}
