package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Run release steps.
func (t *Tool) Release(
	// top level source code directory
	// +defaultPath="/"
	src *dagger.Directory,
	// GitHub token
	// +optional
	token *dagger.Secret,
) *Release {
	return &Release{
		Source: src,
		Token:  token,
	}
}

const (
	releaseNotesDir = "releases"
	changelogPath   = "CHANGELOG.md"
)

// Release provides utilties for preparing and publishing releases
// with git-cliff.
type Release struct {
	// source code directory
	// +defaultPath="/"
	Source *dagger.Directory

	// GitHub token
	// +optional
	Token *dagger.Secret
}

// Update the version, changelog, and release notes.
func (r *Release) Prepare(ctx context.Context) (*dagger.Directory, error) {
	targetVersion, err := r.Version(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolving release target version: %w", err)
	}

	changelogFile, err := r.Changelog(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating changelog: %w", err)
	}

	releaseNotesFile, err := r.Notes(ctx, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}

	releaseNotesPath := filepath.Join(releaseNotesDir, targetVersion+".md")
	return dag.Directory().
		WithFile(changelogPath, changelogFile).
		WithFile(releaseNotesPath, releaseNotesFile).
		WithNewFile("VERSION", strings.TrimPrefix(targetVersion+"\n", "v")), nil
}

// Publish the current release. This should be tagged.
func (t *Release) Publish(ctx context.Context,
	// source code directory
	// +defaultPath="/"
	src *dagger.Directory,
	// github personal access token
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
	version, err := src.File("VERSION").Contents(ctx)
	if err != nil {
		return "", err
	}
	version = strings.TrimSpace(version)
	vVersion := "v" + version

	notesPath := filepath.Join("releases", vVersion+".md")

	return dag.Goreleaser(t.Source).
		WithSecretVariable("GITHUB_TOKEN", token).
		WithSecretVariable("SSH_PRIVATE_KEY", sshPrivateKey).
		WithEnvVariable("RELEASE_AUTHOR", author).
		WithEnvVariable("RELEASE_AUTHOR_EMAIL", email).
		WithEnvVariable("RELEASE_LATEST", strconv.FormatBool(latest)).
		Release().
		WithFailFast().
		WithNotes(t.Source.File(notesPath)).
		Run().
		Stdout(ctx)
	// return GoReleaser(src).
	// 	WithSecretVariable("GITHUB_TOKEN", token).
	// 	WithSecretVariable("SSH_PRIVATE_KEY", sshPrivateKey).
	// 	WithEnvVariable("RELEASE_AUTHOR", author).
	// 	WithEnvVariable("RELEASE_AUTHOR_EMAIL", email).
	// 	WithEnvVariable("RELEASE_LATEST", strconv.FormatBool(latest)).
	// 	WithExec([]string{"goreleaser", "release", "--fail-fast", "--release-notes", notesPath}).
	// 	Stdout(ctx)
}

// Generate the change log from conventional commit messages (see cliff.toml).
func (r *Release) Changelog(ctx context.Context) (*dagger.File, error) {
	// generate and prepend to changelog
	changelogPath := "CHANGELOG.md"
	return dag.GitCliff(r.Source).
		WithBump().
		WithStrip("footer").
		WithUnreleased().
		WithPrepend(changelogPath).
		Run().
		File(changelogPath), nil
}

// Generate the next version from conventional commit messages (see cliff.toml). Includes 'v' prefix.
func (r *Release) Version(ctx context.Context) (string, error) {
	targetVersion, err := dag.GitCliff(r.Source).
		BumpedVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving release target version: %w", err)
	}

	return strings.TrimSpace(targetVersion), err
}

// Generate the initial release notes.
func (r *Release) Notes(ctx context.Context,
	// release version
	version string,
) (*dagger.File, error) {

	// generate and export release notes
	notes, err := dag.GitCliff(r.Source).
		WithBump().
		WithUnreleased().
		WithStrip("all").
		Run().
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating release notes: %w", err)
	}

	// Note: Changes to existing or inclusions of additional image references
	// should be reflected here, see published images in ../bin/release.sh publish stage.
	b := &strings.Builder{}
	b.WriteString("| Images |\n")
	b.WriteString("| ---------------------------------------------------- |\n")
	fmt.Fprintf(b, "| ghcr.io/act3-ai/data-tool:%s |\n\n", version)

	b.WriteString("### ")
	notes = strings.Replace(notes, "### ", b.String(), 1)

	notesFilePath := "release-notes.md"
	return dag.Directory().
		WithNewFile(notesFilePath, notes).
		File(notesFilePath), nil
}
