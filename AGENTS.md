# Agent guide

This project implements a Go package (`github.com/maloquacious/wxx`) to read, manipulate, and write Worldographer data files (WXX).
See *Repository layout* below for the directory layout and [CODECS.md](./CODECS.md) for the codec design that decoders/encoders must follow.

Follow [TOOLING.md](./TOOLING.md) for when to build a reusable tool instead of running ad-hoc shell commands.

## Worldographer background

Worldographer is a Java map generator.
It stores data as XML that is GZip compressed and UTF-16 big-endian encoded, with a BOM.

Two generations of the program produce WXX files; we name them by year:

1. **H2017** — original "Worldographer" / "Worldographer classic".
   XML 1.0, no schema version attribute on `<map>`.
   **wxx no longer reads or writes it** (issue #103, [ADR 0005](./docs/adr/0005-remove-classic-format.md)): the decoder refuses a classic file with `wxx.ErrClassicMap`, and Worldographer 2025 converts a classic map itself.
2. **H2025** — "Worldographer 2025".
   XML 1.1, schema version stored as an attribute of `<map>`.
   The only format wxx supports.

Upstream documentation is sparse.
The file format is described by the wxx reference grammars in [`schema/`](./schema/README.md), one per schema version (`schema/1.06.rnc` is current).
Record format facts there, each labelled observed or inferred, and cite the grammar and check fixture bytes when reasoning about the format; do not infer it from Go structs.

When a claim is about what a WXX file contains, cite `schema/1.06.rnc` (the pattern and its `# observed` / `# inferred` label) and check the fixture bytes (`gunzip -c f.wxx | iconv -f UTF-16BE -t UTF-8`).
`Map_t` and the codec's schema types show what wxx models, not what the file holds.
`# inferred` and `TBD` rules are hypotheses.
See [schema/README.md](./schema/README.md).

## Bugs before features

**Open bugs outrank feature work.** When choosing what to do next, or when asked to start a feature while bugs labeled `bug` are open, say so and propose the bugs first.
Do not begin feature work on the assumption that the bug list will be dealt with later — that is how the identity bugs (#28, #32, #41, #45) accumulated behind each other.

This is a rule about **what to work on next**, not a hard interlock on every commit:

- A bug fix, a test that pins a bug, a doc correction, or a refactor in service of a bug fix is never blocked.
- Feature work proceeds only when the maintainer says so explicitly, having been told what is still open.
  "Do the feature anyway" is a legitimate answer — the requirement is that it is a decision made with the bug list in view, not one made by default.
- Do not use "it is unrelated to the open bugs" as self-authorization.
  Independence is an argument to put to the maintainer, not a reason to skip asking.

When proposing bugs, rank them by whether they can produce a wrong file on disk.
A bug that writes silently-wrong output (e.g. #45) outranks one that loses a field (#35) or misnames a diagnostic (#44).

## Tag every version change

When a merge to `main` changes `version.go`, tag that commit `v<version>` (lightweight) and push the tag.
See [CONTRIBUTING.md](./CONTRIBUTING.md) step 7.

## Repository layout

- `wxx.go`, `map.go`, `errors.go`, `version.go` — top-level package: the `Map_t` type, the `Decoder` / `Encoder` interfaces, sentinel errors, and `Version()` (semver, currently `0.54.0-beta`).
- `xmlio/` — XML decode/encode entry points and shared transforms (`decoder.go`, `encoder.go`, `xml_header.go`).
  - `xmlio/internal/v1_06/` — H2025 (schema 1.06) decoder, encoder, and schema types; the only codec.
- Hex-grid math (cube/offset coordinates, layouts) is the external module `github.com/maloquacious/hexg`; `Tile_t.Coords` is a `hexg.Hex`, and `Map_t.GridOrientation()` derives the offset layout from `HexOrientation`.
- `cmd/` — CLI tools used to exercise the package: `bounds`, `copy`, `import`, `info`, `merge`, `resize`, `schema`, `version`, and the umbrella `wxx` tool (subcommands: `export`).
- `schema/` — the wxx reference grammars, one per schema version (`1.06.rnc`), and a test that checks every 2.08 fixture against the grammar.
  See [schema/README.md](./schema/README.md).
- `testdata/` — every fixture the test harness reads, flat in the root (e.g. `2025-2.06-13x11-941577-blank.wxx`).
  Tracked, so `go test ./...` runs from a clean clone.
- `scratch/` — local scratch: tool output, debug dumps, terrain textures, and inputs for the WIP tools.
  Git-ignored; nothing here is required by a test.
- `tools/` — maintenance scripts (e.g. `update-mod.sh`).

## Codec conventions

- The public surface lives on `wxx.Decoder` / `wxx.Encoder` interfaces in [wxx.go](./wxx.go).
  Version-specific implementations live under `xmlio/internal/<codec version>/` and are **not** publicly importable.
- Follow [CODECS.md](./CODECS.md): `Decode(io.Reader) (*Map_t, error)` and `Encode(io.Writer, *Map_t) error`; expose transforms (gunzip, UTF-16↔UTF-8, XML header fix) as composable functions; tune behavior via options.
- `Map_t` models the H2025 format, the only one wxx reads.
  Decoders populate it; encoders consume it.
  It carries no fields for formats wxx does not read.
- **An encoder takes an application version, never a schema version, and hands out no codec** (issue #41).
  A caller names a target only by its verbatim `map/@version` string — `xmlio.MarshalXML(m, "2.06")` or `xmlio.NewEncoder("2.06")` — and the registry in `xmlio/codecs.go` resolves it to the one codec that accepts it, which writes that version's identity.
  There is no public way to name a schema or hold an encoder: the codecs and the `codec.Codec` interface live under `xmlio/internal/`.
  - Do not add a public symbol that accepts a schema or returns a codec.
    Tests that legitimately choose an encoder import `xmlio/internal/...` directly — `package xmlio_test` lives inside `xmlio/`, and Go's internal rule is directory-based, so that works with no escape hatch in the package.

## CLI conventions

- Use [`github.com/peterbourgon/ff/v4`](https://pkg.go.dev/github.com/peterbourgon/ff/v4) for command-line parsing.
  Do **not** introduce Cobra (`spf13/cobra`) or similar frameworks.
- Multi-command tools follow the `ff.Command` pattern: a root `ff.Command` with `Subcommands` appended, each subcommand owning its own `ff.NewFlagSet(...).SetParent(rootFlags)`.
  See [cmd/wxx](./cmd/wxx) for the reference layout.
- This is transitional.
  Under [ADR 0007](./docs/adr/0007-starlark-script-host.md), the separate `cmd/*` binaries fold into `wxx`, which becomes a Starlark script host (`wxx script`, `wxx shell`), and `ff/v4` is removed.
  Until that lands, new subcommands still follow the `ff.Command` pattern; do not add Starlark or remove `ff/v4` as part of unrelated work.

## Coding conventions

- **Copyright header:** every `.go` file starts with `// Copyright (c) <year> Michael D Henderson. All rights reserved.`
- **Package comments:** each package has a doc comment on the `package` line.
- **Minimal dependencies:** add no dependency that can be avoided.
  `go.mod` is the list; this file does not repeat it.
- **Naming:** major data types take a `_t` suffix (e.g. `Map_t`).
  CLI tools are lowercase single-word names.
- **Tests:** the standard `testing` package only; no external test frameworks.
- **No CI pipelines:** "CI" means running the validation commands below locally.
- **Pipeline architecture:** encoding and decoding are composable transformation stages, not monolithic functions.
- **Diagnostics over debug logging:** optional `Diagnostics` structs capture intermediate pipeline data instead of log statements.
- **Line endings:** LF, enforced by `.gitattributes`.

## Roadmap

- Finish the H2025 encoder.
- Implement the SQLite3 data store (schema + load/store of `Map_t`) after the xmlio decoders and encoders are complete.

## Building and running tools

Binaries go under `dist/local/` (gitignored).
One tool per `cmd/` subdir.

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
- `go test ./...` — run unit tests.
- `go vet ./...` — sanity check before declaring work done.
