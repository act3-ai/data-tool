package bicbackend

import (
	"context"
	"sync"
	"time"

	"github.com/opencontainers/go-digest"
	"k8s.io/utils/set"

	"github.com/act3-ai/go-common/pkg/logger"
)

// locationKey only exists to make lookup in knownLocations easier.
type locationKey struct {
	transport  string
	scope      BICContentScope
	blobDigest digest.Digest
}

// MemoryBlobInfoCache implements an in-memory-only BlobInfoCache.
type MemoryBlobInfoCache struct {
	mutex sync.Mutex
	// The following fields can only be accessed with mutex held.
	untransformedDigests   map[digest.Digest]digest.Digest
	digestsByUntransformed map[digest.Digest]set.Set[digest.Digest]           // stores a set of digests for each untransformed digest
	knownLocations         map[locationKey]map[BICLocationReference]time.Time // stores last known existence time for each location reference
	transformers           map[digest.Digest]string                           // stores a transformer name, or Unknown (not UnknownTransformer), for each digest
}

// MemoryCache returns a BlobInfoCache implementation which is in-memory only. This is primarily intended for tests,
// but would also apply in cases where bulk transfers are occurring and persistent storage for the BIC isn't available.
func MemoryCache() BlobInfoCache {
	return &MemoryBlobInfoCache{
		untransformedDigests:   map[digest.Digest]digest.Digest{},
		digestsByUntransformed: map[digest.Digest]set.Set[digest.Digest]{},
		knownLocations:         map[locationKey]map[BICLocationReference]time.Time{},
		transformers:           map[digest.Digest]string{},
	}
}

// Open does nothing for memory BIC.
func (mem *MemoryBlobInfoCache) Open() {
}

// Close does nothing for memory BIC.
func (mem *MemoryBlobInfoCache) Close() {
}

// UntransformedDigest returns an untransformed digest corresponding to anyDigest.
// May return anyDigest if it is known to be untransformed.
// Returns "" if nothing is known about the digest (it may be transformed or untransformed).
func (mem *MemoryBlobInfoCache) UntransformedDigest(ctx context.Context, anyDigest digest.Digest) digest.Digest {
	mem.mutex.Lock()
	defer mem.mutex.Unlock()
	return mem.untransformedDigestLocked(anyDigest)
}

// untransformedDigestLocked implements BlobInfoCache.UntransformedDigest, but must be called only with mem.mutex held.
func (mem *MemoryBlobInfoCache) untransformedDigestLocked(anyDigest digest.Digest) digest.Digest {
	if d, ok := mem.untransformedDigests[anyDigest]; ok {
		return d
	}
	// Presence in digestsByUntransformed implies that anyDigest must already refer to an untransformed digest.
	// This way we don't have to waste storage space with trivial (untransformed, untransformed) mappings
	// when we already record a (transformed, untransformed) pair.
	if s, ok := mem.digestsByUntransformed[anyDigest]; ok && s.Len() != 0 {
		return anyDigest
	}
	return ""
}

// RecordDigestUntransformedPair records that the untransformed version of anyDigest is untransformed.
// It’s allowed for anyDigest == untransformed.
func (mem *MemoryBlobInfoCache) RecordDigestUntransformedPair(ctx context.Context, anyDigest digest.Digest, untransformed digest.Digest) {
	log := logger.FromContext(ctx)
	mem.mutex.Lock()
	defer mem.mutex.Unlock()
	if previous, ok := mem.untransformedDigests[anyDigest]; ok && previous != untransformed {
		log.WarnContext(ctx, "untransformed digest for blob has been modified", "anyDigest", anyDigest, "previous", previous, "untransformed", untransformed)
	}
	mem.untransformedDigests[anyDigest] = untransformed

	anyDigestSet, ok := mem.digestsByUntransformed[untransformed]
	if !ok {
		anyDigestSet = set.New[digest.Digest]()
		mem.digestsByUntransformed[untransformed] = anyDigestSet
	}
	anyDigestSet.Insert(anyDigest)
}

