// Package encoding implements the protocol used to serialize and deserialize data across the wire.
package encoding

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"

	"git.act3-ace.com/ace/go-common/pkg/ioutil"
)

const blobsDir = "blobs"

// when reading the checkpoint from file we add digests to the "existing" map until we reach a offset that is higher than action.ResumeFromOffset

// OCILayoutSerializer handles writing of different types of data to the tar writer.
type OCILayoutSerializer struct {
	w          io.Writer
	tw         *tar.Writer
	blobsDir   bool                          // blobsDir created
	algorithms map[digest.Algorithm]struct{} // algorithms that have already been created in the blobs directory in the tar archive

	count  *ioutil.WriterCounter
	ledger *json.Encoder

	existingBlobs map[digest.Digest]ocispec.Descriptor
}

// NewOCILayoutSerializer creates a new serializer.
func NewOCILayoutSerializer(dest io.Writer) *OCILayoutSerializer {
	return &OCILayoutSerializer{
		w:             dest,
		tw:            tar.NewWriter(dest),
		blobsDir:      false,
		algorithms:    make(map[digest.Algorithm]struct{}),
		existingBlobs: make(map[digest.Digest]ocispec.Descriptor),
	}
}

// NewOCILayoutSerializerWithLedger serialized data to dest and writes the ledger to ledger.
func NewOCILayoutSerializerWithLedger(dest, ledger io.Writer) *OCILayoutSerializer {
	cw := new(ioutil.WriterCounter)
	serializer := NewOCILayoutSerializer(io.MultiWriter(dest, cw))
	serializer.count = cw
	serializer.ledger = json.NewEncoder(ledger)
	return serializer
}

// Close will close the serializer.
func (ow *OCILayoutSerializer) Close() error {
	return ow.tw.Close()
}

// SaveOCILayout write the OCI layout file to the tar stream.
func (ow *OCILayoutSerializer) SaveOCILayout() error {
	lo := ocispec.ImageLayout{
		Version: ocispec.ImageLayoutVersion,
	}

	b, err := json.Marshal(lo)
	if err != nil {
		return fmt.Errorf("error marshalling oci-layout: %w", err)
	}

	return ow.writeFileBytes(ocispec.ImageLayoutFile, b)
}

// SaveBlob writes a blob to the tar archive.
// This can be a layer, config, or manifest.
func (ow *OCILayoutSerializer) SaveBlob(ctx context.Context, fetcher content.Fetcher, blob ocispec.Descriptor) error {
	if _, ok := ow.existingBlobs[blob.Digest]; ok {
		// already written or denoted as skipped
		return nil
	}

	r, err := fetcher.Fetch(ctx, blob)
	if err != nil {
		return fmt.Errorf("fetch blob: %w", err)
	}
	defer r.Close()

	if err := ow.writeBlob(blob, r); err != nil {
		return fmt.Errorf("writing layer: %w", err)
	}

	// so we never write this blob again to this archive
	ow.SkipBlob(blob)

	return nil
}

// SaveIndex writes out the top level index.json file.
func (ow *OCILayoutSerializer) SaveIndex(index ocispec.Index) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return ow.writeFileBytes(ocispec.ImageIndexFile, data)
}

// SkipBlob tells the serializer to never write a blob with the given digest.
func (ow *OCILayoutSerializer) SkipBlob(desc ocispec.Descriptor) {
	ow.existingBlobs[desc.Digest] = desc
}

// writeFileBytes writes the file (as small slice of bytes) to the tar archive.
func (ow *OCILayoutSerializer) writeFileBytes(filename string, data []byte) error {
	return ow.createFileEntry(filename, int64(len(data)), bytes.NewReader(data))
}

// writeBlob writes the blob from the io.Reader to the tar archive.
func (ow *OCILayoutSerializer) writeBlob(desc ocispec.Descriptor, r io.Reader) error {
	// ensure we have the blobs directory
	if !ow.blobsDir {
		if err := ow.createDirEntry(blobsDir); err != nil {
			return err
		}
		ow.blobsDir = true
	}

	// ensure we have the algorithm directory
	alg := desc.Digest.Algorithm()
	if _, ok := ow.algorithms[alg]; !ok {
		if err := ow.createDirEntry(path.Join(blobsDir, alg.String())); err != nil {
			return err
		}
		ow.algorithms[alg] = struct{}{}
	}

	// verify the read content
	vr := content.NewVerifyReader(r, desc)

	if err := ow.createFileEntry(path.Join(blobsDir, alg.String(), desc.Digest.Encoded()), desc.Size, vr); err != nil {
		return err
	}

	if err := vr.Verify(); err != nil {
		return fmt.Errorf("verifing blob: %w", err)
	}

	// Handle the checkpoint ledger
	if ow.count != nil && ow.ledger != nil {
		// might not be necessary
		if err := ow.tw.Flush(); err != nil {
			return fmt.Errorf("ledger flushing tar: %w", err)
		}

		// prepare the descriptor (we do not want to modify the function's desc.Annotations)
		d := desc
		d.Annotations = make(map[string]string, len(desc.Annotations)+1)
		for k, v := range desc.Annotations {
			d.Annotations[k] = v
		}
		d.Annotations[AnnotationArchiveOffset] = fmt.Sprint(*ow.count)

		if err := ow.ledger.Encode(d); err != nil {
			return fmt.Errorf("writing ledger: %w", err)
		}
	}

	return nil
}

// createFileEntry is a low level function for writing a file to the tar stream.
func (ow *OCILayoutSerializer) createFileEntry(filename string, size int64, r io.Reader) error {
	hdr := tar.Header{
		Name:     filename,
		Size:     size,
		Mode:     0o666,
		ModTime:  time.Unix(0, 0).UTC(),
		Typeflag: tar.TypeReg,
	}

	if err := ow.tw.WriteHeader(&hdr); err != nil {
		return fmt.Errorf("writing file header: %w", err)
	}

	// f, err := os.Create("/tmp/thefile")
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()
	// r = io.TeeReader(r, f)

	n, err := io.Copy(ow.tw, r)
	if err != nil {
		return fmt.Errorf("copying data into archive: %w", err)
	}
	// covering size mismatch from descriptor to what we actually downloaded
	if n != size {
		return fmt.Errorf("copied %d B but expected %d B: %w", n, size, io.ErrShortWrite)
	}

	return nil
}

// createDirEntry is a low level function for writing a directory entry to the tar stream.
func (ow *OCILayoutSerializer) createDirEntry(dirname string) error {
	hdr := tar.Header{
		Name:     dirname,
		Mode:     0o777,
		ModTime:  time.Unix(0, 0).UTC(),
		Typeflag: tar.TypeDir,
	}

	if err := ow.tw.WriteHeader(&hdr); err != nil {
		return fmt.Errorf("writing directory header: %w", err)
	}

	return nil
}
