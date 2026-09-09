#!/usr/bin/env bash
# Pull the latest moleship image and restart the running service.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

log_info "Updating moleship..."

parse_common_flags "$@"
require_rootful_privileges

require_cmd podman
require_cmd systemctl

DEST_FILE="$(quadlet_dir)/${UNIT_NAME}"
if [[ ! -f "${DEST_FILE}" ]]; then
  log_error "No existing install found at ${DEST_FILE}. Run install.sh first."
fi

log_info "Pulling ${IMAGE_TAG}..."
podman pull "${IMAGE_TAG}"

log_info "Restarting moleship..."
systemctl_ daemon-reload
systemctl_ restart "${SERVICE_NAME}"

log_success "Updated. Check status with: $(journal_hint)"
