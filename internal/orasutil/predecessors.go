package orasutil

import "oras.land/oras-go/v2/content"

// PredecessorStorage is an oras Storage, with the ability to discover predecessors.
type PredecessorStorage interface {
	content.Storage
	content.PredecessorFinder
}
