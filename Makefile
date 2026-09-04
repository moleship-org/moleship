# Binary output name
BIN ?= moleship

# Package name
PKG := github.com/moleship-org/moleship

# Architecture
ARCH ?= $(shell go env GOOS)-$(shell go env GOARCH)

# Program version
VERSION ?= main

# Output directory
OUTPUT_DIR ?= _oputput/bin

# Go environment
platform = $(subst -, ,$(ARCH))
GOOS = $(word 1, $(platform))
GOARCH = $(word 2, $(platform))
GOPROXY ?= "https://proxy.golang.org,direct"

.PHONY: all clean tidy run

all:
	@$(MAKE) build

build: $(OUTPUT_DIR)/$(GOOS)/$(GOARCH)/$(BIN)

$(OUTPUT_DIR)/$(GOOS)/$(GOARCH)/$(BIN): build-dirs
	@echo "building: $@"
		GOOS=$(GOOS) \
		GOARCH=$(GOARCH) \
		VERSION=$(VERSION) \
		PKG=$(PKG) \
		BIN=$(BIN) \
		OUTPUT_DIR=$$(pwd)/$(OUTPUT_DIR)/$(GOOS)/$(GOARCH) \
		./scripts/build.sh

build-dirs:
	@mkdir -p $(OUTPUT_DIR)/$(GOOS)/$(GOARCH)

clean:
	@rm -rf $(OUTPUT_DIR)

tidy:
	go mod tidy

run: build
	$$(pwd)/$(OUTPUT_DIR)/$(GOOS)/$(GOARCH)/$(BIN) $(ARGS)
