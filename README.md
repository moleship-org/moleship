# Moleship

Moleship RESTful API backend service for managing Quadlet unit files.

## Requirements

Core dependencies:

* systemd 257+
* podman v5+
* go v1.27+

Build dependencies:

* gpgme
* btrfs
* devmapper
* seccomp

### Debian/Ubuntu Install

```sh
apt install -y \
    podman \
    podman-compose \
    libgpgme-dev \
    libbtrfs-dev \
    libdevmapper-dev \
    libseccomp-dev
```

## Building

If you have `make` installed, run:

```sh
make
```

You'll find the resulting binary in `_output/bin/{GOOS}/{GOARCH}/{BIN}`.

**Important**: The build process uses [scripts/build.sh](./scripts/build.sh).

## Docs

### Godoc

```sh
go doc -http
```
