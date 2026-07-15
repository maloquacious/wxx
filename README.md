# wxx

`wxx` is a Go package and command-line toolkit for reading, writing, inspecting, modifying, and scripting [Worldographer](https://worldographer.com/) files.

The name comes from `.wxx`, Worldographer’s default file extension.

The project provides:

* A Go API for working with Worldographer data
* Reading and writing `.wxx` files
* Inspecting and validating maps
* Modifying and generating map content
* Automating map-processing workflows
* Running embedded Lua scripts with [GopherLua](https://github.com/yuin/gopher-lua)
* Interactively scripting against loaded maps

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
string written to `@version`. Which releases are supported, and the full on-disk
identity of each, is the [release registry](xmlio/registry.go) — adding a release
is an entry there.

These on-disk values are **not** semantic versions: `"2.06"` through a `semver`
round-trip comes back as `"2.6"`, a different string and therefore a different
file. They are modeled as `Dotted`, whose `Raw` is authoritative for output and
whose components exist only to compare. `wxx` re-emits **verbatim** whatever it
read; nothing is ever re-rendered from components. The 2025 numbers are also
known to be buggy (`1.x` where `2.x` was intended), which is why they are treated
as opaque identifiers; reconciling them to the true release version is tracked in
#13.

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

The public API is intended to support parsing, inspecting, modifying, validating, and writing Worldographer data without requiring the command-line tool.

The exact package path and API are still evolving.

## Command-line tool

The `wxx` command exposes the package’s functionality through a subcommand-based interface.

```console
wxx info world.wxx
wxx validate world.wxx
wxx run script.lua world.wxx
wxx shell world.wxx
```

The exact commands and syntax are still evolving.

## Lua scripting

`wxx` embeds Lua using GopherLua, allowing scripts to inspect and modify Worldographer data.

```console
wxx run generate.lua world.wxx
```

Scripts can use the Lua API exposed by `wxx` to read map data, make changes, and write the resulting file.

An interactive shell can also be opened for a map:

```console
wxx shell world.wxx
```

This keeps `wxx` as the name of the overall project while treating scripting and the interactive shell as parts of a broader toolkit.

## Why `wxx`?

The name directly reflects the file format the project works with. It is short, distinctive, and suitable for both a Go package and a command-line executable.

Other names considered included `wsh`, short for “Worldographer shell.” However, `wsh` is already strongly associated with Microsoft Windows Script Host, and it describes only the scripting portion of the project.

Using `wxx` leaves room for the project to cover the full Worldographer workflow:

* Use the Go package in other applications
* Inspect or convert files from the command line
* Automate changes with Lua scripts
* Explore map data through an interactive shell

