package bottle

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"git.act3-ace.com/ace/data/tool/internal/bottle/label"
	"git.act3-ace.com/ace/data/tool/internal/storage"
	"git.act3-ace.com/ace/go-common/pkg/logger"
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
	return func(info storage.PartInfo, status PartStatus) (bool, error) {
		name := info.GetName()
		log := logger.FromContext(ctx).With("name", name)

		switch status {
		case StatusDeleted:
			log.InfoContext(ctx, "Bottle part removed")
			btl.RemovePartMetadata(name)
		case StatusChanged:
			log.InfoContext(ctx, "Changed part flagged for reprocessing")
			// TODO why do we not reset the content size?
			/// Seems like we need to be setting content size somewhere else (where we update everything else), not preserving it here.
			btl.UpdatePartMetadata(name,
				info.GetContentSize(), "",
				nil, // preserve part labels
				info.GetLayerSize(), "",
				"",
				&time.Time{},
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

	log.InfoContext(ctx, "Adding archive metadata to data bottle")
	btl.AddPartMetadata(pth+"/",
		0, "",
		0, "",
		mediatype.MediaTypeLayerTarZstd,
		time.Now(), // TODO why is this not the newest mod time of the files in this archive?
	)
	return nil
}
