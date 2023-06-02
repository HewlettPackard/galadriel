#!/bin/bash
# Builds Galadriel artifacts for Linux for all supported architectures.
# Usage: build_artifacts.sh

set -e

supported_architectures=(amd64 arm64)

export version_tag=
if [[ "$GITHUB_REF" =~ ^refs/tags/v[0-9.]+$ ]]; then
  # Strip off the leading "v" from the release tag. Release artifacts are
  # named just with the version number (e.g. v0.9.3 tag produces
  # galadriel-0.9.3-linux-x64.tar.gz).
  version_tag="${GITHUB_REF##refs/tags/v}"
fi

for architecture in "${supported_architectures[@]}"; do
  # Build the server and harvester binaries for the current architecture
  if GOARCH=$architecture go build -o bin/galadriel-server ./cmd/server && \
     GOARCH=$architecture go build -o bin/galadriel-harvester ./cmd/harvester; then

    echo "Artifacts successfully built for architecture: ${architecture}"
    tarball="galadriel-${version_tag}-linux-${architecture}-glibc.tar.gz"

    # Create a tarball with the binaries, license, and conf files
    tar -czvf "$tarball" -C bin/ . -C ../ LICENSE conf/

    # Generate a SHA-256 checksum for the tarball
    sha256sum "$tarball" > "$tarball.sha256sum.txt"
  else
    echo "Error encountered while building artifact for architecture: ${architecture}"
    exit 1
  fi
done

echo "Build completed successfully for all architectures"
