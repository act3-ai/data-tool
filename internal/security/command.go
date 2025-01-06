package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/act3-ai/data-tool/internal/ui"
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

func clamavBytes(ctx context.Context, data io.ReadCloser) (*VirusScanResults, error) {
	cmd := exec.CommandContext(ctx, "clamscan", "--no-summary", "--infected", "--stdout", "-")
	cmd.Stdin = data
	res, err := cmd.CombinedOutput()
	foundPattern := regexp.MustCompile(`^\s*(.+?)\s+FOUND\b`)
	output := string(res)
	if match := foundPattern.FindStringSubmatch(output); len(match) > 1 {
		lines := strings.Split(strings.TrimSpace(match[1]), ":")
		return &VirusScanResults{
			File:    match[0],
			Finding: lines[1],
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning data: %s %w", res, err)
	}

	if len(res) != 0 {
		lines := strings.Split(strings.TrimSpace(string(res)), ":")
		return &VirusScanResults{
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
