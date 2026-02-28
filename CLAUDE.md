# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Lingualeo CLI helper - a console application for translating words via the Lingualeo API. Supports word pronunciation, dictionary management, and image visualization.

## Build and Test Commands

```bash
# Build binary
make build

# Install to GOPATH/bin
make install

# Run all tests with race detection
make test
# Or directly:
go test -v -race -tags=unit ./...

# Run specific test
go test -v -race -tags=unit ./internal/translator/... -run TestRussianWord

# Lint (runs fmt, vet, goimports, golangci-lint)
make lint

# Generate mocks
make generate

# Full CI workflow (used in GitHub Actions)
make clean cache generate lint test
```

## Environment

- Go 1.26 with `GOEXPERIMENT=jsonv2` enabled (required for json v2 encoding)
- golangci-lint v2.10.0 for linting
- mockery v3 for mock generation

## Key Patterns

**Dependency Injection via Options:**
The `Lingualeo` struct uses functional options (`Option func(*Lingualeo error)`) for injecting mock implementations:
- `WithTranslator(t Translator)` - for mocking API calls
- `WithDownloader(d Downloader)` - for mocking file downloads
- `WithPronouncer(p Pronouncer)` - for mocking audio playback
- `WithOutputer(o Outputer)` - for mocking visualization

**Concurrent Channel Pipeline:**
Translation uses a channel-based pipeline with `sync.WaitGroup.Go()` for concurrent word processing. The `channel.OrDone()` pattern handles context cancellation.

**Config Precedence:**
Default configs (`~/lingualeo.[toml|yml|yaml|json]`) → CLI flags → Explicit `-c` config file

## Git commit

**Do not use Claude Code as contributor or coauthor**
