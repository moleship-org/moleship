#!/usr/bin/env bash
# Shared helpers for the moleship setup scripts (install.sh, update.sh,
# uninstall.sh). Not meant to be run directly.

SETUP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SETUP_DIR}/../.." && pwd)"

CONTAINERFILE="${ROOT_DIR}/containers/Containerfile"
UNIT_NAME="moleship.container"
SERVICE_NAME="moleship.service"

IMAGE_TAG="localhost/moleship:latest"
ROOTFUL=0

log()  { printf '\033[1;34m[moleship]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[moleship]\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[1;31m[moleship]\033[0m %s\n' "$*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "'$1' is required but was not found in PATH."
}

# Scans (without consuming) the action's arguments for flags shared across
# install/update/uninstall, so each script can still do its own parsing pass
# for action-specific flags afterwards.
parse_common_flags() {
  local args=("$@")
  local i=0
  while [[ $i -lt ${#args[@]} ]]; do
    case "${args[$i]}" in
      --rootful)
        ROOTFUL=1
        ;;
      --tag)
        i=$((i + 1))
        IMAGE_TAG="${args[$i]:-${IMAGE_TAG}}"
        ;;
    esac
    i=$((i + 1))
  done
}

require_rootful_privileges() {
  if [[ "${ROOTFUL}" -eq 1 && "${EUID}" -ne 0 ]]; then
    die "--rootful requires running as root, e.g.: sudo $0 --rootful"
  fi
}

# Wraps systemctl so every caller automatically targets the right manager
# (user session vs. system-wide) based on the --rootful flag.
systemctl_() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    systemctl "$@"
  else
    systemctl --user "$@"
  fi
}

quadlet_dir() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf '%s' "/etc/containers/systemd"
  else
    printf '%s' "${HOME}/.config/containers/systemd"
  fi
}

# Picks the right template based on --rootful. Both get installed under the
# same UNIT_NAME (moleship.container) in their respective quadlet_dir, so
# the resulting service is always named moleship.service either way.
quadlet_src() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf '%s' "${ROOT_DIR}/containers/systemd/moleship-rootful.container"
  else
    printf '%s' "${ROOT_DIR}/containers/systemd/moleship-rootless.container"
  fi
}

# The following three mirror moleship's own defaults
# (internal/config/config.go) and must match the Volume= host paths in
# containers/systemd/moleship-rootless.container (%h/.config/moleship etc.)
# and moleship-rootful.container (/etc/moleship etc.). They're bind-mount
# sources for the running container, so they need to exist upfront with
# the right ownership -- otherwise Podman may auto-create them with
# ownership that doesn't line up with the UID the container runs as, and
# moleship will fail to write its JWT secret / config on first start.
config_home_dir() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf '%s' "/etc/moleship"
  else
    printf '%s' "${HOME}/.config/moleship"
  fi
}

cache_home_dir() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf '%s' "/var/cache/moleship"
  else
    printf '%s' "${HOME}/.cache/moleship"
  fi
}

data_home_dir() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf '%s' "/var/lib/moleship"
  else
    printf '%s' "${HOME}/.local/share/moleship"
  fi
}

# Creates all of the host directories moleship's container bind-mounts into
# itself, with the right ownership, before the service ever starts.
prepare_state_dirs() {
  local dir
  for dir in "$(config_home_dir)" "$(cache_home_dir)" "$(data_home_dir)" "$(quadlet_dir)"; do
    mkdir -p "${dir}"
  done
}

journal_hint() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf 'journalctl -u %s -e' "${SERVICE_NAME}"
  else
    printf 'journalctl --user -u %s -e' "${SERVICE_NAME}"
  fi
}

# Text form of the systemctl_ wrapper above, for use in log messages that
# tell the user what command to run themselves.
systemctl_cmd() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf 'systemctl'
  else
    printf 'systemctl --user'
  fi
}
