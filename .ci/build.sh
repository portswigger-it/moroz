#!/usr/bin/env bash

set -eux -o pipefail

cd cmd/moroz

export GOOS="linux"

for GOARCH in "amd64" "arm64"
do
    export GOARCH="${GOARCH}"
    mkdir -p ../../build/${GOOS}/${GOARCH}
    go build -o ../../build/${GOOS}/${GOARCH}/moroz .
done