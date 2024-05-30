// Package bottle provides functions for managing transfer of bottle objects to and from an OCI registry, including
// configuring a pulled bottle, and establishing local metadata and file structure.
package bottle

import (
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
)

const defaultConcurrency = 5

// TransferOptions configures bottle transfers between a remote registry and localhost.
// Used to compose PullOptions and PushOptions.
type TransferOptions struct {
	// Optional, with defaults
	Concurrency int // default: 5

	// Optional, used for tracking virtual parts when pulling bottles with selectors
	// as well as caching blobs
	CachePath string
}

// concurrency returns the concurrency value, or the default if it's not set or invalid.
func (p *TransferOptions) concurrency() int {
	if p.Concurrency < 1 {
		p.Concurrency = defaultConcurrency
	}
	return p.Concurrency
}

// PartSelectorOptions are options for creating a PartSelector.
// Bottle parts may be selected by label, part name, or included
// public artifact.
// It is highly recommended to use the CachePath option in TransferOptions
// alongside these options to enable pushing of a partially pulled bottle,
// otherwise it will fail.
type PartSelectorOptions = bottle.PartSelectorOptions

// PullOptions provides options for pulling bottles from remote registry to the localhost.
type PullOptions struct {
	TransferOptions
	PartSelectorOptions
}
