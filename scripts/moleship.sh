#!/usr/bin/env bash
# Command handler for moleship

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MOLESHIP_DIR="${SCRIPT_DIR}/moleship"

usage() {
  cat <<'EOF'
Usage: moleship.sh [action] [options...]

Actions:
  install     Build the image, deploy the Quadlet unit, and start moleship (default)
  update      Rebuild the image and restart the running service
  uninstall   Stop and remove the Quadlet deployment
  build       Build the image locally
  publish     Build and push the image to its registry (e.g. ghcr.io)

Common options:
  --rootful       Manage a rootful (system-wide) deployment instead of rootless
  --tag <image>   Image tag to build/use (default: ghcr.io/moleship-org/moleship:latest)
  --purge         (uninstall only) also remove persisted data and the image
  --version <v>   (publish only) also build/push a version-tagged image
  -h, --help      Show this help
EOF
}

case "${1:-}" in
  -h|--help)
    usage
    exit 0
    ;;
  install|update|uninstall|publish|build)
    ACTION="$1"
    shift
    ;;
  ""|--*)
    # No action given, default to install.
    ACTION="install"
    ;;
  *)
    echo "Unknown action: $1" >&2
    usage
    exit 1
    ;;
esac

exec "${MOLESHIP_DIR}/${ACTION}.sh" "$@"
