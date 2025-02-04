package bottle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/tool/internal/archive"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/data/tool/internal/util"
)

// archiveConcurrency is the concurrency used when archiving parts.
const archiveConcurrency = 5

// compressionRatioCheckSize is the count of bytes used for comparing compression ratios.  During compression for each
// object, up to twice this value is buffered in memory before the compression ratio check is done.
const compressionRatioCheckSize = 100000

// archivePipeline is a container for pipestream components that form a stream pipeline
// for calculating counts and digests during archival.  The uncompressed and final sizes
// and digests are made available.
type archivePipeline struct {
	LayerSize     int64
	LayerDigest   digest.Digest
	ContentSize   int64
	ContentDigest digest.Digest

	layerCount   *archive.PipeCounter
	layerDig     *archive.PipeDigest
	contentCount *archive.PipeCounter
	contentDig   *archive.PipeDigest
	compCheck    *archive.PipeCompressThreshold
}

// buildPipeline creates an archive pipeline with the destination output path.  The
// output path is created, or overwritten.  The returned PipeWriter should be provided
// to the archiver as its write stream
// The pipeline performs non-buffered simultaneous calculations while the output file
// is being written:
// archiver -> digest (unc) -> counter (unc) -> compressor ->  digest (comp) -> counter (comp)
// After writing, the counters and digesters can be accessed to retrieve statistics.
func (ap *archivePipeline) buildPipeline(ctx context.Context, progress *ui.Progress, outpath string, mediaType string, doCompress bool, compressionLevel string) (io.WriteCloser, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Creating output pipeline starting from output file")

	var nextComp io.WriteCloser
	nextComp, err := archive.NewPipeFileCreator(outpath)
	if err != nil {
		return nil, err
	}

	// Build pipeline in reverse order, starting with counting the final output bytes
	log.InfoContext(ctx, "Creating layer size counter", "file", outpath)
	ap.layerCount = archive.NewPipeCounter()
	log.InfoContext(ctx, "Adding progress tracker", "file", outpath)
	ap.layerCount.AddProgressTracker(func(complete int64) { progress.Update(complete, 0) })
	nextComp = ap.layerCount.ConnectOut(nextComp)

	// Final digest
	log.InfoContext(ctx, "Creating layer digest calculator")
	ap.layerDig = archive.NewPipeDigest()
	nextComp = ap.layerDig.ConnectOut(nextComp)

	if mediatype.IsCompressed(mediaType) && doCompress {
		// Compressor (not needed for statistics, so not retained)
		// TODO we should parse the media type with mime.ParseMediaType()
		if strings.HasSuffix(mediaType, "zstd") {
			log.InfoContext(ctx, "Selected zstd compressor")

			// Detect level of desired compression
			lv := archive.AssignCompressionLevel(compressionLevel)
			ap.compCheck = archive.NewPipeCompressThreshold(lv, compressionRatioCheckSize, false)
			nextComp = ap.compCheck.ConnectOut(nextComp)
		} else {
			log.InfoContext(ctx, "Selected gz compressor")
			// gzip compressor does not support configurable compression levels, we use around 1 MB for compression t
			ap.compCheck = archive.NewPipeCompressThreshold(0, compressionRatioCheckSize, true)
			nextComp = ap.compCheck.ConnectOut(nextComp)
		}

		// Uncompressed size
		log.InfoContext(ctx, "Creating content size counter")
		ap.contentCount = archive.NewPipeCounter()
		nextComp = ap.contentCount.ConnectOut(nextComp)

		// Uncompressed digest
		log.InfoContext(ctx, "Creating content digest calculator", "outpath", outpath)
		ap.contentDig = archive.NewPipeDigest()
		nextComp = ap.contentDig.ConnectOut(nextComp)
	} else {
		ap.contentCount = ap.layerCount
		ap.contentDig = ap.layerDig
	}

	// Return the input side of the pipeline
	return nextComp, nil
}

// checkCompressionRatio examines the compCheck compression threshold pipe writer, and returns false if the ratio
// didn't meet the desired threshold for compression defined by the CompressionRatioThreshold constant in archive.
//
//nolint:sloglint
func (ap *archivePipeline) checkCompressionRatio(log *slog.Logger) bool {
	logger.V(log, 1).Info("Checking compression ratio of file")
	if ap.compCheck.IsCompressible() {
		logger.V(log, 1).Info("Effective compression", "compression ratio", ap.compCheck.FinalRatio)
	} else {
		logger.V(log, 1).Info("Ineffective compression", "compression ratio", ap.compCheck.FinalRatio)
		return false
	}
	return true
}

