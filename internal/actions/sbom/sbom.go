// Package sbom defines the sbom actions.
package sbom

import (
	"github.com/act3-ai/data-tool/internal/actions"
)

// Action represents a general security action.
type Action struct {
	*actions.DataTool
}
