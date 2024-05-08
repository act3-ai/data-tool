// Package oci facilitates transferring of git and git-lfs OCI artifacts.
package oci

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// Helper assists in pushing to or fetching from an OCI compliant registry.
type Helper struct {
	Target     oras.GraphTarget
	Tag        string
	FStore     *file.Store
	FStorePath string
}

// NewOCIHelper returns a ociHelper object used in pushing to or fetching from an OCI compliant registry.
// The OCI compliant registry must support the Referrer's API. This constructor initializes an oras filestore
// which is the caller's responsibility to close, may be done with sync.cleanup().
func NewOCIHelper(tmpDir string, target oras.GraphTarget, tag string) (*Helper, error) {

	fs, err := file.New(tmpDir)
	if err != nil {
		return &Helper{}, fmt.Errorf("initializing filestore: %w", err)
	}

	return &Helper{
		Target:     target,
		Tag:        tag,
		FStore:     fs,
		FStorePath: tmpDir,
	}, nil
}

// CopyLFSFromOCI copies a git-lfs file stored as an OCI layer, written to objDest.
//
// TODO: This is inefficient and not a good interface. See issue https://git.act3-ace.com/ace/data/tool/-/issues/504.
func (o *Helper) CopyLFSFromOCI(ctx context.Context, objDest string, layerDesc ocispec.Descriptor) error {
	log := logger.FromContext(ctx)

	// prepare destination
	oidDir := filepath.Dir(objDest)

	log.InfoContext(ctx, "initializing path to oid", "oidDir", oidDir)
	err := os.MkdirAll(oidDir, 0777)
	if err != nil {
		return fmt.Errorf("creating path to oid file: %w", err)
	}

	log.InfoContext(ctx, "creating oid file", "objDest", objDest)
	oidFile, err := os.Create(objDest)
	if err != nil {
		return fmt.Errorf("creating oid file: %w", err)
	}

	// download
	r, err := o.Target.Fetch(ctx, layerDesc)
	if err != nil {
		return fmt.Errorf("fetching LFS layer: %w", err)
	}

	n, err := io.Copy(oidFile, r)
	if err != nil {
		return fmt.Errorf("copying LFS layer to file: %w", err)
	}
	if n < layerDesc.Size {
		return fmt.Errorf("total bytes copied from LFS layer does not equal LFS layer size, layerSize: %d, copied: %d", layerDesc.Size, n)
	}

	if err := oidFile.Close(); err != nil {
		return fmt.Errorf("closing oid file: %w", err)
	}

	return nil
}
