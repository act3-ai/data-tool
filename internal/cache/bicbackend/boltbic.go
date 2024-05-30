package bicbackend

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/opencontainers/go-digest"
	bolt "go.etcd.io/bbolt"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// BoltCache creates a BoltDB backed blob info cache. Note the BIC isn't opened until read/write operations.
func BoltCache(cachePath string) BlobInfoCache {
	bic := &BoltBlobInfoCache{cachePath}
	return bic
}

var (
	// transformedDigestBucket stores a mapping from any transformed digest to an original digest.
	transformedDigestBucket = []byte("transformedDigest")
	// digestTransformerBucket stores a mapping from any digest to a transformer, or Untransformed (not UnknownTransformer).
	digestTransformerBucket = []byte("digestTransformer")
	// digestByUntransformedBucket stores a bucket per digest, with the bucket containing a set of digests for that untransformed digest
	// (as a set of key=digest, value="" pairs).
	digestByUntransformedBucket = []byte("digestByUntransformed")
	// knownLocationsBucket stores a nested structure of buckets, keyed by (transport name, scope string, blob digest), ultimately containing
	// a bucket of (opaque location reference, BinaryMarshaller-encoded time.Time value).
	knownLocationsBucket = []byte("knownLocations")
)

// BoltBlobInfoCache is a BlobInfoCache implementation which uses a BoltDB file at the specified path.
//
// Note that we don’t keep the database open across operations, because that would lock the file and block any other
// users; instead, we need to open/close it for every single write or lookup.
type BoltBlobInfoCache struct {
	path string
}

// Open sets up the MemoryBlobInfoCache for future accesses, potentially acquiring costly state. Each Open() must be paired with a Close().
// Note that public callers may call the BlobInfoCache operations without Open()/Close().
func (bdc *BoltBlobInfoCache) Open() {
}

// Close destroys state created by Open().
func (bdc *BoltBlobInfoCache) Close() {
}

// view returns runs the specified fn within a read-only transaction on the database.
func (bdc *BoltBlobInfoCache) view(fn func(tx *bolt.Tx) error) (retErr error) {
	// bolt.Open(bdc.path, 0600, &bolt.Options{ReadOnly: true}) will, if the file does not exist,
	// nevertheless create it, but with an O_RDONLY file descriptor, try to initialize it, and fail — while holding
	// a read lock, blocking any future writes.
	// Hence this preliminary check, which is RACY: Another process could remove the file
	// between the Lstat call and opening the database.
	if _, err := os.Lstat(bdc.path); err != nil && os.IsNotExist(err) {
		return err
	}

	lockPath(bdc.path)
	defer unlockPath(bdc.path)
	db, err := bolt.Open(bdc.path, 0o600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("failed to open bolt db %w", err)
	}
	defer func() {
		if err := db.Close(); retErr == nil && err != nil {
			retErr = fmt.Errorf("close failure for boltdb %w", err)
		}
	}()

	err = db.View(fn)
	if err != nil {
		return fmt.Errorf("view failure for boltdb %w", err)
	}
	return nil
}

// update returns runs the specified fn within a read-write transaction on the database.
func (bdc *BoltBlobInfoCache) update(fn func(tx *bolt.Tx) error) (retErr error) {
	lockPath(bdc.path)
	defer unlockPath(bdc.path)
	db, err := bolt.Open(bdc.path, 0o600, nil)
	if err != nil {
		return fmt.Errorf("open failure for boltdb in update %w", err)
	}
	defer func() {
		if err := db.Close(); retErr == nil && err != nil {
			retErr = fmt.Errorf("close failure for boltdb in update %w", err)
		}
	}()

	err = db.Update(fn)
	if err != nil {
		return fmt.Errorf("update failure for boltdb%w", err)
	}
	return nil
}

