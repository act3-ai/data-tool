package bottle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"git.act3-ace.com/ace/data/tool/internal/archive"
	"git.act3-ace.com/ace/data/tool/internal/bottle/label"
	"git.act3-ace.com/ace/data/tool/internal/util"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// SaveOptions is a structure for supplying options to the SaveUpdatesToSet function. By default, all options
// are "on", eg, the options to disable functions are all false.
type SaveOptions struct {
	NoArchive     bool
	NoDigest      bool
	NoCommit      bool
	CompressLevel string
}

// SaveUpdatesToSet performs archival, digest, and cache commission to bottle components, and saves bottle metadata.
func SaveUpdatesToSet(ctx context.Context, btl *Bottle, options SaveOptions) error {
	var tmpFileMap sync.Map

	log := logger.FromContext(ctx)

	if !options.NoArchive {
		log.InfoContext(ctx, "Checking if files need to be archived")
		if err := archiveParts(ctx, btl, options.CompressLevel, &tmpFileMap); err != nil {
			return err
		}
	}
	if !options.NoDigest {
		log.InfoContext(ctx, "Checking if files need to be digested")
		if err := digestParts(ctx, btl); err != nil {
			return err
		}
	}
	if !options.NoCommit {
		log.InfoContext(ctx, "Committing new files to cache")
		if err := commitParts(ctx, btl, &tmpFileMap); err != nil {
			return err
		}

		// build the latest manifest for the bottle based on any updated information generated above. We don't need the
		// manifest handler here, but note that it is saved within the bottle.
		err := btl.ConstructManifest()
		if err != nil {
			return err
		}
	}

	log.InfoContext(ctx, "Saving updated information to bottle")
	return btl.Save()
}

// PrepareUpdatedParts performs bottle part processing as a delegate when scanning for changed parts.  The bottle
// information is updated with the changed data, preserving existing data where possible.   Mostly, this involves
// removing file entries, resetting file entries (removing size/digest to trigger recalc), and adding file entries.
func PrepareUpdatedParts(ctx context.Context, btl *Bottle) Visitor {
	fsys := os.DirFS(btl.GetPath())

	// TODO why does this function not return an error, return an error
	return func(info PartInfo, status PartStatus) (bool, error) {
		name := info.GetName()
		log := logger.FromContext(ctx).With("name", name)

		switch status {
		case StatusDeleted:
			log.InfoContext(ctx, "Bottle part removed")
			btl.RemovePartMetadata(name)
		case StatusChanged:
			log.InfoContext(ctx, "Changed part flagged for reprocessing")
			// the only information we know for certain is the (immutable) part name
			// and the content size.
			// TODO: Can we override the mediatype as to determine the remaining
			// information when we reprocess the part and to ensure we try to compress?
			// TODO: add plumbing for compression type (zstd or gzip)
			// mt := mediatype.MediaTypeLayerZstd
			// if mediatype.IsArchived(info.GetMediaType()) {
			// 	mt = mediatype.MediaTypeLayerTarZstd
			// }
			modTime := info.GetModTime()
			btl.UpdatePartMetadata(name,
				info.GetContentSize(), "",
				nil, // preserve part labels
				0, "",
				"",
				&modTime,
			)
		case StatusNew:
			log.InfoContext(ctx, "New part flagged for processing")
			fullPath := btl.NativePath(name)
			fInfo, err := os.Stat(fullPath)
			if err != nil {
				log.InfoContext(ctx, "file discovered during push, but not able to stat")
				return false, err
			}

			if strings.HasSuffix(name, "/") {
				if err := addDirToBottle(ctx, fsys, btl, name, fInfo); err != nil {
					return false, fmt.Errorf("unable to add directory to bottle at %s: %w", fullPath, err)
				}
			} else {
				if err := addFileToBottle(ctx, btl, name, fInfo); err != nil {
					return false, fmt.Errorf("unable to add file to bottle %s: %w", fullPath, err)
				}
			}
		case StatusCached, StatusDigestMatch, StatusExists, StatusVirtual:
			return false, nil
		}
		return false, nil
	}
}

