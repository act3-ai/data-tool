package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// Helper assists in running git and git-lfs commands. Its methods often
// combine and parse git or git-lfs commands to determine information
// about a repository's state.
type Helper struct {
	Options
	Git
	LFS

	dir string
}

// NewHelper returns a cmdHelper object used for running git and git-lfs commands.
// It validates the compatibility of git and git-lfs. Displays a warning if git-lfs
// is not installed.
func NewHelper(ctx context.Context, gitDir string, opts *Options) (*Helper, error) {
	ch := &Helper{
		Options: *opts,
		dir:     gitDir,
	}
	ch.Git = &gitCmd{
		dir:        gitDir,
		altGitExec: opts.AltGitExec,
	}

	if opts.LFSOptions == nil { // prevent panic if no LFS opts are specified
		ch.LFSOptions = &LFSOptions{
			WithLFS: true, // default behavior
		}
	}
	ch.LFS = &gitLFSCmd{
		dir:           gitDir,
		altGitLFSExec: ch.AltLFSExec,
	}

	return ch, nil
}

// ValidateVersions checks if installed git and git-lfs versions meet minimum requirements.
func (c *Helper) ValidateVersions(ctx context.Context) error {
	u := ui.FromContextOrNoop(ctx)

	_, err := CheckGitVersion(ctx, c.AltGitExec)
	if err != nil {
		return fmt.Errorf("validating git version: %w", err)
	}

	if c.WithLFS {
		version, err := CheckGitLFSVersion(ctx, c.AltLFSExec)
		switch {
		case errors.Is(err, ErrLFSCmdNotFound):
			u.Infof("Warning: git-lfs is not installed. Continuing without syncing git-lfs files.")
			c.WithLFS = false // override LFS setting
			return nil        // recover
		case err != nil:
			u.Infof("Warning: git-lfs version is incompatible. Found version %s, minimum is %s. Continuing without syncing git-lfs files.", version, minGitLFSVersion)
			c.WithLFS = false // override LFS setting
			return nil        // recover
		}
	} else {
		u.Infof("Warning: Overriding git-lfs syncing may prevent pushing to the destination with 'ace-dt git from-oci'. Continuing without syncing git-lfs files.")
	}
	return nil
}

// Dir returns the path of the git directory.
func (c *Helper) Dir() string {
	return c.dir
}

// InitializeRepo initializes the temporary directory as an empty bare git repository. This repository
// functions as an intermediate repo of which changes are collected/applied and then handled accordingly.
func (c *Helper) InitializeRepo(ctx context.Context) error {
	if err := c.Init(ctx, "--bare"); err != nil {
		return fmt.Errorf("creating bare repository: %w", err)
	}

	return nil
}

// LocalCommitsRefs returns the local references and the commits they reference
// split into two slices, with indicies matching the pairs. If argRevList is empty
// all references will be returned.
func (c *Helper) LocalCommitsRefs(ctx context.Context, argRevList ...string) ([]string, []string, error) {
	commitsRefs, err := c.ShowRefs(ctx, argRevList...)
	if err != nil {
		return nil, nil, fmt.Errorf("resolving local commits and references: %w", err)
	}

	commits, fullRefs := parseOIDRefs(commitsRefs...)
	return commits, fullRefs, nil
}

// RemoteCommitsRefs returns the remote references and the commits they reference
// split into two slices, with indicies matching the pairs. If argRevList is empty
// all references will be returned.
func (c *Helper) RemoteCommitsRefs(ctx context.Context, remote string, argRevList ...string) ([]string, []string, error) {
	args := make([]string, 0, len(argRevList)+2)
	args = append(args, "--tags", "--heads", "--refs", remote)
	args = append(args, argRevList...)
	refsCommits, err := c.LSRemote(ctx, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving remote commits and references: %w", err)
	}

	commits, fullRefs := parseOIDRefs(refsCommits...)
	return commits, fullRefs, nil
}

// IsAncestor returns nil if the parent commit is an ancestor of the child commit, ErrNotAncestor
// if not, and the original git error if one occurs.
func (c *Helper) IsAncestor(ctx context.Context, parent, child Commit) error {
	out, err := c.Git.MergeBase(ctx, "--is-ancestor", string(parent), string(child))
	var exitErr *exec.ExitError
	switch {
	case errors.As(err, &exitErr) && exitErr.ExitCode() == 1:
		return ErrNotAncestor
	case err != nil:
		// exit code > 1
		return fmt.Errorf("running git merge-base: output: %s: %w", out, err)
	default:
		// exit code = 0
		return nil
	}
}
