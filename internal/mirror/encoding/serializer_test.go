package encoding

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
)

func TestSerializerWriteOCILayout(t *testing.T) {
	rne := require.New(t).NoError

	// create destFile
	tmp := t.TempDir()
	destFile := filepath.Join(tmp, "test.tar")
	// open the destination file/tape carefully to append only
	dest, err := os.OpenFile(destFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer dest.Close()

	// setup the destination
	t.Logf("Creating serializer with destination %s", destFile)
	s, err := NewOCILayoutSerializer(dest, "")
	rne(err)
	defer s.Close()

	err = s.SaveOCILayout()
	rne(err)

	rne(s.Close())
	rne(dest.Close())
	// TODO verify that the OCI Layout file is in the correct format
	file, err := os.Open(destFile)
	rne(err)
	defer file.Close()
	tr := tar.NewReader(file)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		rne(err)
		if hdr.Name != "oci-layout" {
			rne(fmt.Errorf("error unexpected file: %s", hdr.Name))
		}
		t.Log(hdr.Name)
	}
	rne(file.Close())
}

func TestSerializerWriteBlobLayer(t *testing.T) {
	ctx := context.Background()

	rne := require.New(t).NoError

	// create destFile
	tmp := t.TempDir()
	destFile := filepath.Join(tmp, "test.tar")
	// open the destination file/tape carefully to append only
	dest, err := os.OpenFile(destFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer dest.Close()

	// setup the destination
	t.Logf("Creating serializer with destination %s", destFile)
	s, err := NewOCILayoutSerializer(dest, "")
	rne(err)
	defer s.Close()

	cas := memory.New()

	// a blob to the CAS
	data := []byte("some data for a blob")
	blob := content.NewDescriptorFromBytes(ocispec.MediaTypeImageLayerGzip, data)
	rne(cas.Push(ctx, blob, bytes.NewReader(data)))

	// DUT
	err = s.SaveBlob(ctx, cas, blob)
	rne(err)
	rne(s.Close())

	// verify that there is a blobs directory
	// verify that the layer exists in the tar file
	// "blobs", "blobs/sha256", "blobs/sha256/digest"
	file, err := os.Open(destFile)
	rne(err)
	defer file.Close()

	// populate our expected map
	expected := map[string]string{
		"blobs":                      "",
		path.Join("blobs", "sha256"): "",
		path.Join("blobs", "sha256", blob.Digest.Encoded()): "",
	}

	tr := tar.NewReader(file)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		delete(expected, hdr.Name)
	}
	if len(expected) != 0 {
		rne(fmt.Errorf("tar entry not found: %+v", expected))
	}

	rne(file.Close())
}

func TestSerializerWriteIndex(t *testing.T) {
	rne := require.New(t).NoError

	// create destFile
	tmp := t.TempDir()
	destFile := filepath.Join(tmp, "test.tar")
	// open the destination file/tape carefully to append only
	dest, err := os.OpenFile(destFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer dest.Close()

	// setup the destination
	t.Logf("Creating serializer with destination %s", destFile)
	s, err := NewOCILayoutSerializer(dest, "")
	rne(err)
	defer s.Close()
	rne(s.SaveIndex(ocispec.Index{}))
	rne(s.Close())
	rne(dest.Close())

	// verify that there is a blobs directory
	// verify that the layer exists in the tar file
	// "blobs", "blobs/sha256", "blobs/sha256/digest"
	file, err := os.Open(destFile)
	rne(err)
	defer file.Close()

	tr := tar.NewReader(file)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		assert.Equal(t, hdr.Name, ocispec.ImageIndexFile)
	}

	rne(file.Close())
}

