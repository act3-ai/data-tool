// Package actions contains all the major functional actions that can be taken by the cli.
package actions

import (
	"gitlab.com/act3-ai/asce/data/tool/pkg/conf"
)

// DataTool represents the base data tool action.
type DataTool struct {
	version string
	Config  *conf.Configuration
}

// NewTool creates a new tool action.
func NewTool(version string) *DataTool {
	return &DataTool{
		version: version,
		Config:  conf.NewConfiguration("ace-dt/+" + version),
	}
}

// Version returns the version of ace data tool.
func (action *DataTool) Version() string {
	return action.version
}

// TelemetryOptions defines the options for telemetry server use.
type TelemetryOptions struct {
	URL string // overrides the current telemetry host url
}
