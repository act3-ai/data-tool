package sign

import (
	"context"
	"crypto"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	tlog "gitlab.com/act3-ai/asce/go-common/pkg/test"
)

// TODO: We'll likely have to update the bottle's signature if we change
// the payload to the actual descriptor (which previously used the notation
// example but with the subject digest)

// bottle
const testBottlePath = "testdata/testbottle"

var testBottleManDesc = ocispec.Descriptor{
	MediaType: ocispec.MediaTypeImageManifest,
	Digest:    "sha256:1ab73444bf20635135ac8a289401a10c9990023e6f37993d93172974fe87704d",
	Size:      453,
}

// keys
// const testBottleKeysPath = "testdata/testkeys"
// const testPubKey = "test.pub"
// const testPrivKey = "test.key"
// const testCert = "test.crt"

// Test_SignNotaryCert creates a new bottle and signing keys (with cert), then
// signs it using the private key and notary certificate.
//
// Creating a bottle and signing keys from scratch prevents us from affecting
// the existing testdata used in verification tests as the action of signing
// ultimately writes the signatures to disk.
func Test_SignNotaryCert(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	// create a fresh bottle - we don't want to risk overwriting the testdata.
	testBottleDir := t.TempDir()
	manifestDesc := CreateSampleBottle(t, testBottleDir)

	// get priv key and cert - we can safely use existing keys from testdata.
	keysDir := filepath.Join("testdata", "testkeys")
	privKeyPath := filepath.Join(keysDir, "test.key")
	certPath := filepath.Join(keysDir, "test.crt")
	pkp := NewFilePrivateKeyProvider(privKeyPath, certPath)

	// create an empty sigsHandler
	sigsHandler := NotarySignatures{
		Subject:      manifestDesc,
		SigManifests: []SigsManifestHandler{},
		LocalPath:    testBottleDir,
		HashFunc:     crypto.SHA256,
	}

	// sign
	unsignedAnnos := make(map[string]string, 3)
	unsignedAnnos[AnnotationUserID] = "testUserIdentity"
	unsignedAnnos[AnnotationVerifyAPI] = "cert-basic"
	unsignedAnnos[AnnotationKeyID] = "testKeyID"

	if err := sigsHandler.Sign(ctx, pkp, unsignedAnnos, nil); err != nil {
		t.Fatalf("signing test bottle: %v", err)
	}
}

// Test_VerifyNotaryCert uses existing testdata to validate a bottle signature signed
// with a notary certificate.
func Test_VerifyNotaryCert(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	// load signatures from testdata
	sigsHandler, err := LoadLocalSignatures(ctx, testBottleManDesc, filepath.Join(testBottlePath, ".signature"))
	if err != nil {
		t.Fatalf("loading signatures handler: %v", err)
	}

	// verify
	pass, err := sigsHandler.Verify(ctx)
	if err != nil {
		t.Errorf("Testing Bottle failed verification with an error")
	}

	if pass {
		// signature verification passed.
		t.Logf("Testing Bottle passed verification.")
	} else {
		// signature verification failed.
		t.Errorf("Testing Bottle failed verification.")
	}
	for _, fp := range sigsHandler.VerifiedSignatures() {
		t.Logf("    Verified Signature Digest: %s", fp)
	}
	for _, fp := range sigsHandler.FailedSignatures() {
		t.Errorf("    Failed Signature Digest: %s", fp)
	}
	if err != nil {
		t.Errorf("Test Bottle verify: %v", err) // verification failed.
	}
}

// Builds a sample bottle writing its data to the provided directory, returning a descriptor of the manifest.
func CreateSampleBottle(t *testing.T, dir string) ocispec.Descriptor {
	t.Helper()

	err := os.MkdirAll(filepath.Join(dir, ".dt"), 0o777)
	if err != nil {
		t.Fatalf("creating .dt dir: %s", err)
	}

	// build and write a layer
	layerDesc := BuildWriteLayer(t, dir)

	// build and write the config
	configDesc := BuildWriteConfig(t, []ocispec.Descriptor{layerDesc}, dir)

	// build and write the manifest
	desc := BuildWriteManifest(t, mediatype.MediaTypeBottle, configDesc, []ocispec.Descriptor{layerDesc}, map[string]string{}, filepath.Join(dir, ".dt", ".manifest.json"))

	return desc
}

