// Package mirror defines commands for performing mirror operations for OCI objects and bottles.
package mirror

import (
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
)

// Action represents a general mirror action.
type Action struct {
	*actions.DataTool

	Insecure  bool // allow insecure registry access
	Recursive bool // also copy referrer recursively
}
