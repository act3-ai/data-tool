// Package actions contains all the major functional actions that can be taken by the cli.
package actions

import (
	"fmt"

	"oras.land/oras-go/v2/registry/remote/credentials"

	"github.com/act3-ai/data-tool/pkg/conf"
)

// DataTool represents the base data tool action.
type DataTool struct {
	version string
	Config  *conf.Configuration
}

// NewTool creates a new tool action.
func NewTool(version string) *DataTool {
	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		panic(fmt.Sprintf("failed to get credential store: %s", err))
	}

	return &DataTool{
		version: version,
		Config:  conf.New(conf.WithCredentialStore(credStore)),
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
