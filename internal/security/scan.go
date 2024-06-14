// Package security contains the logic for fetching artifact details, attaching SBOM's, etc. for ace-dt security scan.
package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"
	"sync"

	notationreg "github.com/notaryproject/notation-go/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/data/schema/pkg/selectors"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
)

const (
	// ArtifactTypeSPDX is the standard artifact type for SPDX formatted SBOMs.
	ArtifactTypeSPDX = "application/spdx+json"
)

// ArtifactDetails contains all of the details needed for a given artifact.
type ArtifactDetails struct {
	repository           *remote.Repository
	Source               mirror.Source
	Platforms            []string
	Size                 int64
	IsOCICompliant       bool
	Predecessors         []ocispec.Descriptor
	SignatureDigest      string
	SBOMDigest           string
	SBOM                 map[string][]byte
	OriginatingReference string // only needed for gather artifacts
}

type detailsStorage struct {
	details []ArtifactDetails
	mu      sync.Mutex
}

// ResolveScanReferences will fetch the artifact details for each image in a source file or a mirror (gather) artifact.
func ResolveScanReferences(ctx context.Context,
	sourceFile, artifactReference string,
	repoFunction func(context.Context, string) (*remote.Repository, error),
	concurrency int,
	dryRun bool,
) ([]ArtifactDetails, error) {
	storage := detailsStorage{
		details: []ArtifactDetails{},
		mu:      sync.Mutex{},
	}
	switch {
	case sourceFile != "":
		sources, err := mirror.ProcessSourcesFile(ctx, sourceFile, selectors.LabelSelectorSet{}, concurrency)
		if err != nil {
			return nil, err
		}
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(concurrency)
		for _, source := range sources {
			g.Go(func() error {
				// dryRun is manually overridden (set to true) for source-file input so we don't try to send SBOM's to directories in which we cannot push.
				sourceDetails, err := getManifestDetails(gctx, source.Name, repoFunction, true)
				if err != nil {
					return err
				}
				storage.mu.Lock()
				storage.details = append(storage.details, *sourceDetails)
				storage.mu.Unlock()
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}

	case artifactReference != "":
		sources, err := extractSourcesFromMirrorArtifact(ctx, artifactReference, repoFunction)
		if err != nil {
			return nil, err
		}
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(concurrency)
		for originalReference, source := range sources {
			g.Go(func() error {
				sourceDetails, err := getManifestDetails(gctx, source.Name, repoFunction, dryRun)
				if err != nil {
					return err
				}
				sourceDetails.OriginatingReference = originalReference
				storage.mu.Lock()
				storage.details = append(storage.details, *sourceDetails)
				storage.mu.Unlock()
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("must pass either a source file or a remote reference to a Gather artifact")
	}

	return storage.details, nil
}

func extractSourcesFromMirrorArtifact(ctx context.Context, reference string, repoFunction func(context.Context, string) (*remote.Repository, error)) (map[string]mirror.Source, error) {
	sources := make(map[string]mirror.Source)
	repo, err := repoFunction(ctx, reference)
	if err != nil {
		return nil, err
	}
	// fetch the reference index
	_, data, err := oras.FetchBytes(ctx, repo, repo.Reference.ReferenceOrDefault(), oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, fmt.Errorf("error fetching the artifact index: %w", err)
	}
	var idx ocispec.Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("unmarshalling artifact index: %w", err)
	}

	for _, manifest := range idx.Manifests {
		// create a source
		sources[manifest.Annotations[ref.AnnotationSrcRef]] = mirror.Source{
			Name:   strings.Join([]string{reference, manifest.Digest.String()}, "@"),
			Labels: nil,
		}
	}
	return sources, nil
}

func getManifestDetails(ctx context.Context, reference string, repoFunction func(context.Context, string) (*remote.Repository, error), dryRun bool) (*ArtifactDetails, error) {
	maniDetails := &ArtifactDetails{}
	maniDetails.Source.Name = reference
	maniDetails.SBOM = make(map[string][]byte)
	var platforms []string
	var size int64
	var predecessors []ocispec.Descriptor
	r, err := registry.ParseReference(reference)
	if err != nil {
		return maniDetails, fmt.Errorf("error parsing ref %s: %w", reference, err)
	}
	repo, err := repoFunction(ctx, reference)
	if err != nil {
		return maniDetails, err
	}
	maniDetails.repository = repo
	d, data, err := oras.FetchBytes(ctx, repo, r.ReferenceOrDefault(), oras.DefaultFetchBytesOptions)
	if err != nil {
		return maniDetails, fmt.Errorf("fetching the manifest bytes: %w", err)
	}
	maniDetails.IsOCICompliant = encoding.IsOCICompliant(d.MediaType)
	p, err := repo.Predecessors(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("fetching predecessors: %w", err)
	}

	// Fetch predecessors/SBOMs/Signatures
	predecessors = append(predecessors, p...)

	if encoding.IsIndex(d.MediaType) {
		var idx ocispec.Index
		err = json.Unmarshal(data, &idx)
		if err != nil {
			return nil, err
		}
		size += d.Size
		for _, man := range idx.Manifests {
			var img ocispec.Manifest
			manDesc, manData, err := oras.FetchBytes(ctx, repo, man.Digest.String(), oras.DefaultFetchBytesOptions)
			if err != nil {
				return nil, fmt.Errorf("fetching the manifest bytes: %w", err)
			}
			if err := json.Unmarshal(manData, &img); err != nil {
				return nil, err
			}
			size += manDesc.Size
			for _, blob := range img.Layers {
				size += blob.Size
			}
			p, err := repo.Predecessors(ctx, man)
			if err != nil {
				return maniDetails, fmt.Errorf("fetching predecessors: %w", err)
			}
			// Fetch predecessors/SBOMs/Signatures
			predecessors = append(predecessors, p...)
			platformString := formatPlatformString(man.Platform)
			platforms = append(platforms, platformString)
		}
	} else {
		size += d.Size
		if d.Platform != nil {
			platformString := formatPlatformString(d.Platform)
			platforms = append(platforms, platformString)
		} else {
			// pull the config and see if we can decode from there
			var img ocispec.Manifest
			if err := json.Unmarshal(data, &img); err != nil {
				return nil, err
			}
			_, rc, err := repo.Blobs().FetchReference(ctx, img.Config.Digest.String())
			if err != nil {
				return nil, fmt.Errorf("fetching reference %s: %w", img.Config.Digest.String(), err)
			}
			var platform ocispec.Platform
			if err = json.NewDecoder(rc).Decode(&platform); err != nil && errors.Is(io.EOF, err) {
				return nil, fmt.Errorf("decoding platform %+v: %w", &platform, err)
			}
			platformString := formatPlatformString(&platform)
			platforms = append(platforms, platformString)
		}
	}
	maniDetails.Platforms = platforms
	maniDetails.Size = size
	maniDetails.Predecessors = predecessors

	if err := maniDetails.handlePredecessors(ctx, dryRun); err != nil {
		return nil, err
	}
	return maniDetails, nil
}

func formatPlatformString(platform *ocispec.Platform) string {
	if platform != nil {
		return path.Join(platform.OS, platform.Architecture, platform.Variant)
	}
	return ""
}

func (ad *ArtifactDetails) handlePredecessors(ctx context.Context, dryRun bool) error {
	var hasSBOM bool
	for _, p := range ad.Predecessors {
		switch p.ArtifactType {
		case notationreg.ArtifactTypeNotation:
			ad.SignatureDigest = p.Digest.String()
		case ArtifactTypeSPDX:
			// try and extract sbom
			_, manifestBytesSBOM, err := oras.FetchBytes(ctx, ad.repository, p.Digest.String(), oras.DefaultFetchBytesOptions)
			if err != nil {
				return fmt.Errorf("fetching SBOM manifest bytes for %s: %w", p.Digest.String(), err)
			}
			var man ocispec.Manifest
			if err := json.Unmarshal(manifestBytesSBOM, &man); err != nil {
				continue
			}
			for _, l := range man.Layers {
				b, err := content.FetchAll(ctx, ad.repository, l)
				if err != nil {
					return fmt.Errorf("fetching layer for %s: %w", ad.Source.Name, err)
				}
				ad.SBOM[l.Digest.String()] = b
				hasSBOM = true
			}
		default:
			continue
		}
	}
	// if there is no SBOM present and it is not a dry run, generate and upload SBOM for the artifact
	if !hasSBOM && !dryRun {
		digestSBOM, b, err := GenerateSBOM(ctx, ad.Source.Name, ad.repository)
		if err != nil {
			return fmt.Errorf("error uploading SBOM for reference %s: %w", ad.Source.Name, err)
		}
		ad.SBOM[digestSBOM] = b
		ad.SBOMDigest = digestSBOM
	}
	return nil
}

// GenerateSBOM will generate and attach an SBOM for a given artifact.
// Returns the digest of the SBOM and the bytes or an error.
func GenerateSBOM(ctx context.Context, reference string, repository oras.GraphTarget) (string, []byte, error) {

	// fetch the main descriptor
	desc, err := repository.Resolve(ctx, reference)
	if err != nil {
		return "", nil, fmt.Errorf("error resolving descriptor for %s: %w", reference, err)
	}

	// exec out to syft to generate the SBOM
	log.InfoContext(ctx, "creating sbom", "reference", reference)
	cmd := exec.CommandContext(ctx, "syft", "scan", reference, "-o", "spdx-json")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", nil, fmt.Errorf("error executing command: %s\n %w\n output: %s", cmd, err, string(res))
	}
	log.InfoContext(ctx, "SBOM created", "reference", reference, "SBOM Bytes", len(res))

	// Upload the SBOM... create a manifest and encode SBOM into a layer.
	// The SBOM manifest Subject field must point to the digest of the main reference passed.
	descSBOM, err := oras.PushBytes(ctx, repository, ocispec.MediaTypeImageLayer, res)
	if err != nil {
		return "", nil, fmt.Errorf("pushing SBOM: %w", err)
	}

	// create an empty config and marshal it
	config := ocispec.ImageConfig{}
	cfg, err := json.Marshal(config)
	if err != nil {
		return "", nil, fmt.Errorf("marshalling empty config: %w", err)
	}

	// push the empty config
	descCfg, err := oras.PushBytes(ctx, repository, ocispec.MediaTypeEmptyJSON, cfg)
	if err != nil {
		return "", nil, fmt.Errorf("pushing empty config: %w", err)
	}

	packOpts := oras.PackManifestOptions{
		Subject:          &desc,
		Layers:           []ocispec.Descriptor{descSBOM},
		ConfigDescriptor: &descCfg,
	}

	log.InfoContext(ctx, "pushing SBOM", "reference", reference)
	_, err = oras.PackManifest(ctx, repository, oras.PackManifestVersion1_1, ArtifactTypeSPDX, packOpts)
	if err != nil {
		return "", nil, fmt.Errorf("pushing SBOM manifest: %w", err)
	}
	log.InfoContext(ctx, "successfully pushed SBOM", "reference", reference)
	return descSBOM.Digest.String(), res, nil
}
