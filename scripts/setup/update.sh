#!/usr/bin/env bash
# Rebuild the moleship image and restart the running Quadlet service to pick
# it up.
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

log "Rebuilding ${IMAGE_TAG}..."
podman build -f "${CONTAINERFILE}" -t "${IMAGE_TAG}" "${ROOT_DIR}"

log "Restarting moleship..."
systemctl_ daemon-reload
systemctl_ restart "${SERVICE_NAME}"

log "Updated. Check status with: $(journal_hint)"
