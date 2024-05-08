// Package bottle provides functions for managing transfer of bottle objects to and from an OCI registry, including
// configuring a pulled bottle, and establishing local metadata and file structure.
package bottle

import (
	"context"

	"oras.land/oras-go/v2"

	telemv1alpha1 "git.act3-ace.com/ace/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"git.act3-ace.com/ace/data/tool/internal/bottle"
	"git.act3-ace.com/ace/data/tool/pkg/conf"
	reg "git.act3-ace.com/ace/data/tool/pkg/registry"
)

// TransferConfig configures bottle transfers between the local host
// and remote OCI registries.
type TransferConfig struct {
	// Required
	Reference string
	PullPath  string

	// Optional, with defaults
	NewGraphTargetFn reg.NewGraphTargetFn
	Concurrency      int
	telemHosts       []telemv1alpha1.Location
	telemUserName    string

	// Optional
	matchBottleID string
	bottleIDFile  string
	cachePath     string
	partSelector  bottle.PartSelectorOptions

	// Optional, internal
	preConfigHandler  configHandlerFn
	postConfigHandler configHandlerFn
}

// TransferOption applies options or default overrides to a TransferConfig.
type TransferOption func(*TransferConfig)

// NewTransferConfig initializes a TransferConfig with the required configuration and
// applies default overrides and options.
func NewTransferConfig(ctx context.Context, ref, pullPath string, config *conf.Configuration, opts ...TransferOption) *TransferConfig {
	cfg := config.Get(ctx)
	// define default configuration
	transferCfg := &TransferConfig{
		// set required
		Reference: ref,
		PullPath:  pullPath,
		// set defaults
		NewGraphTargetFn: func(ctx context.Context, ref string) (oras.GraphTarget, error) {
			return config.ConfigureRepository(ctx, ref)
		},
		Concurrency:   cfg.ConcurrentHTTP,
		telemHosts:    cfg.Telemetry,
		telemUserName: cfg.TelemetryUserName,

		cachePath: cfg.CachePath,
	}

	// apply options and override defaults
	for _, o := range opts {
		o(transferCfg)
	}

	return transferCfg
}

// WithTelemetry overrides a pull to discover bottle locations with telemetry hosts.
// SHOULD be used with a bottle reference scheme, e.g. bottle:<digest>,
// otherwise telemetry will not be used.
func WithTelemetry(telemHosts []telemv1alpha1.Location, telemUserName string) TransferOption {
	return func(tc *TransferConfig) {
		tc.telemHosts = telemHosts
		tc.telemUserName = telemUserName
	}
}

// WithConcurrency overrides the concurrency for http connections.
// Concurrency defaults to 10.
func WithConcurrency(c int) TransferOption {
	return func(tc *TransferConfig) {
		tc.Concurrency = c
	}
}

// WithNewGraphTargetFn overrides the default newTargetFn used to create oras.GraphTargets.
// NewTargetFn defaults to reg.DefaultNewTargetFn.
func WithNewGraphTargetFn(newTargetFn reg.NewGraphTargetFn) TransferOption {
	return func(tc *TransferConfig) {
		tc.NewGraphTargetFn = newTargetFn
	}
}

// WithPartSelection enables the selection of parts by lable, name, or public artifact type.
func WithPartSelection(labelSelectors, partNames, artifacts []string) TransferOption {
	// prevent part selection failure if a nil slice is passed
	if labelSelectors == nil {
		labelSelectors = []string{}
	}
	if partNames == nil {
		partNames = []string{}
	}
	if artifacts == nil {
		artifacts = []string{}
	}
	return func(tc *TransferConfig) {
		tc.partSelector.Selectors = labelSelectors
		tc.partSelector.Parts = partNames
		tc.partSelector.Artifacts = artifacts
	}
}

// WithNoParts enables pulling only bottle metadata.
func WithNoParts() TransferOption {
	return func(tc *TransferConfig) {
		tc.partSelector.Empty = true
	}
}

// WithMatchBottleID enables matching the bottleID on pull.
// Only applicable if a bottleID scheme is used, e.g. bottle://.
func WithMatchBottleID(bottleID string) TransferOption {
	return func(tc *TransferConfig) {
		tc.matchBottleID = bottleID
	}
}

// WithBottleIDFile enables the bottleID to be saved to the specified path.
func WithBottleIDFile(path string) TransferOption {
	return func(tc *TransferConfig) {
		tc.bottleIDFile = path
	}
}

// WithCachePath enables bottle part caching by providing a path to a blob cache.
func WithCachePath(path string) TransferOption {
	return func(tc *TransferConfig) {
		tc.cachePath = path
	}
}

// configHandlerFn is a function callback that allows a caller to process a bottle
// configuration, performing initialization as desired.
type configHandlerFn func(btl *bottle.Bottle, rawConfig []byte) error

// defaultConfigHandler processes a bottle configuration by creating the config
// file and writing it to disk.
func defaultConfigHandler(btl *bottle.Bottle, rawConfig []byte) error {
	return bottle.CreateBottle(btl.GetPath(), true)
}
