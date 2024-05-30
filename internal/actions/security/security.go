// Package security defines the security actions.
package security

import (
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
)

// Action represents a general security action.
type Action struct {
	*actions.DataTool

	Insecure  bool // allow insecure registry access
	Recursive bool // also copy referrer recursively
}
