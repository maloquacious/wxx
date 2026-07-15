# W2025 (h2025v1) codec coverage

Per-element read/write coverage for the Worldographer 2025 XML codec
(`xmlio/h2025v1`). This mirrors `wog/FEATURES.md` from the sibling ottomap repo
and exists to make stub-drift visible: this whole ticket (#7) began because a
stub encoder hid behind a passing round-trip test.

"**implemented**" here means the round-trip **at the `Map_t` level** is proven
by the named test: decode -> encode -> decode reproduces the same in-memory
model. It does **not** promise byte-for-byte on-disk fidelity. Fields that are
present in real Worldographer output but have no field in `schema.go` would be
silently dropped on decode; because encode never re-emits them either, the
`Map_t` round-trip still passes while the on-disk data is lost -- that class of
gap is exactly what this matrix exists to surface.

The six W2025-native fields that were formerly dropped (`maplayer/@opacity`,
`labelstyle/@dropShadow*`, `shapestyle/@lineCap`+`@lineJoin`,
`map/@hScrollbarPos`+`@vScrollbarPos`, `<blurTerrainBG>`, `<extraTerrain>`) are
now modeled additively and wired through decode+encode; **issue #11 closed the
"Known un-modeled fields" section** and each of the six is exercised by
`TestW2025CoverageMatrix` (see the **CoverageMatrix** test below).

Statuses: **implemented** (full `Map_t` round-trip) / **stub** (parsed into the
model but only as raw chardata, not structured) / **no-op(intentional)** (encoder
deliberately emits an empty wrapper and drops decoded content, documented +
guarded by a test) / **lossy** (some on-disk detail is not preserved).

### Relationship to `wog/FEATURES.md` legend

The sibling ottomap repo's `wog/FEATURES.md` uses `✅ implemented / ⚠️ partial /
❌ not implemented`. The mapping is: **implemented** → ✅; **stub / lossy /
no-op(intentional)** → ⚠️ (partial, with documented caveats); **not modeled /
not emitted** → ❌ (for the affected direction). The richer vocabulary is kept
here because it distinguishes *how* a field is partial (raw-chardata stub vs.
constant-block lossy vs. symmetric drop), which is exactly the distinction that
lets stub-drift hide. The classic matrix (`xmlio/h2017v1/COVERAGE.md`) uses this
same vocabulary.

Tests referenced (in `xmlio/roundtrip_2025_test.go` unless noted, package
`xmlio_test`):

- **RoundTrip** = `TestW2025RoundTrip` (in-memory codec over the real
  `testdata/2025-2.06-13x11-941577-blank.wxx` sample)
- **PublicRoundTrip** = `TestW2025PublicRoundTrip` (full gzip/UTF-16/header
  pipeline over the same sample)
- **DecodeBoth** = `TestW2025Decode_BothSamples`
- **DecodePopulated** = `TestW2025DecodePopulated` (decode-side assertions over
  `testdata/w2025-populated.xml`)
- **PopulatedRoundTrip** = `TestW2025PopulatedRoundTrip` (in-memory codec over
  the populated fixture, which fills features/labels/shapes/notes the blank
  sample leaves empty)
- **PopulatedPublicRoundTrip** = `TestW2025PopulatedPublicRoundTrip` (full
  gzip/UTF-16/header pipeline over the populated fixture, proving the transport
  layers round-trip populated shapes/notes/features/labels too)
- **ConfigEmpty** = `TestW2025ConfigSectionsEmpty`
- **RowsRoundTrip** = `TestW2025RowsRoundTrip` (in
  `xmlio/rows_encode_2025_test.go`; in-memory encode->decode over an asymmetric
  2x3 ROWS grid, asserting orientation and per-cell position fidelity)
- **CoverageMatrix** = `TestW2025CoverageMatrix` (in `xmlio/coverage_2025_test.go`;
  decode->encode->decode over both the populated fixture and the real sample,
  asserting per-element counts and key field values -- including the six
  W2025-native fields modeled in #11)

| `<map>` child element | Decode | Encode | Test(s) | Notes |
|---|---|---|---|---|
| `<map>` root + scalar attributes | implemented | implemented | RoundTrip, PublicRoundTrip, DecodeBoth, CoverageMatrix | `hScrollbarPos` / `vScrollbarPos` now modeled (#11). |
| `<gridandnumbering>` (30 attrs) | implemented | implemented | RoundTrip, PublicRoundTrip | All 30 attributes modeled and re-emitted. |
| `<terrainmap>` | implemented | implemented | RoundTrip, DecodeBoth | Tab-delimited name/slot table parsed into `TerrainMap_t`. |
| `<maplayer>` | implemented | implemented | RoundTrip, PublicRoundTrip, CoverageMatrix | `opacity` now modeled (#11); `name` + `isVisible` + `opacity` round-trip. |
| `<tiles>` / `<tilerow>` | implemented | implemented | RoundTrip, PublicRoundTrip, DecodeBoth, RowsRoundTrip | Decode handles COLUMNS and ROWS; **encoder now supports COLUMNS and ROWS** (`tiles.go` `encodeTiles`). The physical `<tilerow>` emission is orientation-independent — decode stores tiles in file-physical `Tiles[x][y]` order (`tilesWide` rows of `tilesHigh` lines) for both orientations, so ROWS emits the identical structure; orientation only affects the OddQ/OddR coordinate interpretation and the RowsHigh/ColumnsWide labels. The on-disk `.wxx` sample is COLUMNS; ROWS is covered by `TestW2025RowsRoundTrip`, which builds an asymmetric 2x3 ROWS grid in memory and asserts every cell round-trips to the same position (catching any transpose). |
| tile data (terrain, elevation, isIcy, isGMOnly, resources, customBackgroundColor) | implemented | implemented | RoundTrip, PublicRoundTrip | 6/7/11/12-column forms + `Z`-compressed resources; opaque-black `customBackgroundColor` folds to nil per `decodeRgba`. |
| `<mapkey>` | implemented | implemented | RoundTrip, PublicRoundTrip | All attributes modeled. (Decode is nested inside the tilerow loop but runs given >=1 tilerow.) |
| `<features>` / `<feature>` | implemented | implemented | DecodePopulated, PopulatedRoundTrip | Real blank sample has no features; populated fixture exercises them. |
| feature `<location>` | implemented | implemented | PopulatedRoundTrip | viewLevel/x/y. |
| feature inline `<label>` (optional) | implemented | implemented | DecodePopulated, PopulatedRoundTrip | `Feature.Label` is `*Label_t`; decode nil-guards a labelless feature so encode omits `<label>` (DecodePopulated asserts `Features[1].Label == nil`). |
| `<labels>` / `<label>` (standalone) | implemented | implemented | PopulatedRoundTrip (1 label in fixture), RoundTrip (empty in sample) | Shares `encodeLabel` with the inline feature label. |
| label `<location>` (with `scale`) | implemented | implemented | PopulatedRoundTrip | |
| `<shapes>` / `<shape>` (+ `<p>` points) | implemented | implemented | DecodePopulated, PopulatedRoundTrip | Real sample has no shapes; fixture has 2 shapes with points (DecodePopulated checks `Points[0]`). `<shape>` DOES model `lineCap`/`lineJoin`. |
| `<notes>` / `<note>` (+ `<notetext>`) | implemented | implemented | DecodePopulated, PopulatedRoundTrip | `notetext` CDATA body preserved verbatim; fixture has 2 notes. |
| `<informations>` / `<information>` (+ nested `<information>` detail) | implemented | implemented | RoundTrip, PublicRoundTrip | Real sample has 68 `<information>` elements incl. nested detail. |
| configuration `<terrain-config>` | stub | no-op(intentional) | ConfigEmpty | Parsed as raw chardata only; encoder emits empty wrapper. Lossless only because real samples leave it empty (guarded by ConfigEmpty). |
| configuration `<feature-config>` | stub | no-op(intentional) | ConfigEmpty | Same as terrain-config. |
| configuration `<texture-config>` | stub | no-op(intentional) | ConfigEmpty | Same as terrain-config. |
| configuration `<text-config>` / `<labelstyle>` | implemented | implemented | RoundTrip, PublicRoundTrip, CoverageMatrix | 7 labelstyles in sample round-trip; `dropShadowColor` (nullable string) / `dropShadowRadius` / `dropShadowSpread` now modeled (#11). |
| configuration `<shape-config>` / `<shapestyle>` | implemented | implemented | RoundTrip, PublicRoundTrip, CoverageMatrix | 7 shapestyles in sample round-trip; `lineCap` / `lineJoin` now modeled (#11). |
| `<blurTerrainBG>` | implemented | implemented | CoverageMatrix | Optional top-level element modeled as `*BlurTerrainBG_t` (nil = absent); 6 attrs round-trip (#11). |
| `<extraTerrain>` | **stub** | implemented | CoverageMatrix, ClassicDowngradeStubError | Optional top-level element modeled as `*ExtraTerrain_t` (nil = absent); the container round-trips but its **content is opaque raw innerxml**, not structured (#11). Both shapes are tracked: `…-blank.wxx` carries an empty container (innerxml `"\n"`), `…-layers.wxx` carries 183 bytes — a `<mapLayer name="Terrain Layer">` holding a `<terrainAndLocation>`. Decode is **stub**, not implemented: nothing in `Map_t` understands those children. |

## Known un-modeled fields

**None.** Issue #11 closed this section: the six W2025-native fields that were
formerly listed here are now modeled additively in `schema.go` + `Map_t`, wired
through h2025 decode and encode, and each is exercised by a green 2025 round trip
(`TestW2025CoverageMatrix`). For the record, the six -- and where they now live --
were:

- **`<maplayer opacity>`** -- `MapLayer_t.Opacity` (float) / schema `MapLayer_t.Opacity`. Round-trips via CoverageMatrix (`MapLayers[0].Opacity == 1.0`).
- **`<labelstyle dropShadowColor / dropShadowRadius / dropShadowSpread>`** -- `LabelStyle_t.DropShadowColor` (nullable string, preserves `"null"`), `.DropShadowRadius`, `.DropShadowSpread` (floats). CoverageMatrix asserts `DropShadowColor == "null"` and zero radius/spread.
- **`<shapestyle lineCap / lineJoin>`** -- `ShapeStyle_t.LineCap` / `.LineJoin` (strings), mirroring `Shape_t`. CoverageMatrix asserts `SQUARE` / `ROUND`.
- **`<map hScrollbarPos / vScrollbarPos>`** -- `Map_t.HScrollbarPos` / `.VScrollbarPos` (floats) / schema root attrs. CoverageMatrix asserts they do not drift.
- **`<blurTerrainBG>`** -- `Map_t.BlurTerrainBG *BlurTerrainBG_t` (nil = absent); 6 attrs modeled. CoverageMatrix asserts non-nil with attrs preserved.
- **`<extraTerrain>`** -- `Map_t.ExtraTerrain *ExtraTerrain_t` (nil = absent); the container is preserved via **raw innerxml**, so its content remains a **stub**. CoverageMatrix asserts non-nil. The tracked `…-layers.wxx` fixture carries 183 bytes of real children (`<mapLayer name="Terrain Layer">` / `<terrainAndLocation>`), so "present-but-empty" describes only the `…-blank.wxx` fixture (innerxml `"\n"`). Because nothing in `Map_t` understands those children, a downgrade to a target with no `<extraTerrain>` **hard-errors** rather than dropping them silently (#32; `xmlio/downgrade.go`) — this is the ADR 0004 "stub coverage is a precondition for honest loss reporting" case, and it resolves when #11 models `terrainAndLocation`.

## RelaxNG cross-check

The formal RelaxNG schema in `schema/utf-8-xml.rnc` (imported in B1) is **classic
`version="1.73"` scope only** — it predates the W2025 format (see
`schema/README.md`). It is therefore a **partial** checklist for h2025: only the
elements W2025 *shares* with classic are cross-checkable; the schema says nothing
about W2025 additions.

- **Shared elements are all modeled by h2025.** Every element the RelaxNG schema
  defines — `map`, `gridandnumbering`, `terrainmap`, `maplayer`, `tiles`/`tilerow`,
  `mapkey`, `features`/`feature`, `location`, `labels`/`label`, `shapes`/`shape`/`p`,
  `notes`, `informations`/`information`, `configuration` + its five sub-configs,
  `labelstyle`, `shapestyle` — has a corresponding type in `xmlio/h2025v1/schema.go`
  and appears as **implemented** in the table above. No RelaxNG element is
  unmodeled by the h2025 decoder.
- **The six W2025-native fields modeled in #11 lie outside the schema's scope.**
  Each is a **W2025 addition** the classic RelaxNG schema neither describes nor
  could have flagged: the schema defines `<maplayer>` with only `isVisible`/`name`
  (no `opacity`), and defines no `blurTerrainBG`, no `extraTerrain`, no
  `dropShadow*` on `<labelstyle>`, no `lineCap`/`lineJoin` on `<shapestyle>`, and
  no `hScrollbarPos`/`vScrollbarPos` on `<map>`. `schema/README.md` independently
  flags `maplayer/@opacity`, `blurTerrainBG`, and `extraTerrain` as verified
  W2025 deltas absent from the schema — corroborating that these were real format
  additions, not modeling oversights the classic schema could have warned about.
  They are now modeled additively and proven by `TestW2025CoverageMatrix`.

**Conclusion:** the classic-scoped RelaxNG schema confirms h2025 models 100% of
the *shared* format, and confirms the six #11 fields are genuine W2025 extensions
the schema does not (and cannot) describe. A W2025-scoped formal schema would be
needed to mechanically check the additions; until then `TestW2025CoverageMatrix`
is the executable checklist for them.
