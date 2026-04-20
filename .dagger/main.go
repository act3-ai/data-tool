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
	"dagger/tool/internal/dagger"
)

const (
	// images
	imageGrype      = "anchore/grype:latest"
	imageSyft       = "anchore/syft:latest"
	imageRegistry   = "docker.io/library/registry:3.0.0-rc.3"
	imageTelemetry  = "ghcr.io/act3-ai/data-telemetry/slim:latest"
	imageChainguard = "cgr.dev/chainguard/static"
	imagePostgres   = "postgres:17-alpine"
	imageGoReleaser = "ghcr.io/goreleaser/goreleaser:v2.15.3"
)

type DataTool struct {
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
) *DataTool {
	return &DataTool{
		Source:         src,
		RegistryConfig: dag.RegistryConfig(),
	}
}

// Add credentials for a registry.
func (m *DataTool) WithRegistryAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *DataTool {
	m.RegistryConfig = m.RegistryConfig.WithRegistryAuth(address, username, secret)
	return m
}

// Removes credentials for a registry.
func (m *DataTool) WithoutRegistryAuth(
	// registry's hostname
	address string,
) *DataTool {
	m.RegistryConfig = m.RegistryConfig.WithoutRegistryAuth(address)
	return m
}

// Add netrc credentials for a private git repository.
func (m *DataTool) WithNetrc(
	// NETRC credentials
	netrc *dagger.Secret,
) *DataTool {
	m.Netrc = netrc
	return m
}
