package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/git/cache"
	"git.act3-ace.com/ace/data/tool/internal/git/cmd"
	"git.act3-ace.com/ace/data/tool/internal/git/oci"
	"git.act3-ace.com/ace/data/tool/internal/ui"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// ToOCI represents a git to OCI sync action.
type ToOCI struct {
	sync

	srcGitRemote string
	argRevList   []string
}

// NewToOCI returns a ToOCI object after validating git and/or git-lfs compatibility.
func NewToOCI(ctx context.Context, target oras.GraphTarget, tag, srcGitRemote string, argRevList []string, syncOpts SyncOptions, cmdOpts *cmd.Options) (*ToOCI, error) {
	u := ui.FromContextOrNoop(ctx)

	toOCI := &ToOCI{
		sync{
			base:     syncBase{},
			lfs:      syncLFS{},
			cache:    nil,
			syncOpts: syncOpts,
		},
		srcGitRemote,
		argRevList,
	}

	var err error
	toOCI.ociHelper, err = oci.NewOCIHelper(syncOpts.TmpDir, target, tag)
	if err != nil {
		return nil, fmt.Errorf("creating new ociHelper: %w", err)
	}

	toOCI.cmdHelper, err = cmd.NewHelper(ctx, syncOpts.TmpDir, cmdOpts)
	if err != nil {
		return nil, fmt.Errorf("creating new cmdHelper: %w", err)
	}

	if syncOpts.CacheDir != "" {
		if err := os.MkdirAll(syncOpts.CacheDir, 0777); err != nil {
			u.Infof("Unable to create cache directory, continuing without caching.")
			return toOCI, nil
		}

		toOCI.cache, err = cache.NewCache(ctx, syncOpts.CacheDir, cmdOpts)
		if err != nil {
			u.Infof("Unable to access git object cache, continuing without caching.")
			return toOCI, nil
		}
	}

	return toOCI, nil
}

// Cleanup cleans up any temporary files created during the ToOCI process.
func (t *ToOCI) Cleanup() error {
	err := t.sync.cleanup()
	if err != nil {
		return fmt.Errorf("cleaning up sync: %w", err)
	}

	return nil
}

