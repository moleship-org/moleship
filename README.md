# Moleship

Moleship is a tool for managing Quadlet unit files with a RESTful API.

## Requirements

Core dependencies:

* systemd 257+
* podman v5+
* go v1.27+

### Debian/Ubuntu Install

```sh
apt install -y podman podman-compose
```

## Building

If you have `make` installed, run:

```sh
make
```

You'll find the resulting binary in `_output/bin/{GOOS}/{GOARCH}/{BIN}`.

**Important**: The build process uses [scripts/build.sh](./scripts/build.sh).

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

### Godoc

    go doc -http
