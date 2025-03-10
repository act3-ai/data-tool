package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"git.act3-ace.com/ace/data/tool/internal/ui"
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

func grypeReference(ctx context.Context, reference string) (*Results, error) {
	vulnerabilities := Results{}
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

func grypeSBOM(ctx context.Context, sbom io.ReadCloser) (*Results, error) {
	vulnerabilities := Results{}
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
	cmd := exec.CommandContext(ctx, "grype", "db", "status")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getting the grype db checksum: %w", err)
	}
	pattern := `Checksum:\s*([a-z0-9\:]+)`
	re := regexp.MustCompile(pattern)
	checksum := re.FindSubmatch(res)
	return string(checksum[1]), nil
}
