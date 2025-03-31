package security

import (
	"fmt"
	"strings"
)

func calculateResults(results *Results) (*ArtifactScanReport, error) {
	var securityResults ArtifactScanReport
	for _, res := range results.Matches {
		// maybe move the filter logic in here?
		switch res.Vulnerabilities.Severity {
		case "Critical":
			securityResults.CriticalVulnerabilities = append(securityResults.CriticalVulnerabilities, res)
		case "High":
			securityResults.HighVulnerabilities = append(securityResults.HighVulnerabilities, res)
		case "Medium":
			securityResults.MediumVulnerabilities = append(securityResults.MediumVulnerabilities, res)
		case "Low":
			securityResults.LowVulnerabilities = append(securityResults.LowVulnerabilities, res)
		case "Unknown":
			securityResults.UnknownVulnerabilities = append(securityResults.UnknownVulnerabilities, res)
		case "Negligible":
			securityResults.NegligibleVulnerabilities = append(securityResults.NegligibleVulnerabilities, res)
		default:
			// filter out low/negligible/unknown
			continue
		}
	}

	return &securityResults, nil
}

// GetVulnerabilityCount parses the results json and returns a count for the given vulnerability severity level.
func (cr *ArtifactScanReport) GetVulnerabilityCount(severity string) int {
	switch severity {
	case "critical":
		return len(cr.CriticalVulnerabilities)
	case "high":
		return len(cr.HighVulnerabilities)
	case "medium":
		return len(cr.MediumVulnerabilities)
	case "low":
		return len(cr.LowVulnerabilities)
	case "negligible":
		return len(cr.NegligibleVulnerabilities)
	case "unknown":
		return len(cr.UnknownVulnerabilities)
	default:
		return 0
	}
}

// GetVulnerabilityCVEs parses the results json and returns a slice of CVE IDs for the given severity level.
func (cr *ArtifactScanReport) GetVulnerabilityCVEs(severity string) []string {
	var vulnerabilities []string
	switch severity {
	case "critical":
		for _, v := range cr.CriticalVulnerabilities {
			vulnerabilities = append(vulnerabilities, v.Vulnerabilities.ID)
		}
	case "high":
		for _, v := range cr.HighVulnerabilities {
			vulnerabilities = append(vulnerabilities, v.Vulnerabilities.ID)
		}
	case "medium":
		for _, v := range cr.MediumVulnerabilities {
			vulnerabilities = append(vulnerabilities, v.Vulnerabilities.ID)
		}
	case "low":
		for _, v := range cr.LowVulnerabilities {
			vulnerabilities = append(vulnerabilities, v.Vulnerabilities.ID)
		}
	case "negligible":
		for _, v := range cr.NegligibleVulnerabilities {
			vulnerabilities = append(vulnerabilities, v.Vulnerabilities.ID)
		}
	case "unknown":
		for _, v := range cr.UnknownVulnerabilities {
			vulnerabilities = append(vulnerabilities, v.Vulnerabilities.ID)
		}
	default:
		return nil
	}
	return vulnerabilities
}

// filterResults returns a map of the CVE matches for a given minimum vulnerabilityLevel and higher.
func filterResults(grypeResults *Results, vulnerabilityLevel string) (map[string]Matches, error) {
	matches := map[string]Matches{}
	minLevel, exists := SeverityLevels[strings.ToLower(vulnerabilityLevel)]
	if !exists {
		return nil, fmt.Errorf("invalid vulnerability level %s", vulnerabilityLevel)
	}
	// filter the matches for display purposes (there will be duplicates for multi-architecture images)
	for _, match := range grypeResults.Matches {
		if SeverityLevels[strings.ToLower(match.Vulnerabilities.Severity)] >= minLevel {
			matches[match.Vulnerabilities.ID] = match
		}

	}
	return matches, nil
}
