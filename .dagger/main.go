// A generated module for Tool functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/tool/internal/dagger"
)

const (
	// images
	imageGitCliff   = "docker.io/orhunp/git-cliff:2.8.0"
	imageGrype      = "anchore/grype:latest"
	imageSyft       = "anchore/syft:latest"
	imageRegistry   = "docker.io/library/registry:3.0.0-rc.3"
	imageTelemetry  = "ghcr.io/act3-ai/data-telemetry/slim:latest"
	imageChainguard = "cgr.dev/chainguard/static"
	imagePostgres   = "postgres:17-alpine"
	imageGoReleaser = "ghcr.io/goreleaser/goreleaser:v2.8.2"

	// go tools
	goControllerGen = "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.2"
	goCrdRefDocs    = "github.com/elastic/crd-ref-docs@v0.1.0"
)

type Tool struct {
	// source code directory
	Source *dagger.Directory

	// +private
	RegistryConfig *dagger.RegistryConfig
	// +private
	Netrc *dagger.Secret
}

func New(
	// top level source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) *Tool {
	return &Tool{
		Source:         src,
		RegistryConfig: dag.RegistryConfig(),
	}
}

// Add credentials for a registry.
func (t *Tool) WithRegistryAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *Tool {
	t.RegistryConfig = t.RegistryConfig.WithRegistryAuth(address, username, secret)
	return t
}

// Removes credentials for a registry.
func (t *Tool) WithoutRegistryAuth(
	// registry's hostname
	address string,
) *Tool {
	t.RegistryConfig = t.RegistryConfig.WithoutRegistryAuth(address)
	return t
}

// Add netrc credentials for a private git repository.
func (t *Tool) WithNetrc(
	// NETRC credentials
	netrc *dagger.Secret,
) *Tool {
	t.Netrc = netrc
	return t
}

func (t *Tool) Renovate(ctx context.Context,
	// GitHub token with API access to the project(s) being renovated
	token *dagger.Secret,
) (string, error) {
	return dag.Renovate("act3-ai/data-tool", token, "https://github.com").
		Update(ctx)
}
