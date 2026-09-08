#!/usr/bin/env bash
# Pull the moleship image, deploy it as a Podman Quadlet unit, and start it.
#
# Usage:
#   scripts/setup/install.sh [--rootful] [--tag <image:tag>]
#
# This pulls a prebuilt image rather than building from source -- see
# scripts/setup/publish.sh to build and push one. If the image is private,
# log in to its registry first (e.g. `podman login ghcr.io`).
#
# Rootless (default): installs to ~/.config/containers/systemd and manages
# the service under `systemctl --user`. Requires no special privileges,
# except for a one-time `loginctl enable-linger`, which the script requests
# via sudo.
#
# Rootful (--rootful): installs to /etc/containers/systemd and manages the
# service under the system `systemctl`. Must be run as root (e.g. via sudo).

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

parse_common_flags "$@"
require_rootful_privileges

require_cmd podman
require_cmd systemctl

log "Pulling ${IMAGE_TAG}..."
podman pull "${IMAGE_TAG}"

if [[ "${ROOTFUL}" -eq 0 ]]; then
  log "Enabling the rootless Podman socket..."
  systemctl_ enable --now podman.socket

  log "Enabling linger for $(whoami) (keeps your session alive across logout)..."
  if command -v loginctl >/dev/null 2>&1; then
    sudo loginctl enable-linger "$(whoami)" \
      || warn "Could not enable linger automatically; run manually: sudo loginctl enable-linger $(whoami)"
  else
    warn "'loginctl' not found; skipping linger setup."
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

log "Starting moleship..."
systemctl_ daemon-reload
systemctl_ start "${SERVICE_NAME}"

PORT="${MOLESHIP_PORT:-5000}"
if command -v curl >/dev/null 2>&1; then
  log "Waiting for moleship to become healthy..."
  for _ in $(seq 1 30); do
    if curl -fs "http://localhost:${PORT}/api/v1/health/" >/dev/null 2>&1; then
      log "moleship is up: http://localhost:${PORT}"
      log ""
      log "Log in once to set your password (username must be '$(whoami)'):"
      log "  curl -X POST http://localhost:${PORT}/api/v1/auth/login \\"
      log "    -H 'Content-Type: application/json' \\"
      log "    -d '{\"username\":\"$(whoami)\",\"password\":\"<choose-one>\"}'"
      exit 0
    fi
    sleep 1
  done
  warn "moleship did not report healthy within 30s. Check logs with:"
  warn "  $(journal_hint)"
  exit 1
fi

log "moleship service started. Check status with: $(journal_hint)"
