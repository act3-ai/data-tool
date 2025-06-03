#!/usr/bin/env bash

help() {
    cat <<EOF

Name:
    release.sh - Run a release process in stages.

Usage:
    release.sh COMMAND [-f | --force] [-i | --interactive] [-s | --silent] [-h | --help]

Commands:
    prepare - prepare a release locally by running linters, tests, and producing the changelog, notes, assets, etc.

    approve - commit and tag your approved release.

    publish - push tag and publish the release to GitHab by uploading assets, images, helm chart, etc.

Options:
    -h, --help
        Prints usage and other helpful information.

    -i, --interactive
        Run the release process interactively, prompting for approval to continue the release process for each stage: prepare, approve, and publish. By default it begins with the prepare stage, otherwise it "resumes" the process at a specified stage.

    -s, --silent
        Run dagger silently, e.g. 'dagger --silent'.

    -f, --force
        Skip git status checks, e.g. uncommitted changes. Only recommended for development.

Required Environment Variables:
    - GITHUB_API_TOKEN     - repo access
    - GITHUB_REG_TOKEN     - write:packages access
    - GITHUB_REG_USER      - username of GITHUB_REG_TOKEN owner
    - RELEASE_AUTHOR       - username of release author, for homebrew tap
    - RELEASE_AUTHOR_EMAIL - email of release author, for homebrew tap

Dependencies:
    - dagger
    - make
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
silent=false      # silence dagger

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
# we risk unexpected behavior, e.g. 'release.sh -f' implies prepare.
if [ "$interactive" = "true" ] && [ -z "$cmd" ]; then
    cmd="prepare"
fi

# continue requests user input until a yes or no option is provided.
# Inputs:
#   - $1 : name of next stage to continue to.
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

# prepare runs linters and unit tests, bumps the version, and generates the changelog.
# runs 'approve' if interactive mode is enabled.
prepare() {
    echo "Running prepare stage..."

    dagger -s="$silent" call release check

    # git fetch --tags

    # dagger -s="$silent" call release prepare export --path="."

    echo -e "Successfully ran prepare stage.\n"
    if [ "$interactive" = "true" ]; then
        if [ "$(prompt_continue "approve")" = true ]; then
            approve
        fi
    else
        version=v$(cat VERSION)
        echo "Please review the local changes, especially releases/$version.md"
    fi
}

# approve commits changes and adds a release tag locally.
# runs 'publish' if interactive mode is enabled.
approve() {
    echo "Running approve stage..."

    # version=v$(cat VERSION)
    # notesPath="releases/$version.md"
    # # release material
    # git add VERSION CHANGELOG.md "$notesPath"
    # # documentation changes (from make gendoc, apidoc, swagger)
    # git add \*.md
    # # signed commit
    # git commit -S -m "chore(release): prepare for $version"
    # # annotated and signed tag
    # git tag -s -a -m "Official release $version" "$version"

    echo -e "Successfully ran approve stage.\n"
    if [ "$interactive" = "true" ]; then
        if [ "$(prompt_continue "publish")" = true ]; then
            publish
        fi
    fi
}

# publish pushing the release tag, uploads release assets, and publishes images.
publish() {
    echo "Running publish stage..."

    # push this branch and the associated tags
    # git push --follow-tags

    # version=$(cat VERSION)
    # registry="ghcr.io"
    # registryRepo=$registry/act3-ai/data-tool
    # imageRepoRef="${registryRepo}:${fullVersion}"
    # echo "$imageRepoRef" > artifacts.txt
    
    # # create release, upload artifacts
    # dagger -s="$silent" call \
    #     with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN \
    #     release \
    #     publish --token=env:GITHUB_API_TOKEN --ssh-private-key=env:SSH_PRIVATE_KEY --author=env:RELEASE_AUTHOR --email=env:RELEASE_AUTHOR_EMAIL

    # dagger -s="$silent" call with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN scan --sources artifacts.txt

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