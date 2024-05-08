package python

import (
	"strings"

	"k8s.io/apimachinery/pkg/labels"
)

// The version format for a distribution are defined in https://peps.python.org/pep-0440/
// need to write a parser string -> Version struct
// need to write a comparison function on Version
// requirement line, version specifier to a list of clauses
// matching function for clauses on a Version struct

func labelsForVersion(version string) labels.Set {
	// see https://peps.python.org/pep-0440/ for how versions are defined for python

	// we only want the "public version identifier"
	version, _, _ = strings.Cut(version, "+")

	// split on periods to get the parts
	parts := strings.Split(version, ".")

	lb := labels.Set{}
	if len(parts) >= 1 {
		lb["version.major"] = parts[0]
	}
	if len(parts) >= 2 {
		lb["version.minor"] = parts[1]
	}
	if len(parts) >= 3 {
		lb["version.patch"] = parts[2]
	}
	// TODO handle better the cases like 1.2.dev5 where there is no real "patch" number.  We want to stop short.

	return lb
}
