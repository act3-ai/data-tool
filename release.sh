#!/usr/bin/env bash

# For custom changes, see https://daggerverse.dev/mod/github.com/act3-ai/dagger/release for dagger release module usage.

# Custom Variables
version_path="VERSION"
changelog_path="CHANGELOG.md"
notes_dir="releases"

# Remote Dependencies
mod_release="release"
mod_gitcliff="git-cliff"

help() {
    cat <<EOF

Name:
    release.sh - Run a release process in stages.

Usage:
    release.sh COMMAND [-f | --force] [-i | --interactive] [-s | --silent] [--version VERSION] [-h | --help]

Commands:
    prepare - prepare a release locally by running linters, tests, and producing the changelog, notes, assets, etc.

    approve - commit and tag your approved release.

    publish - push tag and publish the release to a remote by uploading assets, images, helm chart, etc.

Options:
    -h, --help
        Prints usage and other helpful information.

    -i, --interactive
        Run the release process interactively, prompting for approval to continue for each stage: prepare, approve, and publish. By default it begins with the prepare stage, otherwise it "resumes" the process at a specified stage.

    -s, --silent
        Run dagger silently, e.g. 'dagger --silent'.

    -f, --force
        Skip git status checks, e.g. uncommitted changes. Only recommended for development.

    --version VERSION
        Run the release process for a specific semver version, ignoring git-cliff's configured bumping strategy.

Required Environment Variables:
    - GITHUB_API_TOKEN     - repo:api access
    - GITHUB_REG_TOKEN     - write:packages access
    - GITHUB_REG_USER      - username of GITHUB_REG_TOKEN owner
    - RELEASE_AUTHOR       - username of release author, for homebrew tap
    - RELEASE_AUTHOR_EMAIL - email of release author, for homebrew tap
    - RELEASE_LATEST       - tag release as latest

Dependencies:
    - dagger
    - git
EOF
    exit 1
}

# insufficient args
if [ "$#" -eq 0 ]; then
    help
fi

set -euo pipefail

# Defaults
cmd=""
force=false       # skip git status checks
interactive=false # interactive mode
silent=false      # silence dagger (dagger --silent)
explicit_version=""  # release for a specific version

# Get commands and flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    # Commands
    "prepare" | "approve" | "publish")
       cmd=$1
       shift
       ;;
    # Flags
    "-h" | "--help")
       help
       ;;
    "--version")
       shift
       explicit_version=$1
       shift
       ;;
    "-i" | "--interactive")
       interactive=true
       shift
       ;;
    "-s" | "--silent")
       silent=true
       shift
       ;;
    "-f" | "--force")
       force=true
       shift
       ;;
    *)
       echo "Unknown option: $1"
       help
       ;;
  esac
done

# Interactive mode begins with prepare by default, otherwise continue the release
# process at the specified stage. Must occur after parsing commands and flags, else
# we risk unexpected behavior, e.g. 'release.sh -f' would imply prepare.
if [ "$interactive" = "true" ] && [ -z "$cmd" ]; then
    cmd="prepare"
fi

# prompt_continue requests user input until a valid y/n option is provided.
# Inputs:
#   - $1 : name of next stage to continue to.
# disable read without -r backslash mangling for this func
# shellcheck disable=SC2162
prompt_continue() {
    read -p "Continue to $1 stage (y/n)?" choice
    case "$choice" in
    y|Y )
        echo -n "true"
    ;;
    n|N )
        echo -n "false"
        ;;
    * )
        echo "Invalid input '$choice'" >&2
        prompt_continue "$1"
        ;;
    esac
}

# check_upstream ensures remote upstream matches local commit.
# Inputs:
#  - $1 : commit, often HEAD or HEAD~1
check_upstream() {
    if [ "$force" != "true" ]; then
        echo "Comparing local $1 to remote upstream"
        git diff "@{upstream}" "$1" --stat --exit-code
    fi
}

# prepare runs linters and unit tests, bumps the version, and generates the changelog.
# runs 'approve' if interactive mode is enabled.
prepare() {
    echo "Running prepare stage..."

    old_version=v$(cat "$version_path")

    # linters and unit tests
    dagger -s="$silent" --src="." call release check

    git fetch --tags
    check_upstream "HEAD"

    # bump version, generate changelogs
    vVersion=""
    if [ "$explicit_version" != "" ]; then
        vVersion="$explicit_version"
    else
        vVersion=$(dagger -m="$mod_gitcliff" -s="$silent" --src="." call bumped-version)
    fi

       vVersion=v$(cat "$version_path") # use file as source of truth
    # verify release version with gorelease
    dagger -m="$mod_release" -s="$silent" --src="." call go verify --target-version="$vVersion" --current-version="$old_version"

    # bump version, generate changelogs
    dagger -s="$silent" --src="." call release prepare --ignore-error="$force" --version="$vVersion" export --path="."

    echo -e "Successfully ran prepare stage.\n"
    echo -e "Please review the local changes, especially releases/$vVersion.md\n"
    if [ "$interactive" = "true" ] && [ "$(prompt_continue "approve")" = "true" ]; then
            approve
    fi
}

# approve commits changes and adds a release tag locally.
# runs 'publish' if interactive mode is enabled.
approve() {
    echo "Running approve stage..."

    git fetch --tags
    check_upstream "HEAD"

    vVersion=v$(cat "$version_path")
    notesPath="${notes_dir}/${vVersion}.md"

    # stage release material
    git add "$version_path" "$changelog_path" "$notesPath"
    git add \*.md
    # signed commit
    git commit -S -m "chore(release): prepare for $vVersion"
    # annotated and signed tag
    git tag -s -a -m "Official release $vVersion" "$vVersion"

    echo -e "Successfully ran approve stage.\n"
    if [ "$interactive" = "true" ] && [ "$(prompt_continue "publish")" = "true" ]; then
            publish
    fi
}

# publish pushes the release tag, uploads release assets, and publishes images.
publish() {
    echo "Running publish stage..."

    git fetch --tags
    check_upstream "HEAD~1"

    # push this branch and the associated tags
    git push --follow-tags
    
    # build image OCI reference
    vVersion=v$(cat VERSION)
    registry="ghcr.io"
    registryRepo=$registry/act3-ai/data-tool
    imageRepoRef="${registryRepo}:${vVersion}"
    echo "$imageRepoRef" > artifacts.txt

    # create release, upload artifacts
    dagger -s="$silent" call \
        with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN \
        release \
        publish --token=env:GITHUB_API_TOKEN --ssh-private-key=env:SSH_PRIVATE_KEY --author=env:RELEASE_AUTHOR --email=env:RELEASE_AUTHOR_EMAIL
    # publish SBOM and CVE results
    dagger -s="$silent" call with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN scan --sources artifacts.txt

    echo -e "Successfully ran publish stage.\n"
    echo "Release process complete."
}

# Run the release script.
case $cmd in
"prepare")
    prepare
    ;;
"approve")
    approve
    ;;
"publish")
    publish
    ;;
*)
    help
    ;;
esac