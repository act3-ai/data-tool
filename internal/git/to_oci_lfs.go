package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/git/cmd"
	"git.act3-ace.com/ace/data/tool/internal/ui"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// RunLFS creates or updates an LFS manifest with git LFS tracked files modified between two sets of commit tips. It is acceptable for oldTips to be empty, but this
// is only recommended in cases of a clean manifest. Returns a descriptor for the LFS manifest.
//
// RunLFS SHOULD be called after Run, MUST NOT be called if a user does not have git-lfs installed, and
// MAY be called if a repository does not have LFS enabled.
func (t *ToOCI) runLFS(ctx context.Context, oldCommitManDesc, newCommitManDesc ocispec.Descriptor) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	reachableLFSObjs, err := t.cmdHelper.ListReachableLFSFiles(t.argRevList...)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("resolving status of LFS files: %w", err)
	}

	if len(reachableLFSObjs) < 1 {
		// repository does not have any LFS files
		return ocispec.Descriptor{}, cmd.ErrLFSNotEnabled
	}

	if err := t.FetchLFSManifestConfig(ctx, oldCommitManDesc, t.syncOpts.Clean); err != nil && !errors.Is(err, errdef.ErrNotFound) {
		return ocispec.Descriptor{}, err
	}

	if err := t.prepRepoForLFS(); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("prepping tempoarary intermediate git repository for handling lfs: %w", err)
	}

	log.InfoContext(ctx, "Fetching new reachable LFS files", "srcGitRemote", t.srcGitRemote, "argRevList", t.argRevList)
	if err := t.fetchLFSFilesGit(ctx, t.srcGitRemote, t.argRevList...); err != nil { // caching handled here
		return ocispec.Descriptor{}, fmt.Errorf("fetching new reachable lfs files: %w", err)
	}

	objsInRegistry := t.getExistingLFSFiles() // do not modify lfs manifest before this
	newLFSObjs := excludeLFSFiles(reachableLFSObjs, objsInRegistry)

	log.InfoContext(ctx, "Pushing LFS sync")
	lfsManDesc, err := t.sendLFSSync(ctx, &newCommitManDesc, newLFSObjs)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("sending LFS manifest: %w", err)
	}

	return lfsManDesc, nil
}

// createFakeLFSFiles creates LFS files, that already exist in the destination registry,
// within the intermediate repo to prevent fetching with `git-lfs fetch`.
func (t *ToOCI) createFakeLFSFiles() error {
	objsInRegistry := t.getExistingLFSFiles() // do not modify lfs manifest before this
	for obj, size := range objsInRegistry {
		oidPath := filepath.Join(t.cmdHelper.Dir(), t.cmdHelper.ResolveLFSOIDPath(obj))
		if err := os.MkdirAll(filepath.Dir(oidPath), 0777); err != nil {
			return fmt.Errorf("creating path to empty lfs obj: %w", err)
		}

		oidFile, err := os.Create(oidPath)
		if err != nil {
			return fmt.Errorf("creating empty lfs obj: %w", err)
		}

		_, err = oidFile.WriteAt([]byte{1}, size-1)
		if err != nil {
			return fmt.Errorf("writing to obj file at offset %d: %w", size-1, err)
		}
		if err := oidFile.Close(); err != nil {
			return fmt.Errorf("closing obj file: %w", err)
		}
	}

	return nil
}

