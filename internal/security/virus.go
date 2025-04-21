package security

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	cfgdef "github.com/act3-ai/bottle-schema/pkg/apis/data.act3-ace.io/v1"
	"github.com/act3-ai/bottle-schema/pkg/mediatype"
	"github.com/act3-ai/data-tool/internal/actions/pypi"
	gitoci "github.com/act3-ai/data-tool/internal/git/oci"
	"github.com/act3-ai/data-tool/internal/mirror/encoding"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
)

// VirusScan scans an artifact using clamav and reutrns a slice of virus scanning report manifests.
func VirusScan(ctx context.Context,
	desc ocispec.Descriptor,
	target oras.GraphTarget,
	clamavChecksums []ClamavDatabase,
	pushResults bool,
	cachePath string) ([]*VirusScanManifestReport, error) {

	// do reports exist with matching clamav checksums?
	// output any errors but do not fail until the end.
	var reports []*VirusScanManifestReport
	if encoding.IsIndex(desc.MediaType) {
		var idx ocispec.Index
		b, err := content.FetchAll(ctx, target, desc)
		if err != nil {
			return nil, fmt.Errorf("fetching index manifest: %w", err)
		}
		if err := json.Unmarshal(b, &idx); err != nil {
			return nil, fmt.Errorf("parsing the image manifest: %w", err)
		}
		for _, man := range idx.Manifests {
			rl, err := VirusScan(ctx, man, target, clamavChecksums, pushResults, cachePath)
			if err != nil {
				return nil, err
			}
			reports = append(reports, rl...)
		}
	} else {
		report := VirusScanManifestReport{}
		rl, err := scanManifestForViruses(ctx, desc, target, cachePath)
		if err != nil {
			return nil, err
		}
		report.Results = rl
		report.ManifestDigest = desc.Digest.String()
		reports = append(reports, &report)
		if pushResults {
			if err := pushEmptyDesc(ctx, target, digest.SHA256); err != nil {
				return nil, err
			}

			data, err := json.Marshal(clamavChecksums)
			if err != nil {
				return nil, fmt.Errorf("marshalling the virus database checksums: %w", err)
			}
			// attach the results report
			rd, err := attachResultsReport(ctx, desc, ocispec.DescriptorEmptyJSON, reports, target, map[string]string{AnnotationVirusDatabaseChecksum: string(data)}, ArtifactTypeVirusScanReport)
			if err != nil {
				return nil, err
			}
			for _, result := range rd {
				slog.InfoContext(ctx, "pushed results", "reference", result.Digest.String())
			}
		}
	}
	return reports, nil
}

func scanManifestForViruses(ctx context.Context,
	desc ocispec.Descriptor,
	repository oras.GraphTarget,
	cachePath string) ([]*VirusScanResults, error) {
	var clamavResults []*VirusScanResults

	// create a blob tracker
	tracker := map[digest.Digest]string{}

	// pull the image manifest
	rc, err := repository.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}
	var manifest ocispec.Manifest
	decoder := json.NewDecoder(rc)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("parsing the image manifest: %w", err)
	}

	if manifest.ArtifactType == pypi.MediaTypePythonDistribution {
		return clamavPypiArtifact(ctx, cachePath, manifest.Layers, repository, tracker)
	}

	// pull the config
	rcCfg, err := repository.Fetch(ctx, manifest.Config)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest config: %w", err)
	}

	switch manifest.Config.MediaType {
	case gitoci.MediaTypeSyncConfig:
		slog.InfoContext(ctx, "Git Artifact Scanning is not supported at this time")
		return nil, nil
		// return clamavGitArtifact(ctx, cachePath, desc, repository, desc.Digest.String())
	case mediatype.MediaTypeBottleConfig:
		return clamavBottle(ctx, rcCfg, manifest.Layers, repository)
	default:
		for _, layer := range manifest.Layers {
			lrc, err := repository.Fetch(ctx, layer)
			if err != nil {
				return nil, fmt.Errorf("fetching the image layer: %w", err)
			}
			r, err := clamavBytes(ctx, lrc, "")
			if err != nil {
				return nil, err
			}

			if r != nil {
				r.LayerDigest = layer.Digest.String()
				clamavResults = append(clamavResults, r)
			}
		}
	}

	return clamavResults, nil
}

