#!/usr/bin/env bash
# Pull the latest moleship image and restart the running Quadlet service to
# pick it up. See scripts/setup/publish.sh to build and push a new image
# first. If the image is private, log in to its registry first (e.g.
# `podman login ghcr.io`).
#
# Usage:
#   scripts/setup/update.sh [--rootful] [--tag <image:tag>]

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

parse_common_flags "$@"
require_rootful_privileges

require_cmd podman
require_cmd systemctl

DEST_FILE="$(quadlet_dir)/${UNIT_NAME}"
if [[ ! -f "${DEST_FILE}" ]]; then
  die "No existing install found at ${DEST_FILE}. Run install.sh first."
fi

log "Pulling ${IMAGE_TAG}..."
podman pull "${IMAGE_TAG}"

log "Restarting moleship..."
systemctl_ daemon-reload
systemctl_ restart "${SERVICE_NAME}"

log "Updated. Check status with: $(journal_hint)"
