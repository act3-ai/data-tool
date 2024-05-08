package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/ui"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// runLFS updates a remote git repository by fetching git LFS tracked files from an LFS manifest
// and pushes the changes to the remote git reference.
//
// runLFS should only be called after Run.
func (f *FromOCI) runLFS(ctx context.Context, refList []string, commitManifest ocispec.Descriptor) error {
	log := logger.FromContext(ctx)

	if err := f.cmdHelper.ConfigureLFS(); err != nil {
		return fmt.Errorf("configuring LFS: %w", err)
	}

	log.InfoContext(ctx, "Fetching LFS manifest and config")
	err := f.FetchLFSManifestConfig(ctx, commitManifest, false)
	switch {
	case errors.Is(err, errdef.ErrNotFound):
		return fmt.Errorf("LFS manifest does not exist: %w", err)
	case err != nil:
		return fmt.Errorf("fetching existing LFS manifest: %w", err)
	case len(f.lfs.manifest.Layers) < 1:
		return fmt.Errorf("no LFS layers present, LFS rebuild is not possible")
	}

	log.InfoContext(ctx, "Fetching LFS files from registry")
	if err := f.fetchLFSFilesOCI(ctx); err != nil { // caching handled here
		return fmt.Errorf("fetching LFS files from LFS manifest layers: %w", err)
	}

	log.InfoContext(ctx, "Pushing to git remote")
	if err := f.pushLFSRemote(refList, f.dstGitRemote); err != nil {
		return fmt.Errorf("pushing to remote: %w", err)
	}

	log.InfoContext(ctx, "Successfully pushed LFS tracked files to remote repository")
	return nil
}

// fetchLFSFilesOCI fetches the necessary LFS layers in the OCI manifest.
// Handles caching if it is used.
func (f *FromOCI) fetchLFSFilesOCI(ctx context.Context) error {
	log := logger.FromContext(ctx)
	u := ui.FromContextOrNoop(ctx)

	switch {
	case f.syncOpts.CacheDir != "":
		log.InfoContext(ctx, "Updating cache with git-lfs files")
		err := f.cache.UpdateLFSFromOCI(ctx, f.ociHelper.Target, f.cmdHelper.Options, f.lfs.manifest.Layers)
		if err != nil {
			log.DebugContext(ctx, "Cache failed to update git-lfs objects", "error", err)
			u.Infof("Failed to update cache with git-lfs objects, continuing without caching...")
		} else {
			log.InfoContext(ctx, "Linking cache lfs files to intermediate repository")
			if err := f.cmdHelper.Config("--add", "lfs.storage", filepath.Join(f.cache.CachePath(), "lfs")); err != nil {
				return fmt.Errorf("setting lfs.storage config to cache: %w", err)
			}
			break
		}
		fallthrough // recover to default if cache fails
	default:
		for _, layerDesc := range f.lfs.manifest.Layers {
			oid := layerDesc.Digest.Hex()
			oidPath := filepath.Join(f.cmdHelper.Dir(), f.cmdHelper.ResolveLFSOIDPath(oid))
			log.InfoContext(ctx, "Fetching git-lfs object to intermediate repo without caching", "oidPath", oidPath)
			if err := f.ociHelper.CopyLFSFromOCI(ctx, oidPath, layerDesc); err != nil {
				return fmt.Errorf("")
			}
		}
	}

	return nil
}

// pushLFSRemote pushes all LFS files to the remote, followed by a standard git push.
func (f *FromOCI) pushLFSRemote(refList []string, gitRemote string) error {

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
