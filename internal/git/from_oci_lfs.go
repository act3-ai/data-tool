package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// runLFS updates a remote git repository by fetching git LFS tracked files from an LFS manifest
// and pushes the changes to the remote git reference.
//
// runLFS should only be called after Run.
func (f *FromOCI) runLFS(ctx context.Context, refList []string, commitManifest ocispec.Descriptor, remoteLFSFiles []string) error {
	log := logger.FromContext(ctx)

	if err := f.cmdHelper.ConfigureLFS(); err != nil {
		return fmt.Errorf("configuring LFS: %w", err)
	}

	log.InfoContext(ctx, "Fetching LFS manifest and config")
	lfsManDesc, err := f.FetchLFSManifestConfig(ctx, commitManifest, false)
	switch {
	case errors.Is(err, errdef.ErrNotFound):
		return fmt.Errorf("LFS manifest does not exist: %w", err)
	case err != nil:
		return fmt.Errorf("fetching existing LFS manifest: %w", err)
	case len(f.lfs.manifest.Layers) < 1:
		return fmt.Errorf("no LFS layers present, LFS rebuild is not possible")
	}

	log.InfoContext(ctx, "Fetching LFS files from remote")
	err = f.updateLFSFromOCI(ctx, lfsManDesc, remoteLFSFiles)
	if err != nil {
		return fmt.Errorf("updating LFS files from OCI: %w", err)
	}

	log.InfoContext(ctx, "Pushing to git remote")
	if err := f.pushLFSRemote(f.dstGitRemote, refList...); err != nil {
		return fmt.Errorf("pushing to remote: %w", err)
	}

	log.InfoContext(ctx, "Successfully pushed LFS tracked files to remote repository")
	return nil
}

// pushLFSRemote pushes all LFS files to the remote, followed by a standard git push.
func (f *FromOCI) pushLFSRemote(gitRemote string, refList ...string) error {
	if len(refList) == 0 {
		refList = []string{"--all"}
	}

	// push LFS files
	if err := f.cmdHelper.LFS.Push(gitRemote, refList...); err != nil {
		return fmt.Errorf("pushing git lfs files to remote: %w", err)
	}

	// regular git push
	if err := f.pushRemote(gitRemote, refList, f.cmdHelper.Force); err != nil {
		return fmt.Errorf("pushing commits and refs to remote: %w", err)
	}

	return nil
}

// updateLFSFromOCI copies the minimum number of LFS files needed from an OCI repository, caching them accordingly.
func (f *FromOCI) updateLFSFromOCI(ctx context.Context, lfsManDesc ocispec.Descriptor, remoteLFSFiles []string) error {
	log := logger.FromContext(ctx)
	u := ui.FromContextOrNoop(ctx)

	switch {
	case f.syncOpts.Cache != nil:
		_, err := f.syncOpts.Cache.UpdateLFSFromOCI(ctx, f.ociHelper.Target, lfsManDesc)
		if err != nil {
			log.DebugContext(ctx, "Cache failed to update git-lfs objects", "error", err)
			u.Infof("Failed to update cache with git-lfs objects, continuing without caching...")
		} else {
			// log.InfoContext(ctx, "cached LFS files from OCI", "oids", cachedOIDs)
			log.InfoContext(ctx, "Linking cache lfs files to intermediate repository")
			if err := f.cmdHelper.Config("--add", "lfs.storage", filepath.Join(f.syncOpts.Cache.CachePath(), "lfs")); err != nil {
				return fmt.Errorf("setting lfs.storage config to cache: %w", err) // TODO: recover?
			}
			break
		}
		fallthrough // recover to default if cache fails

	default:
		// we can still optimize some even if caching is not used.
		resolver := make(map[string]struct{}, len(remoteLFSFiles))
		for _, oid := range remoteLFSFiles {
			resolver[oid] = struct{}{}
		}

		// compare new LFS manifest to the LFS files at the remote destination
		existingLFSFiles := make(map[string]int64, len(remoteLFSFiles))  // LFS files to skip
		newDescs := make([]ocispec.Descriptor, 0, len(existingLFSFiles)) // LFS files to fetch
		for _, layer := range f.lfs.manifest.Layers {
			_, ok := resolver[layer.Annotations[ocispec.AnnotationTitle]] // safer to use title instead of digest, for cryptographic agility
			if ok {
				existingLFSFiles[layer.Annotations[ocispec.AnnotationTitle]] = layer.Size // tricking LFS to think a file exists requires a file of the same size
			} else {
				newDescs = append(newDescs, layer)
			}
		}

		// do not update local copy of remote repo until existing LFS files are resolved
		if err := cmd.CreateFakeLFSFiles(f.cmdHelper.Dir(), existingLFSFiles); err != nil {
			return fmt.Errorf("creating fake LFS files: %w", err)
		}

		copyOpts := oras.CopyGraphOptions{
			// TODO: Using oras' default concurrency (3), until plumbing arrives here...
			PostCopy:       oci.PostCopyLFS(f.ociHelper.FStorePath, f.syncOpts.IntermediateDir),
			FindSuccessors: oci.FindSuccessorsLFS(newDescs),
		}

		if err := oras.CopyGraph(ctx, f.ociHelper.Target, f.ociHelper.FStore, lfsManDesc, copyOpts); err != nil {
			return fmt.Errorf("copying LFS files as OCI layers: %w", err)
		}
	}

	return nil
}