// untransformedDigest implements BlobInfoCache.UntransformedDigest within the provided read-only transaction.
func (bdc *BoltBlobInfoCache) untransformedDigest(tx *bolt.Tx, anyDigest digest.Digest) digest.Digest {
	if b := tx.Bucket(transformedDigestBucket); b != nil {
		if untransformedBytes := b.Get([]byte(anyDigest.String())); untransformedBytes != nil {
			d, err := digest.Parse(string(untransformedBytes))
			if err == nil {
				return d
			}
		}
	}
	// Presence in digestByUntransformedBucket implies that anyDigest must already refer to an untransformed digest.
	// This way we don't have to waste storage space with trivial (untransformed, untransformed) mappings
	// when we already record a (transformed, untransformed) pair.
	if b := tx.Bucket(digestByUntransformedBucket); b != nil {
		if b = b.Bucket([]byte(anyDigest.String())); b != nil {
			c := b.Cursor()
			if k, _ := c.First(); k != nil { // The bucket is non-empty
				return anyDigest
			}
		}
	}
	return ""
}

// UntransformedDigest returns an untransformed digest corresponding to anyDigest.
// May return anyDigest if it is known to be untransformed.
// Returns "" if nothing is known about the digest (it may be transformed or untransformed).
func (bdc *BoltBlobInfoCache) UntransformedDigest(ctx context.Context, anyDigest digest.Digest) digest.Digest {
	var res digest.Digest
	if err := bdc.view(func(tx *bolt.Tx) error {
		res = bdc.untransformedDigest(tx, anyDigest)
		return nil
	}); err != nil { // Including os.IsNotExist(err)
		return ""
	}
	return res
}

// RecordDigestUntransformedPair records that the untransformed version of anyDigest is untransformed.
// It’s allowed for anyDigest == untransformed.
// WARNING: Only call this for LOCALLY VERIFIED data; don’t record a digest pair just because some remote author claims so (e.g.
// because a manifest/config pair exists); otherwise the MemoryBlobInfoCache could be poisoned and allow substituting unexpected blobs.
// (Eventually, the DiffIDs in image config could detect the substitution, but that may be too late, and not all image formats contain that data.)
func (bdc *BoltBlobInfoCache) RecordDigestUntransformedPair(ctx context.Context, anyDigest digest.Digest, untransformed digest.Digest) {
	log := logger.FromContext(ctx)
	_ = bdc.update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(transformedDigestBucket)
		if err != nil {
			return fmt.Errorf("bucket creation failure 1 for boltdb in RecordDigestUntransformedPair %w", err)
		}
		key := []byte(anyDigest.String())
		if previousBytes := b.Get(key); previousBytes != nil {
			previous, err := digest.Parse(string(previousBytes))
			if err != nil {
				return fmt.Errorf("digest parse error in RecordDigestUntransformedPair %w", err)
			}
			if previous != untransformed {
				log.WarnContext(ctx, "untransformed digest for blob has been modified", "anyDigest", anyDigest, "previous", previous, "untransformed", untransformed)
			}
		}
		if err := b.Put(key, []byte(untransformed.String())); err != nil {
			return fmt.Errorf("data put failure 1 in RecordDigestUntransformedPair %w", err)
		}

		b, err = tx.CreateBucketIfNotExists(digestByUntransformedBucket)
		if err != nil {
			return fmt.Errorf("bucket creation failure 2 for boltdb in RecordDigestUntransformedPair %w", err)
		}
		b, err = b.CreateBucketIfNotExists([]byte(untransformed.String()))
		if err != nil {
			return fmt.Errorf("bucket creation failure 3 for boltdb in RecordDigestUntransformedPair %w", err)
		}
		if err := b.Put([]byte(anyDigest.String()), []byte{}); err != nil {
			return fmt.Errorf("data put failure 2 in RecordDigestUntransformedPair %w", err)
		}
		return nil
	})
}

