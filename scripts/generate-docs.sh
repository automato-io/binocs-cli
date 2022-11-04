#!/usr/bin/env bash

set -e

# set version in 0.7.4 format
BINOCS_VERSION="$1"

pushd ..
go run main.go docgen
mkdir ~/Code/automato/binocs-website/resources/docs/v${BINOCS_VERSION}/
cp -a docs/* ~/Code/automato/binocs-website/resources/docs/v${BINOCS_VERSION}/
popd