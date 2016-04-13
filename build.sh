#!/bin/bash

set -e

OS="darwin linux windows"
ARCH="amd64"

echo "Getting build dependencies"
go get -t github.com/skytap/docker-machine-driver-skytap/docker/driver

for GOOS in $OS; do
    for GOARCH in $ARCH; do
        arch="$GOOS-$GOARCH"
        binary="bin/docker-machine-driver-skytap.$arch"
        echo "Building $binary"
        GOOS=$GOOS GOARCH=$GOARCH go build -o $binary github.com/skytap/docker-machine-driver-skytap/docker/driver/cmd/
    done
done
