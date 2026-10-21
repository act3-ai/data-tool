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
	githubHost = "github.com"
	acedtCLI   = "ace-dt"

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
	// +private
	GithubUser string
	// +private
	GithubToken *dagger.Secret
	// +private
	SSHKey *dagger.Secret
	// +private
	Email string
}

func New(
	// GitHub username
	// +optional
	githubUser string,
	// GitHub token
	// +optional
	githubToken *dagger.Secret,
	// ssh key for homebrew cask PR
	// +optional
	sshKey *dagger.Secret,
	// email for ssh key
	// +optional
	email string,
) *DataTool {
	return &DataTool{
		GithubUser:  githubUser,
		GithubToken: githubToken,
	}
}
