// Package storage defines a storage provider interface for abstracting storage operations during transfers.  A storage
// provider provides writer and reader interface functions, keying data objects by an oci descriptor.  These functions
// are implemented by a storage management object, such as a BottleFileCache or OCIFileCache
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	notaryreg "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/labels"
	orascontent "oras.land/oras-go/v2/content"

	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/tool/internal/archive"
	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
	"gitlab.com/act3-ai/asce/data/tool/internal/util"
)

// CacheProvider is an interface describing an object that can facilitate retrieval of a cache manager and storage path.
type CacheProvider interface {
	// GetCache provides a cache management interface for working with a cache
	GetCache() cache.MoteCache
	// GetPath returns a string containing a path or reference to cache storage
	GetPath() string
}

// Provider defines an interface that provides Layer reading and writing capability based on context and
// Descriptors, as well as the ability to determine if a Layer already exists, and to copy from an alternate source
// such as a cache.
type Provider interface {
	// LayerExists returns true if a cache contains data associated with the descriptor
	LayerExists(ctx context.Context, descriptor ocispec.Descriptor) bool

	// CopyFromCache copy descriptor from the cache and put it in destination and returns success and any error
	CopyFromCache(ctx context.Context, descriptor ocispec.Descriptor, destination string) (bool, error)
}

// DataStore provides content from the file system.
type DataStore struct {
	DisableOverwrite          bool
	AllowPathTraversalOnWrite bool

	// Reproducible enables stripping times from added files
	// Reproducible bool

	root       string
	descriptor *sync.Map // map[digest.Digest]ocispec.Descriptor

	configData []byte
	dloc       CacheProvider

	// predecessors is a map of subject to referring descriptors
	predecessors *sync.Map // map[digest.Digest][]ocispec.Descriptor
	// looseData is a map of digest to bytes, for handling uncached/loose data
	looseData *sync.Map // map[digest.Digest][]byte

	// disable in stream digest calculation
	DisableDigestCalc bool
}

// PartInfo is an interface for oci file entry data retrieval.
type PartInfo interface {
	// GetName returns a file name
	GetName() string
	// GetContentSize returns the size of a file in the bottle config (uncompressed, unarchived form)
	GetContentSize() int64
	// GetContentDigest returns the digest of a file as it exists in uncompressed/unarchived state, and is the digest
	// referened in bottle config
	GetContentDigest() digest.Digest
	// GetLabels returns a map of label keys to values associated with a file
	GetLabels() labels.Set

	// GetLayerSize is the size of the layer (often compresses size)
	GetLayerSize() int64
	// GetLayerDigest returns the digest of a file, as stored in cache or on the server.  This can be the compressed and/or
	// archived form of the file, and in OCI terms is the LayerDigest
	GetLayerDigest() digest.Digest
	// GetMediaType returns the content media type string for a file
	GetMediaType() string

	// GetModTime returns the last known modified time of a file, and can be used to determine if a file on disk has
	// been modified since the last bottle operation
	GetModTime() time.Time
}

// NewDataStore creates a new data store.
func NewDataStore(dataloc CacheProvider) *DataStore {
	return &DataStore{
		root:         dataloc.GetPath(),
		descriptor:   &sync.Map{},
		looseData:    &sync.Map{},
		predecessors: &sync.Map{},
		dloc:         dataloc,
	}
}

// verifying that DataStore implements the oras.content.Storage and oras.content.ReadOnlyStorage interfaces.
var _ orascontent.Storage = &DataStore{}

// Fetch implements the oras.content.ReadOnlyStorage interface for providing content through a ReadCloser.
func (s *DataStore) Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	// if loose data is available matching the descriptor send that
	if byteAny, isLoose := s.looseData.Load(target.Digest); isLoose {
		byteData := byteAny.([]byte)
		return io.NopCloser(bytes.NewReader(byteData)), nil
	}

	c := s.dloc.GetCache()
	if c == nil {
		return nil, fmt.Errorf("cache not configured for oras accessed data store fetch call")
	}
	rdr, err := c.MoteReader(target.Digest)
	if errors.Is(err, cache.ErrNotFound) {
		return nil, fmt.Errorf("layer not found in cache: %s", target.Digest)
	}

	return rdr, err
}

// Exists implements the oras.content.ReadOnlyStorage interface for checking whether a target descriptor exists in
// the storage (underlying cache in this case).
//
// Oras calls this function immediately upon discovering a layer in a graph, and skips transfer of the layer if found.
// This means that preCopy and postCopy are not called by Oras, which means we cannot perform cache actions to move
// data to the final destination if Exists returns true.  Thus, to enable pre- and post-copy activity, set the
// DelayExistCheck property to true.  Internal DataStore functions do not call Exists, so this will only affect external
// calls such as the ones from Oras.
func (s *DataStore) Exists(ctx context.Context, target ocispec.Descriptor) (bool, error) {
	switch {
	case target.MediaType == ocispec.MediaTypeImageManifest:
		// manifests are not cached, but true will skip an oras copy
		return false, nil
	case mediatype.IsBottleConfig(target.MediaType) ||
		target.MediaType == notaryreg.ArtifactTypeNotation:
		// configs (bottle or signatures) are not cached, instead added as loose data
		if _, hasLoose := s.looseData.Load(target.Digest); hasLoose {
			return true, nil
		}
		return false, fmt.Errorf("bottle or signature configs must be added as loose data")
	case target.MediaType == jws.MediaTypeEnvelope ||
		target.MediaType == cose.MediaTypeEnvelope:
		// signature layers are not cached, instead added as loose data
		if _, hasLoose := s.looseData.Load(target.Digest); hasLoose {
			return true, nil
		}
		return false, fmt.Errorf("signature layers must be added as loose data")
	case mediatype.IsLayer(target.MediaType):
		c := s.dloc.GetCache()
		if c == nil {
			return false, fmt.Errorf("cache not configured for oras accessed data store exists call")
		}
		// TODO: underlying cache does not return errors, which can be refactored, but other callers will need to handle
		// the error
		return c.MoteExists(target.Digest), nil
	default:
		return false, fmt.Errorf("unexpected mediatype encountered when checking data store '%s'", target.MediaType)
	}
}

