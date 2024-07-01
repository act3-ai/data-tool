package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
)

const (
	mainBranchName         = "main"
	modifyFileText         = "modifying..."
	mainUpdateFileName     = "updateFile.txt"
	feature1UpdateFileName = "feature1.txt"
	feature2UpdateFileName = "feature2.txt"
	testEmail              = "user@example.com"
	testUser               = "user"
)

type subCmd struct {
	cmd  string // sometimes a func
	args []string
}

// createTestRepo creates a specially crafted git repository used for testing toOCI and fromOCI.
// see ./testdata/testing.md for a visual representation of this "script".
func createTestRepo(ctx context.Context, ch *cmd.Helper) error {
	dir := ch.Dir()

	cmdList := []subCmd{
		// initialize repo
		{"init", []string{"--initial-branch", mainBranchName}},
		{"config", []string{"user.email", testEmail}},
		{"config", []string{"user.name", testUser}},

		// v1.0.0
		{"modifyFile", []string{filepath.Join(dir, mainUpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for v1.0.0"}},
		{"tag", []string{"v1.0.0", "HEAD"}},

		// v1.0.1
		{"modifyFile", []string{filepath.Join(dir, mainUpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for v1.0.1"}},
		{"tag", []string{"v1.0.1", "HEAD"}},

		// create Feature1 branch
		{"createBranch", []string{"v1.0.1", "Feature1"}},
		{"checkout", []string{"Feature1"}},

		// extend Feature1
		{"modifyFile", []string{filepath.Join(dir, feature1UpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for feature1 branch"}},
		{"modifyFile", []string{filepath.Join(dir, feature1UpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for feature1 branch"}},

		// add v1.0.2 tag to Feature1
		{"modifyFile", []string{filepath.Join(dir, feature1UpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for feature1 branch"}},
		{"tag", []string{"v1.0.2", "HEAD"}},

		// create Feature2 branch
		{"createBranch", []string{"v1.0.1", "Feature2"}},
		{"checkout", []string{"Feature2"}},

		// extend Feature2
		{"modifyFile", []string{filepath.Join(dir, feature2UpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for feature2 branch"}},
		{"modifyFile", []string{filepath.Join(dir, feature2UpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for featur2 branch"}},

		// add v1.0.3 tag to Feature2
		{"modifyFile", []string{filepath.Join(dir, feature2UpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for feature2 branch"}},
		{"tag", []string{"v1.0.3", "HEAD"}},

		// merge Feature1 into main
		{"checkout", []string{mainBranchName}},
		{"merge", []string{"Feature1"}},

		// extend main
		{"modifyFile", []string{filepath.Join(dir, mainUpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for v1.2.0"}},
		{"modifyFile", []string{filepath.Join(dir, mainUpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for v1.2.0"}},

		// add v1.2.0 tag to main
		{"tag", []string{"v1.2.0", "HEAD"}},

		// extend main
		{"modifyFile", []string{filepath.Join(dir, mainUpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for v1.2.0"}},
		{"modifyFile", []string{filepath.Join(dir, mainUpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for v1.2.0"}},
	}

	return runCmdList(ctx, ch, cmdList)
}

// updateTestRepo updates a specially crafted git repository made with createTestRepo used for testing toOCI and fromOCI.
// see ./testdata/testing.md for a visual representation of this "script".
func updateTestRepo(ctx context.Context, ch *cmd.Helper) error {
	cmdList := []subCmd{
		// update head of Feature2
		{"checkout", []string{"Feature2"}},
		{"modifyFile", []string{filepath.Join(ch.Dir(), feature2UpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for updating Feature 2 branch head"}},

		// update v1.2.0 tag
		{"checkout", []string{mainBranchName}},
		{"update-ref", []string{"refs/tags/v1.2.0", "HEAD"}},
	}

	return runCmdList(ctx, ch, cmdList)
}

// createLFSRepo creates a git repository used for git LFS testing.
//
// It installs git-lfs to a bare repository during initialization. After adding
// an lfs file on main, two branches are created with their own lfs files.
func createLFSRepo(ctx context.Context, ch *cmd.Helper) error {
	dir := ch.Dir()
	cmdList := []subCmd{
		// initialize repo
		{"init", []string{"--initial-branch", mainBranchName}},
		{"config", []string{"user.email", testEmail}},
		{"config", []string{"user.name", testUser}},
		{"install", []string{}},
		// {"checkout", []string{mainBranchName}},
		{"track", []string{"*.txt"}}, // TODO: Don't risk user filtering these out
		{"addAttributes", []string{}},

		// Add lfs file on main
		{"modifyFile", []string{filepath.Join(dir, mainUpdateFileName), modifyFileText}},
		{"add", []string{"--all"}},
		{"commit", []string{"committing..."}},

		// add lfs file on a branch
		{"createBranch", []string{mainBranchName, "Feature1"}},
		{"checkout", []string{"Feature1"}},
		{"modifyFile", []string{filepath.Join(dir, feature1UpdateFileName), modifyFileText + " add uniqueness for 1"}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for feature1 branch"}},

		// add lfs file on another branch
		{"createBranch", []string{mainBranchName, "Feature2"}},
		{"checkout", []string{"Feature2"}},
		{"modifyFile", []string{filepath.Join(dir, feature2UpdateFileName), modifyFileText + " add uniqueness for 2"}},
		{"add", []string{"--all"}},
		{"commit", []string{"commit for feature2 branch"}},
	}

	return runCmdList(ctx, ch, cmdList)
}

// runAction runs a specific git command, git lfs command, or helper function.
func runAction(ctx context.Context, ch *cmd.Helper, action subCmd) error {
	switch action.cmd {
	case "init":
		if err := ch.Init(ctx, action.args...); err != nil {
			return err
		}
	case "config":
		if err := ch.Config(ctx, action.args[0], action.args[1]); err != nil {
			return err
		}
	case "modifyFile":
		if err := modifyFile(action.args[0], action.args[1]); err != nil {
			return err
		}
	case "add":
		if err := stage(ctx, ch, "--all"); err != nil {
			return err
		}
	case "commit":
		if err := commit(ctx, ch, action.args[0]); err != nil {
			return err
		}
	case "tag":
		if err := tag(ctx, ch, action.args[0], action.args[1]); err != nil {
			return err
		}
	case "checkout":
		if err := checkout(ctx, ch, action.args[0]); err != nil {
			return err
		}
	case "merge":
		if err := merge(ctx, ch, action.args[0]); err != nil {
			return err
		}
	case "createBranch":
		if err := createBranch(ctx, ch, action.args[0], action.args[1]); err != nil {
			return err
		}
	case "update-ref":
		if err := ch.UpdateRef(ctx, action.args[0], action.args[1]); err != nil {
			return err
		}
	case "addAttributes":
		if err := addAttributes(ctx, ch); err != nil {
			return err
		}
	case "install":
		if err := install(ctx, ch); err != nil {
			return err
		}
	case "track":
		if err := track(ctx, ch, action.args[0]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unrecognized action %s, args = %s", action.cmd, action.args)
	}

	return nil
}

// runCmdList runs a list of commands, which is essentially a "script".
func runCmdList(ctx context.Context, ch *cmd.Helper, list []subCmd) error {
	for i, action := range list {
		if err := runAction(ctx, ch, action); err != nil {
			return fmt.Errorf("action at cmdList[%d], cmd = %s: %w", i, action, err)
		}
	}
	return nil
}

// modifyFile modifies, creating if necessary, the file at path by appending the text
func modifyFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening file for modification: %w", err)
	}
	defer f.Close()

	_, err = f.Write([]byte(fmt.Sprintf("%s\n", text)))
	if err != nil {
		return fmt.Errorf("modifying file: %w", err)
	}

	return nil
}

// createBranch makes a new branch off of the commit idenfified by fromRef, using the provided newBranch name.
func createBranch(ctx context.Context, ch *cmd.Helper, fromRef, newBranch string) error {
	out, err := ch.ShowRefs(ctx, fromRef)
	if err != nil {
		return fmt.Errorf("getting commit from %s to create branch %s: %w", fromRef, newBranch, err)
	}

	split := strings.Split(out[0], " ")
	commit := split[0]

	err = branch(ctx, ch, newBranch, commit)
	if err != nil {
		return fmt.Errorf("creating branch %s: %w", newBranch, err)
	}

	return nil
}

// gitListHeads calls `git bundle list-heads <bundlePath>`
//
// i.e. returns a list of all tag and head references in the bundle.
func listHeads(ctx context.Context, ch *cmd.Helper, bundlePath string) ([]string, error) {
	return ch.Git.Run(ctx, "bundle", "list-heads", bundlePath)
}

func branch(ctx context.Context, ch *cmd.Helper, name, commit string) error {
	_, err := ch.Git.Run(ctx, "branch", name, commit)
	return err
}

func checkout(ctx context.Context, ch *cmd.Helper, branch string) error {
	_, err := ch.Git.Run(ctx, "checkout", branch)
	return err
}

func commit(ctx context.Context, ch *cmd.Helper, message string) error {
	_, err := ch.Git.Run(ctx, "commit", "-m", message)
	return err
}

func stage(ctx context.Context, ch *cmd.Helper, args ...string) error {
	_, err := ch.Git.Run(ctx, "add", args...)
	return err
}

func tag(ctx context.Context, ch *cmd.Helper, name, commit string) error {
	_, err := ch.Git.Run(ctx, "tag", name, commit)
	return err
}

func merge(ctx context.Context, ch *cmd.Helper, mergeTarget string) error {
	_, err := ch.Git.Run(ctx, "merge", mergeTarget, "-m", "merging")
	return err
}

// RevList calls `git rev-list <argRevList>...`
func revList(ctx context.Context, ch *cmd.Helper, args ...string) ([]string, error) {
	return ch.Git.Run(ctx, "rev-list", args...)
}

// VerifyPack calls `git verify-pack *.idx -s` for all *.idx files in packDir,
// returning the sum of objects in the packfiles.
func verifyPack(ctx context.Context, ch *cmd.Helper, idxPath string) (int, error) {
	out, err := ch.Git.Run(ctx, "verify-pack", idxPath, "-s")
	if err != nil {
		return -1, err
	}

	re := regexp.MustCompile("[0-9]+")
	nums := re.FindAllString(out[0], -1) // expecting one line
	num, err := strconv.Atoi(nums[0])    // line should only have one int
	if err != nil {
		return -1, fmt.Errorf("converting string to int: %w", err)
	}
	return num, nil
}

// install calls `git lfs install --local`.
func install(ctx context.Context, ch *cmd.Helper) error {
	_, err := ch.LFS.Run(ctx, "install", "--local")
	return err
}

// track calls `git lfs track <pattern>`.
func track(ctx context.Context, ch *cmd.Helper, pattern string) error {
	_, err := ch.LFS.Run(ctx, "track", fmt.Sprintf(`"%s"`, pattern))
	return err
}

// addAttributes calls `git add .gitattribues`
//
// Used for supporting git-lfs tracked files.
func addAttributes(ctx context.Context, ch *cmd.Helper) error {
	_, err := ch.Git.Run(ctx, "add", ".gitattributes")
	return err
}
