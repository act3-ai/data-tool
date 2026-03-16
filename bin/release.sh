#!/usr/bin/env bash

source lib.release.sh

############
# Custom code is below

extra_help="
Please see .env.example for required variables to be set in your .env file.
"

# additional files that change during the release process
changed_files+=(
    # TODO
)

# command to run the prepare phase
function dagger_prepare() {
    dagger "${dagger_args[@]}" call prepare-release
    # TODO once checks are callable from dagger functions we can remove this
    dagger "${dagger_args[@]}" check
}

main "$@"
