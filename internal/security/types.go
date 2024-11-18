package security

const (
	// ArtifactTypeSPDX is the standard artifact type for SPDX formatted SBOMs.
	ArtifactTypeSPDX = "application/spdx+json"
	// ArtifactTypeHarborSBOM is the artifact type given to SBOMs generated via the Harbor UI.
	ArtifactTypeHarborSBOM = "application/vnd.goharbor.harbor.sbom.v1"
	// ArtifactTypeVulnerabilityReport is the artifact type given to ASCE vulnerability results.
	ArtifactTypeVulnerabilityReport = "application/vnd.act3-ace.data.cve.results+json"
	// AnnotationGrypeDatabaseChecksum is the checksum of the grype database that is attached to the vulnerability results.
	AnnotationGrypeDatabaseChecksum = "vnd.act3-ace.scan.database.checksum"
	// MediaTypeHelmChartConfig defines the expected media type of a helm chart config manifest.
	MediaTypeHelmChartConfig = "application/vnd.cncf.helm.config.v1+json"
)

// VulnerabilityScanResults holds the vulnerability data for all given artifacts.
type VulnerabilityScanResults struct {
	Matches []Matches `json:"matches"`
}

// Matches represents the vulnerability matches and details for a given artifact.
type Matches struct {
	Vulnerabilities Vulnerability `json:"vulnerability"`
	Artifact        Artifact      `json:"artifact"`
}

// Vulnerability represents a specific vulnerability for a given artifact.
type Vulnerability struct {
	ID          string `json:"id"`
	Source      string `json:"dataSource"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	State       string `json:"state"`
}

// Artifact represents the identifying details for a given artifact.
type Artifact struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ArtifactScanReport formats the artifact's pertinent grype JSON results for printing.
type ArtifactScanReport struct {
	CriticalVulnerabilities   []Matches `json:"CriticalVulnerabilities"`
	HighVulnerabilities       []Matches `json:"HighVulnerabilities"`
	MediumVulnerabilities     []Matches `json:"MediumVulnerabilities"`
	LowVulnerabilities        []Matches `json:"LowVulnerabilities"`
	UnknownVulnerabilities    []Matches `json:"UnknownVulnerabilities"`
	NegligibleVulnerabilities []Matches `json:"NegligibleVulnerabilities"`
}

// SeverityLevels enumerates the string severity levels for filtering.
var SeverityLevels = map[string]int{
	"critical":   5,
	"high":       4,
	"medium":     3,
	"low":        2,
	"negligible": 1,
	"unknown":    0,
}

// Virus scanning

type VirusScanResults struct {
	File        string `json:"File"`
	MalwareName string `json:"MalwareName"`
}
