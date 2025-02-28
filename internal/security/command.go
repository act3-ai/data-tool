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

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"

	cfgdef "github.com/act3-ai/bottle-schema/pkg/apis/data.act3-ace.io/v1"
	"github.com/act3-ai/data-tool/internal/actions/pypi"
	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func syftReference(ctx context.Context, reference string) ([]byte, error) {
	// log := logger.FromContext(ctx)
	// exec out to syft to generate the SBOM
	// log.InfoContext(ctx, "creating sbom", "reference", reference)
	cmd := exec.CommandContext(ctx, "syft", "scan", reference, "-o", "spdx-json")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing command: %s\n %w\n output: %s", cmd, err, string(res))
	}
	// log.InfoContext(ctx, "created SBOM", "reference", reference)
	return res, nil
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

func clamavFile(ctx context.Context, filename string) (*VirusScanResults, error) {
	cmd := exec.CommandContext(ctx, "clamscan", filename, "--no-summary", "--infected", "--stdout")
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

func grypeReference(ctx context.Context, reference string) (*VulnerabilityScanResults, error) {
	vulnerabilities := VulnerabilityScanResults{}
	cmd := exec.CommandContext(ctx, "grype", reference, "-o", "json")
	cmd.Env = append(os.Environ(),
		"GRYPE_DB_AUTO_UPDATE=false",
	)
	res, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(res), "oci-registry: unknown layer media type:") {
			return &vulnerabilities, nil
		}
		return nil, fmt.Errorf("error executing command: %s\n%w\n%s", cmd, err, res)
	}

	// catch grype warnings
	i := bytes.Index(res, []byte("{")) // where warnings end and json begins
	warnings := string(res[:i])
	if warnings != "" {
		ui.FromContextOrNoop(ctx).Info("Found grype warnings", warnings)
		res = res[i:]
	}

	if err := json.Unmarshal(res, &vulnerabilities); err != nil {
		return nil, fmt.Errorf("parsing vulnerabilities: %w", err)
	}

	return &vulnerabilities, nil
}

func grypeSBOM(ctx context.Context, sbom io.ReadCloser) (*VulnerabilityScanResults, error) {
	vulnerabilities := VulnerabilityScanResults{}
	cmd := exec.CommandContext(ctx, "grype", "-o", "json")
	cmd.Env = append(os.Environ(),
		"GRYPE_DB_AUTO_UPDATE=false",
	)
	cmd.Stdin = sbom
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing command: %s\n %w\n output: %s", cmd, err, string(res))
	}

	// catch grype warnings
	i := bytes.Index(res, []byte("{")) // where warnings end and json begins
	warnings := string(res[:i])
	if warnings != "" {
		ui.FromContextOrNoop(ctx).Info("Found grype warnings", warnings)
		res = res[i:]
	}

	if err := json.Unmarshal(res, &vulnerabilities); err != nil {
		return nil, fmt.Errorf("parsing vulnerabilities: %w", err)
	}

	return &vulnerabilities, nil
}

func getGrypeDBChecksum(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "grype", "db", "status", "--output", "json")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getting the grype db checksum: %w", err)
	}

	var grypeStatus grypeDBChecksum
	if err := json.Unmarshal(res, &grypeStatus); err != nil {
		return "", fmt.Errorf("decoding grype db checksum output: %w", err)
	}

	return grypeStatus.value()
}

// The following is for documenting changes to the output of grype db status
// due to frequent changes.
// v0.87.0 and prior:
//
//	struct {
//		SchemaVersion int
//		Built         string // time
//		Location      string // filepath
//		Checksum      string // sha256 hash
//		Error     	  struct
//	}
//
// v0.88.0 & v0.89.0:
//
//	struct {
//		SchemaVersion string // type is now string
//		Built         string
//		Path      	  string // Location renamed to Path
//		Checksum      string // change in format, likely a non-cryptographic digest
//		Error     	  string // error no longer a struct
//	}
//
// v0.90.0 (latest at this time):
//
//	struct {
//		SchemaVersion string // v6.0.2
//		Built         string
//		Path      	  string
//		From          string // Checksum renamed to From, another format change to a URL reference to a tar.zst file (with checksum)
//		Valid     	  bool   // Error renamed to Valid, type is now boolean
//	}

// grypeDBChecksum contains all possible keys for the "checksum" field as output
// by 'grype db status'.
type grypeDBChecksum struct {
	Checksum string `json:"checksum,omitempty"` // v0.89.0 and prior
	From     string `json:"from,omitempty"`     // v0.90.0 (latest at this time)
}

func (g *grypeDBChecksum) value() (string, error) {
	switch {
	case g.From != "":
		// grype v0.90.0 and later
		return g.From, nil
	case g.Checksum != "":
		// grype v0.89.0 and prior
		return g.Checksum, nil
	default:
		// theoretically impossible as 'grype db status' should throw an error first
		return "", fmt.Errorf("both checksum and from fields are empty, please run 'grype db status' to validate")
	}
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

func generateTempDir(cachePath string) (string, error) {
	// initialize the storage cache

	tmpDir, err := os.MkdirTemp(filepath.Join(cachePath, "tmp"), "")
	if err != nil {
		return "", fmt.Errorf("creating tmp git dir: %w", err)
	}

	return tmpDir, nil
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
