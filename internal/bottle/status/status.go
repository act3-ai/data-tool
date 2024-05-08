// Package status handles inspecting the filesystem for changes
package status

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/tool/internal/archive"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle/label"
	"gitlab.com/act3-ai/asce/data/tool/internal/storage"
	"gitlab.com/act3-ai/asce/data/tool/internal/util"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// partStatuses is a collection of file entry lists that organize them into various
// status lists for display.
type partStatuses struct {
	New        []storage.PartInfo
	Cached     []storage.PartInfo
	Changed    []storage.PartInfo
	Deleted    []storage.PartInfo
	DirDetails map[string][]string
}

// Visitor is a function delegate for performing an action based on a provided file info and status indicator
// return true to halt processing, or false to continue.
type Visitor func(storage.PartInfo, bottle.PartStatus) (bool, error)

// AddEntry processes a file status and adds a file entry to the appropriate status collection.
func (fss *partStatuses) AddEntry(e storage.PartInfo, s bottle.PartStatus, dirpaths []string) {
	if s&bottle.StatusExists != bottle.StatusExists {
		fss.New = append(fss.New, e)
	} else {
		// Existing items must be removed from the "Deleted" list
		name := e.GetName()
		for i, f := range fss.Deleted {
			if f.GetName() != name {
				continue
			}
			// found in list, copy end to found index and truncate slice
			lp := len(fss.Deleted) - 1
			fss.Deleted[i] = fss.Deleted[lp]
			fss.Deleted[lp] = nil
			fss.Deleted = fss.Deleted[:lp]
			// name must be unique so it's safe to exit early
			break
		}
	}
	if s&(bottle.StatusChanged|bottle.StatusDigestMatch) == bottle.StatusChanged {
		fss.Changed = append(fss.Changed, e)
	} else if s&bottle.StatusCached == bottle.StatusCached {
		fss.Cached = append(fss.Cached, e)
	}

	if len(dirpaths) != 0 {
		fss.DirDetails[e.GetName()] = dirpaths
	}
}

// Display outputs a set of file statuses.
func (fss *partStatuses) Display(out *strings.Builder) {
	for _, f := range fss.Cached {
		_, err := out.WriteString(fmt.Sprintf("File cached with name: %v\n", f.GetName()))
		if err != nil {
			// expect string builder to be properly passed, if err: panic
			panic(err)
		}
	}

	for _, f := range fss.Changed {
		_, err := out.WriteString(fmt.Sprintf("File changed with name: %v\n", f.GetName()))
		if err != nil {
			// expect string builder to be properly passed, if err: panic
			panic(err)
		}
		if l, ok := fss.DirDetails[f.GetName()]; ok {
			for _, p := range l {
				_, err := out.WriteString(fmt.Sprintf("directory details: %v\n", p))
				if err != nil {
					// expect string builder to be properly passed, if err: panic
					panic(err)
				}
			}
		}
	}

	for _, f := range fss.New {
		_, err := out.WriteString(fmt.Sprintf("New file with name: %v\n", f.GetName()))
		if err != nil {
			// expect string builder to be properly passed, if err: panic
			panic(err)
		}
	}

	for _, f := range fss.Deleted {
		_, err := out.WriteString(fmt.Sprintf("File deleted with name: %v\n", f.GetName()))
		if err != nil {
			// expect string builder to be properly passed, if err: panic
			panic(err)
		}
	}
}

// VisitAll calls a status visitor for each file info in the status structure.
func (fss *partStatuses) VisitAll(visitor Visitor) error {
	for _, v := range fss.Cached {
		stop, err := visitor(v, bottle.StatusCached)
		if err != nil {
			return fmt.Errorf("failed visiting cached: %w", err)
		}
		if stop {
			break
		}
	}
	for _, v := range fss.Changed {
		stop, err := visitor(v, bottle.StatusChanged)
		if err != nil {
			return fmt.Errorf("failed visiting changed: %w", err)
		}
		if stop {
			break
		}
	}
	for _, v := range fss.Deleted {
		stop, err := visitor(v, bottle.StatusDeleted)
		if err != nil {
			return fmt.Errorf("failed visiting deleted: %w", err)
		}
		if stop {
			break
		}
	}
	for _, v := range fss.New {
		stop, err := visitor(v, bottle.StatusNew)
		if err != nil {
			return fmt.Errorf("failed visiting new: %w", err)
		}
		if stop {
			break
		}
	}
	return nil
}

