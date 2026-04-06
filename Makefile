# Binary output name
BIN ?= moleship

# Package name
PKG := github.com/moleship-org/moleship

# Architecture
ARCH ?= $(shell go env GOOS)-$(shell go env GOARCH)

# Program version
VERSION ?= main

# Output directory
OUTPUT_DIR ?= _oputput

# Go environment
platform = $(subst -, ,$(ARCH))
GOOS = $(word 1, $(platform))
GOARCH = $(word 2, $(platform))
GOPROXY ?= "https://proxy.golang.org,direct"

all:
	@$(MAKE) build

build: _output/bin/$(GOOS)/$(GOARCH)/$(BIN)

_output/bin/$(GOOS)/$(GOARCH)/$(BIN): build-dirs
	@echo "building: $@"
		GOOS=$(GOOS) \
		GOARCH=$(GOARCH) \
		VERSION=$(VERSION) \
		PKG=$(PKG) \
		BIN=$(BIN) \
		OUTPUT_DIR=$$(pwd)/_output/bin/$(GOOS)/$(GOARCH) \
		./scripts/build.sh

build-dirs:
	@mkdir -p _output/bin/$(GOOS)/$(GOARCH)

clean:
	@rm -rf _output

tidy:
	go mod tidy

run: build
	$$(pwd)/_output/bin/$(GOOS)/$(GOARCH)/$(BIN) $(ARGS)

.PHONY: all clean tidy run