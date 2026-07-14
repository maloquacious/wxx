# W2025 (h2025v1) codec coverage

Per-element read/write coverage for the Worldographer 2025 XML codec
(`xmlio/h2025v1`). This mirrors `wog/FEATURES.md` from the sibling ottomap repo
and exists to make stub-drift visible: this whole ticket (#7) began because a
stub encoder hid behind a passing round-trip test.

"**implemented**" here means the round-trip **at the `Map_t` level** is proven
by the named test: decode -> encode -> decode reproduces the same in-memory
model. It does **not** promise byte-for-byte on-disk fidelity. Fields that are
present in real Worldographer output but have no field in `schema.go` are
silently dropped on decode; because encode never re-emits them either, the
`Map_t` round-trip still passes while the on-disk data is lost. Those are listed
under "Known un-modeled fields" below -- that gap is exactly what this matrix
exists to surface.

Statuses: **implemented** (full `Map_t` round-trip) / **stub** (parsed into the
model but only as raw chardata, not structured) / **no-op(intentional)** (encoder
deliberately emits an empty wrapper and drops decoded content, documented +
guarded by a test) / **lossy** (some on-disk detail is not preserved).

Tests referenced (all in `xmlio/roundtrip_2025_test.go`, package `xmlio_test`):

- **RoundTrip** = `TestW2025RoundTrip` (in-memory codec over the real
  `data/2025-2.05.wxx` sample)
- **PublicRoundTrip** = `TestW2025PublicRoundTrip` (full gzip/UTF-16/header
  pipeline over the same sample)
- **DecodeBoth** = `TestW2025Decode_BothSamples`
- **DecodePopulated** = `TestW2025DecodePopulated` (decode-side assertions over
  `testdata/input/w2025-populated.xml`)
- **PopulatedRoundTrip** = `TestW2025PopulatedRoundTrip` (in-memory codec over
  the populated fixture, which fills features/labels/shapes/notes the blank
  sample leaves empty)
- **PopulatedPublicRoundTrip** = `TestW2025PopulatedPublicRoundTrip` (full
  gzip/UTF-16/header pipeline over the populated fixture, proving the transport
  layers round-trip populated shapes/notes/features/labels too)
- **ConfigEmpty** = `TestW2025ConfigSectionsEmpty`

| `<map>` child element | Decode | Encode | Test(s) | Notes |
|---|---|---|---|---|
| `<map>` root + scalar attributes | implemented | implemented | RoundTrip, PublicRoundTrip, DecodeBoth | `hScrollbarPos` / `vScrollbarPos` un-modeled (see below). |
| `<gridandnumbering>` (30 attrs) | implemented | implemented | RoundTrip, PublicRoundTrip | All 30 attributes modeled and re-emitted. |
| `<terrainmap>` | implemented | implemented | RoundTrip, DecodeBoth | Tab-delimited name/slot table parsed into `TerrainMap_t`. |
| `<maplayer>` | implemented | lossy | RoundTrip, PublicRoundTrip | `opacity` attr un-modeled -> dropped (see below). Only `name` + `isVisible` round-trip. |
| `<tiles>` / `<tilerow>` | implemented | implemented | RoundTrip, PublicRoundTrip, DecodeBoth | Decode handles COLUMNS and ROWS; **encoder supports COLUMNS only -- ROWS returns an error** (`encode.go` `encodeTiles`). Sample is COLUMNS. |
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
| configuration `<text-config>` / `<labelstyle>` | implemented | lossy | RoundTrip, PublicRoundTrip | 7 labelstyles in sample fully round-trip at `Map_t` level; but `dropShadowColor` / `dropShadowRadius` / `dropShadowSpread` un-modeled -> dropped (see below). |
| configuration `<shape-config>` / `<shapestyle>` | implemented | lossy | RoundTrip, PublicRoundTrip | 7 shapestyles in sample round-trip at `Map_t` level; but `lineCap` / `lineJoin` un-modeled -> dropped (see below). |
| `<blurTerrainBG>` | not modeled | not emitted | -- | No field in `XMLSchema`; present in sample, dropped (see below). |
| `<extraTerrain>` | not modeled | not emitted | -- | No field in `XMLSchema`; present in sample, dropped (see below). |

## Known un-modeled fields

These attributes/elements appear in the real `data/2025-2.05.utf8` sample (and in
Worldographer output generally) but have **no corresponding field in
`xmlio/h2025v1/schema.go`**. They are therefore silently discarded on decode and
never written on encode. Because decode and encode ignore them **symmetrically**,
the `Map_t`-level round-trip tests still pass -- but the data is **not preserved
to disk**. This is precisely the kind of drift this matrix is meant to catch.

Each item below was verified by grepping the sample and reading `schema.go`:

- **`<maplayer opacity>`** -- sample: `<maplayer name="Labels" isVisible="true" opacity="1.0"/>`. `MapLayer_t` models only `Name` and `IsVisible`; no `Opacity` field. Dropped.
- **`<labelstyle dropShadowColor / dropShadowRadius / dropShadowSpread>`** -- sample: `<labelstyle name="Nation" ... dropShadowColor="null" dropShadowRadius="0" dropShadowSpread="0" />`. `LabelStyle_t` has no `dropShadow*` fields. Dropped.
- **`<shapestyle lineCap / lineJoin>`** -- sample: `<shapestyle name="Trail" ... lineCap="SQUARE" lineJoin="ROUND" />`. `ShapeStyle_t` has no `LineCap`/`LineJoin` fields. Dropped. (Note: the `<shape>` element -- a different type, `Shape_t` -- *does* model `lineCap`/`lineJoin`; only `<shapestyle>` drops them.)
- **`<map hScrollbarPos / vScrollbarPos>`** -- sample: `hScrollbarPos="0.0"` and `vScrollbarPos="0.0"` on the root `<map>` element. `XMLSchema` models neither. Dropped (these are UI scroll positions, so the loss is cosmetic).
- **`<blurTerrainBG>`** -- top-level element in the sample: `<blurTerrainBG blur="false" topBleed="0.33" bottomBleed="0.65" randomness="0.1" blurStart="0.4" blurEnd="0.95"/>`. No field in `XMLSchema`; dropped on decode and omitted on encode.
- **`<extraTerrain>`** -- top-level element in the sample (`<extraTerrain>`). No field in `XMLSchema`; dropped on decode and omitted on encode.

All six were confirmed absent from `schema.go` (grep for `hScrollbarPos`,
`vScrollbarPos`, `dropShadow*`, `blurTerrainBG`, `extraTerrain`, and inspection
of `MapLayer_t` / `ShapeStyle_t` returned no match).
