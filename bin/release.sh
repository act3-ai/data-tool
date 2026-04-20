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

main "$@"
