package mirror

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	golog "log"
	"math/rand"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/google/go-containerregistry/pkg/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/data-tool/internal/archive"
	"github.com/act3-ai/data-tool/internal/print"
	"github.com/act3-ai/go-common/pkg/logger"
	"github.com/act3-ai/go-common/pkg/test"
)

func TestDeserialize(t *testing.T) {
	defer leaktest.Check(t)() //nolint:revive

	log := test.Logger(t, 0)
	ctx := logger.NewContext(context.Background(), log)

	rne := require.New(t).NoError

	// Set up a fake registry
	s := httptest.NewServer(registry.New())
	defer s.Close()
	u, err := url.Parse(s.URL)
	rne(err)

	dir := GetOrCreateTestDir(t)
	ociDir := filepath.Join(dir, "oci")
	cas, err := oci.New(ociDir)
	rne(err)

	// Populating the registry with a few images
	rng := rand.New(rand.NewSource(1))

	// index 1
	idx1, err := pushRandomIndex(ctx, cas, rng, "v2")
	rne(err)
	// idx1.Annotations = map[string]string{ocispec.AnnotationRefName: "something"}
	t.Log("idx1", idx1)

	index := ocispec.Index{
		Manifests: []ocispec.Descriptor{idx1},
	}
	indexData, err := json.Marshal(index)
	rne(err)
	rne(os.WriteFile(filepath.Join(ociDir, ocispec.ImageIndexFile), indexData, 0o600))

	// tar it up
	tape := filepath.Join(dir, "tape.tar")
	f, err := os.Create(tape)
	rne(err)
	rne(archive.TarToStream(ctx, os.DirFS(ociDir), f))

	// build up the action
	tAction := actions.NewTool("0.0.0")
	// create a config file that defines the registry as HTTP
	config := filepath.Join(dir, "config.yaml")
	CreateConfigWithRegHTTP(t, config, u.Host)
	// add it to the config files
	tAction.Config.ConfigFiles = []string{config}
	mAction := &Action{
		DataTool: tAction,
	}

	t.Run("missing blobs", func(t *testing.T) {
		// get a list of the blobs
		blobsDir := filepath.Join(ociDir, "blobs", "sha256")
		files, err := os.ReadDir(blobsDir)
		// choose the last blob to delete
		lastBlob := files[len(files)-1]
		rne(err)
		// delete the blob from the cas
		t.Logf("removing %s", filepath.Join(blobsDir, lastBlob.Name()))
		err = os.Remove(path.Join(blobsDir, lastBlob.Name()))

		rne(err)
		// create a new tar/tape file
		missingTape := filepath.Join(dir, "missing-tape.tar")
		f, err := os.Create(missingTape)
		rne(err)
		rne(archive.TarToStream(ctx, os.DirFS(ociDir), f))

		// run deserialize with that
		deserialize := Deserialize{
			Action:     mAction,
			Strict:     false,
			BufferSize: 512 * 1024,
		}
		t.Logf(`Commands to help with debugging:\nmkdir "%[1]s/oci"; tar xvf "%[1]s/tape.tar" -C "%[1]s/oci"; ace-dt oci tree -d "%[1]s/oci"`, dir)

		dest := u.Host + "/mirror:sync-1"

		err = deserialize.Run(ctx, missingTape, dest)

		assert.ErrorContains(t, err, "missing blobs")
		t.Logf("re-adding %s", filepath.Join(blobsDir, lastBlob.Name()))
		rne(f.Close())
	})

	t.Run("basic", func(t *testing.T) {
		deserialize := Deserialize{
			Action:     mAction,
			Strict:     false,
			BufferSize: 512 * 1024,
		}

		t.Logf(`Commands to help with debugging:\nmkdir "%[1]s/oci"; tar xvf "%[1]s/tape.tar" -C "%[1]s/oci"; ace-dt oci tree -d "%[1]s/oci"`, dir)

		dest := u.Host + "/mirror:sync-1"
		err = deserialize.Run(ctx, tape, dest)
		rne(err)
	})

	t.Run("referrers", func(t *testing.T) {
		tmp := GetOrCreateTestDir(t)
		cas2, err := oci.New(ociDir)
		rne(err)

		// create a manifest with a referrer and a referrer to that referrer
		manifestDescriptor, err := pushRandomManifest(ctx, cas2, rng, nil, "original", nil)
		rne(err)
		referrer, err := pushRandomManifest(ctx, cas2, rng, &manifestDescriptor, "referrer", nil)
		rne(err)
		referrer2, err := pushRandomManifest(ctx, cas2, rng, &referrer, "referrer2", nil)
		rne(err)
		filename := path.Join(tmp, "referrer-tape.tar")
		tf, err := os.Create(filename)
		rne(err)
		tw := tar.NewWriter(tf)

		// write the oci-layout file
		lo := ocispec.ImageLayout{
			Version: ocispec.ImageLayoutVersion,
		}
		b, err := json.Marshal(lo)
		rne(err)
		rne(writeFileToTar(tw, ocispec.ImageLayoutFile, b))

		// write the index as it would exist without the referrer to the tar file
		idx := ocispec.Index{
			Manifests: []ocispec.Descriptor{manifestDescriptor},
		}
		idxData, err := json.Marshal(idx)
		rne(err)
		rne(writeFileToTar(tw, ocispec.ImageIndexFile, idxData))

		// write the main manifest
		var m ocispec.Manifest
		maniBytes, err := content.FetchAll(ctx, cas2, manifestDescriptor)
		rne(err)
		rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", manifestDescriptor.Digest.Algorithm(), manifestDescriptor.Digest.Hex()), maniBytes))
		// unmarshal the  manifest bytes so we can extract and write the layers to the tar file
		rne(json.Unmarshal(maniBytes, &m))
		// append the config to the layers slice so it gets written
		m.Layers = append(m.Layers, m.Config)
		// write the manifest layers
		for _, layer := range m.Layers {
			layerBytes, err := content.FetchAll(ctx, cas2, layer)
			rne(err)
			rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", layer.Digest.Algorithm(), layer.Digest.Hex()), layerBytes))
		}

		// write the referrer manifest
		referrerBytes, err := content.FetchAll(ctx, cas2, referrer)
		rne(err)
		rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", referrer.Digest.Algorithm(), referrer.Digest.Hex()), referrerBytes))
		// get and write the layers of the referrer manifest
		rne(json.Unmarshal(referrerBytes, &m))
		// append the config to the layers slice so it gets written
		m.Layers = append(m.Layers, m.Config)
		// write the manifest layers
		for _, layer := range m.Layers {
			layerBytes, err := content.FetchAll(ctx, cas2, layer)
			rne(err)
			rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", layer.Digest.Algorithm(), layer.Digest.Hex()), layerBytes))
		}

		// write a second referrer manifest
		// write the referrer manifest
		referrer2Bytes, err := content.FetchAll(ctx, cas2, referrer2)
		rne(err)
		rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", referrer2.Digest.Algorithm(), referrer2.Digest.Hex()), referrer2Bytes))
		// get and write the layers of the referrer manifest
		rne(json.Unmarshal(referrer2Bytes, &m))
		// append the config to the layers slice so it gets written
		m.Layers = append(m.Layers, m.Config)
		// write the manifest layers
		for _, layer := range m.Layers {
			layerBytes, err := content.FetchAll(ctx, cas2, layer)
			rne(err)
			rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", layer.Digest.Algorithm(), layer.Digest.Hex()), layerBytes))
		}

		// write the second index.json WITH the referrer manifest
		followUpIdx := ocispec.Index{
			Manifests: []ocispec.Descriptor{manifestDescriptor, referrer, referrer2},
		}

		followUpIdxData, err := json.Marshal(followUpIdx)
		rne(err)
		rne(writeFileToTar(tw, ocispec.ImageIndexFile, followUpIdxData))

		m2Action := &Action{
			DataTool:  tAction,
			Recursive: true,
		}
		deserialize := Deserialize{
			Action: m2Action,
			Strict: true,
		}
		dest := u.Host + "/testreferrer:sync-1"
		err = deserialize.Run(ctx, filename, dest)
		rne(err)
		// verify that the referrer exists
		cas3, err := remote.NewRepository(dest)
		rne(err)
		var manif ocispec.Manifest
		cas3.PlainHTTP = true
		_, rc, err := oras.Fetch(ctx, cas3, strings.Join([]string{u.Host + "/testreferrer", referrer.Digest.String()}, "@"), oras.DefaultFetchOptions)
		rne(err)
		decoder := json.NewDecoder(rc)
		err = decoder.Decode(&manif)
		rne(err)
		// verify that the referrer subject matches the original manifest descriptor
		assert.Equal(t, &manifestDescriptor, manif.Subject)
	})

	t.Run("referrers with more than 2 indexes", func(t *testing.T) {
		tmp := GetOrCreateTestDir(t)
		cas2, err := oci.New(ociDir)
		rne(err)

		// create a manifest with a referrer and a referrer to that referrer
		manifestDescriptor, err := pushRandomManifest(ctx, cas2, rng, nil, "original", nil)
		rne(err)
		referrer, err := pushRandomManifest(ctx, cas2, rng, &manifestDescriptor, "referrer", nil)
		rne(err)
		referrer2, err := pushRandomManifest(ctx, cas2, rng, &referrer, "referrer2", nil)
		rne(err)
		filename := path.Join(tmp, "referrer-tape.tar")
		tf, err := os.Create(filename)
		rne(err)
		tw := tar.NewWriter(tf)

		// write the oci-layout file
		lo := ocispec.ImageLayout{
			Version: ocispec.ImageLayoutVersion,
		}
		b, err := json.Marshal(lo)
		rne(err)
		rne(writeFileToTar(tw, ocispec.ImageLayoutFile, b))

		// write the index as it would exist without the referrer to the tar file
		idx := ocispec.Index{
			Manifests: []ocispec.Descriptor{manifestDescriptor},
		}
		idxData, err := json.Marshal(idx)
		rne(err)
		rne(writeFileToTar(tw, ocispec.ImageIndexFile, idxData))

		// write the main manifest
		var m ocispec.Manifest
		maniBytes, err := content.FetchAll(ctx, cas2, manifestDescriptor)
		rne(err)
		rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", manifestDescriptor.Digest.Algorithm(), manifestDescriptor.Digest.Hex()), maniBytes))
		// unmarshal the  manifest bytes so we can extract and write the layers to the tar file
		rne(json.Unmarshal(maniBytes, &m))
		// append the config to the layers slice so it gets written
		m.Layers = append(m.Layers, m.Config)
		// write the manifest layers
		for _, layer := range m.Layers {
			layerBytes, err := content.FetchAll(ctx, cas2, layer)
			rne(err)
			rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", layer.Digest.Algorithm(), layer.Digest.Hex()), layerBytes))
		}

		// write the referrer manifest
		referrerBytes, err := content.FetchAll(ctx, cas2, referrer)
		rne(err)
		rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", referrer.Digest.Algorithm(), referrer.Digest.Hex()), referrerBytes))
		// get and write the layers of the referrer manifest
		rne(json.Unmarshal(referrerBytes, &m))
		// append the config to the layers slice so it gets written
		m.Layers = append(m.Layers, m.Config)
		// write the manifest layers
		for _, layer := range m.Layers {
			layerBytes, err := content.FetchAll(ctx, cas2, layer)
			rne(err)
			rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", layer.Digest.Algorithm(), layer.Digest.Hex()), layerBytes))
		}

		// write the second index.json WITH the first referrer manifest
		followUpIdx := ocispec.Index{
			Manifests: []ocispec.Descriptor{manifestDescriptor, referrer},
		}

		followUpIdxData, err := json.Marshal(followUpIdx)
		rne(err)
		rne(writeFileToTar(tw, ocispec.ImageIndexFile, followUpIdxData))

		// write a second referrer manifest
		// write the referrer manifest
		referrer2Bytes, err := content.FetchAll(ctx, cas2, referrer2)
		rne(err)
		rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", referrer2.Digest.Algorithm(), referrer2.Digest.Hex()), referrer2Bytes))
		// get and write the layers of the referrer manifest
		rne(json.Unmarshal(referrer2Bytes, &m))
		// append the config to the layers slice so it gets written
		m.Layers = append(m.Layers, m.Config)
		// write the manifest layers
		for _, layer := range m.Layers {
			layerBytes, err := content.FetchAll(ctx, cas2, layer)
			rne(err)
			rne(writeFileToTar(tw, fmt.Sprintf("blobs/%s/%s", layer.Digest.Algorithm(), layer.Digest.Hex()), layerBytes))
		}

		// write the third index.json WITH both referrer manifests
		followUpIdx2 := ocispec.Index{
			Manifests: []ocispec.Descriptor{manifestDescriptor, referrer, referrer2},
		}

		followUpIdxData2, err := json.Marshal(followUpIdx2)
		rne(err)
		rne(writeFileToTar(tw, ocispec.ImageIndexFile, followUpIdxData2))

		m2Action := &Action{
			DataTool:  tAction,
			Recursive: true,
		}
		deserialize := Deserialize{
			Action: m2Action,
			Strict: false,
		}
		dest := u.Host + "/testreferrer:sync-1"
		err = deserialize.Run(ctx, filename, dest)
		rne(err)
		// verify that the referrer exists
		cas3, err := remote.NewRepository(dest)
		rne(err)
		var manif ocispec.Manifest
		cas3.PlainHTTP = true
		_, rc, err := oras.Fetch(ctx, cas3, strings.Join([]string{u.Host + "/testreferrer", referrer.Digest.String()}, "@"), oras.DefaultFetchOptions)
		rne(err)
		decoder := json.NewDecoder(rc)
		err = decoder.Decode(&manif)
		rne(err)
		// verify that the referrer subject matches the original manifest descriptor
		assert.Equal(t, &manifestDescriptor, manif.Subject)

		// test with strict mode (should fail)
		m2Action = &Action{
			DataTool:  tAction,
			Recursive: true,
		}
		deserialize = Deserialize{
			Action: m2Action,
			Strict: true,
		}
		dest = u.Host + "/testreferrer:sync-2"
		err = deserialize.Run(ctx, filename, dest)
		assert.ErrorContains(t, err, "deserializing: expected 2 index.json files in Strict mode but received 3")
	})
	/*
		t.Run("Cached Storage Fetch Missing Manifest", func(t *testing.T) {
			desc := ocispec.Descriptor{
				MediaType: "application/vnd.oci.image.index.v1+json",
				Digest:    idx1.Digest,
				Size:      idx1.Size,
			}

			err = cas.Delete(ctx, desc)
			rne(err)

			// create the cached storage
			var s content.Storage = cas
			rf := newCachedStorage(s)

			// try to fetch
			_, err = rf.Fetch(ctx, desc)
			assert.ErrorContains(t, err, "getting manifest as blob")
		})

		t.Run("Cached Storage Blob Push", func(t *testing.T) {

			n := rng.Intn(100) + 1
			data := make([]byte, n)
			_, err := rng.Read(data)
			rne(err)
			desc := content.NewDescriptorFromBytes("", data)
			r := bytes.NewReader(data)

			// create the cached storage
			var s content.Storage = cas
			rf := newCachedStorage(s)

			// push the blob
			err = rf.Push(ctx, desc, r)
			rne(err)

			// retrieve the blob
			exists, err := s.Exists(ctx, desc)
			rne(err)
			assert.True(t, exists)
		})
	*/
}

