#!/bin/env bash
set -e
set -x

REGISTRY=$1

VERSION=v0.25.4
for DATA in basic subparts
do
    BOTTLE=$DATA/bottle-$VERSION
    OCI=$DATA/oci-$VERSION
    rm -rf $BOTTLE
    rm -rf $OCI
    cp -r $DATA/original $BOTTLE
    go run git.act3-ace.com/ace/data/tool/cmd/ace-dt@$VERSION bottle init -d $BOTTLE
    go run git.act3-ace.com/ace/data/tool/cmd/ace-dt@$VERSION bottle push -d $BOTTLE http://"$REGISTRY"/testdata/$DATA:$VERSION --force
    go run git.act3-ace.com/ace/data/tool/cmd/ace-dt@master oci pull "$REGISTRY"/testdata/$DATA:$VERSION $OCI
done

VERSION=master
for DATA in basic subparts
do
    BOTTLE=$DATA/bottle-$VERSION
    OCI=$DATA/oci-$VERSION
    rm -rf $BOTTLE
    rm -rf $OCI
    cp -r $DATA/original $BOTTLE
    go run git.act3-ace.com/ace/data/tool/cmd/ace-dt@$VERSION bottle init -d $BOTTLE
    go run git.act3-ace.com/ace/data/tool/cmd/ace-dt@$VERSION bottle push -d $BOTTLE http://"$REGISTRY"/testdata/$DATA:$VERSION
    go run git.act3-ace.com/ace/data/tool/cmd/ace-dt@master oci pull "$REGISTRY"/testdata/$DATA:$VERSION $OCI
done