// RecordDigestTransformerName records that the blob with digest anyDigest was transformed with the specified
// transformer, or is Untransformed.
func (bdc *BoltBlobInfoCache) RecordDigestTransformerName(ctx context.Context, anyDigest digest.Digest, transformerName string) {
	log := logger.FromContext(ctx)
	_ = bdc.update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(digestTransformerBucket)
		if err != nil {
			return fmt.Errorf("bucket creation failure for boltdb in RecordDigestTransformerName %w", err)
		}
		key := []byte(anyDigest.String())
		if previousBytes := b.Get(key); previousBytes != nil {
			if string(previousBytes) != transformerName {
				log.WarnContext(ctx, "transformer for blob has been modified", "previous", string(previousBytes), "blobDigest", anyDigest, "transformerName", transformerName)
			}
		}
		if transformerName == UnknownTransformer {
			err = b.Delete(key)
			if err != nil {
				return fmt.Errorf("data delete error in RecordDigestTransformerName %w", err)
			}
			return nil
		}
		err = b.Put(key, []byte(transformerName))
		if err != nil {
			return fmt.Errorf("data write error in RecordDigestTransformerName %w", err)
		}
		return nil
	})
}

// RecordKnownLocation records that a blob with the specified digest exists within the specified (transport, scope) scope,
// and can be reused given the opaque location data.
func (bdc *BoltBlobInfoCache) RecordKnownLocation(ctx context.Context, transport string, scope BICContentScope, blobDigest digest.Digest, location BICLocationReference) {
	_ = bdc.update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(knownLocationsBucket)
		if err != nil {
			return fmt.Errorf("bucket creation failure 1 for boltdb in RecordKnownLocation %w", err)
		}
		b, err = b.CreateBucketIfNotExists([]byte(transport))
		if err != nil {
			return fmt.Errorf("bucket creation failure 2 for boltdb in RecordKnownLocation %w", err)
		}
		b, err = b.CreateBucketIfNotExists([]byte(scope.Opaque))
		if err != nil {
			return fmt.Errorf("bucket creation failure 3 for boltdb in RecordKnownLocation %w", err)
		}
		b, err = b.CreateBucketIfNotExists([]byte(blobDigest.String()))
		if err != nil {
			return fmt.Errorf("bucket creation failure 4 for boltdb in RecordKnownLocation %w", err)
		}
		value, err := time.Now().MarshalBinary()
		if err != nil {
			return fmt.Errorf("failure to format time data in boltdb RecordKnownLocation %w", err)
		}
		if err := b.Put([]byte(location.Opaque), value); err != nil {
			return fmt.Errorf("data write error in boltdb RecordKnownLocation %w", err)
		}
		return nil
	})
}

// appendReplacementCandidates creates prioritize.CandidateWithTime values for digest in scopeBucket with corresponding transformation info from transformationBucket (if transformationBucket is not nil), and returns the result of appending them to candidates.
func (bdc *BoltBlobInfoCache) appendReplacementCandidates(candidates []CandidateWithTime, scopeBucket, transformBucket *bolt.Bucket, digest digest.Digest) []CandidateWithTime {
	digestKey := []byte(digest.String())
	b := scopeBucket.Bucket(digestKey)
	if b == nil {
		return candidates
	}
	transformerName := UnknownTransformer
	if transformBucket != nil {
		// the bucket won't exist if the MemoryBlobInfoCache was created by a v1 implementation and
		// hasn't yet been updated by a v2 implementation
		if transformerNameValue := transformBucket.Get(digestKey); len(transformerNameValue) > 0 {
			transformerName = string(transformerNameValue)
		}
	}

	_ = b.ForEach(func(k, v []byte) error {
		t := time.Time{}
		if err := t.UnmarshalBinary(v); err != nil {
			return fmt.Errorf("unmarshaling error for time data %w", err)
		}
		candidates = append(candidates, CandidateWithTime{
			Candidate: BICReplacementCandidate{
				Digest:          digest,
				TransformerName: transformerName,
				Location:        BICLocationReference{Opaque: string(k)},
			},
			LastSeen: t,
		})
		return nil
	})
	return candidates
}

// CandidateLocations returns a sorted, number of blobs and their locations that could possibly be reused
// within the specified (transport scope) (if they still exist, which is not guaranteed).
//
// If !canSubstitute, the returned candidates will match the submitted digest exactly; if canSubstitute,
// data from previous RecordDigestUntransformedPair calls is used to also look up variants of the blob which have the same
// untransformed digest.
func (bdc *BoltBlobInfoCache) CandidateLocations(ctx context.Context, transport string, scope BICContentScope, primaryDigest digest.Digest, canSubstitute bool) []BICReplacementCandidate {
	return bdc.candidateLocations(transport, scope, primaryDigest, canSubstitute, true)
}