// Run leverages git bundles to store a git repository in an OCI registry. The bundle is appended to the manifest identified
// by a tag in the target repository if a bundle is necessary.
//
// Not all calls to ToOCI result in a new bundle. Sometimes we only need to update the references, which is done in the manifest config.
func (t *ToOCI) Run(ctx context.Context) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)
	u := ui.FromContextOrNoop(ctx)

	// sometimes we want to update everything
	if len(t.argRevList) < 1 {
		log.InfoContext(ctx, "no specified list of references, resolving all remote references")
		_, fullRefs, err := t.remoteCommitsRefs()
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("resolving all remote references: %w", err)
		}

		// clean ref list, populate argRevList
		t.argRevList = make([]string, 0, len(fullRefs))
		reg := regexp.MustCompile("^refs/(tags|heads)/")
		for _, fullRef := range fullRefs {
			cleanRef := reg.ReplaceAllString(fullRef, "")
			t.argRevList = append(t.argRevList, cleanRef)
		}
	}

	// see where the current sync is at
	oldManDesc, err := t.FetchBaseManifestConfig(ctx)
	if err != nil && !errors.Is(err, errdef.ErrNotFound) {
		return ocispec.Descriptor{}, err
	}

	// try to cache, and recover if it fails
	if t.syncOpts.CacheDir != "" {
		log.DebugContext(ctx, "Utilizing git cache", "cacheDir", t.syncOpts.CacheDir)
		if err := t.cache.UpdateFromGit(t.srcGitRemote, t.cmdHelper.Options, t.argRevList...); err != nil {
			log.DebugContext(ctx, "Cache failed to update git objects", "error", err)
			u.Infof("Failed to update cache with git objects, continuing without caching...")
		}
	}

	log.InfoContext(ctx, "Cloning git repo", "repo", t.srcGitRemote, "cache", t.syncOpts.CacheDir)
	if err := t.cloneRemote(); err != nil { // if cache DNE, objs are cloned to intermediate repo
		return ocispec.Descriptor{}, fmt.Errorf("cloning git remote %s to %s: %w", t.srcGitRemote, t.cmdHelper.Dir(), err)
	}

	log.InfoContext(ctx, "Bundling changes")
	newBundlePath, err := t.bundleChanges(t.argRevList...)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("bundling changes: %w", err)
	}

	var newBundleDesc ocispec.Descriptor
	if newBundlePath != "" { // sometimes a bundle is not necessary
		newBundleDesc, err = t.addBundleToManifest(ctx, newBundlePath)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("adding new bundle to manifest: %w", err)
		}
	}

	// prep config for sending
	if err := t.updateBaseConfig(ctx); err != nil { // reference updates are always necessary
		return ocispec.Descriptor{}, fmt.Errorf("updating base manifest config: %w", err)
	}

	// when pushing to OCI we must push the base manifest before the lfs manifest.
	var newManDesc ocispec.Descriptor
	log.InfoContext(ctx, "updating base manifest")
	newManDesc, err = t.sendBaseSync(ctx, newBundleDesc)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	log.InfoContext(ctx, "Commit manifest updated", "digest", newManDesc.Digest)

	if t.cmdHelper.WithLFS { // we always set this to true, unless the git-lfs command was not found
		log.InfoContext(ctx, "updating LFS manifest")
		lfsManDesc, err := t.runLFS(ctx, oldManDesc, newManDesc)
		switch {
		case errors.Is(err, cmd.ErrLFSNotEnabled):
			log.InfoContext(ctx, "repository does not have LFS enabled")
			return newManDesc, nil
		case err != nil:
			return ocispec.Descriptor{}, fmt.Errorf("continuing to-oci with git LFS: %w", err)
		default:
			log.InfoContext(ctx, "LFS manifest updated", "digest", lfsManDesc.Digest)

		}
	} else {
		// even if we don't sync LFS files try to update the lfs manifest's subject, if it exists
		log.InfoContext(ctx, "Attempting to update LFS manifest's subject to new descriptor")
		if err := t.updateLFSManSubject(ctx, oldManDesc, newManDesc); err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("updating LFS manifest's subject: %w", err)
		}
	}

	return newManDesc, nil
}

// cloneRemote clones a remote source git repository to the intermediate directory. The clone
// always attempts to reference the cache. A non-existent cache is not fatal.
func (t *ToOCI) cloneRemote() error {
	if t.srcGitRemote == "" {
		return fmt.Errorf("no source git remote specified, unable to clone")
	}

	return t.cmdHelper.CloneWithShared(t.srcGitRemote, t.syncOpts.CacheDir)
}

// updateBaseConfig updates the commit manifest's config. It should be called after
// appending a new layer to the existing manifest.Layers[].
func (t *ToOCI) updateBaseConfig(ctx context.Context) error {
	log := logger.FromContext(ctx)

	layerResolver := t.sortRefsByLayer()

	log.InfoContext(ctx, "Resolving new references and commits")
	newCommits, fullRefs, err := t.localCommitsRefs(t.argRevList...) // fullRefs[i] corresponds to newCommits[i]
	if err != nil {
		return fmt.Errorf("resolving references and new commits: %w", err)
	}

	// Update config refs
	for i, fullRef := range fullRefs {
		trimmedRef := filepath.Base(fullRef)
		newCommit := newCommits[i]

		switch {
		case strings.HasPrefix(fullRef, cmd.TagRefPrefix):
			refInfo := t.base.config.Refs.Tags[trimmedRef]
			if newCommit != refInfo.Commit {
				oldestLayer, err := t.resolveLayer(layerResolver, newCommit)
				if err != nil {
					return fmt.Errorf("resolving layer containing commit '%s': %w", newCommit, err)
				}
				refInfo.Layer = oldestLayer
				refInfo.Commit = newCommit
				t.base.config.Refs.Tags[trimmedRef] = refInfo
			}

		case strings.HasPrefix(fullRef, cmd.HeadRefPrefix):
			refInfo := t.base.config.Refs.Heads[trimmedRef]
			if newCommit != refInfo.Commit {
				oldestLayer, err := t.resolveLayer(layerResolver, newCommit)
				if err != nil {
					return fmt.Errorf("resolving layer containing commit '%s': %w", newCommit, err)
				}
				refInfo.Layer = oldestLayer
				refInfo.Commit = newCommit
				t.base.config.Refs.Heads[trimmedRef] = refInfo
			}

		default:
			// we filter out other references
			log.InfoContext(ctx, "skipping unsupported reference", "ref", fullRef)
		}
	}

	return nil
}

