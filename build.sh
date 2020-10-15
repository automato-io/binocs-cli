#!/usr/bin/env bash

 set -eux

 go build -o binocs .

 EXT=""
 if [ $GOOS == "windows" ]; then
     EXT=".exe"
 fi

 echo "binocs${EXT}"
