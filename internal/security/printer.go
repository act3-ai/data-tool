package security

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
	"golang.org/x/term"
)

// These are the ANSI color code definitions.
const (
	BoldRed = "\033[1;31m"
	Red     = "\033[31m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Green   = "\033[32m"
	Gray    = "\033[37m"
	Reset   = "\033[0m"
)

func prettifySeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return BoldRed + severity + Reset
	case "high":
		return Red + severity + Reset
	case "medium":
		return Yellow + severity + Reset
	case "low":
		return Blue + severity + Reset
	case "negligible", "unknown":
		return Gray + severity + Reset
	default:
		return severity
	}
}

// PrintJSON prints out the ArtifactDetails in JSON format to the io.Writer defined.
func PrintJSON(out io.Writer, results []*ArtifactDetails) error {
	for _, r := range results {
		marshalledData, err := json.MarshalIndent(*r, "", "    ")
		if err != nil {
			return fmt.Errorf("marshalling results: %w", err)
		}
		_, err = fmt.Fprintln(out, string(marshalledData))
		if err != nil {
			return fmt.Errorf("error printing JSON output: %w", err)
		}
	}
	return nil
}

// PrintMarkdown prints out the ArtifactDetails in markdown format to the io.Writer defined. Includes mermaid and platform coverage charts.
func PrintMarkdown(out io.Writer, results []*ArtifactDetails, vulnerabilityLevel string) error {
	platformMap := analyzePlatforms(results)
	// sort artifacts by critical count
	newResults := sortArtifacts(results, 10)
	newResults = formatArtifactNames(newResults)
	xaxis := make([]string, len(newResults))
	criticalCount := make([]int, len(newResults))
	rows := make([]string, len(newResults))

	for i, res := range newResults {
		xaxis[i] = res.shortenedName
		criticalCount[i] = len(res.CalculatedResults.CriticalVulnerabilities)
		platformStringList := strings.Join(res.platforms, ", ")
		tablefmt := "|" + strings.Join([]string{res.originatingReference, strconv.Itoa(len(res.CalculatedResults.CriticalVulnerabilities)), strconv.Itoa(len(res.CalculatedResults.HighVulnerabilities)), strconv.Itoa(len(res.CalculatedResults.MediumVulnerabilities)), platformStringList, strconv.FormatBool(res.isOCICompliant), strconv.FormatBool(res.manifestDigestSBOM != ""), strconv.FormatBool(res.signatureDigest != "")}, "|") + "|"
		rows[i] = tablefmt
	}

	// TODO think about if we want graphs
	// formattedCriticalCount := formatAxis(criticalCount)
	// formattedXaxis := make([]string, len(xaxis))
	// for i, v := range xaxis {
	// 	formattedXaxis[i] = fmt.Sprintf(`"%s"`, v)
	// }
	// joined := strings.Join(formattedXaxis, ",")
	// result := fmt.Sprintf(`[%s]`, joined)
	// graphConfig := `%%{init: {"xyChart":{"width": 1200, "chartOrientation": "horizontal","xAxis":{"labelFontSize": 12}, "yAxis":{"showTitle": false}}}}%%`
	// criticalGraph := fmt.Sprintf("# Vulnerabilities\n\n```mermaid\n %s\nxychart-beta\n title \"Critical Vulnerabilities\"\n x-axis %s\n y-axis \"Critical vulnerabiliites\"\n bar %s\n```\n", graphConfig, result, formattedCriticalCount)
	// _, err := fmt.Fprintln(out, criticalGraph)
	// if err != nil {
	// 	return fmt.Errorf("printing the critical graph to markdown: %w", err)
	// }

	// print markdown table
	table := `| Reference | Critical Vulnerabilities | High Vulnerabilities | Medium Vulnerabilities | Platforms | OCI Compliance | SBOM exists | Signed |
|-----|-----|-----|-----|-----|-----|-----|-----|`

	_, err := fmt.Fprintln(out, table)
	if err != nil {
		return fmt.Errorf("printing the vulnerability table to markdown: %w", err)
	}

	for _, row := range rows {
		_, err = fmt.Fprintln(out, row)
		if err != nil {
			return fmt.Errorf("printing the vulnerability row to markdown %s: %w", row, err)
		}
	}

	// print platform data
	tablePlatform := `
| Platform | Platform Support |
|------|------|`
	_, err = fmt.Fprintln(out, tablePlatform)
	if err != nil {
		return fmt.Errorf("printing the platform table to markdown: %w", err)
	}

	var numHelmFiles int
	for _, res := range results {
		if res.isNotScanSupported {
			numHelmFiles++
		}
	}

	var i int
	data := make([][]string, len(platformMap))
	for k, v := range platformMap {
		coverage := (float64(v) / float64(len(results)-numHelmFiles)) * 100
		if coverage > 100.00 {
			coverage = 100.00
		}
		coveragefmt := strconv.FormatFloat(coverage, 'f', 2, 64)
		// data = append(data, []string{k, coverage})
		data[i] = []string{k, coveragefmt}
		i++
	}
	sortDataByCoverage(data)
	for _, line := range data {
		if err != nil {
			return fmt.Errorf("getting platform count: %w", err)
		}
		if line[0] == "" {
			continue
		}
		_, err = fmt.Fprintln(out, "|", strings.Join([]string{line[0], line[1] + "%"}, "|"), "|")
		if err != nil {
			return fmt.Errorf("printing the vulnerability table to markdown: %w", err)
		}
	}

	return nil
}