// sortRefsByLayer organizes the refs in the current config by layer,
// returning a map of layer digests to a slice of commits contained in that layer.
func (t *ToOCI) sortRefsByLayer() map[digest.Digest][]Commit {

	layerResolver := make(map[digest.Digest][]Commit) // layer digest : []commits
	for _, info := range t.base.config.Refs.Heads {
		layerResolver[info.Layer] = append(layerResolver[info.Layer], info.Commit)
	}
	for _, info := range t.base.config.Refs.Tags {
		layerResolver[info.Layer] = append(layerResolver[info.Layer], info.Commit)
	}

	return layerResolver
}

// resolveLayer returns the oldest layer in the current manifest containing a commit. Should be called
// after layer updates.
func (t *ToOCI) resolveLayer(layerResolver map[digest.Digest][]Commit, targetCommit Commit) (digest.Digest, error) {
	for i, layer := range t.base.manifest.Layers {
		if i == len(t.base.manifest.Layers)-1 { // we have reached the final layer, so it must be here.
			return layer.Digest, nil
		}
		for _, commit := range layerResolver[layer.Digest] {
			err := t.cmdHelper.MergeBase("--is-ancestor", string(targetCommit), string(commit)) // is the targetCommit an ancestor of commit?
			switch {
			case errors.Is(err, cmd.ErrNotAncestor):
				continue
			case err != nil:
				return "", fmt.Errorf("checking if commit %s is an ancestor of commit %s: %w", targetCommit, commit, err)
			default:
				return layer.Digest, nil
			}
		}
	}

	// default to the base layer
	return t.base.manifest.Layers[0].Digest, nil
}

// addBundleToManifest prepares the shared filestore with the new bundle layer
// as well as adds it to the manifest layers.
func (t *ToOCI) addBundleToManifest(ctx context.Context, newBundlePath string) (ocispec.Descriptor, error) {

	newBundleDesc, err := t.ociHelper.FStore.Add(ctx, filepath.Base(newBundlePath), MediaTypeBundleLayer, newBundlePath)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("adding bundle to filestore: %w", err)
	}
	t.base.manifest.Layers = append(t.base.manifest.Layers, newBundleDesc)

	return newBundleDesc, nil
}

// sendBaseSync pushes all components of a base sync to the target repository. Components
// include the updated manifest, config, and a new bundle if necessary.
//
// If newBundlePath is an empty string this func will skip pushing a new bundle.
// Sending/updating a manifest does not update the local copy in the ToOCI structure.
func (t *ToOCI) sendBaseSync(ctx context.Context, newBundleDesc ocispec.Descriptor) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	if newBundleDesc.Digest != "" {
		log.DebugContext(ctx, "Pushing new bundle layer", "digest", newBundleDesc.Digest)
		if err := oras.CopyGraph(ctx, t.ociHelper.FStore, t.ociHelper.Target, newBundleDesc, oras.DefaultCopyGraphOptions); err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("copying bundle layer to target repository: %w", err)
		}
	} else {
		log.DebugContext(ctx, "Skipping bundle push")
	}

	configBytes, err := json.Marshal(t.base.config)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("encoding base manifest config")
	}
	log.DebugContext(ctx, "Pushing base config")
	configDesc, err := oras.PushBytes(ctx, t.ociHelper.Target, MediaTypeSyncConfig, configBytes)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing base config to repo: %w", err)
	}

	log.DebugContext(ctx, "Pushing base manifest")
	manOpts := oras.PackManifestOptions{
		Layers:              t.base.manifest.Layers, // if a new bundle was made, it was already added to the manifest
		ConfigDescriptor:    &configDesc,
		ManifestAnnotations: map[string]string{ocispec.AnnotationCreated: "1970-01-01T00:00:00Z", AnnotationDTVersion: t.syncOpts.DTVersion}, // POSIX epoch
	}

	manDesc, err := oras.PackManifest(ctx, t.ociHelper.Target, oras.PackManifestVersion1_1, ArtifactTypeSyncManifest, manOpts)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("packing and pushing base manifest: %w", err)
	}

	log.DebugContext(ctx, "Tagging base manifest")
	err = t.ociHelper.Target.Tag(ctx, manDesc, t.ociHelper.Tag)
	if err != nil {
		return manDesc, fmt.Errorf("tagging base manifest: %w", err)
	}

	return manDesc, nil
}