// CreateSignedBottle creates and signs a bottle for testing purposes. All internal errors are fatal.
func CreateSignedBottle(t *testing.T, testBottleDir string) {
	t.Helper()
	ctx := context.Background()
	log := tlog.Logger(t, -2)
	ctx = logger.NewContext(ctx, log)

	manifestDesc := CreateSampleBottle(t, testBottleDir)

	certPair := MakeEcdsaCertPair("ace-dt self signed signing cert", nil, nil)
	pkp := filePrivateKeyProvider{ // we can avoid writing to a file if we pre-populate the fields
		pKey: certPair.PrivateKey,
		cert: certPair.Cert,
	}

	sigsHandler, err := LoadLocalSignatures(ctx, manifestDesc, testBottleDir)
	if err != nil {
		t.Fatalf("loading signatures handler: %v", err)
	}

	unsignedAnnos := make(map[string]string, 3)
	unsignedAnnos[AnnotationUserID] = "testUserIdentity"
	unsignedAnnos[AnnotationVerifyAPI] = "cert-basic"
	unsignedAnnos[AnnotationKeyID] = "testKeyID"

	if err := sigsHandler.Sign(ctx, &pkp, unsignedAnnos, nil); err != nil {
		t.Fatalf("signing test bottle: %v", err)
	}
}

// BuildWriteConfig creates a config with the given layers, writes it to path, and returns a
// descriptor for the generated config.
func BuildWriteConfig(t *testing.T, layerDescs []ocispec.Descriptor, path string) ocispec.Descriptor {
	t.Helper()
	// build the config

	layerDigests := make([]digest.Digest, 0, len(layerDescs))
	for _, desc := range layerDescs {
		layerDigests = append(layerDigests, desc.Digest)
	}

	rootFS := ocispec.RootFS{
		Type:    "layers",
		DiffIDs: layerDigests,
	}

	// build the config
	config := &ocispec.Image{
		RootFS: rootFS,
	}

	// encode the config
	configBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshaling conifg: %s", err)
	}

	// create the config file
	configFile, err := os.Create(filepath.Join(path, ".dt", ".config.json"))
	if err != nil {
		t.Fatalf("Creating config file: %s", err)
	}
	defer configFile.Close()

	// write to the config file
	_, err = configFile.Write(configBytes)
	if err != nil {
		t.Fatalf("Writing to config file: %s", err)
	}

	// return a descriptor for the config
	return ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(configBytes),
		Size:      int64(len(configBytes)),
	}
}

// BuildWriteLayer creates and writes an arbitrary text file to dest, returning
// an oci descriptor for the file.
func BuildWriteLayer(t *testing.T, dest string) ocispec.Descriptor {
	t.Helper()

	// create a random layer file
	layerFile, err := os.Create(filepath.Join(dest, "tempLayer.txt"))
	if err != nil {
		t.Fatal("Creating layer file")
	}
	defer layerFile.Close()

	// write something to the layer file
	text := []byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.")
	_, err = layerFile.Write(text)
	if err != nil {
		t.Fatalf("Writing to layer file: %s", err)
	}

	// return a descriptor for the layer
	return ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayer,
		Digest:    digest.FromBytes(text),
		Size:      int64(len(text)),
	}
}

// BuildWriteManifest puts together a simple manifest, writing it to dest, and returing a descriptor of the manifest.
func BuildWriteManifest(t *testing.T, artifactType string, configDesc ocispec.Descriptor, layerDescs []ocispec.Descriptor, annotations map[string]string, dest string) ocispec.Descriptor {
	t.Helper()

	// build the manifest
	manifest := &ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		MediaType:    ocispec.MediaTypeImageManifest,
		ArtifactType: artifactType,
		Config:       configDesc,
		Layers:       layerDescs,
		Annotations:  annotations,
	}

	// encode the manifest
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshaling signature configuration: %s", err)
	}

	manifestFile, err := os.Create(dest)
	if err != nil {
		t.Fatalf("creating signature manifest file: %s", err)
	}
	defer manifestFile.Close()

	// write manifest to the signautre directory
	_, err = manifestFile.Write(manifestBytes)
	if err != nil {
		t.Fatalf("writing signature manifest to file: %s", err)
	}

	return ocispec.Descriptor{
		MediaType: mediatype.MediaTypeBottle,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
}
