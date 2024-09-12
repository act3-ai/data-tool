package sign

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	notationreg "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	content "oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/oci"
)

// Pull copies all signatures referring to the subject descriptor from source to btlPath.
// Safe to call if no signatures exist.
func Pull(ctx context.Context, btlPath string, source content.ReadOnlyGraphStorage, subject ocispec.Descriptor) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Discovering Bottle Signatures")
	sigManDescs, err := registry.Referrers(ctx, source, subject, notationreg.ArtifactTypeNotation)
	if err != nil {
		return fmt.Errorf("resolving bottle signature referrers: %w", err)
	}

	if len(sigManDescs) == 0 {
		log.InfoContext(ctx, "No signatures Found")
		return nil
	}
	log.InfoContext(ctx, "Fetching signatures", "signaturesFound", len(sigManDescs))

	sigsHandler := NotarySignatures{
		Subject:   subject,
		HashFunc:  crypto.SHA256, // TODO: Cryptographic agility
		LocalPath: bottle.SigDir(btlPath),
	}

	for _, desc := range sigManDescs {
		log.InfoContext(ctx, "Fetching signature", "sigManifestDigest", desc.Digest)
		handler, err := fetchNotarySig(ctx, source, desc)
		if err != nil {
			return fmt.Errorf("fetching notary signature: %w", err)
		}
		sigsHandler.SigManifests = append(sigsHandler.SigManifests, handler)
	}

	log.InfoContext(ctx, "Writing signatures to disk")
	if err := sigsHandler.WriteDisk(subject.Digest); err != nil {
		return fmt.Errorf("writing signatures: %w", err)
	}

	return nil
}

// fetchNotarySig fetches all contents of a notary signature manifest.
func fetchNotarySig(ctx context.Context, source content.ReadOnlyGraphStorage, manifestDesc ocispec.Descriptor) (SigsManifestHandler, error) {
	// fetch the manifest itself
	rawManifest, err := content.FetchAll(ctx, source, manifestDesc)
	if err != nil {
		return nil, fmt.Errorf("fetching signature manifest: %w", err)
	}

	var sigManifest ocispec.Manifest
	if err := json.Unmarshal(rawManifest, &sigManifest); err != nil {
		return nil, fmt.Errorf("parsing signature manifest: %w", err)
	}

	// fetch layers, there should only be one
	layerMap := make(map[digest.Digest][]byte, len(sigManifest.Layers))
	for _, layerDesc := range sigManifest.Layers {
		rawLayer, err := content.FetchAll(ctx, source, layerDesc) // signatures a typically small
		if err != nil {
			return nil, fmt.Errorf("fetching signature layer: %w", err)
		}
		layerMap[layerDesc.Digest] = rawLayer
	}

	return &SigsManifest{
		Descriptor: manifestDesc,
		Manifest:   sigManifest,
		Config:     ocispec.DescriptorEmptyJSON.Data, // notation configs are empty
		rawLayers:  layerMap,
		ManifestStatusInfo: oci.ManifestStatusInfo{
			Status:     oci.ManifestOK,
			StatusInfo: "",
			Error:      nil,
		},
	}, nil
}

// PrepareSigsGraph prepares signatures referring to subject for pushing by adding them
// to the file-backed oras filestore. Safe to call if no signatures exist.
// The additional descriptor is for the referrers fallback mechanism, which only supports
// a single tag. Due to the data-race nature of the fallback mechanism it is likely undesired
// behavior will occur if more than one fallback manifest is found in the .signature directory.
func PrepareSigsGraph(ctx context.Context, btlPath string, storage content.GraphStorage, subject ocispec.Descriptor) error {
	log := logger.V(logger.FromContext(ctx), 1)

	sigsDir := bottle.SigDir(btlPath)
	manifestPaths, err := ResolveSigManifestNames(sigsDir, subject.Digest)
	if err != nil {
		return fmt.Errorf("loading manifests at path %s: %w", btlPath, err)
	}

	if len(manifestPaths) > 0 {
		log.InfoContext(ctx, "prepparing bottle signatures for transfer", "signatureManifests", len(manifestPaths))
	} else {
		log.InfoContext(ctx, "no bottle signatures found")
		return nil
	}

	var addedNotationCfg bool // we only need to add a notation config once

	for _, manifestPath := range manifestPaths {
		manifestRaw, err := os.ReadFile(filepath.Join(sigsDir, manifestPath))
		if err != nil {
			return fmt.Errorf("reading manifest file: %w", err)
		}

		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestRaw, &manifest); err != nil {
			return fmt.Errorf("decoding raw manifest: %w", err)
		}

		// conform to oci distribution spec, see https://github.com/opencontainers/distribution-spec/blob/main/spec.md#listing-referrers
		sigType := manifest.ArtifactType
		if manifest.ArtifactType == "" {
			sigType = manifest.Config.MediaType
		}

		switch sigType {
		case notationreg.ArtifactTypeNotation: // notation sig manifest
			if !addedNotationCfg {
				err := storage.Push(ctx, ocispec.DescriptorEmptyJSON, bytes.NewReader([]byte("{}")))
				if err != nil {
					return fmt.Errorf("adding notation config to file store: %w", err)
				}
				addedNotationCfg = true
			}
		default:
			return fmt.Errorf("unsupported signature manifest type '%s'", sigType)
		}

		// add layers
		for _, sigLayer := range manifest.Layers {
			layerFilePath := filepath.Join(sigsDir, sigLayer.Digest.Hex())
			layerFile, err := os.Open(layerFilePath)
			if err != nil {
				return fmt.Errorf("opening signature file: %w", err)
			}
			defer layerFile.Close()

			if err := storage.Push(ctx, sigLayer, layerFile); err != nil {
				return fmt.Errorf("pushing signature to storage: %w", err)
			}

			if err := layerFile.Close(); err != nil {
				return fmt.Errorf("closing signature file: %w", err)
			}
			log.InfoContext(ctx, "pushed signature layer to storage", "descriptor", sigLayer)
		}

		// add sig manifest itself
		manifestDesc := ocispec.Descriptor{
			Digest:    digest.FromBytes(manifestRaw),
			MediaType: sigType,
			Size:      int64(len(manifestRaw)),
		}
		if err := storage.Push(ctx, manifestDesc, bytes.NewReader(manifestRaw)); err != nil {
			return fmt.Errorf("pushing signature manifest to storage: %w", err)
		}

		log.InfoContext(ctx, "pushed signature manifest to storage", "manifestSubject", manifest.Subject, "bottleManDesc", subject) // manifest.Subject should equal bottle descriptor
	}

	return nil
}
