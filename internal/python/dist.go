package python

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
)

// ErrInvalidPythonDistributionFilename is an error to signify that the python filename is not valid
// or our parser needs updating :).
var ErrInvalidPythonDistributionFilename = errors.New("invalid python distribution filename")

// Distribution represents a python distribution package.
type Distribution interface {
	// Project name
	Project() string

	// Version is the version string
	Version() string

	// Labels returns all the distribution information expanded as a set of labels if any of the labels.Set matches then
	// this distribution is compatible
	Labels() []labels.Set

	// OS() []string
	// Arch() []string
}

// NewDistribution parses a DistributionEntry and returns the a Distribution useful for filtering.
func NewDistribution(filename string) (Distribution, error) {
	// we support wheels and source distributions right now.
	// There should be no need to support .egg (not sufficient standards around it anyway)

	switch {
	case strings.HasSuffix(filename, ".whl"):
		return newWheel(filename)
	case strings.HasSuffix(filename, ".tar.gz"):
		return newSourceDistribution(filename)
	}

	return nil, nil
}

// sourceDistribution represents a python source distribution in the form of a archive (sdist).
type sourceDistribution struct {
	project string
	version string
}

// for all possible sdist extensions see https://www.geeksforgeeks.org/source-distribution-and-built-distribution-in-python/
var sdistExtensions = []string{".zip", ".tar.gz", ".tar.bz2", ".tar.xz", ".tar.Z", ".tar"}

func newSourceDistribution(filename string) (*sourceDistribution, error) {
	// source distributions look like
	// pyzmq-14.0.0.zip
	// pyzmq-14.0.0.tar.gz
	// aiohttp-cors-0.7.0.tar.gz

	// for all possible sdist extensions see https://www.geeksforgeeks.org/source-distribution-and-built-distribution-in-python/
	var name string
	for _, ext := range sdistExtensions {
		if strings.HasSuffix(filename, ext) {
			name = strings.TrimSuffix(filename, ext)
		}
	}
	if name == "" {
		return nil, fmt.Errorf("%s is not a source distribution (sdist): %w", filename, ErrInvalidPythonDistributionFilename)
	}

	lastInd := strings.LastIndex(name, "-")
	if lastInd == -1 {
		return nil, fmt.Errorf("source distribution (sdist) %s is missing the dash separator between name and version: %w", filename, ErrInvalidPythonDistributionFilename)
	}

	return &sourceDistribution{
		project: Normalize(name[:lastInd]),
		version: name[lastInd+1:],
	}, nil
}

// Project implements Distribution.
func (sdist *sourceDistribution) Project() string {
	return sdist.project
}

// Version implements Distribution.
func (sdist *sourceDistribution) Version() string {
	return sdist.version
}

// Labels implements Distribution.
func (sdist *sourceDistribution) Labels() []labels.Set {
	lbl := labelsForVersion(sdist.version)
	lbl["project"] = sdist.project
	lbl["type"] = "sdist"

	return []labels.Set{lbl}
}

// wheelDistribution represents a python binary distribution in the form of a wheel (bdist_wheel).
type wheelDistribution struct {
	project  string
	version  string
	build    *string // optional
	python   []string
	abi      []string
	platform []string
}

func newWheel(filename string) (*wheelDistribution, error) {
	// parse the filename
	// See https://peps.python.org/pep-0491/#file-name-convention
	// {distribution}-{version}(-{build tag})?-{python tag}-{abi tag}-{platform tag}.whl
	// pyzmq-23.2.1-pp39-pypy39_pp73-manylinux_2_17_x86_64.manylinux2014_x86_64.whl
	// pip-22.1.2-py3-none-any.whl

	if !strings.HasSuffix(filename, ".whl") {
		return nil, fmt.Errorf("%s is not a wheel: %w", filename, ErrInvalidPythonDistributionFilename)
	}

	// There are also tag sets https://peps.python.org/pep-0425/ so we split on periods
	name := strings.TrimSuffix(filename, ".whl")
	parts := strings.Split(name, "-")
	n := len(parts)
	if (n != 5) && (n != 6) {
		return nil, fmt.Errorf("wheel %s has %d parts but should have 5 or 6: %w", filename, n, ErrInvalidPythonDistributionFilename)
	}

	whl := &wheelDistribution{
		project:  Normalize(parts[0]),
		version:  parts[1],
		python:   strings.Split(parts[n-3], "."),
		abi:      strings.Split(parts[n-2], "."),
		platform: strings.Split(parts[n-1], "."),
	}
	if n == 6 {
		whl.build = &parts[2]
	}
	return whl, nil
}

