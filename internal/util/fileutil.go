// Package util implements general utility functionality for operations such as:
//
// - file operations.
// - simple http operations.
// - a set collection implementation.
// - an unlimited channel (buffered channel) implementation.
package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"
)

// CopyFile copies a file from src to dest.
func CopyFile(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Ensure destination path exists
	err = CreatePathForFile(dest)
	if err != nil {
		return fmt.Errorf("unable to create path for destination: %w", err)
	}

	destFile, err := os.Create(dest) // creates if file doesn't exist
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	if err := srcFile.Close(); err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	return destFile.Close()
}

// CopyToStream opens the input file specified by src and copies it to the stream
// specified by w.
func CopyToStream(src string, w io.Writer) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()
	size, err := io.Copy(w, srcFile) // check first var for number of bytes copied
	if err != nil {
		return size, fmt.Errorf("copying source file to stream: %w", err)
	}
	return size, srcFile.Close()
}

// DigestFile calculates a SHA256 hash for a given file path.
// TODO pass an argument to be able to change the digest algorithm.
func DigestFile(path string) (digest.Digest, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file for digesting: %w", err)
	}
	defer f.Close()
	dgst, err := digest.FromReader(f)
	if err != nil {
		return dgst, fmt.Errorf("digesting file: %w", err)
	}
	return dgst, f.Close()
}

// CreatePathForFile creates a directory if one does not exist that can contain the provided file path.
// no file is created, only the directory path is created.
func CreatePathForFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("create file path: %w", err)
	}
	return nil
}

// IsDirEmpty returns true, nil if a directory is empty
// from https://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func IsDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("opening dir: %w", err)
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // only one entry needs to be identified to determine emptiness
	if errors.Is(err, io.EOF) {
		return true, nil
	}

	if err != nil {
		return false, fmt.Errorf("reading directory names: %w", err)
	}
	return false, f.Close()
}
