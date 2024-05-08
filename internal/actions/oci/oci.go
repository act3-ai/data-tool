// Package oci defines general purpose oci push and pull operations backed by oras.
package oci

import (
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
)

// Action represents a general oci action.
type Action struct {
	*actions.DataTool
}
