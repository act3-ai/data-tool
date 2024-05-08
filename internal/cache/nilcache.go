package cache

import (
	"io"

	"github.com/opencontainers/go-digest"
)

// NilCache is an empty implementation of MoteCache that provides empty functionality
// for cases when caching is disabled.
type NilCache struct {
}

// Initialize does nothing.
func (nc *NilCache) Initialize() error {
	return nil
}

// IsDirty returns false.
func (nc *NilCache) IsDirty() bool {
	return false
}

// Refresh does nothing.
func (nc *NilCache) Refresh() (bool, error) {
	return false, nil
}

// MoteExists returns false.
func (nc *NilCache) MoteExists(dgst digest.Digest) bool {
	return false
}

// MoteWriter returns an empty mote.
func (nc *NilCache) MoteWriter(dgst digest.Digest) (io.WriteCloser, error) {
	return Mote{}, nil
}

// MoteReader returns an empty mote.
func (nc *NilCache) MoteReader(dgst digest.Digest) (io.ReadCloser, error) {
	return Mote{}, nil
}

// MoteReaderAt returns an empty mote.
func (nc *NilCache) MoteReaderAt(dgst digest.Digest) (ReaderAt, error) {
	return Mote{}, nil
}

// Find returns no info, and false to indicate nothing was found.
func (nc *NilCache) Find(dgst digest.Digest) (MoteInfo, bool) {
	return nil, false
}

// MoteRef returns an empty string.
func (nc *NilCache) MoteRef(dgst digest.Digest) string {
	return ""
}

// CommitMote does nothing.
func (nc *NilCache) CommitMote(ref string, mote Mote, commitMode int) error {
	return nil
}

// CreateMote creates a mote with the provided info. The mote created with this contains the
// provided information, but is otherwise non functional.
func (nc *NilCache) CreateMote(dgst digest.Digest, mediaType string, size int64) (Mote, error) {
	m := Mote{
		Digest:    dgst,
		MediaType: mediaType,
		DataSize:  size,
	}

	return m, nil
}

// RemoveMote does nothing.
func (nc *NilCache) RemoveMote(dgst digest.Digest) error {
	return nil
}

// Prune does nothing.
func (nc *NilCache) Prune(maxSize int64) error {
	return nil
}