// Project implements Distribution.
func (whl *wheelDistribution) Project() string {
	return whl.project
}

// Version implements Distribution.
func (whl *wheelDistribution) Version() string {
	return whl.version
}

// Labels implements Distribution.
func (whl *wheelDistribution) Labels() []labels.Set {
	// for a python implementation see https://peps.python.org/pep-0425/#compressed-tag-sets
	// updated here https://packaging.python.org/en/latest/specifications/platform-compatibility-tags/#compressed-tag-sets
	labelSets := make([]labels.Set, 0, len(whl.python)*len(whl.abi)*len(whl.platform))

	for _, plat := range whl.platform {
		// derived labels
		os, arch := decodePlatform(plat)

		for _, py := range whl.python {
			for _, abi := range whl.abi {
				lb := labelsForVersion(whl.version)
				lb["project"] = whl.project
				lb["type"] = "bdist_wheel"

				lb["python"] = py
				lb["abi"] = abi
				lb["platform"] = plat

				lb["os"] = os
				lb["arch"] = arch

				labelSets = append(labelSets, lb)
			}
		}
	}

	return labelSets
}

// Mapping from legacy to modern platform tags.
var legacyAliases = map[string]string{
	"manylinux1_x86_64":     "manylinux_2_5_x86_64",
	"manylinux1_i686":       "manylinux_2_5_i686",
	"manylinux2010_x86_64":  "manylinux_2_12_x86_64",
	"manylinux2010_i686":    "manylinux_2_12_i686",
	"manylinux2014_x86_64":  "manylinux_2_17_x86_64",
	"manylinux2014_i686":    "manylinux_2_17_i686",
	"manylinux2014_aarch64": "manylinux_2_17_aarch64",
	"manylinux2014_armv7l":  "manylinux_2_17_armv7l",
	"manylinux2014_ppc64":   "manylinux_2_17_ppc64",
	"manylinux2014_ppc64le": "manylinux_2_17_ppc64le",
	"manylinux2014_s390x":   "manylinux_2_17_s390x",
}

// decodePlatform extracts the OS and ARCH from the platform tag.
func decodePlatform(platform string) (string, string) {

	// For definitions of the platform tags see https://github.com/pypa/manylinux
	// Here is the latest spec https://peps.python.org/pep-0600/ and how to convert old ones to use it.

	// musllinux https://peps.python.org/pep-0656/

	if platform == "win32" {
		return "windows", ""
	}

	if newPlatform, exists := legacyAliases[platform]; exists {
		platform = newPlatform
	}

	os, arch, found := strings.Cut(platform, "_")
	if !found {
		return os, arch
	}

	switch os {
	case "manylinux", "musllinux", "macosx":
		parts := strings.SplitN(arch, "_", 3)
		arch = parts[len(parts)-1] // last element
	case "win":
		os = "windows"
	}

	return os, arch
}

/*
// goOSArch returns the GOOS and GOARCH from the python platform tag.
func goOSArch(platform string) (os string, arch string) {

	// For definitions of the platform tags see https://github.com/pypa/manylinux
	// Here is the latest spec https://peps.python.org/pep-0600/ and how to convert old ones to use it.

	// musllinux https://peps.python.org/pep-0656/

	if newPlatform, exists := legacyAliases[platform]; exists {
		platform = newPlatform
	}

	first, rest, found := strings.Cut(platform, "_")
	if !found {
		return
	}

	switch first {
	case "manylinux", "musllinux":
		os = "linux"
	case "macosx":
		os = "darwin"
	case "win32", "win64":
		os = "windows"
	default:
		os = first
	}

	if os == "linux" {
		parts := strings.SplitN(rest, "_", 3)
		arch = parts[len(parts)-1] // last element
	}

	switch arch {
	case "x86_64":
		arch = "amd64"
	case "aarch64":
		arch = "arm64"
	case "aarch32":
		arch = "arm"
	case "i686":
		arch = "386"
	}

	return os, arch
}
*/

var re = regexp.MustCompile("[-_.]+")

// Normalize computes the normalized name as per https://peps.python.org/pep-0503/#normalized-names
func Normalize(s string) string {
	return strings.ToLower(re.ReplaceAllString(s, "-"))
}