func clamavBytes(ctx context.Context, data io.ReadCloser, filename string) (*VirusScanResults, error) {
	cmd := exec.CommandContext(ctx, "clamscan", "--no-summary", "--infected", "--stdout", "-")
	cmd.Stdin = data
	res, err := cmd.CombinedOutput()
	foundPattern := regexp.MustCompile(`^\s*(.+?)\s+FOUND\b`)
	output := string(res)
	if match := foundPattern.FindStringSubmatch(output); len(match) > 1 {
		lines := strings.Split(strings.TrimSpace(match[1]), ":")
		return &VirusScanResults{
			File:    filename,
			Finding: lines[1],
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning data: %s %w", res, err)
	}

	if len(res) != 0 {
		lines := strings.Split(strings.TrimSpace(string(res)), ":")
		return &VirusScanResults{
			File:    filename,
			Finding: lines[1],
		}, nil
	}
	return nil, nil
}

func clamavBottle(ctx context.Context, cfg io.ReadCloser, layers []ocispec.Descriptor, repository oras.GraphTarget) ([]*VirusScanResults, error) {
	// Create the return slice of results for each layer
	var scanResults []*VirusScanResults

	// inspect the config for the hash -->  file mappings
	filenames := make(map[digest.Digest]string, len(layers))

	b, err := io.ReadAll(cfg)
	if err != nil {
		return nil, fmt.Errorf("reading the bottle config: %w", err)
	}

	r := bytes.NewReader(b)
	rc := io.NopCloser(r)

	// scan the config first
	cfgResults, err := clamavBytes(ctx, rc, "config")
	if err != nil {
		return nil, err
	}
	if cfgResults != nil {
		scanResults = append(scanResults, cfgResults)
	}

	var bottle cfgdef.Bottle
	if err := json.Unmarshal(b, &bottle); err != nil {
		return nil, fmt.Errorf("decoding the bottle config: %w", err)
	}
	for _, part := range bottle.Parts {
		filenames[part.Digest] = part.Name
	}
	for _, layer := range layers {
		v, ok := filenames[layer.Digest]
		layerBytes, err := content.FetchAll(ctx, repository, layer)
		if err != nil {
			return nil, fmt.Errorf("fetching the bottle layer %s: %w", layer.Digest.String(), err)
		}
		r := bytes.NewReader(layerBytes)
		rc := io.NopCloser(r)
		if ok {
			results, err := clamavBytes(ctx, rc, v)
			if err != nil {
				return nil, err
			}
			if results != nil {
				scanResults = append(scanResults, results)
			}
		} else {
			results, err := clamavBytes(ctx, rc, layer.Digest.String())
			if err != nil {
				return nil, err
			}
			if results != nil {
				scanResults = append(scanResults, results)
			}
		}
	}
	return scanResults, nil
}

func clamavPypiArtifact(ctx context.Context, cachePath string, layers []ocispec.Descriptor, repository oras.GraphTarget, tracker map[digest.Digest]string) ([]*VirusScanResults, error) {
	// iterate through the layers, if the layer is a wheel, unzip into tmp dir.
	// If the layer is metadata, can scan as is.
	// add scanned layers to tracker map
	var results []*VirusScanResults

	tmpDir, err := generateTempDir(cachePath)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	for _, layer := range layers {
		_, ok := tracker[layer.Digest]
		if ok {
			continue
		}
		// pull the layer
		lrc, err := repository.Fetch(ctx, layer)
		if err != nil {
			return nil, fmt.Errorf("fetching the layer: %w", err)
		}
		// initially scan the layer as a zip, if there is a positive virus scan result, the copied buffered reader is used to unzip and inspect the files in each layer.
		var buf bytes.Buffer
		tr := io.TeeReader(lrc, &buf)
		trc := io.NopCloser(tr)
		initialResults, err := clamavBytes(ctx, trc, layer.Digest.String())
		if err != nil {
			return nil, fmt.Errorf("scanning the pypi artifact layer %s: %w", layer.Digest.String(), err)
		}
		if initialResults == nil {
			// we don't want to pull out all of the files if there is no hit on the initial scan of the zip.
			continue
		}
		if layer.MediaType == pypi.MediaTypePythonDistributionWheel {
			buffReadCloser := io.NopCloser(&buf)
			res, err := unzipAndScanLayerFiles(ctx, cachePath, layer, buffReadCloser)
			if err != nil {
				return nil, err
			}
			if res != nil {
				results = append(results, res)
			}
			tracker[layer.Digest] = ""
		} else {
			// scan as is
			res, err := clamavBytes(ctx, lrc, layer.Digest.String())
			if err != nil {
				return nil, err
			}
			if res != nil {
				results = append(results, res)
			}
		}
	}
	return results, nil
}

func unzipAndScanLayerFiles(ctx context.Context, cachePath string, layer ocispec.Descriptor, layerReadCloser io.ReadCloser) (*VirusScanResults, error) {
	fp := filepath.Join(cachePath, "blobs", string(layer.Digest.Algorithm()), layer.Digest.Encoded())
	// unzip
	r, err := zip.OpenReader(fp)
	if err != nil {
		return nil, fmt.Errorf("initializing zip reader for pypi blob: %w", err)
	}
	defer r.Close()
	for _, f := range r.File {
		slog.InfoContext(ctx, "scanning", "filename", f.Name)
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening file within zip %s: %w", f.Name, err)
		}
		res, err := clamavBytes(ctx, rc, f.Name)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, nil
}

func getClamAVChecksum(ctx context.Context) ([]ClamavDatabase, error) {
	clamavDBChecksums := []ClamavDatabase{}
	// initialize the regex pattern for checksum parsing
	pattern := `Digital signature:\s*([a-zA-Z0-9+\/=]+)`
	re := regexp.MustCompile(pattern)
	// find all clamav database files in the default location
	cmd := exec.CommandContext(ctx, "sh", "-c", "ls /var/lib/clamav/*.cvd")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("looking for clamav databases in /var/lib/clamav/: %s %w", string(res), err)
	}
	dbFiles := strings.Fields(string(res))
	// iterate through the db files and get the checksum using sigtool, a clamav-installed tool
	for _, file := range dbFiles {
		// get the checksum on the db
		cmd := exec.CommandContext(ctx, "sigtool", "--info", file)
		res, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("getting checksum for clamav db file %s: %w", file, err)
		}
		checksum := re.FindSubmatch(res)
		if len(checksum) != 0 {
			// clamavDBChecksums[file] = string(checksum[1])
			clamavDBChecksums = append(clamavDBChecksums, ClamavDatabase{
				File:     file,
				Checksum: string(checksum[1]),
			})
		}
	}
	return clamavDBChecksums, nil
}
