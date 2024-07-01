package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// Git provides access to git commands.
type Git interface {
	Init(ctx context.Context, args ...string) error
	CloneWithShared(ctx context.Context, gitRemote, reference string) error
	Config(ctx context.Context, args ...string) error
	Push(ctx context.Context, gitRemote string, refs ...string) error
	Fetch(ctx context.Context, args ...string) error
	BundleCreate(ctx context.Context, destFile string, revList []string) error
	ShowRefs(ctx context.Context, refs ...string) ([]string, error)
	UpdateRef(ctx context.Context, ref string, commit string) error
	RemoteAdd(ctx context.Context, shortname, remoteTarget string) error
	RemoteRemove(ctx context.Context, shortname string) error
	LSRemote(ctx context.Context, args ...string) ([]string, error)
	MergeBase(ctx context.Context, args ...string) error
	CatFile(ctx context.Context, args ...string) error
	Run(ctx context.Context, subCmd string, args ...string) ([]string, error)
}

// gitCmd contains a logger and directory of execution for a git command.
// gitCmd implements Git.
type gitCmd struct {
	dir string

	altGitExec string // optional
}

// Run executes a git command, returning the parsed output.
func (gc *gitCmd) Run(ctx context.Context, subCmd string, args ...string) ([]string, error) {
	log := logger.FromContext(ctx)

	var cmd *exec.Cmd
	switch {
	case gc.altGitExec != "":
		cmd = exec.CommandContext(ctx, gc.altGitExec)
	default:
		cmd = exec.CommandContext(ctx, "git")
	}

	cmd.Args = append(cmd.Args, subCmd)
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = gc.dir

	// We only want stdout for processing but we want stderr for errors
	out, err := cmd.Output()
	log.InfoContext(ctx, "Ran git command", "command", args, "directory", gc.dir, "output", string(out))
	parsedOut := parseGitOutput(out)
	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			log.InfoContext(ctx, "Command exit error", "err", exitError.Stderr)
			errStr := string(exitError.Stderr)
			switch {
			case strings.Contains(errStr, "fatal: Refusing to create empty bundle."):
				return parsedOut, errors.Join(err, ErrEmptyBundle)
			case strings.Contains(errStr, "fatal: Could not read from remote repository."):
				return parsedOut, errors.Join(err, ErrRepoNotExistOrPermDenied)
			default:
				// extract git's error, join incase of future unwrapping (e.g. used by git merge-base)
				return parsedOut, errors.Join(err, fmt.Errorf("git %q produced an error: %s", cmd.Args, exitError.Stderr))
			}
		}
		return parsedOut, fmt.Errorf("git %q produced an error: %w", cmd.Args, err)
	}

	return parsedOut, nil
}

// ShowRefs calls `git show-ref <refs>...`
//
// i.e. returns the "commit SP fullRef" pair for all refs as resolved by git. (SP = space).
func (gc *gitCmd) ShowRefs(ctx context.Context, refs ...string) ([]string, error) {
	return gc.Run(ctx, "show-ref", refs...)
}

// UpdateRef calls `git update-ref <ref> <commit>` within the gitCmd's  directory.
func (gc *gitCmd) UpdateRef(ctx context.Context, ref string, commit string) error {
	_, err := gc.Run(ctx, "update-ref", ref, string(commit))
	return err
}

// RemoteAdd calls `git remote add <shortname> <remoteTarget>` within the gitCmd's  directory.
func (gc *gitCmd) RemoteAdd(ctx context.Context, shortname, remoteTarget string) error {
	_, err := gc.Run(ctx, "remote", "add", shortname, remoteTarget)
	return err
}

// RemoteRemove calls `git remote remove <shortname>` within the gitCmd's  directory.
func (gc *gitCmd) RemoteRemove(ctx context.Context, shortname string) error {
	_, err := gc.Run(ctx, "remote", "remove", shortname)
	return err
}

// LSRemote calls `git ls-remote <args>...`.
func (gc *gitCmd) LSRemote(ctx context.Context, args ...string) ([]string, error) {
	return gc.Run(ctx, "ls-remote", args...)
}

// Fetch calls `git fetch <args>...` within the gitCmd's  directory.
func (gc *gitCmd) Fetch(ctx context.Context, args ...string) error {
	_, err := gc.Run(ctx, "fetch", args...)
	return err
}

// Push calls `git push <gitRef> --tags` within the gitCmd's  directory.
//
// i.e. pushes a local git repository to the local/remote reference, with tags.
func (gc *gitCmd) Push(ctx context.Context, gitRemote string, refs ...string) error {
	args := []string{gitRemote}
	args = append(args, refs...)
	_, err := gc.Run(ctx, "push", args...)
	return err
}

// Init calls `git init` within the gitCmd's  directory.
func (gc *gitCmd) Init(ctx context.Context, args ...string) error {
	_, err := gc.Run(ctx, "init", args...)
	return err
}

// CloneWithShared calls `git clone --shared --reference-if-able <reference> --bare <gitRef> <gc.dir>`.
//
// Cloning with the shared option prevents copying objects to the clone. This is a safe operation
// as long as the cache is not pRuned between cloning and managing the clone.
func (gc *gitCmd) CloneWithShared(ctx context.Context, gitRemote, reference string) error {
	_, err := gc.Run(ctx, "clone", "--shared", "--reference-if-able", reference, "--bare", gitRemote, gc.dir)
	return err
}

// BundleCreate calls `git bundle create <destFile> <revList>...` within the gitCmd's  directory.
//
// i.e. creates a git bundle including all layers specified in revList, writing the bundle to the
// destination path.
func (gc *gitCmd) BundleCreate(ctx context.Context, destFile string, revList []string) error {
	args := make([]string, 0, len(revList)+2)
	args = append(args, "create", destFile)
	args = append(args, revList...)
	_, err := gc.Run(ctx, "bundle", args...)
	return err
}

// MergeBase calls `git merge-base <args>...` within the gitCmd's directory.
func (gc *gitCmd) MergeBase(ctx context.Context, args ...string) error {
	out, err := gc.Run(ctx, "merge-base", args...)

	var exitErr *exec.ExitError
	switch {
	case errors.As(err, &exitErr) && exitErr.ExitCode() == 1: // exit code 1 = false, if an actual err occurs then code > 1.
		return ErrNotAncestor
	case err != nil:
		return fmt.Errorf("running git merge-base \nOutput: %s: %w", out, err)
	default:
		return nil
	}
}

// Config calls `git config <args>...`
//
// Used for setting git config options.
func (gc *gitCmd) Config(ctx context.Context, args ...string) error {
	_, err := gc.Run(ctx, "config", args...)
	return err
}

func (gc *gitCmd) CatFile(ctx context.Context, args ...string) error {
	_, err := gc.Run(ctx, "cat-file", args...)
	return err
}