// finalize updates the final values.
func (ap *archivePipeline) finalize() {
	if ap.layerCount != nil {
		ap.LayerSize = ap.layerCount.Count
	}
	if ap.layerDig != nil {
		ap.LayerDigest = ap.layerDig.GetDigest()
	}
	if ap.contentCount != nil {
		ap.ContentSize = ap.contentCount.Count
	} else {
		ap.ContentSize = ap.LayerSize
	}
	// pre compression/encryption digest exists only when those operations
	// are performed, otherwise, the final digest matches the content digest
	if ap.contentDig != nil {
		ap.ContentDigest = ap.contentDig.GetDigest()
	} else {
		ap.ContentDigest = ap.LayerDigest
	}
}

// makeArchivePath generates an output file path based on the local path and file entry information.
func makeArchivePath(scratchPath string) (string, error) {
	// TODO delete this function.  We should be using os.CreateTemp directly and the returned os.File (not it's path).
	tf, err := os.CreateTemp(scratchPath, "*")
	if err != nil {
		return "", fmt.Errorf("unable to create archive file: %w", err)
	}

	archFile := tf.Name()
	// TODO this is an anti-pattern.  We should not create a temp file and then close it.  We should be returning the File object instead of just the path.
	err = tf.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close archive file: %w", err)
	}

	return archFile, nil
}

// archiveParts examines the file entries in a bottle and archives them
// if necessary.  Items that have an archive/compression format specified
// are archived if the matching archive file does not already exist, and only
// files that do not have digests that appear in the cache are archived.
func archiveParts(ctx context.Context, btl *Bottle, compressionLevel string, tmpFileMap *sync.Map) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Archiving files/directories in bottle")

	// protects btl.Parts
	var btlPartMutex sync.Mutex

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.SetLimit(archiveConcurrency)

	progress := ui.FromContextOrNoop(ctx).SubTaskWithProgress("Archiving Parts")
	defer progress.Complete()

	var total int64
	for i, part := range btl.GetParts() {
		if part.GetLabels()["data.act3-ace.io/compression"] == "none" ||
			(btl.VirtualPartTracker != nil && btl.VirtualPartTracker.HasContent(part.GetContentDigest())) {
			// numDone++
			progress.Infof("%s completed", part.GetName())
			continue
		}
		total += part.GetContentSize()

		// Start a goroutine for each part. Compressing and archiving if necessary
		errGroup.Go(func() error {
			return archivePart(ctx, &btlPartMutex, progress, &btl.Parts[i], btl, compressionLevel, tmpFileMap)
		})
	}
	progress.Update(0, total)

	// check for any errors from the goroutines
	return errGroup.Wait()
}

