package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/conc/pool"
	"oras.land/oras-go/v2/registry"
)

// Build an executable for the specified platform, named "ace-dt-v{VERSION}-{GOOS}-{GOARCH}".
//
// Supported Platform Matrix:
//
//	GOOS: linux, windows, darwin
//	GOARCH: amd64, arm64
func (m *DataTool) Build(ctx context.Context,
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
	// Build target platform
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
	// snapshot build, skip goreleaser validations
	// +optional
	snapshot bool,
) *dagger.File {
	return build(src, platform, snapshot)
}

// Create a container with an ace-dt executable.
func (m *DataTool) Container(ctx context.Context,
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
	// image version
	version string,
	// Build target platform
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
) *dagger.Container {
	ctr := dag.Container(dagger.ContainerOpts{Platform: platform}).
		From(imageChainguard).
		WithFile(filepath.Join("/usr/local/bin", acedtCLI), m.Build(ctx, src, platform, false)).
		WithEntrypoint([]string{acedtCLI}).
		WithWorkdir("/")
	return withCommonLabels(ctr, version)
}

// Create and publish a multi-platform image index.
func (m *DataTool) imageIndex(ctx context.Context,
	src *dagger.Directory,
	// image version
	version string,
	// build platforms
	platforms []dagger.Platform,
	// OCI Reference, without tag
	address string,
	// extraTags
	extraTags []string,
) (string, error) {
	ref, err := registry.ParseReference(address)
	if err != nil {
		return "", fmt.Errorf("parsing address: %w", err)
	}
	imgURL := "https://" + path.Join(ref.Registry, ref.Repository)

	// build platforms
	p := pool.NewWithResults[*dagger.Container]().WithContext(ctx)
	for _, platform := range platforms {
		p.Go(func(ctx context.Context) (*dagger.Container, error) {
			img := m.Container(ctx, src, version, platform).
				WithLabel("org.opencontainers.image.url", imgURL).
				WithLabel("org.opencontainers.image.source", "https://github.com/act3-ai/data-tool")
			return img, nil
		})
	}

	platformVariants, err := p.Wait()
	if err != nil {
		return "", fmt.Errorf("building images: %w", err)
	}

	var result strings.Builder

	// publish index
	taggedRef := fmt.Sprintf("%s:%s", address, version)
	dgstRef, err := dag.Container().
		Publish(ctx, taggedRef, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
		})
	if err != nil {
		return "", fmt.Errorf("publishing index to %s: %w", taggedRef, err)
	}
	result.WriteString("Successfully published image index to:\n")
	result.WriteString(fmt.Sprintf("%s\n", dgstRef))

	// apply extra tags to remote index
	if len(extraTags) > 0 {
		_, err := dag.Release(nil).AddTags(ctx, taggedRef, extraTags)
		if err != nil {
			return result.String(), fmt.Errorf("adding extra tags to image: %w", err)
		}
		for _, tag := range extraTags {
			result.WriteString(fmt.Sprintf("%s:%s\n", address, tag))
		}
	}
	return result.String(), nil
}

func build(
	src *dagger.Directory,
	platform dagger.Platform,
	// snapshot build, skip goreleaser validations
	snapshot bool,
) *dagger.File {
	return dag.Goreleaser(src).
		Build().
		WithTimeout("10m").
		WithAutoSnapshot().
		With(func(r *dagger.GoreleaserBuild) *dagger.GoreleaserBuild {
			if snapshot {
				r = r.WithSnapshot()
			}
			return r
		}).
		Platform("ace-dt", dagger.GoreleaserBuildPlatformOpts{Platform: platform})
}

// withCommonLabels applies common labels to a container, e.g. maintainers, vendor, etc.
func withCommonLabels(ctr *dagger.Container, version string) *dagger.Container {
	return ctr.
		WithLabel("maintainers", "Nathan D. Joslin <nathan.joslin@udri.udayton.edu>").
		WithLabel("org.opencontainers.image.vendor", "AFRL ACT3").
		WithLabel("org.opencontainers.image.version", version).
		WithLabel("org.opencontainers.image.title", "Tool").
		WithLabel("org.opencontainers.image.url", "ghcr.io/act3-ai/data-tool").
		WithLabel("org.opencontainers.image.source", "https://github.com/act3-ai/data-tool").
		WithLabel("org.opencontainers.image.description", "ACE Data Tool")
}