func (fss *partStatuses) ChangesDetected() bool {
	if len(fss.Changed) != 0 || len(fss.New) != 0 || len(fss.Deleted) != 0 {
		return true
	}
	return false
}

// getDirArchiveSize gets the uncompressed size of a directory by archiving it, without
// compression. The archived data is ignored.
func getDirArchiveSize(ctx context.Context, btl *bottle.Bottle, path string) (int64, error) {
	counter := archive.NewPipeCounter()
	counter.ConnectOut(&archive.PipeTerm{})
	defer counter.Close()
	err := archive.TarToStream(ctx, os.DirFS(path), counter)
	if err != nil {
		return 0, err
	}
	err = counter.Close()
	return counter.Count, err
}

// Options define a set of options for processing file statuses.
type Options struct {
	// WantDetails enables directory tree walking for gathering files changed within directories
	WantDetails bool

	// Visitor is a delegate for processing individual file infos with status indicators, set to null to perform a default Display action
	Visitor Visitor
}

// statusFileVisitor is a callback function for processing files discovered in
// a data bottle path (hidden files are skipped.)  If a file is found in
// the bottle, its last modified time is checked against the record in
// the bottle.
// path is the relative file path within the bottle.
func statusFileVisitor(ctx context.Context, btl *bottle.Bottle, fss *partStatuses, opts Options, path string, info fs.FileInfo) error {
	log := logger.FromContext(ctx)
	logger.V(log, 1).InfoContext(ctx, "found file", "path", path)
	if info.IsDir() {
		panic("expected a file")
	}

	// ignore hidden files
	if strings.HasPrefix(info.Name(), ".") {
		logger.V(log, 2).InfoContext(ctx, "skipping hidden file.")
		return nil
	}

	entry := bottle.PartTrack{
		Part: latest.Part{
			Name: path,
			Size: info.Size(),
		},
		MediaType: mediatype.MediaTypeLayerZstd,
		Modified:  info.ModTime(),
	}

	// We may want to support zero length files for  marker files, etc.  But for now, the zero length fails schema
	// validation so here we throw an error to inform the user
	if entry.Size == 0 {
		return fmt.Errorf("zero length part file detected, which is not currently supported: %s", path)
	}

	status := btl.GetPartStatus(entry)

	// HACK - Ultimately, UpdatePartMetadata will overwrite the layer size. Adding it here
	// even if archival and digesting is necessary is safe as it will just be overwritten.
	if (status & bottle.StatusExists) == bottle.StatusExists {
		// validating exists guarantees that btl.GetPartByName does not return nil
		pInfo := btl.GetPartByName(entry.Name)
		entry.LayerSize = pInfo.GetLayerSize()
	}

	fss.AddEntry(&entry, status, []string{})

	return nil
}

