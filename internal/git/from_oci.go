package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cache"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// FromOCI represents an OCI to git sync action.
type FromOCI struct {
	sync

	dstGitRemote string
}

// NewFromOCI returns a FromOCI object after validating git and/or git-lfs compatibility.
func NewFromOCI(ctx context.Context, target oras.GraphTarget, tag, dstGitRemote string, syncOpts SyncOptions, cmdOpts *cmd.Options) (*FromOCI, error) {
	u := ui.FromContextOrNoop(ctx)

	fromOCI := &FromOCI{
		sync{
			base:     syncBase{},
			lfs:      syncLFS{},
			cache:    nil,
			syncOpts: syncOpts,
		},
		dstGitRemote,
	}

	var err error
	fromOCI.ociHelper, err = oci.NewOCIHelper(syncOpts.TmpDir, target, tag)
	if err != nil {
		return nil, fmt.Errorf("creating new ociHelper: %w", err)
	}

	fromOCI.cmdHelper, err = cmd.NewHelper(ctx, syncOpts.TmpDir, cmdOpts)
	if err != nil {
		return nil, fmt.Errorf("createing new cmdHelper: %w", err)
	}

	if syncOpts.CacheDir != "" {
		if err := os.MkdirAll(syncOpts.CacheDir, 0o777); err != nil {
			u.Infof("Unable to create cache directory, continuing without caching.")
			return fromOCI, nil
		}

		fromOCI.cache, err = cache.NewCache(ctx, syncOpts.CacheDir, cmdOpts)
		if err != nil {
			u.Infof("Unable to access git object cache, continuing without caching.")
			return fromOCI, nil
		}
	}

	return fromOCI, nil
}

// Cleanup cleans up any temporary files created during the FromOCI process.
func (f *FromOCI) Cleanup() error {
	err := f.sync.cleanup()
	if err != nil {
		return fmt.Errorf("cleaning up sync: %w", err)
	}

	return nil
}

// Run updates a remote git repository by fetching changes from a commit manifest, pushes the changes to the remote git
// reference and returns a slice of all updated tag and head references.
func (f *FromOCI) Run(ctx context.Context) ([]string, error) {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Fetching commit manifest and config")
	manifestDesc, err := f.FetchBaseManifestConfig(ctx)
	switch {
	case err != nil:
		return nil, fmt.Errorf("fetching commit manifest and config: %w", err)
	case len(f.base.manifest.Layers) < 1:
		return nil, fmt.Errorf("no bundle layers present, rebuild is not possible")
	}

	log.InfoContext(ctx, "Cloning git repo", "repo", f.dstGitRemote, "cache", f.syncOpts.CacheDir)
	if err := f.cloneRemote(); err != nil {
		return nil, fmt.Errorf("cloning git remote %s to %s: %w", f.dstGitRemote, f.cmdHelper.Dir(), err)
	}

	bundleDescs, err := f.resolveLayersNeeded()
	if err != nil {
		return nil, fmt.Errorf("resolving bundle layers containing uncached objects: %w", err)
	}

	log.InfoContext(ctx, "Copying bundles", "bundleDescriptors", bundleDescs)
	if err := f.copyBundles(ctx, bundleDescs...); err != nil {
		return nil, fmt.Errorf("fetching bundle contents: %w", err)
	}

	if err := f.fetchFromBundles(ctx, bundleDescs...); err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "Updating references in intermediate git repository")
	updated, err := f.updateAllRefs(f.dstGitRemote) // not all refs are guaranteed to be explicitly added to the bundles
	if err != nil {
		return nil, fmt.Errorf("updated tag and head references: %w", err)
	}
	log.InfoContext(ctx, "Updated references in intermediate git repository", "updated", updated)

	refList, err := f.RefList()
	if err != nil {
		return nil, fmt.Errorf("getting reference list: %w", err)
	}

	// if WithLFS, we'll wait to push everything all at once. We must push lfs files before
	// the git references.
	if f.cmdHelper.WithLFS {
		log.InfoContext(ctx, "Updating git LFS tracked files")
		err := f.runLFS(ctx, refList, manifestDesc)
		if err != nil && !errors.Is(err, errdef.ErrNotFound) {
			return updated, fmt.Errorf("running LFS: %w", err)
		}
	}
	log.InfoContext(ctx, "Pushing commits and references to remote repository")
	err = f.pushRemote(f.dstGitRemote, refList, f.cmdHelper.Force)
	if err != nil {
		return nil, fmt.Errorf("pushing commits and refs to remote repository: %w", err)
	}

	log.InfoContext(ctx, "Successfully pushed commits and references to remote repository")
	return updated, nil
}

