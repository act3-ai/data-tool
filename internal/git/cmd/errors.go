package cmd

import "errors"

// ErrLFSNotEnabled indicates that no LFS files were found.
var ErrLFSNotEnabled = errors.New("repository does not contain git LFS files")

// ErrLFSCmdNotFound indicates that the git-lfs command is not installed.
var ErrLFSCmdNotFound = errors.New("git-lfs command not found")

// ErrTagUpdate is the error returned when updating an existing git tag fails because it already exists.
var ErrTagUpdate = errors.New("updates were rejected because the tag already exists in the remote")

// ErrRepoNotExistOrPermDenied indicates a remote repository may not exist or user has insufficient permissions.
var ErrRepoNotExistOrPermDenied = errors.New("repository does not exist or insufficient permissions")

// ErrEmptyBundle indicates the set of commits resolved by git-rev-list-args is empty.
var ErrEmptyBundle = errors.New("refusing to create empty bundle")

// ErrNotAncestor indicates a target commit is not an ancestor (parent) of another.
var ErrNotAncestor = errors.New("parent is not an ancestor of child")

var errInsufficientGitVersion = errors.New("detected git version does not meet minimum requirements")
