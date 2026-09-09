# Agents

## About

The name of this project is "moleship". It is a RESTful API backend service that allows you to create, read, update, and delete Quadlet-generated systemd unit files for Podman containers.

Moleship is designed to run on the same device as the Podman containers it manages. It does not work as a distributed system.

## Core Dependencies

* golang
* systemd
* podman

## How to build

To build the project, run `make`, which uses `Makefile` and `scripts/build.sh`.

All targetable binaries live in the `cmd/` directory. You can change the target binary by setting the `BIN` environment variable; its value must match a valid binary name in `cmd/`. The default target is `moleship`.

The build process generates compiled binaries under `_output/bin/{GOOS}/{ARCH}`.

## Must follow

1. Human-working memory is small. Keep things in memory that can be easily recomputed.
2. Knowing the answer is not doing the answer. The friction between "got it" and "done it" is where work dies.
3. Starting is the hardest step. The first action must be obvious, small, and doable now.
4. Time estimates feel uniform. "A bit of work" and "a few hours" register the same. Vague estimates fail.
5. Visible progress is the best indicator of success. It should be easy to see what work has been done and what remains to be done.
6. Simple is better than complex. Prefer simple solutions over complex ones. Simple do not imply "repeat yourself" or "be lazy". You need to keep things clear, concise, modular and cohesive.
7. When you make a decision, consider the ramifications of your decision in the global context.
8. Any enumeration of steps must be clear and concise. Avoid jargon and use crystal clear language.
9. No preambles, no recaps, no closing pleasantries, just go straight to the point.
10. You can make exceptions to the above guidelines when necessary. You may need to deviate from the guidelines when the situation requires it. Don't struggle to follow the guidelines when they are not applicable.
