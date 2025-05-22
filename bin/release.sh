#!/usr/bin/env bash

set -euo pipefail

help() {
    cat <<EOF
Try using one of the following commands:
prepare - prepare a release locally by producing the changelog, notes, assets, etc.
approve - commit, tag, and push your approved release.
publish - publish the release to GitLab by uploading assets, images, helm chart, etc.

Dependencies: dagger, oras, make, and git.

Required Environment Variables:
Without Defaults:
    - GITHUB_API_TOKEN     - repo access
    - GITHUB_REG_TOKEN     - write:packages access
    - GITHUB_REG_USER      - username of GITHUB_REG_TOKEN owner
    - RELEASE_AUTHOR       - username of release author, for homebrew tap
    - RELEASE_AUTHOR_EMAIL - email of release author, for homebrew tap

EOF
    exit 1
}

if [ "$#" != "1" ]; then
    help
fi

set -x

registry="ghcr.io"
registryRepo=$registry/act3-ai/data-tool

case $1 in
prepare)
    dagger call release check

    dagger call release prepare export --path="."

    version=v$(cat VERSION)
    echo "Please review the local changes, especially releases/$version.md"
    ;;

approve)
    version=v$(cat VERSION)
    notesPath="releases/$version.md"
    # release material
    git add VERSION CHANGELOG.md "$notesPath"
    # documentation changes (from make gendoc, apidoc, swagger)
    git add \*.md # updated
    # signed commit
    git commit -S -m "chore(release): prepare for $version"
    # annotated and signed tag
    git tag -s -a -m "Official release $version" "$version"
    ;;
publish)
    # push this branch and the associated tags
    git push --follow-tags

    version=$(cat VERSION)
    imageRepoRef="${registryRepo}:${fullVersion}"

    # TODO: This is the usage without wrapping the gorelease module with our own module, resulting in repetitive use of env var names
    # dagger -m goreleaser call \
    #     with-secret-variable --name="GITHUB_TOKEN" --secret=env:GITHUB_TOKEN \
    #     with-secret-variable --name="SSH_PRIVATE_KEY" --secret=env:SSH_PRIVATE_KEY \
    #     with-env-variable --name="RELEASE_AUTHOR" --value="$RELEASE_AUTHOR" \
    #     with-env-variable --name="RELEASE_AUTHOR_EMAIL" --value="$RELEASE_AUTHOR_EMAIL" \
    #     with-env-variable --name="RELEASE_LATEST" --value="$RELEASE_LATEST" \
    #     release \
    #     with-fail-fast \
    #     with-notes --notes="releases/v${version}.md" \
    #     run
    
    # publish release
    dagger call \
        with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN \
        release \
        publish --token=env:GITHUB_API_TOKEN --ssh-private-key=env:SSH_PRIVATE_KEY --author=env:RELEASE_AUTHOR --email=env:RELEASE_AUTHOR_EMAIL

    # TODO: How do we want to handle extra tags? An oras container would be cleaner, but a bit unnecessary.
    # shellcheck disable=SC2046
    # oras tag "$(oras discover "$imageRepoRef" | head -n 1)" $(resolveExtraTags "$registryRepo" "$fullVersion")
    extraTags=$(dagger -m release call --src="." with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN extra-tags --ref="$imageRepoRef" --version="v${version}")
    oras tag "${imageRepoRef}:v${version}" $extraTags

    # scan images with ace-dt
    echo "$imageRepoRef" > artifacts.txt
    dagger call with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN scan --sources artifacts.txt
    ;;

*)
    help
    ;;
esac