package git

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"oras.land/oras-go/v2/content/memory"

	"git.act3-ace.com/ace/data/tool/internal/git/cmd"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	tlog "git.act3-ace.com/ace/go-common/pkg/test"
)

func Test_GitCache(t *testing.T) {
	ctx := context.Background()
	ctx = logger.NewContext(ctx, tlog.Logger(t, -2))

	lfsSrc, lfsSrcHandler, srcServer, lfsDst, lfsDstHandler, dstServer := setupLFSServerHandlers(t, ctx)
	defer lfsSrcHandler.cleanup() //nolint
	defer srcServer.Close()
	defer lfsDstHandler.cleanup() //nolint
	defer dstServer.Close()

	target := memory.New() // oci target
	toOCICache := t.TempDir()
	fromOCICache := t.TempDir()

	// run LFS tests
	for _, tt := range lfsTests {

		args := []string{"--objects"} // --objects is crucial, as this will tell us all objects that are fetched to the cache; else its just commits
		args = append(args, tt.t.args.argRevList...)
		reachableGitObjs, err := revList(lfsSrcHandler.cmdHelper, args...)
		if err != nil {
			t.Errorf("resolving reachable git objects: %v", err)
		}

		reachableLFSFiles, err := lfsSrcHandler.cmdHelper.ListReachableLFSFiles(tt.t.args.argRevList...)
		if err != nil {
			t.Errorf("resolving reachable lfs files: %v", err)
		}

		t.Run(tt.t.name+":ToOCILFS Cache", func(t *testing.T) {

			syncOpts := SyncOptions{TmpDir: t.TempDir(), CacheDir: toOCICache}
			cmdOpts := cmd.Options{LFSOptions: &cmd.LFSOptions{WithLFS: true, ServerURL: srcServer.URL}}
			toOCITester, err := NewToOCI(ctx, target, tt.t.args.tag, lfsSrc, tt.t.args.argRevList, syncOpts, &cmdOpts)
			if err != nil {
				t.Errorf("creating ToOCI: %v", err)
			}
			defer toOCITester.Cleanup() //nolint

			_, err = toOCITester.Run(ctx)
			if err != nil {
				t.Errorf("ToOCI() error = %v, wantErr %v", err, tt.t.wantErr)
			}

			if err := validateToOCICache(toOCITester.syncOpts.CacheDir, len(reachableGitObjs), len(reachableLFSFiles)); err != nil {
				t.Errorf("resulting cache is invalid: %v", err)
			}

			if err := toOCITester.Cleanup(); err != nil {
				t.Errorf("cleaning up toOCITester handler: %v", err)
			}
		})

		t.Run(tt.t.name+"FromOCILFS Cache", func(t *testing.T) {

			syncOpts := SyncOptions{TmpDir: t.TempDir(), CacheDir: fromOCICache}
			cmdOpts := cmd.Options{LFSOptions: &cmd.LFSOptions{WithLFS: true, ServerURL: dstServer.URL}}
			fromOCITester, err := NewFromOCI(ctx, target, tt.t.args.tag, lfsDst, syncOpts, &cmdOpts)
			if err != nil {
				t.Errorf("creating FromOCI: %v", err)
			}
			defer fromOCITester.Cleanup() //nolint

			updatedRefs, err := fromOCITester.Run(ctx)
			if err != nil {
				t.Fatalf("from oci: %v", err)
			}
			t.Logf("updated refs: %s", updatedRefs)

			if err := validateFromOCICache(ctx, fromOCITester.syncOpts.CacheDir, len(reachableGitObjs), len(reachableLFSFiles)); err != nil {
				t.Errorf("resulting cache is invalid: %v", err)
			}

			if err := fromOCITester.Cleanup(); err != nil {
				t.Errorf("cleaning up fromOCITester handler: %v", err)
			}
		})
	}

	if err := lfsSrcHandler.Cleanup(); err != nil {
		t.Errorf("cleaning up source handler: %v", err)
	}

	if err := lfsDstHandler.Cleanup(); err != nil {
		t.Errorf("cleaning up destination handler: %v", err)
	}

	t.Log("test complete")
}

// validateToCache validates the number of git objects in a git cache directory.
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
		packObjs, err := verifyPack(ch, filepath.Join(packPath, packIdx))
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
