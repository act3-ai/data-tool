package git

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"

	"github.com/act3-ai/data-tool/internal/git/cache"
	"github.com/act3-ai/data-tool/internal/git/cmd"
	"github.com/act3-ai/go-common/pkg/logger"
	tlog "github.com/act3-ai/go-common/pkg/test"
)

func Test_GitCache(t *testing.T) {
	ctx := context.Background()
	ctx = logger.NewContext(ctx, tlog.Logger(t, 0))

	// init
	srcRepo, srcHandler, srcServer, dstRepo, dstHandler, dstServer := setupLFSServerHandlers(t, ctx)
	defer srcHandler.cleanup() //nolint
	defer srcServer.Close()
	defer dstHandler.cleanup() //nolint
	defer dstServer.Close()

	target := memory.New() // oci target
	toOCICacheDir := t.TempDir()
	fromOCICacheDir := t.TempDir()

	// run LFS tests
	for i, tt := range lfsTests {
		// prepare ground-truth
		gitOids, lfsOids := reachableOIDs(t, ctx, tt.t.args.argRevList, srcHandler.cmdHelper)

		if i == 0 {
			ToFromOCICacheTestFunc(ctx, tt, target, ocispec.Descriptor{}, srcHandler, srcRepo, srcServer.URL, dstHandler, dstRepo, dstServer.URL, toOCICacheDir, fromOCICacheDir, gitOids, lfsOids)(t)
		} else {
			existingDesc, err := target.Resolve(ctx, tt.t.args.tag)
			if err != nil {
				t.Fatalf("resolving base manifest descriptor: error = %s", err)
			}

			ToFromOCICacheTestFunc(ctx, tt, target, existingDesc, srcHandler, srcRepo, srcServer.URL, dstHandler, dstRepo, dstServer.URL, toOCICacheDir, fromOCICacheDir, gitOids, lfsOids)(t)
		}
	}

	// cleanup
	if err := srcHandler.Cleanup(); err != nil {
		t.Errorf("cleaning up source handler: %v", err)
	}
	if err := dstHandler.Cleanup(); err != nil {
		t.Errorf("cleaning up destination handler: %v", err)
	}

	t.Log("test complete")
}

//nolint:gocognit
func ToFromOCICacheTestFunc(ctx context.Context, tt lfsTest, target oras.GraphTarget, existingDesc ocispec.Descriptor,
	srcHandler *ToOCI, srcRepo, srcServer string,
	dstHandler *FromOCI, dstRepo, dstServer string,
	toOCICacheDir, fromOCICacheDir string,
	reachableGitOids, reachableLFSOids []string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		var newDesc ocispec.Descriptor
		t.Run(tt.t.name+": ToOCI-Cache", func(t *testing.T) {
			toOCIFStorePath := t.TempDir()
			toOCIFStore, err := file.New(toOCIFStorePath)
			if err != nil {
				t.Fatalf("initializing to ocicache file store: %v", err)
			}
			defer toOCIFStore.Close()
			toOCICmdOpts := cmd.Options{
				LFSOptions: &cmd.LFSOptions{WithLFS: true, ServerURL: srcServer},
			}
			toOCICache, err := cache.NewCache(ctx, toOCICacheDir, toOCIFStorePath, toOCIFStore, &toOCICmdOpts)
			if err != nil {
				t.Fatalf("initializing to oci base cache: %v", err)
			}

			toOCICacheLink, err := toOCICache.NewLink(ctx, toOCICmdOpts)
			if err != nil {
				t.Errorf("establishing to oci new cache link: %v", err)
			}

			syncOpts := SyncOptions{IntermediateDir: toOCIFStorePath, IntermediateStore: toOCIFStore, Cache: toOCICacheLink} // new intermediate dir, same to-oci cache dir
			toOCITester, err := NewToOCI(ctx, target, existingDesc, srcRepo, tt.t.args.argRevList, syncOpts, &toOCICmdOpts)
			if err != nil {
				t.Errorf("creating ToOCI: %v", err)
			}
			defer toOCITester.Cleanup() //nolint

			newDesc, err = toOCITester.Run(ctx)
			if err != nil {
				t.Errorf("ToOCI() error = %v, wantErr %v", err, tt.t.wantErr)
			}

			if err := validateToOCICache(toOCITester.syncOpts.Cache.CachePath(), len(reachableGitOids), len(reachableLFSOids)); err != nil {
				t.Errorf("resulting cache is invalid: %v", err)
			}

			if err := toOCITester.Cleanup(); err != nil {
				t.Errorf("cleaning up toOCITester handler: %v", err)
			}

			if err := target.Tag(ctx, newDesc, tt.t.args.tag); err != nil {
				t.Fatalf("tagging new sync manifest: error = %s", err)
			}
		})

		t.Run(tt.t.name+": FromOCI-Cache", func(t *testing.T) {
			fromOCIFStorePath := t.TempDir()
			fromOCIFStore, err := file.New(fromOCIFStorePath)
			if err != nil {
				t.Fatalf("initializing from oci cache file store: %v", err)
			}
			defer fromOCIFStore.Close()
			fromOCICmdOpts := cmd.Options{
				LFSOptions: &cmd.LFSOptions{WithLFS: true, ServerURL: dstServer},
			}
			fromOCICache, err := cache.NewCache(ctx, fromOCICacheDir, fromOCIFStorePath, fromOCIFStore, &fromOCICmdOpts)
			if err != nil {
				t.Fatalf("initializing from oci base cache: %v", err)
			}

			fromOCICacheLink, err := fromOCICache.NewLink(ctx, fromOCICmdOpts)
			if err != nil {
				t.Errorf("establishing from oci new cache link: %v", err)
			}

			syncOpts := SyncOptions{IntermediateDir: fromOCIFStorePath, IntermediateStore: fromOCIFStore, Cache: fromOCICacheLink} // new intermediate dir, same from-oci cache dir
			fromOCITester, err := NewFromOCI(ctx, target, newDesc, dstRepo, syncOpts, &fromOCICmdOpts)
			if err != nil {
				t.Errorf("creating FromOCI: %v", err)
			}
			defer fromOCITester.Cleanup() //nolint

			updatedRefs, err := fromOCITester.Run(ctx)
			if err != nil {
				t.Fatalf("from oci: %v", err)
			}
			t.Logf("updated refs: %s", updatedRefs)

			if err := validateFromOCICache(ctx, fromOCITester.syncOpts.Cache.CachePath(), len(reachableGitOids), len(reachableLFSOids)); err != nil {
				t.Errorf("resulting cache is invalid: %v", err)
			}

			if err := fromOCITester.Cleanup(); err != nil {
				t.Errorf("cleaning up fromOCITester handler: %v", err)
			}
		})
	}
}