// updateAllRefs updates all tag and head references based on a sync config. It compares
// the updates to the remote repository to determine which refs will be updated on a push,
// but does not actually update the remote repo.
func (f *FromOCI) updateAllRefs(gitRemote string) ([]string, error) {
	if err := f.cmdHelper.RemoteAdd("remote", gitRemote); err != nil {
		return nil, err
	}

	// parse existing tags
	rt, err := f.cmdHelper.LSRemote("--tags", gitRemote)
	switch {
	case errors.Is(err, cmd.ErrRepoNotExistOrPermDenied):
		// recover
	case err != nil:
		return nil, err
	}

	remoteTags := make(map[string]Commit, len(rt))
	if len(rt) > 0 { // split result of remote ls
		for _, existingRef := range rt {
			split := strings.Fields(existingRef)
			oldCommit, fullRef := split[0], split[1]
			remoteTags[fullRef] = Commit(oldCommit)
		}
	}

	// parse existing heads
	rh, err := f.cmdHelper.LSRemote("--heads", gitRemote)
	switch {
	case errors.Is(err, cmd.ErrRepoNotExistOrPermDenied):
		// recover
	case err != nil:
		return nil, err
	}

	remoteHeads := make(map[string]Commit, len(rh))
	if len(rh) > 0 { // split result of remote ls
		for _, existingRef := range rh {
			split := strings.Fields(existingRef)
			oldCommit, fullRef := split[0], split[1]
			remoteHeads[fullRef] = Commit(oldCommit)
		}
	}

	// update intermediate repo refs
	updated := make([]string, 0) // unable to predict number of possible updates
	if len(f.base.config.Refs.Tags) > 0 {
		tagUpdates, err := f.updateRefList(cmd.TagRefPrefix, f.base.config.Refs.Tags, remoteTags)
		if err != nil {
			return nil, fmt.Errorf("updating tag references: %w", err)
		}
		updated = append(updated, tagUpdates...)
	}
	if len(f.base.config.Refs.Heads) > 0 {
		headUpdates, err := f.updateRefList(cmd.HeadRefPrefix, f.base.config.Refs.Heads, remoteHeads)
		if err != nil {
			return nil, fmt.Errorf("updating head references: %w", err)
		}
		updated = append(updated, headUpdates...)
	}

	err = f.cmdHelper.RemoteRemove("remote")
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// updateRefList updates references, returning a slice of the references and their corresponding
// commits that will ultimately be updated at the remote destination.
// TODO: Do we want to show which commit a ref was updated from? This may make it easier for users to do diffs.
// Also this is the typical behavior of a pull.
func (f *FromOCI) updateRefList(prefixType string, refs map[string]ReferenceInfo, remoteRefs map[string]Commit) ([]string, error) {
	var updated []string
	for ref, refInfo := range refs {
		fullRef := prefixType + ref
		oldCommit, ok := remoteRefs[fullRef]
		if !ok || oldCommit != refInfo.Commit {
			updated = append(updated, fmt.Sprintf("%s %s", refInfo.Commit, fullRef))
		}
		err := f.cmdHelper.UpdateRef(fullRef, string(refInfo.Commit))
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}

// resolveLayersNeeded returns a list of the minimum bundle layers needed to update the remote.
func (f *FromOCI) resolveLayersNeeded() ([]ocispec.Descriptor, error) {
	layerNumResolver := make(map[digest.Digest]int, len(f.base.manifest.Layers))
	for i, layerDesc := range f.base.manifest.Layers {
		layerNumResolver[layerDesc.Digest] = i
	}

	layerCutoff := len(f.base.manifest.Layers) // start cutoff after the total num of layers
	for tag, refInfo := range f.base.config.Refs.Tags {
		// only check for a possible update if the ref is before the cutoff
		if layerNumResolver[refInfo.Layer] < layerCutoff {
			fullTagRef := filepath.Join(cmd.TagRefPrefix, tag)
			refCommits, err := f.cmdHelper.ShowRefs(fullTagRef) // returned slice should be of length 1
			if err != nil {
				// try to recover by assuming this ref DNE
				layerCutoff = layerNumResolver[refInfo.Layer]
				continue
			}

			split := strings.Fields(refCommits[0])
			remoteCommit := Commit(split[0]) // not technically the remote, but our intermediate repo should be identical at this point
			if refInfo.Commit != remoteCommit {
				layerCutoff = layerNumResolver[refInfo.Layer]
			}
		}
	}

	for head, refInfo := range f.base.config.Refs.Heads {
		// only check for a possible update if the ref is before the cutoff
		if layerNumResolver[refInfo.Layer] < layerCutoff {
			fullHeadRef := filepath.Join(cmd.HeadRefPrefix, head)
			refCommits, err := f.cmdHelper.ShowRefs(fullHeadRef) // returned slice should be of length 1
			if err != nil {
				// try to recover by assuming this ref DNE
				layerCutoff = layerNumResolver[refInfo.Layer]
				continue
			}

			split := strings.Fields(refCommits[0])
			remoteCommit := Commit(split[0]) // not technically the remote, but our intermediate repo should be identical at this point
			if refInfo.Commit != remoteCommit {
				layerCutoff = layerNumResolver[refInfo.Layer]
			}
		}
	}

	return f.base.manifest.Layers[layerCutoff:], nil
}

// pushRemote pushes all commits reachable from refList to the remote repository.
func (f *FromOCI) pushRemote(gitRemote string, refList []string, force bool) (err error) {
	if force {
		err = f.cmdHelper.Git.Push(gitRemote, "--mirror") // implies --force but also includes all refs/*, which is only head and tag refs as created by sync
		if err != nil {
			return err
		}
	} else {
		err = f.cmdHelper.Git.Push(gitRemote, refList...)
		if err != nil {
			return fmt.Errorf("pushing to remote: %w", err)
		}
	}
	return nil
}

// copyBundles copies a set of bundle layers from an oci target.
func (f *FromOCI) copyBundles(ctx context.Context, bundleDescs ...ocispec.Descriptor) error {
	log := logger.FromContext(ctx)

	for _, bundleDesc := range bundleDescs {
		if bundleDesc.MediaType != MediaTypeBundleLayer {
			return fmt.Errorf("expected bundle media type %s, got %s", MediaTypeBundleLayer, bundleDesc.MediaType)
		}

		bundleName := bundleDesc.Annotations[ocispec.AnnotationTitle]

		// fetch a bundle
		log.InfoContext(ctx, "Retrieving bundle layer", "bundle", bundleName)
		err := oras.CopyGraph(ctx, f.ociHelper.Target, f.ociHelper.FStore, bundleDesc, oras.DefaultCopyGraphOptions)
		if err != nil {
			return fmt.Errorf("fetching layer %s bytes: %w", bundleDesc.Digest, err)
		}
	}

	return nil
}

// fetchFromBundles fetches to the cache or intermediate repo from git bundles stored in an oci filestore.
func (f *FromOCI) fetchFromBundles(ctx context.Context, bundleDescs ...ocispec.Descriptor) error {
	log := logger.V(logger.FromContext(ctx), 1)
	u := ui.FromContextOrNoop(ctx)

	bundleErrs := make([]error, 0)
	var cacheErr bool
	for _, desc := range bundleDescs {
		bundleName := desc.Annotations[ocispec.AnnotationTitle]
		bundlePath := filepath.Join(f.ociHelper.FStorePath, bundleName)

		switch {
		case f.syncOpts.CacheDir != "":
			log.InfoContext(ctx, "fetching bundle contents into cache", "bundlePath", bundlePath)
			err := f.cache.FetchFromBundle(bundlePath, f.cmdHelper.Options)
			if err == nil {
				continue
			}
			log.DebugContext(ctx, "Cache failed to update git objects", "error", err)
			cacheErr = true
			fallthrough // a cache failure is not fatal
		default:
			log.InfoContext(ctx, "fetching bundle contents into intermediate repository", "bundlePath", bundlePath)
			err := f.cmdHelper.FetchFromBundle(bundlePath)
			if err != nil {
				bundleErrs = append(bundleErrs, fmt.Errorf("fetching from bundle %s: %w", bundleName, err))
			}
		}
	}
	if cacheErr { // only print cache failure notice once
		u.Infof("Failed to update cache with contents of some git bundles, continuing without caching...")
	}
	return errors.Join(bundleErrs...)
}

// GetTagRefs returns the ReferenceInfo for tags from the commit manifest's config.
func (f *FromOCI) GetTagRefs() (map[string]ReferenceInfo, error) {
	if f.sync.base.config.Refs.Tags == nil {
		return nil, fmt.Errorf("tag reference info map in config has not been initialized")
	}
	return f.sync.base.config.Refs.Tags, nil
}

// GetHeadRefs returns the ReferenceInfo for heads from the commit manifest's config.
func (f *FromOCI) GetHeadRefs() (map[string]ReferenceInfo, error) {
	if f.sync.base.config.Refs.Heads == nil {
		return nil, fmt.Errorf("head reference info map in config has not been initialized")
	}
	return f.sync.base.config.Refs.Heads, nil
}

// RefList returns a list of all refences in the commit manifest's config, filtering out
// the commits they reference.
func (f *FromOCI) RefList() ([]string, error) {
	tagCommits, err := f.GetTagRefs()
	if err != nil {
		return nil, err
	}
	headCommits, err := f.GetHeadRefs()
	if err != nil {
		return nil, err
	}

	// filter out commits
	refs := make([]string, 0, len(tagCommits)+len(headCommits))
	for tag := range tagCommits {
		refs = append(refs, tag)
	}
	for head := range headCommits {
		refs = append(refs, head)
	}

	return refs, nil
}

// cloneRemote clones a remote destination git repository to the intermediate directory. The clone
// always attempts to reference the cache. A non-existent cache is not fatal.
func (f *FromOCI) cloneRemote() error {
	if f.dstGitRemote == "" {
		return fmt.Errorf("no source git remote specified, unable to clone")
	}

	err := f.cmdHelper.CloneWithShared(f.dstGitRemote, f.syncOpts.CacheDir)
	switch {
	case errors.Is(err, cmd.ErrRepoNotExistOrPermDenied):
		if err := os.MkdirAll(f.syncOpts.TmpDir, 0o777); err != nil {
			return fmt.Errorf("creating intermediate repository directory: %w", err)
		}
		if err := f.cmdHelper.InitializeRepo(); err != nil {
			return fmt.Errorf("initializing intermediate repository when attempting to recover from failed clone: %w", err)
		}
		return nil // recover successful
	case err != nil:
		return err
	}

	return nil
}