func archivePart(ctx context.Context, btlPartMutex sync.Locker, progress *ui.Progress, part PartInfo, btl *Bottle, compressionLevel string, tmpFileMap *sync.Map) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Active file", "filename", part.GetName())

	defer progress.Infof("%v completed", part.GetName())

	// Skip if the part has a digest (has been archived and digested before), AND the digest matches the cache
	if part.GetLayerDigest() != "" {
		exists, err := btl.cache.Exists(ctx, ocispec.Descriptor{Digest: part.GetLayerDigest()})
		if err != nil {
			logger.V(log, 1).ErrorContext(ctx, "checking for part in cache", "error", err)
		}
		if exists {
			logger.V(log, 1).InfoContext(ctx, "Skipping archive because it was found in cache", "filename", part.GetName(), "digest", part.GetLayerDigest())
			return nil
		}
	}

	// Skip if the part is already archived / compressed or marked oci RAW
	mt := part.GetMediaType()
	if !(mediatype.IsArchived(mt) || mediatype.IsCompressed(mt)) || mediatype.IsRaw(mt) {
		logger.V(log, 1).InfoContext(ctx, "Skipping archive because of format", "mediaType", mt)
		return nil
	}

	// Check source file or directory
	srcFile := btl.NativePath(part.GetName())
	log.InfoContext(ctx, "Processing part", "path", srcFile)

	// Use temporary file for archive output, this can be committed to the cache later
	archFile, err := makeArchivePath(btl.ScratchPath())
	if err != nil {
		return err
	}
	tmpFileMap.Store(part.GetName(), archFile)

	// Create output pipeline; which allows calculation of sizes, digests and compression ratio during stream
	archpipe := archivePipeline{}

	output, err := archpipe.buildPipeline(ctx, progress, archFile, part.GetMediaType(), true, compressionLevel)
	if err != nil {
		return err
	}

	// Archive to the pipeline or copy to the pipeline if not an archive format
	if mediatype.IsArchived(mt) {
		log.InfoContext(ctx, "Desired format is archived", "mediaType", mt)
		err = archive.TarToStream(ctx, os.DirFS(srcFile), output)
		// Set uncompressed size to archive size
		archpipe.ContentSize = archpipe.contentCount.Count
	} else {
		log.InfoContext(ctx, "Desired format is single file", "mediaType", mt)
		archpipe.ContentSize, err = util.CopyToStream(srcFile, output)
	}
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}

	if err := output.Close(); err != nil {
		return err
	}

	if !archpipe.checkCompressionRatio(log.With("part name", part.GetName())) {
		// Compression was too inefficient for this file/directory, so the output is just the original data, either
		// archived or plain.  We need to update the media type accordingly
		if mediatype.IsArchived(part.GetMediaType()) {
			log.InfoContext(ctx, "Creating archive only for path")
			// An archive was created (eg, for directories), but without compression.  Change the format to archive only
			mt = mediatype.MediaTypeLayerTar
		} else {
			log.InfoContext(ctx, "Using raw source file")
			// File output is just the raw file data, without compression. Change the format to plain layer
			mt = mediatype.MediaTypeLayer
		}
	}

	// Update bottle information with statistics from archive pipeline
	archpipe.finalize()
	btlPartMutex.Lock()
	btl.UpdatePartMetadata(part.GetName(),
		archpipe.ContentSize,
		archpipe.ContentDigest,
		nil,
		archpipe.LayerSize,
		archpipe.LayerDigest,
		mt,
		nil,
	)
	btlPartMutex.Unlock()

	log.InfoContext(ctx, "Archive file created", "path", archFile)
	return nil
}

// digestParts calculates digest values for any files in the bottle, and records them in the
// bottle structure.  The process is skipped if a digest is already present in the bottle. For
// most cases, the digest is calculated as part of the archival process using a stream pipeline.
func digestParts(ctx context.Context, btl *Bottle) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Digesting files/directories in bottle")
	defer log.InfoContext(ctx, "Digesting completed")

	// protects btl.Parts
	var btlPartMutex sync.Mutex

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.SetLimit(5)

	progress := ui.FromContextOrNoop(ctx).SubTaskWithProgress("Digesting Parts")
	defer progress.Complete()

	numParts := btl.NumParts()
	for i := 0; i < numParts; i++ {
		progress.Update(0, btl.Parts[i].GetContentSize())

		// Start a goroutine for each part
		errGroup.Go(func() error {
			return digestPart(ctx, &btlPartMutex, progress, &btl.Parts[i], btl)
		})
	}

	// check for any errors from the goroutines
	return errGroup.Wait()
}

func digestPart(ctx context.Context, btlPartMutex sync.Locker, p *ui.Progress, part PartInfo, btl *Bottle) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Checking if digest for file already exists", "filename", part.GetName())

	defer p.Infof("%v completed", part.GetName())

	if part.GetLayerDigest() != "" && part.GetContentDigest() != "" {
		logger.V(log, 1).InfoContext(ctx, "Digest already exists for file, skipping", "filename", part.GetName())
		return nil
	}
	// Get the file name if not an archive, or the archive file name.
	fname := btl.NativePath(part.GetName())
	log.InfoContext(ctx, "Raw file path", "path", fname)
	mt := part.GetMediaType()
	if mediatype.IsArchived(mt) || mediatype.IsCompressed(mt) {
		log.InfoContext(ctx, "appending archive/compression media type extension format", "mediaType", mt)
		log.InfoContext(ctx, "archive format", "name", part.GetName(),
			"layerDigest", part.GetLayerDigest(), "fileDigest", part.GetContentDigest())
		fname = filepath.Join(btl.ScratchPath(), part.GetName())
	}

	log.InfoContext(ctx, "Checking if file exists", "path", fname)
	// Check if file exists and can be accessed
	finfo, err := os.Stat(fname)
	if err != nil {
		return fmt.Errorf("bottle file %s unavailable for digest calculation: %w", fname, err)
	}

	log.InfoContext(ctx, "Checking if path is a directory", "path", fname)
	// Sanity check to ensure a directory didn't sneak through
	if finfo.IsDir() {
		panic(fmt.Sprintf("expected part file %s should not be a directory", fname))
	}

	log.InfoContext(ctx, "Creating metadata generator for digest calculation")
	filedigest, err := digestFileWithProgress(digest.SHA256, fname, p)
	if err != nil {
		return fmt.Errorf("unable to calculate digest for %s: %w", fname, err)
	}

	btlPartMutex.Lock()
	btl.UpdatePartMetadata(part.GetName(),
		finfo.Size(),
		filedigest,
		nil,
		finfo.Size(),
		filedigest,
		"",
		nil,
	)
	btlPartMutex.Unlock()

	log.InfoContext(ctx, "File digest calculated")
	return nil
}

