#!/usr/bin/env bash

ver="$1"

actual=.gorelease-actual.txt
allowed=.gorelease-allowed.txt

echo "Running gorelease"
if gorelease -base="$(git describe --abbrev=0 --tags)" -version="v$ver" >$actual; then
    cat $actual
else
    cat $actual
    echo "To ignore these changes run 'cp $actual $allowed && git add $allowed' then commit and push."
    if [ -f "$allowed" ]; then
        echo "Comparing against $allowed"
        diff $actual $allowed || exit 2
    else
        exit 1
    fi
fi
