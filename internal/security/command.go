package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

func syftReference(ctx context.Context, reference string) ([]byte, error) {
	log := logger.FromContext(ctx)
	// exec out to syft to generate the SBOM
	log.InfoContext(ctx, "creating sbom", "reference", reference)
	cmd := exec.CommandContext(ctx, "syft", "scan", reference, "-o", "spdx-json")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing command: %s\n %w\n output: %s", cmd, err, string(res))
	}
	log.InfoContext(ctx, "created SBOM", "reference", reference)
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
		return nil, fmt.Errorf("error executing command: %s\n%w", cmd, err)
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
	if err := json.Unmarshal(res, &vulnerabilities); err != nil {
		return nil, fmt.Errorf("parsing vulnerabilities: %w", err)
	}

	return &vulnerabilities, nil
}
