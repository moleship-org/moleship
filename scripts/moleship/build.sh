#!/usr/bin/env bash
# Build the moleship image using podman.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

IMAGE_TAG="$(default_image_tag)"
VERSION=""

parse_common_flags "$@"

require_cmd podman

BUILD_ARGS=()
if [[ -n "${VERSION}" ]]; then
  BUILD_ARGS+=(--build-arg "VERSION=${VERSION}")
fi

log_info "Building ${IMAGE_TAG}..."
podman build -f "${CONTAINERFILE}" -t "${IMAGE_TAG}" "${BUILD_ARGS[@]}" "${ROOT_DIR}"

log_success "Successfully built ${IMAGE_TAG}"
