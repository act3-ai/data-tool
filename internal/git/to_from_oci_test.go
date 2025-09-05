package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/errdef"

	"github.com/act3-ai/data-tool/internal/git/cmd"
	"github.com/act3-ai/data-tool/internal/git/oci"
	"github.com/act3-ai/go-common/pkg/logger"
	tlog "github.com/act3-ai/go-common/pkg/test"
)

const dtVersion = "devel-test"
const syncTag = "sync"

type args struct {
	argRevList          []string // references used by ToOCI
	expectedTagList     []string // expected results of tag refs in manifest config
	expectedHeadList    []string // expected results of head refs in manifest config
	expectedBundleCount int      // expected number of bundle layers in manifest
	expectedRebuildRefs []string // expected updates returned by FromOCI
	tag                 string   // tag of the manifest in the OCI registry
}

type test struct {
	name    string
	args    args
	wantErr bool
}

// Use Cases
var tests = []test{
	// NOTES:
	// - The order of these tests is crucial.
	// - All tests validate the OCI manifest and config created by ToOCI.
	// - All tests validate the resulting repository created by FromOCI.

	// Base Layer Tests:
	// - Creating a valid bundle of a commit history up to a reference.
	// - Creating a bundle based on a tag reference.
	{
		name: "Base Layer",
		args: args{
			argRevList:          []string{"v1.0.1"},
			expectedTagList:     []string{"v1.0.1"},
			expectedHeadList:    []string{},
			expectedBundleCount: 1,
			expectedRebuildRefs: []string{"refs/tags/v1.0.1"},
			tag:                 syncTag,
		},
		wantErr: false,
	},

	// Add Tag Ref to New Commit Tests:
	// - Appending a thin bundle to an existing manifest.
	// - Adding a new tag reference to manifest config.
	// - Only one tag reference is updated by FromOCI.
	{
		name: "Add Tag Ref to New Commit",
		args: args{
			argRevList:          []string{"v1.0.2"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2"},
			expectedHeadList:    []string{},
			expectedBundleCount: 2,
			expectedRebuildRefs: []string{"refs/tags/v1.0.2"},
			tag:                 syncTag,
		},
		wantErr: false,
	},

	// Add Head Ref to New Commit Tests:
	// - Appending a thin bundle to an existing manifest.
	// - Adding a new head reference to manifest config.
	// - Only one head reference is updated by FromOCI.
	{
		name: "Add Head Ref to New Commit",
		args: args{
			argRevList:          []string{"Feature2"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2"},
			expectedHeadList:    []string{"Feature2"},
			expectedBundleCount: 3,
			expectedRebuildRefs: []string{"refs/heads/Feature2"},
			tag:                 syncTag,
		},
		wantErr: false,
	},

	// Add Tag Ref to Existing Branch Head Ref Tests:
	// - No additional bundle is created.
	// - Adding a tag reference to a commit that's already included in the manifest via head reference.
	{
		name: "Add Tag Ref to Existing Branch Head Ref",
		args: args{
			argRevList:          []string{"v1.0.3"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2", "v1.0.3"},
			expectedHeadList:    []string{"Feature2"},
			expectedBundleCount: 3,
			expectedRebuildRefs: []string{"refs/tags/v1.0.3"},
			tag:                 syncTag,
		},
		wantErr: false,
	},

	// Add Branch Head Ref to Existing Tag Ref Tests:
	// - No additional bundle is created.
	// - Adding a head reference to a commit that's already included in the manifest via tag reference.
	{
		name: "Add Branch Head Ref to Existing Tag Ref",
		args: args{
			argRevList:          []string{"Feature1"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2", "v1.0.3"},
			expectedHeadList:    []string{"Feature2", "Feature1"},
			expectedBundleCount: 3,
			expectedRebuildRefs: []string{"refs/heads/Feature1"},
			tag:                 syncTag,
		},
		wantErr: false,
	},

	// Add Tag Ref to New Commit Tests:
	// - Mostly for prepparing for the update tests.
	{
		name: "Add Tag Ref to New Commit",
		args: args{
			argRevList:          []string{"v1.2.0"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2", "v1.0.3", "v1.2.0"},
			expectedHeadList:    []string{"Feature2", "Feature1"},
			expectedBundleCount: 4,
			expectedRebuildRefs: []string{"refs/tags/v1.2.0"},
			tag:                 syncTag,
		},
		wantErr: false,
	},
}

// Expected Error Cases
var errorTests = []test{
	// Unnecessary Tag Ref Sync Tests"
	// - An expected failure for a tag reference that does not need to be updated.
	{
		name: "Unnecessary Tag Ref Sync",
		args: args{
			argRevList:          []string{"v1.0.1"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2", "v1.0.3", "v1.2.0"},
			expectedHeadList:    []string{"Feature2", "Feature1"},
			expectedBundleCount: 4,
			tag:                 syncTag,
		},
		wantErr: true,
	},

	//  Unnecessary Head Ref Sync Tests:
	// - An expected failure for a head reference that does not need to be updated.
	{
		name: "Unnecessary Head Ref Sync",
		args: args{
			argRevList:          []string{"Feature2"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2", "v1.0.3", "v1.2.0"},
			expectedHeadList:    []string{"Feature2", "Feature1"},
			expectedBundleCount: 4,
			tag:                 syncTag,
		},
		wantErr: true,
	},
}

// Update Use Cases - requires an update to be made to an existing reference.
var updateTests = []test{
	// Update Branch Head Ref Tests:
	// - Updating a head reference that already exists in the manifest config to a child commit.
	{
		name: "Update Branch Head Ref",
		args: args{
			argRevList:          []string{"Feature2"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2", "v1.0.3", "v1.2.0"},
			expectedHeadList:    []string{"Feature2", "Feature1"},
			expectedBundleCount: 5,
			expectedRebuildRefs: []string{"refs/heads/Feature2"},
			tag:                 syncTag,
		},
		wantErr: false,
	},
}

// Force Update Use Cases - requires an update to an existing tag reference.
var forceTests = []test{
	// Update Tag Ref Tests:
	// - Updating a tag reference that already exists in the manifest config to a child commit.
	{
		name: "Update Tag Ref",
		args: args{
			argRevList:          []string{"v1.2.0"},
			expectedTagList:     []string{"v1.0.1", "v1.0.2", "v1.0.3", "v1.2.0"},
			expectedHeadList:    []string{"Feature2", "Feature1"},
			expectedBundleCount: 6,
			expectedRebuildRefs: []string{"refs/tags/v1.2.0"},
			tag:                 syncTag,
		},
		wantErr: false,
	},
}

func Test_ToFromOCI(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	gitVersion, err := cmd.CheckGitVersion(ctx, "")
	if err != nil {
		t.Errorf("validating git version: %v", err)
	}
	t.Logf("Testing with git version %s", gitVersion)

	// Setup git repo
	t.Log("Setting up testing git repository")
	srcGitRemote := t.TempDir() // sync target
	srcRepoCH, err := cmd.NewHelper(ctx, srcGitRemote, &cmd.Options{})
	if err != nil {
		t.Fatalf("creating source repo command helper: %v", err)
	}
	if err := createTestRepo(ctx, srcRepoCH); err != nil {
		t.Fatalf("creating test repository: %v", err)
	}

	t.Log("Setting up rebuild git repository")
	dstGitRemote := t.TempDir() // rebuild target
	destRepoCH, err := cmd.NewHelper(ctx, dstGitRemote, &cmd.Options{})
	if err != nil {
		t.Fatalf("creating destination repo command helper: %v", err)
	}
	err = destRepoCH.Init(ctx)
	if err != nil {
		t.Fatalf("setting up git rebuild dir: %v", err)
	}

	target := memory.New() // oci target

	// Use Cases
	for i, tt := range tests {
		var existingDesc ocispec.Descriptor
		if i != 0 {
			existingDesc, err = target.Resolve(ctx, tt.args.tag)
			if err != nil {
				t.Fatalf("resolving base manifest descriptor: error = %s", err)
			}
		}
		ToFromOCITestFunc(ctx, tt, target, existingDesc, srcRepoCH, destRepoCH, &cmd.Options{}, true)(t)
	}

	// Error Cases
	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			toOCIFStorePath := t.TempDir()
			toOCIFStore, err := file.New(toOCIFStorePath)
			if err != nil {
				t.Fatalf("initializing to ocicache file store: %v", err)
			}
			defer toOCIFStore.Close()

			existingDesc, err := target.Resolve(ctx, tt.args.tag)
			if err != nil {
				t.Fatalf("resolving base manifest descriptor: error = %s", err)
			}

			syncOpts := SyncOptions{UserAgent: dtVersion, IntermediateDir: toOCIFStorePath, IntermediateStore: toOCIFStore}
			toOCITester, err := NewToOCI(ctx, target, existingDesc, srcGitRemote, tt.args.argRevList, syncOpts, &cmd.Options{})
			if err != nil {
				t.Errorf("creating ToOCI: %v", err)
			}
			defer toOCITester.Cleanup() //nolint

			_, err = toOCITester.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToOCI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := toOCITester.Cleanup(); err != nil {
				t.Errorf("cleaning up toOCITester handler: %v", err)
			}
		})
	}

	// Use Cases - that require an update to an existing reference
	t.Log("Updating up testing git repository")
	if err := updateTestRepo(ctx, srcRepoCH); err != nil {
		t.Fatalf("updating test repository: %v", err)
	}

	for _, tt := range updateTests {
		existingDesc, err := target.Resolve(ctx, tt.args.tag)
		if err != nil {
			t.Fatalf("resolving base manifest descriptor: error = %s", err)
		}
		ToFromOCITestFunc(ctx, tt, target, existingDesc, srcRepoCH, destRepoCH, &cmd.Options{}, true)(t)
	}

	for _, tt := range forceTests {
		existingDesc, err := target.Resolve(ctx, tt.args.tag)
		if err != nil {
			t.Fatalf("resolving base manifest descriptor: error = %s", err)
		}
		ToFromOCITestFunc(ctx, tt, target, existingDesc, srcRepoCH, destRepoCH, &cmd.Options{GitOptions: cmd.GitOptions{Force: true}}, true)(t)
	}
}

func Test_ToFromOCIRewrite(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	gitVersion, err := cmd.CheckGitVersion(ctx, "")
	if err != nil {
		t.Errorf("validating git version: %v", err)
	}
	t.Logf("Testing with git version %s", gitVersion)

	// Setup source git repo
	t.Log("Setting up testing git repository")
	srcGitRemote := t.TempDir()
	srcRepoCH, err := cmd.NewHelper(ctx, srcGitRemote, &cmd.Options{})
	if err != nil {
		t.Fatalf("creating source repo command helper: %v", err)
	}
	if err := createTestRepoRewrite(ctx, srcRepoCH); err != nil {
		t.Fatalf("creating test repository: %v", err)
	}

	// Setup destination git repo
	t.Log("Setting up rebuild git repository")
	dstGitRemote := t.TempDir()
	destRepoCH, err := cmd.NewHelper(ctx, dstGitRemote, &cmd.Options{})
	if err != nil {
		t.Fatalf("creating destination repo command helper: %v", err)
	}
	err = destRepoCH.Init(ctx, "--bare")
	if err != nil {
		t.Fatalf("setting up git rebuild dir: %v", err)
	}

	target := memory.New()

	// Base case, setup some history to rewrite
	baseTest := test{
		name: "Base",
		args: args{
			argRevList:          []string{"main", "Feature1"},
			expectedTagList:     []string{},
			expectedHeadList:    []string{"main", "Feature1"},
			expectedBundleCount: 1,
			expectedRebuildRefs: []string{"refs/heads/main", "refs/heads/Feature1"},
			tag:                 syncTag,
		},
		wantErr: false,
	}
	t.Run("Rewrite", ToFromOCITestFunc(ctx, baseTest, target, ocispec.Descriptor{}, srcRepoCH, destRepoCH, &cmd.Options{}, true))

	// Diverge branch history
	if err := updateTestRepoDiverge(ctx, srcRepoCH); err != nil {
		t.Fatalf("diverging git history: %v", err)
	}
	if err := deleteDanglingCommits(ctx, srcRepoCH); err != nil {
		t.Fatalf("cleaning up diverged history: %v", err)
	}

	divergeTest := test{
		name: "Diverge History",
		args: args{
			argRevList:          []string{"main", "Feature1"},
			expectedTagList:     []string{},
			expectedHeadList:    []string{"main", "Feature1"},
			expectedBundleCount: 2,
			expectedRebuildRefs: []string{"refs/heads/main", "refs/heads/Feature1"},
			tag:                 syncTag,
		},
		wantErr: false,
	}

	existingDesc, err := target.Resolve(ctx, divergeTest.args.tag)
	if err != nil {
		t.Fatalf("resolving base manifest descriptor: error = %s", err)
	}

	// we can check bundles here due to the implementation of bundle validation
	// taking the newest reference in any bundle.
	t.Run("Rewrite", ToFromOCITestFunc(ctx, divergeTest, target, existingDesc, srcRepoCH, destRepoCH, &cmd.Options{GitOptions: cmd.GitOptions{Force: true}}, true))

	// save a bad commit
	refCommits, err := srcRepoCH.ShowRefs(ctx, "Feature1")
	if err != nil {
		t.Fatalf("resolving reference and commit: %v", err)
	}
	if len(refCommits) != 1 {
		t.Fatalf("expected 1 reference resolved, got %d", len(refCommits))
	}
	split := strings.Fields(refCommits[0])
	badCommit := split[0]

	// save this version of the sync manifest
	// HACK: rework with commented out code once dup config push is fixed
	// hackTarget := memory.New()
	// syncTagDup := syncTag + "-dup"
	// _, err = oras.Copy(ctx, target, syncTag, hackTarget, syncTagDup, oras.CopyOptions{})
	// if err != nil {
	// 	t.Fatalf("copying sync manifest: %v", err)
	// }

	// desc, err := target.Resolve(ctx, syncTag)
	// if err != nil {
	// 	t.Fatalf("resolving sync manifest descriptor: %v", err)
	// }
	// if err := target.Tag(ctx, desc, syncTagDup); err != nil {
	// 	t.Fatalf("tagging manifest again: %v", err)
	// }

	if err := updateTestRepoRevert(ctx, srcRepoCH); err != nil {
		t.Fatalf("reverting git history: %v", err)
	}
	if err := deleteDanglingCommits(ctx, srcRepoCH); err != nil {
		t.Fatalf("cleaning up reverted history: %v", err)
	}

	// Reset branch history
	resetTest := test{name: "Revert History",
		args: args{
			argRevList:          []string{"main", "Feature1", badCommit},
			expectedTagList:     []string{},
			expectedHeadList:    []string{"main", "Feature1"},
			expectedBundleCount: 2,
			expectedRebuildRefs: []string{"refs/heads/Feature1"},
			tag:                 syncTag,
		},
		wantErr: false,
	}

	existingDesc, err = target.Resolve(ctx, resetTest.args.tag)
	if err != nil {
		t.Fatalf("resolving base manifest descriptor: error = %s", err)
	}
	// don't check the bundles, the newest reference is overridden by the config.
	t.Run("Rewrite", ToFromOCITestFunc(ctx, resetTest, target, existingDesc, srcRepoCH, destRepoCH, &cmd.Options{GitOptions: cmd.GitOptions{Force: true}}, false))

	// Clean sync - no oci artifact
	hackTarget := memory.New()
	cleanSyncTest := test{name: "Clean Sync",
		args: args{
			argRevList:          []string{"main", "Feature1", badCommit},
			expectedTagList:     []string{},
			expectedHeadList:    []string{"main", "Feature1"},
			expectedBundleCount: 1,
			expectedRebuildRefs: []string{"refs/heads/main", "refs/heads/Feature1"},
			tag:                 syncTag,
		},
		wantErr: false,
	}

	// HACK: Our validation function checks the destination based off of the update list returned, not what's actually in the destination
	dstGitRemote2 := t.TempDir()
	destRepoCH2, err := cmd.NewHelper(ctx, dstGitRemote2, &cmd.Options{})
	if err != nil {
		t.Fatalf("creating destination repo command helper: %v", err)
	}
	err = destRepoCH2.Init(ctx, "--bare")
	if err != nil {
		t.Fatalf("setting up git rebuild dir: %v", err)
	}
	// don't check the bundles, the newest reference is overridden by the config.
	t.Run("Rewrite", ToFromOCITestFunc(ctx, cleanSyncTest, hackTarget, ocispec.Descriptor{}, srcRepoCH, destRepoCH2, &cmd.Options{GitOptions: cmd.GitOptions{Force: true}}, false))
}

// test an alternate git executable provided via path.
func Test_AltGitExec(t *testing.T) {
	ctx := context.Background()

	gitDir := t.TempDir()

	whichCmd := exec.CommandContext(t.Context(), "which", "git")
	out, err := whichCmd.Output()
	if err != nil {
		t.Fatalf("executing 'which git' command: %v", err)
	}

	cmdOpts := cmd.Options{GitOptions: cmd.GitOptions{AltGitExec: strings.TrimSpace(string(out))}}
	_, err = cmd.NewHelper(ctx, gitDir, &cmdOpts) // calls alt exec when validating compatibility and initializing intermediate repo
	if err != nil {
		t.Errorf("creating cmd helper: %v", err)
	}
}

// test an alternate git executable provided via path.
func Test_AltGitLFSExec(t *testing.T) {
	ctx := context.Background()

	gitDir := t.TempDir()

	whichCmd := exec.CommandContext(t.Context(), "which", "git-lfs")
	out, err := whichCmd.Output()
	if err != nil {
		t.Fatalf("executing 'which git-lfs' command: %v", err)
	}

	cmdOpts := cmd.Options{LFSOptions: &cmd.LFSOptions{AltLFSExec: strings.TrimSpace(string(out))}}
	_, err = cmd.NewHelper(ctx, gitDir, &cmdOpts) // calls alt exec when validating compatibility and initializing intermediate repo
	if err != nil {
		t.Errorf("creating cmd helper: %v", err)
	}
}

// test an attempted rebuild of a manifest that does not exist.
func Test_FromOCINonExistantManifest(t *testing.T) {
	ctx := context.Background()
	target := memory.New()
	nonExistantDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.Digest("sha256:e9d3f25129237340dfb34b98aff9c5adf1ed2713264729fa9f470a5233be8a5d"),
		Size:      636,
	}

	fs, err := file.New(t.TempDir())
	if err != nil {
		t.Fatalf("initializing file store: %v", err)
	}
	defer fs.Close()

	syncOpts := SyncOptions{IntermediateDir: t.TempDir(), IntermediateStore: fs}
	cmdOpts := &cmd.Options{
		GitOptions: cmd.GitOptions{Force: true, AltGitExec: ""},
		LFSOptions: &cmd.LFSOptions{},
	}
	fromOCITester, err := NewFromOCI(ctx, target, nonExistantDesc, "", syncOpts, cmdOpts)
	if err != nil {
		t.Errorf("creating FromOCI: %v", err)
	}
	defer fromOCITester.Cleanup() //nolint

	_, err = fromOCITester.Run(ctx)
	if !errors.Is(err, errdef.ErrNotFound) {
		t.Errorf("FromOCI() error = %v, wantErr = %v", err, errdef.ErrNotFound)
	}

	if err := fromOCITester.Cleanup(); err != nil {
		t.Errorf("cleaning up fromOCITester handler: %v", err)
	}
}

//nolint:gocognit
func ToFromOCITestFunc(ctx context.Context, tt test, target oras.GraphTarget,
	existingDesc ocispec.Descriptor, srcRepoCH, dstRepoCH *cmd.Helper,
	cmdOpts *cmd.Options, checkBundles bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run(tt.name, func(t *testing.T) {
			// want - but as maps to validate we got against what we want across bundles
			expectedTags := make(map[string]bool, 0)
			for _, tag := range tt.args.expectedTagList {
				expectedTags[tag] = false
			}

			expectedHeads := make(map[string]bool, 0)
			for _, head := range tt.args.expectedHeadList {
				expectedHeads[head] = false
			}

			var newDesc ocispec.Descriptor
			t.Run(tt.name+": ToOCI", func(t *testing.T) {
				toOCIFStorePath := t.TempDir()
				toOCIFStore, err := file.New(toOCIFStorePath)
				if err != nil {
					t.Fatalf("initializing to ocicache file store: %v", err)
				}
				defer toOCIFStore.Close()

				syncOpts := SyncOptions{
					UserAgent:         dtVersion,
					IntermediateDir:   toOCIFStorePath,
					IntermediateStore: toOCIFStore,
				}
				toOCITester, err := NewToOCI(ctx, target, existingDesc, srcRepoCH.Dir(), tt.args.argRevList, syncOpts, cmdOpts)
				if err != nil {
					t.Errorf("creating ToOCI: %v", err)
				}
				defer toOCITester.Cleanup() //nolint

				newDesc, err = toOCITester.Run(ctx)
				if err != nil {
					t.Fatalf("ToOCI() error = %v, wantErr %v", err, tt.wantErr)
				}

				if newDesc.ArtifactType != oci.ArtifactTypeSyncManifest {
					t.Errorf("unexpected artifact type in descriptor of commit manifest, got '%s' want '%s'", newDesc.ArtifactType, oci.ArtifactTypeSyncManifest)
				}

				successors, err := content.Successors(ctx, target, newDesc)
				if err != nil {
					t.Fatalf("fetching successors for sync manifest, err = %v", err)
				}
				if len(successors) != tt.args.expectedBundleCount+1 { // +1 for config
					t.Errorf("expected %d successors, got %d successors", tt.args.expectedBundleCount+1, len(successors))
				}

				pulledBundles := t.TempDir()
				bundleRebuild := t.TempDir()

				t.Log("Validating sync config and bundles")
				err = validateSync(ctx, pulledBundles, bundleRebuild, successors, expectedTags, expectedHeads, checkBundles, target, dtVersion)
				if err != nil {
					t.Fatalf("sync failed verification, args = %+v, err = %v", tt.args, err)
				}
				t.Logf("Information in bundle(s) matches config")

				if err := toOCITester.Cleanup(); err != nil {
					t.Errorf("cleaning up toOCITester handler: %v", err)
				}

				if err := target.Tag(ctx, newDesc, tt.args.tag); err != nil {
					t.Fatalf("tagging new sync manifest: error = %s", err)
				}
			})

			expectedRebuildMap := make(map[string]bool, len(tt.args.expectedRebuildRefs))
			for _, ref := range tt.args.expectedRebuildRefs {
				expectedRebuildMap[ref] = false
			}

			// test rebuild
			t.Run(tt.name+": FromOCI", func(t *testing.T) {
				fromOCIFStorePath := t.TempDir()
				fromOCIFStore, err := file.New(fromOCIFStorePath)
				if err != nil {
					t.Fatalf("initializing from oci cache file store: %v", err)
				}
				defer fromOCIFStore.Close()

				syncOpts := SyncOptions{
					UserAgent:         dtVersion,
					IntermediateDir:   fromOCIFStorePath,
					IntermediateStore: fromOCIFStore,
				}
				fromOCITester, err := NewFromOCI(ctx, target, newDesc, dstRepoCH.Dir(), syncOpts, cmdOpts)
				if err != nil {
					t.Errorf("creating FromOCI: %v", err)
				}
				defer fromOCITester.Cleanup() //nolint

				updatedRefs, err := fromOCITester.Run(ctx)
				if err != nil {
					t.Fatalf("from oci: %v", err)
				}
				t.Logf("updated refs: %s", updatedRefs)

				err = validateRebuild(expectedRebuildMap, updatedRefs)
				if err != nil {
					t.Fatalf("rebuild failed verification, args = %+v, err = %v", tt.args, err)
				}
				t.Logf("Destination repository result matches expectation")

				if err := fromOCITester.Cleanup(); err != nil {
					t.Errorf("cleaning up fromOCITester handler: %v", err)
				}
			})
		})
	}

}

// * * Validation Functions * * //

// validateSync validates the successors of a sync manifest, i.e. it's config and bundle layers.
func validateSync(ctx context.Context, pulledBundlesDir, bundleRebuild string,
	successors []ocispec.Descriptor, expectedTags, expectedHeads map[string]bool,
	checkBundles bool, target oras.GraphTarget, dtVersion string) error {

	gc, err := cmd.NewHelper(ctx, bundleRebuild, &cmd.Options{})
	if err != nil {
		return fmt.Errorf("creating validation command helper: %w", err)
	}

	if err := gc.Init(ctx, "--bare"); err != nil {
		return fmt.Errorf("initializing bundle rebuild repository: %w", err)
	}

	// ensure we get the config first, so we can compare it against bundles later.
	// This may not be necessary, but added to be safe
	if successors[0].MediaType != oci.MediaTypeSyncConfig {
		for i, desc := range successors {
			if desc.MediaType == oci.MediaTypeSyncConfig {
				successors[0], successors[i] = successors[i], successors[0]
				break
			}
		}
	}
	syncErrs := make([]error, 0)

	config, allBundleTags, allBundleHeads, err := prepSyncValidation(ctx, gc, target, successors, pulledBundlesDir)
	syncErrs = append(syncErrs, err)

	// validate if config has expected tags and heads
	err = validateConfig(*config, dtVersion, expectedTags, expectedHeads)
	if err != nil {
		syncErrs = append(syncErrs, fmt.Errorf("invalid config: %w", err))
	}

	// reset expected maps
	for tag := range expectedTags {
		expectedTags[tag] = false
	}
	for head := range expectedHeads {
		expectedHeads[head] = false
	}

	// validate if bundles have expected tags and heads, and the referenced commits match the config
	if checkBundles {
		if err = validateBundleRefs(ctx, *config, allBundleTags, allBundleHeads, expectedTags, expectedHeads); err != nil {
			syncErrs = append(syncErrs, fmt.Errorf("invalid references in bundle(s): %w", err))
		}
	}

	return errors.Join(syncErrs...)
}

func prepSyncValidation(ctx context.Context, ch *cmd.Helper, target oras.GraphTarget,
	successors []ocispec.Descriptor, pathToBundles string) (*oci.Config, map[string]cmd.Commit, map[string]cmd.Commit, error) {

	var config = &oci.Config{}
	allBundleTags := make(map[string]cmd.Commit, 0) // ref:commit key:val pair
	allBundleHeads := make(map[string]cmd.Commit, 0)
	syncErrs := make([]error, 0)

	// verify the config and bundle
	fstore, err := file.New(pathToBundles)
	if err != nil {
		return nil, allBundleTags, allBundleHeads, fmt.Errorf("initializing filestore: %w", err)
	}
	defer fstore.Close()

	// prepare for validation
	for _, desc := range successors {
		switch desc.MediaType {
		case oci.MediaTypeSyncConfig:

			// fetch & unmarshal config, fail instead of collecting errors here as we can't validate much without a config
			cfgReader, err := target.Fetch(ctx, desc)
			if err != nil {
				return nil, allBundleTags, allBundleHeads, fmt.Errorf("fetching config from target: %w", err)
			}
			cfgBytes, err := io.ReadAll(cfgReader)
			if err != nil {
				return nil, allBundleTags, allBundleHeads, fmt.Errorf("reading config: %w", err)
			}
			cfgReader.Close()

			err = json.Unmarshal(cfgBytes, config)
			if err != nil {
				return nil, allBundleTags, allBundleHeads, fmt.Errorf("unmarshaling config: %w", err)
			}

		case oci.MediaTypeBundleLayer:

			// fetch the bundle
			bundleName := desc.Annotations[ocispec.AnnotationTitle]
			targetBundlePath := filepath.Join(pathToBundles, bundleName)

			err = oras.CopyGraph(ctx, target, fstore, desc, oras.DefaultCopyGraphOptions)
			if err != nil {
				return config, allBundleTags, allBundleHeads, fmt.Errorf("fetching layer %s bytes: %w", desc.Digest, err)
			}

			// get and organize the bundle's references
			bundleRefs, err := listHeads(ctx, ch, targetBundlePath)
			if err != nil {
				return config, allBundleTags, allBundleHeads, fmt.Errorf("listing references in fetched bundle: %w", err)
			}

			for _, entry := range bundleRefs {
				split := strings.Fields(entry)
				commit := cmd.Commit(split[0])
				fullBundleRef := split[1]

				// if a ref already exists, it will be overwritten by the latest bundle - which is expected behavior
				switch {
				case strings.HasPrefix(fullBundleRef, cmd.TagRefPrefix):
					allBundleTags[strings.TrimPrefix(fullBundleRef, cmd.TagRefPrefix)] = commit
				case strings.HasPrefix(fullBundleRef, cmd.HeadRefPrefix):
					allBundleHeads[strings.TrimPrefix(fullBundleRef, cmd.HeadRefPrefix)] = commit
				default:
					syncErrs = append(syncErrs, fmt.Errorf("%s contains invalid reference prefix, ref = %s, commit = %s", bundleName, fullBundleRef, commit))
				}
			}
		}
	}

	return config, allBundleTags, allBundleHeads, errors.Join(syncErrs...)
}

// validateBundleRefs validates that all references extracted from the bundle(s) are expected and that the
// commit objects they refer to match the sync config. Note this function does not validate that the refs
// found in the bundles exactly matches the config, i.e. in some cases a ref only exists in the config
// because it's commit already exists in a prior bundle; this case is checked in validateConfig.
//
// validateBundleRefs answers 3 questions for both tag and head references:
// 1) Is the reference expected?
// 2) Does the reference in the bundle exist in the config?
// 3) Does the commit referenced by the tag in the bundle match the config?
func validateBundleRefs(ctx context.Context, config oci.Config, foundTags, foundHeads map[string]cmd.Commit, expectedTags, expectedHeads map[string]bool) error {
	bundleErrs := make([]error, 0)

	for tag, bundleCommit := range foundTags {

		// is the tag expected?
		_, ok := expectedTags[tag]
		if !ok {
			bundleErrs = append(bundleErrs, fmt.Errorf("tag in bundle(s) is not expected, tag = %s", tag))
			continue
		}
		expectedTags[tag] = true // found

		// does the tag in the bundle exist in the config?
		refInfo, ok := config.Refs.Tags[tag]
		if !ok {
			bundleErrs = append(bundleErrs, fmt.Errorf("tag in bundle(s) is not in config, tag = %s", tag))
		}
		cfgCommit := refInfo.Commit

		// does the commit referenced by the tag in the bundle match the config?
		if bundleCommit != cfgCommit {
			bundleErrs = append(bundleErrs, fmt.Errorf("commit referenced by tag %s in bundle(s) does not match config, bundle commit = %s, config commit = %s", tag, bundleCommit, cfgCommit))
		}
	}

	for head, bundleCommit := range foundHeads {

		// is the head expected?
		_, ok := expectedHeads[head]
		if !ok {
			bundleErrs = append(bundleErrs, fmt.Errorf("head in bundle(s) is not expected, head = %s", head))
			continue
		}
		expectedHeads[head] = true // found

		// does the head in the bundle exist in the config?
		refInfo, ok := config.Refs.Heads[head]
		if !ok {
			bundleErrs = append(bundleErrs, fmt.Errorf("head in bundle(s) is not in config, head = %s", head))
		}
		cfgCommit := refInfo.Commit

		// does the commit referenced by the head in the bundle match the config?
		if bundleCommit != cfgCommit {
			bundleErrs = append(bundleErrs, fmt.Errorf("commit referenced by head %s in bundle(s) does not match config, bundle commit = %s, config commit = %s", head, bundleCommit, cfgCommit))
		}
	}

	return errors.Join(bundleErrs...)
}

// validateConfig checks to see if a sync config contains the expected tag and head references, as well as a valid api field.
func validateConfig(config oci.Config, expectedAPIVersion string, expectedTagsMap,
	expectedHeadsMap map[string]bool) error {

	configErrs := make([]error, 0)

	// catch extraneous tags
	for tagRef, commit := range config.Refs.Tags {
		if _, ok := expectedTagsMap[tagRef]; !ok {
			configErrs = append(configErrs, fmt.Errorf("unexpected tag found in config, tag = %s, commit = %s", tagRef, commit))
			continue
		}
		expectedTagsMap[tagRef] = true
	}

	// catch missing tags
	if len(config.Refs.Tags) < len(expectedTagsMap) {
		for tag, found := range expectedTagsMap {
			if !found {
				configErrs = append(configErrs, fmt.Errorf("expected tag %s not found in config", tag))
			}
		}
	}

	// catch extraneous heads
	for headRef, commit := range config.Refs.Heads {
		if _, ok := expectedHeadsMap[headRef]; !ok {
			configErrs = append(configErrs, fmt.Errorf("unexpected head found in config, head = %s, commit = %s", headRef, commit))
			continue
		}
		expectedHeadsMap[headRef] = true
	}

	// catch missing heads
	if len(config.Refs.Heads) < len(expectedHeadsMap) {
		for head, found := range expectedHeadsMap {
			if !found {
				configErrs = append(configErrs, fmt.Errorf("expected head %s not found in config", head))
			}
		}
	}

	return errors.Join(configErrs...)
}

// validate rebuild validates the results of FromOCI.
func validateRebuild(expectedRefs map[string]bool, gotRefs []string) error {
	rebuildErrs := make([]error, 0)

	foundRefs := 0
	for _, ref := range gotRefs {
		split := strings.Fields(ref)
		commit := split[0] // typically, but in case of rewrite over a tag it's "Deleted"
		fullRef := split[1]

		switch {
		case commit == "Deleted":
			expected := strings.Join(split[0:4], " ")
			_, ok := expectedRefs[expected] // "Deleted tag 'split[3]'"
			if !ok {
				rebuildErrs = append(rebuildErrs, fmt.Errorf("unexpected tag deleted from rebuild repository, tag = %s", strings.Replace(split[3], "'", "", 2)))
			}
			expectedRefs[expected] = true
			foundRefs++
		case strings.HasPrefix(fullRef, cmd.TagRefPrefix):
			_, ok := expectedRefs[fullRef]
			if !ok {
				rebuildErrs = append(rebuildErrs, fmt.Errorf("unexpected tag updated in rebuilt repository, tag = %s, commit = %s", fullRef, commit))
				break
			}
			expectedRefs[fullRef] = true
			foundRefs++

		case strings.HasPrefix(fullRef, cmd.HeadRefPrefix):
			_, ok := expectedRefs[fullRef]
			if !ok {
				rebuildErrs = append(rebuildErrs, fmt.Errorf("unexpected head updated in rebuilt repository, head = %s, commit = %s", fullRef, commit))
				break
			}
			expectedRefs[fullRef] = true
			foundRefs++

		default:
			rebuildErrs = append(rebuildErrs, fmt.Errorf("unexpected reference type updated in rebuilt repository, ref = %s, commit = %s", fullRef, commit))
		}
	}

	if foundRefs < len(expectedRefs) {
		for ref, found := range expectedRefs {
			if !found {
				rebuildErrs = append(rebuildErrs, fmt.Errorf("expected ref not found in rebuilt repository, ref = %s", ref))
			}
		}
	}

	return errors.Join(rebuildErrs...)
}

func deleteDanglingCommits(ctx context.Context, ch *cmd.Helper) error {
	// Delete dangling commits, i.e. actually delete the history we diverged from
	if _, err := ch.Git.Run(ctx, "reflog", "expire", "--expire-unreachable=now", "--all"); err != nil {
		return fmt.Errorf("expiring dangling commits: %w", err)
	}

	if _, err := ch.Git.Run(ctx, "gc", "--prune=now"); err != nil {
		return fmt.Errorf("deleting dangling commits: %w", err)
	}

	return nil
}
