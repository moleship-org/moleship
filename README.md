# Moleship

Moleship is a tool for managing Quadlet unit files with a cohesive API.

## Requirements

Core dependencies:

* systemd 257.9^
* podman v5.4.2^
* go v1.26.1^
* libgpgme v1.24.2^
* libassuan v3.0.2^
* libgpg-error v1.51^

### Debian/Ubuntu Install
  
    apt install -y podman podman-compose libgpgme-dev libassuan-dev libgpg-error-dev

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

## Docs

### Swaggger

    swag init -g cmd/moleship/main.go -d ./ --parseDependency --parseInternal

### Godoc

    go doc -http