// commitParts commits newly added/archived files to the cache.
func commitParts(ctx context.Context, btl *Bottle, tmpFileMap *sync.Map) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Committing new files/directories in bottle to cache")
	for _, part := range btl.GetParts() {
		// virtual parts cannot be committed, since they do not exist locally
		if btl.VirtualPartTracker != nil && btl.VirtualPartTracker.HasContent(part.GetContentDigest()) {
			continue
		}

		if err := commitPart(ctx, btl, part, tmpFileMap); err != nil {
			return fmt.Errorf("committing part: %w", err)
		}
	}
	return nil
}

func commitPart(ctx context.Context, btl *Bottle, part PartInfo, tmpFileMap *sync.Map) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "File about to be committed", "filename", part.GetName())
	if part.GetLayerDigest() == "" {
		return fmt.Errorf("digest for file %s not found in bottle", part.GetName())
	}

	// Get the file name if not an archive, or the archive file name.
	fname := btl.NativePath(part.GetName())
	mt := part.GetMediaType()
	var tempExists = false
	archivedOrCompressed := mediatype.IsArchived(mt) || mediatype.IsCompressed(mt)
	// any object within the tmpFileMap will need to be moved to the final cache destination.
	if n, ok := tmpFileMap.Load(part.GetName()); ok {
		fname = n.(string)
		tempExists = true
	} else if archivedOrCompressed {
		// TODO: this object doesn't appear in the tmpFileMap, but is archived/compressed, can this state occur?
		fname = filepath.Join(btl.ScratchPath(), part.GetName())
	}
	// determine if a temp archive file exists that will need to be moved to cache
	// this is mainly for trace messaging
	if _, err := os.Stat(fname); !errors.Is(err, fs.ErrNotExist) && archivedOrCompressed {
		tempExists = true
		log.InfoContext(ctx, "Temporary archive exist", "path", fname)
	}

	exists, err := btl.cache.Exists(ctx, ocispec.Descriptor{Digest: part.GetLayerDigest()})
	if err != nil {
		logger.V(log, 1).ErrorContext(ctx, "checking for part in cache", "error", err)
	}
	if exists {
		logger.V(log, 1).InfoContext(ctx, "File found in cache", "filename", part.GetName(), "digest", part.GetLayerDigest())
		if tempExists && archivedOrCompressed {
			logger.V(log, 1).InfoContext(ctx, "Removing temporarily created archive")
			if err := os.Remove(fname); err != nil {
				return fmt.Errorf("failed to remove archive: %w", err)
			}
		}
		return nil
	}

	log.InfoContext(ctx, "adding part file to cache")
	f, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("opening part file: %w", err)
	}
	defer f.Close()

	if err := btl.cache.Push(ctx, ocispec.Descriptor{Digest: part.GetLayerDigest(), Size: part.GetLayerSize()}, f); err != nil {
		return fmt.Errorf("pushing part file to cache: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing part file: %w", err)
	}

	if tempExists && archivedOrCompressed {
		logger.V(log, 1).InfoContext(ctx, "Removing temporarily created archive")
		if err := os.Remove(fname); err != nil {
			return fmt.Errorf("failed to remove archive: %w", err)
		}
	}

	return nil
}

func digestFileWithProgress(alg digest.Algorithm, fname string, p *ui.Progress) (digest.Digest, error) {
	f, err := os.Open(fname)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	digester := alg.Digester()
	if _, err := io.Copy(io.MultiWriter(digester.Hash(), p), f); err != nil {
		return "", fmt.Errorf("unable to compute digest: %w", err)
	}

	return digester.Digest(), nil
}
