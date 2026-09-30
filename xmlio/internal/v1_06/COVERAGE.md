# W2025 (v1_06) codec coverage

Per-element read/write coverage for the Worldographer 2025 XML codec
(`xmlio/internal/v1_06`). This mirrors `wog/FEATURES.md` from the sibling ottomap repo
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
lets stub-drift hide. The classic matrix (`xmlio/internal/v0_77/COVERAGE.md`) uses this
same vocabulary.

Tests referenced (in `xmlio/roundtrip_2025_test.go` unless noted, package
`xmlio_test`):

- **RoundTrip** = `TestW2025RoundTrip` (in-memory codec over the real
  `testdata/2025-2.06-13x11-941577-blank.wxx` sample)
- **PublicRoundTrip** = `TestW2025PublicRoundTrip` (full gzip/UTF-16/header
  pipeline over the same sample)
- **DecodeBoth** = `TestW2025Decode_BothSamples`
- **NotesShapesRoundTrip** = `TestW2025NotesShapesRoundTrip` (in-memory codec
  over `testdata/2025-2.07-13x11-941577-notes-shapes.wxx`, which fills the
  features, inline labels, shapes, notes and `<extraTerrain>` placements the
  blank sample leaves empty; encoded as 2.06, so `MetaData` is not compared)
- **NotesShapesPublicRoundTrip** = `TestW2025NotesShapesPublicRoundTrip` (full
  gzip/UTF-16/header pipeline over the same 2.07 map, proving the transport
  layers round-trip shapes/notes/features/labels too)
- **ConfigEmpty** = `TestW2025ConfigSectionsEmpty`
- **RowsRoundTrip** = `TestW2025RowsRoundTrip` (in
  `xmlio/rows_encode_2025_test.go`; in-memory encode->decode over an asymmetric
  2x3 ROWS grid, asserting orientation and per-cell position fidelity)
