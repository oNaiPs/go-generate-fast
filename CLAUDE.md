# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`go-generate-fast` is a drop-in replacement for `go generate` that speeds up code generation through intelligent caching. It intercepts `go:generate` directives, detects input/output files, and reuses cached outputs when inputs haven't changed.

## Build and Test Commands

### Build
```bash
make build
# Output: bin/go-generate-fast
```

### Test
```bash
# Run unit tests
make test

# Run e2e tests
make e2e

# Run linter
make lint

# Fix linting issues
make lint-fix
```

### Running Tests for Specific Packages
```bash
# Run tests in a specific package directory
go test -C src ./path/to/package

# Run a single test file
go test -C src ./path/to/package -run TestName
```

### Install Dependencies
```bash
make install-deps
```

## Architecture

### Core Flow

1. **Entry Point** (`main.go`): Initializes config, logger, and plugin factory, then calls `generate.RunGenerate()`

2. **Generate Engine** (`src/core/generate/generate/generate.go`):
   - Scans Go files for `//go:generate` directives
   - Also processes `//go:generate_input` and `//go:generate_output` for manual input/output specification
   - Parses command arguments and creates `GenerateOpts` struct
   - Matches commands against registered plugins
   - Orchestrates cache verification, restoration, or generation

3. **Plugin System** (`src/plugins/`):
   - Each plugin implements the `Plugin` interface with `Matches()` and `ComputeInputOutputFiles()`
   - Plugins auto-register via `init()` functions imported in `src/plugin_factory/plugin_factory.go`
   - Plugins parse tool-specific flags to determine input/output files

4. **Cache System** (`src/core/cache/`):
   - Creates unique hash from: command directory, command words, input/output files, input file contents, and executable metadata
   - Cache directory structure: `{cacheDir}/{hash[0:1]}/{hash[1:3]}/{hash[3:]}/`
   - Each cached entry contains hashed output files and a `config.json` with metadata
   - Supports cache restoration by copying files from cache to destination with preservation of modification times

### Plugin Architecture

Each plugin in `src/plugins/` follows this pattern:
- Implements `Plugin` interface (`Name()`, `Matches()`, `ComputeInputOutputFiles()`)
- Parses tool-specific flags using Go's `flag` package
- Returns `InputOutputFiles` struct with detected inputs/outputs and optional output glob patterns
- Registers itself via `init()` calling `plugins.RegisterPlugin()`

Supported plugins: stringer, mockgen, moq, protobuf, gqlgen, esc, go-bindata, genny, controller-gen

### Key Data Structures

- `plugins.GenerateOpts`: Contains parsed command info (path, words, executable name/path, go package, sanitized args, extra input/output patterns)
- `plugins.InputOutputFiles`: Lists input files, output files, output glob patterns, and extra hash data
- `cache.VerifyResult`: Contains plugin match, cache hit status, cache directory path, and save capability

## Adding a New Plugin

1. Create new file in `src/plugins/{toolname}/{toolname}.go`
2. Define struct implementing `plugins.Plugin` interface
3. Implement `Name()` to return plugin identifier
4. Implement `Matches()` to identify when this plugin should handle a command (check `ExecutableName` or `GoPackage`)
5. Implement `ComputeInputOutputFiles()` to parse tool flags and return input/output files
6. Add `init()` function that calls `plugins.RegisterPlugin()`
7. Import the new plugin in `src/plugin_factory/plugin_factory.go` with blank import
8. Add e2e test in `e2e/{toolname}/` directory

## Configuration

Environment variables (see `src/core/config/config.go`):
- `GO_GENERATE_FAST_DIR`: Base directory for config and cache
- `GO_GENERATE_FAST_CACHE_DIR`: Cache files location
- `GO_GENERATE_FAST_DEBUG`: Enable debug logging
- `GO_GENERATE_FAST_DISABLE`: Disable caching completely
- `GO_GENERATE_FAST_READ_ONLY`: Use cache but don't create new entries
- `GO_GENERATE_FAST_FORCE_USE_CACHE`: Fail if cache doesn't exist
- `GO_GENERATE_FAST_RECACHE`: Force regenerate and overwrite cache

## Testing Strategy

- Unit tests located alongside source files (`*_test.go`)
- E2E tests in `e2e/` directory, each subdirectory tests a specific tool
- E2E runner script: `e2e/run.sh`
- Test utilities in `src/test/util.go`

## Key Implementation Notes

- The generator changes working directory to the command's directory (`opts.Dir()`) before running plugins, then restores it afterward
- Output glob patterns are resolved AFTER command execution to capture dynamically generated files
- When using `go run [pkg]@version`, only specific versions (not "latest") can be cached since executable hash is needed
- Cache restoration skips files with matching modification times for performance
- Plugin matching happens via `plugins.MatchPlugin()` which iterates registered plugins until one matches