func FuzzDeserialize_Random(f *testing.F) {
	defer leaktest.Check(f)() //nolint

	ctx := context.Background()

	f.Add(int64(0))

	rne := require.New(f).NoError

	dir := f.TempDir()
	ociDir := filepath.Join(dir, "oci")
	cas, err := oci.New(ociDir)
	rne(err)

	// Randomly generate a few images
	// TODO set the seed in a Fuzz test
	rng := rand.New(rand.NewSource(0))
	idx1, err := pushRandomIndex(ctx, cas, rng, "idx1")
	rne(err)
	// TODO make a better uber index

	// add the tag annotation for OCI layout
	idx1.Annotations = map[string]string{
		ocispec.AnnotationRefName: "root",
	}

	index := ocispec.Index{
		Manifests: []ocispec.Descriptor{idx1},
	}
	indexData, err := json.Marshal(index)
	rne(err)
	rne(os.WriteFile(filepath.Join(ociDir, "index.json"), indexData, 0o600))

	{
		f.Logf(`To see the OCI tree run: ace-dt oci tree --oci-layout %s:root`, dir)
		store, err := oci.NewFromFS(ctx, os.DirFS(ociDir))
		rne(err)

		out := &bytes.Buffer{}
		rne(print.Successors(ctx, out, store, idx1, print.Options{}))
		f.Logf("\n%s\n", out.String())
	}

	// add extraneous files
	for i := 0; i < 3; i++ {
		rne(os.WriteFile(filepath.Join(ociDir, fmt.Sprintf("bogus-%d", i)), []byte("bogus content"), 0o666))
	}

	f.Fuzz(func(t *testing.T, seed int64) {
		deserializeRandom(t, seed, os.DirFS(ociDir))
	})
}

