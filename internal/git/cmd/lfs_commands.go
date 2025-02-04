package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// LFS provides access to git-lfs commands.
type LFS interface {
	Version(ctx context.Context) (string, error)
	Track(ctx context.Context, pattern string) error
	Fetch(ctx context.Context, gitRemote string, argRevList ...string) error
	Push(ctx context.Context, gitRemote string, argRevList ...string) error
	LSFiles(ctx context.Context, args ...string) ([]string, error)
	Run(ctx context.Context, subCmd string, args ...string) ([]string, error)
}

// gitLFSCmd contains a logger and directory of execution for a git lfs command.
// gitLFSCmd implements LFS.
type gitLFSCmd struct {
	dir string

	altGitLFSExec string // optional
}

// Run executes a git lfs command, returning the parsed output.
func (lfs *gitLFSCmd) Run(ctx context.Context, subCmd string, args ...string) ([]string, error) {
	log := logger.FromContext(ctx)
	var cmd *exec.Cmd
	switch {
	case lfs.altGitLFSExec != "":
		cmd = exec.CommandContext(ctx, lfs.altGitLFSExec)
	default:
		cmd = exec.CommandContext(ctx, "git-lfs")
	}
	cmd.Args = append(cmd.Args, subCmd)
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = lfs.dir

	out, err := cmd.Output()
	log.InfoContext(ctx, "Ran git-lfs Command", "command", cmd.Args, "output", string(out))
	parsedOut := parseGitOutput(out)
	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			// in the base git command we log the error as we have a chance to recover,
			// but here we have no hope. Extract git's error.
			return parsedOut, errors.Join(err, fmt.Errorf("git %q produced an error: %s", cmd.Args, exitError.Stderr))
		}

		// if git-lfs is not installed it won't be an exit error
		execErr := &exec.Error{}
		if errors.As(err, &execErr) {
			log.ErrorContext(ctx, "exec error", "cmd", cmd.Args, "err", execErr.Error())
			if strings.Contains(string(execErr.Error()), "executable file not found in $PATH") {
				return parsedOut, errors.Join(err, ErrLFSCmdNotFound)
			}
		}
		return parsedOut, fmt.Errorf("git-lfs %s produced an error: %w", cmd.Args, err)
	}

	return parsedOut, nil
}

// Fetch calls `git lfs fetch <args>...`
//
// i.e. downloads git lfs objects at the provided refs. Requires a
// remote to be provided if a default is not set (same default as git fetch).
func (lfs *gitLFSCmd) Fetch(ctx context.Context, gitRemote string, argRevList ...string) error {
	args := []string{gitRemote}
	args = append(args, argRevList...)
	_, err := lfs.Run(ctx, "fetch", args...)
	return err
}

// Push calls `git lfs push <gitRemote> <args>...`.
func (lfs *gitLFSCmd) Push(ctx context.Context, gitRemote string, args ...string) error {
	a := []string{gitRemote}
	a = append(a, args...)
	_, err := lfs.Run(ctx, "push", a...)
	return err
}

// LSFiles calls `git lfs ls-files <args>...`
//
// i.e. lists paths of git lfs files found in the tree at a provided ref.
// If two refs are given, a list of files modified between the two are shown.
func (lfs *gitLFSCmd) LSFiles(ctx context.Context, args ...string) ([]string, error) {
	return lfs.Run(ctx, "ls-files", args...)
}

// Track calls `git lfs track <pattern>`.
func (lfs *gitLFSCmd) Track(ctx context.Context, pattern string) error {
	_, err := lfs.Run(ctx, "track", fmt.Sprintf(`"%s"`, pattern))
	return err
}

// Version calls `git lfs version`.
func (lfs *gitLFSCmd) Version(ctx context.Context) (string, error) {
	out, err := lfs.Run(ctx, "version")
	if err != nil {
		return "", err
	}
	return out[0], err
}
