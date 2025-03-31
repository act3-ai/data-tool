// Package sbom defines the sbom actions.
package sbom

import (
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
)

// Action represents a general security action.
type Action struct {
	*actions.DataTool
}
