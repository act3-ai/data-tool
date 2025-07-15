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

// Run release steps.
func (t *Tool) Release() *Releaser {
	return &Releaser{
		Tool: t,
	}
}

// Releaser provides utilties for preparing and publishing releases
// with git-cliff.
type Releaser struct {
	Tool *Tool
}

// Run linters, unit, functional, and integration tests.
func (r *Releaser) Check(ctx context.Context) (string, error) {
	err := r.diffGenAll(ctx)
	if err != nil {
		return "", err
	}

	// lint, unit test
	_, err = dag.Release(r.Tool.Source).
		Go().
		Check(ctx,
			dagger.ReleaseGolangCheckOpts{
				UnitTestBase: dag.Go().
					WithSource(r.Tool.Source).
					Container().
					WithExec([]string{"apt", "update"}).
					WithExec([]string{"apt", "install", "-y", "git-lfs"}),
			},
		)
	if err != nil {
		return "", fmt.Errorf("running linters and unit tests: %w", err)
	}

	// functional test
	_, err = r.Tool.Test().Functional(ctx)
	if err != nil {
		return "", fmt.Errorf("running functional tests: %w", err)
	}

	// integration test
	// _, err = r.Tool.Test().Integration(ctx)
	// if err != nil {
	// 	return "", fmt.Errorf("running integration tests: %w", err)
	// }

	return "Successfully passed linters and tests", nil
}

// Update the version, changelog, and release notes.
func (r *Releaser) Prepare(ctx context.Context,
	// ignore git status checks
	// +optional
	ignoreError bool,
	// release with a specific version
	// +optional
	version string,
) (*dagger.Directory, error) {
	var err error

	targetVersion := version
	if targetVersion == "" {
		targetVersion, err = r.Version(ctx)
		if err != nil {
			return nil, fmt.Errorf("resolving release target version: %w", err)
		}
	}

	// Note: Changes to existing or inclusions of additional image references
	// should be reflected here, see published images in ../bin/release.sh publish stage.
	b := &strings.Builder{}
	b.WriteString("| Images |\n")
	b.WriteString("| ---------------------------------------------------- |\n")
	fmt.Fprintf(b, "| %s/%s:%s |\n\n", reg, repo, targetVersion)

	return dag.Release(r.Tool.Source).
		Prepare(dagger.ReleasePrepareOpts{
			Version:     targetVersion,
			ExtraNotes:  b.String(),
			IgnoreError: ignoreError},
		), nil
}

// Create release and publish artifacts. This should already be tagged.
func (r *Releaser) Publish(ctx context.Context,
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
	version, err := r.Tool.Source.File("VERSION").Contents(ctx)
	if err != nil {
		return "", err
	}
	version = strings.TrimSpace(version)
	vVersion := "v" + version
	notesPath := filepath.Join("releases", vVersion+".md")
	imagePlatforms := []dagger.Platform{"linux/amd64", "linux/arm64"}

	_, err = dag.Goreleaser(r.Tool.Source, dagger.GoreleaserOpts{Version: "v2.9"}).
		// env vars defined in .goreleaser.yaml
		WithSecretVariable("GITHUB_TOKEN", token).
		WithSecretVariable("SSH_PRIVATE_KEY", sshPrivateKey).
		WithEnvVariable("RELEASE_AUTHOR", author).
		WithEnvVariable("RELEASE_AUTHOR_EMAIL", email).
		WithEnvVariable("RELEASE_LATEST", strconv.FormatBool(latest)).
		Release().
		WithFailFast().
		WithNotes(r.Tool.Source.File(notesPath)).
		Run(ctx)
	if err != nil {
		return "", fmt.Errorf("creating release: %w", err)
	}

	regRepo := path.Join("%s/%s", reg, repo)
	extraTags, err := dag.Release(r.Tool.Source).ExtraTags(ctx, regRepo, vVersion)
	if err != nil {
		return "", fmt.Errorf("resolving extra image tags: %w", err)
	}
	_, err = r.Tool.ImageIndex(ctx, vVersion, imagePlatforms, regRepo, extraTags)
	if err != nil {
		return "", fmt.Errorf("publishing image index: %w", err)
	}

	return "Successfully created release and uploaded images", nil
}

// Generate the next version from conventional commit messages (see cliff.toml). Includes 'v' prefix.
func (r *Releaser) Version(ctx context.Context) (string, error) {
	targetVersion, err := dag.GitCliff(r.Tool.Source).
		BumpedVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving release target version: %w", err)
	}

	return strings.TrimSpace(targetVersion), err
}

// diffGenAll runs all auto generators, comparing it's output to what currently exists in the source.
func (r *Releaser) diffGenAll(ctx context.Context) error {
	existing := dag.Directory().
		WithDirectory(cliDocsPath, r.Tool.Source.Directory(cliDocsPath)).
		WithDirectory(apiDocsPath, r.Tool.Source.Directory(apiDocsPath).Filter(dagger.DirectoryFilterOpts{Exclude: []string{"schemas/"}})).
		WithDirectory(pkgPath, r.Tool.Source.Directory(pkgPath))

	regen := r.Tool.GenAll(ctx)

	existingDgst, err := existing.Digest(ctx)
	if err != nil {
		return err
	}

	regenDgst, err := regen.Digest(ctx)
	if err != nil {
		return err
	}

	if existingDgst != regenDgst {
		return fmt.Errorf("found changes from running auto generators, please run 'dagger call gen-all export --path=.'")
	}
	return nil
}
