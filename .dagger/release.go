package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Run release steps.
func (t *Tool) Release(
	// top level source code directory
	// +defaultPath="/"
	src *dagger.Directory,
	// gitlab token
	// +optional
	token *dagger.Secret,
) *Release {
	return &Release{
		Source: src,
		Token:  token,
	}
}

// Release provides utilties for preparing and publishing releases
// with git-cliff.
type Release struct {
	// source code directory
	// +defaultPath="/"
	Source *dagger.Directory

	// GitLab token
	// +optional
	Token *dagger.Secret
}

// Update the changelog, release notes, version, and helm chart versions.
func (r *Release) Prepare(ctx context.Context) (*dagger.Directory, error) {
	changelog := r.Changelog(ctx)
	// version, err := r.Version(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	version := "1.15.8"

	notes, err := r.Notes(ctx, version)
	if err != nil {
		return nil, err
	}

	notesPath := filepath.Join("releases", fmt.Sprintf("v%s.md", version))
	return dag.Directory().
			WithFile("CHANGELOG.md", changelog).
			WithNewFile("VERSION", version+"\n").
			WithNewFile(notesPath, notes),
		nil
}

// Publish the current release. This should be tagged.
func (t *Tool) Publish(ctx context.Context,
	// source code directory
	// +defaultPath="/"
	src *dagger.Directory,
	// gitlab personal access token
	token *dagger.Secret,
) (string, error) {
	version, err := src.File("VERSION").Contents(ctx)
	if err != nil {
		return "", err
	}
	version = strings.TrimSpace(version)

	notesFileName := fmt.Sprintf("v%s.md", version)
	notes := src.File(filepath.Join("releases", notesFileName))

	return t.createRelease(ctx, version, notes, token)
}

// Generate the change log from conventional commit messages (see cliff.toml)
func (r *Release) Changelog(ctx context.Context) *dagger.File {
	const changelogPath = "/app/CHANGELOG.md"
	return r.gitCliffContainer().
		// WithExec([]string{"git-cliff", "--bump", "--strip=footer", "-o", changelogPath}).
		WithExec([]string{"git-cliff", "06f591bec27905323cb69295b390ec0ad589c6eb..50feb02fd08178e758f6e5ac6abf7b4e0e28485b", "--tag", "v1.15.8", "--prepend", changelogPath}).
		File(changelogPath)
}

