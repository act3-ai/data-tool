package bottle

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFSUtilAndClose(t *testing.T) {
	fsys, err := NewFSUtil("test")
	require.NoError(t, err, "NewFSUtil should not return an error")

	// Check if the directory exists and is a directory
	dirInfo, err := os.Stat(fsys.RootDir)
	require.NoError(t, err, "RootDir should exist")
	assert.True(t, dirInfo.IsDir(), "RootDir should be a directory")

	// Create a test file in the root directory
	testFilePath := filepath.Join(fsys.RootDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("test"), 0644)
	require.NoError(t, err, "Creating a test file should not return an error")

	// Close FSUtil and remove the temporary directory
	err = fsys.Close()
	require.NoError(t, err, "Close should not return an error")

	// Check if the temporary directory was removed
	_, err = os.Stat(fsys.RootDir)
	assert.Error(t, err, "RootDir should not exist after Close")
	assert.True(t, os.IsNotExist(err), "RootDir should be removed")

}

func TestAddFileWithData(t *testing.T) {
	// This is needed because Windows does not recognize / as an absolute path
	absoluteFilepath, err := filepath.Abs("/invalid/data.txt")
	require.NoError(t, err, "filepath.Abs should not return an error")
	testCases := []struct {
		name   string
		path   string
		data   []byte
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "data.txt",
			data: []byte("Hello, world!"),
		},
		{
			name: "Nested relative path",
			path: "nested/data.txt",
			data: []byte("Nested file content"),
		},
		{
			name: "Empty file",
			path: "empty.txt",
			data: []byte{},
		},
		{
			name:   "Invalid absolute path",
			path:   absoluteFilepath,
			data:   []byte("Invalid absolute path"),
			errMsg: fmt.Sprintf("path %s is absolute. All FSUtil paths are relative", absoluteFilepath),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fsys.Close()

			err = fsys.AddFileWithData(tc.path, tc.data)
			if tc.errMsg != "" {
				require.Error(t, err, "AddFileWithData should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddFileWithData should not return an error")

				// Check if the file exists and has the correct content
				filePath := filepath.Join(fsys.RootDir, tc.path)
				fileContent, err := os.ReadFile(filePath)
				require.NoError(t, err, "File should be readable")
				assert.Equal(t, tc.data, fileContent, "File content should match the provided data")
			}
		})
	}
}

func TestAddFileOfSize(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		size   int64
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "random.txt",
			size: 1024,
		},
		{
			name: "Nested relative path",
			path: "nested/random.txt",
			size: 512,
		},
		{
			name: "Empty file",
			path: "empty.txt",
			size: 0,
		},
		{
			name:   "Invalid absolute path",
			path:   "//invalid/random.txt",
			size:   1024,
			errMsg: "path //invalid/random.txt is absolute. All FSUtil paths are relative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fs.Close()

			err = fs.AddFileOfSize(tc.path, tc.size)
			if tc.errMsg != "" {
				require.Error(t, err, "AddFileOfSize should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddFileOfSize should not return an error")

				// Check if the file exists and has the correct size
				filePath := filepath.Join(fs.RootDir, tc.path)
				fileInfo, err := os.Stat(filePath)
				require.NoError(t, err, "File should be stat-able")
				assert.Equal(t, tc.size, fileInfo.Size(), "File size should match the provided size")
			}
		})
	}
}

func TestAddDir(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "random",
		},
		{
			name: "Nested relative path",
			path: "nested/random",
		},
		{
			name: "Empty",
			path: "",
		},
		{
			name:   "Invalid absolute path",
			path:   "//invalid/random",
			errMsg: "path //invalid/random is absolute. All FSUtil paths are relative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fs.Close()

			err = fs.AddDir(tc.path)
			if tc.errMsg != "" {
				require.Error(t, err, "AddDir should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddDir should not return an error")

				// Check if the file exists and has the correct size
				dirPath := filepath.Join(fs.RootDir, tc.path)
				dirInfo, err := os.Stat(dirPath)
				require.NoError(t, err, "File should be stat-able")
				assert.True(t, dirInfo.IsDir(), "File should be a directory")
			}
		})
	}
}

func TestAddFileOfSizeDeterministic(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		size   int64
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "deterministic.txt",
			size: 1024,
		},
		{
			name: "Nested relative path",
			path: "nested/deterministic.txt",
			size: 512,
		},
		{
			name: "Empty file",
			path: "empty.txt",
			size: 0,
		},
		{
			name:   "Invalid absolute path",
			path:   "//invalid/deterministic.txt",
			size:   1024,
			errMsg: "path //invalid/deterministic.txt is absolute. All FSUtil paths are relative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fs.Close()

			err = fs.AddFileOfSizeDeterministic(tc.path, tc.size)
			if tc.errMsg != "" {
				require.Error(t, err, "AddFileOfSizeDeterministic should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddFileOfSizeDeterministic should not return an error")

				// Check if the file exists and has the correct size
				filePath := filepath.Join(fs.RootDir, tc.path)
				fileInfo, err := os.Stat(filePath)
				require.NoError(t, err, "File should be stat-able")
				assert.Equal(t, tc.size, fileInfo.Size(), "File size should match the provided size")

				// Check if the content is deterministic (all zeros)
				fileContent, err := os.ReadFile(filePath)
				require.NoError(t, err, "File should be readable")
				for _, b := range fileContent {
					assert.Equal(t, byte(0), b, "File content should be all zeros")
				}
			}
		})
	}
}

func TestCanDirFS(t *testing.T) {
	fsUtil, err := NewFSUtil("test")
	require.NoError(t, err, "NewFSUtil should not return an error")
	defer fsUtil.Close()

	err = fsUtil.AddFileWithData("testfile.txt", []byte("Test data"))
	require.NoError(t, err, "AddFileWithData should not return an error")

	readOnlyFS := os.DirFS(fsUtil.RootDir)

	data, err := fs.ReadFile(readOnlyFS, "testfile.txt")
	require.NoError(t, err, "ReadFile should not return an error")
	assert.Equal(t, []byte("Test data"), data, "File content should match the provided data")
}

func TestFSUtil_Open(t *testing.T) {
	fsys, err := NewFSUtil("test")
	require.NoError(t, err, "NewFSUtil should not return an error")
	defer fsys.Close()

	// Create a test file in the root directory
	testFilePath := filepath.Join(fsys.RootDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("test"), 0644)
	require.NoError(t, err, "Creating a test file should not return an error")

	// Test opening an existing file
	file, err := fsys.Open("test.txt")
	require.NoError(t, err, "Open should not return an error")
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	require.NoError(t, err, "ReadAll should not return an error")
	assert.Equal(t, []byte("test"), content, "File content should match the expected value")

	// Test opening a non-existing file
	_, err = fsys.Open("non-existing.txt")
	assert.Error(t, err, "Open should return an error for non-existing files")
	assert.ErrorIs(t, err, os.ErrNotExist, "Open should return an ErrNotExist error for non-existing files")

	// Test opening an absolute path (not allowed with fsutil)
	_, err = fsys.Open("/tmp/abs/non-existing.txt")
	assert.Error(t, err, "Open should return an error for absolute paths")
}
