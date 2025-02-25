package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// FromOCI represents an OCI to git sync action.
type FromOCI struct {
	sync

	dstGitRemote string
}

// NewFromOCI returns a FromOCI object after validating git and/or git-lfs compatibility. Unlike NewToOCI, the base manifest descriptor is not optional.
func NewFromOCI(ctx context.Context, target oras.GraphTarget, desc ocispec.Descriptor, dstGitRemote string, syncOpts SyncOptions, cmdOpts *cmd.Options) (*FromOCI, error) {
	fromOCI := &FromOCI{
		sync{
			base: syncBase{
				manDesc: desc,
			},
			lfs:      syncLFS{},
			syncOpts: syncOpts,
		},
		dstGitRemote,
	}

	fromOCI.ociHelper = &oci.Helper{
		Target:     target,
		FStore:     syncOpts.IntermediateStore,
		FStorePath: syncOpts.IntermediateDir,
	}

	var err error
	fromOCI.cmdHelper, err = cmd.NewHelper(ctx, syncOpts.IntermediateDir, cmdOpts)
	if err != nil {
		return nil, fmt.Errorf("createing new cmdHelper: %w", err)
	}

	if err := fromOCI.cmdHelper.ValidateVersions(ctx); err != nil {
		return nil, err
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
	u := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "Fetching commit manifest and config")
	err := f.FetchBaseManifestConfig(ctx)
	switch {
	case err != nil:
		return nil, fmt.Errorf("fetching commit manifest and config: %w", err)
	case len(f.base.manifest.Layers) < 1:
		return nil, fmt.Errorf("no bundle layers present, rebuild is not possible")
	}

	log.InfoContext(ctx, "Cloning git repo", "repo", f.dstGitRemote)
	if err := f.cloneRemote(ctx); err != nil {
		return nil, fmt.Errorf("cloning git remote %s to %s: %w", f.dstGitRemote, f.cmdHelper.Dir(), err)
	}

	refList, err := f.RefList()
	if err != nil {
		return nil, fmt.Errorf("getting reference list: %w", err)
	}

	// determine LFS files existing in our local copy of the remote
	// do not update local copy until existing LFS files are resolved
	var remoteLFSFiles []string
	if f.cmdHelper.WithLFS {
		_, refs, err := f.cmdHelper.LocalCommitsRefs(ctx)
		// if err, try to recover by assuming no reachable LFS files
		if err == nil {
			remoteLFSFiles, err = f.cmdHelper.ListReachableLFSFiles(ctx, refs)
			if err != nil {
				return nil, fmt.Errorf("resolving reachable LFS files at remote: %w", err)
			}
		}
	}

	switch {
	case f.syncOpts.Cache != nil:
		err := f.syncOpts.Cache.UpdateFromOCI(ctx, f.ociHelper.Target, f.base.manDesc)
		if err == nil {
			log.InfoContext(ctx, "utilizing cache", "path", f.syncOpts.Cache)
			break
		}
		log.DebugContext(ctx, "Cache failed to update git objects from bundles", "error", err)
		u.Infof("Failed to update cache with git objects, continuing without caching...")
		fallthrough // recover to default if cache fails

	default:
		log.InfoContext(ctx, "no cache specified")
		err := f.updateFromOCI(ctx, f.base.manDesc)
		if err != nil {
			return nil, fmt.Errorf("updating intermediate repo from OCI: %w", err)
		}
	}

	log.InfoContext(ctx, "Updating references in intermediate git repository")
	updated, err := f.updateAllRefs(ctx, f.dstGitRemote) // not all refs are guaranteed to be explicitly added to the bundles
	if err != nil {
		return nil, fmt.Errorf("updated tag and head references: %w", err)
	}
	log.InfoContext(ctx, "Updated references in intermediate git repository", "updated", updated)

	// if WithLFS, we'll wait to push everything all at once. We must push lfs files before
	// the git references.
	if f.cmdHelper.WithLFS {
		log.InfoContext(ctx, "Updating git LFS tracked files")
		err := f.runLFS(ctx, refList, f.base.manDesc, remoteLFSFiles)
		if err != nil && !errors.Is(err, errdef.ErrNotFound) {
			return updated, fmt.Errorf("running LFS: %w", err)
		}
	}
	log.InfoContext(ctx, "Pushing commits and references to remote repository")
	err = f.pushRemote(ctx, f.dstGitRemote, refList, f.cmdHelper.Force)
	if err != nil {
		return nil, fmt.Errorf("pushing commits and refs to remote repository: %w", err)
	}

	log.InfoContext(ctx, "Successfully pushed commits and references to remote repository")
	return updated, nil
}

