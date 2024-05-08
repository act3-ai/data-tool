package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// LFS provides access to git-lfs commands.
type LFS interface {
	Version() (string, error)
	Track(pattern string) error
	Fetch(gitRemote string, argRevList ...string) error
	Push(gitRemote string, argRevList ...string) error
	LSFiles(args ...string) ([]string, error)
	Run(subCmd string, args ...string) ([]string, error)
}

// gitLFSCmd contains a logger and directory of execution for a git lfs command.
// gitLFSCmd implements LFS.
type gitLFSCmd struct {
	logger *slog.Logger
	dir    string

	altGitLFSExec string // optional
}

// newGitLFSCmd returns a gitLFSCmd using the provided logger and directory of execution.
func newGitLFSCmd(log *slog.Logger, dir, altGitLFSExec string) *gitLFSCmd {
	return &gitLFSCmd{
		logger:        log,
		dir:           dir,
		altGitLFSExec: altGitLFSExec,
	}
}

func (lfs *gitLFSCmd) log(args []string, out []byte) {
	lfs.logger.Info("Executed git-lfs Command", "command", args, "output", string(out)) //nolint:sloglint
}

// Run executes a git lfs command, returning the parsed output.
func (lfs *gitLFSCmd) Run(subCmd string, args ...string) ([]string, error) {
	var cmd *exec.Cmd
	switch {
	case lfs.altGitLFSExec != "":
		cmd = exec.Command(lfs.altGitLFSExec)
	default:
		cmd = exec.Command("git-lfs")
	}
	cmd.Args = append(cmd.Args, subCmd)
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = lfs.dir

	out, err := cmd.Output()
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
			lfs.log(cmd.Args, []byte(execErr.Error())) // log the exec error
			if strings.Contains(string(execErr.Error()), "executable file not found in $PATH") {
				return parsedOut, errors.Join(err, ErrLFSCmdNotFound)
			}
		}
		return parsedOut, fmt.Errorf("git-lfs %s produced an error: %w", cmd.Args, err)
	}

	lfs.log(cmd.Args, out) // log the raw output
	return parsedOut, nil
}

// Fetch calls `git lfs fetch <args>...`
//
// i.e. downloads git lfs objects at the provided refs. Requires a
// remote to be provided if a default is not set (same default as git fetch).
func (lfs *gitLFSCmd) Fetch(gitRemote string, argRevList ...string) error {
	args := []string{gitRemote}
	args = append(args, argRevList...)
	_, err := lfs.Run("fetch", args...)
	return err
}

// Push calls `git lfs push --all <gitRemote>`.
func (lfs *gitLFSCmd) Push(gitRemote string, argRevList ...string) error {
	args := []string{gitRemote}
	args = append(args, argRevList...)
	_, err := lfs.Run("push", args...)
	return err
}

// LSFiles calls `git lfs ls-files <args>...`
//
// i.e. lists paths of git lfs files found in the tree at a provided ref.
// If two refs are given, a list of files modified between the two are shown.
func (lfs *gitLFSCmd) LSFiles(args ...string) ([]string, error) {
	return lfs.Run("ls-files", args...)
}

// Track calls `git lfs track <pattern>`.
func (lfs *gitLFSCmd) Track(pattern string) error {
	_, err := lfs.Run("track", fmt.Sprintf(`"%s"`, pattern))
	return err
}

// Version calls `git lfs version`.
func (lfs *gitLFSCmd) Version() (string, error) {
	out, err := lfs.Run("version")
	if err != nil {
		return "", err
	}
	return out[0], err
}