func deserializeRandom(tb testing.TB, seed int64, fsys fs.FS) {
	tb.Helper()
	// defer leaktest.Check(t)()

	log := test.Logger(tb, 0)

	rne := require.New(tb).NoError

	ctx := logger.NewContext(context.Background(), log)

	// collect all the files
	var files []string
	rne(fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	}))

	// We want the order OS agnostic at this point
	sort.Strings(files)

	// TODO Test error handling.  We also want to hit some errors by omitting files.  We might want to even write out files in the wrong format.

	// We want to be able to re-order the tar file entries (blobs, index.json, and oci-layout).
	// shuffle the file order
	rng := rand.New(rand.NewSource(seed))

	// duplicate some of the files for fun.
	for i := 0; i < 3; i++ {
		files = append(files, files[rng.Intn(len(files))])
	}

	rng.Shuffle(len(files), func(i, j int) {
		files[j], files[i] = files[i], files[j]
	})

	// aid in debugging
	tb.Logf("File processing order:\n%s", strings.Join(files, "\n"))

	tape := filepath.Join(tb.TempDir(), "tape.tar")
	tapeFile, err := os.Create(tape)
	rne(err)

	tw := tar.NewWriter(tapeFile)

	for _, filename := range files {
		addFile(tb, tw, fsys, filename)
	}

	// TODO ensure these are called even if the t.Fatal() is called (which exits the goroutine). t.Cleanup()?
	rne(tw.Close())
	rne(tapeFile.Close())

	// Set up a fake registry
	s := httptest.NewServer(registry.New(registry.Logger(golog.New(io.Discard, "", 0))))
	defer s.Close()
	u, err := url.Parse(s.URL)
	rne(err)

	dir := GetOrCreateTestDir(tb)
	// build up the action
	tAction := actions.NewTool("0.0.0")
	// create a config file that defines the registry as HTTP
	config := filepath.Join(dir, "config.yaml")
	CreateConfigWithRegHTTP(tb, config, u.Host)
	// add it to the config files
	tAction.Config.ConfigFiles = []string{config}
	mAction := &Action{
		DataTool: tAction,
	}

	deserialize := Deserialize{
		Action: mAction,
		Strict: false,
	}

	// Run the DUT
	destination := u.Host + "/high/mirror:sync-1"
	err = deserialize.Run(ctx, tape, destination)
	rne(err)
}

func addFile(tb testing.TB, tw *tar.Writer, fsys fs.FS, filename string) {
	tb.Helper()
	rne := require.New(tb).NoError
	f, err := fsys.Open(filename)
	rne(err)
	defer f.Close()
	stat, err := f.Stat()
	rne(err)

	rne(tw.WriteHeader(&tar.Header{
		Name:     filename,
		Size:     stat.Size(),
		Typeflag: tar.TypeReg,
	}))

	_, err = io.Copy(tw, f)
	rne(err)
}

func writeFileToTar(tw *tar.Writer, filename string, data []byte) error {
	if err := tw.WriteHeader(&tar.Header{
		Name:     filename,
		Size:     int64(len(data)),
		Mode:     0666,
		ModTime:  time.Unix(0, 0).UTC(),
		Typeflag: tar.TypeReg,
	}); err != nil {
		return fmt.Errorf("writing tar header for %s: %w", filename, err)
	}
	_, err := tw.Write(data)
	if err != nil {
		return fmt.Errorf("writing tar data for %s: %w", filename, err)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flushing tar data for %s: %w", filename, err)
	}
	return nil
}