// updateAllRefs updates all tag and head references based on a sync config. It compares
// the updates to the remote repository, returning the refs and corresponding updated commits that
// will be updated on a push, but does not actually push updates to the remote repo.
func (f *FromOCI) updateAllRefs(ctx context.Context, gitRemote string) ([]string, error) {

	if err := f.cmdHelper.RemoteAdd(ctx, "remote", gitRemote); err != nil {
		return nil, err
	}

	// parse existing tags
	rt, err := f.cmdHelper.LSRemote(ctx, "--tags", gitRemote)
	switch {
	case errors.Is(err, cmd.ErrRepoNotExistOrPermDenied):
		// recover
	case err != nil:
		return nil, err
	}

	remoteTags := make(map[string]cmd.Commit, len(rt))
	if len(rt) > 0 { // split result of remote ls
		for _, existingRef := range rt {
			split := strings.Fields(existingRef)
			oldCommit, fullRef := split[0], split[1]
			remoteTags[fullRef] = cmd.Commit(oldCommit)
		}
	}

	// parse existing heads
	rh, err := f.cmdHelper.LSRemote(ctx, "--heads", gitRemote)
	switch {
	case errors.Is(err, cmd.ErrRepoNotExistOrPermDenied):
		// recover
	case err != nil:
		return nil, err
	}

	remoteHeads := make(map[string]cmd.Commit, len(rh))
	if len(rh) > 0 { // split result of remote ls
		for _, existingRef := range rh {
			split := strings.Fields(existingRef)
			oldCommit, fullRef := split[0], split[1]
			remoteHeads[fullRef] = cmd.Commit(oldCommit)
		}
	}

	// update intermediate repo refs
	updated := make([]string, 0) // unable to predict number of possible updates
	if len(f.base.config.Refs.Tags) > 0 {
		tagUpdates, err := f.updateRefList(ctx, cmd.TagRefPrefix, f.base.config.Refs.Tags, remoteTags)
		if err != nil {
			return nil, fmt.Errorf("updating tag references in intermediate repo: %w", err)
		}
		updated = append(updated, tagUpdates...)

	}
	if len(f.base.config.Refs.Heads) > 0 {
		headUpdates, err := f.updateRefList(ctx, cmd.HeadRefPrefix, f.base.config.Refs.Heads, remoteHeads)
		if err != nil {
			return nil, fmt.Errorf("updating head referencesin intermediate repo: %w", err)
		}
		updated = append(updated, headUpdates...)
	}

	err = f.cmdHelper.RemoteRemove(ctx, "remote")
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// updateRefList updates references in the intermediate repository, returning a
// slice of the references and their corresponding updated commits that will
// ultimately be updated at the remote destination.
func (f *FromOCI) updateRefList(ctx context.Context, prefixType string,
	refs map[string]oci.ReferenceInfo, remoteRefs map[string]cmd.Commit) ([]string, error) {
	var updated []string
	for ref, refInfo := range refs {
		fullRef := path.Join(prefixType, ref)
		oldCommit, ok := remoteRefs[fullRef]
		if !ok || oldCommit != refInfo.Commit {
			updated = append(updated, fmt.Sprintf("%s %s", refInfo.Commit, fullRef))
		}
		err := f.cmdHelper.UpdateRef(ctx, fullRef, string(refInfo.Commit))
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}

// resolveLayersNeeded returns a list of the minimum bundle layers needed to update the remote
// by comparing a recently fetched git manifest to the destination git remote.
func (f *FromOCI) resolveLayersNeeded(ctx context.Context) ([]ocispec.Descriptor, error) {

	// layer digest to layer index resolver
	layerNumResolver := make(map[digest.Digest]int, len(f.base.manifest.Layers))
	for i, layerDesc := range f.base.manifest.Layers {
		layerNumResolver[layerDesc.Digest] = i
	}

	layerCutoff := len(f.base.manifest.Layers) // start cutoff after the total num of layers
	for tag, refInfo := range f.base.config.Refs.Tags {
		// only check for a possible update if the ref is before the cutoff
		if layerNumResolver[refInfo.Layer] < layerCutoff {
			fullTagRef := path.Join(cmd.TagRefPrefix, tag)           // references don't use OS-specific path separators
			refCommits, err := f.cmdHelper.ShowRefs(ctx, fullTagRef) // returned slice should be of length 1
			if err != nil {
				// try to recover by assuming this ref DNE
				layerCutoff = layerNumResolver[refInfo.Layer]
				continue
			}

			split := strings.Fields(refCommits[0])
			remoteCommit := cmd.Commit(split[0]) // not technically the remote, but our intermediate repo should be identical at this point
			if refInfo.Commit != remoteCommit {
				layerCutoff = layerNumResolver[refInfo.Layer]
			}
		}
	}

	for head, refInfo := range f.base.config.Refs.Heads {
		// only check for a possible update if the ref is before the cutoff
		if layerNumResolver[refInfo.Layer] < layerCutoff {
			fullHeadRef := path.Join(cmd.HeadRefPrefix, head)         // references don't use OS-specific path separators
			refCommits, err := f.cmdHelper.ShowRefs(ctx, fullHeadRef) // returned slice should be of length 1
			if err != nil {
				// try to recover by assuming this ref DNE
				layerCutoff = layerNumResolver[refInfo.Layer]
				continue
			}

			split := strings.Fields(refCommits[0])
			remoteCommit := cmd.Commit(split[0]) // not technically the remote, but our intermediate repo should be identical at this point
			if refInfo.Commit != remoteCommit {
				layerCutoff = layerNumResolver[refInfo.Layer]
			}
		}
	}

	return f.base.manifest.Layers[layerCutoff:], nil
}

// pushRemote pushes all commits reachable from refList to the remote repository.
func (f *FromOCI) pushRemote(ctx context.Context, gitRemote string, refList []string, force bool) (err error) {
	if force {
		err = f.cmdHelper.Git.Push(ctx, gitRemote, "--mirror") // implies --force but also includes all refs/*, which is only head and tag refs as created by sync
		if err != nil {
			return err
		}
	} else {
		err = f.cmdHelper.Git.Push(ctx, gitRemote, refList...)
		if err != nil {
			return fmt.Errorf("pushing to remote: %w", err)
		}
	}
	return nil
}

// updateFromOCI copies the minimal set of git bundles, and fetches their contents into the
// intermediate repo, or the cache if appropriate.
func (f *FromOCI) updateFromOCI(ctx context.Context, manDesc ocispec.Descriptor) error {
	log := logger.FromContext(ctx)

	// determine min set of bundles and copy
	bundleDescs, err := f.resolveLayersNeeded(ctx)
	if err != nil {
		return fmt.Errorf("resolving bundle layers containing missing objects: %w", err)
	}
	if len(bundleDescs) < 1 {
		// nothing to update
		return nil
	}

	copyOpts := oras.CopyGraphOptions{
		// PreCopy func simply logs
		PreCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
			if desc.MediaType == oci.MediaTypeBundleLayer {
				log.InfoContext(ctx, "Retrieving bundle layer", "bundle", desc.Annotations[ocispec.AnnotationTitle])
			}
			return nil
		},
		FindSuccessors: oci.FindSuccessorsBundles(manDesc, bundleDescs),
	}

	err = oras.CopyGraph(ctx, f.ociHelper.Target, f.ociHelper.FStore, manDesc, copyOpts)
	if err != nil {
		return fmt.Errorf("copying bundles: %w", err)
	}

	// fetch from copied bundles
	log.InfoContext(ctx, "adding bundles as remotes")
	shortnames := make([]string, 0, len(bundleDescs))
	for _, desc := range bundleDescs {
		// resolve bundle path
		bundleName := desc.Annotations[ocispec.AnnotationTitle]
		bundlePath := filepath.Join(f.ociHelper.FStorePath, bundleName)

		// add bundle as a remote
		shortname := strings.TrimSuffix(bundleName, ".bundle")
		err := f.cmdHelper.RemoteAdd(ctx, shortname, bundlePath)
		if err != nil {
			return fmt.Errorf("adding bundle '%s' as remote: %w", bundlePath, err)
		}
		shortnames = append(shortnames, shortname)
	}

	args := append(shortnames, "--tags", "--multiple") //nolint
	if f.cmdHelper.Force {
		args = append(args, "--force")
		log.InfoContext(ctx, "force fetching from bundles")
	} else {
		log.InfoContext(ctx, "fetching from bundles")
	}
	if err := f.cmdHelper.Git.Fetch(ctx, args...); err != nil {
		return fmt.Errorf("fetching from bundles: %w", err)
	}

	// remove remotes
	// TODO: is this necessary?
	log.InfoContext(ctx, "removing bundles from remotes")
	for _, shortname := range shortnames {
		err := f.cmdHelper.RemoteRemove(ctx, shortname)
		if err != nil {
			return fmt.Errorf("removing remote bundle: %w", err)
		}
	}

	return nil
}

// GetTagRefs returns the ReferenceInfo for tags from the commit manifest's config.
func (f *FromOCI) GetTagRefs() (map[string]oci.ReferenceInfo, error) {
	if f.sync.base.config.Refs.Tags == nil {
		return nil, fmt.Errorf("tag reference info map in config has not been initialized")
	}
	return f.sync.base.config.Refs.Tags, nil
}

// GetHeadRefs returns the ReferenceInfo for heads from the commit manifest's config.
func (f *FromOCI) GetHeadRefs() (map[string]oci.ReferenceInfo, error) {
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
func (f *FromOCI) cloneRemote(ctx context.Context) error {
	if f.dstGitRemote == "" {
		return fmt.Errorf("no source git remote specified, unable to clone")
	}

	cachePath := ""
	if f.syncOpts.Cache != nil {
		cachePath = f.syncOpts.Cache.CachePath()
	}

	err := f.cmdHelper.CloneWithShared(ctx, f.dstGitRemote, cachePath)
	switch {
	case errors.Is(err, cmd.ErrRepoNotExistOrPermDenied):
		if err := os.MkdirAll(f.syncOpts.IntermediateDir, 0777); err != nil {
			return fmt.Errorf("creating intermediate repository directory: %w", err)
		}
		if err := f.cmdHelper.InitializeRepo(ctx); err != nil {
			return fmt.Errorf("initializing intermediate repository when attempting to recover from failed clone: %w", err)
		}
		return nil // recover successful
	case err != nil:
		return err
	}

	return nil
}
