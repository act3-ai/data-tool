package archive

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkHdr(t *testing.T, hdr *tar.Header) {
	t.Helper()
	assert.Equal(t, 0, hdr.Uid, 0)
	assert.Equal(t, 0, hdr.Gid, 0)
	assert.Equal(t, "", hdr.Uname)
	assert.Equal(t, "", hdr.Gname)

	et := time.Unix(0, 0).UTC()
	t.Logf("expected time is %s", et)
	assert.Equal(t, time.Time{}, hdr.AccessTime)
	assert.Equal(t, time.Time{}, hdr.ChangeTime)
	assert.True(t, et.Equal(hdr.ModTime))
	assert.Equal(t, et, hdr.ModTime.UTC()) // TODO is this actually stored in the archive as local time?
}

// Test the archive of a single file in a directory.
func Test_CreateFromPath_OneFile(t *testing.T) {

	fsys := fstest.MapFS{
		"testfile1.txt": &fstest.MapFile{
			Data: []byte("Random Data"),
		},
	}

	buf := new(bytes.Buffer)
	require.NoError(t, TarToStream(context.Background(), fsys, buf))

	assert.Equal(t, "sha256:e00389079424742073a0d09ea29a267818e00be3c22a2348b559a284af99dff0",
		digest.FromBytes(buf.Bytes()).String())

	t.Log("Check the tar archive")
	tr := tar.NewReader(buf)

	hdr, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "testfile1.txt", hdr.Name)
	checkHdr(t, hdr)
	assert.Equal(t, hdr.Size, int64(11))
	assert.False(t, hdr.FileInfo().IsDir())
}

func Test_CreateFromPath_ManyFiles(t *testing.T) {
	bigZero := bytes.Repeat([]byte{0}, 1e6)

	fsys1 := fstest.MapFS{
		"testfile1.txt":        &fstest.MapFile{Data: []byte("Random Data")},
		"testfile2.txt":        &fstest.MapFile{Data: []byte("Random Data2")},
		"somedir/file.txt":     &fstest.MapFile{Data: []byte("data in directory")},
		"somedir/.hidden.txt":  &fstest.MapFile{Data: []byte("hidden data in directory")}, // This should not be included
		".somedir/.hidden.txt": &fstest.MapFile{Data: []byte("hidden directory")},         // This should not be included
		"emptydir":             &fstest.MapFile{Mode: fs.ModeDir},                         // This should not be included
		"bigZero.dat":          &fstest.MapFile{Data: bigZero},
	}

	t.Run("simple", func(t *testing.T) {
		checkFS(t, fsys1)
	})

	dir := t.TempDir()
	otherDir := t.TempDir()

	// We  need to skip some of the test if the underlying FS does not support Symlinks, Hardlinks, etc.
	// TODO verify that all combos of these pass
	testSymlinks := true
	testSymlinkDirs := true
	testHardlinks := true
	testSparseFiles := false

	if runtime.GOOS == "windows" {
		// The windows filesystem does not reallty support these types
		testSymlinks = false
		testSymlinkDirs = false
		testHardlinks = false
		testSparseFiles = false
	}

	fj := filepath.Join
	rne := require.New(t).NoError

	// fill up the otherDir with goodies
	rne(os.WriteFile(fj(otherDir, "myfile.txt"), []byte("Random Data"), 0666))
	rne(os.Mkdir(fj(otherDir, "someotherdir"), 0777))
	rne(os.WriteFile(fj(otherDir, "someotherdir", "file.txt"), []byte("data in directory"), 0666))
	rne(os.WriteFile(fj(otherDir, "someotherdir", ".hidden.txt"), []byte("hidden data in directory"), 0666)) // This should not be included
	rne(os.WriteFile(fj(otherDir, "other-testfile2.txt"), []byte("Random Data2"), 0666))

	// fill up the dir with goodies
	if testSymlinks {
		rne(os.Symlink(fj(otherDir, "myfile.txt"), fj(dir, "testfile1.txt"))) // symlinked file should behave like a regular file
	} else {
		rne(os.WriteFile(fj(dir, "testfile1.txt"), []byte("Random Data"), 0666))
	}
	if testHardlinks {
		rne(os.Link(fj(otherDir, "other-testfile2.txt"), fj(dir, "testfile2.txt"))) // hard links should behave like a regular file
	} else {
		rne(os.WriteFile(fj(dir, "testfile2.txt"), []byte("Random Data2"), 0666))
	}
	if testSymlinkDirs {
		rne(os.Symlink(fj(otherDir, "someotherdir"), fj(dir, "somedir"))) // should behave like a regular directory
	} else {
		rne(os.Mkdir(fj(dir, "somedir"), 0777))
		rne(os.WriteFile(fj(dir, "somedir", "file.txt"), []byte("data in directory"), 0666))
		rne(os.WriteFile(fj(dir, "somedir", ".hidden.txt"), []byte("hidden data in directory"), 0666)) // This should not be included
	}
	rne(os.Mkdir(fj(dir, ".somedir"), 0777))                                               // This should not be included
	rne(os.WriteFile(fj(dir, ".somedir", "hidden.txt"), []byte("hidden directory"), 0666)) // This should not be included
	rne(os.Mkdir(fj(dir, "emptydir"), 0777))                                               // This should not be included
	if testSparseFiles {
		f, err := os.Create(fj(dir, "bigZero.dat"))
		rne(err)
		rne(f.Truncate(1e6)) // 1MB
		rne(f.Close())
	} else {
		rne(os.WriteFile(fj(dir, "bigZero.dat"), bigZero, 0666))
	}

	fsys2 := os.DirFS(dir)

	t.Run("real", func(t *testing.T) {
		checkFS(t, fsys2)
	})

	t.Run("playground", func(t *testing.T) {
		rne := require.New(t).NoError

		if testSymlinkDirs {
			// Test with Open()
			{
				f, err := fsys2.Open("somedir") // symlink to directory
				rne(err)
				defer f.Close()

				stat, err := f.Stat()
				rne(err)

				assert.True(t, stat.IsDir())
				assert.Equal(t, "somedir", stat.Name())
				assert.Equal(t, fs.FileMode(0), stat.Mode().Type()&fs.ModeSymlink)
				rne(f.Close())
			}

			// Test with Stat()
			{
				stat, err := fs.Stat(fsys2, "somedir") // symlink to directory
				rne(err)

				assert.True(t, stat.IsDir())
				assert.Equal(t, "somedir", stat.Name())
				assert.Equal(t, fs.FileMode(0), stat.Mode().Type()&fs.ModeSymlink)
			}
		}

		if testSymlinks {
			// test with Open()
			{
				f, err := fsys2.Open("testfile1.txt") // symlink to directory
				rne(err)
				defer f.Close()

				stat, err := f.Stat()
				rne(err)

				assert.False(t, stat.IsDir())
				assert.Equal(t, "testfile1.txt", stat.Name())
				assert.Equal(t, fs.FileMode(0), stat.Mode().Type()&fs.ModeSymlink)
				rne(f.Close())
			}

			// Test with Stat()
			{
				stat, err := fs.Stat(fsys2, "testfile1.txt") // symlink to directory
				rne(err)

				assert.False(t, stat.IsDir())
				assert.Equal(t, "testfile1.txt", stat.Name())
				assert.Equal(t, fs.FileMode(0), stat.Mode().Type()&fs.ModeSymlink)
			}
		}
	})
}