// fetchLFSFilesGit fetches all git LFS tracked files from gitRemote reachable from argRevList.
// Handles caching if it is used.
func (t *ToOCI) fetchLFSFilesGit(ctx context.Context, gitRemote string, argRevList ...string) error {
	log := logger.FromContext(ctx)
	u := ui.FromContextOrNoop(ctx)

	switch {
	case t.syncOpts.CacheDir != "":
		commits, _, err := t.localCommitsRefs(argRevList...)
		if err != nil {
			return fmt.Errorf("resolving argRevList to commits: %w", err)
		}

		// TODO: Not a big fan of this side effect of defining a custom type...
		commitsAsStr := make([]string, 0, len(commits))
		for _, commit := range commits {
			commitsAsStr = append(commitsAsStr, string(commit))
		}

		if err := t.cache.UpdateLFSFromGit(gitRemote, t.cmdHelper.Options, commitsAsStr...); err != nil {
			log.DebugContext(ctx, "Cache failed to update git-lfs objects", "error", err)
			u.Infof("Failed to update cache with git-lfs objects, continuing without caching...")
		} else {

			log.InfoContext(ctx, "Linking cache lfs files to intermediate repository")
			if err := t.cmdHelper.Config("--add", "lfs.storage", filepath.Join(t.cache.CachePath(), cmd.LFSObjsPath)); err != nil {
				return fmt.Errorf("setting lfs.storage config to cache: %w", err)
			}
			break
		}
		fallthrough // recover to default if cache fails

	default:
		if err := t.createFakeLFSFiles(); err != nil { // we can still optimize some even if caching is not used.
			return fmt.Errorf("creating fake LFS files to trick git-lfs: %w", err)
		}

		args := []string{"--all"}
		args = append(args, argRevList...)
		err := t.cmdHelper.LFS.Fetch(gitRemote, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

// sendLFSSync transfers the new LFS layers, config, and manifest to the target repository. Returns an OCI
// descriptor of the LFS Manifest.
func (t *ToOCI) sendLFSSync(ctx context.Context, subject *ocispec.Descriptor, newLFSObjs []string) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	lfsObjDescs := make([]ocispec.Descriptor, 0, len(newLFSObjs))
	for _, oid := range newLFSObjs {

		var lfsDir string
		if t.syncOpts.CacheDir != "" {
			lfsDir = t.cache.CachePath()
		} else {
			lfsDir = t.cmdHelper.Dir()
		}
		desc, err := t.ociHelper.FStore.Add(ctx, oid, MediaTypeLFSLayer, filepath.Join(lfsDir, t.cmdHelper.ResolveLFSOIDPath(oid)))
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("adding LFS object to filestore: %w", err)
		}
		lfsObjDescs = append(lfsObjDescs, desc)
		// t.lfs.config.Objects[desc.Digest.String()] = oid

		log.InfoContext(ctx, "Pushing new LFS layer", "digest", desc.Digest)
		if err := oras.CopyGraph(ctx, t.ociHelper.FStore, t.ociHelper.Target, desc, oras.DefaultCopyGraphOptions); err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("copying LFS layer to target repository: %w", err)
		}
	}

	log.InfoContext(ctx, "Pushing LFS manifest")
	manOpts := oras.PackManifestOptions{
		Subject: subject,
		Layers:  append(t.lfs.manifest.Layers, lfsObjDescs...),
		//ConfigDescriptor:    &configDesc,
		ManifestAnnotations: map[string]string{ocispec.AnnotationCreated: "1970-01-01T00:00:00Z", AnnotationDTVersion: t.syncOpts.DTVersion}, // POSIX epoch
	}

	manDesc, err := oras.PackManifest(ctx, t.ociHelper.Target, oras.PackManifestVersion1_1, ArtifactTypeLFSManifest, manOpts)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("packing and pushing LFS manifest: %w", err)
	}

	return manDesc, nil
}

// updateLFSManSubject updates a referring LFS manifest's subject to the new base manifest's descriptor.
// Does not return an error if a referring LFS manifest does not exist.
func (t *ToOCI) updateLFSManSubject(ctx context.Context, oldBaseManDesc, newBaseManDesc ocispec.Descriptor) error {
	log := logger.FromContext(ctx)
	u := ui.FromContextOrNoop(ctx)

	err := t.FetchLFSManifestConfig(ctx, oldBaseManDesc, false)
	switch {
	case errors.Is(err, errdef.ErrNotFound):
		log.InfoContext(ctx, "LFS manifest does not exist, updating subject is unnecessary")
		return nil
	case err != nil:
		return fmt.Errorf("fetching LFS manifest: %w", err)
	default:
		log.InfoContext(ctx, "Pushing LFS manifest")
		manOpts := oras.PackManifestOptions{
			Subject: &newBaseManDesc,
			Layers:  t.lfs.manifest.Layers,
			//ConfigDescriptor:    &configDesc,
			ManifestAnnotations: map[string]string{ocispec.AnnotationCreated: "1970-01-01T00:00:00Z", AnnotationDTVersion: t.syncOpts.DTVersion}, // POSIX epoch
		}

		lfsManDesc, err := oras.PackManifest(ctx, t.ociHelper.Target, oras.PackManifestVersion1_1, ArtifactTypeLFSManifest, manOpts)
		if err != nil {
			return fmt.Errorf("packing and pushing LFS manifest: %w", err)
		}

		u.Infof("Warning: LFS files added since the prior sync have not been updated. Run again with the --lfs option to update LFS files.")
		u.Infof("Updated LFS Manifest digest: %s", lfsManDesc.Digest)
		return nil
	}
}

// getExistingLFSFiles returns a slice of LFS file names (not paths) that exist in the LFS manifest.
// If this function is called immediately after fetching an LFS manifest, it may be assumed that these LFS files
// exist in the registry.
func (t *ToOCI) getExistingLFSFiles() map[string]int64 {

	objects := make(map[string]int64, len(t.lfs.manifest.Layers))
	for _, layer := range t.lfs.manifest.Layers {
		objects[layer.Annotations[ocispec.AnnotationTitle]] = layer.Size
	}

	return objects
}

// prepRepoForLFS prepares the intermediate git repository for pushing LFS files.
func (t *ToOCI) prepRepoForLFS() error {
	// apply optional lfs server url override
	if t.cmdHelper.ServerURL != "" {
		if err := t.cmdHelper.Config("lfs.url", t.cmdHelper.ServerURL); err != nil {
			return fmt.Errorf("setting up git config with LFS server URL for base lfs repo: %w", err)
		}
	}
	return nil
}

// excludeLFSFiles removes the excluded files from lfsFilePaths in-place.
func excludeLFSFiles(lfsObjs []string, exclusions map[string]int64) []string {
	j := 0
	for i, obj := range lfsObjs {
		if _, ok := exclusions[obj]; ok {
			lfsObjs[j], lfsObjs[i] = lfsObjs[i], lfsObjs[j]
			j++
		}
	}

	return lfsObjs[j:]
}
