package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	version "github.com/hashicorp/go-version"

	"github.com/act3-ai/go-common/pkg/logger"
)

// Minimum git version.
// Note: The login script requires git >= 2.34 for commit signing.
// v2.29.0 was chosen as this is the minimum version to support git bundles.
const minGitVersion = "2.29.0"

// Minimum git lfs version.
// v2.11.0 allows us to fetch lfs files without a worktree. https://github.com/git-lfs/git-lfs/releases/tag/v2.11.0
const minGitLFSVersion = "2.11.0"

// CheckGitVersion gets and validates a user's git version.
func CheckGitVersion(ctx context.Context, altExec string) (string, error) {
	log := logger.FromContext(ctx)

	if altExec != "" {
		log.InfoContext(ctx, "Using alternate git executable", "path", altExec)
	}

	gitVersion, err := getGitVersion(ctx, altExec)
	if err != nil {
		return "", fmt.Errorf("checking git version: %w", err)
	}
	log.InfoContext(ctx, "Git version resolved", "version", gitVersion)

	err = validGitVersion(gitVersion)
	if err != nil {
		return gitVersion, fmt.Errorf("validating git version: %w", err)
	}

	return gitVersion, nil
}

var gitVersionRegex = regexp.MustCompile(`git version (\d*\.\d*\.\d*)`)

// getGitVersion shells out and parses the version of git being used. Returns major, minor, patch.
// gitExec is the path to the git executable (default is "git").
func getGitVersion(ctx context.Context, gitExec string) (string, error) {
	if gitExec == "" {
		gitExec = "git"
	}
	buf, err := exec.CommandContext(ctx, gitExec, "version").Output()
	if err != nil {
		return "", fmt.Errorf("running git cmd: %w", err)
	}

	matches := gitVersionRegex.FindSubmatch(buf)
	if len(matches) != 2 {
		return "", fmt.Errorf("git version does not match expected format %s: %w", gitVersionRegex, err)
	}
	return string(matches[1]), nil
}

// validGitVersion returns true if the provided version meets the globally specified minimum requirement.
func validGitVersion(v string) error {
	return validVersion(v, minGitVersion)
}

// CheckGitLFSVersion gets and validates a user's git lfs version.
func CheckGitLFSVersion(ctx context.Context, altExec string) (string, error) {
	log := logger.FromContext(ctx)

	if altExec != "" {
		log.InfoContext(ctx, "Using alternate git-lfs executable", "path", altExec)
	}

	gitLFSVersion, err := getGitLFSVersion(ctx, altExec)
	if err != nil {
		return "", fmt.Errorf("getting git lfs version: %w", err)
	}
	log.InfoContext(ctx, "Git lfs version resolved", "version", gitLFSVersion)

	err = validGitLFSVersion(gitLFSVersion)
	if err != nil {
		return gitLFSVersion, fmt.Errorf("validating git lfs version: %w", err)
	}

	return gitLFSVersion, nil
}

// getGitLFSVersion shells out and parses the version of git lfs being used. Returns major, minor, patch.
func getGitLFSVersion(ctx context.Context, altExec string) (string, error) {
	lfsgc := &gitLFSCmd{
		dir:           "",
		altGitLFSExec: altExec,
	}

	v, err := lfsgc.Version(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving git lfs version: %w", err)
	}

	v = strings.TrimPrefix(v, "git-lfs/")
	cutIdx := strings.Index(v, " ")
	return v[:cutIdx], nil
}

// validGitLFSVersion returns true if the provided version meets the globally specified minimum requirement.
func validGitLFSVersion(v string) error {
	return validVersion(v, minGitLFSVersion)
}

// validVersion compares a version to a minimum requirement.
func validVersion(got, minimum string) error {
	gotVer, err := version.NewSemver(strings.TrimSpace(got))
	if err != nil {
		return fmt.Errorf("parsing received semantic version %s: %w", got, err)
	}

	minVer, err := version.NewSemver(minimum)
	if err != nil {
		return fmt.Errorf("parsing minimum semantic version %s: %w", minimum, err)
	}

	if gotVer.LessThan(minVer) {
		return fmt.Errorf("got: %s < minimum: %s: %w", got, minimum, errInsufficientGitVersion)
	}

	return nil
}
