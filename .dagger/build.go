package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"path"
	"strings"

	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"oras.land/oras-go/v2/registry"
)

// Generate a directory of ace-dt executables built for all supported platforms, concurrently.
func (t *Tool) BuildPlatforms(ctx context.Context,
	// release version
	// +optional
	version string,
) (*dagger.Directory, error) {
	// build matrix
	gooses := []string{"linux", "windows", "darwin"}
	goarches := []string{"amd64", "arm64"}

	ctx, span := Tracer().Start(ctx, "Build Platforms", trace.WithAttributes(attribute.StringSlice("GOOS", gooses), attribute.StringSlice("GOARCH", goarches)))
	defer span.End()

	buildsDir := dag.Directory()
	p := pool.NewWithResults[*dagger.File]().WithContext(ctx)

	for _, goos := range gooses {
		for _, goarch := range goarches {
			p.Go(func(ctx context.Context) (*dagger.File, error) {
				platform := fmt.Sprintf("%s/%s", goos, goarch)
				bin := t.Build(ctx, dagger.Platform(platform), version, "latest")
				return bin, nil
			})
		}
	}

	bins, err := p.Wait()
	if err != nil {
		return nil, err
	}
	return buildsDir.WithFiles(".", bins), nil
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
	// Release version, included in file name
	// +optional
	version string,
	// value of GOFIPS140, accepts modes "off", "latest", and "v1.0.0"
	// +optional
	// +default="latest"
	fipsMode string,
) *dagger.File {
	return build(ctx, t.Source, t.Netrc, platform, version, fipsMode)
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
		From("cgr.dev/chainguard/static").
		WithFile("/usr/local/bin/ace-dt", t.Build(ctx, platform, "", "latest")).
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
				WithLabel("org.opencontainers.image.url", imgURL)
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
	netrc *dagger.Secret,
	platform dagger.Platform,
	version string,
	fipsMode string,
) *dagger.File {
	name := binaryName(string(platform), version)

	_, span := Tracer().Start(ctx, fmt.Sprintf("Build %s", name))
	defer span.End()

	return dag.Go(
		dagger.GoOpts{
			Container: dag.Container().
				From(imageGo).                            // same as dag.Go, but...
				WithMountedSecret("/root/.netrc", netrc), // allows us to mount this secret
		}).
		WithSource(src).
		WithCgoDisabled().
		WithEnvVariable("GO_PRIVATE", gitlabHost).
		WithEnvVariable("GOFIPS140", fipsMode).
		Build(dagger.GoWithSourceBuildOpts{
			Pkg:      "./cmd/ace-dt",
			Platform: platform,
			Ldflags:  []string{"-s", "-w", fmt.Sprintf("-X 'main.version=%s'", version)},
			Trimpath: true,
		}).
		WithName(name)
}

// binaryName constructs the name of a ace-dt executable based on build params.
// All arguments are optional, building up to "ace-dt-v{VERSION}-fips-{GOOS}-{GOARCH}".
func binaryName(platform string, version string) string {
	str := strings.Builder{}
	str.Grow(32) // est. max = len("ace-dt-v1.11.11-fips-linux-amd64")
	str.WriteString("ace-dt")

	if version != "" {
		str.WriteString("-v")
		str.WriteString(version)
	}

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
		// TODO: Different keys, same values feels off; but reflects the legacy release process
		WithLabel("org.opencontainers.image.url", path.Join(gitlabHost, gitlabProject)).
		WithLabel("org.opencontainers.image.source", path.Join(gitlabHost, gitlabProject)).
		WithLabel("org.opencontainers.image.description", "ACE Data Tool")
}
