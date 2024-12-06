package sign

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	notationreg "github.com/notaryproject/notation-go/registry"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/oci"
)

const (
	// SignatureFileExt is the file extension for a signature manifest.
	SignatureFileExt string = ".sig"
)

// LoadLocalSignatures loads a signature image manifest at a local path, returning a handler for the signature collection.
func LoadLocalSignatures(ctx context.Context, targetDescriptor ocispec.Descriptor, localPath string) (SigsHandler, error) {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Loading existing signatures")
	// attempt to load an existing signature manifest
	sigsManifestHandlers, err := SigManifestsFromPath(localPath, targetDescriptor.Digest)
	if err != nil {
		return nil, fmt.Errorf("loading signature image manifests: %w", err)
	}

	log.InfoContext(ctx, "Signatures loaded", "count", len(sigsManifestHandlers))
	// return a SigsHandler
	return &NotarySignatures{
		Subject:      targetDescriptor,
		SigManifests: sigsManifestHandlers,
		LocalPath:    localPath,
		HashFunc:     crypto.SHA256,
	}, nil
}

// NewEmptySigsManifest creates an empty sigs manifest for situations when no local sig.
func NewEmptySigsManifest() *SigsManifest {
	return &SigsManifest{
		Descriptor: ocispec.Descriptor{},
		Manifest:   ocispec.Manifest{},
		ManifestStatusInfo: oci.ManifestStatusInfo{
			Status:     oci.ManifestNotFound,
			StatusInfo: "New signature image manifest required",
			Error:      nil,
		},
	}
}

// SigManifestsFromPath returns a slice of SigsManifestHandlers for all signature image manifests referring to the subject, discovered by path.
func SigManifestsFromPath(localPath string, subjectDigest digest.Digest) ([]SigsManifestHandler, error) {
	manifestNames, err := ResolveSigManifestNames(localPath, subjectDigest)
	if err != nil {
		return nil, fmt.Errorf("loading signature manifests at path %s: %w", localPath, err)
	}

	// read the manifest file
	if len(manifestNames) == 0 {
		return []SigsManifestHandler{}, nil
	}
	sigsManifestHandlers := make([]SigsManifestHandler, 0, len(manifestNames))
	for _, manifestName := range manifestNames {
		manifestPath := filepath.Join(localPath, manifestName)
		manifestRaw, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("reading manifest file: %w", err)
		}
		// construct the manifest handler
		sigsManifestHandler := SigsManifestFromData(ocispec.MediaTypeImageManifest, manifestRaw)

		sigManifest := sigsManifestHandler.GetManifestData()

		// ensure supported signature type
		if sigManifest.Config.MediaType != notationreg.ArtifactTypeNotation {
			continue
		}

		// load the existing layers
		layers := sigsManifestHandler.GetManifestData().Layers
		layerMap := make(map[digest.Digest][]byte, len(layers))
		for _, layerDesc := range layers {
			rawLayer, err := os.ReadFile(filepath.Join(localPath, layerDesc.Digest.Hex()))
			if err != nil {
				return nil, fmt.Errorf("reading layer file: %w", err)
			}
			layerMap[layerDesc.Digest] = rawLayer

		}

		sigsManifestHandler.SetRawLayers(layerMap)
		err = sigsManifestHandler.SetLayerDescriptors(layers)
		if err != nil {
			return nil, fmt.Errorf("setting layer descriptors: %w", err)
		}

		sigsManifestHandlers = append(sigsManifestHandlers, sigsManifestHandler)
	}
	return sigsManifestHandlers, nil
}

// SigsManifestFromData constructs a SigsManifestHandler from data.
func SigsManifestFromData(contentType string, data []byte) SigsManifestHandler {
	switch contentType {
	case ocispec.MediaTypeImageManifest:
		var ociman ocispec.Manifest
		status := oci.ManifestOK
		statusInfo := ""
		err := json.Unmarshal(data, &ociman)
		if err != nil {
			status = oci.ManifestBadFormat
			statusInfo = "Failed to parse manifest info from data: " + err.Error()
			err = fmt.Errorf("failed to parse manifest info from data: %w", err)
		}
		return &SigsManifest{
			Descriptor: ocispec.Descriptor{
				Digest:    digest.FromBytes(data),
				Size:      int64(len(data)),
				MediaType: contentType,
			},
			Manifest: ociman,
			ManifestStatusInfo: oci.ManifestStatusInfo{
				Status:     status,
				StatusInfo: statusInfo,
				Error:      err,
			},
		}
	default:
		// return an empty
		return &SigsManifest{
			Descriptor: ocispec.Descriptor{},
			Manifest:   ocispec.Manifest{},
			ManifestStatusInfo: oci.ManifestStatusInfo{
				Status:     oci.ManifestUnsupportedType,
				StatusInfo: "Unsupported manifest media type: " + contentType,
				Error:      errors.New("Unsupported manifest media type " + contentType),
			},
		}
	}
}

// ResolveSigManifestName resolves the filename of a signature manifest.
// Ex: <algorithm>-<hash_hex>.sig.
func ResolveSigManifestName(targetDigest digest.Digest) string {
	return targetDigest.Algorithm().String() + "-" + targetDigest.Hex() + SignatureFileExt
}

// ResolveSigManifestNames creates and returns a list of filenames of signature manifests containing a subject field
// that points to the provided subject digest.
func ResolveSigManifestNames(localPath string, subjectDigest digest.Digest) ([]string, error) {
	// loop through all files ending in ".sig" in local path, and match subjectDigest within the "subject" field in the json contents of each file. If the digest is found, save the filename in a slice of strings
	files, err := filepath.Glob(filepath.Join(localPath, "*"+SignatureFileExt)) // this will return all signature manifests, regardless of digest algorithm
	if err != nil {
		return nil, fmt.Errorf("searching for signature manifest files: %w", err)
	}
	manifestNames := make([]string, 0)
	for _, file := range files {
		manifestBytes, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading signature manifest file: %w", err)
		}
		var manifest ocispec.Manifest
		err = json.Unmarshal(manifestBytes, &manifest)
		if err != nil {
			return nil, fmt.Errorf("decoding signature manifest: %w", err)
		}
		if manifest.Subject.Digest == subjectDigest {
			manifestNames = append(manifestNames, filepath.Base(file))
		}
	}
	return manifestNames, nil
}
