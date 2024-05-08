package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirSizeSymlinks(t *testing.T) {
	// Test my test code

	d := t.TempDir()

	// add some files directories
	err := os.Mkdir(filepath.Join(d, "subdir"), 0777)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(d, "file1"), []byte("the data"), 0666)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(d, "subdir", "file2"), []byte("in sub dir"), 0666)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Symlink(filepath.Join(d, "file1"), filepath.Join(d, "symlink-to-file1"))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Link(filepath.Join(d, "subdir", "file2"), filepath.Join(d, "hardlink-to-file2"))
	if err != nil {
		t.Fatal(err)
	}

	var size int64
	size, err = DirSize(os.DirFS(d))
	if err != nil {
		t.Fatal(err)
	}

	if size != 8+10 {
		t.Errorf("expected 18 B but got %d B", size)
	}
}

/*
func TestReadDirSortedByAccessTime(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		setupFS     func() fs.FS
		path        string
		expectedErr error
	}{
		{
			name: "normal",
			setupFS: func() fs.FS {
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: time.Now(),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: time.Now(),
				}
				return fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
				}
			},
			path:        ".",
			expectedErr: nil,
		},
		{
			name: "error_getting_file_info",
			setupFS: func() fs.FS {
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: time.Now(),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: time.Now(),
				}
				mapFS := fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
					"error_info.txt": &fstest.MapFile{
						Data: []byte("error info content"),
						Mode: 0644,
					},
				}
				return mapFS
			},
			path:        ".",
			expectedErr: fmt.Errorf("error getting file info: %w", errors.New("Info error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := tc.setupFS()

			infos, err := ReadDirSortedByAccessTime(fsys, tc.path)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)

				// Check if the files are sorted by access time
				for i := 0; i < len(infos)-1; i++ {
					assert.True(t, atime.Get(infos[i]).Before(atime.Get(infos[i+1])))
				}
			}
		})
	}
}
*/
/*
func TestGetDirLastUpdate(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		setupFS     func() fs.FS
		expectedErr error
	}{
		{
			name: "normal",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now,
				}
				return fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
				}
			},
			expectedErr: nil,
		},
		{
			name: "error_getting_file_info",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now,
				}
				mapFS := fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
					"error_info.txt": &fstest.MapFile{
						Data: []byte("error info content"),
						Mode: 0644,
					},
				}
				return mapFS
			},
			expectedErr: fmt.Errorf("error getting file info: %w", errors.New("Info error")),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := tc.setupFS()

			lastUpdate, err := GetDirLastUpdate(fsys)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)

				fileBInfo, err := fsys.Open("fileB.txt")
				require.NoError(t, err)
				fileBStat, err := fileBInfo.Stat()
				require.NoError(t, err)

				assert.Equal(t, fileBStat.ModTime(), lastUpdate)
			}
		})
	}
}
*/

/*
func TestGetDirUpdatedPaths(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		setupFS     func() fs.FS
		earliest    time.Time
		expected    []string
		expectedErr error
	}{
		{
			name: "normal",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-2 * time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				return fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
				}
			},
			earliest: time.Now().Add(-90 * time.Minute),
			expected: []string{"fileB.txt"},
		},
		{
			name: "error_getting_file_info",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-2 * time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				mapFS := fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
					"error_info.txt": &fstest.MapFile{
						Data: []byte("error info content"),
						Mode: 0644,
					},
				}
				return mapFS
			},
			earliest:    time.Now().Add(-90 * time.Minute),
			expectedErr: fmt.Errorf("error getting file info: %w", errors.New("Info error")),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := tc.setupFS()

			updatedPaths, err := GetDirUpdatedPaths(fsys, tc.earliest)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, updatedPaths)
			}
		})
	}
}
*/
