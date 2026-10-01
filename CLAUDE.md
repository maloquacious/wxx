# CLAUDE.md

This file provides guidance for AI assistants working on the WXX codebase.

## Project Overview

WXX is a Go package (`github.com/maloquacious/wxx`) for reading, writing, and manipulating Worldographer WXX map files. Worldographer is a Java-based map generator that stores data as gzip-compressed, UTF-16 big-endian encoded XML files.

One Worldographer format is supported:
- **W2025** - "Worldographer 2025" (XML 1.1, schema version in `map` element)

The original H2017 ("classic", Worldographer 1.x) format is no longer read or written (issue #103, ADR 0005). The decoder refuses a classic file with `wxx.ErrClassicMap`; Worldographer 2025 converts a classic map itself.

Current version: **0.44.0-alpha** (see `version.go`).

## Bugs Before Features

**Open bugs outrank feature work.** When choosing what to do next, or when asked to start a
feature while bugs labeled `bug` are open, say so and propose the bugs first. Do not begin
feature work on the assumption that the bug list will be dealt with later — that is how the
identity bugs (#28, #32, #41, #45) accumulated behind each other.

This is a rule about **what to work on next**, not a hard interlock on every commit:

- A bug fix, a test that pins a bug, a doc correction, or a refactor in service of a bug fix
  is never blocked.
- Feature work proceeds only when the maintainer says so explicitly, having been told what is
  still open. "Do the feature anyway" is a legitimate answer — the requirement is that it is a
  decision made with the bug list in view, not one made by default.
- Do not use "it is unrelated to the open bugs" as self-authorization. Independence is an
  argument to put to the maintainer, not a reason to skip asking.

When proposing bugs, rank them by whether they can produce a wrong file on disk. A bug that
writes silently-wrong output (e.g. #45) outranks one that loses a field (#35) or misnames a
diagnostic (#44).

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
├── version.go          # Semantic version (0.44.0-alpha)
├── xmlio/              # XML encoding/decoding pipeline
│   ├── decoder.go      # Generic decoder with functional options
│   ├── encoder.go      # Generic encoder with functional options
│   ├── xml_header.go   # XML header utilities
│   └── internal/       # codec packages; unimportable outside xmlio/
│       ├── README.md   # the codec-version convention (what v1_06 means)
│       ├── appver/     # a codec's accepted-application-version declaration
│       ├── codec/      # the Codec interface the dispatcher holds
│       └── v1_06/      # W2025 schema 1.06: per-element decode/encode files
├── cmd/                # CLI tools (each has its own main.go)
│   ├── bounds/         # Extract map dimensions
│   ├── copy/           # Copy WXX files with optional transformations
│   ├── import/         # Import terrain layers (WIP)
│   ├── info/           # Display WXX file information
│   ├── merge/          # Merge multiple maps (WIP)
│   ├── resize/         # Resize/expand/crop maps
│   ├── schema/         # Extract XML schema hierarchy
│   ├── server/         # Web server for hex grid visualization
│   └── version/        # Display package version
├── schema/             # wxx reference grammars, one per schema version
│   ├── README.md       # how to read them, label vocabulary, format notes
│   ├── 1.06.rnc        # schema 1.06 (Worldographer 2025 2.06, 2.07, 2.08)
│   └── grammar_test.go # checks every 2.08 fixture against 1.06.rnc
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

The decoder reads the `<map>` element's `release` attribute to dispatch to the schema-specific decoder:
- `release="2025"` -> `v1_06` (`version` and `schema` do not gate dispatch)
- empty `release` + a `1.x` `version` (a classic map) -> refused with `wxx.ErrClassicMap`
- anything else -> refused with `wxx.ErrUnsupportedMapMetadata`

W2025 support covers 2.06 (the first post-beta build), 2.07 (#92) and 2.08 (#73). All three state `release="2025"` and `schema="1.06"`, so they share the `v1_06` codec and differ only in `map/@version`. 2.08 is the baseline new work targets (#73). Earlier 2025 builds are out of scope.

The **encoder** dispatches the other way round, and the rule is a contract rather than a convenience (issue #41):

- A caller names a target **only** by its verbatim application version (`map/@version`): `xmlio.MarshalXML(m, "2.06")`, `xmlio.NewEncoder("2.06")` or `xmlio.WriteFile(path, m, "2.06")`. `""` is not a sentinel — it names no release and errors.
- The registry (`byApp` in `xmlio/codecs.go`) maps that string to the one codec that accepts it; each codec declares its accepted application versions in its own `apps.go`. An unregistered version is an error (`wxx.ErrUnsupportedMapVersion`).
- **No public symbol accepts a schema version or returns a codec.** The codec packages live under `xmlio/internal/`, unreachable from `cmd/*` and from outside the module.
- The codec writes the identity (`@release`, `@version`, `@schema`) of the application version it was asked for, never the identity the source map states (issue #45; `encodeMap` in `xmlio/internal/v1_06/map.go`). So a file's declared identity and its content format cannot disagree. `xmlio/chimera_test.go` pins this: a W2025 map carrying #41's classic identity (`release="" version="1.77" schema=""`) encodes as a well-formed 2.06 file, through the internal codec and through the public API.

## Coding Conventions

- **Go version**: 1.24.4 (specified in `go.mod`)
- **Copyright header**: Every `.go` file starts with `// Copyright (c) <year> Michael D Henderson. All rights reserved.`
- **Package comments**: Each package has a doc comment on the `package` line
- **Minimal dependencies**: Add no dependency that can be avoided. `go.mod` is the list; this file does not repeat it.
- **Line endings**: LF enforced via `.gitattributes` for all source files
- **Naming**: Types use `_t` suffix for major data types (e.g., `Map_t`). CLI tools are lowercase single-word names.
- **No external test frameworks**: Uses Go standard `testing` package only
- **No CI/CD**: No automated pipelines; test locally with `go test ./...`
- **Pipeline architecture**: Encoding/decoding is done as composable transformation stages, not monolithic functions
- **Diagnostics over debug logging**: Optional `Diagnostics` structs capture intermediate pipeline data instead of using log statements
- **CLI tools in cmd/**: Each tool is a separate `main` package under `cmd/<name>/main.go`, built independently

## Existing Documentation

Read these files for deeper context:
- `AGENTS.md` - High-level project overview, Worldographer background, building instructions
- `CODECS.md` - Guiding principles for codec design, API specifications, implementation patterns
- `PROJECT.md` - Directory structure overview
- `README.md` - Project overview, version mapping table, pipeline documentation
- `schema/README.md` - The wxx reference grammars for the file format, their label vocabulary, and format notes

## Reasoning About the File Format

When a claim is about what a WXX file contains, cite `schema/1.06.rnc` (the pattern
and its `# observed` / `# inferred` label) and check the fixture bytes
(`gunzip -c f.wxx | iconv -f UTF-16BE -t UTF-8`). Do not infer the format from Go
structs: `Map_t` and the codec's schema types show what wxx models, not what the file
holds. `# inferred` and `TBD` rules are hypotheses. See `schema/README.md`.

## Important Notes

- The W2025 decoder (`xmlio/internal/v1_06/`) is incomplete and a work in progress
- The `cmd/import` and `cmd/merge` tools are also WIP
- An Sqlite3 data store is planned for the future (after xmlio codecs are complete)
- `Map_t` models the W2025 format, the only one wxx reads; decoders target it, encoders source from it. It carries no fields for formats wxx does not read (the classic-only ones were removed by #103)
- WXX files are binary (gzip-compressed); every fixture a test reads lives flat in `testdata/` and is tracked, so the suite runs from a clean clone. Do not add a fixture the tests need to `scratch/` — it is git-ignored
- The `dist/` and `scratch/` directories are for local output and are not committed
