package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// Git provides access to git commands.
type Git interface {
	Init(args ...string) error
	CloneWithShared(gitRemote, reference string) error
	Config(args ...string) error
	Push(gitRemote string, refs ...string) error
	Fetch(gitRemote string, args ...string) error
	BundleCreate(destFile string, revList []string) error
	ShowRefs(refs ...string) ([]string, error)
	UpdateRef(ref string, commit string) error
	RemoteAdd(shortname, remoteTarget string) error
	RemoteRemove(shortname string) error
	LSRemote(args ...string) ([]string, error)
	MergeBase(args ...string) error
	Run(subCmd string, args ...string) ([]string, error)
}

// gitCmd contains a logger and directory of execution for a git command.
// gitCmd implements Git.
type gitCmd struct {
	logger *slog.Logger
	dir    string

	altGitExec string // optional
}

// newGitCmd returns a gitCmd using the provided logger and directory of execution.
func newGitCmd(log *slog.Logger, dir string, altGitExec string) *gitCmd {
	return &gitCmd{
		logger:     log,
		dir:        dir,
		altGitExec: altGitExec,
	}
}

func (gc *gitCmd) log(args []string, out []byte) {
	gc.logger.Info("Ran command", "command", args, "directory", gc.dir, "output", string(out)) //nolint:sloglint
}

// Run executes a git command, returning the parsed output.
func (gc *gitCmd) Run(subCmd string, args ...string) ([]string, error) {
	gitCmd := "git"
	if gc.altGitExec != "" {
		gitCmd = gc.altGitExec
	}
	cmd := exec.Command(gitCmd)

	cmd.Args = append(cmd.Args, subCmd)
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = gc.dir

	// We only want stdout for processing but we want stderr for errors
	out, err := cmd.Output()
	gc.log(cmd.Args, out) // log the raw output
	parsedOut := parseGitOutput(out)
	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			gc.log(cmd.Args, exitError.Stderr) // log the exit error
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
func (gc *gitCmd) ShowRefs(refs ...string) ([]string, error) {
	return gc.Run("show-ref", refs...)
}

// UpdateRef calls `git update-ref <ref> <commit>` within the gitCmd's  directory.
func (gc *gitCmd) UpdateRef(ref string, commit string) error {
	_, err := gc.Run("update-ref", ref, string(commit))
	return err
}

// RemoteAdd calls `git remote add <shortname> <remoteTarget>` within the gitCmd's  directory.
func (gc *gitCmd) RemoteAdd(shortname, remoteTarget string) error {
	_, err := gc.Run("remote", "add", shortname, remoteTarget)
	return err
}

// RemoteRemove calls `git remote remove <shortname>` within the gitCmd's  directory.
func (gc *gitCmd) RemoteRemove(shortname string) error {
	_, err := gc.Run("remote", "remove", shortname)
	return err
}

// LSRemote calls `git ls-remote <args>...`.
func (gc *gitCmd) LSRemote(args ...string) ([]string, error) {
	return gc.Run("ls-remote", args...)
}

// Fetch calls `git fetch <gitRemote> <args>...` within the gitCmd's  directory.
//
// i.e. fetches from a bundle with the --tags flag, expecting a remote HEAD ref to have been set.
func (gc *gitCmd) Fetch(gitRemote string, args ...string) error {
	a := []string{gitRemote}
	a = append(a, args...)
	_, err := gc.Run("fetch", a...)
	return err
}

// Push calls `git push <gitRef> --tags` within the gitCmd's  directory.
//
// i.e. pushes a local git repository to the local/remote reference, with tags.
func (gc *gitCmd) Push(gitRemote string, refs ...string) error {
	args := []string{gitRemote}
	args = append(args, refs...)
	_, err := gc.Run("push", args...)
	return err
}

// Init calls `git init` within the gitCmd's  directory.
func (gc *gitCmd) Init(args ...string) error {
	_, err := gc.Run("init", args...)
	return err
}

// CloneWithShared calls `git clone --shared --reference-if-able <reference> --bare <gitRef> <gc.dir>`.
//
// Cloning with the shared option prevents copying objects to the clone. This is a safe operation
// as long as the cache is not pRuned between cloning and managing the clone.
func (gc *gitCmd) CloneWithShared(gitRemote, reference string) error {
	_, err := gc.Run("clone", "--shared", "--reference-if-able", reference, "--bare", gitRemote, gc.dir)
	return err
}

// BundleCreate calls `git bundle create <destFile> <revList>...` within the gitCmd's  directory.
//
// i.e. creates a git bundle including all layers specified in revList, writing the bundle to the
// destination path.
func (gc *gitCmd) BundleCreate(destFile string, revList []string) error {
	args := make([]string, 0, len(revList)+2)
	args = append(args, "create", destFile)
	args = append(args, revList...)
	_, err := gc.Run("bundle", args...)

	return err
}

// MergeBase calls `git merge-base <args>...` within the gitCmd's directory.
func (gc *gitCmd) MergeBase(args ...string) error {
	out, err := gc.Run("merge-base", args...)

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
func (gc *gitCmd) Config(args ...string) error {
	_, err := gc.Run("config", args...)
	if err != nil {
		return err
	}

	return nil
}