// reachableOIDs determines git and git-lfs objects reachable from a set of git references within the source repository.
// The results are used for validation.
func reachableOIDs(t *testing.T, ctx context.Context, argRevList []string, lfsSrcCmdHelper *cmd.Helper) ([]string, []string) { //nolint
	t.Helper()
	args := []string{"--objects"} // --objects is crucial, as this will tell us all objects that are fetched to the cache; else its just commits
	args = append(args, argRevList...)
	reachableGitOids, err := revList(ctx, lfsSrcCmdHelper, args...)
	if err != nil {
		t.Errorf("resolving reachable git objects: %v", err)
	}

	reachableLFSOids, err := lfsSrcCmdHelper.ListReachableLFSFiles(ctx, argRevList)
	if err != nil {
		t.Errorf("resolving reachable lfs files: %v", err)
	}

	return reachableGitOids, reachableLFSOids
}

// validateToCache validates the number of git objects in a git cache directory.
// TODO: validate the OIDs themselves as well.
func validateToOCICache(cachePath string, expectedObjs, expectedLFSObjs int) error {
	// git objects
	objsDir := filepath.Join(cachePath, "objects")
	oids, err := listGitObjects(objsDir)
	if err != nil {
		return fmt.Errorf("resolving all git objects: %w", err)
	}
	if len(oids) != expectedObjs {
		return fmt.Errorf("unexpected number of git objects incache, want = %d, got = %d", expectedObjs, len(oids))
	}

	// git-lfs objects
	lfsObjsDir := filepath.Join(cachePath, "lfs", "objects")
	lfsOids, err := listLFSObjects(lfsObjsDir)
	if err != nil {
		return fmt.Errorf("resolving all git-lfs objects: %w", err)
	}
	if len(lfsOids) != expectedLFSObjs {
		return fmt.Errorf("unexpected number of git lfs objects incache, want = %d, got = %d", expectedLFSObjs, len(lfsOids))
	}

	return nil
}