func TestSerializerWithLedger(t *testing.T) {
	ctx := context.Background()

	rne := require.New(t).NoError

	// create destFile
	tmp := t.TempDir()
	destFile := filepath.Join(tmp, "test.tar")
	// open the destination file/tape carefully to append only
	dest, err := os.OpenFile(destFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer dest.Close()

	// set up the ledger
	l := filepath.Join(tmp, "ledger")
	ledger, err := os.OpenFile(l, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer ledger.Close()

	// setup the destination
	t.Logf("Creating serializer with destination %s", destFile)
	s, err := NewOCILayoutSerializerWithLedger(dest, ledger, "")
	rne(err)
	defer s.Close()

	cas := memory.New()

	// a blob to the CAS
	data := []byte("some data for a blob")
	blob := content.NewDescriptorFromBytes(ocispec.MediaTypeImageLayerGzip, data)
	rne(cas.Push(ctx, blob, bytes.NewReader(data)))

	// write the layer
	err = s.SaveBlob(ctx, cas, blob)
	rne(err)

	// close the files being written to
	rne(s.Close())
	rne(dest.Close())
	rne(ledger.Close())

	// verify that there is a blobs directory
	// verify that the layer exists in the tar file
	// "blobs", "blobs/sha256", "blobs/sha256/digest"
	file, err := os.Open(l)
	rne(err)
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		desc := ocispec.Descriptor{}
		err := decoder.Decode(&desc)
		if errors.Is(err, io.EOF) {
			break
		}
		rne(err)
		if desc.Digest != blob.Digest {
			rne(fmt.Errorf("unexpected digest received: %s. Expected: %s", desc.Digest, blob.Digest))
		}
	}
	rne(file.Close())
}

func TestSerializerWithLedgerAndGzipCompression(t *testing.T) {
	ctx := context.Background()

	rne := require.New(t).NoError

	// create destFile
	tmp := t.TempDir()
	destFile := filepath.Join(tmp, "test.tar.gz")
	// open the destination file/tape carefully to append only
	dest, err := os.OpenFile(destFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer dest.Close()
	// set up the ledger
	l := filepath.Join(tmp, "ledger")
	ledger, err := os.OpenFile(l, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer ledger.Close()

	// setup the destination
	t.Logf("Creating serializer with destination %s", destFile)
	s, err := NewOCILayoutSerializerWithLedger(dest, ledger, "gzip")
	rne(err)
	defer s.Close()

	cas := memory.New()

	// a blob to the CAS
	data := []byte("some data for a blob")
	blob := content.NewDescriptorFromBytes(ocispec.MediaTypeImageLayerGzip, data)
	rne(cas.Push(ctx, blob, bytes.NewReader(data)))

	// write the layer
	err = s.SaveBlob(ctx, cas, blob)
	rne(err)

	// close the files being written to
	rne(s.Close())
	rne(dest.Close())
	rne(ledger.Close())

	// verify that there is a blobs directory
	// verify that the layer exists in the tar file
	// "blobs", "blobs/sha256", "blobs/sha256/digest"
	file, err := os.Open(l)
	rne(err)
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		desc := ocispec.Descriptor{}
		err := decoder.Decode(&desc)
		if errors.Is(err, io.EOF) {
			break
		}
		rne(err)
		if desc.Digest != blob.Digest {
			rne(fmt.Errorf("unexpected digest received: %s. Expected: %s", desc.Digest, blob.Digest))
		}
	}
	rne(file.Close())

	// verify that dest file is gzip formatted
	file, err = os.Open(destFile)
	rne(err)
	defer file.Close()
	buf := make([]byte, 10)
	_, err = file.Read(buf)
	rne(err)
	// these are special header bytes that gzip sets to identify the file as gzip compressed.
	if len(buf) <= 1 || buf[0] != 0x1F && buf[1] != 0x8B {
		rne(fmt.Errorf("destfile is not gzip formatted: %x", buf))
	}
	rne(file.Close())
}

func TestSerializerWithLedgerAndZstdCompression(t *testing.T) {
	ctx := context.Background()

	rne := require.New(t).NoError

	// create destFile
	tmp := t.TempDir()
	destFile := filepath.Join(tmp, "test.tar.zst")
	// open the destination file/tape carefully to append only
	dest, err := os.OpenFile(destFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer dest.Close()
	// set up the ledger
	l := filepath.Join(tmp, "ledger")
	ledger, err := os.OpenFile(l, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	rne(err)
	defer ledger.Close()

	// setup the destination
	t.Logf("Creating serializer with destination %s", destFile)
	s, err := NewOCILayoutSerializerWithLedger(dest, ledger, "zstd")
	rne(err)
	defer s.Close()

	cas := memory.New()

	// a blob to the CAS
	data := []byte("some data for a blob")
	blob := content.NewDescriptorFromBytes(ocispec.MediaTypeImageLayerGzip, data)
	rne(cas.Push(ctx, blob, bytes.NewReader(data)))

	// write the layer
	err = s.SaveBlob(ctx, cas, blob)
	rne(err)

	// close the files being written to
	rne(s.Close())
	rne(dest.Close())
	rne(ledger.Close())

	// verify that there is a blobs directory
	// verify that the layer exists in the tar file
	// "blobs", "blobs/sha256", "blobs/sha256/digest"
	file, err := os.Open(l)
	rne(err)
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		desc := ocispec.Descriptor{}
		err := decoder.Decode(&desc)
		if errors.Is(err, io.EOF) {
			break
		}
		rne(err)
		if desc.Digest != blob.Digest {
			rne(fmt.Errorf("unexpected digest received: %s. Expected: %s", desc.Digest, blob.Digest))
		}
	}
	rne(file.Close())

	// verify that dest file is gzip formatted
	file, err = os.Open(destFile)
	rne(err)
	defer file.Close()
	buf := make([]byte, 10)
	_, err = file.Read(buf)
	rne(err)
	// these are special header bytes that zstd sets to identify the file as gzip compressed.
	if len(buf) <= 3 || buf[0] != 0x28 && buf[1] != 0xB5 && buf[2] != 0x2F && buf[3] != 0xFD {
		rne(fmt.Errorf("destfile is not zstd formatted: %x", buf))
	}
	rne(file.Close())
}
