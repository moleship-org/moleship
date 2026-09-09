#!/usr/bin/env bash
# Build and push the moleship image to its container registry.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

log_info "Publishing moleship image..."

IMAGE_TAG="$(default_image_tag)"
VERSION=""

log_info "Building ${IMAGE_TAG}..."
./build.sh

require_cmd podman

VERSION_TAG=""
if [[ -n "${VERSION}" ]]; then
  VERSION_TAG="${IMAGE_TAG%:*}:${VERSION}"
  log "Tagging ${VERSION_TAG}..."
  podman tag "${IMAGE_TAG}" "${VERSION_TAG}"
fi

log_info "Pushing ${IMAGE_TAG}..."
podman push "${IMAGE_TAG}"

if [[ -n "${VERSION_TAG}" ]]; then
  log "Pushing ${VERSION_TAG}..."
  podman push "${VERSION_TAG}"
fi

log_success "Published:"
log "  ${IMAGE_TAG}"
if [[ -n "${VERSION_TAG}" ]]; then
  log "  ${VERSION_TAG}"
fi
