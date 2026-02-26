# lingualeo Makefile

.DEFAULT_GOAL := help

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
HAS_GOLANGCI := $(shell command -v golangci-lint;)
HAS_GOIMPORTS := $(shell command -v goimports;)
HAS_MOCKERY := $(shell command -v mockery;)
HAS_GOVULNCHECK := $(shell command -v govulncheck;)
GOLANGCI_LINT_VERSION := v2.10.0
MOCKERY_VERSION := v3.6.4
GO_RUNTIME_VERSION := $(shell go env GOVERSION)

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
	@echo "Creating GOBIN directory"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags $(LDFLAGS) -o build/lingualeo-$(GOOS)-$(GOARCH) $(CMD_PACKAGE)

cache:
	go clean --cache

install:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go install -ldflags $(LDFLAGS) $(CMD_PACKAGE)

test: unit

lint: work tools fmt fix vet goimports golangci govulncheck

unit: work
	go test -count 1 -tags=unit $(TESTARGS) ./...

fmt:
	go fmt ./...

fix:
	go fix ./...

goimports:
ifndef HAS_GOIMPORTS
	@echo "Installing goimports"
	go install golang.org/x/tools/cmd/goimports@latest
endif
	@unformatted="$$(goimports -l $(shell find . -path ./.go -prune -o -type f -iname "*.go" -print))"; \
	if [ -n "$$unformatted" ]; then \
		echo "goimports check failed. Run goimports on:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	go vet ./...

golangci: ensure-golangci
	golangci-lint run

govulncheck:
ifndef HAS_GOVULNCHECK
	@echo "Installing govulncheck"
	go install golang.org/x/vuln/cmd/govulncheck@latest
endif
	govulncheck ./...

generate: ensure-mockery
	go generate ./...

cover: work
	go test $(TESTARGS) -tags=unit -coverprofile=coverage.out -coverpkg=./ ./...

tidy:
	go mod tidy
	git diff --exit-code go.mod go.sum

tools: $(GOBIN) ensure-golangci ensure-mockery
ifndef HAS_GOIMPORTS
	@echo "Installing goimports"
	go install golang.org/x/tools/cmd/goimports@latest
endif
ifndef HAS_GOVULNCHECK
	@echo "Installing govulncheck"
	go install golang.org/x/vuln/cmd/govulncheck@latest
endif

ensure-golangci:
	@set -e; \
	if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)"; \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION); \
	else \
		installed_version="$$(golangci-lint version | awk '{for (i=1;i<=NF;i++) if ($$i=="version") {print "v" $$(i+1); exit}}')"; \
		built_with_go="$$(go version -m "$$(command -v golangci-lint)" | awk 'NR==1 {print $$2}')"; \
		built_with_go_base="$${built_with_go%%-*}"; \
		go_runtime_base="$(GO_RUNTIME_VERSION)"; \
		go_runtime_base="$${go_runtime_base%%-*}"; \
		if [ "$$installed_version" != "$(GOLANGCI_LINT_VERSION)" ] || [ "$$built_with_go_base" != "$$go_runtime_base" ]; then \
			echo "Reinstalling golangci-lint $(GOLANGCI_LINT_VERSION) (found $$installed_version built with $$built_with_go, need $(GO_RUNTIME_VERSION))"; \
			curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION); \
		fi; \
	fi

ensure-mockery:
	@set -e; \
	if ! command -v mockery >/dev/null 2>&1; then \
		echo "Installing mockery $(MOCKERY_VERSION)"; \
		go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION); \
	else \
		installed_version="$$(mockery version)"; \
		built_with_go="$$(go version -m "$$(command -v mockery)" | awk 'NR==1 {print $$2}')"; \
		built_with_go_base="$${built_with_go%%-*}"; \
		go_runtime_base="$(GO_RUNTIME_VERSION)"; \
		go_runtime_base="$${go_runtime_base%%-*}"; \
		if [ "$$installed_version" != "$(MOCKERY_VERSION)" ] || [ "$$built_with_go_base" != "$$go_runtime_base" ]; then \
			echo "Reinstalling mockery $(MOCKERY_VERSION) (found $$installed_version built with $$built_with_go, need $(GO_RUNTIME_VERSION))"; \
			go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION); \
		fi; \
	fi

shell:
	$(SHELL) -i

clean: work
	rm -rf $(BINARY)
	rm -rf build/

version:
	@echo ${VERSION}

ci: clean cache generate lint test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: install build cover work fmt fix test version clean tools generate lint check ci tidy help govulncheck ensure-golangci ensure-mockery
