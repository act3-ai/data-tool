#!/usr/bin/env bash

# For custom changes, see https://daggerverse.dev/mod/github.com/act3-ai/dagger/release for dagger release module usage.# Custom Variables
git_remote="<no value>"


# Remote Dependencies
mod_release="github.com/act3-ai/dagger/release@release/v0.1.1"
mod_goreleaser="github.com/act3-ai/dagger/goreleaser@goreleaser/v0.1.0"


help() {
    cat <<EOF

Name:
    release.sh - Run a release process in stages.

Usage:
    release.sh COMMAND [-f | --force] [-i | --interactive] [-s | --silent] [-h | --help]

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

Required Environment Variables:
    TODO: Add as desired
    - GITHUB_API_TOKEN     - repo:api access

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
# we risk unexpected behavior, e.g. 'release.sh -f' would imply prepare.
if [ "$interactive" = "true" ] && [ -z "$cmd" ]; then
    cmd="prepare"
fi

# prompt_continue requests user input until a valid y/n option is provided.
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

# check_upstream ensures remote upstream matches local HEAD.
check_upstream() {
    git diff @{upstream} --stat --exit-code || {
        echo "Local HEAD does not match upstream"
        echo "Please review 'git diff @{upstream}' and match remote upstream or use --force"
        exit 1
    }
}

# prepare runs linters and unit tests, bumps the version, and generates the changelog.
# runs 'approve' if interactive mode is enabled.
prepare() {
    echo "Running prepare stage..."

    old_version=v$(cat "$version_file")

    # linters and unit tests
    dagger -m="$mod_release" -s="$silent" --src="." call go check

    git fetch --tags
    check_upstream
    # bump version, generate changelogs
    dagger -m="$mod_release" -s="$silent" --src="." call prepare export --path="."

    version=v$(cat "$version_file")
    # verify release version with gorelease
    dagger -m="$mod_release" -s="$silent" --src="." call go verify --target-version="$version" --current-version="$old_version"

    
    echo -e "Successfully ran prepare stage.\n"
    echo -e "Please review the local changes, especially releases/$version.md\n"
    if [ "$interactive" = "true" ] && [ "$(prompt_continue "approve")" = "true" ]; then
            approve
    fi
}

# approve commits changes and adds a release tag locally.
# runs 'publish' if interactive mode is enabled.
approve() {
    echo "Running approve stage..."

    git fetch --tags
    check_upstream

    version=v$(cat "$version_file")
    notesPath="releases/$version.md"

    # stage release material
    git add "VERSION" "CHANGELOG.md" "$notesPath"
    git add \*.md
    
    # signed commit
    git commit -S -m "chore(release): prepare for $version"
    # annotated and signed tag
    git tag -s -a -m "Official release $version" "$version"

    echo -e "Successfully ran approve stage.\n"
    if [ "$interactive" = "true" ] && [ "$(prompt_continue "publish")" = "true" ]; then
            publish
    fi
}

# publish pushes the release tag, uploads release assets, and publishes images.
publish() {
    echo "Running publish stage..."

    git fetch --tags
    check_upstream

    # push this branch and the associated tags
    git push --follow-tags

    version=v$(cat "$version_file")

    dagger -m="$mod_goreleaser" -s="$silent" --src="." call \
    with-secret-variable --name="GITHUB_API_TOKEN" --secret=env:GITHUB_API_TOKEN \
    with-env-variable --name="RELEASE_LATEST" --value="$RELEASE_LATEST" \
    release

    
    # For resolving extra image tags, see https://daggerverse.dev/mod/github.com/act3-ai/dagger/release#Release.extraTags
    # extra_tags=$(dagger -m="$mod_release" -s="$silent" --src="."  call release extra-tags --ref=<OCI_REF> --version="$version")
    # For applying extra image tags, see https://daggerverse.dev/mod/github.com/act3-ai/dagger/release#Release.addTags OR if the docker module is used, provide them directly to --tags
    
    # publish image
    # TODO:
    # - Docker dagger module - https://daggerverse.dev/mod/github.com/act3-ai/dagger/docker
    # - Native dagger containers - https://docs.dagger.io/cookbook#perform-a-multi-stage-build
    # - Or other methods

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