// RecordKnownLocation records that a blob with the specified digest exists within the specified (transport, scope) scope,
// and can be reused given the opaque location data.
func (mem *MemoryBlobInfoCache) RecordKnownLocation(ctx context.Context, transport string, scope BICContentScope, blobDigest digest.Digest, location BICLocationReference) {
	mem.mutex.Lock()
	defer mem.mutex.Unlock()
	key := locationKey{transport: transport, scope: scope, blobDigest: blobDigest}
	locationScope, ok := mem.knownLocations[key]
	if !ok {
		locationScope = map[BICLocationReference]time.Time{}
		mem.knownLocations[key] = locationScope
	}
	locationScope[location] = time.Now() // Possibly overwriting an older entry.
}

// RecordDigestTransformerName records that the blob with the specified digest is either transformed with the specified
// algorithm, or untransformed, or that we no longer know.
func (mem *MemoryBlobInfoCache) RecordDigestTransformerName(ctx context.Context, blobDigest digest.Digest, transformerName string) {
	log := logger.FromContext(ctx)
	mem.mutex.Lock()
	defer mem.mutex.Unlock()
	if previous, ok := mem.transformers[blobDigest]; ok && previous != transformerName {
		log.WarnContext(ctx, "transformer for blob has been modified", "previous", previous, "blobDigest", blobDigest, "transformerName", transformerName)
	}
	if transformerName == UnknownTransformer {
		delete(mem.transformers, blobDigest)
		return
	}
	mem.transformers[blobDigest] = transformerName
}

// appendReplacementCandidates creates CandidateWithTime values for (transport, scope, digest), and returns the result of appending them to candidates.
func (mem *MemoryBlobInfoCache) appendReplacementCandidates(candidates []CandidateWithTime, transport string, scope BICContentScope, digest digest.Digest) []CandidateWithTime {
	locations := mem.knownLocations[locationKey{transport: transport, scope: scope, blobDigest: digest}] // nil if not present
	for l, t := range locations {
		transformerName, transformerKnown := mem.transformers[digest]
		if !transformerKnown {
			transformerName = UnknownTransformer
		}
		candidates = append(candidates, CandidateWithTime{
			Candidate: BICReplacementCandidate{
				Digest:          digest,
				TransformerName: transformerName,
				Location:        l,
			},
			LastSeen: t,
		})
	}
	return candidates
}

// CandidateLocations returns a prioritized, limited, number of blobs and their locations that could possibly be reused
// within the specified (transport scope) (if they still exist, which is not guaranteed).
//
// If !canSubstitute, the returned candidates will match the submitted digest exactly; if canSubstitute,
// data from previous RecordDigestUntransformedPair calls is used to also look up variants of the blob which have the same
// untransformed digest.
func (mem *MemoryBlobInfoCache) CandidateLocations(ctx context.Context, transport string, scope BICContentScope, primaryDigest digest.Digest, canSubstitute bool) []BICReplacementCandidate {
	return mem.candidateLocations(transport, scope, primaryDigest, canSubstitute)
}

func (mem *MemoryBlobInfoCache) candidateLocations(transport string, scope BICContentScope, primaryDigest digest.Digest, canSubstitute bool) []BICReplacementCandidate {
	mem.mutex.Lock()
	defer mem.mutex.Unlock()
	var res []CandidateWithTime
	res = mem.appendReplacementCandidates(res, transport, scope, primaryDigest)
	var untransformedDigest digest.Digest // = ""
	if canSubstitute {
		if untransformedDigest = mem.untransformedDigestLocked(primaryDigest); untransformedDigest != "" {
			if otherDigests, ok := mem.digestsByUntransformed[untransformedDigest]; ok {
				for _, d := range otherDigests.UnsortedList() {
					if d != primaryDigest && d != untransformedDigest {
						res = mem.appendReplacementCandidates(res, transport, scope, d)
					}
				}
			}
			if untransformedDigest != primaryDigest {
				res = mem.appendReplacementCandidates(res, transport, scope, untransformedDigest)
			}
		}
	}
	return SortReplacementCandidates(res, primaryDigest, untransformedDigest)
}
