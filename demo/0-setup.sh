#!/bin/bash
set -e

SPIRE_VERSION=${SPIRE_VERSION:=v1.4.4}

cwd=$(pwd)
temp_clone=/tmp/spire-${SPIRE_VERSION}
mkdir -p ./bin

cleanup() {
    rm -rf ${temp_clone}
    exit
}
trap cleanup EXIT

# Build SPIRE
git clone https://github.com/spiffe/spire.git --single-branch --branch ${SPIRE_VERSION} -c advice.detachedHead=false ${temp_clone}
(cd ${temp_clone}; \
    make build; \
    cp bin/spire-server bin/spire-agent ${cwd}/bin/)

# Build Galadriel
(cd ../; \
    make build; \
    cp ./bin/galadriel-server ./bin/galadriel-harvester ${cwd}/bin/)

# Build greeter demo
(cd greeter; \
    CGO_ENABLED=0 go build -o ../bin/greeter-server ./cmd/greeter-server/main.go;
    CGO_ENABLED=0 go build -o ../bin/greeter-client ./cmd/greeter-client/main.go)