// bundleChanges creates a bundle of changes from the prior sync to the commits referenced by argRevList, returning
// the path of the bundle. An empty bundle path alongside a nil error indicates that a bundle of objects is not needed
// but reference updates should still occur.
func (t *ToOCI) bundleChanges(argRevList ...string) (string, error) {

	// make new bundle rev-list
	excludeCommits := make([]string, 0, len(t.base.config.Refs.Tags)+len(t.base.config.Refs.Heads)+len(argRevList))
	for _, refInfo := range t.base.config.Refs.Tags {
		excludeCommits = append(excludeCommits, "^"+string(refInfo.Commit))
	}
	for _, refInfo := range t.base.config.Refs.Heads {
		excludeCommits = append(excludeCommits, "^"+string(refInfo.Commit))
	}
	revList := append(excludeCommits, argRevList...) //nolint
	newBundlePath := filepath.Join(t.ociHelper.FStorePath, "changes"+fmt.Sprintf("%d", len(t.base.manifest.Layers)+1)+".bundle")

	err := t.cmdHelper.BundleCreate(newBundlePath, revList)
EmptyBundleCheck:
	switch {
	case errors.Is(err, cmd.ErrEmptyBundle):
		newBundlePath = "" // a "" bundle path indicates we're only updating refs

		newCommits, newRefs, err := t.localCommitsRefs(argRevList...)
		if err != nil {
			return "", fmt.Errorf("resolving references and new commits: %w", err)
		}

		// check to see if we're updating a reference
		for i, fullRef := range newRefs {

			switch {
			case strings.HasPrefix(fullRef, cmd.TagRefPrefix):
				oldTaggedInfo, inTags := t.base.config.Refs.Tags[strings.TrimPrefix(fullRef, cmd.TagRefPrefix)]
				if !inTags || newCommits[i] != oldTaggedInfo.Commit {
					break EmptyBundleCheck
				}
			case strings.HasPrefix(fullRef, cmd.HeadRefPrefix):
				oldHeadInfo, inHeads := t.base.config.Refs.Heads[strings.TrimPrefix(fullRef, cmd.HeadRefPrefix)]
				if !inHeads || newCommits[i] != oldHeadInfo.Commit {
					break EmptyBundleCheck
				}
			}
		}

		return "", nil // update not discovered

	case err != nil:
		return "", err
	}
	return newBundlePath, nil
}

// localCommitsRefs returns the local references and the commits they reference
// split into two slices, with indicies matching the pairs. If argRevList is empty
// all references will be returned.
func (t *ToOCI) localCommitsRefs(argRevList ...string) ([]Commit, []string, error) {
	commitStr, fullRefs, err := t.cmdHelper.LocalCommitsRefs(argRevList...)
	if err != nil {
		return nil, nil, err
	}

	commits := make([]Commit, 0, len(commitStr))
	for _, commit := range commitStr {
		commits = append(commits, Commit(commit))
	}

	return commits, fullRefs, nil
}

// remoteCommitsRefs returns the remote references and the commits they reference
// split into two slices, with indicies matching the pairs. If argRevList is empty
// all references will be returned.
func (t *ToOCI) remoteCommitsRefs(argRevList ...string) ([]Commit, []string, error) {
	commitStr, fullRefs, err := t.cmdHelper.RemoteCommitsRefs(t.srcGitRemote, argRevList...)
	if err != nil {
		return nil, nil, err
	}

	commits := make([]Commit, 0, len(commitStr))
	for _, commit := range commitStr {
		commits = append(commits, Commit(commit))
	}

	return commits, fullRefs, nil
}
