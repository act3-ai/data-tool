// Package cache implements local cached storage of part data.
//
// The cache is accessed using an interface defined as MoteCache, and implemented with BottleFileCache
// Generally, the cache is intended to be used in a pull-through fashion, with data being transferred first to the
// cache and then automatically to its final destination.  Items in the cache are identified using a digest, though the
// source of those digests are generally within an oci descriptor.
package cache

import (
	"context"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"gitlab.com/act3-ai/asce/data/tool/internal/cache/bicbackend"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
)

// SourceProvider is an interface enabling the retrieval of file sources, which is a mapping of digest string to
// list of OCI Refs in string format, allowing content to be retrieved from a remote source.
type SourceProvider interface {
	// GetSources returns a map of digests to a list of OCI references (in string format) where each digest is known to
	// exist. This allows the content to be retrieved from a remote source.
	GetSources() map[digest.Digest][]string
}

// BIC is a type alias for a blob info cache.
type BIC bicbackend.BlobInfoCache

// TransportFromImageName extracts a transport string from the provided ref string, if one exists.  This allows
// filtering on location, such as images located in an oci-layout dir, archive, etc.  If a transport string isn't
// found in the refString, (format is transport:ref), or if an unknown transport is specified, "oci" is returned.
func TransportFromImageName(refString string) string {
	parts := strings.SplitN(refString, ":", 2)
	if len(parts) == 2 {
		switch parts[0] {
		case "docker":
			fallthrough
		case "oci":
			return "oci"
		case "dir":
			return "dir"
		case "docker-archive":
			return "docker-archive"
		case "oci-archive":
			return "oci-archive"
		case "tarball":
			return "tarball"
		default:
			return "oci"
		}
	}
	return "oci"
}

// NewCache returns the active blob info cache.  Opening and initialization is delayed until access.  If cachePath is
// an empty string, a memory cache is returned, otherwise it is a persistent cache backed by boltdb.
func NewCache(cachePath string) BIC {
	if cachePath == "" {
		return bicbackend.MemoryCache()
	}
	return bicbackend.BoltCache(cachePath)
}

// LayerBlobInfo is a structure implementing the SourceProvider interface, enabling lookup and server-to server
// transfer for layers based on recorded source information in blobinfocache.
type LayerBlobInfo struct {
	layerIDs []digest.Digest
}

// GetSources returns a map of layerID to known source list for all layerIDs in a LayerBlobInfo, implementing the
// SourceProvider interface.
func (lbi *LayerBlobInfo) GetSources(ctx context.Context, bic BIC) map[digest.Digest][]string {
	sources := make(map[digest.Digest][]string)
	destRef := ref.RepoFromString("https://any.host/any/reg")
	for _, l := range lbi.layerIDs {
		refs := LocateLayerDigest(ctx, bic, l, destRef, false)
		var refStrs []string
		for _, r := range refs {
			refStrs = append(refStrs, r.String())
		}
		sources[l] = refStrs
	}
	return sources
}

// RecordLayerSource adds the layer described by desc to the local blob info cache, with the provided srcRef as the
// known source location.  The desc should describe the manifest-level layer (eg, possibly compressed/encrypted).
func RecordLayerSource(ctx context.Context, bic BIC, desc ocispec.Descriptor, srcRef ref.Ref) {
	refStr := srcRef.URL()
	srcTransport := TransportFromImageName(refStr)
	// TODO do we really want to do this?
	if bic != nil {
		bic.RecordKnownLocation(ctx, srcTransport, bicbackend.BICContentScope{Opaque: bicbackend.LayerContent}, desc.Digest, bicbackend.BICLocationReference{Opaque: refStr})
	}
}

// LocateLayer returns a set of bottle Refs containing a list of known locations for the provided layer descriptor.  The
// layer descriptor should be manifest-level (eg. possibly compressed/encrypted).  Candidates returned will be located
// on the destination registry if matchReg is true.
func LocateLayer(ctx context.Context, bic BIC, desc ocispec.Descriptor, destRef ref.Ref, matchReg bool) []ref.Ref {
	refStr := destRef.String()
	destTransport := TransportFromImageName(refStr)
	candidates := bic.CandidateLocations(ctx, destTransport, bicbackend.BICContentScope{Opaque: bicbackend.LayerContent}, desc.Digest, false)

	retRefs := make([]ref.Ref, 0, len(candidates))
	for _, c := range candidates {
		r, err := ref.FromString(c.Location.Opaque)
		if err != nil {
			continue
		}
		if matchReg && r.Reg != destRef.Reg {
			continue
		}
		retRefs = append(retRefs, r)
	}
	return retRefs
}

// LocateLayerDigest performs a layer location query based on a digest representation of the layer digest.
// This works the same as LocateLayer.
func LocateLayerDigest(ctx context.Context, bic BIC, dgst digest.Digest, destRef ref.Ref, matchReg bool) []ref.Ref {
	desc := ocispec.Descriptor{Digest: dgst}
	return LocateLayer(ctx, bic, desc, destRef, matchReg)
}
