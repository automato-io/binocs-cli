#!/usr/bin/env bash

set -eux

go build $1

EXT=""
if [ $GOOS == "windows" ]; then
    EXT=".exe"
fi

echo "binocs${EXT}"