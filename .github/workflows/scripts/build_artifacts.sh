#!/bin/bash
# Builds Galadriel artifacts for Linux for all supported architectures.
# Usage: build_artifacts.sh

set -e

supported_architectures=(amd64 arm64)

function get_version_tag() {
  # Strip off the leading "v" from the release tag. Release artifacts are
  # named just with the version number (e.g. v0.9.3 tag produces
  # galadriel-0.9.3-linux-x64.tar.gz).
  if [[ "$GITHUB_REF" =~ ^refs/tags/v[0-9.]+$ ]]; then
    echo "${GITHUB_REF##refs/tags/v}"
  else
    echo "Error: No valid version tag found. Aborting."
    exit 1
  fi
}

function create_tarball() {
  local architecture=$1
  local version_tag=$2
  local tarball
  local staging_dir

  tarball="galadriel-${version_tag}-linux-${architecture}-glibc.tar.gz"
  staging_dir="galadriel-${version_tag}"

  mkdir "${staging_dir}"
  cp -r bin conf LICENSE "${staging_dir}"

  # Create a tarball with the binaries, license, and conf files
  tar -czvf "$tarball" "${staging_dir}"

  # Generate a SHA-256 checksum for the tarball
  sha256sum "$tarball" >"$tarball.sha256sum.txt"

  rm -rf "${staging_dir}"

  echo "Tarball successfully created for architecture: ${architecture}"
}

function build_and_package_artifact() {
  local architecture=$1
  local version_tag=$2

  if GOARCH=$architecture make build; then
    create_tarball "$architecture" "$version_tag"
  else
    echo "Error encountered while building artifact for architecture: ${architecture}"
    exit 1
  fi
}

version_tag=$(get_version_tag)
export version_tag

for architecture in "${supported_architectures[@]}"; do
  build_and_package_artifact "$architecture" "$version_tag"
done

echo "Build completed successfully for all architectures"
