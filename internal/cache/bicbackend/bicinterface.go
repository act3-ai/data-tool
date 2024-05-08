// Package bicbackend implements the management and storage of blob info cache data, including boltdb and memory
// implementations.
package bicbackend

import (
	"context"

	"github.com/opencontainers/go-digest"
)

// BlobInfoCache records data useful for reusing blobs, or substituting equivalent ones, to avoid unnecessary blob copies.
// It records two kinds of data:
//
//   - Sets of corresponding digest vs. untransformed digest ("DiffID") pairs:
//     One of the two digests is known to be untransformed, and a single untransformed digest may correspond to more than one transformed digest.
//     This allows matching transformed layer blobs to existing local untransformed layers (to avoid unnecessary download and untransformation),
//     or untransformed layer blobs to existing remote transformed layers (to avoid unnecessary transformation and upload)/
//
//     It is allowed to record an (untransformed digest, the same untransformed digest) correspondence, to express that the digest is known
//     to be untransformed (i.e. that a conversion from schema1 does not have to untransform the blob to compute a DiffID value).
//
//   - Known blob locations, managed by individual transports:
//     The transports call RecordKnownLocation when encountering a blob that could possibly be reused (typically in GetBlob/PutBlob/TryReusingBlob),
//     recording transport-specific information that allows the transport to reuse the blob in the future;
//     then, TryReusingBlob implementations can call CandidateLocations to look up previously recorded blob locations that could be reused.
//
//     Each transport defines its own “scopes” within which blob reuse is possible (e.g. in, the docker/distribution case, blobs
//     can be directly reused within a registry, or mounted across registries within a registry server.)
//
// None of the methods return an error indication: errors when neither reading from, nor writing to, the MemoryBlobInfoCache, should be fatal;
// users of the MemoryBlobInfoCache should just fall back to copying the blobs the usual way.
//
// NOTE: Code for BlobInfoCache is derived from github.com/containers/image/v5/pkg/blobinfocache/, ported with changes
// for generic transformers versus compression, as well as removing unneeded support for legacy database versions.
type BlobInfoCache interface {
	// UntransformedDigest returns an untransformed digest corresponding to anyDigest.
	// May return anyDigest if it is known to be untransformed.
	// Returns "" if nothing is known about the digest (it may be transformed or untransformed).
	UntransformedDigest(ctx context.Context, anyDigest digest.Digest) digest.Digest

	// RecordDigestUntransformedPair records that the untransformed version of anyDigest is untransformed.
	RecordDigestUntransformedPair(ctx context.Context, anyDigest digest.Digest, untransformed digest.Digest)

	// RecordKnownLocation records that a blob with the specified digest exists within the specified (transport, scope) scope,
	// and can be reused given the opaque location data.
	RecordKnownLocation(ctx context.Context, transport string, scope BICContentScope, digest digest.Digest, location BICLocationReference)

	// RecordDigestTransformerName records a transformer for the blob with the specified digest, or Untransformed
	// or UnknownTransformer.
	RecordDigestTransformerName(ctx context.Context, anyDigest digest.Digest, transformerName string)
	// CandidateLocations returns a sorted list of blobs and their locations
	// that could possibly be reused within the specified (transport scope) (if they still
	// exist, which is not guaranteed).
	//
	// If !canSubstitute, the returned candidates will match the submitted digest exactly; if
	// canSubstitute, data from previous RecordDigestUntransformedPair calls is used to also look
	// up variants of the blob which have the same untransformed digest.
	//
	// The TransformerName fields in returned data must never be UnknownTransformer.
	CandidateLocations(ctx context.Context, transport string, scope BICContentScope, digest digest.Digest, canSubstitute bool) []BICReplacementCandidate

	// Open sets up the BlobInfoCache for future accesses, potentially acquiring costly state. Each Open() must be paired with a Close().
	// Note that public callers may call the BlobInfoCache operations without Open()/Close().
	Open()
	// Close destroys state created by Open().
	Close()
}

const (
	// Untransformed is the value we store in a blob info BlobInfoCache to indicate that we know that
	// the blob in the corresponding location is not transformed.
	Untransformed = "untransformed"
	// UnknownTransformer is the value we store in a blob info BlobInfoCache to indicate that we don't
	// know if the blob in the corresponding location is transformed (and if so, how) or not.
	UnknownTransformer = "unknown"
)

// Content Scope type string declarations for common content categories.
const (
	// LayerContent is the content scope for layers (blobs).
	LayerContent = "Layer"
)

// BICContentScope contains a description of the type of content being recorded, such as "Layer", "Artifact", etc.  This
// reduces the search space by a broad categorization of content type.
type BICContentScope struct {
	Opaque string
}

// BICLocationReference encapsulates transport-dependent representation of a blob location within a BICContentScope.
// Each transport can store arbitrary data using BlobInfoCache.RecordKnownLocation, and ImageDestination.TryReusingBlob
// can look it up using BlobInfoCache.CandidateLocations.
type BICLocationReference struct {
	Opaque string
}

// BICReplacementCandidate is an item returned by BlobInfoCache.CandidateLocations.
type BICReplacementCandidate struct {
	Digest          digest.Digest
	TransformerName string // either the name of a known transformer algorithm, or Untransformed or UnknownTransformer
	Location        BICLocationReference
}
