#!/usr/bin/env bash

set -e

# set version in 0.7.4 format
BINOCS_VERSION="$1"

pushd ..
git add cmd/root.go
git commit -m 'bump version to ${BINOCS_VERSION}'
git tag -a v${BINOCS_VERSION} -m 'release v${BINOCS_VERSION}'
git push origin master
git push origin v${BINOCS_VERSION}
popd
