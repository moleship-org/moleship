# Agents

## About

The name of this project is "moleship". It is a RESTful API backend service that allows you to create, read, update, and delete Quadlet-generated systemd unit files for Podman containers.

Moleship is designed to run on the same device as the Podman containers it manages. It does not work as a distributed system.

## Dependencies

* go 1.27+
* systemd 257+
* podman 5.4+
* gpgme 1.24+
* make 4.4+

## How to build

To build the project, run `make`, which uses `Makefile` and `scripts/build.sh`.

All targetable binaries live in the `cmd/` directory. You can change the target binary by setting the `BIN` environment variable; its value must match a valid binary name in `cmd/`. The default target is `moleship`.

The build process generates compiled binaries under `_output/bin/{GOOS}/{ARCH}`.
