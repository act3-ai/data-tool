package security

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

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
func PrintMarkdown(out io.Writer, results []*ArtifactDetails) error {
	platformMap := analyzePlatforms(results)
	// sort artifacts by critical count
	newResults := sortArtifacts(results, 10)
	newResults = formatArtifactNames(newResults)
	xaxis := make([]string, len(newResults))
	criticalCount := make([]int, len(newResults))
	rows := make([]string, len(newResults))

	for i, res := range newResults {
		xaxis[i] = res.shortenedName
		criticalCount[i] = res.CalculatedResults.CriticalVulnCount
		platformStringList := strings.Join(res.Platforms, ", ")
		tablefmt := "|" + strings.Join([]string{res.OriginatingReference, strconv.Itoa(res.CalculatedResults.CriticalVulnCount), strconv.Itoa(res.CalculatedResults.HighVulnCount), strconv.Itoa(res.CalculatedResults.MediumVulnCount), platformStringList, strconv.FormatBool(res.IsOCICompliant), strconv.FormatBool(res.SBOMDigest != ""), strconv.FormatBool(res.SignatureDigest != "")}, "|") + "|"
		rows[i] = tablefmt
	}

	formattedCriticalCount := formatAxis(criticalCount)
	formattedXaxis := make([]string, len(xaxis))
	for i, v := range xaxis {
		formattedXaxis[i] = fmt.Sprintf(`"%s"`, v)
	}
	joined := strings.Join(formattedXaxis, ",")
	result := fmt.Sprintf(`[%s]`, joined)
	graphConfig := `%%{init: {"xyChart":{"width": 1200, "chartOrientation": "horizontal","xAxis":{"labelFontSize": 12}, "yAxis":{"showTitle": false}}}}%%`
	criticalGraph := fmt.Sprintf("# Vulnerabilities\n\n```mermaid\n %s\nxychart-beta\n title \"Critical Vulnerabilities\"\n x-axis %s\n y-axis \"Critical vulnerabiliites\"\n bar %s\n```\n", graphConfig, result, formattedCriticalCount)
	_, err := fmt.Fprintln(out, criticalGraph)
	if err != nil {
		return fmt.Errorf("printing the critical graph to markdown: %w", err)
	}

	// print markdown table
	table := `| Reference | Critical Vulnerabilities | High Vulnerabilities | Medium Vulnerabilities | Platforms | OCI Compliance | SBOM exists | Signed |
|-----|-----|-----|-----|-----|-----|-----|-----|`

	_, err = fmt.Fprintln(out, table)
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
| Platform | Count | Percent Coverage |
|------|------|------|`
	_, err = fmt.Fprintln(out, tablePlatform)
	if err != nil {
		return fmt.Errorf("printing the platform table to markdown: %w", err)
	}
	for k, v := range platformMap {

		_, err = fmt.Fprintln(out, "|", strings.Join([]string{k, strconv.Itoa(v), strconv.FormatFloat((float64(v)/float64(len(results)))*100, 'f', 2, 64) + "%"}, "|"), "|")
		if err != nil {
			return fmt.Errorf("printing the vulnerability table to markdown: %w", err)
		}
	}
	return nil
}

// PrintCSV prints out the ArtifactDetails in CSV format to the io.Writer defined.
func PrintCSV(out io.Writer, results []*ArtifactDetails) error {
	table := make([][]string, len(results)+1)
	header := []string{"reference", "critical", "high", "medium"}
	w := csv.NewWriter(out)
	table[0] = header
	for i, res := range results {
		table[i+1] = []string{res.OriginatingReference, strconv.Itoa(res.CalculatedResults.CriticalVulnCount), strconv.Itoa(res.CalculatedResults.HighVulnCount), strconv.Itoa(res.CalculatedResults.MediumVulnCount)}
	}
	if err := w.WriteAll(table); err != nil {
		return fmt.Errorf("writing csv table: %w", err)
	}
	return nil
}

// PrintTable prints out the ArtifactDetails in a printed table format to the io.Writer defined.
func PrintTable(out io.Writer, results []*ArtifactDetails) error {
	table := make([][]string, len(results)+1)
	tableHeader := []string{"reference", "critical", "high", "medium"}
	table[0] = tableHeader
	for i, res := range results {
		// platformStringList := strings.Join(res.Platforms, ",") // list does not look good in table
		table[i+1] = []string{res.OriginatingReference, strconv.Itoa(res.CalculatedResults.CriticalVulnCount), strconv.Itoa(res.CalculatedResults.HighVulnCount), strconv.Itoa(res.CalculatedResults.MediumVulnCount)}
	}
	w := tabwriter.NewWriter(out, 1, 1, 1, ' ', 0)

	for _, row := range table {
		_, err := fmt.Fprintln(w, strings.Join(row, "\t|"))
		if err != nil {
			return fmt.Errorf("printing the results table: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("printing table: %w", err)
	}
	return nil
}

func formatAxis(intSlice []int) string {
	// Create a slice to hold the string representations of the integers
	strSlice := make([]string, len(intSlice))

	// Convert each integer to a string
	for i, v := range intSlice {
		strSlice[i] = strconv.Itoa(v)
	}

	// Join the string representations with commas
	return "[" + strings.Join(strSlice, ", ") + "]"
}

func sortArtifacts(results []*ArtifactDetails, filterTopCount int) []*ArtifactDetails {
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].CalculatedResults.CriticalVulnCount > results[j].CalculatedResults.CriticalVulnCount
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
		image := path.Base(artifact.OriginatingReference)
		// we want to evaluate if the artifact name already exists in the mapper, if it does, we want to make them both unique
		value, ok := mapper[image]
		if ok {
			ref1, ref2 := dilineateArtifactNames(value.OriginatingReference, artifact.OriginatingReference)
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
		for _, platform := range res.Platforms {
			platformMap[platform]++
		}
	}
	return platformMap
}
