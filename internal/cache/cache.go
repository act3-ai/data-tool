package cache

import (
	"errors"
	"io"

	"github.com/opencontainers/go-digest"
)

var (
	// ErrNotFound is an error returned when attempting to access a mote that isn't present in the cache.
	ErrNotFound = errors.New("specified item is not found in the cache")

	// ErrSizeMismatch is an error raised when a file write is terminated before expected.
	ErrSizeMismatch = errors.New("size of transferred item does not match expected value")
)

// ReaderAt extends the standard io.ReaderAt interface with reporting of Size and io.Closer.
type ReaderAt interface {
	io.ReaderAt
	io.Closer
	Size() int64
}

// MoteInfo is an interface representing the public information available for a mote.
type MoteInfo interface {
	// GetDigest retrieves the digest of a cached item.  This is not a calculated value, but instead refers to the
	// digest identifier of the cached item (Though the two should match)
	GetDigest() digest.Digest
	// Size retrieves the size of the data in bytes for the cached item
	Size() int64
	// GetMediaType returns a string identifying the media type for the cached item
	GetMediaType() string
}

// MoteCache is an interface for working with cache storage.
type MoteCache interface {
	// Initialize should load metadata for a cache, or instantiate a new cache if one does not exist.  The initialized
	// or created cache is returned (which may not be the one that was used to call initialize
	Initialize() error

	// IsDirty returns true if the cache has been updated since the last Initialization
	IsDirty() bool

	// Refresh triggers a reload of cache metadata from the index.  This should be called prior to performing an update.
	// The refresh operation should return true if data was updated, and false if not. An error should be returned if
	// the process fails
	Refresh() (bool, error)

	// MoteExists should return true if an item matching digest exists in the cache
	MoteExists(dgst digest.Digest) bool

	// MoteWriter returns an io.WriteCloser interface for writing data to an item. It is expected that the digest is
	// known to the cache at the time of requesting a writer (CreateMote called first), so this returns ErrNotFound if
	// the mote is not found.
	MoteWriter(dgst digest.Digest) (io.WriteCloser, error)

	// MoteReader returns an io.ReadCloser interface for reading data from a cached item based on its digest
	MoteReader(dgst digest.Digest) (io.ReadCloser, error)

	// MoteReaderAt returns an io.ReadAtCloser interface for reading data from a cached item based on its digest
	MoteReaderAt(dgst digest.Digest) (ReaderAt, error)

	// Find returns a mote information interface based on a digest search, and true if the item is found, false if not
	Find(dgst digest.Digest) (MoteInfo, bool)

	// MoteRef returns information about a mote location or reference, for instance an explicit file path
	MoteRef(dgst digest.Digest) string

	// CommitMote finalizes the addition of a mote or update to the cache. ref can be a source file path or other
	// reference information, while mote contains the mote information to be added to the index.  This silently does
	// nothing if no update is required.
	CommitMote(ref string, mote Mote, commitMode int) error

	// CreateMote creates a new item and returns an error if not successful
	CreateMote(dgst digest.Digest, mediaType string, size int64) (Mote, error)

	// RemoveMote removes an item and returns an error if not successful
	RemoveMote(dgst digest.Digest) error

	// Prune removes cached items until the size is less than or equal to maxSize
	Prune(maxSize int64) error
}

// Mote is a representation of a single item in a cache, an abstraction that may
// be backed by a file, memory, or other data blob. Access to motes and manipulation
// of them should be done via the Mote* interfaces.
type Mote struct {
	// Digest is a digest identifier for an item
	Digest digest.Digest
	// MediaType is the oci media type
	MediaType string
	// DataSize is the real size in bytes of an item
	DataSize int64

	// function overrides for read/write/close.  These can be set to customize
	// the relevant operations by a controlling MoteCache
	readFn   func([]byte) (int, error)
	readAtFn func([]byte, int64) (int, error)
	writeFn  func([]byte) (int, error)
	closeFn  func() error
}

// Read implements the io.Reader interface, optionally forwarding to a
// handler for reading data from a cache mote.
func (m Mote) Read(p []byte) (n int, err error) {
	if m.readFn == nil {
		return len(p), nil
	}
	return m.readFn(p)
}

// Write implements the io.Writer interface, optionally forwarding to a
// handler for writing data to a cache mote.  Note, a writer should write
// to a temporary location, then perform a cache update (e.g. movefile)
// during the close operation.
func (m Mote) Write(p []byte) (n int, err error) {
	if m.writeFn == nil {
		return len(p), nil
	}
	return m.writeFn(p)
}

// Close implements the io.Closer interface, optionally forwaring to a
// handler for finalizing a read/write operation.  For writers, the close
// handler should perform atomic or locked cache updates.
func (m Mote) Close() error {
	if m.closeFn == nil {
		return nil
	}
	return m.closeFn()
}

// ReadAt implements the io.ReaderAt interface, opionally forwarding to a
// handler for performing the functionality.
func (m Mote) ReadAt(p []byte, off int64) (n int, err error) {
	if m.readAtFn == nil {
		return len(p), nil
	}
	return m.readAtFn(p, off)
}

// GetDigest returns the digest of the mote.
func (m Mote) GetDigest() digest.Digest {
	return m.Digest
}

// Size returns the size of the mote.
func (m Mote) Size() int64 {
	return m.DataSize
}

// GetMediaType returns the known media type for the mote.
func (m Mote) GetMediaType() string {
	return m.MediaType
}

// Commit Modes for a call to CommitMote

// CommitCopy copies a file, leaving the original in place.
const CommitCopy = 1

// CommitMove Moves a file, removing the original.
const CommitMove = 2
