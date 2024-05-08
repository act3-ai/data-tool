package cmd

import (
	"bufio"
	"bytes"
	"strings"
)

// TagRefPrefix is the git prefex for tag references.
const TagRefPrefix = "refs/tags/"

// HeadRefPrefix is the git prefex for head references.
const HeadRefPrefix = "refs/heads/"

// BundleSuffix is the git bundle file extension.
const BundleSuffix = ".bundle"

// parseGitOutput parses a git output by separating each line.
func parseGitOutput(out []byte) []string {
	s := bufio.NewScanner(bytes.NewReader(out))
	s.Split(bufio.ScanLines)

	parsed := make([]string, 0)
	for s.Scan() {
		parsed = append(parsed, s.Text())
	}

	return parsed
}

// parseOIDRefs parses the standard git output of (commit) OIDs and their
// references, returning index matching slices of the two. It expects
// a clean slice of OID + ref lines; some git commands have a "prefix" line.
//
// Expected format: <oid> [TAB | SP] <ref> LF .
func parseOIDRefs(lines ...string) ([]string, []string) {
	fullRefs := make([]string, len(lines))
	commits := make([]string, len(lines))
	for i, entry := range lines {
		split := strings.Fields(entry)
		commits[i], fullRefs[i] = split[0], split[1]
	}

	return commits, fullRefs
}
