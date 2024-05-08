// Package git facilitates executing actions for git repositories stored in an OCI registry.
package git

import (
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
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
