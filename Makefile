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

work: $(GOBIN) ## Ensure working directory is ready

build: ## Build binary for current OS/arch
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags $(LDFLAGS) -o build/lingualeo-$(GOOS)-$(GOARCH) $(CMD_PACKAGE)

cache: ## Clean Go build cache
	go clean --cache

install: ## Install binary to GOPATH/bin
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go install -ldflags $(LDFLAGS) $(CMD_PACKAGE)

test: unit ## Run all tests

lint: work tools fmt fix vet goimports golangci govulncheck ## Run all linters

unit: work ## Run unit tests
	go test -count 1 -tags=unit $(TESTARGS) ./...

fmt: ## Format Go code
	go fmt ./...

fix: ## Apply go fix
	go fix ./...

goimports: ## Check imports formatting
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

vet: ## Run go vet
	go vet ./...

golangci: ## Run golangci-lint
ifndef HAS_GOLANGCI
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
endif
	golangci-lint run

govulncheck: ## Run vulnerability check
ifndef HAS_GOVULNCHECK
	@echo "Installing govulncheck"
	go install golang.org/x/vuln/cmd/govulncheck@latest
endif
	govulncheck ./...

generate: ## Generate code (mocks, etc.)
ifndef HAS_MOCKERY
	@echo "Installing mockery $(MOCKERY_VERSION)"
	go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
endif
	go generate ./...

cover: work ## Run tests with coverage
	go test $(TESTARGS) -tags=unit -coverprofile=coverage.out -coverpkg=./ ./...

tidy: ## Check go.mod is tidy
	go mod tidy
	git diff --exit-code go.mod go.sum

tools: $(GOBIN) ## Install development tools
ifndef HAS_GOLANGCI
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
endif
ifndef HAS_GOIMPORTS
	@echo "Installing goimports"
	go install golang.org/x/tools/cmd/goimports@latest
endif
ifndef HAS_MOCKERY
	@echo "Installing mockery $(MOCKERY_VERSION)"
	go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
endif
ifndef HAS_GOVULNCHECK
	@echo "Installing govulncheck"
	go install golang.org/x/vuln/cmd/govulncheck@latest
endif

shell: ## Start interactive shell
	$(SHELL) -i

clean: work ## Clean build artifacts
	rm -rf $(BINARY)
	rm -rf build/

version: ## Print version
	@echo ${VERSION}

ci: clean cache generate lint test ## Run full CI workflow

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: install build cover work fmt fix test version clean tools generate lint check ci tidy help govulncheck
