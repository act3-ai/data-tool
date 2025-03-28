package oci

import (
	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
)

const (
	// ArtifactTypeSyncManifest is the artifact type for an sync manifest.
	ArtifactTypeSyncManifest = "application/vnd.act3-ace.git.repo.v1+json"

	// ArtifactTypeLFSManifest is the artifact type for an git-lfs manifest.
	ArtifactTypeLFSManifest = "application/vnd.act3-ace.git-lfs.repo.v1+json"

	// MediaTypeSyncConfig is the media type for a sync config.
	MediaTypeSyncConfig = "application/vnd.act3-ace.git.config.v1+json"

	// MediaTypeLFSConfig is the media type for a git-lfs config. Currently not used.
	MediaTypeLFSConfig = "application/vnd.act3-ace.git-lfs.config.v1+json"

	// MediaTypeBundleLayer is the media type for a git bundle stored as an oci layer.
	MediaTypeBundleLayer = "application/vnd.act3-ace.git.bundle.v1"

	// MediaTypeLFSLayer is the media type used for git-lfs layers.
	MediaTypeLFSLayer = "application/vnd.act3-ace.git-lfs.object.v1"

	// AnnotationDTVersion is the key for the annotation to denote the ace-dt version used during gather.
	AnnotationDTVersion = "vnd.act3-ace.data.version"
)

// Config holds information related to the git repository's references.
type Config struct {
	// Typically, bundles will explicitly contain references. However,
	// some workflows result in only reference update(s) which requires an additional
	// bundle or rebuilding the bundle containing the referenced commit. Neither approach
	// is ideal, so we choose to update the references from the config in FromOCI.Run.

	Refs References `json:"refs"`
}

// LFSConfig holds the oid information contained in the LFS manifest.
// Currently empty.
type LFSConfig struct {
	// LFS bundles may be useful in the future. It may be useful to store
	// a which bundle an LFS file is in, similar to the current git bundle mechanism.
	// See https://git.act3-ace.com/ace/data/tool/-/issues/503.

	// Objects map[string]string `json:"objects"` // layer digest : OID
}

// References hold the mappings tag and head refs to tuples of commit and layer digest pairs.
type References struct {
	Tags  map[string]ReferenceInfo `json:"tags"`  // tag reference : (commit, layer)
	Heads map[string]ReferenceInfo `json:"heads"` // head reference : (commit, layer)
}

// ReferenceInfo holds informations about git references stored in bundle layers.
type ReferenceInfo struct {
	// The additional layer information allows for optimizations on the rebuild (FromOCI) side,
	// such that one may determine the least number of bundle layers to fetch (from OCI) and git pull from.

	Commit cmd.Commit    `json:"commit"` // commit pointed to by reference
	Layer  digest.Digest `json:"layer"`  // OCI layer, the bundle with the commit
}
