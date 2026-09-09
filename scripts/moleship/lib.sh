#!/usr/bin/env bash
# Shared helpers for the moleship setup scripts. Not meant to be run directly.

SETUP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SETUP_DIR}/../.." && pwd)"

CONTAINERFILE="${ROOT_DIR}/containers/Containerfile"
UNIT_NAME="moleship.container"
SERVICE_NAME="moleship.service"

default_image_tag() { printf '%s' "ghcr.io/moleship-org/moleship:latest"; }

IMAGE_TAG="$(default_image_tag)"
ROOTFUL=0

# --- output helpers ----------------------------------------------------------

if [ -t 1 ]; then
  C_RESET=$'\033[0m'
  C_RED=$'\033[31m'
  C_GREEN=$'\033[32m'
  C_YELLOW=$'\033[33m'
  C_BLUE=$'\033[34m'
else
  C_RESET=""
  C_RED=""
  C_GREEN=""
  C_YELLOW=""
  C_BLUE=""
fi

log_info()    { printf '%s[info]%s %s\n'  "$C_BLUE"   "$C_RESET" "$*"; }
log_success() { printf '%s[ ok ]%s %s\n'  "$C_GREEN"  "$C_RESET" "$*"; }
log_warn()    { printf '%s[warn]%s %s\n'  "$C_YELLOW" "$C_RESET" "$*" >&2; }
log_error()   { printf '%s[fail]%s %s\n'  "$C_RED"    "$C_RESET" "$*" >&2; }

# Generic-purpose logger
log() { printf '%s\n' "$@"; }

# --- command helpers ---------------------------------------------------------

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || log_error "'$1' is required but was not found in PATH."
}

# Scans (without consuming) the action's argument flags
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
    log_error "--rootful requires running as root, e.g.: sudo $0 --rootful"
  fi
}

# Wraps systemctl so every caller automatically targets the right manager
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

quadlet_src() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf '%s' "${ROOT_DIR}/containers/systemd/moleship-rootful.container"
  else
    printf '%s' "${ROOT_DIR}/containers/systemd/moleship-rootless.container"
  fi
}

# --- path helpers -------------------------------------------------------------

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

systemctl_cmd() {
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    printf 'systemctl'
  else
    printf 'systemctl --user'
  fi
}

unit_exists() {
  local unit="${1}.service"
  if [[ "${ROOTFUL}" -eq 1 ]]; then
    systemctl list-unit-files --no-legend | awk '{print $1}' | grep -qxF "${unit}"
  else
    systemctl --user list-unit-files --no-legend | awk '{print $1}' | grep -qxF "${unit}"
  fi
}

quadlet_installed() {
  local file="${1:-${UNIT_NAME}}"
  [[ -e "$(quadlet_dir)/${file}" ]]
}
