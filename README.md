# Moleship

Moleship is a tool for managing files Quadlet unit files with a cohesive API.

## Requirements

Core dependencies:

* systemd 257.9^
* podman v5.4.2^
* go v1.26.1^
* libgpgme v1.24.2^
* libassuan v3.0.2^
* libgpg-error v1.51^

## Building

If you have `make` installed, run:

    make

You'll find the resulting binary in `_output/bin/{GOOS}/{GOARCH}/{BIN}`.

**Import**: The build process uses [scripts/build.sh](./scripts/build.sh) and `CGO_ENABLED=1`.

## Project structure

    - _output/                  # Binaries
    - cmd/                      # Entry points
    - internal/                 # Source
      - adapter/                # External providers
      - core/                   # Core features, application and handlers
      - domain/                 # Business logic entities and interface
        - model/                # Data structs or records
        - port/                 # In and out interfaces
