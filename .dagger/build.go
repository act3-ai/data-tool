package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"path"
	"strings"

	"github.com/sourcegraph/conc/pool"
	"oras.land/oras-go/v2/registry"
)

// Generate a directory of ace-dt executables built for all supported platforms, concurrently.
//
//	 Platform Matrix:
//
//		GOOS: linux, windows, darwin
//		GOARCH: amd64, arm64
func (t *Tool) BuildPlatforms(ctx context.Context,
	// snapshot build, skip goreleaser validations
	// +optional
	snapshot bool,
) *dagger.Directory {
	return GoReleaser(t.Source).
		WithExec([]string{"goreleaser", "build", "--clean", "--auto-snapshot", "--timeout=10m", fmt.Sprintf("--snapshot=%v", snapshot)}).
		Directory("dist")
}

// Build an executable for the specified platform, named "ace-dt-v{VERSION}-{GOOS}-{GOARCH}".
//
// Supported Platform Matrix:
//
//	GOOS: linux, windows, darwin
//	GOARCH: amd64, arm64
func (t *Tool) Build(ctx context.Context,
	// Build target platform
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
	// snapshot build, skip goreleaser validations
	// +optional
	snapshot bool,
) *dagger.File {
	return build(ctx, t.Source, platform, snapshot)
}

// Create an image with an ace-dt executable.
func (t *Tool) Image(ctx context.Context,
	// image version
	version string,
	// Build target platform
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
) *dagger.Container {
	ctr := dag.Container(dagger.ContainerOpts{Platform: platform}).
		From(imageChainguard).
		WithFile("/usr/local/bin/ace-dt", t.Build(ctx, platform, false)).
		WithEntrypoint([]string{"ace-dt"}).
		WithWorkdir("/")
	return withCommonLabels(ctr, version)
}

// Create and publish a multi-platform image index.
func (t *Tool) ImageIndex(ctx context.Context,
	// image version
	version string,
	// OCI Reference
	address string,
	// build platforms
	platforms []dagger.Platform,
) (string, error) {
	ref, err := registry.ParseReference(address)
	if err != nil {
		return "", fmt.Errorf("parsing address: %w", err)
	}
	imgURL := "https://" + path.Join(ref.Registry, ref.Repository)

	p := pool.NewWithResults[*dagger.Container]().WithContext(ctx)
	for _, platform := range platforms {
		p.Go(func(ctx context.Context) (*dagger.Container, error) {
			img := t.Image(ctx, version, platform).
				WithLabel("org.opencontainers.image.url", imgURL).
				WithLabel("org.opencontainers.image.source", "https://github.com/act3-ai/data-tool")
			return img, nil
		})
	}

	platformVariants, err := p.Wait()
	if err != nil {
		return "", fmt.Errorf("building images: %w", err)
	}

	return dag.Container().
		Publish(ctx, address, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
		})
}

func build(ctx context.Context,
	src *dagger.Directory,
	platform dagger.Platform,
	// snapshot build, skip goreleaser validations
	snapshot bool,
) *dagger.File {
	name := binaryName(string(platform))

	_, span := Tracer().Start(ctx, fmt.Sprintf("Build %s", name))
	defer span.End()

	os, arch, _ := strings.Cut(string(platform), "/")
	return GoReleaser(src).
		WithEnvVariable("GOOS", os).
		WithEnvVariable("GOARCH", arch).
		WithExec([]string{"goreleaser", "build", "--auto-snapshot", "--timeout=10m", "--single-target", "--output", name, fmt.Sprintf("--snapshot=%v", snapshot)}).
		File(name)
}

// binaryName constructs the name of a ace-dt executable, as generated by goreleaser.
func binaryName(platform string) string {
	str := strings.Builder{}
	str.WriteString("ace-dt")

	if platform != "" {
		platform = strings.ReplaceAll(string(platform), "/", "-")
		str.WriteString("-")
		str.WriteString(platform)
	}

	return str.String()
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
