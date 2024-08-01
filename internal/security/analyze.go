package security

import (
	"sync"
)

// ScanningResults represents the list of security results and contains a mutex for concurrent writes.
type ScanningResults struct {
	results []*ArtifactDetails
	mu      sync.Mutex
}

// ArtifactScanResults formats the artifact's pertinent grype JSON results for printing.
type ArtifactScanResults struct {
	CriticalVulnCount int `json:"critical_vulnerabilities"`
	HighVulnCount     int `json:"highVulnerabilites"`
	MediumVulnCount   int `json:"mediumVulnerabilities"`
}

func calculateResults(results *Results) (*ArtifactScanResults, error) {
	var securityResults ArtifactScanResults
	// count crits, high, medium and add to results
	for _, res := range results.Matches {
		switch res.Vulnerabilities.Severity {
		case "Critical":
			securityResults.CriticalVulnCount++
		case "High":
			securityResults.HighVulnCount++
		case "Medium":
			securityResults.MediumVulnCount++
		default:
			// filter out low/negligible/unknown
			continue
		}
	}

	return &securityResults, nil
}