func checkFS(t *testing.T, fsys fs.FS) {
	t.Helper()
	buf := new(bytes.Buffer)
	require.NoError(t, TarToStream(context.Background(), fsys, buf))

	assert.Equal(t, "sha256:d62187986de71c4ba160fbff9120265fa8f6faa60a1120fb843fb13610bc26f2",
		digest.FromBytes(buf.Bytes()).String())

	// assert.NoError(t, os.WriteFile("/tmp/CreateFromPath_ManyFiles.tar", buf.Bytes(), 0666))

	t.Log("Check the tar archive")
	tr := tar.NewReader(buf)

	hdr, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "bigZero.dat", hdr.Name)
	checkHdr(t, hdr)
	assert.Equal(t, int64(1e6), hdr.Size)
	assert.False(t, hdr.FileInfo().IsDir())

	hdr, err = tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "emptydir/", hdr.Name)
	checkHdr(t, hdr)
	assert.True(t, hdr.FileInfo().IsDir())

	hdr, err = tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "somedir/", hdr.Name)
	checkHdr(t, hdr)
	assert.True(t, hdr.FileInfo().IsDir())

	hdr, err = tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "somedir/file.txt", hdr.Name)
	checkHdr(t, hdr)
	assert.Equal(t, int64(17), hdr.Size)
	assert.False(t, hdr.FileInfo().IsDir())

	hdr, err = tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "testfile1.txt", hdr.Name)
	checkHdr(t, hdr)
	assert.Equal(t, int64(11), hdr.Size)
	assert.False(t, hdr.FileInfo().IsDir())

	hdr, err = tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "testfile2.txt", hdr.Name)
	checkHdr(t, hdr)
	assert.Equal(t, int64(12), hdr.Size)
	assert.False(t, hdr.FileInfo().IsDir())

	_, err = tr.Next()
	require.ErrorIs(t, err, io.EOF)
}

/*
// This is just a toy test to show that I have no idea how to make special files (such as symlinks) with MapFS
// Maybe I will figure it out some day
func Test_SpecialFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"link.txt": &fstest.MapFile{
			Data: []byte("testfile1.txt"),
			Mode: fs.ModeSymlink,
		},
		"testfile1.txt": &fstest.MapFile{Data: []byte("Random Data")},
	}

	info, err := fs.Stat(fsys, "link.txt")
	assert.NoError(t, err)
	t.Logf("info = %s", info)

	data, err := fs.ReadFile(fsys, "link.txt")
	assert.NoError(t, err)
	t.Logf("data = %s", string(data))
}
*/
