# Repository Guidelines

## Project Structure & Module Organization
The CLI entrypoint is `cmd/lingualeo/main.go`. Core application code lives under `internal/` and is split by domain (for example `internal/api`, `internal/translator`, `internal/files`, `internal/visualizer`, `internal/validator`). Keep new packages focused and cohesive; prefer adding to an existing domain package before creating a new top-level folder. Build artifacts are written to `build/`. CI workflows are in `.github/workflows/`.

## Build, Test, and Development Commands
- `make build`: builds the binary to `build/lingualeo-<os>-<arch>` with version ldflags.
- `make test` or `make unit`: runs `go test -tags=unit -v -race ./...`.
- `make lint`: runs formatting, `go fix`, `go vet`, `goimports`, and `golangci-lint`.
- `make cover`: runs unit tests with coverage enabled.
- `make generate`: runs `go generate ./...` (mock generation via `mockery`).
- `make clean cache`: removes the local binary and Go build cache.

For local sanity before opening a PR, run: `make ci`.
After finishing the task always run `make ci`.

## Coding Style & Naming Conventions
Use standard Go style (`gofmt` output, tabs for indentation). Package names should be short, lowercase, and noun-like. Exported identifiers use `CamelCase`; unexported identifiers use `camelCase`. Test files must end with `_test.go`. Keep functions small enough to satisfy configured linters in `.golangci.yaml` (notably `revive`, `gocritic`, `staticcheck`, `gosec`).

## Testing Guidelines
Tests use Go’s `testing` package plus `testify`. Place tests next to implementation files inside the same package. Prefer table-driven tests for parser/validator/translator logic. Run targeted tests with `go test ./internal/<package> -run <TestName>` and full checks with `make test`.

## Commit & Pull Request Guidelines
Recent commits follow short, imperative messages (for example: `Fix linters`, `Create api.Client interface for better testability`). Keep commit scope narrow and messages specific.

PRs should include:
- clear problem/solution summary,
- linked issue (if applicable),
- notes on behavior changes and risks,
- evidence of validation (`make lint test` output or equivalent).

If CLI output or UX changes, include sample command/output snippets in the PR description.
