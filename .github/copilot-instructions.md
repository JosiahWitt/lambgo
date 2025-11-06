# Lambgo AI Coding Instructions

## Project Overview

Lambgo is a CLI tool for building AWS Lambda functions in Go. It reads `.lambgo.yml` config files and compiles Go binaries for Linux/amd64, zipping them for Lambda deployment. The tool supports parallel builds and custom build flags.

## General Instructions

- This is a CLI, so avoid breaking changes to publicly exposed pieces (eg. commands, `.lambgo.yml`, etc.) unless absolutely necessary.
- Follow Go conventions and idiomatic patterns.
- Prioritize simplicity over complexity.

## Architecture

### Core Components

- **cmd/lambgo**: CLI entry point with dependency injection pattern (see `main.go` for wiring)
- **internal/cmd**: CLI commands using urfave/cli/v2. `App` struct holds all dependencies
- **internal/lambgofile**: Config loader that searches up directories for `go.mod`, then loads `.lambgo.yml`
- **internal/builder**: Orchestrates parallel Lambda builds with Go toolchain
- **internal/runcmd**: Wraps `os/exec` for running `go build` commands
- **internal/zipper**: Creates reproducible zip files (hardcoded 2009-11-10 timestamp)

### Data Flow

1. `Loader.LoadConfig(pwd)` walks up to find `go.mod`, then loads `.lambgo.yml` alongside it
2. `Builder.BuildBinaries()` first builds dependencies (if >1 Lambda) to populate build cache
3. Parallel workers (configurable via `--num-parallel`) build each Lambda: `go build -trimpath` → zip
4. All errors collected via `erk/erg` error group, reported together at end

## Testing & Mocking

### ensure Test Framework

All tests use `github.com/JosiahWitt/ensure`:

```go
ensure := ensure.New(t)
ensure.Run("description", func(ensure ensuring.E) {
    ensure(result).Equals(expected)
})
```

Use `ensure.RunTableByIndex()` for table-driven tests.

### Mock Generation

- Config: `.ensure.yml` defines interfaces to mock
- Generate: `make generate-mocks` or `ensure mocks generate`
- Mocks live in `internal/mocks/` and use `gomock`
- Pattern: Every `*API` interface gets mocked (e.g., `RunnerAPI`, `ZipAPI`)

## Error Handling with erk

Uses `github.com/JosiahWitt/erk` for typed errors with context:

```go
var ErrCannotOpenFile = erk.New(ErkCannotLoadConfig{}, "Cannot open '{{.path}}': {{.err}}")
return erk.WrapWith(ErrCannotOpenFile, err, erk.Params{"path": configFilePath})
```

Multi-errors via `erk/erg`: see `builder.go` lines 22-23, 51-55 for parallel error collection.

## Development Workflow

### Key Commands

- `make test`: Run all tests
- `make test-coverage`: Generate HTML coverage report → `tests/coverage.html`
- `make lint`: Run golangci-lint
- `make generate-mocks`: Regenerate mocks from `.ensure.yml`
- `go install ./cmd/lambgo`: Install CLI locally
- `lambgo build`: Build Lambdas per `.lambgo.yml` (try `examples/simple/`)

### Build System

- Makefile wraps common tasks (see `Makefile` for all targets)
- CI runs tests + lint + regression across Go 1.18-1.25 (see `.github/workflows/ci.yml`)
- Regression test: builds `examples/simple` to verify end-to-end

## Code Conventions

### Dependency Injection

All components use interface dependencies (e.g., `LambdaBuilderAPI`, `LoaderAPI`). Main wires concrete types; tests inject mocks. See `cmd/lambgo/main.go` for the pattern.

### Cross-compilation Defaults

Always set `GOOS=linux` and `GOARCH=amd64` for Lambda compatibility (see `builder.go:166-170`). User can override via `.lambgo.yml`.

### Parallel Execution

- Builder uses worker pool pattern with channels (see `builder.go:68-75`)
- `NumParallel` supports: `all`, `<int>`, or `<float>x` (CPU multiplier)
- Dependency pre-build avoids cache conflicts (see `builder.go:85-107`)

### File Structure

- `internal/`: Private packages (Go module semantics)
- `examples/`: End-to-end examples; `simple/` used in CI regression tests
- No `//go:generate` directives; use `make generate-mocks` instead

## Common Patterns

### Config Loading

`Loader.LoadConfig()` recursively searches parent dirs for `go.mod` (see `lambgofile.go:83-94`). This allows running `lambgo` from subdirectories.

### Reproducible Builds

Zip files use fixed timestamp (2009-11-10) for determinism (see `zipper.go:44`). Combined with `go build -trimpath` for reproducible artifacts.

### CLI Flag Parsing

Use urfave/cli/v2. See `build.go:37-67` for flag definitions. Complex parsing (e.g., `--num-parallel`) has dedicated functions.

## Examples

Run `cd examples/simple && lambgo build` to see end-to-end workflow. Output: `tmp/lambdas/hello_world.zip`.