// PrintCSV prints out the ArtifactDetails in CSV format to the io.Writer defined.
func PrintCSV(out io.Writer, results []*ArtifactDetails, vulnerabilityLevel string) error {
	orderedSeverities := []string{"critical", "high", "medium", "low", "negligible", "unknown"}
	table := make([][]string, len(results)+1)
	minLevel := SeverityLevels[strings.ToLower(vulnerabilityLevel)]
	header := []string{"reference"}
	for _, severity := range orderedSeverities {
		if SeverityLevels[severity] >= minLevel {
			header = append(header, severity)
		}
	}
	w := csv.NewWriter(out)
	table[0] = header
	// data := []string{res.OriginatingReference, strconv.Itoa(len(res.CalculatedResults.CriticalVulnerabilites)), strconv.Itoa(len(res.CalculatedResults.HighVulnerabilities)), strconv.Itoa(len(res.CalculatedResults.MediumVulnerabilites))}
	for i, res := range results {
		data := []string{res.originatingReference}
		for i := 1; i < len(header); i++ {
			data = append(data, strconv.Itoa(res.CalculatedResults.GetVulnerabilityCount(header[i])))
		}
		table[i+1] = data
	}
	if err := w.WriteAll(table); err != nil {
		return fmt.Errorf("writing csv table: %w", err)
	}
	return nil
}

// PrintTable prints out the ArtifactDetails in a printed table format to the io.Writer defined.
func PrintTable(out io.Writer, results []*ArtifactDetails, vulnerabilityLevel string, displayCVEs, displayPlatforms, displayMalware bool) error {
	table := [][]string{}
	orderedSeverities := []string{"critical", "high", "medium", "low", "negligible", "unknown"}
	minLevel := SeverityLevels[strings.ToLower(vulnerabilityLevel)]
	columns := 2
	tableHeader := []string{"reference", "size"}
	for _, severity := range orderedSeverities {
		if SeverityLevels[severity] >= minLevel {
			tableHeader = append(tableHeader, severity)
			columns++
		}
	}

	table = append(table, []string{}, tableHeader)
	for _, res := range results {
		data := []string{res.originatingReference, humanize.Bytes(uint64(res.size))}
		for i := 2; i < len(tableHeader); i++ {
			data = append(data, strconv.Itoa(res.CalculatedResults.GetVulnerabilityCount(tableHeader[i])))
		}

		table = append(table, data)
	}
	table = append(table, []string{})

	if err := PrintCustomTable(out, table); err != nil {
		return err
	}
	// use switch/fallthrough instead?
	if displayMalware {
		if err := printMalwareTable(out, results); err != nil {
			return err
		}
	}
	if displayCVEs {
		if err := printCVETable(out, results, vulnerabilityLevel); err != nil {
			return err
		}
	}
	if displayPlatforms {
		if err := printPlatformsTable(out, results); err != nil {
			return err
		}
	}
	return nil
}

