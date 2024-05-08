package oci

import (
	"fmt"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// platformValue implements cobra.Value for the "platform" flag
// borrowed from crane.

type platformValue struct {
	platform *ocispec.Platform
}

func (pv *platformValue) Set(platform string) error {
	p, err := ParsePlatform(platform)
	if err != nil {
		return err
	}
	pv.platform = p
	return nil
}

func (pv *platformValue) String() string {
	return PlatformToString(pv.platform)
}

func (pv *platformValue) Type() string {
	return "platform"
}

// PlatformToString accepts a pointer to an ocispec.Platform and turns it into an "os/architecture/variant"-formatted string.
func PlatformToString(p *ocispec.Platform) string {
	if p == nil {
		return "all"
	}
	platform := ""
	if p.OS != "" && p.Architecture != "" {
		platform = p.OS + "/" + p.Architecture
	}
	if p.Variant != "" {
		platform += "/" + p.Variant
	}
	return platform
}

// ParsePlatform accepts an "os/architecture/variant"-formatted string and returns a pointer to an ocispec.Platform.
func ParsePlatform(platform string) (*ocispec.Platform, error) {
	if platform == "all" {
		return nil, nil
	}

	p := &ocispec.Platform{}

	parts := strings.SplitN(platform, ":", 2)
	if len(parts) == 2 {
		p.OSVersion = parts[1]
	}

	parts = strings.Split(parts[0], "/")

	if len(parts) < 2 {
		return nil, fmt.Errorf("failed to parse platform '%s': expected format os/arch[/variant]", platform)
	}
	if len(parts) > 3 {
		return nil, fmt.Errorf("failed to parse platform '%s': too many slashes", platform)
	}

	p.OS = parts[0]
	p.Architecture = parts[1]
	if len(parts) > 2 {
		p.Variant = parts[2]
	}

	return p, nil
}
