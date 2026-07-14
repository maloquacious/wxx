# Worldographer Projects Inventory

An inventory of Go projects on this machine that read, write, or model Worldographer
WXX map files. Compiled to identify which older attempts hold format insights worth
carrying into a single consolidated project (`github.com/maloquacious/wxx`).

Dates are the last git commit (or go.mod mtime where no git history exists).
"Integration" = whether Worldographer/WXX is the project's core purpose or a side feature.

---

## Tier 1 — Dedicated / independent WXX codecs (the format knowledge)

| Path | Package (go.mod) | Date | Integration |
|------|------------------|------|-------------|
| `~/Software/mdhender/ottomap/wog` | `github.com/mdhender/ottomap` (subpkg `wog`) | 2026-06-06 | **Core** — the codec |
| `~/Jetbrains/worldographer/wxconv` | `github.com/mdhender/wxconv` | 2024-02-11 | Core |
| `~/Jetbrains/worldographer/tnwxx` | `github.com/playbymail/tnwxx` | 2024-01-09 | Core |
| `~/Jetbrains/tcfna` (`internal/wxx`) | `github.com/mdhender/tcfna` | 2024-08-06 | Core (writer) |
| `~/Software/mdhender/gemgem/apps/api` | `github.com/mdhender/gemgem` | 2026-03-16 | Core (renderer) |
| `~/Software/maloquacious/wxx` **(this one)** | `github.com/maloquacious/wxx` | 2026-07-13 | Core — consolidation target |

### `wog` — the single best artifact ⭐
Bidirectional codec (decode **and** encode) split by schema version, format plumbing fully isolated.
- **`wog/pipeline.go`** — cleanest statement of the WXX container: gzip (`0x1f8b`) → UTF-16 with
  BOM detection (BE `FEFF` / LE `FFFE`, UTF-8 BOM, falls back to BE) → XML. Documents that
  Worldographer emits `<?xml version='1.1' encoding='utf-16'?>` with **single quotes** that Go's
  `encoding/xml` rejects, and rewrites the declaration on read/write.
- **`wog/read.go`** — version sniffing: `release="2025"` → V2025; `version="1.x"` + no release → V2017.
- **`wog/v2017/schema.go` + `wog/v2025/schema.go`** — full XML structs, doc comments pinning the exact
  V2017↔V2025 diffs (no `release`, no `maplayer/@opacity`, no `blurTerrainBG`/`extraTerrain` in 2017).
- **`wog/internal/wxxio/tile.go`** — definitive `<tilerow>` spec: tab-separated, field counts **6/7/11/12**,
  the **`"Z"` sentinel** (field 5 = all non-Animal resources zero), optional trailing float-RGBA background.
- **`wxxio/terrain.go`** — `<terrainmap>` is `Name<TAB>Index`; **slot 0 is always "Blank"**, slots interned lazily.
- **`wxxio/pixel.go`** — richest geometry insight: features/labels anchor to an **"ideal" 300×300 hex**
  independent of render size; flat-top columns step 225 (¾), odd columns stagger down 150 (½). Also: the
  grid emits `tilesWide` `<tilerow>`s each holding `tilesHigh` lines (rowIdx is the *column*).
- **`wxxio/color.go`** — float-RGBA `"r,g,b,a"`; both `"null"` and `"0.0,0.0,0.0,1.0"` mean "no color."
- **`wog/FEATURES.md`** — per-feature read/write coverage matrix for the whole V2025 schema.

### `wxconv`
Clean adapters/domains/models refactor of the codec. Best **encode-side** reference
(`adapters/wmap_to_wxx_encoder.go`: re-emitting `"Z"`, trimming `.0` floats, nullable RGBA).
Explicit version dispatch (`WMAPToWXXEncoding`, `ErrUnsupportedVersion`) though only v1.73 is wired up.
Three-layer model split: raw-XML (`wxml173`) vs domain (`wxx`) vs all-string encode model (`tmap173`).

### `tnwxx`
Richest **raw** insight. Ships an actual **RelaxNG schema** of the WXX XML
(`testdata/utf-8-xml.rnc` / `.rng`) — attribute types, color NMTOKENs, enum NCNames. Working
`text/template` encoder (`xml.gohtml`) showing type-dependent `<information>` attributes
(Culture/Nation/Religion) and CDATA inner text. v1.73 only. Note: contains TribeNet-specific
terrain hacks a codec author should ignore.

### `tcfna` (`internal/wxx/`)
Independent **second implementation** with **no ottomap dependency** — confirms the encode format.
Captures a *different* schema/orientation: **`version="1.76"`, `hexOrientation="ROWS"`** (pointy-top),
vs the COLUMNS/flat-top projects. `internal/wxx/points.go` + `hex_vertex_offsets.png` hold the
hex-vertex geometry (flat-top vs pointy vertex offsets, label placement).

