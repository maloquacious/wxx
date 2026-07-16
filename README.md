# wxx

`wxx` is a Go package and command-line toolkit for reading, writing, inspecting, and modifying [Worldographer](https://worldographer.com/) files.

The name comes from `.wxx`, Worldographer’s default file extension.

> **This README marks planned work as planned.** Anything under a *(planned)*
> heading or in a **Planned** list does not exist yet. Everything else describes
> the tree you are looking at.

**What exists today:**

* A Go API for working with Worldographer data
* Reading and writing `.wxx` files
* Inspecting maps, and modifying them (crop, resize, copy)
* A `wxx export` subcommand, plus a set of separate single-purpose binaries

**Planned — not built yet:**

* Folding the separate binaries into `wxx` and retiring them
* Turning `wxx` into a Lua script host using [GopherLua](https://github.com/yuin/gopher-lua)
* Interactive scripting against a loaded map
* Validating maps ([#20](https://github.com/maloquacious/wxx/issues/20))

## Worldographer versions

Worldographer has two file-format families, and `wxx` reads and writes both:

* **Classic** — the original Worldographer / Hexographer 2 format (XML 1.0, no schema version in the file).
* **2025** — the newer Worldographer 2025 format (XML 1.1, with a schema version in the `map` element).

**Classic support is frozen.** It will continue to read and write existing files and will receive **security bug fixes only** — no new features. **Future development focuses on the 2025 version.** (One narrow exception is in progress: reconciling classic version *identity* metadata — see the note below and `docs/adr/0004-version-struct-and-release-registry.md`.)

### Version identity

Every decoded map carries `MetaData.Version`, the two independent version axes a
file states — and nothing else (ADR 0004). The verbatim on-disk strings are also
kept in `MetaData.Worldographer.*`.

```go
type Version_t struct {
	App    Dotted  // map/@version -- the build that wrote the file
	Schema *Dotted // map/@schema  -- nil when the file states none
}
```

| release | `map/@release` | `map/@version` (App) | `map/@schema` (Schema) |
|---|---|---|---|
| classic (Hexographer 2) | *absent* | `1.73`, `1.74`, `1.77` | *absent* → `nil` |
| Worldographer 2025 | `2025` | `2.06` | `1.06` |

A `nil` Schema is meaningful: it identifies the one **implicit legacy** schema
that classic `1.73`/`1.74`/`1.77` share, rather than an unknown one.

**The schema selects the codec**; the application version is caller-chosen data.
Two application versions sharing a schema use one codec and differ only in the
string written to `@version` — which is why classic `1.73`, `1.74` and `1.77` all
run through one codec. Which releases are supported, and the full on-disk identity
of each, is the [release registry](xmlio/registry.go).

Callers name an **application version**, never a schema and never a codec:
`xmlio.MarshalXML(m, "2.06")`. The codecs themselves live under
`xmlio/internal/`, where an external caller cannot reach them; each declares the
application versions it accepts, and
[`xmlio/internal/README.md`](xmlio/internal/README.md) explains why the packages
are named `v0_77` and `v1_06`. Adding support for a further application version
means an entry in that codec's `apps.go` **and** one in the release registry.

These on-disk values are **not** semantic versions: `"2.06"` through a `semver`
round-trip comes back as `"2.6"`, a different string and therefore a different
file. They are modeled as `Dotted`, whose `Raw` is authoritative for output and
whose components exist only to compare. `wxx` re-emits **verbatim** whatever it
read; nothing is ever re-rendered from components. That — not any defect in the
numbers — is why they are treated as opaque identifiers.

An earlier claim that the 2025 numbers are *buggy* (`1.x` written where `2.x` was
intended) is **retired**: it does not survive the two-axis model, and
[#13](https://github.com/maloquacious/wxx/issues/13) was closed as not planned
for that reason. The claim read `schema="1.06"` as a mis-numbered *version*,
which is the single-slot conflation ADR 0002 was built on and ADR 0004
superseded. Separate the axes and both numbers are correct: `@version="2.06"` is
the application version and *is* `2.x`; `@schema="1.06"` is the schema version, a
different axis whose numbering starts at 1 by design.

See `docs/adr/0003-version-axes.md` for the two-axis model and
`docs/adr/0004-version-struct-and-release-registry.md` for the identity and
registry design. ADR 0002, which modeled identity as a single semver keyed by a
family year, is **superseded**.

## Go package

The `wxx` package can be imported into other Go applications that need to work with Worldographer files.

The root `wxx` package holds the data types (`wxx.Map_t` and friends); the
file-level read/write helpers live in the `xmlio` package (`xmlio` imports
`wxx`, so the helpers cannot live in `wxx` without an import cycle).

```go
package main

import (
	"fmt"
	"log"

	"github.com/maloquacious/wxx/xmlio"
)

func main() {
	world, err := xmlio.ReadFile("world.wxx")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(world)

	// Write it back out (defaults to the release the map itself states):
	if err := xmlio.WriteFile("world.wxx", world); err != nil {
		log.Fatal(err)
	}
}
```

The public API supports parsing, inspecting, modifying and writing Worldographer
data without the command-line tool. Validation is *(planned)*: `Map_t` has no
`Validate()` today — see [#20](https://github.com/maloquacious/wxx/issues/20).

The exact package path and API are still evolving, and the API has already made
breaking changes at `0.x`.

## Command-line tool

Today, the `wxx` command has exactly **one** subcommand:

```console
wxx export world.wxx
```

Everything else is a **separate binary**, built individually:

```console
go build -o dist/local/info ./cmd/info
dist/local/info world.wxx
```

The full set is `bounds`, `copy`, `crop`, `import`, `info`, `merge`, `resize`,
`schema`, `server` and `version`. `import` and `merge` are works in progress.

### Where this is going *(planned)*

Those separate binaries fold into `wxx` and are retired, and `wxx` becomes a Lua
script host rather than a subcommand tree. That also drops the `ff/v4`
dependency, which exists only to parse subcommands.

So the current subcommand surface is transitional and not worth learning.

## Lua scripting *(planned)*

**Not implemented.** GopherLua is not a dependency, and there is no `script`,
`run` or `shell` subcommand.

The intent is for `wxx` to embed Lua using
[GopherLua](https://github.com/yuin/gopher-lua) so that scripts can inspect and
modify Worldographer data, with `wxx` acting as the script host. An interactive
shell against a loaded map is part of the same idea. This keeps `wxx` as the name
of the overall project while treating scripting and the shell as parts of a
broader toolkit.

The API a script would use to discover which application versions it may write is
being designed in [#46](https://github.com/maloquacious/wxx/issues/46).

## Why `wxx`?

The name directly reflects the file format the project works with. It is short, distinctive, and suitable for both a Go package and a command-line executable.

Other names considered included `wsh`, short for “Worldographer shell.” However, `wsh` is already strongly associated with Microsoft Windows Script Host, and it describes only the scripting portion of the project.

Using `wxx` leaves room for the project to cover the full Worldographer workflow:

* Use the Go package in other applications
* Inspect or convert files from the command line
* Automate changes with Lua scripts
* Explore map data through an interactive shell

