# ADR 0007 — `wxx script` embeds Starlark, not GopherLua or a custom DSL

- **Status:** **Accepted (2026-11-18)**: maintainer decision.
- **Date:** 2026-11-18
- **Context tickets:** [#46](https://github.com/maloquacious/wxx/issues/46)
  (the discovery API the script host needs),
  [#41](https://github.com/maloquacious/wxx/issues/41) (a caller names an
  application version, never a codec),
  [#45](https://github.com/maloquacious/wxx/issues/45) (the encoder owns every
  identity byte it writes), [#103](https://github.com/maloquacious/wxx/issues/103)
  (classic removed, see ADR 0005).
- **Affects:** the *Lua scripting (planned)* section of `README.md`, which
  names GopherLua; ADR 0006, whose reasons include staying on Go 1.24.4.

---

## Context

The plan is for `wxx` to become a script host: the separate `cmd/*` binaries
fold into it, `wxx script <path>` runs a script against the API, and
`wxx shell` is an interactive REPL against the same API. The `ff/v4`
dependency, which exists only to parse subcommands, goes when that lands. The
README and #46 have assumed the embedded language is Lua, through
[GopherLua](https://github.com/yuin/gopher-lua). Nothing is implemented yet,
so the choice is still free.

A script does three things: reads a map, changes it, and writes it as a named
application version. The last step carries the rules the codec work has paid
for:

- **A script names an application version, never a codec** (#41 requirement
  5, #46 point 4). `v1_06` must be unreachable from a script.
- **Versions are verbatim strings, not numbers** (ADR 0004 Decision 1, #46
  point 6). `"2.06"` is not `"2.6"`, and `2.10` is not `"2.10"`.
- **The write target comes from the caller, never from the map** (#45). A map
  read from a 2.06 file is written as 2.08 because the script said so.

Either way, the REPL needs a line editor (history, cursor movement). Both
`go.starlark.net/repl` and GopherLua's own `cmd/glua` use
`github.com/chzyer/readline`, so readline is not a cost peculiar to one
option. The difference is that Starlark ships the REPL as a library and
GopherLua does not.

## Options considered

### A. GopherLua (Lua 5.1)

- **For:** mature and small. `github.com/yuin/gopher-lua` v1.1.2 declares Go
  1.23 and adds no compiled dependency (its `chzyer/readline` requirement
  serves only `cmd/glua`). Lua is familiar to game and map scripters.
  Userdata and metatables bind Go types, coroutines and `pcall` are there,
  and everything is mutable.
- **Against:**
  - **Strings coerce to numbers and back.** `"2.06" + 0` is `2.06`, and a
    script that writes `app = 2.10` gets `"2.1"` from `tostring`. That is the
    exact mistake ADR 0004 Decision 1 exists to prevent, and the language
    makes it easy.
  - **Arrays are 1-based.** `Tiles_t.Tiles` is `[col][row]` from 0, and so is
    every coordinate wxx prints. Each binding has to choose to translate or
    not, and either choice will be wrong in somebody's script.
  - **Every number is a float64** (Lua 5.1 has no integer type).
  - **Not sandboxed by default.** `os` and `io` are open unless the host
    loads libraries selectively, so a script can do anything the user can.
  - **No REPL library.** We would write the loop ourselves (about 50–100
    lines: compile each line, and keep reading while the error ends in
    `<eof>`) and choose a line editor separately.
  - **Binding is by hand,** or through `layeh.com/gopher-luar` (reflection),
    which is another dependency and would expose Go field names and every
    exported field, provenance included.

### B. Starlark (`go.starlark.net`)

- **For:**
  - **No implicit conversions.** Passing `2.10` where a string is wanted is a
    type error. Verified with the sketch below:
    `map.write: for parameter "app": got float, want string`.
  - **0-based indexing and Python syntax.** Scripts address tiles the way the
    Go API and the CLI tools do.
  - **Hermetic.** A script has no file, OS or network access. Its only I/O is
    what the host provides (here, `wxx.read` and `map.write`), and `load()`
    is resolved by the host.
  - **Exact integers.**
  - **The REPL is a library.** `repl.REPLOptions(opts, thread, globals)` gives
    line editing, history, multi-line blocks and expression echo.
  - **A written spec,** good error messages with backtraces
    (`EvalError.Backtrace`), and active maintenance by Google, which uses it
    for Bazel and other tools.
- **Against:**
  - **More binding code than reflection.** Each exposed Go type implements
    `starlark.Value` plus `HasAttrs` / `HasSetField` / `HasSetIndex` by hand.
    The sketch below binds `Map_t` and `Tile_t` (tiles only) in about 220 lines. Hand
    binding is also the point: it decides exactly what a script can see.
  - **A deliberately small language.** `while`, recursion, top-level `if` /
    `for` and global reassignment are off unless the host turns them on in
    `syntax.FileOptions` (we would turn on all but recursion). There are no
    exceptions: `fail()` aborts. Top-level values are frozen when a script
    file finishes executing, but REPL chunks are not.
  - **No tagged releases.** `go.starlark.net` publishes only pseudo-versions
    (latest: `v0.0.0-20261021141710-4f4267fb140d`), so we pin a commit.
  - **Go 1.25.** That commit declares `go 1.25.0`. `go.mod` would move off
    1.24.4, which ADR 0006 gave as one reason to reject
    `klauspost/compress`. I have not checked whether an older commit still
    builds with 1.24.
  - **Transitive modules.** `go.starlark.net` requires
    `chzyer/readline` (compiled in only through `repl`), `x/sys`, `x/term`,
    `go-cmp` and `protobuf`. `protobuf` serves only `lib/proto`, which we
    would not import, but it enters `go.sum`.
  - **The README and #46 say Lua.** Both need rewording.

### C. A custom DSL

A wxx-specific language: either a command language
(`set tile 3,4 terrain Forest`, `write out.wxx as 2.08`) or a small
expression language with variables and loops.

- **For:** no dependency at all. The syntax fits the domain exactly. Version
  strings, tile coordinates and terrain names can be first-class tokens, so
  the #41 and verbatim-version rules can be enforced by the grammar. Nothing
  inherited needs sandboxing.
- **Against:**
  - **We write and maintain a language:** lexer, parser, evaluator, error
    positions and messages, a REPL with continuation lines, documentation
    and tests. None of that is the project's subject.
  - **It grows.** A command language is fine until the first script needs
    "for every tile where …". Then it needs variables, conditions, loops and
    functions, and becomes a general-purpose language designed in a hurry.
    Starlark is that language, already designed.
  - **No ecosystem.** No editor support or syntax highlighting, and no
    knowledge users bring with them.
  - **It still needs a line editor for the REPL,** the same as A.

## Decision

**Option B.** `wxx script <path>` and `wxx shell` embed Starlark through
`go.starlark.net`, with its `repl` package for the shell.

1. **The script API is a hand-written binding,** in one package under
   `cmd/wxx`. It is not reflection over `wxx.Map_t`. A script sees only the
   names the binding defines, written in snake_case, never a Go field name
   and never a codec.
2. **Tiles are addressed `(column, row)` from 0,** the same as
   `Tiles_t.Tiles`.
3. **Terrain is a name in scripts, never a `TerrainMap` index.**
4. **Writing needs an explicit application version string:**
   `m.write(path, app=...)`. There is no default. `wxx.current_app()`
   returns the newest version for a script that wants it.
5. **Discovery is `wxx.versions()`:** a frozen list of frozen
   `struct(app=..., schema=...)`, oldest first, holding verbatim strings
   (`Dotted.Raw`). It is backed by a new `xmlio.SupportedVersions()
   []wxx.Version_t`, built from the codecs' own declarations
   (`codec.AcceptedApps()`), the same source the registry is built from. This
   answers #46 questions 1 to 3: the existing tuple type, no new descriptor,
   and no codec field to leak.
6. **File options:** `TopLevelControl`, `GlobalReassign`, `While` and `Set`
   are on, and `Recursion` is off. Scripts are programs, not configuration
   files.
7. **The Go bump is accepted** as the cost of this decision. Starlark needs
   at least Go 1.25; the maintainer expects to move `go.mod` to Go 1.27 when
   this is implemented. That would also lift the `go.mod` reason in ADR 0006,
   which stays decided on its other grounds.

## Sketch

This is not the final code. It was compiled and run against `go.starlark.net`
`v0.0.0-20261021141710-4f4267fb140d` and the
`2025-2.08-13x11-941577-blank.wxx` fixture: read, edit 13 tiles, write as
2.08, read back. The failure cases listed after the script were run too.

### Discovery, in `xmlio`

```go
// SupportedVersions returns every application version the registry can write,
// paired with the schema that version's files state, oldest first.
//
// It is derived from the codecs' own declarations, the same source init builds
// the registry from, so it cannot drift from what an encode accepts. It names no
// codec: wxx.Version_t has no field for one (issue #41 requirement 5).
func SupportedVersions() []wxx.Version_t {
	var out []wxx.Version_t
	for _, c := range codecs() {
		set := c.AcceptedApps()
		for _, a := range set.Apps {
			app, _ := wxx.ParseDotted(a.Version) // init has proved every one parses
			v := wxx.Version_t{App: app}
			if set.Schema != "" {
				schema, _ := wxx.ParseDotted(set.Schema)
				v.Schema = &schema
			}
			out = append(out, v)
		}
	}
	slices.SortFunc(out, func(a, b wxx.Version_t) int {
		n, _ := a.App.Compare(b.App) // init has proved they order
		return n
	})
	return out
}
```

### The `wxx` module

```go
func Module() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "wxx",
		Members: starlark.StringDict{
			"read":        starlark.NewBuiltin("wxx.read", read),
			"versions":    starlark.NewBuiltin("wxx.versions", versions),
			"current_app": starlark.NewBuiltin("wxx.current_app", currentApp),
		},
	}
}

// read is not "load": load is a Starlark keyword.
func read(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	m, err := xmlio.ReadFile(path) // a classic file fails with wxx.ErrClassicMap
	if err != nil {
		return nil, err
	}
	return &Map{m: m}, nil
}

func versions(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
		return nil, err
	}
	var out []starlark.Value
	for _, v := range xmlio.SupportedVersions() {
		out = append(out, versionValue(v))
	}
	list := starlark.NewList(out)
	list.Freeze()
	return list, nil
}

// versionValue renders the verbatim strings. An absent schema would be None;
// no supported version has one since #103.
func versionValue(v wxx.Version_t) starlark.Value {
	var schema starlark.Value = starlark.None
	if v.Schema != nil {
		schema = starlark.String(v.Schema.Raw)
	}
	s := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"app":    starlark.String(v.App.Raw),
		"schema": schema,
	})
	s.Freeze()
	return s
}
```

### `Map_t` and `Tile_t`

```go
type Map struct {
	m      *wxx.Map_t
	frozen bool
}

func (v *Map) Type() string          { return "wxx.map" }
func (v *Map) Freeze()               { v.frozen = true }
func (v *Map) Truth() starlark.Bool  { return starlark.True }
func (v *Map) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable: %s", v.Type()) }
// String, columns() and rows() (len of Tiles_t.Tiles and its first column) omitted.

func (v *Map) AttrNames() []string {
	return []string{"columns", "rows", "orientation", "source", "terrains", "tile", "write"}
}

func (v *Map) Attr(name string) (starlark.Value, error) {
	switch name {
	case "columns":
		return starlark.MakeInt(v.columns()), nil
	case "rows":
		return starlark.MakeInt(v.rows()), nil
	case "orientation":
		return starlark.String(v.m.HexOrientation), nil
	case "source": // what wrote the file: provenance, never the write target
		return versionValue(v.m.MetaData.Version), nil
	case "terrains": // names in the map's terrain table
		var names []starlark.Value
		for name := range v.m.TerrainMap.Data {
			names = append(names, starlark.String(name))
		}
		return starlark.NewList(names), nil
	case "tile":
		return starlark.NewBuiltin("map.tile", v.tile), nil
	case "write":
		return starlark.NewBuiltin("map.write", v.write), nil
	}
	return nil, nil // Starlark reports "wxx.map has no .name field"
}

func (v *Map) tile(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var col, row int
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 2, &col, &row); err != nil {
		return nil, err
	}
	if col < 0 || col >= v.columns() || row < 0 || row >= v.rows() {
		return nil, fmt.Errorf("%s: (%d, %d) is off the map", b.Name(), col, row)
	}
	return &Tile{owner: v, t: v.m.Tiles.Tiles[col][row]}, nil
}

// write takes the target from the script, never from the map (#45).
func (v *Map) write(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path, app string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path, "app", &app); err != nil {
		return nil, err
	}
	return starlark.None, xmlio.WriteFile(path, v.m, app)
}

type Tile struct {
	owner *Map // for the terrain table and the frozen flag
	t     *wxx.Tile_t
}

func (v *Tile) AttrNames() []string {
	return []string{"column", "row", "terrain", "elevation", "is_icy", "is_gm_only"}
}

// Attr returns column and row (read-only), terrain as a name (reverse lookup
// in TerrainMap.Data), elevation as a float, and the two flags as bools.

func (v *Tile) SetField(name string, val starlark.Value) error {
	if v.owner.frozen {
		return fmt.Errorf("cannot set .%s: map is frozen", name)
	}
	switch name {
	case "terrain":
		s, ok := starlark.AsString(val)
		if !ok {
			return fmt.Errorf("terrain: got %s, want string", val.Type())
		}
		index, ok := v.owner.m.TerrainMap.Data[s]
		if !ok { // a typo must not add a terrain; see Open questions
			return fmt.Errorf("terrain %q is not in this map's terrain table", s)
		}
		v.t.Terrain = index
	case "elevation":
		f, ok := starlark.AsFloat(val)
		if !ok {
			return fmt.Errorf("elevation: got %s, want number", val.Type())
		}
		v.t.Elevation = f
	// is_icy and is_gm_only: accept starlark.Bool only.
	default:
		return starlark.NoSuchAttrError(fmt.Sprintf("wxx.tile has no settable field .%s", name))
	}
	return nil
}
```

### The host

```go
opts := &syntax.FileOptions{TopLevelControl: true, GlobalReassign: true, While: true, Set: true}
thread := &starlark.Thread{Name: "wxx"}
predeclared := starlark.StringDict{"wxx": script.Module()}

// wxx shell
repl.REPLOptions(opts, thread, predeclared)

// wxx script <path>
_, err := starlark.ExecFileOptions(opts, thread, path, nil, predeclared)
```

### A script

```python
print([(v.app, v.schema) for v in wxx.versions()], wxx.current_app())
# [("2.06", "1.06"), ("2.07", "1.06"), ("2.08", "1.06")] 2.08

m = wxx.read("testdata/2025-2.08-13x11-941577-blank.wxx")
for col in range(m.columns):
    t = m.tile(col, 0)
    t.elevation = 100
m.write("out.wxx", app=wxx.current_app())
```

Errors observed from the same build:

| script | result |
|---|---|
| `m.write("x.wxx", app=2.10)` | `map.write: for parameter "app": got float, want string` |
| `m.write("x.wxx", app="2.6")` | `version "2.6": not a supported application version` |
| `m.tile(3, 0).terrain = "Forrest"` | `terrain "Forrest" is not in this map's terrain table` |
| `m.tile(99, 0)` | `map.tile: (99, 0) is off the map` |

## Consequences

- `go.mod` gains `go.starlark.net` (pinned to a pseudo-version) and its
  transitive modules, and moves to Go 1.25 or later (1.27 expected). It loses `ff/v4` when the
  subcommand tree is retired.
- `xmlio` gains one public function, `SupportedVersions()`. It returns the
  existing `wxx.Version_t` and exposes no codec, so #41 holds.
- The binding package becomes the script API's single source of truth. Every
  field a script can read or write is listed there, so a script reference can
  be written straight from it.
- `README.md` stops naming Lua and GopherLua. The scripting section becomes
  *Starlark scripting (planned)*, and the "Automate changes with Lua scripts"
  bullet changes to match.
- `CLAUDE.md`'s "only 2 direct deps" statement, already stale (see #46), is
  corrected when the dependency lands.

## Open questions

- **Unknown terrain names.** The sketch rejects a name missing from the map's
  terrain table, so a typo cannot add an entry. Worldographer lists only the
  terrains a map uses, though, so painting a new terrain needs a deliberate
  `m.add_terrain(name)` built on `TerrainMap_t.EnsureTerrain`. Without a list
  of Worldographer's terrain names, wxx cannot tell a new name from a typo.
- **Frozen maps.** `ExecFile` freezes a script's globals when it finishes, so
  a host that wants to keep using a map after the script ends must not treat
  `frozen` as meaningful in Go. REPL chunks are never frozen.
- **`read` and `write` reach any path the user can.** That is the script's
  only I/O, and it is the user's own script, but a `-allow-write` style flag
  is possible if that changes.
- **Coverage of `Map_t`.** The sketch binds tiles only. Features, labels,
  shapes, notes and layers each need a binding type. They should be added as
  scripts need them, not up front.

## Revisit when

- `go.starlark.net` stops being maintained, or a Go release it needs is one
  the project will not take, or
- scripts routinely need something Starlark's restrictions forbid and no
  binding can provide.

Reopening this decision means a superseding ADR, not an edit to this one.
