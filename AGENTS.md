# Agent guide

This project implements a Go package (`github.com/maloquacious/wxx`) to read,
manipulate, and write Worldographer data files (WXX). See
[PROJECT.md](./PROJECT.md) for the directory layout and [CODECS.md](./CODECS.md)
for the codec design that decoders/encoders must follow.

## Worldographer background

Worldographer is a Java map generator. It stores data as XML that is GZip
compressed and UTF-16 big-endian encoded, with a BOM.

Two generations of the program produce WXX files; we name them by year:

1. **H2017** — original "Worldographer" / "Worldographer classic". XML 1.0,
   no schema version attribute on `<map>`.
2. **H2025** — "Worldographer 2025". XML 1.1, schema version stored as an
   attribute of `<map>`.

Track schema differences in package docs as they are discovered; upstream
documentation is sparse.

## Repository layout

- `wxx.go`, `map.go`, `errors.go`, `version.go` — top-level package: the
  `Map_t` superset type, the `Decoder` / `Encoder` interfaces, sentinel
  errors, and `Version()` (semver, currently `0.41.0-alpha`).
- `xmlio/` — XML decode/encode entry points and shared transforms
  (`decoder.go`, `encoder.go`, `xml_header.go`).
  - `xmlio/internal/v0_77/` — H2017 decoder, encoder, and schema types.
  - `xmlio/internal/v1_06/` — H2025 decoder (encoder pending).
- `hexg/` — hex-grid math (cube/offset/doubled coordinates, layouts,
  orientations, TribeNet adapter). See [hexg/HEXES.md](./hexg/HEXES.md).
- `cmd/` — CLI tools used to exercise the package: `bounds`, `copy`, `crop`,
  `import`, `info`, `merge`, `resize`, `schema`, `server`, `version`, and
  the umbrella `wxx` tool (subcommands: `export`).
- `testdata/` — every fixture the test harness reads, flat in the root
  (e.g. `2025-2.06-13x11-941577-blank.wxx`). Tracked, so `go test ./...`
  runs from a clean clone.
- `scratch/` — local scratch: tool output, debug dumps, terrain textures, and inputs
  for the WIP tools. Git-ignored; nothing here is required by a test.
- `tools/` — maintenance scripts (e.g. `update-mod.sh`).

## Codec conventions

- The public surface lives on `wxx.Decoder` / `wxx.Encoder` interfaces in
  [wxx.go](./wxx.go). Version-specific implementations live under `xmlio/<schema>/`.
- Follow [CODECS.md](./CODECS.md): `Decode(io.Reader) (*Map_t, error)` and
  `Encode(io.Writer, *Map_t) error`; expose transforms (gunzip, UTF-16↔UTF-8,
  XML header fix) as composable functions; tune behavior via options.
- `Map_t` is a superset of all known schema versions. Decoders populate it;
  encoders consume it. Never narrow `Map_t` to a single schema.

## CLI conventions

- Use [`github.com/peterbourgon/ff/v4`](https://pkg.go.dev/github.com/peterbourgon/ff/v4)
  for command-line parsing. Do **not** introduce Cobra (`spf13/cobra`) or
  similar frameworks.
- Multi-command tools follow the `ff.Command` pattern: a root `ff.Command`
  with `Subcommands` appended, each subcommand owning its own
  `ff.NewFlagSet(...).SetParent(rootFlags)`. See [cmd/wxx](./cmd/wxx) for
  the reference layout.

## Roadmap

- Finish the H2025 encoder.
- Implement the SQLite3 data store (schema + load/store of `Map_t`) after
  the xmlio decoders and encoders are complete.

## Building and running tools

Binaries go under `dist/local/` (gitignored). One tool per `cmd/` subdir.

```sh
# build
go build -o dist/local/version ./cmd/version
go build -o dist/local/info    ./cmd/info
go build -o dist/local/wxx     ./cmd/wxx

# run
dist/local/version
dist/local/info path/to/file.wxx
dist/local/wxx export --utf-8 out.xml path/to/file.wxx
```

## Validation

- `go build ./...` — compile everything.
- `go test ./...` — run unit tests (notably under `hexg/`).
- `go vet ./...` — sanity check before declaring work done.