- **CoverageMatrix** = `TestW2025CoverageMatrix` (in `xmlio/coverage_2025_test.go`;
  decode->encode->decode over both the 2.07 notes-shapes map and the real sample,
  asserting per-element counts and key field values -- including the six
  W2025-native fields modeled in #11)

| `<map>` child element | Decode | Encode | Test(s) | Notes |
|---|---|---|---|---|
| `<map>` root + scalar attributes | implemented | implemented | RoundTrip, PublicRoundTrip, DecodeBoth, CoverageMatrix | `hScrollbarPos` / `vScrollbarPos` now modeled (#11). |
| `<gridandnumbering>` (30 attrs) | implemented | implemented | RoundTrip, PublicRoundTrip | All 30 attributes modeled and re-emitted. |
| `<terrainmap>` | implemented | implemented | RoundTrip, DecodeBoth | Tab-delimited name/slot table parsed into `TerrainMap_t`. The encoder writes the table in index order (#87). Worldographer does not always: 2.07 saved the notes-shapes map as `Blank 0`, `Classic/Water Sea 2`, `Classic/Flat Farmland 1`, so re-encoding it reorders the entries while each name keeps its index. Whether the order matters to Worldographer is untested in the app; NotesShapesRoundTrip compares the table in index order. |
| `<maplayer>` | implemented | implemented | RoundTrip, PublicRoundTrip, CoverageMatrix | `opacity` now modeled (#11); `name` + `isVisible` + `opacity` round-trip. |
| `<tiles>` / `<tilerow>` | implemented | implemented | RoundTrip, PublicRoundTrip, DecodeBoth, RowsRoundTrip | Decode handles COLUMNS and ROWS; **encoder now supports COLUMNS and ROWS** (`tiles.go` `encodeTiles`). The physical `<tilerow>` emission is orientation-independent — decode stores tiles in file-physical `Tiles[x][y]` order (`tilesWide` rows of `tilesHigh` lines) for both orientations, so ROWS emits the identical structure; orientation only affects the OddQ/OddR coordinate interpretation and the RowsHigh/ColumnsWide labels. The on-disk `.wxx` sample is COLUMNS; ROWS is covered by `TestW2025RowsRoundTrip`, which builds an asymmetric 2x3 ROWS grid in memory and asserts every cell round-trips to the same position (catching any transpose). |
| tile data (terrain, elevation, isIcy, isGMOnly, resources, customBackgroundColor) | implemented | implemented | RoundTrip, PublicRoundTrip | 6/7/11/12-column forms + `Z`-compressed resources; opaque-black `customBackgroundColor` folds to nil per `decodeRgba`. |
| `<mapkey>` | implemented | implemented | RoundTrip, PublicRoundTrip, **IntegerAttributeSpelling**, **NonIntegralValueRefused** | All attributes modeled. (Decode is nested inside the tilerow loop but runs given >=1 tilerow.) **Five attributes are INTEGERS on disk and must never carry a decimal point** — `@height`, `@backgroundopacity`, `@titleScale`, `@scaleScale`, `@entryScale`. Worldographer reads them with `Integer.parseInt` and **refuses to open the file** otherwise (issue #64: `NumberFormatException: For input string: "-1.0"` at `LoadMapTask.readMapKey`, reproduced by experiment). They are typed `Int_t` in `schema.go` and rendered through it; a caller setting a non-integral value is refused before any output is written. |
| `<features>` / `<feature>` | implemented | implemented | CoverageMatrix, NotesShapesRoundTrip | Real blank sample has no features; the 2.07 notes-shapes map has 3. |
| feature `<location>` | implemented | implemented | CoverageMatrix, NotesShapesRoundTrip | viewLevel/x/y. |
| feature inline `<label>` (optional) | implemented | implemented | CoverageMatrix, NotesShapesRoundTrip | `Feature.Label` is `*Label_t`; decode nil-guards a labelless feature so encode omits `<label>`. No app-saved fixture has a labelless feature, so that path is untested. |
| `<labels>` / `<label>` (standalone) | implemented | **partial** | RoundTrip, NotesShapesRoundTrip (empty in both) | Shares `encodeLabel` with the inline feature label. **`dropShadowColor` / `dropShadowRadius` / `dropShadowSpread` are NOT modeled on `Label_t`** and are dropped on a same-release round trip (#35) -- the trio is modeled on `LabelStyle_t`, not here. |
| label `<location>` (with `scale`) | implemented | implemented | CoverageMatrix, NotesShapesRoundTrip | Exercised by inline feature labels only. |
| `<shapes>` / `<shape>` (+ `<p>` points) | implemented | implemented | CoverageMatrix, NotesShapesRoundTrip, **ShapesMatchSource**, **ShapeExtraLineZeroOmitted**, **ShapeIntegerXYFractionalRefused**, **ShapePointSpellingDecode** (in `xmlio/shapes_w2025_test.go`) | Real sample has no shapes; the 2.07 notes-shapes map has 4 with points (CoverageMatrix checks `Points[0]`). `<shape>` DOES model `lineCap`/`lineJoin`. Issue #94 closed three same-version differences, all observed on the notes-shapes fixtures (2.06, 2.07, 2.08 and the autosaves): (1) the four `@extraLineDistance @extraLineLength @extraLineWidth @extraLineSeparation`, which every shape states, were dropped; they are now modeled, and written all four or none -- none when all four are zero, which is what a classic-decoded shape holds (**app check pending**: whether `extraLineWidth="0.0"` reads the same as no attribute is untested). (2) `@fillRule` and `p/@type` were written as `""` where the source states none; they are now written only when non-empty. No 2.06/2.07/2.08 fixture states either as `""`: the tile-border polygons state no `@fillRule` and each path states `fillRule="NON_ZERO"` and `type="m"` on its first point only. (3) Every coordinate was written with a decimal point; the app spells tile-border polygon points as integers (`x="2700" y="150"`) and path points as decimals (`x="1950.0"`). `schema.go` now reads `p/@x`/`@y` as strings and `wxx.Point_t.IntegerXY` records that both were integer-spelled (a point spelled one each way decodes false; no fixture has one); the encoder writes that spelling back, and refuses a point marked `IntegerXY` with a fractional coordinate before writing anything (`ErrInvalidIntegerAttribute`). Nothing is derived from `isMatchTileBorders` or the shape type. The encoder writes the attributes in 2.07's order, with the app's `<shape  type=` and ` <p ` layout; the encoded `<shapes>` element differs from the source only in the app's `y = "..."` spelling on path points (spaces around `=`, the same XML), on every notes-shapes fixture. (4) Curve control points: `testdata/2025-2.08-13x11-941577-populated.wxx` (observed in 2.08) has a curved path (`isCurve="true"`) whose `type="c"` point states `@cx1 @cy1 @cx2 @cy2`, which were dropped. They are modeled as `wxx.Point_t.Control *CurveControl_t` (nil = the point states none, so a control point at 0,0 survives), written all four or none, and a point stating only some is refused on decode. All four values in the fixture are decimal-spelled, so they have no integer flag. The same fixture has the first long fractional coordinates (`x="1715.028150714595"`, `cx1="2390.6453009961024"`); `floats` reproduces them byte for byte, and its encoded `<shapes>` again differs from the source only in the app's spaced `y = `, `cy1 = `, `cy2 = `. |
| `<notes>` / `<note>` (+ `<notetext>`, `<location>`) | implemented | implemented | **NotesMatchSource**, **NotesDecodeLocation**, **NoteWithoutLocationRefused**, **NoteKeyMismatchRefused** (in `xmlio/notes_w2025_test.go`) | Modeled as 2.07, 2.08 and 2.06 write it (issue #94): `@key @originalViewLevel @filename @parent @color @isWorld @isContinent @isKingdom @isProvince @title`, then `<notetext>` (CDATA, verbatim), then `<location viewLevel x y />`. Until #94 the codec read an older spelling (`@viewLevel @x @y @isGMOnly`, no `<location>`) that no build in scope writes, so every note lost its position on decode and was written at `viewLevel="" x="0.0" y="0.0"`; a `Map_t` round trip could not see it. `@key` repeats the location as `<viewLevel>,<x>,<y>`: `wxx.Note_t` has no `Key`, the encoder derives it from `Location`, and decode refuses a key naming a different place (`ErrNoteKeyMismatch`). A note with no `Location` is refused before any output (`ErrNoteWithoutLocation`), never written at 0,0. The encoded `<notes>` element is byte-identical to the source on every notes-shapes fixture (2.06, 2.07, 2.08, and the autosaves). |
| `<informations>` / `<information>` (+ nested `<information>`, any depth) | implemented | implemented | RoundTrip, PublicRoundTrip, InformationAttrsMatchSource | `Information_t.Details` is recursive. Real files nest three deep (Information > pantheon > god): 91 entries in the 2.06 blank and layers-beta fixtures, 74 in the 2.07 blank. Until #69, this row said "implemented" while decode copied two levels and dropped the third (76 of 91 entries). The `Map_t`-to-`Map_t` round-trip tests could not see that. `InformationAttrsMatchSource` compares every start tag against the source document, and the element count is what now holds it. |
| configuration `<terrain-config>` | stub | no-op(intentional) | ConfigEmpty | Parsed as raw chardata only; encoder emits empty wrapper. Lossless only because real samples leave it empty (guarded by ConfigEmpty). |
| configuration `<feature-config>` | stub | no-op(intentional) | ConfigEmpty | Same as terrain-config. |
| configuration `<texture-config>` | stub | no-op(intentional) | ConfigEmpty | Same as terrain-config. |
| configuration `<text-config>` / `<labelstyle>` | implemented | implemented | RoundTrip, PublicRoundTrip, CoverageMatrix, **AttrsMatchSource**, **BlackBackgroundIsNotNull**, **IntegerAttributeSpelling** | 7 labelstyles in sample round-trip; `dropShadowColor` (nullable string) / `dropShadowRadius` / `dropShadowSpread` now modeled (#11). **Both nullable colours are exact as of #62**: `backgroundColor="null"` used to come back as `"0.0,0.0,0.0,1.0"` on every label style of every file this codec wrote, because `decodeRgba` folded `"null"` and opaque black into the same nil and `rgbas` rendered nil as black. Decode now uses `decodeZeroableRgba` so nil means `"null"` and nothing else, and encode uses `rgbaOrNull`; `rgbans` was rejected as the fix because it decides on the formatted string and would have laundered a genuine opaque black into `"null"` instead. `TestW2025LabelStyleAttrsMatchSource` compares every attribute against the source document — the audit classic always had and W2025 lacked, which is why the structural round-trip tests could not see this — and `TestW2025LabelStyleBlackBackgroundIsNotNull` synthesizes the black case no fixture carries. `dropShadowRadius` and `dropShadowSpread` are **integers on disk** and are now written as such (issue #64); they were emitted `"0.0"`, which Worldographer refuses to load. The element is byte-identical to the source apart from inter-attribute whitespace. |
| configuration `<shape-config>` / `<shapestyle>` | implemented | implemented | RoundTrip, PublicRoundTrip, CoverageMatrix | 7 shapestyles in sample round-trip; `lineCap` / `lineJoin` now modeled (#11). |
| `<blurTerrainBG>` | implemented | implemented | CoverageMatrix | Optional top-level element modeled as `*BlurTerrainBG_t` (nil = absent); 6 attrs round-trip (#11). |
| `<extraTerrain>` (+ `<mapLayer>` / `<terrainAndLocation>`) | implemented | implemented | ExtraTerrainMatchesSource, ExtraTerrainRoundTrip, ExtraTerrainDecodeRefusals, ClassicDowngradeExtraTerrain | Modeled structurally by #34 as `ExtraTerrain_t` → `[]ExtraTerrainLayer_t` → `[]TerrainAndLocation_t` (terrain name, integer elevation, icy, GM-only, resources, and `location` kept as the raw x,y point). Encode reproduces the source byte for byte on all three samples. Decode refuses any child element or attribute it does not model instead of dropping it, because the element used to be carried verbatim. `…-blank.wxx` carries an empty container, `…-layers.wxx` one placement; multi-layer shapes are synthesized. |

## Integer attributes

**22 attributes of this schema are Java `int` fields**, written without a decimal
point, and Worldographer reads them with `Integer.parseInt`: a decimal point in
any of them is a hard load failure, not a formatting difference (issue #64).
`schema.go` declares them as integer types, which is the single statement both
halves of the codec read — 15 were already `int` and the 7 the struct mistyped as
`float64` were exactly the ones that shipped broken.

`Map_t` deliberately does not carry this knowledge. It is the superset of every
supported format (ADR 0004 Decision 6) and keeps these fields `float64`; the
conversion happens at the codec boundary, and a value the schema cannot state is
refused there rather than rounded onto disk.

`TestW2025IntegerAttributeSpelling` enforces the rule without a list to maintain:
every attribute a tracked document always spells integrally must be emitted
integrally. It reads RAW attribute values — `xmlAggregate`'s `normVal`
canonicalizes `"0"` and `"0.0"` to the same string, which is right for the loss
inventory and blind to this.

The classic codec is **not** covered and is safe only because its `<mapkey>` is a
hard-coded constant carrying the integer spellings verbatim: see issue #67.

## Known un-modeled fields

**None.** The two gaps found after #11 are both closed:

- **`<extraTerrain>` children (`<mapLayer>` / `<terrainAndLocation>`)** were an
  opaque `InnerXML` stub. #34 modeled them, and a downgrade now reports the loss
  instead of failing.
- **`<label>` `@dropShadowColor` / `@dropShadowRadius` / `@dropShadowSpread`**
  were dropped on a 2025 → 2025 round trip. #35 modeled them on `Label_t`.

For the record, the six fields #11 modeled -- and where they now live -- were:

- **`<maplayer opacity>`** -- `MapLayer_t.Opacity` (float) / schema `MapLayer_t.Opacity`. Round-trips via CoverageMatrix (`MapLayers[0].Opacity == 1.0`).
- **`<labelstyle dropShadowColor / dropShadowRadius / dropShadowSpread>`** -- `LabelStyle_t.DropShadowColor` (nullable string, preserves `"null"`), `.DropShadowRadius`, `.DropShadowSpread` (floats). CoverageMatrix asserts `DropShadowColor == "null"` and zero radius/spread.
- **`<shapestyle lineCap / lineJoin>`** -- `ShapeStyle_t.LineCap` / `.LineJoin` (strings), mirroring `Shape_t`. CoverageMatrix asserts `SQUARE` / `ROUND`.
- **`<map hScrollbarPos / vScrollbarPos>`** -- `Map_t.HScrollbarPos` / `.VScrollbarPos` (floats) / schema root attrs. CoverageMatrix asserts they do not drift.
- **`<blurTerrainBG>`** -- `Map_t.BlurTerrainBG *BlurTerrainBG_t` (nil = absent); 6 attrs modeled. CoverageMatrix asserts non-nil with attrs preserved.
- **`<extraTerrain>`** -- `Map_t.ExtraTerrain *ExtraTerrain_t` (nil = absent). #11 modeled only the container, carrying its content as raw innerxml, a **stub** that made a downgrade hard-error. #34 replaced the stub with structured types (see the matrix row above).

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
  `labelstyle`, `shapestyle` — has a corresponding type in `xmlio/internal/v1_06/schema.go`
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
