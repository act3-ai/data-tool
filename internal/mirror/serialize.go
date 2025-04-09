package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/act3-ai/data-tool/internal/mirror/blockbuf"
	"github.com/act3-ai/data-tool/internal/mirror/encoding"
	"github.com/act3-ai/data-tool/internal/print"
	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/act3-ai/go-common/pkg/ioutil"
)

// serializationVersion is the serialization format version.
// If we change how we write data to tape then we need to increment this value.
const serializationVersion = 2

// ResumeFromLedger contains the data necessary to resume from a checkpoint.
type ResumeFromLedger struct {
	// Path is the path to the local file where the checkpoint ledger is stored
	Path string

	// Offset number of bytes to resume from.  We assume all blobs referenced before this number exist at the destination.
	Offset int64
}

// BlockBufOptions define the requirements to serialize with blockbuf.
type BlockBufOptions struct {
	Buffer        int
	BlockSize     int
	HighWaterMark int
}

// SerializeOptions define the requirements to run a serialize operation.
type SerializeOptions struct {
	BufferOpts          BlockBufOptions
	ExistingCheckpoints []ResumeFromLedger
	ExistingImages      []string
	Recursive           bool
	RepoFunc            func(context.Context, string) (*remote.Repository, error)
	SourceStorage       content.ReadOnlyStorage
	SourceReference     string
	Compression         string
	SourceDesc          ocispec.Descriptor
	WithManifestJSON    bool
}

