// Package push provides cross-repository blob mounting features for use during a push operation.  Cross repository
// blob mounting works across repositories on the same registry.  Note, cross-registry blob transfer is possible using a
// OCIFileCache backing store.
package push

import (
	"context"

	"github.com/opencontainers/go-digest"

	"git.act3-ace.com/ace/data/tool/internal/cache"
	"git.act3-ace.com/ace/data/tool/internal/ref"
)

// MountStatus is a status indicator type for a mountLayers attempt.
type MountStatus int

const (
	// MountStatusSkipped indicates a case where no mountable sources were found.
	MountStatusSkipped = iota

	// MountStatusMounted indicates a successful mount was performed.
	MountStatusMounted

	// MountStatusUnavailable indicates alternate sources are specified, but do not match current registry.
	MountStatusUnavailable

	// MountStatusPushStarted indicates that a layer could not be mounted so the server started a push operation for
	// file transfer.
	MountStatusPushStarted

	// MountStatusMountRequired indicates that the "must mount" flag was set for a layer, but mounting was not
	// successful.
	MountStatusMountRequired

	// MountStatusFailure indicates a failure during the mount operation prevented success.
	MountStatusFailure
)

// MountableLayers is an interface for providing information about known alternate repository sources for layerIDs.
// LayerIDs are intended to be digests with algorithm included, eg "sha256:b242feafa...".
type MountableLayers interface {
	// IsMountable returns true if the provided layerID (digest) exists in a mountable layers provider. The destMount
	// parameter determines if the source and dest registries match, and also filters repositories that match exactly
	IsMountable(ctx context.Context, layerID digest.Digest, destMount ref.Ref) bool

	// MustMount returns true if the provided layerID MUST be mounted (eg. because it's unavailable for file transfer)
	MustMount(ctx context.Context, layerID digest.Digest) bool

	// Sources returns a slice of repository references that exist on the destination registry, and don't exactly
	// match the destination repository (not including tag)
	Sources(ctx context.Context, layerID digest.Digest, destMount ref.Ref) []ref.Ref

	// AddSource associates a given ref with the provided layer id (digest)
	AddSource(ctx context.Context, layerID digest.Digest, source ref.Ref, mustMount bool)
}

// LayerRepoList contains a list of repositories to attempt to use as the source of a layer in a cross-repo mount
// operation. For each list a boolean value stores whether or not the layer is flagged as "must mount", which
// can prevent an attempt to perform a standard file transfer if the mount attempt fails.
type LayerRepoList struct {
	repos     []ref.Ref
	mustMount bool
}

// LayerRepos is a mapping of layer ids (digests) to a list of repos where the layer is suspected (or known) to exist.
type LayerRepos struct {
	layers map[digest.Digest]LayerRepoList
	bic    cache.BIC
}

// AddSource adds a layerID, ref mapping to the layer repos collection.
func (lr *LayerRepos) AddSource(ctx context.Context, layerID digest.Digest, source ref.Ref, mustMount bool) {
	if m, ok := lr.layers[layerID]; ok {
		for i := range m.repos {
			if m.repos[i].Match(source, ref.RefMatchRegRepo) {
				return
			}
		}
		m.mustMount = mustMount
		m.repos = append(m.repos, source)
		lr.layers[layerID] = m
	} else {
		m := LayerRepoList{mustMount: mustMount, repos: []ref.Ref{source}}
		lr.layers[layerID] = m
	}
}

// IsMountable returns true if the layerID is found on the destination registry in a different repo. False if not on
// the registry, or if the repository already contains the layerID.
func (lr *LayerRepos) IsMountable(ctx context.Context, layerID digest.Digest, destMount ref.Ref) bool {
	if sources, ok := lr.layers[layerID]; ok {
		for _, s := range sources.repos {
			if s.Match(destMount, ref.RefMatchRegRepo) {
				// already in repo
				return false
			}
			if s.Match(destMount, ref.RefMatchReg) {
				// found on registry in a different repo
				return true
			}
		}
	}
	if lr.bic == cache.BIC(nil) {
		return false
	}
	if len(cache.LocateLayerDigest(ctx, lr.bic, layerID, destMount, true)) != 0 {
		return true
	}
	return false
}

// MustMount returns true if the must mount flag is set for the layerID.
func (lr *LayerRepos) MustMount(ctx context.Context, layerID digest.Digest) bool {
	if sources, ok := lr.layers[layerID]; ok {
		if sources.mustMount {
			return true
		}
	}
	return false
}

// Sources returns a list of possible repository sources for the provided layerID on the same registry, and not
// including the destination repository itself.
func (lr *LayerRepos) Sources(ctx context.Context, layerID digest.Digest, destMount ref.Ref) []ref.Ref {
	var outSources []ref.Ref
	if sources, ok := lr.layers[layerID]; ok {
		for _, s := range sources.repos {
			if s.Match(destMount, ref.RefMatchRegRepo) {
				continue
			}
			if s.Match(destMount, ref.RefMatchReg) {
				outSources = append(outSources, s)
			}
		}
	}
	if lr.bic != cache.BIC(nil) {
		bicSources := cache.LocateLayerDigest(ctx, lr.bic, layerID, destMount, true)
		outSources = append(outSources, bicSources...)
	}
	return outSources
}

// LayerReposOption defines a function that operates on a LayerRepos to configure options.
type LayerReposOption func(*LayerRepos) error

// NewMountableLayerList creates a new collection of mountable layers and their known sources.  By default,
// this is an empty list, and the MountableLayers Add() interface can be used to add repos,.
func NewMountableLayerList(opt ...LayerReposOption) (MountableLayers, error) {
	lr := &LayerRepos{layers: make(map[digest.Digest]LayerRepoList)}
	for _, o := range opt {
		if err := o(lr); err != nil {
			return nil, err
		}
	}
	return lr, nil
}
