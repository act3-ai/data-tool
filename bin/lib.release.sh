#shellcheck shell=bash

# Bash library for release process helper functions
# The release script that is sourcing this file may override/edit any of the functions/variables to meet their specific needs.

set -euo pipefail

function help() {
    cat <<EOF

Name:
    release.sh - Run a release process in stages.

Usage:
    release.sh COMMAND [-f | --force] [-i | --interactive] [-s | --silent]  [--version VERSION] [-h | --help]

Commands:
    prepare - prepare a release locally by running linters, tests, and producing the changelog, notes, assets, etc.

    approve - commit and tag your approved release.

    publish - push tag and publish the release to a remote by uploading assets, images, helm chart, etc.

Options:
    -h, --help
        Prints usage and other helpful information.

    -i, --interactive
        Run the release process interactively, prompting for approval to continue for each stage: prepare, approve, and publish. By default it begins with the prepare stage, otherwise it "resumes" the process at a specified stage. Alternatively, set \$INTERACTIVE.

    -s, --silent
        Run dagger silently, e.g. 'dagger --silent'. Alternatively, set \$SILENT.

    -f, --force
       Skip git status checks, e.g. uncommitted changes, in all stages and linters in prepare. Alternatively, set \$FORCE.

    --version VERSION
        Run the release process for a specific semver version, ignoring git-cliff's configured bumping strategy. Alternatively, set \$VERSION.

Dependencies:
    - dagger
    - git

${extra_help:-}
EOF
    exit 1
}

# Defaults. Overriden by flag equivalents, when applicable.
cmd=""
interactive="${INTERACTIVE:-false}" # interactive mode
silent="${SILENT:-false}"           # silence dagger (dagger --silent)
explicit_version="${VERSION:-""}"   # release for a specific version

function parse_args() {
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

    if [ -z "$cmd" ]; then
        help
    fi
}

# determine if we should continue or exit
function continueQ() {
    if [ "$interactive" = "true" ]; then
        read -r -p "Continue to $cmd stage? [y/N] " response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            return
        fi
        exit 2
    else
        cmd=""
    fi
}

# Run the release script.
function main() {
    parse_args "$@"
    while [[ -n "$cmd" ]]; do
        case $cmd in
        "prepare")
            echo "Running prepare stage..."
            prepare
            echo "Successfully ran prepare stage."
            cmd="approve"
            continueQ
            ;;
        "approve")
            echo "Running approve stage..."
            approve
            echo "Successfully ran approve stage."
            cmd="publish"
            continueQ
            ;;
        "publish")
            echo "Running publish stage..."
            publish
            echo "Successfully ran publish stage."
            cmd=""
            ;;
        *)
            help
            ;;
        esac

    done
    echo "Release process complete."
}

# ensure provided envs are set
function required_envs() {
    fail=0
    for env_name in "$@"; do
        set +u
        value="${!env_name}"
        set -u
        if [[ -z ${value} ]]; then
            echo "Env var $env_name is not set!" >&2
            fail=1
        fi
    done
    if [[ "$fail" == 1 ]]; then
        exit 1
    fi
}

# prepare runs linters and unit tests, bumps the version, and generates the changelog.
function prepare() {
    # use given version if provided
    local args=()
    if [[ -n "$explicit_version" ]]; then
        args+=("--version" "$explicit_version")
    fi

    dagger_prepare "${args[@]}"

    local version
    version=$(<VERSION) # use file as source of truth

    echo "Please review the local changes, especially releases/v$version.md"
}

# approve commits changes and adds a release tag locally.
function approve() {
    # Head off a race-condition if someone else has already released.
    git fetch --tags

    local version
    version=$(<VERSION)

    changed_files+=(
        VERSION
        CHANGELOG.md
        "releases/v${version}.md"
    )

    # stage release material
    git add "${changed_files[@]}"

    # signed commit
    git commit -S -m "chore(release): prepare release with version $version"
    # annotated and signed tag
    git tag -s -a -m "Official release of version $version" "v$version"
}

# publish pushes the release tag, uploads release assets, and publishes images.
function publish() {
    # Head off a race-condition if someone else has already released.
    git fetch --tags

    # push this branch and the associated tags
    git push --follow-tags

    dagger_publish
}

dagger_args=(--quiet --silent="$silent" --auto-apply)
if [ "${DEBUG-}" == "1" ]; then
    dagger_args=(--interactive)
    set -x
fi

##############################################################################
# User may override anything in this library but the below are the most likely

# command to run the prepare phase
function dagger_prepare() {
    dagger "${dagger_args[@]}" call prepare-release
    # TODO once checks are callable from dagger functions we can remove this
    dagger "${dagger_args[@]}" check
}

# paths to "git add"
changed_files=()

# command to run the publish phase
function dagger_publish() {
    dagger "${dagger_args[@]}" call publish-release
}
