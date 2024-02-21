# The binary to build (just the basename).
BIN=promc

# Where to push the docker image.
OUT_DIR=./out

# This version-strategy uses git tags to set the version string
VERSION=$(shell git describe --tags --always --dirty)

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Build the project
all: build

build:
	@echo "Building ${BIN} to ${OUT_DIR}"
	@mkdir -p ${OUT_DIR}
	@go build ${LDFLAGS} -o ${OUT_DIR}/${BIN} cmd/promc/main.go
	@echo "Build complete"

clean:
	@echo "Cleaning"
	@rm -rf ${OUT_DIR}/${BIN}
	@echo "Clean complete"

.PHONY: build clean
