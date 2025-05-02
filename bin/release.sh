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
    - GITHUB_API_TOKEN - repo access
    - GITHUB_REG_TOKEN - write:packages access
    - GITHUB_REG_USER  - username of GITHUB_REG_TOKEN owner

EOF
    exit 1
}

if [ "$#" != "1" ]; then
    help
fi

set -x

registry="ghcr.io"
registryRepo=$registry/act3-ai/data-tool

# Extract the major version of a release.
parseMajor() {
    echo -n "$(echo -n "$1" | sed 's/v//' | cut -d "." -f 1)"
}

# Extract the minor version of a release.
parseMinor() {
    echo -n "$(echo -n "$1" | sed 's/v//' | cut -d "." -f 2)"
}

# Extract the patch version of a release.
parsePatch() {
    echo -n "$(echo -n "$1" | sed 's/v//' | cut -d "." -f 3)"
}

# Determines if the target version is the latest patch release of all releases
# with the same major and minor version.
isLatestPatch() {
    allTags="$1"
    targetMajor="$2"
    targetMinor="$3"
    targetPatch="$4"

    sameMajorMinors=$(echo "$allTags" | grep -P "^v$targetMajor\.$targetMinor\.\d+$")

    result="true"
    for v in $sameMajorMinors
    do
        if [ "$(parsePatch "$v")" -gt "$targetPatch" ]; then
            result="false"
            break
        fi
    done

    echo -n "$result"
}

# Determines if the target version is the latest minor release of all releases
# with the same major version.
isLatestMinor() {
    allTags="$1"
    targetMajor="$2"
    targetMinor="$3"

    sameMajors=$(echo "$allTags" | grep -P "^v$targetMajor\.\d+\.\d+$")

    result="true"
    for v in $sameMajors
    do
        if [ "$(parseMinor "$v")" -gt "$targetMinor" ]; then
            result="false"
            break
        fi
    done

    echo -n "$result"
}

# Determines if the target version is the latest major release.
isLatestMajor() {
    allTags="$1"
    targetMajor="$2"

    allFullTags=$(echo "$allTags" | grep -P "^v\d+\.\d+\.\d+$")

    result="true"
    for v in $allFullTags
    do
        if [ "$(parseMajor "$v")" -gt "$targetMajor" ]; then
            result="false"
            break
        fi
    done

    echo -n "$result"
}

# Determine extra tags based on existing release tags, e.g. should release v1.2.3
# also tag images for v1.2, v1, and latest. It does not check if a tag already
# exists. Only considers tags of the form '^v\d+\.\d+\.\d+$', e.g. beta releases
# are excluded.
# Input: OCI repository reference, without tag.
# Output: space separated list of tags, as a string.
resolveExtraTags() {
    ref="$1"
    targetVersion="$2"

    allTags=$(oras repo tags --exclude-digest-tags "$ref" | grep -P "^v\d+\.\d+\.\d+$" | sort -Vr)

    targetMajor=$(parseMajor "$targetVersion")
    targetMinor=$(parseMinor "$targetVersion")
    targetPatch=$(parsePatch "$targetVersion")

    latestPatch=$(isLatestPatch "$allTags" "$targetMajor" "$targetMinor" "$targetPatch")
    latestMinor=$(isLatestMinor "$allTags" "$targetMajor" "$targetMinor")
    latestMajor=$(isLatestMajor "$allTags" "$targetMajor")

    extraTags=""

    # if latest patch (for the same major.minor releases), add "vX.X" tag
    if [ "$latestPatch" = "true"  ]; then
        extraTags="v${targetMajor}.${targetMinor}"
        # if also latest minor (for the same major releases), add "vX" tag
        if [ "$latestMinor" = "true" ]; then
            extraTags="$extraTags v${targetMajor}"
            # if also latest major add "latest" tag
            if [ "$latestMajor" = "true" ]; then
                extraTags="$extraTags latest"
            fi
        fi
    fi

    echo -n "$extraTags"
}

case $1 in
prepare)
    if [[ $(git diff --stat) != '' ]]; then
        echo 'Git repo is dirty, aborting'
        exit 2
    fi

    # TODO: Check for required env vars
    # TODO: Step 1 Check - consistency and testing

    # auto-gen kube api
    # TODO: test not do
    dagger call \
        generate \
        export --path=./pkg/apis/config.dt.act3-ace.io

    # generate API docs
    dagger call \
        apidocs \
        export --path=./docs/apis/config.dt.act3-ace.io

    # generate CLI docs
    dagger call \
        clidocs \
        export --path=./docs/cli

    # run all linters
    dagger call lint all

    # run unit, functional, and integration tests
    dagger call \
        with-registry-auth --address="$registry" --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN \
        test all

    # govulncheck
    dagger -m govulncheck call run-with-source --source="." stdout

    # no changes before this point
    # TODO: ensure committed and clean, no stages, no funny business
    if [[ $(git diff --stat) != '' ]]; then
        echo 'Git repo is dirty, aborting' # TODO: Note that the dirty changes are from the script
        exit 2
    fi

    # TODO: End Step 1

    # TODO: Step 2 Prepare
    # resolve target version
    vVersion=$(dagger call -m git-cliff --source="." bumped-version)
    version=$(echo -n $vVersion | cut -c 2-) # remove 'v'
    echo "Bumping ace-dt to $version" >&2


    # generate release notes, changelog, and version file
    dagger call release prepare export --path="."

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

    # TODO: Replace remaining with:
    # dagger -m go-pipeline publish --source=URL

    version=$(cat VERSION)
    fullVersion=v${version}
    platforms=linux/amd64,linux/arm64
    
    # publish release
    dagger call \
        release \
        publish --token=env:GITHUB_API_TOKEN --ssh-private-key=env:SSH_PRIVATE_KEY --author=env:RELEASE_AUTHOR --email=env:RELEASE_AUTHOR_EMAIL

    # publish image
    # Note: Changes to existing or inclusions of additional image references should be reflected in the release notes generated in ../.dagger/release.go
    imageRepoRef="${registryRepo}:${fullVersion}"
    dagger call \
        with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN \
        image-index --version="$fullVersion" --platforms="$platforms" --address "$imageRepoRef"

    # shellcheck disable=SC2046
    oras tag "$(oras discover "$imageRepoRef" | head -n 1)" $(resolveExtraTags "$registryRepo" "$fullVersion")

    # scan images with ace-dt
    echo "$imageRepoRef" > artifacts.txt
    dagger call with-registry-auth --address=$registry --username="$GITHUB_REG_USER" --secret=env:GITHUB_REG_TOKEN scan --sources artifacts.txt
    ;;

*)
    help
    ;;
esac