// Push implements the oras.content.Push interface method for writing content to storage.  The data is copied from the
// content reader synchronously to the cache.
func (s *DataStore) Push(ctx context.Context, expected ocispec.Descriptor, contentRead io.Reader) error {
	c := s.dloc.GetCache()
	if c == nil {
		return fmt.Errorf("cache not configured for oras accessed data store push call")
	}
	_, err := c.CreateMote(expected.Digest, expected.MediaType, expected.Size)
	if err != nil {
		return err
	}
	w, err := c.MoteWriter(expected.Digest)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, contentRead)
	w.Close()
	if err != nil {
		return fmt.Errorf("copying mote: %w", err)
	}

	return nil
}

// Predecessors returns a list of known referring descriptors to the target -- these descriptors are essentially
// manifests with a subject field containing the requested node. This implements the PredecessorLister interface for
// oras, which is part of the ReadOnlyGraphStorage interface.
func (s *DataStore) Predecessors(ctx context.Context, subject ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	if listAny, found := s.predecessors.Load(subject.Digest); found {
		descList := listAny.([]ocispec.Descriptor)
		return descList, nil
	}
	return nil, nil
}

// AddLooseData adds data from an io.Reader to a collection of loose data that can be accessed in addition to cached
// data.  After reading, the size and digest are returned within the returned descriptor.  If subject is not nil, the
// data is treated as a referring object to another descriptor (eg, manifest with a subject field matching the subject),
// and a referring link to the subject is established for future CopyExtendedGraph operations.
func (s *DataStore) AddLooseData(ctx context.Context, reader io.Reader, mediaType string, subject *ocispec.Descriptor) (ocispec.Descriptor, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to read from loose file reader: %w", err)
	}
	dgst := digest.FromBytes(data)
	size := len(data)
	s.looseData.Store(dgst, data)

	desc := ocispec.Descriptor{
		MediaType:   mediaType,
		Digest:      dgst,
		Size:        int64(size),
		Annotations: make(map[string]string),
	}

	if subject != nil {
		updatedList := []ocispec.Descriptor{desc}
		if listAny, found := s.predecessors.Load(subject.Digest); found {
			updatedList = append(updatedList, listAny.([]ocispec.Descriptor)...)
		}
		s.predecessors.Store(subject.Digest, updatedList)
	}

	return desc, nil
}

// Close frees up resources used by the data store.
func (s *DataStore) Close() error {
	return nil
}

// ErrUnknownLayerMediaType is the error if the layer media type is unknown.
var ErrUnknownLayerMediaType = errors.New("unknown layer media type")

// HandlePartMedia copies the layer into the part file/directory given by partName.
func (s *DataStore) HandlePartMedia(ctx context.Context, desc ocispec.Descriptor, partName string) error {
	destPath := filepath.Join(s.root, filepath.FromSlash(partName))

	cman := s.dloc.GetCache()
	cachefile := cman.MoteRef(desc.Digest)

	switch desc.MediaType {
	case mediatype.MediaTypeLayerTar:
		return archive.ExtractTar(ctx, cachefile, destPath)
	case mediatype.MediaTypeLayerTarOld, mediatype.MediaTypeLayerTarLegacy:
		return archive.ExtractTarCompat(ctx, cachefile, s.root)
	case mediatype.MediaTypeLayerTarZstd:
		return archive.ExtractTarZstd(ctx, cachefile, destPath)
	case mediatype.MediaTypeLayerTarZstdOld, mediatype.MediaTypeLayerTarZstdLegacy:
		return archive.ExtractTarZstdCompat(ctx, cachefile, s.root)
	case mediatype.MediaTypeLayerZstd:
		return archive.ExtractZstd(ctx, cachefile, destPath)
	case mediatype.MediaTypeLayerTarGzip, mediatype.MediaTypeLayerTarGzipLegacy:
		return errors.New("gzip is not implemented")
	case mediatype.MediaTypeLayer, mediatype.MediaTypeLayerRawOld, mediatype.MediaTypeLayerRawLegacy:
		return util.CopyFile(cachefile, destPath)
	default:
		return ErrUnknownLayerMediaType
	}
}

// CopyFromCache allows data store to preempt data transfer by locating data in a cache
// and handling the media from there, according to a defined HandlePartMedia.
func (s *DataStore) CopyFromCache(ctx context.Context, desc ocispec.Descriptor, name string) (bool, error) {
	// check cache for existing data and preempt operations by handling the media from there
	cman := s.dloc.GetCache()
	if cman.MoteExists(desc.Digest) {
		if err := s.HandlePartMedia(ctx, desc, name); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// LayerExists returns true if the layer is present in the data store cache.
func (s *DataStore) LayerExists(ctx context.Context, descriptor ocispec.Descriptor) bool {
	cman := s.dloc.GetCache()
	return cman.MoteExists(descriptor.Digest)
}