// Serialize takes the artifact created in a gather operation and serializes it to tar.
func Serialize(ctx context.Context, destFile, checkpointFile, dataToolVersion string, opts SerializeOptions) error {
	// TODO we fetch manifests more than once.  We should cache them by wrapping repo.
	rootUI := ui.FromContextOrNoop(ctx)

	defer rootUI.Info("Serialize action completed")

	g, ctx := errgroup.WithContext(ctx)

	progress := rootUI.SubTaskWithProgress("Writing to archive")
	defer progress.Complete()

	// open the destination file/tape carefully to append only
	file, err := os.OpenFile(destFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	if err != nil {
		return fmt.Errorf("destination file: %w", err)
	}
	defer file.Close()

	// setup the destination
	var dest io.WriteCloser = file

	if opts.BufferOpts.Buffer > 0 {
		// Use the blockbuf
		r, w := io.Pipe()
		cw := new(ioutil.WriterCounter)
		pr := io.TeeReader(r, cw)
		defer w.Close() // this is closed at the end of the function and the error is checked.

		g.Go(func() error {
			return blockbuf.Copy(file, pr, opts.BufferOpts.Buffer, opts.BufferOpts.BlockSize, opts.BufferOpts.HighWaterMark)
		})
		dest = w
	}

	var serializer *encoding.OCILayoutSerializer
	if checkpointFile != "" {
		ledger, err := os.Create(checkpointFile)
		if err != nil {
			return fmt.Errorf("create checkpoint ledger file: %w", err)
		}
		defer ledger.Close()
		serializer, err = encoding.NewOCILayoutSerializerWithLedger(dest, ledger, opts.Compression)
		if err != nil {
			return fmt.Errorf("creating serializer: %w", err)
		}
	} else {
		serializer, err = encoding.NewOCILayoutSerializer(dest, opts.Compression)
		if err != nil {
			return fmt.Errorf("creating serializer: %w", err)
		}
	}

	defer serializer.Close() // this is closed at the end of the function and the error is checked.

	// TODO: We can remove the opts.RepoFunc requirement if we perform the reference resolution ahead of time.
	if err := processExisting(ctx, rootUI, opts.ExistingImages, serializer.SkipBlob, opts.RepoFunc); err != nil {
		return err
	}

	if err := resumeFrom(serializer.SkipBlob, opts.ExistingCheckpoints); err != nil {
		return err
	}

	// Add the reference name into the annotation for book keeping
	if opts.SourceDesc.Annotations == nil {
		opts.SourceDesc.Annotations = map[string]string{}
	}
	// This is similar to calling .Tag() on a CAS
	opts.SourceDesc.Annotations[ocispec.AnnotationRefName] = opts.SourceReference

	// get the deduplicated size from the annotations of the gather manifest and use it in the progress UI.
	var gatherIdxManifest ocispec.Index
	gatherManifestBytes, err := content.FetchAll(ctx, opts.SourceStorage, opts.SourceDesc)
	if err != nil {
		return fmt.Errorf("fetching the gather manifest: %w", err)
	}
	if err := json.Unmarshal(gatherManifestBytes, &gatherIdxManifest); err != nil {
		return fmt.Errorf("unmarshalling index manifest bytes: %w", err)
	}

	ddb := gatherIdxManifest.Annotations[encoding.AnnotationLayerSizeDeduplicated]
	if ddb == "" {
		// <ace-dt v1.13 serialized files will not have this annotation so we need to check it for backwards compatibility
		// we will set it to 0
		ddb = "0"
	}
	deduplicatedBytes, err := strconv.Atoi(ddb)
	if err != nil {
		return fmt.Errorf("getting deduplicated bytes from the manifest annotations: %w", err)
	}
	progress.Update(opts.SourceDesc.Size, int64(deduplicatedBytes))

	// Add caching
	// fsBlobCache := cache.NewFilesystemCache(cfg.CachePath)
	// idx = cache.ImageIndex(idx, fsBlobCache)

	/* The following is roughly equivalent to...
	crane pull --format oci some-ref oci-dir
	# then
	tar cf tape.tar -C oci-dir oci-layout index.json blobs
	*/

	// we always start with the oci-layout file (no matter what).
	if err := serializer.SaveOCILayout(); err != nil {
		return err
	}

	// write the index.json file first
	index := ocispec.Index{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		MediaType: ocispec.MediaTypeImageIndex,
		Manifests: []ocispec.Descriptor{opts.SourceDesc},
		Annotations: map[string]string{
			encoding.AnnotationGatherVersion:         dataToolVersion,
			encoding.AnnotationSerializationVersion:  fmt.Sprint(serializationVersion),
			encoding.AnnotationLayerSizeDeduplicated: ddb,
		},
	}
	if err := serializer.SaveIndex(index); err != nil {
		return err
	}

	// TODO populate this from writeDescriptor
	// The ORAS content/oci approach is to save all manifests in the index.json
	// We need to at least save all the roots for Predecessors to be discoverable.
	// Finding the subset that is the roots is someone hard so we store them all instead (like ORAS).
	// In fact a descriptor can be done many ways so we return all manifests that we come across.
	mt := newManifestTracker()

	if err := writeDescriptor(ctx, rootUI, progress, opts.Recursive, opts.SourceStorage, serializer, mt, opts.SourceDesc); err != nil {
		return fmt.Errorf("writing top level descriptor: %w", err)
	}

	// write out the index.json (again) but this time with all the newly discovered manifests
	index.Manifests = mt.Manifests()
	if err := serializer.SaveIndex(index); err != nil {
		return err
	}

	if opts.WithManifestJSON {
		mj, err := encoding.BuildManifestJSON(ctx, opts.SourceStorage, index.Manifests)
		if err != nil {
			return fmt.Errorf("building manifest.json: %w", err)
		}

		if err = serializer.SaveManifestJSON(mj); err != nil {
			return fmt.Errorf("saving manifest.json to archive: %w", err)
		}
	}

	// clean up
	if err := serializer.Close(); err != nil {
		return err
	}

	if err := dest.Close(); err != nil {
		return err
	}

	return g.Wait()
}

// resumeFrom allows resuming from a checkpoint (by knowing that some digests are already known).
func resumeFrom(haveLayerDigest func(ocispec.Descriptor), existingCheckpoints []ResumeFromLedger) error {
	// iterate through the checkpoint files
	for _, cp := range existingCheckpoints {
		// if the offset is 0 this checkpoint has nothing to offer us
		if cp.Offset == 0 {
			continue
		}

		if err := processCheckpoint(cp.Path, cp.Offset, haveLayerDigest); err != nil {
			return err
		}
	}

	return nil
}

// processCheckpoint reads in the ledger up to maxOffset record.
// digests are all reported with the callback function, haveBlobDigest.
func processCheckpoint(filename string, maxOffset int64, haveBlob func(ocispec.Descriptor)) error {
	// open the checkpoint file
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("unable to open checkpoint file: %w", err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	for {
		desc := ocispec.Descriptor{}
		err := decoder.Decode(&desc)
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return fmt.Errorf("ledger ill-formatted: %w", err)
		}

		//  extract the offset
		offset, err := strconv.ParseInt(desc.Annotations[encoding.AnnotationArchiveOffset], 10, 64)
		if err != nil {
			return fmt.Errorf("extracting offset: %w", err)
		}

		// only want digests that were serialized to the media successfully
		if maxOffset <= offset {
			break
		}

		// report that we have the digest to the "blob exists cache"
		haveBlob(desc)
	}

	return nil
}

// writeDescriptor writes the descriptor to the archive.
// within the function we determine if it is an index or image based on the media type.
func writeDescriptor(ctx context.Context,
	task *ui.Task, progress *ui.Progress,
	referrers bool,
	fetcher content.Fetcher,
	serializer *encoding.OCILayoutSerializer,
	mt *manifestTracker,
	desc ocispec.Descriptor,
) error {
	// short circuit if this descriptor has already been seen (digest, size, and media type)
	if mt.Exists(desc) {
		return nil
	}

	if !encoding.IsManifest(desc.MediaType) {
		return writeBlob(ctx, task, progress, fetcher, serializer, desc)
	}

	if err := writeManifest(ctx, task, progress, referrers, fetcher, serializer, mt, desc); err != nil {
		return err
	}

	if referrers {
		// write out predecessors if supported by fetcher
		// NOTE we do not look for predecessors of non-manifests (blobs)
		if pf, ok := fetcher.(content.PredecessorFinder); ok {
			predecessors, err := pf.Predecessors(ctx, desc)
			if err != nil {
				return fmt.Errorf("getting predecessors for node %v: %w", desc, err)
			}

			// follow predecessors down the tree
			for _, desc := range predecessors {
				if err := writeDescriptor(ctx, task, progress, referrers, fetcher, serializer, mt, desc); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// writeManifest writes an index or image manifest to the archive.
func writeManifest(ctx context.Context,
	task *ui.Task, progress *ui.Progress,
	referrers bool,
	fetcher content.Fetcher,
	serializer *encoding.OCILayoutSerializer,
	mt *manifestTracker,
	desc ocispec.Descriptor,
) error {
	task = task.SubTask("Manifest " + print.ShortDigest(desc.Digest))
	defer task.Complete()
	task.Info("Processing manifest ", desc.Digest)

	mt.Add(desc)

	// write the index itself (as a blob)
	if err := serializer.SaveBlob(ctx, fetcher, desc); err != nil {
		return fmt.Errorf("writing index as a blob: %w", err)
	}

	successors, err := encoding.Successors(ctx, fetcher, desc)
	if err != nil {
		return fmt.Errorf("finding successors of image: %w", err)
	}

	// follow successors
	for _, desc := range successors {
		if err := writeDescriptor(ctx, task, progress, referrers, fetcher, serializer, mt, desc); err != nil {
			return err
		}
	}
	return nil
}

// writeBlob writes a single blob (layer or config) to the archive.
func writeBlob(ctx context.Context,
	task *ui.Task, progress *ui.Progress,
	fetcher content.Fetcher,
	serializer *encoding.OCILayoutSerializer,
	desc ocispec.Descriptor,
) error {
	task = task.SubTask("Blob " + print.ShortDigest(desc.Digest))
	defer task.Complete()

	// update the total write size
	defer progress.Update(desc.Size, 0)
	task.Infof("Writing blob (%s) %s", print.Bytes(desc.Size), print.ShortDigest(desc.Digest))

	return serializer.SaveBlob(ctx, fetcher, desc)
}

// processExisting extracts all blobs (layers and config) from the images in existingImages and calls exists() on each.
func processExisting(ctx context.Context, rootUI *ui.Task, existingImages []string, exists func(ocispec.Descriptor), repoFunc func(context.Context, string) (*remote.Repository, error)) error {
	// log := logger.FromContext(ctx).With("existing images", existingImages)
	task := rootUI.SubTask("Existing")
	defer task.Complete()

	// TODO do this in parallel with errgroup
	// but that requires exists() to be concurrency safe
	for _, ref := range existingImages {
		srepo, err := repoFunc(ctx, ref)
		if err != nil {
			return fmt.Errorf("error parsing source ref: %w", err)
		}

		desc, err := srepo.Resolve(ctx, ref)
		if err != nil {
			return fmt.Errorf("error getting the remote descriptor: %w", err)
		}
		// Walk down the tree and extract digests for blobs (layers and config)
		if err := extractBlobs(ctx, exists, srepo, desc); err != nil {
			return err
		}
	}

	return nil
}