// CopyFromCache copies a bottle part from the cache, handling extraction and decompression
// based on part mediatypes.
func CopyFromCache(ctx context.Context, btl *Bottle, desc ocispec.Descriptor, name string, btlPartMutex *sync.Mutex) (bool, error) {
	exists, err := btl.cache.Exists(ctx, desc)
	switch {
	case err != nil:
		return false, fmt.Errorf("checking part existence in cache: %w", err)
	case !exists:
		return false, nil
	default:
		if err := handlePartMedia(ctx, btl.localPath, btl.cache, desc, name); err != nil {
			return false, fmt.Errorf("copying part from cache: %w", err)
		}

		// update part modification time, ensuring we base future evaluations of mod time are
		// based off of the actual bottle part not the originally cached version
		fi, err := os.Stat(btl.NativePath(name))
		if err != nil {
			return true, fmt.Errorf("determining part modification time: %w", err)
		}

		btlPartMutex.Lock()
		btl.partByName(name).Modified = fi.ModTime()
		btlPartMutex.Unlock()

		return true, nil
	}
}

// ErrUnknownLayerMediaType is the error if the layer media type is unknown.
var ErrUnknownLayerMediaType = errors.New("unknown layer media type")

// handlePartMedia copies the layer into the part file/directory given by partName.
func handlePartMedia(ctx context.Context, localPath string, fetcher content.Fetcher, desc ocispec.Descriptor, partName string) error {

	destPath := filepath.Join(localPath, filepath.FromSlash(partName))

	rc, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("fetching from cache: %w", err)
	}
	defer rc.Close()

	switch desc.MediaType {
	case mediatype.MediaTypeLayerTar:
		return archive.ExtractTar(ctx, rc, destPath)
	case mediatype.MediaTypeLayerTarOld, mediatype.MediaTypeLayerTarLegacy:
		return archive.ExtractTarCompat(ctx, rc, localPath)
	case mediatype.MediaTypeLayerTarZstd:
		return archive.ExtractTarZstd(ctx, rc, destPath)
	case mediatype.MediaTypeLayerTarZstdOld, mediatype.MediaTypeLayerTarZstdLegacy:
		return archive.ExtractTarZstdCompat(ctx, rc, localPath)
	case mediatype.MediaTypeLayerZstd:
		return archive.ExtractZstd(ctx, rc, destPath)
	case mediatype.MediaTypeLayerTarGzip, mediatype.MediaTypeLayerTarGzipLegacy:
		return errors.New("gzip is not implemented")
	case mediatype.MediaTypeLayer, mediatype.MediaTypeLayerRawOld, mediatype.MediaTypeLayerRawLegacy:
		if err := os.MkdirAll(filepath.Dir(destPath), 0777); err != nil {
			return fmt.Errorf("initializing part parent directories: %w", err)
		}
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("opening destination file: %w", err)
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, rc)
		if err != nil {
			return fmt.Errorf("copying part: %w", err)
		}

		if err := rc.Close(); err != nil {
			return fmt.Errorf("closing part source: %w", err)
		}
		return nil
	default:
		return ErrUnknownLayerMediaType
	}
}

// addFileToBottle adds a single file to the provided bottle, used during init as
// part of the file processing delegate, and push when a new file is identified
// path is a bottle relative path to the part.
func addFileToBottle(ctx context.Context, btl *Bottle, pth string, info fs.FileInfo) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "found file", "path", pth)
	if info.IsDir() {
		panic("expected a file")
	}

	// ignore hidden files
	if strings.HasPrefix(info.Name(), ".") {
		log.DebugContext(ctx, "skipping hidden file.")
		return nil
	}

	log.InfoContext(ctx, "Adding metadata to data bottle")
	btl.AddPartMetadata(pth,
		info.Size(), "", // Digest is empty since we compute it lazily
		info.Size(), "",
		mediatype.MediaTypeLayerZstd,
		info.ModTime(),
	)
	log.InfoContext(ctx, "done with file", "path", pth)
	return nil
}

// addDirToBottle adds a directory to a bottle for processing, used in init during directory processing delegate, and
// during push when a new directory is discovered
// path is a bottle relative path to the part.
func addDirToBottle(ctx context.Context, fsys fs.FS, btl *Bottle, pth string, info fs.FileInfo) error {
	log := logger.FromContext(ctx).With("path", pth)
	log.InfoContext(ctx, "found directory")
	if !info.IsDir() {
		panic("expected a directory")
	}

	pth = path.Clean(pth)

	// ignore hidden dirs
	if strings.HasPrefix(info.Name(), ".") {
		log.DebugContext(ctx, "skipping hidden directory")
		return nil
	}

	subparts, err := label.HasSubparts(fsys, pth)
	if err != nil {
		return err
	}
	if subparts {
		return nil
	}

	latestUpdate, err := util.GetDirLastUpdate(fsys)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "Adding archive metadata to data bottle")
	btl.AddPartMetadata(pth+"/",
		0, "",
		0, "",
		mediatype.MediaTypeLayerTarZstd,
		latestUpdate,
	)
	return nil
}