// statusDirVisitor is a callback function for processing directories discovered
// in a data bottle path (hidden dirs are skipped).  If a directory is found in
// the bottle, it is further walked to determine the latest update time among
// all files, to determine if a change has occurred since the last time recorded
// for the bottle entry.
func statusDirVisitor(ctx context.Context, btl *bottle.Bottle, fS *partStatuses, opts Options, path string, info fs.FileInfo) error {
	log := logger.FromContext(ctx)
	logger.V(log, 1).InfoContext(ctx, "found directory", "path", path)
	if !info.IsDir() {
		panic("expected a directory")
	}

	// ignore hidden dirs
	if strings.HasPrefix(info.Name(), ".") {
		logger.V(log, 2).InfoContext(ctx, "skipping hidden dir")
		return fs.SkipDir
	}

	fsys := os.DirFS(btl.GetPath())
	subparts, err := label.HasSubparts(fsys, path)
	if err != nil {
		return err
	}
	if subparts {
		// this directory is not a part because it has subparts
		return nil
	}

	dirPath := btl.NativePath(path)
	nativeFsys := os.DirFS(dirPath)
	logger.V(log, 1).InfoContext(ctx, "Walking directory for latest modification time")
	latestUpdate, err := util.GetDirLastUpdate(nativeFsys)
	if err != nil {
		return err
	}
	// TODO this is inefficient.  We walk the directory above for find the latest modtime and then we walk it again to find the archived size.  This could be one traversal of the directory tree.
	// TODO this should also use fs.FS since it is read only
	archSize, err := getDirArchiveSize(ctx, btl, dirPath)
	if err != nil {
		return err
	}
	entry := bottle.PartTrack{
		Part: latest.Part{
			Name: path + "/",
			Size: archSize,
		},
		MediaType: mediatype.MediaTypeLayerTarZstd,
		Modified:  latestUpdate,
	}

	status := btl.GetPartStatus(entry)
	var dirpaths []string

	if status&bottle.StatusCached == bottle.StatusCached && opts.WantDetails {
		// details causes a secondary walk of a subdirectory tree to determine which
		// files have been updated, and is thus a bit of a performance hit.  This only
		// checks directories that seem to have changed
		logger.V(log, 1).InfoContext(ctx, "Walking directory again to determine what paths have changed")
		var testTime time.Time
		existing := btl.GetPartByName(entry.Name)
		if existing != nil {
			testTime = existing.GetModTime()
		}
		dirpaths, err = util.GetDirUpdatedPaths(nativeFsys, testTime)
		if err != nil {
			return err
		}
	}

	// HACK - Ultimately, UpdatePartMetadata will overwrite the layer size. Adding it here
	// even if archival and digesting is necessary is safe as it will just be overwritten.
	if (status & bottle.StatusExists) == bottle.StatusExists {
		// validating exists guarantees that btl.GetPartByName does not return nil
		pInfo := btl.GetPartByName(entry.Name)
		entry.LayerSize = pInfo.GetLayerSize()
	}

	fS.AddEntry(&entry, status, dirpaths)
	// want to skip the directory if we've processed it, so the iteration doesn't recurse into it
	return fs.SkipDir
}

// InspectBottleFiles files iterates over all the files in a directory, including
// subdirectories, and determines if the files are cached, new, or updated.
// Returns a string containing bottle info if opts specifies display.
// Also returns a boolean that is true if files changed.
func InspectBottleFiles(ctx context.Context, btl *bottle.Bottle, opts Options) (string, bool, error) {
	log := logger.FromContext(ctx)
	logger.V(log, 1).InfoContext(ctx, "Scanning for files and directories")
	bottlePath := btl.GetPath()
	pS := &partStatuses{
		DirDetails: make(map[string][]string),
		Deleted:    btl.GetParts(),
	}

	// TODO consider passing in the fs.FS object for easier testability
	fsys := os.DirFS(bottlePath)

	logger.V(log, 1).InfoContext(ctx, "Walking files")
	err := fs.WalkDir(fsys, ".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		// resolve the path to the final file/dir (resolves symbolic links)
		info, err := fs.Stat(fsys, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// directory
			err = statusDirVisitor(ctx, btl, pS, opts, path, info)
			if err != nil {
				if errors.Is(err, fs.SkipDir) && !d.IsDir() {
					// simlink directories aren't recursed, but are still processed via status dir visitor, which
					// returns skip dir after processing directory contents.  The combination causes directory walking
					// to terminate early in walkdir err handling, so, we force it to continue here.
					return nil
				}
				return err
			}
			return nil
		}
		// file
		return statusFileVisitor(ctx, btl, pS, opts, path, info)
	})
	if err != nil {
		return "", false, err
	}
	if btl.VirtualPartTracker != nil {
		// build a cache for fast access by content ID/digest
		contentIDMap := make(map[digest.Digest]storage.PartInfo, len(btl.VirtualPartTracker.VirtRecords))
		for _, part := range btl.GetParts() {
			contentIDMap[part.GetContentDigest()] = part
		}
		for _, p := range btl.VirtualPartTracker.VirtRecords {
			part := contentIDMap[p.ContentID]
			pS.AddEntry(part, bottle.StatusVirtual|bottle.StatusExists, []string{})
		}
	}
	var displayStr strings.Builder
	if opts.Visitor == nil {
		logger.V(log, 1).InfoContext(ctx, "Displaying results")
		pS.Display(&displayStr)
	} else {
		if err := pS.VisitAll(opts.Visitor); err != nil {
			return "", false, err
		}
		if pS.ChangesDetected() {
			return "", true, nil
		}
	}
	return displayStr.String(), false, nil
}
