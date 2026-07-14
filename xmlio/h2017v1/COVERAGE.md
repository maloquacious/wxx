# Classic H2017 (h2017v1) codec coverage

Per-element read/write coverage for the classic Worldographer / Hexographer 2
XML codec (`xmlio/h2017v1`), covering on-disk `version="1.73"`, `"1.74"`, and
`"1.77"` files. This uses the **same format** as `xmlio/h2025v1/COVERAGE.md`
(issue #7) so the two per-version matrices read as one artifact, and it exists
for the same reason: to make **stub-drift** visible. Classic decode is nearly
complete, but the encoder still contains several silent stubs — this matrix
records exactly which, with a code citation for every claim.

Classic (h2017v1) is **frozen** per `README.adoc` (security fixes only, no new
features), so these gaps are documented rather than scheduled: the point is that
a caller can see the truth before relying on a round-trip.

## Status vocabulary

Shared with the h2025 matrix:

- **implemented** — the field is fully modeled and moved between XML and `Map_t`
  in the named direction.
- **stub** — parsed into the model but only as raw chardata, not structured.
- **no-op(intentional)** — the encoder deliberately emits an empty wrapper and
  drops decoded content (documented, mirrors a section real classic files leave
  empty).
- **lossy** — some on-disk detail is not preserved / not sourced from `Map_t`.
- **unimplemented(dropped)** — decode reads the data into `Map_t`, but the
  encoder emits nothing for it (an empty wrapper or a hard error): the data is
  **silently lost on write**. This is the stub-drift class this matrix exists to
  surface, and it is much larger on the classic encoder than on h2025.

### Relationship to `wog/FEATURES.md` legend

The sibling ottomap repo's `wog/FEATURES.md` uses `✅ implemented / ⚠️ partial /
❌ not implemented`. The mapping is: **implemented** → ✅; **stub / lossy /
no-op(intentional)** → ⚠️ (partial, with documented caveats); **unimplemented(dropped)
/ not modeled** → ❌ (for the affected direction). The richer vocabulary is kept
here because it is more precise about *how* a field is partial (raw-chardata stub
vs. constant-block lossy vs. hard-error drop), which is exactly the distinction
that lets stub-drift hide.

## Tests

There is **no automated h2017 codec test** in the tree (verified:
`find . -name '*_test.go'` yields only `hexg/*_test.go` and
`xmlio/roundtrip_2025_test.go`; `xmlio/h2017v1/` has no `_test.go`). This absence
is itself a coverage gap — the classic codec is exercised only indirectly through
CLI tools, not asserted. Every status below is therefore cited to a **source
line** (`decode.go` / `encode.go`) rather than to a test, and to a decoded
classic sample where a fixture proves the shape. Decoded samples referenced are
the classic fixtures under `testdata/input/` (`blank-2017-1.73/1.74/1.77-1.0.wxx`,
`2017-1.77-1.0-{columns,rows}-blank.wxx`, `2017-1.77-1.0-{import,merge-01}.wxx`),
inspected by decompressing the gzip/UTF-16BE container to UTF-8 XML.

| `<map>` child element | Decode | Encode | Evidence | Notes |
|---|---|---|---|---|
| `<map>` root + scalar attributes (24) | implemented | implemented | `decode.go:38-76`, `encode.go:27-59` | All 24 modeled attrs round-trip. `version` attr (`1.73`/`1.74`/`1.77`) is preserved in `Map_t.Version` and re-emitted, but `MetaData.DataVersion` is **hardcoded** — see "Version identity" below. |
| `<gridandnumbering>` (30 attrs) | implemented | implemented | `decode.go:78-109`, `encode.go:110-145` | All 30 attributes modeled and re-emitted. |
| `<terrainmap>` | implemented | implemented | `decode.go:111-128`, `encode.go:147-158` | Tab-delimited name/slot table parsed into `TerrainMap_t`; re-emitted sorted by slot. |
| `<maplayer>` (`name`, `isVisible`) | implemented | implemented | `decode.go:130-132`, `encode.go:161-176` | Classic `<maplayer>` has no `opacity` attr (confirmed by RelaxNG + samples), so nothing is dropped. |
| `<tiles>`/`<tilerow>` — COLUMNS | implemented | implemented | `decode.go:151-245`, `encode.go:191-201` | COLUMNS (`OddQ`) grid fully round-trips. |
| `<tiles>`/`<tilerow>` — ROWS | implemented | **unimplemented(dropped)** | decode `decode.go:146-163`; encode `encode.go:202-203` | **ROWS decode works** (`OddR` branch); **ROWS encode returns a hard error** `assert(orientation != "ROWS")`. `2017-1.77-1.0-rows-blank.wxx` decodes but cannot be re-encoded. Same gap exists in h2025. |
| tile data (`terrain`, `elevation`, `isIcy`, `isGMOnly`, resources 6/7/11/12-col, `customBackgroundColor`) | implemented | implemented | `decode.go:166-244`, `encode.go:219-251` | `Z`-compressed and full 6-resource forms; optional trailing RGBA. Encoder auto-compresses when non-Animal resources are all zero. |
| `<mapkey>` | implemented | **lossy (constant block)** | decode `decode.go:247-279`; encode `encode.go:253-258` | Decode reads every `mapkey` attribute into `Map_t.MapKey`. **Encode ignores `Map_t.MapKey` entirely and writes a hardcoded default `<mapkey ...>` string.** A decoded-then-encoded map key is not preserved. |
| `<features>`/`<feature>` (+ `<location>`, inline `<label>`) | implemented | implemented | `decode.go:282-347`, `encode.go:261-321` | All feature attributes + nested location + inline label round-trip. Exercised by e.g. `2017-1.77-1.0-columns-blank.wxx` (4 features). |
| `<labels>`/`<label>` (standalone, + `<location>`) | implemented | implemented | `decode.go:349-384`, `encode.go:323-376` | Standalone labels round-trip; `backgroundColor` omitted on write when it is the `0.0,0.0,0.0,1.0` sentinel (`encode.go:343-345`). |
| `<shapes>`/`<shape>` (+ `<p>` points) | implemented | **unimplemented(dropped)** | decode `decode.go:386-439`; encode `encode.go:389-411` | Decode builds full `Shape_t` + points. **`encodeShape` is a commented-out no-op**: `<shapes></shapes>` is emitted with **no `<shape>` children** — shapes are silently lost on write. (No classic sample contains a populated `<shape>`, so this drop is invisible to the samples but real for any populated map.) |
| `<notes>`/`<note>` | implemented | **unimplemented(dropped)** | decode `decode.go:441-446`; encode `encode.go:413-436` | Decode reads `<note>` chardata into `Note_t.InnerText`. **`encodeNote` is a commented-out no-op** — `<notes></notes>` emitted empty. Every classic sample already has an empty `<notes>` element, so the loss is currently latent. |
| `<informations>`/`<information>` (+ nested detail) | implemented | **unimplemented(dropped)** | decode `decode.go:448-485`; encode `encode.go:438-442` | Decode reads the full lore tree (incl. `<information>` detail children) into `Informations_t`; samples carry 14–86 `<information>` entries. **`encodeInformations` emits only an empty `<informations></informations>` wrapper** — the entire lore tree is dropped on write. |
| configuration `<terrain-config>` | stub | no-op(intentional) | decode `decode.go:489-495`; encode `encode.go:465-469` | Parsed as raw chardata only; encoder emits an empty wrapper. Real samples leave it empty. |
| configuration `<feature-config>` | stub | no-op(intentional) | decode `decode.go:496-501`; encode `encode.go:471-475` | Same as terrain-config. |
| configuration `<texture-config>` | stub | no-op(intentional) | decode `decode.go:502-507`; encode `encode.go:477-481` | Same as terrain-config. |
| configuration `<text-config>`/`<labelstyle>` | implemented | **unimplemented(dropped)** | decode `decode.go:508-532`; encode `encode.go:483-507` | Decode builds structured `LabelStyle_t` (samples have 10 labelstyles). **`encodeLabelStyle` is a commented-out no-op** — `<text-config></text-config>` emitted with **no `<labelstyle>` children**. Labelstyles are lost on write. |
| configuration `<shape-config>`/`<shapestyle>` | implemented | implemented | decode `decode.go:533-575`; encode `encode.go:509-551` | Decode builds structured `ShapeStyle_t` (samples have 9–10 shapestyles); **`encodeShapeStyle` writes all attributes**. Note the asymmetry: shapestyle encodes, labelstyle (a peer sub-config) does not. |

## Version identity — `DataVersion` hardcoded

`decode.go:33` sets `w.MetaData.DataVersion = semver.Version{Major: 2017, Minor: 1}`
**unconditionally**, discarding whether the on-disk file was `version="1.73"`,
`"1.74"`, or `"1.77"`. The consequences:

- The on-disk `version` **attribute string** *is* preserved: it is copied to
  `Map_t.Version` (`decode.go:74`) and re-emitted verbatim by the encoder
  (`encode.go:30`). So a 1.73 file re-encodes with `version="1.73"`.
- But `MetaData.DataVersion` — the semver the public encoder dispatches on
  (`xmlio/encoder.go:154-159`, `2017.1 → h2017v1.Encode`) — collapses all three
  classic revisions to `2017.1`. The codec cannot tell 1.73 / 1.74 / 1.77 apart
  at the metadata level, and there is a single classic encoder for all three.

This is a fidelity gap only if a caller relies on `MetaData.DataVersion` to
distinguish classic sub-revisions; the round-tripped file itself keeps its
`version` attribute.

## Dispatch asymmetry — public decoder cannot read classic

The public `xmlio` pipeline is **write-only for classic**:

- **Encode**: `xmlio/encoder.go:154-159` dispatches `DataVersion.Major == 2017,
  Minor == 1 → h2017v1.Encode`. The public encoder can emit classic XML.
- **Decode**: `xmlio/decoder.go:233-243` routes **only** `release="2025"` to
  `h2025v1`. Classic files carry **no `release` attribute at all** (confirmed by
  every sample and by the RelaxNG schema, which defines no `release`), so a
  classic file falls through to `ErrUnsupportedMapMetadata` (`decoder.go:243`).

So `xmlio.NewDecoder().Decode(classicFile)` fails even though `h2017v1.Decode`
exists and works — the classic decoder is reachable **only** by calling the
package function `h2017v1.Decode` directly. Issue B4 decides the fix; this matrix
records the current state.

## Known un-modeled fields

**None found in the available classic samples.** Every attribute present on every
element across all seven classic fixtures is modeled by a field in
`xmlio/h2017v1/schema.go`. This was verified by sweeping the decoded UTF-8 XML of
all seven samples for `element → {attribute}` sets and diffing against the
`schema.go` structs. Element-by-element:

- `<map>` (23 attrs in samples), `<gridandnumbering>` (30), `<mapkey>` (23),
  `<feature>` (21), `<label>` (15), `<labelstyle>` (9), `<shapestyle>` (27),
  `<maplayer>` (2), `<tiles>` (3), `<location>` (4), `<information>` (11) — **all
  present attributes are modeled**; no leftover attribute in any element.

Two caveats on coverage of the sweep (not gaps, just scope limits):

- **`<shape>` and populated `<note>` are absent from the samples.** All seven
  classic fixtures have an empty `<notes>` element and no `<shape>` children, so
  those two elements' attribute fidelity is verified only against `schema.go`
  and the RelaxNG schema, not a live fixture. `schema.go` models `Shape_t` /
  `Point_t` / `Note_t`, and their fields match the RelaxNG definition (below).
- Un-modeled *behavior* on the **encode** side (shapes / notes / informations /
  labelstyles dropped, mapkey constant) is captured in the table above, not
  here; this section is specifically about attributes with **no field in the
  schema**, of which there are none for classic.

Contrast with h2025, whose "Known un-modeled fields" section lists six real
schema-modeling holes (`maplayer/@opacity`, `labelstyle/@dropShadow*`,
`shapestyle/@lineCap`+`@lineJoin`, `map/@hScrollbarPos`+`@vScrollbarPos`,
`blurTerrainBG`, `extraTerrain`). Those are all **W2025 additions** absent from
the classic format, which is why the classic schema has no equivalent gap.

## RelaxNG cross-check

Cross-checked against `schema/utf-8-xml.rnc` (the formal RelaxNG schema imported
in B1). That schema was reverse-engineered from a real `version="1.73"` classic
export, so it is **directly in scope for this codec** (see `schema/README.md`).

Every element the RelaxNG schema defines is **modeled** in `xmlio/h2017v1/schema.go`:

| RelaxNG element | Modeled in schema.go | Note |
|---|---|---|
| `map` + its 24 attrs | ✅ `XMLSchema` | 1:1. |
| `gridandnumbering` (30 attrs) | ✅ `GridAndNumbering` | 1:1. |
| `terrainmap` (text) | ✅ `TerrainMap_t` | Raw text, parsed downstream. |
| `maplayer` (`isVisible`, `name`) | ✅ `MapLayer_t` | RelaxNG confirms no `opacity` in classic. |
| `tiles`/`tilerow` | ✅ `Tiles_t`/`TileRow_t` | |
| `mapkey` (23 attrs) | ✅ `MapKey_t` | |
| `features`/`feature` (+ `location`, `label`) | ✅ `Feature` | |
| `labels`/`label` (+ `location`) | ✅ `Label_t`/`Labels_t` | |
| `shapes`/`shape`/`p` | ✅ `Shape_t`/`Point_t` | `shape/@fillRule`, `p/@type?` both modeled. |
| `notes` (`empty`) | ✅ `Notes_t`/`Note_t` | RelaxNG types `<notes>` as empty; schema models a `<note>` child for populated (W2025-style) files. |
| `informations`/`information` | ✅ `Informations_t`/`Information_t` | Recursive `information` detail modeled. |
| `configuration` + `terrain/feature/texture/text/shape-config` | ✅ `Configuration_t` | All five sub-configs present. |
| `labelstyle` (9 attrs) | ✅ `LabelStyle_t` | |
| `shapestyle` (27 attrs) | ✅ `ShapeStyle_t` | |

**Conclusion:** there is **no RelaxNG element that neither this codec models nor
lists as a gap** — decode models 100% of the schema. The classic gaps are all on
the **encode** side (ROWS, mapkey, shapes, notes, informations, labelstyles;
tabulated above), which the RelaxNG schema — being a document-shape grammar, not
a codec spec — cannot by itself detect. That is precisely why this matrix is
maintained alongside the schema.
