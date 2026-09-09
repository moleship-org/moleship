#!/usr/bin/env bash
# Stop and remove the moleship Quadlet deployment.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

parse_common_flags "$@"
require_rootful_privileges

PURGE=0
for arg in "$@"; do
  case "$arg" in
    --purge) PURGE=1 ;;
  esac
done

require_cmd systemctl

DEST_FILE="$(quadlet_dir)/${UNIT_NAME}"

log_info "Stopping moleship..."
systemctl_ stop "${SERVICE_NAME}" 2>/dev/null || true
systemctl_ disable "${SERVICE_NAME}" 2>/dev/null || true

log "Removing quadlet unit ${DEST_FILE}..."
rm -f "${DEST_FILE}"

systemctl_ daemon-reload

if [[ "${PURGE}" -eq 1 ]]; then
  log_info "Purging persisted data..."
  rm -rf "$(config_home_dir)" "$(cache_home_dir)" "$(data_home_dir)"

  if command -v podman >/dev/null 2>&1; then
    podman rmi "${IMAGE_TAG}" 2>/dev/null || true
  fi
else
  log_warn "Persisted data left in place under $(config_home_dir), $(cache_home_dir) and $(data_home_dir)."
  log_warn "Re-run with --purge to remove it (and the built image) too."
fi

log_success "moleship uninstalled."