// validateFromOCICache validates the number of git objects in a git cache directory.
// In practice, the objects are packed within packfiles when extracted from bundles.
//
// Note: Errors may occur here if multiple packfiles contain the same obj, however due
// to the "thin"/"shallow" nature of how we manage bundles this should not be an issue.
func validateFromOCICache(ctx context.Context, cachePath string, expectedObjs, expectedLFSObjs int) error {
	// count objs in all packfiles
	packPath := filepath.Join(cachePath, "objects", "pack")
	allPackIdxs, err := resolvePacks(packPath) // actually indexes, which correspond to packs
	if err != nil {
		return fmt.Errorf("resolving packfiles in cache: %w", err)
	}
	if len(allPackIdxs) < 1 {
		return fmt.Errorf("no packfiles found")
	}

	ch, err := cmd.NewHelper(ctx, cachePath, &cmd.Options{})
	if err != nil {
		return fmt.Errorf("initializing cache validation command helper: %w", err)
	}
	var objTotal int
	for _, packIdx := range allPackIdxs {
		packObjs, err := verifyPack(ctx, ch, filepath.Join(packPath, packIdx))
		if err != nil {
			return fmt.Errorf("validating packfile '%s': %w", packIdx, err)
		}

		objTotal += packObjs
	}

	if objTotal != expectedObjs {
		return fmt.Errorf("unexpected number of git objects found in cache, want = %d, got = %d", expectedObjs, objTotal)
	}

	// git-lfs objects
	lfsObjsDir := filepath.Join(cachePath, "lfs", "objects")
	lfsOids, err := listLFSObjects(lfsObjsDir)
	if err != nil {
		return fmt.Errorf("resolving all git-lfs objects: %w", err)
	}
	if len(lfsOids) != expectedLFSObjs {
		return fmt.Errorf("unexpected number of git lfs objects founc in cache, want = %d, got = %d", expectedLFSObjs, len(lfsOids))
	}

	return nil
}

// listGitObjects walks a git objects directory, returning a slice
// of all oids. The path to the oid may be derived with oid[0:2]/oid[2:4]/oid[4:].
func listGitObjects(objectsDir string) (oids []string, err error) {
	// git's path naming, unlike git-lfs', is named as ab/cd/efghi... where
	// the oid is abcdefghi...
	// As such, we reconstruct the actual oid which is not the filename

	objsFS := os.DirFS(objectsDir)
	var relativePath string
	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := fs.Stat(objsFS, path)
		if err != nil {
			return fmt.Errorf("unable to get the actual file info: %w", err)
		}

		// we don't care about directories, just the files in them
		if info.IsDir() {
			relativePath = filepath.Join(relativePath, info.Name())
			return nil
		}

		// object found
		oids = append(oids, info.Name())
		relativePath = ""

		return nil
	}

	// begin walk
	if err := fs.WalkDir(objsFS, ".", walkFn); err != nil { // Note: loading entire obj dir into memory may be problematic
		return oids, fmt.Errorf("walking obj directory: %w", err)
	}

	return oids, nil
}

// listLFSObjects walks an lfs objects directory, returning all oids found.
//
// We could use `git-lfs ls-files --all`, but this is incompatible with a bare repository,
// which is what we're evaluating in the tests.
func listLFSObjects(objectsDir string) (oids []string, err error) {
	lfsObjsFS := os.DirFS(objectsDir)

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := fs.Stat(lfsObjsFS, path)
		if err != nil {
			return fmt.Errorf("unable to get the actual file info: %w", err)
		}

		// we don't care about directories, just the files in them
		if info.IsDir() {
			return nil
		}

		// add relative path
		oids = append(oids, info.Name())

		return nil
	}

	// begin walk
	if err := fs.WalkDir(lfsObjsFS, ".", walkFn); err != nil { // TODO: loading entire obj dir into memory may be problematic
		return oids, fmt.Errorf("walking LFS obj directory: %w", err)
	}

	return oids, nil
}

// resolvePacks returns a slice of all *.idx files within a git repositories pack dir.
// Each *.idx should have a corresponding *.pack containing the objects themselves.
func resolvePacks(packDir string) ([]string, error) {
	var packIndexes []string
	objsFS := os.DirFS(packDir)
	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := fs.Stat(objsFS, path)
		if err != nil {
			return fmt.Errorf("unable to get the actual file info: %w", err)
		}

		// we don't care about directories, just the files in them.
		// here, we should not encounter sub-directories.
		if info.IsDir() {
			return nil
		}

		// searching for *.idx files
		if filepath.Ext(info.Name()) == ".idx" {
			packIndexes = append(packIndexes, info.Name())
		}

		return nil
	}

	// begin walk
	if err := fs.WalkDir(objsFS, ".", walkFn); err != nil { // Note: loading entire obj dir into memory may be problematic
		return packIndexes, fmt.Errorf("walking obj directory: %w", err)
	}

	return packIndexes, nil
}