// printMalwareTable
func printMalwareTable(out io.Writer, results []*ArtifactDetails) error {
	table := [][]string{{}}
	table[0] = []string{"reference", "layer", "malware ID"}
	for _, res := range results {
		if res.MalwareResults == nil {
			table = append(table, []string{res.originatingReference, "", "No Malware Found"})
		}
		for _, malwareResult := range res.MalwareResults {
			table = append(table, []string{res.originatingReference, malwareResult.File, malwareResult.MalwareName})
		}
	}
	if len(table) == 1 {
		_, err := out.Write([]byte("no malware vulnerabilities found\n\n"))
		if err != nil {
			return fmt.Errorf("writing to out: %w", err)
		}
		return nil
	}
	return PrintCustomTable(out, table)
}

// PrintTable prints out the ArtifactDetails in a printed table format to the io.Writer defined.
func printCVETable(out io.Writer, results []*ArtifactDetails, vulnerabilityLevel string) error {
	// are we printing to terminal? If so make it pretty
	var isTerminal bool
	if f, ok := out.(*os.File); ok {
		isTerminal = term.IsTerminal(int(f.Fd()))
	}

	table := [][]string{{}}
	orderedSeverities := []string{"critical", "high", "medium", "low", "negligible", "unknown"}
	minLevel := SeverityLevels[strings.ToLower(vulnerabilityLevel)]
	// create the table header
	tableHeader := []string{"reference", "CVE", "severity"}
	table[0] = tableHeader
	data := [][]string{}
	for _, res := range results {
		for _, sev := range orderedSeverities {
			if SeverityLevels[sev] >= minLevel {
				levelVulnerabilities := res.CalculatedResults.GetVulnerabilityCVEs(sev)
				for _, v := range levelVulnerabilities {
					if isTerminal {
						data = append(data, []string{res.originatingReference, v, prettifySeverity(sev)})
					} else {
						data = append(data, []string{res.originatingReference, v, sev})
					}
				}
			}
		}
		if data == nil {
			continue
		}
	}
	table = append(table, data...)
	if len(table) == 1 {
		_, err := out.Write([]byte("no CVE vulnerabilities found\n\n"))
		if err != nil {
			return fmt.Errorf("writing to out: %w", err)
		}
		return nil
	}
	return PrintCustomTable(out, table)
}

func printPlatformsTable(out io.Writer, results []*ArtifactDetails) error {
	platformMap := analyzePlatforms(results)
	if len(platformMap) == 0 {
		_, err := out.Write([]byte("no platforms found\n\n"))
		if err != nil {
			return fmt.Errorf("writing to out: %w", err)
		}
		return nil
	}
	table := [][]string{}
	table = append(table, []string{"platform", "platform support"})
	var numHelmFiles int
	for _, res := range results {
		if res.isNotScanSupported {
			numHelmFiles++
		}
	}
	data := make([][]string, len(platformMap))
	var i int
	for platform, count := range platformMap {
		support := (float64(count) / float64(len(results)-numHelmFiles)) * 100
		if support > 100.00 {
			support = 100.00
		}
		coverage := strconv.FormatFloat(support, 'f', 2, 64)
		// data = append(data, []string{platform, coverage})
		data[i] = []string{platform, coverage}
		i++
	}
	sortDataByCoverage(data)
	table = append(table, data...)
	return PrintCustomTable(out, table)
}

