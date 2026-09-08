#!/usr/bin/env bash
# One-command setup for moleship: builds the image, deploys it as a Podman
# Quadlet unit, and starts the service.
#
# Usage:
#   ./scripts/quick-setup.sh [install|update|uninstall] [options...]
#
# Defaults to "install" when no action is given.
#
# Common options (see scripts/setup/<action>.sh for the full list):
#   --rootful       Manage a rootful (system-wide) deployment instead of rootless
#   --tag <image>   Image tag to build/use (default: localhost/moleship:latest)
#   --purge         (uninstall only) also remove persisted data and the image

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SETUP_DIR="${SCRIPT_DIR}/setup"

usage() {
  cat <<'EOF'
Usage: quick-setup.sh [action] [options...]

Actions:
  install     Build the image, deploy the Quadlet unit, and start moleship (default)
  update      Rebuild the image and restart the running service
  uninstall   Stop and remove the Quadlet deployment

Common options:
  --rootful       Manage a rootful (system-wide) deployment instead of rootless
  --tag <image>   Image tag to build/use (default: localhost/moleship:latest)
  --purge         (uninstall only) also remove persisted data and the image
  -h, --help      Show this help
EOF
}

case "${1:-}" in
  -h|--help)
    usage
    exit 0
    ;;
  install|update|uninstall)
    ACTION="$1"
    shift
    ;;
  ""|--*)
    # No action given, or the first token is an option (e.g. --tag): default
    # to install and leave it in place for install.sh to parse.
    ACTION="install"
    ;;
  *)
    echo "Unknown action: $1" >&2
    usage
    exit 1
    ;;
esac

exec "${SETUP_DIR}/${ACTION}.sh" "$@"
