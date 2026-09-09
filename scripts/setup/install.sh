#!/usr/bin/env bash
# Pull the moleship image, deploy it as a Podman Quadlet unit, and start it.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

parse_common_flags "$@"
require_rootful_privileges

if quadlet_installed "${UNIT_NAME}"; then
  log_warn "Unit ${UNIT_NAME} already exists; refusing to overwrite."
  read -p "Do you want to overwrite? [y/N] " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

require_cmd podman
require_cmd systemctl

log_info "Pulling ${IMAGE_TAG}..."
podman pull "${IMAGE_TAG}"

if [[ "${ROOTFUL}" -eq 0 ]]; then
  log "Enabling the rootless Podman socket..."
  systemctl_ enable --now podman.socket

  log "Enabling linger for $(whoami) (keeps your session alive across logout)..."
  if command -v loginctl >/dev/null 2>&1; then
    sudo loginctl enable-linger "$(whoami)" \
      || log_warn "Could not enable linger automatically; run manually: sudo loginctl enable-linger $(whoami)"
  else
    log_warn "'loginctl' not found; skipping linger setup."
  fi
else
  log "Enabling the rootful Podman socket..."
  systemctl_ enable --now podman.socket
fi

log "Preparing moleship's state directories..."
prepare_state_dirs

DEST_DIR="$(quadlet_dir)"
DEST_FILE="${DEST_DIR}/${UNIT_NAME}"

log "Installing quadlet unit to ${DEST_FILE}..."
install -m 0644 "$(quadlet_src)" "${DEST_FILE}"

if [[ "${IMAGE_TAG}" != "$(default_image_tag)" ]]; then
  sed -i "s#^Image=.*#Image=${IMAGE_TAG}#" "${DEST_FILE}"
fi

log_info "Starting moleship..."
systemctl_ daemon-reload
systemctl_ start "${SERVICE_NAME}"

PORT="${MOLESHIP_PORT:-5000}"
if command -v curl >/dev/null 2>&1; then
  log "Waiting for moleship to become healthy..."
  for _ in $(seq 1 30); do
    if curl -fs "http://localhost:${PORT}/api/v1/health/" >/dev/null 2>&1; then
      log_success "moleship is up: http://localhost:${PORT}"
      log ""
      log "Log in once to set your password (username must be '$(whoami)'):"
      log "  curl -X POST http://localhost:${PORT}/api/v1/auth/login \\"
      log "    -H 'Content-Type: application/json' \\"
      log "    -d '{\"username\":\"$(whoami)\",\"password\":\"<choose-one>\"}'"
      exit 0
    fi
    sleep 1
  done
  log_warn "moleship did not report healthy within 30s. Check logs with:"
  log_warn "  $(journal_hint)"
  exit 1
fi

log "moleship service started. Check status with: $(journal_hint)"