// PrintCustomTable accepts a slice of string slices and will format it into table format with separator strings and spacing.
func PrintCustomTable(out io.Writer, table [][]string) error {
	formattedTable := [][]string{}
	maxSeparatorLength := []int{}
	divider := []int{}
	var previousKey string

	for i, entry := range table {
		if len(entry) == 0 {
			continue
		}
		if entry[0] != previousKey {
			// this keeps track of which indexes to put the dividers in the table
			divider = append(divider, i)
		}
		for i, v := range entry {
			if len(maxSeparatorLength) < len(entry) {
				// then just append
				maxSeparatorLength = append(maxSeparatorLength, len(v))
			} else if len(v) > maxSeparatorLength[i] {
				maxSeparatorLength[i] = len(v)
			}
		}
		previousKey = entry[0]
	}
	s := separatorString(maxSeparatorLength)
	var previousIndex int
	for _, tableIndex := range divider {
		formattedTable = append(formattedTable, table[previousIndex:tableIndex]...)
		formattedTable = append(formattedTable, s)
		previousIndex = tableIndex
	}
	// print the rest of the table
	formattedTable = append(formattedTable, table[previousIndex:]...)
	formattedTable = append(formattedTable, []string{})
	w := tabwriter.NewWriter(out, 1, 1, 1, ' ', 0)
	for _, row := range formattedTable {
		_, err := fmt.Fprintln(w, strings.Join(row, "\t|"))
		if err != nil {
			return fmt.Errorf("printing the platform table: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing platform table: %w", err)
	}
	return nil
}

func sortDataByCoverage(data [][]string) {
	sort.Slice(data, func(i, j int) bool {
		coverageI, err := strconv.ParseFloat(data[i][1], 64)
		if err != nil {
			return data[i][1] < data[j][1]
		}
		coverageJ, err := strconv.ParseFloat(data[j][1], 64)
		if err != nil {
			return data[i][1] < data[j][1]
		}

		return coverageI > coverageJ
	})
}

// func formatAxis(intSlice []int) string {
// 	// Create a slice to hold the string representations of the integers
// 	strSlice := make([]string, len(intSlice))

// 	// Convert each integer to a string
// 	for i, v := range intSlice {
// 		strSlice[i] = strconv.Itoa(v)
// 	}

// 	// Join the string representations with commas
// 	return "[" + strings.Join(strSlice, ", ") + "]"
// }

func sortArtifacts(results []*ArtifactDetails, filterTopCount int) []*ArtifactDetails {
	sort.SliceStable(results, func(i, j int) bool {
		return len(results[i].CalculatedResults.CriticalVulnerabilities) > len(results[j].CalculatedResults.CriticalVulnerabilities)
	})
	if filterTopCount > len(results) {
		return results
	}
	return results[:filterTopCount]
}

func formatArtifactNames(artifacts []*ArtifactDetails) []*ArtifactDetails {
	// the mapper is just to check for duplicate names
	mapper := make(map[string]ArtifactDetails, len(artifacts))

	for _, artifact := range artifacts {
		image := path.Base(artifact.originatingReference)
		// we want to evaluate if the artifact name already exists in the mapper, if it does, we want to make them both unique
		value, ok := mapper[image]
		if ok {
			ref1, ref2 := dilineateArtifactNames(value.originatingReference, artifact.originatingReference)
			mapper[ref1] = mapper[image]
			art := mapper[ref1]
			// assign the shortened names to the images
			art.shortenedName = ref1
			artifact.shortenedName = ref2
			// we want to rename both to include more information
			delete(mapper, image)
			mapper[ref2] = *artifact
		} else {
			mapper[image] = *artifact
			// assign the shortened name to the new artifact
			artifact.shortenedName = image
		}

	}
	return artifacts
}

func dilineateArtifactNames(image1, image2 string) (string, string) {
	// are they exactly the same?
	if image1 == image2 {
		return image1, image2
	}

	dir1, repo1 := path.Split(image1)
	dir2, repo2 := path.Split(image2)
	if repo1 != repo2 {
		return repo1, repo2
	}
	dir1 = strings.TrimSuffix(dir1, "/")
	dir2 = strings.TrimSuffix(dir2, "/")
	newString1, newString2 := dilineateArtifactNames(dir1, dir2)
	return path.Join(newString1, repo1), path.Join(newString2, repo2)
}

func analyzePlatforms(results []*ArtifactDetails) map[string]int {
	platformMap := make(map[string]int)
	for _, res := range results {
		for _, platform := range res.platforms {
			if res.isNotScanSupported || platform == "" {
				continue
			}
			platformMap[platform]++
		}
	}
	return platformMap
}

func separatorString(maxSeparatorLength []int) []string {
	separator := []string{}
	for _, width := range maxSeparatorLength {
		separator = append(separator, strings.Repeat("-", width))
	}
	return separator
}
