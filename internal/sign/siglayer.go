package sign

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	SignatureTagSuffix string = "sig" // SignatureTagSuffix is the suffix for a tag, follows the digest of the signed manifest
)

// ErrNoSignatureAnnotation indicates that a signature was not found in an annotation.
var ErrNoSignatureAnnotation = errors.New("no signature found in annotations")

// Signature handles a signature layer by providing access to it's data.
type Signature interface {

	// Annotations returns the annotations associated with this layer.
	Annotations() (map[string]string, error)

	// GetDescriptor returns the signature layer descriptor.
	GetDescriptor() ocispec.Descriptor

	// GetPayload fetches the opaque data that is being signed.
	// This will always return data when there is no error.
	GetPayload() ([]byte, error)

	// GetPayloadBase64 returns the payload, base64 encoded.
	GetPayloadBase64() (string, error)

	// GetKeyRetrieverForPayload returns a key retriever determined
	// by the metadata included in the signed annotations of a signature.
	GetKeyRetrieverForPayload() (KeyRetriever, error)

	// Path returns the local path of the signature.
	Path() string
}

var _ Signature = (*notarySigLayer)(nil)

// SigLayer represents an individual signature layer.
// It is managed by a descriptor of the layer, while the Payload is the
// raw signature layer. SigLayer implements Signature.
type notarySigLayer struct {
	desc ocispec.Descriptor

	// localPath is the local filesystem path for this signature layer
	localPath string

	// payload is the JSON payload that is signed.
	payload []byte
}

// GetDescriptor returns the signature layer descriptor.
func (l *notarySigLayer) GetDescriptor() ocispec.Descriptor {
	return l.desc
}

// Annotations returns the ?unsigned? annotations. Annotations implements Signature.
func (l *notarySigLayer) Annotations() (map[string]string, error) {
	m := make(map[string]string, len(l.desc.Annotations)+1)
	for k, v := range l.desc.Annotations {
		m[k] = v
	}
	return m, nil
}

// GetPayload returns the signed payload. In the case of a notary signature, it returns
// the signature envelope. Payload implements Signature.
func (l *notarySigLayer) GetPayload() ([]byte, error) {

	// if the payload has not been set, load it
	if l.payload == nil {
		// open the sig layer file discovered by digest
		notarySigLayerFile, err := os.Open(l.localPath)
		if err != nil {
			return nil, fmt.Errorf("opening sig layer file path = %s: %w", l.localPath, err)
		}
		defer notarySigLayerFile.Close()

		// read the sig layer file
		l.payload, err = io.ReadAll(notarySigLayerFile)
		if err != nil {
			return nil, fmt.Errorf("reading sig layer file: %w", err)
		}
	}

	return l.payload, nil
}

// GetPayloadBase64 returns the payload, base64 encoded. GetPayloadBase64 implements Signature.
func (l *notarySigLayer) GetPayloadBase64() (string, error) {
	payload, err := l.GetPayload()
	if err != nil {
		return "", fmt.Errorf("getting payload: %w", err)
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

// GetKeyRetrieverForPayload calls the correct KeyRetriever constructor determined by the signature's mediatype.
func (l *notarySigLayer) GetKeyRetrieverForPayload() (KeyRetriever, error) {

	payload, err := l.GetPayload()
	if err != nil {
		return nil, fmt.Errorf("getting signature payload, notarySigLayer = %v: %w", l.desc, err)
	}

	// We currently only support jose or cose formatted, certificate based signatures with Notary
	keyRetriever, err := NewCertRetriever(l.desc.MediaType, payload)
	if err != nil {
		return nil, fmt.Errorf("constructing key retriever, mediatype %s: %w", l.desc.MediaType, err)
	}

	return keyRetriever, nil
}

// Path returns the local path of the signature layer.
func (l *notarySigLayer) Path() string {
	return l.localPath
}

// NewNotarySigLayer returns a sig layer used for extracting signature information.  layerData can contain the raw
// signature layer bytes, or nil if GetPayload() is planned to be called later (which will load the data from sigsPath).
func NewNotarySigLayer(layerDescriptor ocispec.Descriptor, sigsPath string, layerData []byte) Signature {
	return &notarySigLayer{
		desc:      layerDescriptor,
		localPath: filepath.Join(sigsPath, layerDescriptor.Digest.Hex()),
		payload:   layerData,
	}
}