func (bdc *BoltBlobInfoCache) candidateLocations(transport string, scope BICContentScope, primaryDigest digest.Digest, canSubstitute, requireTransformInfo bool) []BICReplacementCandidate { //nolint: gocognit
	var res []CandidateWithTime
	var utDigestValue digest.Digest // = ""
	if err := bdc.view(func(tx *bolt.Tx) error {
		scopeBucket := tx.Bucket(knownLocationsBucket)
		if scopeBucket == nil {
			return nil
		}
		scopeBucket = scopeBucket.Bucket([]byte(transport))
		if scopeBucket == nil {
			return nil
		}
		scopeBucket = scopeBucket.Bucket([]byte(scope.Opaque))
		if scopeBucket == nil {
			return nil
		}
		// transformedBucket won't have been created if previous writers never recorded info about transformation,
		// and we don't want to fail just because of that
		transformedBucket := tx.Bucket(digestTransformerBucket)

		res = bdc.appendReplacementCandidates(res, scopeBucket, transformedBucket, primaryDigest)
		if canSubstitute {
			if utDigestValue = bdc.untransformedDigest(tx, primaryDigest); utDigestValue != "" {
				b := tx.Bucket(digestByUntransformedBucket)
				if b != nil {
					b = b.Bucket([]byte(utDigestValue.String()))
					if b != nil {
						if err := b.ForEach(func(k, _ []byte) error {
							d, err := digest.Parse(string(k))
							if err != nil {
								return fmt.Errorf("digest parse error in RecordDigestUntransformedPair %w", err)
							}
							if d != primaryDigest && d != utDigestValue {
								res = bdc.appendReplacementCandidates(res, scopeBucket, transformedBucket, d)
							}
							return nil
						}); err != nil {
							return fmt.Errorf("callback function return error %w", err)
						}
					}
				}
				if utDigestValue != primaryDigest {
					res = bdc.appendReplacementCandidates(res, scopeBucket, transformedBucket, utDigestValue)
				}
			}
		}
		return nil
	}); err != nil { // Including os.IsNotExist(err)
		return []BICReplacementCandidate{}
	}

	return SortReplacementCandidates(res, primaryDigest, utDigestValue)
}

// NOTE: we must limit access to db to one thread at a time See https://www.sqlite.org/src/artifact/c230a7a24?ln=994-1081

// pathLock contains a lock for a specific BoltDB database path.
type pathLock struct {
	refCount int64      // Number of threads/goroutines owning or waiting on this lock.  Protected by global pathLocksMutex, NOT by the mutex field below!
	mutex    sync.Mutex // Owned by the thread/goroutine allowed to access the BoltDB database.
}

var (
	// pathLocks contains a lock for each currently open file.
	// This must be global so that independently created instances of boltDBCache exclude each other.
	// The map is protected by pathLocksMutex.
	pathLocks      = map[string]*pathLock{}
	pathLocksMutex = sync.Mutex{}
)

// lockPath obtains the pathLock for path.
// The caller must call unlockPath eventually.
func lockPath(path string) {
	pl := func() *pathLock { // A scope for defer
		pathLocksMutex.Lock()
		defer pathLocksMutex.Unlock()
		pl, ok := pathLocks[path]
		if ok {
			pl.refCount++
		} else {
			pl = &pathLock{refCount: 1, mutex: sync.Mutex{}}
			pathLocks[path] = pl
		}
		return pl
	}()
	pl.mutex.Lock()
}

// unlockPath releases the pathLock for path.
func unlockPath(path string) {
	pathLocksMutex.Lock()
	defer pathLocksMutex.Unlock()
	pl, ok := pathLocks[path]
	if !ok {
		// Should this return an error instead? BlobInfoCache ultimately ignores errors…
		panic(fmt.Sprintf("Internal error: unlocking nonexistent lock for path %s", path))
	}
	pl.mutex.Unlock()
	pl.refCount--
	if pl.refCount == 0 {
		delete(pathLocks, path)
	}
}
