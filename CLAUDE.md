# CLAUDE.md

This file provides guidance for AI assistants working on the WXX codebase.

## Project Overview

WXX is a Go package (`github.com/maloquacious/wxx`) for reading, writing, and manipulating Worldographer WXX map files. Worldographer is a Java-based map generator that stores data as gzip-compressed, UTF-16 big-endian encoded XML files.

Two Worldographer versions are supported:
- **H2017** - Original "Worldographer / Hexographer 2" (XML 1.0, no schema version in file)
- **W2025** - Newer "Worldographer 2025" (XML 1.1, schema version in `map` element)

Current version: **0.41.0-alpha** (see `version.go`).

## Build & Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Build a specific CLI tool
go build -o dist/local/<tool> ./cmd/<tool>

# Build example
go build -o dist/local/version ./cmd/version

# Run a built tool
dist/local/version

# Format code (standard Go formatting)
go fmt ./...

# Vet code
go vet ./...

# Update Go module dependencies
bash tools/update-mod.sh
```

## Repository Structure

```
wxx/
├── wxx.go              # Core package: Decoder/Encoder interfaces
├── map.go              # Map_t struct (in-memory map representation)
├── errors.go           # Constant error types (type Error string)
├── version.go          # Semantic version (0.41.0-alpha)
├── xmlio/              # XML encoding/decoding pipeline
│   ├── decoder.go      # Generic decoder with functional options
│   ├── encoder.go      # Generic encoder with functional options
│   ├── xml_header.go   # XML header utilities
│   └── internal/       # codec packages; unimportable outside xmlio/
│       ├── v0_77/      # classic schema: decode.go, encode.go, schema.go
│       └── v1_06/      # W2025 schema: decode.go (incomplete)
├── hexg/               # Hexagonal grid coordinate system library
│   ├── cube.go         # Cube coordinates (primary system)
│   ├── offset.go       # Offset coordinates
│   ├── doubled.go      # Doubled coordinates
│   ├── adapters.go     # Coordinate conversions
│   ├── layout.go       # Hex layout management
│   ├── orientation.go  # Hex orientation types
│   ├── point.go        # Point structure
│   └── tribenet.go     # Tribe/settlement network support
├── cmd/                # CLI tools (each has its own main.go)
│   ├── bounds/         # Extract map dimensions
│   ├── copy/           # Copy WXX files with optional transformations
│   ├── crop/           # Crop WXX map edges
│   ├── import/         # Import terrain layers (WIP)
│   ├── info/           # Display WXX file information
│   ├── merge/          # Merge multiple maps (WIP)
│   ├── resize/         # Resize/expand/crop maps
│   ├── schema/         # Extract XML schema hierarchy
│   ├── server/         # Web server for hex grid visualization
│   └── version/        # Display package version
├── testdata/           # Test fixtures, flat in the root; tracked
├── scratch/            # Local scratch: tool output, debug dumps, textures; git-ignored
├── tools/              # Build/utility scripts
└── dist/               # Build output directory
```

## Architecture & Key Patterns

### Data Flow Pipeline

```
WXX File -> Gunzip -> UTF-16/BE to UTF-8 -> Parse XML Header -> Unmarshal XML -> Map_t
Map_t -> Marshal XML -> Insert XML Header -> UTF-8 to UTF-16/BE -> Gzip -> WXX File
```

### Core Interfaces (wxx.go)

```go
type Decoder interface { Decode(io.Reader) (*Map_t, error) }
type Encoder interface { Encode(io.Writer, *Map_t) error }
```

### Functional Options Pattern

Decoders and encoders use functional options for configuration:
```go
decoder := xmlio.NewDecoder(
    xmlio.WithDecoderDiagnostics(&diag),
    xmlio.WithUTF16BEInput(true),
)
```

### Error Handling

Errors are defined as constant string types in `errors.go`:
```go
type Error string
func (e Error) Error() string { return string(e) }
const ErrInvalidXML = Error("invalid xml")
```

Errors are composed using `errors.Join()` to combine context with root causes.

### Version Dispatching

The decoder reads the `<map>` element's `release` attribute to dispatch to the correct schema-specific decoder; `version` and `schema` do not gate dispatch:
- empty `release` + a `1.x` `version` (`1.73`, `1.74`, `1.77`) -> `v0_77`
- `release="2025"` -> `v1_06`

W2025 support is baselined on 2.06 (`release="2025" version="2.06" schema="1.06"`), the first post-beta build; earlier 2025 builds are out of scope.

## Coding Conventions

- **Go version**: 1.24.4 (specified in `go.mod`)
- **Copyright header**: Every `.go` file starts with `// Copyright (c) <year> Michael D Henderson. All rights reserved.`
- **Package comments**: Each package has a doc comment on the `package` line
- **Minimal dependencies**: Only 2 direct deps (`semver`, `golang.org/x/text`). Keep it lean.
- **Line endings**: LF enforced via `.gitattributes` for all source files
- **Naming**: Types use `_t` suffix for major data types (e.g., `Map_t`). CLI tools are lowercase single-word names.
- **No external test frameworks**: Uses Go standard `testing` package only
- **No CI/CD**: No automated pipelines; test locally with `go test ./...`
- **Pipeline architecture**: Encoding/decoding is done as composable transformation stages, not monolithic functions
- **Diagnostics over debug logging**: Optional `Diagnostics` structs capture intermediate pipeline data instead of using log statements
- **CLI tools in cmd/**: Each tool is a separate `main` package under `cmd/<name>/main.go`, built independently

## Existing Documentation

Read these files for deeper context:
- `AGENT.md` - High-level project overview, Worldographer background, building instructions
- `CODECS.md` - Guiding principles for codec design, API specifications, implementation patterns
- `PROJECT.md` - Directory structure overview
- `README.adoc` - Project overview, version mapping table, pipeline documentation
- `hexg/HEXES.md` - Hex coordinate system documentation

## Important Notes

- The W2025 decoder (`xmlio/internal/v1_06/`) is incomplete and a work in progress
- The `cmd/import` and `cmd/merge` tools are also WIP
- An Sqlite3 data store is planned for the future (after xmlio codecs are complete)
- `Map_t` is a superset of all Worldographer schema versions; decoders target it, encoders source from it
- WXX files are binary (gzip-compressed); every fixture a test reads lives flat in `testdata/` and is tracked, so the suite runs from a clean clone. Do not add a fixture the tests need to `scratch/` — it is git-ignored
- The `dist/` and `scratch/` directories are for local output and are not committed