### `gemgem` (`internal/render/ottomap/wxx/`)
The evolved ottomap renderer — full **schema-1.74** emission with the deepest **feature / resource /
settlement / edge** modeling (`writer.go` 982 lines, `blank_map.go`, `points.go`, `colors.go`,
`grids.go`, `merge.go`). `cmd/json2wxx` drives it from JSON. Richer than a plain ottomap copy.

---

## Tier 2 — WXX as primary output (rendering conventions, not codec)

| Path | Package (go.mod) | Date | Integration |
|------|------------------|------|-------------|
| `~/Jetbrains/tribenet/ottomap` | `github.com/playbymail/ottomap` | 2026-03-20 | Core (write-only output) |
| `~/Software/mdhender/ottomap` (top level) | `github.com/mdhender/ottomap` | 2026-06-06 | Side feature (map model) |

- **playbymail/ottomap** — the actual TribeNet turn-report mapper; WXX is its output (write-only, no reader).
  Keep for TribeNet **rendering conventions**: grid system (`grids.go`, 30×21, IDs AA–ZZ),
  elevation-by-terrain conventions (Ocean −3, Lake −1, PolarIce 10, land 1250), the **"flattened hexes"**
  vertex computation (`flattened_hexes.go`), and the complete literal **blank-map template**
  (`blank_map.go`: full `<gridandnumbering>`, `<mapkey>`, `<configuration>` blocks that `wog` omits).
  Edge features (Canal/Ford/Pass/River/StoneRoad by direction) in `types.go`.
- **mdhender/ottomap (top level)** — a clean, format-agnostic hex-grid map **model** (axial coords, opaque
  `Terrain` string); WXX lives only in the `wog` subpackage. Keep `types.go` `Resources` struct and `Scope` enum.

---

## Tier 3 — Consumers of the external `wog` engine (coordinate/stagger notes only)

| Path | Package (go.mod) | Date | Integration |
|------|------------------|------|-------------|
| `~/Software/maloquacious/yage-maps` (`cmd/komwxx`) | `github.com/maloquacious/yage-maps` | 2026-06-29 | Core cmd, external engine |
| `~/Software/mdhender/tpty` (`worldographer/`) | `github.com/mdhender/tpty` | 2026-07-13 | Side (export only) |
| `~/Software/mdhender/opyl` (`cmd/g3wxx`) | `github.com/mdhender/opyl` | 2026-06-06 | Side (export only) |

These delegate all encoding to `github.com/mdhender/ottomap/wog`. Value is in their **notes**, not code:
- **yage-maps/komwxx** — only project doing **both directions** (auto-selects by extension, `wog.Read`/`wog.Write`).
  Documents the round-trip translation-table pattern and the pin-bbox-to-origin trick so missing hexes emit
  as Blank and origin-column parity aligns with the COLUMNS stagger.
- **tpty/worldographer.go** — verified-in-app prose on the **COLUMNS stagger gotcha** (`minColParity`/`offsetBounds`):
  how Worldographer staggers by *array position* and how to pad the min column so odd-q parity lands right.
- **opyl/internal/domain/hex.go** — clean doc-cited odd-q offset↔axial conversion with exact Worldographer formulas.

---

## Excluded (no independent Worldographer knowledge)

| Path | Package | Reason |
|------|---------|--------|
| `~/Jetbrains/amp/otto` | `github.com/playbymail/otto` | Consumer — imports `github.com/maloquacious/wxx` (downstream of this project) |
| `~/Jetbrains/worldographer/devorc` | `github.com/mdhender/devor` | Delaunay triangulation port; **zero** WXX code (misfiled). Doesn't compile. |
| `~/Software/mdhender/goober` | `github.com/mdhender/goober` | 17-line doc-store registration of a WXX mime type; no parsing. |
| `~/Jetbrains/tribenet/tn3/ottomap` | `github.com/mdhender/ottomap` | Stale duplicate (2024-06-23) of pre-`internal/` ottomap `wxx/`. |
| `~/Jetbrains/tribenet/tmp/ottomap` | `github.com/mdhender/ottomap` | Stale duplicate (2024-08-08), slightly newer than tn3, still far behind. |

---

## Consolidation takeaways

1. **`wog` is the keeper for format knowledge** — bidirectional, both schema versions, isolated plumbing,
   best-documented container/tile/terrain/pixel semantics, plus `FEATURES.md` coverage matrix.
2. **`wxconv`'s architecture** (adapters + 3-layer models) is the cleanest structural template.
3. **`tnwxx`'s RelaxNG schema** (`utf-8-xml.rnc`/`.rng`) and **`xml.gohtml` template** are unique raw
   references — carry them over as documentation even if the code isn't reused.
4. **`tcfna`** independently confirms the format and uniquely covers **1.76 / ROWS (pointy-top)** —
   valuable for schema coverage breadth (the codec projects are mostly COLUMNS/flat-top).
5. **`gemgem`** has the deepest feature/resource/settlement/edge modeling (schema 1.74).
6. **playbymail/ottomap** holds the TribeNet **blank-map template** and rendering conventions `wog` omits.
7. Coordinate/parity gotchas are best documented in **tpty**, **yage-maps**, and **opyl** notes.
