# Classic H2017 (v0_77) codec coverage

Per-element read/write coverage for the classic Worldographer / Hexographer 2
XML codec (`xmlio/internal/v0_77`), covering on-disk `version="1.73"`, `"1.74"`, and
`"1.77"` files. This uses the **same format** as `xmlio/internal/v1_06/COVERAGE.md`
(issue #7) so the two per-version matrices read as one artifact, and it exists
for the same reason: to make **stub-drift** visible. Classic decode is nearly
complete, but the encoder still contains several silent stubs — this matrix
records exactly which, with a code citation for every claim.

Classic (v0_77) is **frozen** per `README.adoc` (security fixes only, no new
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

The classic codec's first automated test is the round-trip loss harness
`xmlio/roundtrip_2017_test.go` (issue #25), covered in "Round-trip loss inventory
(executable)" below. It is an **on-disk audit** — it asserts *what the classic
codec drops/alters* on a decode → encode round trip per fixture — not a
per-field fidelity check, and the `xmlio/internal/v0_77/` package itself still has no
`_test.go`. So the per-element statuses in the matrix below remain cited to a
**source line** (`decode.go` / `encode.go`) rather than to a test, and to a
decoded classic sample where a fixture proves the shape. Decoded samples referenced are
the classic fixtures under `testdata/` (`blank-2017-1.73/1.74/1.77-1.0.wxx`,
`2017-1.77-1.0-{columns,rows}-blank.wxx`, `2017-1.77-1.0-{import,merge-01}.wxx`),
inspected by decompressing the gzip/UTF-16BE container to UTF-8 XML.

| `<map>` child element | Decode | Encode | Evidence | Notes |
|---|---|---|---|---|
| `<map>` root + scalar attributes (24) | implemented | implemented | `decode.go:38-81`, `encode.go:27-59` | All 24 modeled attrs round-trip. `version` attr (`1.73`/`1.74`/`1.77`) is preserved verbatim in `Map_t.Version` (re-emitted) and `MetaData.Worldographer.Version`, and parsed into `MetaData.Version.App` with a nil `Schema` (ADR 0004); the encoder resolves the codec from that schema — see "Version identity" below. |
| `<gridandnumbering>` (30 attrs) | implemented | implemented | `decode.go:78-109`, `encode.go:110-145` | All 30 attributes modeled and re-emitted. |
| `<terrainmap>` | implemented | implemented | `decode.go:111-128`, `encode.go:147-158` | Tab-delimited name/slot table parsed into `TerrainMap_t`; re-emitted sorted by slot. |
| `<maplayer>` (`name`, `isVisible`) | implemented | implemented | `decode.go:130-132`, `encode.go:161-176` | Classic `<maplayer>` has no `opacity` attr (confirmed by RelaxNG + samples), so nothing is dropped. |
| `<tiles>`/`<tilerow>` — COLUMNS | implemented | implemented | `decode.go:151-245`, `encode.go:191-201` | COLUMNS (`OddQ`) grid fully round-trips. |
| `<tiles>`/`<tilerow>` — ROWS | implemented | **unimplemented(dropped)** | decode `decode.go:146-163`; encode `encode.go:202-203` | **ROWS decode works** (`OddR` branch); **ROWS encode returns a hard error** `assert(orientation != "ROWS")`. `2017-1.77-1.0-rows-blank.wxx` decodes but cannot be re-encoded. This remains a **documented classic gap by decision** (issue B4): the corresponding gap in h2025 has been closed (h2025 ROWS encode is now implemented, guarded by `TestW2025RowsRoundTrip`), but classic ROWS encode is intentionally left unimplemented since classic is frozen. |
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

## Round-trip loss inventory (executable)

This section is the **on-disk round-trip loss** view that complements the
per-element matrix above. The matrix records the encode gap for each element in
isolation; this section records what those gaps actually cost on a full
**decode → encode** round trip, measured at the XML level (original UTF-8 XML
vs. re-encoded UTF-8 XML), because a `Map_t`-level comparison is structurally
blind to symmetric decode/encode drops.

It is **guarded by `xmlio/roundtrip_2017_test.go`** — the first automated h2017
codec test. That harness decodes each classic fixture, re-encodes it, and diffs
the two documents at the element/attribute-set level (normalizing away attribute
order, whitespace, self-closing form, and numeric formatting). The per-fixture
loss set below is asserted in `classicRoundTripExpect`; any drift (a newly
dropped/altered field, or a previously dropped field that starts surviving)
fails the test so this inventory must be updated deliberately. Run
`go test ./xmlio/ -run RoundTrip2017 -v` to regenerate the raw per-fixture loss
dump.

**Loss vocabulary:** `dropped` (element present in input, absent/emptied in
output) · `altered` (attribute present in both, value changed) · `text-dropped`
(element text present in input, not output) · `encode-hard-error` (re-encode
returns an error; round trip impossible).

### Observed-in-fixture (harness-proven on the samples)

| Path | Classification | Fixtures | Encode citation |
|---|---|---|---|
| `map/informations/information` (+ nested `/information` to depth 2–3) | dropped | all 7 encodable fixtures | `encode.go:438-442` (`encodeInformations` emits only an empty `<informations>` wrapper) |
| `map/configuration/text-config/labelstyle` | dropped | all 7 encodable fixtures | `encode.go:483-507` (`encodeLabelStyle` is a commented-out no-op; `<text-config>` wrapper still emitted, empty) |
| `map/mapkey` `@viewlevel` (`"null"` → `"WORLD"`) | altered | `blank-1.73`, `blank-1.77`, `import`, `merge-01`, `merge-02` | `encode.go:253-258` (`encodeMapKey` writes a hardcoded constant `<mapkey>` block) |
| `<tiles>`/`<tilerow>` (ROWS) | encode-hard-error | `2017-1.77-1.0-rows-blank.wxx` | `encode.go:202-203` (`assert(orientation != "ROWS")`) |

Notes on the observed set:

- **`<informations>` nesting depth is fixture-specific.** `blank-1.74` and
  `2017-1.77-1.0-columns-blank` carry the lore tree only two `<information>`
  levels deep; the other five encodable fixtures reach three levels. The harness
  records the exact deepest chain per fixture, so the expectation for the two
  shallower fixtures omits the depth-3 entry. Only the `<information>` children
  are lost — the `<informations>` wrapper itself survives (emitted empty).
- **`@viewlevel` is the *only* observed `<mapkey>` alteration.** The encoder
  discards `Map_t.MapKey` entirely and writes a constant block, but the blank
  samples' map keys coincide with that block on all 22 other attributes, so only
  `viewlevel` (input `"null"`, constant `"WORLD"`) shows as altered.
  `blank-1.74` and `columns-blank` already carry `viewlevel="WORLD"` on disk, so
  they show **no `<mapkey>` drift at all**. The full constant-block override of
  the remaining attributes is real but **latent** (below).
- **ROWS is a round-trip *failure*, not a silent drop.** The fixture decodes
  successfully (asserted by `TestRoundTrip2017RowsHardError`), but re-encode
  returns `assert(orientation != "ROWS")`, so no diff is possible. Classic ROWS
  encode is intentionally left unimplemented (classic is frozen; issue B4).

### Latent-by-code (encode gap is real, but no classic fixture exercises it)

These drops are proven by the encoder source, not by the samples — every
classic fixture leaves the relevant element empty or matches the constant block,
so the harness cannot observe them. They are recorded from the encode-side
citation and would surface the moment a populated map is round-tripped.

| Path | Classification | Why latent | Encode citation |
|---|---|---|---|
| `map/mapkey` (22 non-`viewlevel` attributes) | altered (constant-block override) | sample map keys equal the hardcoded constants | `encode.go:253-258` |
| `map/shapes/shape` (+ `<p>` points) | dropped | no classic fixture contains a populated `<shape>` | `encode.go:389-411` (`encodeShape` is a no-op) |
| `map/notes/note` (+ text) | dropped | every classic fixture has an empty `<notes>` element | `encode.go:424-436` (`encodeNote` is a no-op) |
| `map/configuration/terrain-config`, `feature-config`, `texture-config` inner content | dropped (no-op wrapper) | samples leave these three sub-configs empty | `encode.go:465-481` |
| `map/labels/label` / `map/features/feature/label` `@backgroundColor` (`0.0,0.0,0.0,1.0` sentinel) | dropped (by design) | no sampled label carries the sentinel background | `encode.go:343-345` (omitted intentionally; see the code comment) |

### What round-trips cleanly (harness confirms *no* loss)

The harness reports **zero** losses for these, across all seven encodable
fixtures, corroborating the "implemented/implemented" rows of the matrix at the
on-disk level: `map` root + its 24 scalar attributes, `<gridandnumbering>`,
`<terrainmap>`, `<maplayer>`, `<tiles>`/`<tilerow>` + tile data (COLUMNS),
`<features>`/`<feature>` (+ `<location>`, inline `<label>`), standalone
`<labels>`/`<label>`, and `<configuration>`'s `<shape-config>`/`<shapestyle>`.
None of these appear in any fixture's loss set.

## Downgrade loss inventory — W2025 → classic (executable)

The two sections above are about what the classic codec loses **to itself**. This
one is about what the classic **format** cannot hold at all, which is a different
claim and must not be confused with it: a codec gap is our encoder not writing
something classic has room for, while a downgrade loss is content classic has
nowhere to put. Only the second is reported as loss by `xmlio/downgrade.go`
(#32, ADR 0004 Decision 7); reporting the first would blame the format for our
encoder.

It was derived with the same harness, by encoding a decoded W2025 2.06 fixture
through the classic target and diffing against the W2025 original, then
subtracting two controls: a **2.06 → 2.06** trip (which isolates h2025 codec
gaps) and the **classic → classic** trip above (which isolates classic codec
gaps). Guarded by `TestClassicDowngradeLossInventory` in `xmlio/downgrade_test.go`,
which re-runs that diff and fails **both** if the encoder reports a loss the
harness does not show and if the harness shows one the encoder does not report.

**Contract:** a **modeled** loss is reported through `EncoderDiagnostics.Dropped`
and the encode succeeds; an **unmodeled stub** is a hard error, because the
encoder cannot describe what such a loss costs. When a stub becomes modeled (#34
tracks the one stub this rule currently bites), its error becomes a diagnostic.

| Path | Class | Evidence on `2025-2.06-13x11-941577-blank.wxx` | Why classic cannot express it |
|---|---|---|---|
| `map/maplayer/@opacity` | modeled → diagnostic | harness: `attr-dropped map/maplayer opacity`; 8 layers at `1.0` | RelaxNG `maplayer` states only `@name`/`@isVisible` (lines 63-66) |
| `map/configuration/shape-config/shapestyle/@lineCap` | modeled → diagnostic | harness: `attr-dropped …/shapestyle lineCap` (`SQUARE`) | RelaxNG `shapestyle` has 27 attrs, no `@lineCap`; classic defines it on `<shape>`, a **different element** (line 157) |
| `map/configuration/shape-config/shapestyle/@lineJoin` | modeled → diagnostic | harness: `attr-dropped …/shapestyle lineJoin` (`ROUND`) | as above (line 158) |
| `map/blurTerrainBG` | modeled → diagnostic | harness: `element-dropped map/blurTerrainBG`; 6 real attrs | classic defines no `<blurTerrainBG>` |
| `map/@hScrollbarPos`, `map/@vScrollbarPos` | modeled → diagnostic, **latent** | harness shows both `attr-dropped`, but **both fixtures carry `0.0`** | classic `<map>` states no scrollbar position |
| `map/extraTerrain` | **unmodeled stub → hard ERROR** | `…-layers.wxx` carries 183 bytes (`<mapLayer name="Terrain Layer">`/`<terrainAndLocation>`); `…-blank.wxx` carries `"\n"` | classic defines no `<extraTerrain>`; classic binds `mapLayer` to features/labels/shapes but **never to tiles**, so per-hex layer assignment collapses (ADR 0004) |

Notes:

- **The scrollbar entry is latent.** `Map_t` models both as plain `float64`, so
  absent and `0.0` are the same value and a zero cannot be reported as a loss
  without inventing one. Both tracked fixtures carry `0.0`, so no fixture
  demonstrates it; `TestClassicDowngradeScrollbarLatent` **synthesizes** a
  non-zero source rather than pretending one does, mirroring
  `TestW2025LabelStyleDropShadowGate`.
- **`<extraTerrain>` emptiness.** The error fires on non-whitespace `InnerXML`
  only. `…-blank.wxx`'s container holds `"\n"` — pretty-printer whitespace, in
  which no element, attribute, or text node can hide — so it loses nothing and
  must not error.
- **Not in this table, deliberately.** `map/features/feature/label/@dropShadow*`
  is dropped on a **2.06 → 2.06** trip too (`Map_t.Label_t` models no drop
  shadow; the trio lives on `LabelStyle_t`), so it is an **h2025 codec gap**, not
  a downgrade loss. `map/@version` altered and `map/@release`/`@schema` dropped
  are **target identity** (`Release_t.identify`), not loss.
- **Masked, and therefore unclaimed.** Classic `<labelstyle>` has no
  `@dropShadow*` (RelaxNG lines 181-190), so a W2025 label style's drop shadow
  *is* beyond the classic format — but the classic encoder drops the entire
  `<labelstyle>` element as a codec gap, so the harness cannot separate the two
  and the downgrade half is **not** claimed here.

## Version identity — `MetaData.Version` and `Worldographer.Version`

**Implemented (ADR 0004, issue #32).** `decode.go` parses the on-disk `version`
attribute into `MetaData.Version` via `classicVersionIdentity(m.Version)`: `App`
is the dotted on-disk revision (`1.73`/`1.74`/`1.77`) and `Schema` is **nil**,
because a classic file states no `@schema` at all. That absence is not a gap — it
identifies the one **implicit legacy schema** every classic revision shares. The
classic XML schema did **not** change across `1.73`→`1.77` (these are application
version bumps), which is the evidence for treating them as one schema; all
sub-revisions share the one `v0_77` codec.

The nil `Schema` is what routes an encode back here: `xmlio/encoder.go`
`MarshalXML` resolves the codec from the target release's schema
(`CodecForSchema`), so the implicit legacy schema selects `v0_77.Encode`. There
is no family-year dispatch key any more — the `2017` in this package's name is a
project coinage that appears in no classic file (ADR 0004).

The on-disk revision is preserved in two forms: the **parsed, comparable**
`Dotted` in `MetaData.Version.App`, whose `Raw` is authoritative, and the
**verbatim** string in `Map_t.Version` / `MetaData.Worldographer.Version`
(re-emitted byte-for-byte by the encoder). Classic files carry no
`release`/`schema` attributes, so `Worldographer.Release`/`.Schema` stay empty.
Guarded by `TestClassicVersionFidelity` and `TestClassicEncodeDispatch`
(`xmlio/classic_dispatch_test.go`), and the XML declaration classic files open
with (`<?xml version='1.0'`) is bound to the release entry and pinned at the byte
level by `TestEncodeXMLHeaderFollowsRelease` (`xmlio/encode_dispatch_test.go`).

## Dispatch symmetry — public decoder reads classic (backfilled)

The public `xmlio` pipeline now **round-trips classic end to end** (issue B4
backfill); it was previously write-only:

- **Encode**: `xmlio/encoder.go` `MarshalXML` resolves the codec from the target
  release's schema, so the implicit legacy schema (`Schema == nil`) selects
  `v0_77.Encode` (ADR 0004). The public encoder can emit classic XML.
- **Decode**: `xmlio/decoder.go` now has a classic dispatch case alongside the
  `release="2025"` case. Classic files carry **no `release` attribute at all**
  (confirmed by every sample and by the RelaxNG schema, which defines no
  `release`), so the dispatcher routes them by the classic version shape:
  `Release == "" && strings.HasPrefix(Version, "1.") → v0_77.Decode`. The
  predicate is deliberately conservative — anything that is neither `release=2025`
  nor a `1.x` classic version still falls through to `ErrUnsupportedMapMetadata`,
  so unknown/future formats are not silently swallowed.

So `xmlio.NewDecoder().Decode(classicFile)` now succeeds. Guarded by
`TestClassicDispatch_Decode` (`xmlio/classic_dispatch_test.go`), which decodes the
1.73 / 1.74 / 1.77 fixtures through the full public gzip/UTF-16BE/XML pipeline.
The dispatch asymmetry recorded in earlier revisions of this matrix is resolved.

## Known un-modeled fields

**None found in the available classic samples.** Every attribute present on every
element across all seven classic fixtures is modeled by a field in
`xmlio/internal/v0_77/schema.go`. This was verified by sweeping the decoded UTF-8 XML of
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

Every element the RelaxNG schema defines is **modeled** in `xmlio/internal/v0_77/schema.go`:

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
