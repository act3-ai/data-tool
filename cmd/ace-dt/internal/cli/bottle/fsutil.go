package bottle

import (
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/act3-ai/go-common/pkg/ioutil"
)

// FSUtil contains common utilities for working with a filesystem.
// NewFSUtil should be used to create a new instance.
// Using the struct directly is not recommended as it may not be initialized properly.
type FSUtil struct {
	RootDir string
	source  rand.Source
}

// NewFSUtil creates a new FSUtil instance. A temporary directory is created with the given prefix.
// The directory needs to be removed by the caller with the built in `Close` method.
// Defer `Close` is not recommended as it ignores any errors while closing the complex filesystem.
func NewFSUtil(prefix string) (*FSUtil, error) {
	return NewFSUtilWithSource(prefix, rand.NewSource(rand.Int63()))
}

// NewFSUtilWithSource is identical to NewFSUtil but allows a custom math/rand.Source to be used.
func NewFSUtilWithSource(prefix string, source rand.Source) (*FSUtil, error) {
	tempDir, err := os.MkdirTemp("", prefix+"*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	return &FSUtil{
		RootDir: tempDir,
		source:  source,
	}, nil
}

// Close removes the root directory.
func (f *FSUtil) Close() error {
	if err := os.RemoveAll(f.RootDir); err != nil {
		return fmt.Errorf("failed to remove root dir: %w", err)
	}
	return nil
}

// AddDir creates a directory.
// fPath is required to be a relative path.
func (f *FSUtil) AddDir(fPath string) error {
	fPath, err := f.joinRelative(fPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(fPath, 0o775); err != nil {
		return fmt.Errorf("failed to create dir %s: %w", fPath, err)
	}
	return nil
}

// AddFileWithData creates a file with the given data.
// fPath is required to be a relative path.
func (f *FSUtil) AddFileWithData(fPath string, data []byte) error {
	file, err := f.createPathAndFile(fPath)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name(), err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", file.Name(), err)
	}
	return nil
}

// AddFileOfSize creates a file with the given size, filled with random data.
// fPath is required to be a relative path.
func (f *FSUtil) AddFileOfSize(fPath string, size int64) error {
	file, err := f.createPathAndFile(fPath)
	if err != nil {
		return err
	}

	rng := rand.New(f.source)

	_, err = io.CopyN(file, rng, size)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name(), err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", file.Name(), err)
	}
	return nil
}

// AddFileOfSizeDeterministic creates a file with the given size, filled with deterministic data.
// fPath is required to be a relative path.
func (f *FSUtil) AddFileOfSizeDeterministic(fPath string, size int64) error {
	file, err := f.createPathAndFile(fPath)
	if err != nil {
		return err
	}

	zeroReader, err := ioutil.NewZeroReader(size)
	if err != nil {
		return fmt.Errorf("failed to create zero reader: %w", err)
	}

	_, err = io.Copy(file, zeroReader)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name(), err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", file.Name(), err)
	}
	return nil
}

// joinRelative joins the given path to the root dir after checking that the path is relative.
func (f *FSUtil) joinRelative(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("path %s is absolute. All FSUtil paths are relative", path)
	}
	return filepath.Join(f.RootDir, path), nil
}

// createPathAndFile creates the path and file.
// need to close file after use.
func (f *FSUtil) createPathAndFile(path string) (*os.File, error) {
	fPath, err := f.joinRelative(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(fPath), 0o775); err != nil {
		return nil, fmt.Errorf("failed to create dir %s: %w", fPath, err)
	}
	file, err := os.Create(fPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", fPath, err)
	}
	return file, nil
}

// Open implements the io/fs.FS interface.
// name is required to be a relative path.
func (f *FSUtil) Open(name string) (fs.File, error) {
	path, err := f.joinRelative(name)
	if err != nil {
		return nil, fmt.Errorf("failed to join relative path: %w", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}
