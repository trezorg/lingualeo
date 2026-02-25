# lingualeo Makefile

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))
# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(shell go env GOPATH)/.go
OLD_PATH := ${PATH}
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
export GO111MODULE := on
export GOEXPERIMENT ?= jsonv2
export PATH=${OLD_PATH}:${GOBIN}
TESTARGS_DEFAULT := -v -race
TESTARGS ?= $(TESTARGS_DEFAULT)
PKG := $(shell awk '/^module/ { print $$2 }' go.mod)
DEST := $(GOPATH)/src/$(GIT_HOST)/$(BASE_DIR)
SOURCES := $(shell find $(DEST) -name '*.go' 2>/dev/null)
HAS_GOLANGCI := $(shell command -v golangci-lint;)
HAS_GOIMPORTS := $(shell command -v goimports;)
HAS_MOCKERY := $(shell command -v mockery;)
GOLANGCI_LINT_VERSION := v2.10.0
MOCKERY_VERSION := v3.6.4

TARGETS		?= darwin/amd64 linux/amd64 linux/386 linux/arm linux/arm64 linux/ppc64le linux/s390x
DIST_DIRS	= find * -type d -exec

TEMP_DIR	:=$(shell mktemp -d)

VERSION		?= $(shell git describe --tags 2> /dev/null || \
			   git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)
GOARCH		:= amd64
TAGS		:=
LDFLAGS		:= "-w -s -X 'main.version=${VERSION}'"
CMD_PACKAGE := ./cmd/lingualeo
BINARY 		:= ./lingualeo
GOOS		?= $(shell go env GOOS)
GOARCH		?= $(shell go env GOARCH)

# CTI targets

$(GOBIN):
	echo "create gobin"
	mkdir -p $(GOBIN)

work: $(GOBIN)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags $(LDFLAGS) -o build/lingualeo-$(GOOS)-$(GOARCH) $(CMD_PACKAGE)

cache:
	go clean --cache

install:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go install -ldflags $(LDFLAGS) $(CMD_PACKAGE)

test: unit

lint: work prepare fmt fix vet goimports golangci
unit: work
	go test -count 1 -tags=unit $(TESTARGS) ./...

fmt:
	go fmt ./...

fix:
	go fix ./...

goimports:
ifndef HAS_GOIMPORTS
	echo "installing goimports"
	go install golang.org/x/tools/cmd/goimports@latest
endif
	@unformatted="$$(goimports -l $(shell find . -path ./.go -prune -o -type f -iname "*.go"))"; \
	if [ -n "$$unformatted" ]; then \
		echo "goimports check failed. Run goimports on:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	go vet ./...

golangci:
ifndef HAS_GOLANGCI
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
endif
	golangci-lint run

generate:
ifndef HAS_MOCKERY
	echo "installing mockery"
	go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
endif
	go generate ./...

cover: work
	go test $(TESTARGS) -tags=unit -cover -coverpkg=./ ./...


prepare:
ifndef HAS_GOLANGCI
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
endif
	echo "golangci-lint already installed"
ifndef HAS_GOIMPORTS
	echo "installing goimports"
	go install golang.org/x/tools/cmd/goimports@latest
endif
	echo "goimports already installed"
ifndef HAS_MOCKERY
	echo "installing mockery"
	go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
endif
	echo "mockery already installed"

shell:
	$(SHELL) -i

clean: work
	rm -rf $(BINARY)

version:
	@echo ${VERSION}

check: clean generate lint test

.PHONY: install build cover work fmt fix test version clean prepare generate lint check
