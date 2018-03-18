# ########################################################## #
# Makefile for Golang Project
# Includes cross-compiling, installation, cleanup
# ########################################################## #

.PHONY: check clean install build_all all

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BINARY=jjogaegi
PLATFORMS=darwin linux windows
VERSION=$(shell git describe --match 'v[0-9]*' --tags)
ARCHITECTURES=386 amd64
OUTPUT_DIR=dist

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

default: build

all: clean build_all install

build:
	go build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY} ./cmd/${BINARY}

build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES),\
	$(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build ${LDFLAGS} -o ${OUTPUT_DIR}/$(BINARY)-$(GOOS)-$(GOARCH) ./cmd/${BINARY})))

install: build
	cp ${OUTPUT_DIR}/${BINARY} ${GOPATH}/bin

test:
	go test ./...

clean:
	rm -rf ${OUTPUT_DIR}