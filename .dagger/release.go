package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	reg  = "ghcr.io"
	repo = "act3-ai/data-tool"
)

// Update the version, changelog, and release notes.
func (m *DataTool) PrepareRelease(ctx context.Context,
	// Git reference to release (can use "." for the current commit)
	// +defaultPath="."
	gitRef *dagger.GitRef,
	// release with a specific version
	// +optional
	version string,
) (*dagger.Changeset, error) {
	release := dag.Release(gitRef)

	// compute the version if not provided
	if version == "" {
		v, err := release.Version(ctx)
		if err != nil {
			return nil, err
		}
		version = v
	}

	// Note: Changes to existing or inclusions of additional image references
	// should be reflected here, see published images in ../bin/release.sh publish stage.
	b := &strings.Builder{}
	b.WriteString("| Images |\n")
	b.WriteString("| ---------------------------------------------------- |\n")
	fmt.Fprintf(b, "| %s/%s:%s |\n\n", reg, repo, version)

	return release.Prepare(version, dagger.ReleasePrepareOpts{
		ExtraNotes: b.String(),
	}), nil
}

// Create release and publish artifacts. This should already be tagged.
func (m *DataTool) PublishReleasen(ctx context.Context,
	// github API token
	token *dagger.Secret,
	// commit ssh private key
	sshPrivateKey *dagger.Secret,
	// releaser username
	author string,
	//releaser email
	email string,
	// tag release as latest
	// +default=true
	// +optional
	latest bool,
) (string, error) {
	version, err := m.Source.File("VERSION").Contents(ctx)
	if err != nil {
		return "", err
	}
	version = strings.TrimSpace(version)
	vVersion := "v" + version
	notesPath := filepath.Join("releases", vVersion+".md")
	imagePlatforms := []dagger.Platform{"linux/amd64", "linux/arm64"}

	_, err = dag.Goreleaser(m.Source, dagger.GoreleaserOpts{Version: "v2.9"}).
		// env vars defined in .goreleaser.yaml
		WithSecretVariable("GITHUB_TOKEN", token).
		WithSecretVariable("SSH_PRIVATE_KEY", sshPrivateKey).
		WithEnvVariable("RELEASE_AUTHOR", author).
		WithEnvVariable("RELEASE_AUTHOR_EMAIL", email).
		WithEnvVariable("RELEASE_LATEST", strconv.FormatBool(latest)).
		Release().
		WithFailFast().
		WithNotes(m.Source.File(notesPath)).
		Run(ctx)
	if err != nil {
		return "", fmt.Errorf("creating release: %w", err)
	}

	regRepo := path.Join("%s/%s", reg, repo)
	extraTags, err := dag.Release(nil).ExtraTags(ctx, regRepo, vVersion)
	if err != nil {
		return "", fmt.Errorf("resolving extra image tags: %w", err)
	}
	_, err = m.ImageIndex(ctx, vVersion, imagePlatforms, regRepo, extraTags)
	if err != nil {
		return "", fmt.Errorf("publishing image index: %w", err)
	}

	return "Successfully created release and uploaded images", nil
}