// Generate the next version from conventional commit messages (see cliff.toml)
func (r *Release) Version(ctx context.Context) (string, error) {
	version, err := r.gitCliffContainer().
		WithExec([]string{"git-cliff", "--bumped-version"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(version)[1:], err
}

// Generate the initial release notes
func (r *Release) Notes(ctx context.Context,
	// helm chart version
	chartVersion string,
) (string, error) {
	return r.gitCliffContainer().
		WithExec([]string{"git-cliff", "--bump", "--unreleased", "--strip=all"}).
		Stdout(ctx)

}

// create a release for an existing tag.
func (t *Tool) createRelease(ctx context.Context,
	// release version
	version string,
	// release notes file
	notes *dagger.File,
	// gitlab personal access token
	token *dagger.Secret,
) (string, error) {
	notesFileName, err := notes.Name(ctx)
	if err != nil {
		return "", err
	}
	return dag.Container().
		From(imageGitlabCLI).
		WithMountedFile(notesFileName, notes).
		WithSecretVariable("GITLAB_TOKEN", token).
		WithEnvVariable("GITLAB_HOST", gitlabHost).
		WithExec([]string{"glab", "release", "create",
			"-R", gitlabProject, // repository
			"v" + version,                 // tag
			"--name=Release v" + version,  // title
			"--notes-file", notesFileName, // description
		}).
		Stdout(ctx)
}

func (r *Release) gitCliffContainer() *dagger.Container {
	return dag.Container().
		From(imageGitCliff).
		With(func(c *dagger.Container) *dagger.Container {
			if r.Token != nil {
				return c.WithSecretVariable("GITLAB_TOKEN", r.Token).
					WithEnvVariable("GITLAB_API_URL", path.Join(gitlabHost, "/api/v4")).
					WithEnvVariable("GITLAB_REPO", gitlabProject)
			}
			return c
		}).
		WithMountedDirectory("/app", r.Source)
}

// UploadAssets publishes binaries as assets to an existing release tag.
func (r *Release) UploadAssets(ctx context.Context,
	// release version
	version string,
	// release assets
	assets *dagger.Directory,
	// gitlab personal access token
	token *dagger.Secret,
) (string, error) {
	releaseAssets, err := assets.Entries(ctx)
	if err != nil {
		return "", err
	}

	ctx, span := Tracer().Start(ctx, "Upload Builds", trace.WithAttributes(attribute.StringSlice("Assets", releaseAssets)))
	defer span.End()

	// remove unwanted items that exist in bin dir
	cleanedAssets := slices.DeleteFunc(releaseAssets, func(s string) bool {
		return !regexp.MustCompile("ace-dt-*").MatchString(s)
	})

	p := pool.NewWithResults[string]().WithContext(ctx)
	for _, asset := range cleanedAssets {
		p.Go(func(ctx context.Context) (string, error) {
			_, err := r.uploadBuild(ctx, version, assets.File(asset), token)
			if err != nil {
				return fmt.Sprintf("Failed to upload asset - %s", asset), err
			}
			return fmt.Sprintf("Asset Uploaded - %s", asset), nil
		})
	}

	result, err := p.Wait()
	return strings.Join(result, "\n"), err
}

func (r *Release) uploadBuild(ctx context.Context,
	// release version
	version string,
	// build file
	build *dagger.File,
	// gitlab personal access token
	token *dagger.Secret,
) (string, error) {
	buildName, err := build.Name(ctx)
	if err != nil {
		return "", err
	}
	ctx, span := Tracer().Start(ctx, fmt.Sprintf("upload release asset %s", buildName))
	defer span.End()

	return dag.Container().
		From(imageGitlabCLI).
		WithMountedFile(buildName, build).
		WithSecretVariable("GITLAB_TOKEN", token).
		WithEnvVariable("GITLAB_HOST", gitlabHost).
		WithExec([]string{"glab", "release", "upload",
			"-R", gitlabProject,
			"v" + version,
			buildName},
		).
		Stdout(ctx)
}

// announce the release on Mattermost
func (r *Release) Announce(ctx context.Context,
	// Mattermost server base URL
	// +default="https://chat.git.act3-ace.com"
	serverURL string,

	// Mattermost team
	// +default="act3"
	team string,

	// Mattermost channel
	// +default="ace-dt"
	channel string,

	// Mattermost personal access token
	token *dagger.Secret,

	// source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) (string, error) {
	version, err := src.File("VERSION").Contents(ctx)
	if err != nil {
		return "", err
	}
	version = strings.TrimSpace(version)

	notes, err := src.File(fmt.Sprintf("releases/v%s.md", version)).Contents(ctx)
	if err != nil {
		return "", err
	}

	bearerToken, err := token.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	message := fmt.Sprintf("🎉 **ACE Data Tool [v%[1]s](https://git.act3-ace.com/ace/data/tool/-/releases/v%[1]s) has been released.** 🎉\n\n%s", version, notes)

	c := model.NewAPIv4Client(serverURL)
	c.AuthToken = bearerToken
	c.AuthType = "Bearer"

	// resolve the team and channel to a channel ID
	ch, resp, err := c.GetChannelByNameForTeamName(ctx, channel, team, "")
	if err != nil {
		log.Fatal(resp, err)
	}

	post, _, err := c.CreatePost(ctx, &model.Post{
		ChannelId: ch.Id,
		IsPinned:  true,
		Message:   message,
	})
	if err != nil {
		return "", err
	}

	return post.Id, nil
}
