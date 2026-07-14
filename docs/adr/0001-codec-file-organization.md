# ADR 0001 — Codec file organization: co-located per-element encode/decode

- **Status:** Proposed (awaiting maintainer decision — Accepted / Rejected)
- **Date:** 2026-07-13
- **Context tickets:** #8 (this task, B3), #7 (Track A drift), B2 (classic stub audit)
- **Gates:** B3b (the actual refactor) runs **only** if this ADR is Accepted with an "adopt" outcome.

---

## Context

### The drift problem

Each version codec in this repo is split into three files by *kind of code*, not
by *element*:

```
xmlio/h2017v1/{schema.go, decode.go, encode.go}
xmlio/h2025v1/{schema.go, decode.go, encode.go}
```

`decode.go` and `encode.go` are large and grow in lockstep with the format:

| file | lines | bytes |
|---|---|---|
| `xmlio/h2017v1/decode.go` | 578 | 22 KB |
| `xmlio/h2017v1/encode.go` | 686 | 27 KB |
| `xmlio/h2017v1/schema.go` | 438 | 17 KB |
| `xmlio/h2025v1/decode.go` | 622 | 24 KB |
| `xmlio/h2025v1/encode.go` | 826 | 35 KB |
| `xmlio/h2025v1/schema.go` | 469 | 18 KB |

The two directions also have **different shapes**, which widens the gap:

- **Decode** is one monolithic `func Decode(input []byte) (*wxx.Map_t, error)`
  (verified: `grep '^func' decode.go` yields only `Decode` in both packages). It
  does a single `xml.Unmarshal` into the per-version `XMLSchema` structs, then
  copies fields into `Map_t` inline. A given element's decode logic is a *section*
  of that one function (e.g. classic `<mapkey>` decode is `decode.go:247-279`).
- **Encode** is ~30 small `encodeXxx(*wxx.Foo_t, *bytes.Buffer)` functions
  (e.g. `encodeMapKey`, `encodeShape`, `encodeNote`) that hand-write XML into a
  buffer.

So for any single `<map>` child element, its parse code and its format code live
in **two different files, in two different styles, often hundreds of lines
apart**. Nothing structural forces them to stay in sync, and nothing makes a
missing half visible at a glance.

### The evidence that this causes real bugs

This is not hypothetical. The split has already produced silent, shipped drift —
an encoder half that quietly does nothing while the decoder half works and a
round-trip test passes:

- **Track A (#7).** The whole of ticket #7 existed because a **stub W2025
  encoder** hid behind a working decoder and a green round-trip test. The
  round-trip passed because decode and encode dropped the same data
  *symmetrically*; the on-disk fidelity was gone. See
  `xmlio/h2025v1/COVERAGE.md` (the "Known un-modeled fields" section documents
  six W2025 fields still dropped symmetrically today, e.g. `maplayer/@opacity`,
  `blurTerrainBG`, `extraTerrain`).

- **Classic stubs (B2).** The classic encoder has the *same* drift pattern in
  several places, now catalogued with source-line citations in
  `xmlio/h2017v1/COVERAGE.md`. Decode reads the data into `Map_t`; encode emits
  nothing:
  - `encodeShape` — commented-out no-op; `<shapes>` written with no children.
  - `encodeNote` — commented-out no-op; `<notes>` written empty.
  - `encodeInformations` — emits only an empty `<informations>` wrapper; the
    entire lore tree (14–86 entries in real samples) is dropped on write.
  - `encodeLabelStyle` — no-op; `<text-config>` written with no `<labelstyle>`
    children (samples carry ~10).
  - `encodeMapKey` — ignores `Map_t.MapKey` and writes a hardcoded constant
    `<mapkey ...>` block.
  - ROWS tile encode — hard `assert(orientation != "ROWS")` error even though
    ROWS decode works.

  These are invisible today only because the classic *sample fixtures* happen to
  leave those sections empty — the loss is latent for any populated map.

- **No classic safety net.** There is currently **no automated h2017 codec
  test** at all (verified: the only `*_test.go` files are `hexg/*_test.go` and
  `xmlio/roundtrip_2025_test.go`). Classic drift is caught by nobody.

The common root cause across all of these: **the two halves of one element's
codec are far apart, so a stubbed or diverged half is not obvious next to its
working counterpart.**

### What wxconv actually does (the reference "style guide")

The task framed wxconv (`~/Jetbrains/worldographer/wxconv/`) as "co-locating each
element's parse+format." Reading the actual source, that framing is **not
literally accurate**, and this ADR records the truth so the maintainer decides on
facts:

- wxconv does **not** co-locate per element. Its `adapters/` directory is
  organized **by pipeline stage / direction**, and each stage is a *single large
  file*:
  - `adapters/wxml_to_wmap.go` (22 KB) — the entire decode adapter
    (`WXMLToWXX` → `wxmlV173ToWXX`), raw XML → domain, all elements in one file.
  - `adapters/wmap_to_tmap.go` (18 KB) — domain → all-string encode model.
  - `adapters/wmap_to_wxx_encoder.go` (19 KB) — domain → encoder, all elements.

  So on the co-location axis, wxconv is *the same direction-split we already
  have*, just at adapter granularity. Co-locating per element is a **distillation
  of the lesson**, not a copy of wxconv's layout.

The parts of wxconv that **are** genuinely distinctive and verified:

1. **A 3-layer model split** (`models/`):
   - `models/wxml173/map.go` — *raw-XML* structs with `xml:` tags and typed
     fields (`int`/`float64`); version-specific ("read a v1.73 file"). The README
     intends this type to be package-internal (it lives behind an `internal`
     folder in `domains/wxx`).
   - `models/wxx/map.go` — the *domain* model with `json:` tags; version-neutral
     superset. This is the analogue of our `Map_t`.
   - `models/tmap173/tmap.go` — an *all-string encode model*: **every field is a
     pre-formatted `string`**, serialized by a `text/template`
     (`tmap173/encode.go` executes the embedded `xml.gohtml` template).

2. **An explicit version-dispatch table with a typed error**
   (`adapters/wmap_to_wxx_encoder.go`):
   ```go
   func WMAPToWXXEncoding(m *wxx.Map, appVersion string) ([]byte, error) {
       t, ok := map[string]func(*wxx.Map, string) (wrx.Encoder, error){
           "1.73": adaptWMAPToWXXEncoder_1_0_0,
       }[appVersion]
       if !ok {
           return nil, ErrUnsupportedVersion
       }
       ...
   }
   ```
   Decode mirrors it with a type switch in `WXMLToWXX` that returns
   `ErrUnsupportedWXMLVersion`. Note wxconv only ever wires **v1.73** (one table
   entry, one switch case; everything else errors).

For comparison, our repo already has the equivalent dispatch, just spread across
two files:

- **Decode** dispatch: `xmlio/decoder.go:233-243` routes `release="2025"` →
  `h2025v1.Decode`, and returns `ErrUnsupportedMapMetadata` otherwise. **There is
  no classic case** — the public decoder does not route H2017 files at all today.
- **Encode** dispatch: `xmlio/encoder.go:153-167` `MarshalXML` switches on
  `version.Major` (2017 → `h2017v1.Encode`, 2025 → `h2025v1.Encode`) and returns
  `ErrUnsupportedSchemaVersion` otherwise.

So of wxconv's three ideas, we already have (2) the dispatch table (in a
different shape) and effectively (part of) the 3-layer split — a per-version raw
`XMLSchema` and a domain `Map_t`. We do **not** have the third layer (the
all-string `tmap` + template encoder), and we do **not** co-locate per element.

---

## Decision / Recommendation

**Recommended default: adopt Option (b) — co-locate per-element encode/decode —
for `xmlio/h2025v1` ONLY. Do not touch `xmlio/h2017v1` (frozen). Do not adopt the
full 3-layer refactor (Option c).**

Rationale (smallest change that kills the drift):

1. **The drift is a co-location problem, so fix co-location — nothing more.** The
   Track A and B2 failures were all "the other half is far away and got stubbed."
   Putting one element's decode and encode in the same small file makes a stub or
   asymmetry obvious in review and while editing. That directly targets the cause.

2. **Honor the frozen-classic constraint.** `README.adoc` freezes h2017v1
   (security fixes only). Re-organizing 1.3 KLOC of frozen code buys nothing
   (its stubs are already documented and, per policy, will not be filled), churns
   a version we've promised to leave alone, and — critically — **has no automated
   test to catch a mechanical slip** (there is no h2017 codec test). The
   risk/reward on classic is strictly negative. Leave it exactly as-is.

3. **h2025v1 is the active codec and has a real safety net.** The round-trip and
   coverage tests in `xmlio/roundtrip_2025_test.go`
   (`TestW2025RoundTrip`, `TestW2025PublicRoundTrip`, `TestW2025PopulatedRoundTrip`,
   `TestW2025PopulatedPublicRoundTrip`, `TestW2025Decode_BothSamples`,
   `TestW2025DecodePopulated`, `TestW2025ConfigSectionsEmpty`) are green and
   exercise both the in-memory codec and the full transport pipeline over both a
   blank and a populated fixture. That is exactly the net a mechanical move needs.

4. **Reject Option (c) as over-scoped.** The all-string `tmap` + template layer is
   a *different encoding strategy* than our buffer-based `encodeXxx` functions;
   adopting it is a rewrite, not a reorganization, and it would have to be done
   twice (or leak into frozen classic). We already have two of wxconv's three
   layers. The third does not pull its weight here, and Option (c) does not more
   effectively kill drift than (b) does.

### Safety net (binding conditions if Accepted → B3b)

If the maintainer accepts "adopt (b) for h2025v1," B3b **must** obey all of:

1. **Zero behavior change.** The move is *pure code motion* — cut a function/
   section, paste it into a new file, fix nothing else. No renames of exported
   symbols, no signature changes, no "while I'm here" fixes.
2. **The green tests stay green throughout.** Run `go test ./...` (plus
   `go build ./...`, `go vet ./...`) **after each element is moved**, not just at
   the end. Any red = revert that step.
3. **Element by element, one commit per element** (or per small group), so any
   regression bisects to a single mechanical move.
4. **Package stays `h2025v1`.** Co-location is *intra-package* file splitting;
   the public API (`h2025v1.Decode`, `h2025v1.Encode`) and the dispatch in
   `xmlio/decoder.go` / `xmlio/encoder.go` are untouched.

### The decode wrinkle B3b must plan for (honest cost note)

Encode already is per-element functions, so moving those is trivial cut/paste.
**Decode is not** — it is one monolithic `Decode` with a single `xml.Unmarshal`
and inline field copies sharing local scope. To co-locate an element's decode
next to its encode, B3b must first *extract* that element's decode section into
its own function. That extraction is the one place a behavior change could sneak
in (shared locals, ordering, the single `Unmarshal` call). It is still low risk
under the safety net above, but it makes (b) a genuine small refactor of the
decode side, not a pure paste. B3b should keep the single top-level `Unmarshal`
in a `decode.go`/`map.go` and have it call per-element `decodeXxx(...)` helpers
that live in the element files.

Expected shape after B3b (illustrative, h2025v1):

```
xmlio/h2025v1/
  schema.go        # unchanged: all XMLSchema structs (or optionally also split)
  map.go           # Decode/Encode entry points + <map> root attrs
  gridandnumbering.go   # decode + encode for <gridandnumbering>
  terrainmap.go
  maplayer.go
  tiles.go         # <tiles>/<tilerow>/tile data
  mapkey.go
  features.go      # <features>/<feature>/<location>/inline <label>
  labels.go
  shapes.go
  notes.go
  informations.go
  configuration.go # + the five sub-configs, labelstyle, shapestyle
  helpers.go       # boold/bools/floatd/floatf/rgbans/terrainMapToSlice/encodeInnerText/...
```

~12–14 element files + a shared `helpers.go`. The shared numeric/RGBA/text
helpers (`boold`, `bools`, `floatd`, `floats`, `rgbans`, `rgbas`,
`terrainMapToSlice`, `encodeInnerText`, …) move once into `helpers.go`.

### Alternative / complementary lighter net (no file move)

There is a cheaper mitigation that attacks the *symptom* without any
reorganization: **a coverage-assertion test for h2025 (and, if desired,
h2017).** It would mechanically assert, per element, that what decode read is
what encode wrote — or, for the documented `no-op(intentional)` / `lossy` cells,
assert exactly that status so an *un*expected change trips the test. This is the
`COVERAGE.md` matrices turned into executable assertions.

- As a **complement** to (b): adopt both — the file move prevents new drift from
  being written; the test fails loudly if it happens anyway. Best coverage.
- As a **substitute** for (b): if the maintainer prefers zero churn, add only the
  coverage-assertion test and mark this ADR "defer." It does not make the code
  easier to *read* the way co-location does, but it does close the "silent stub"
  hole that motivated the whole exercise, and it is the *only* option that can
  also protect frozen classic without editing it.

Honest tradeoff: co-location improves human review (drift is visible while
editing); the assertion test improves machine detection (drift is caught in CI —
except we have no CI, so it is caught on `go test ./...`). They are strongest
together.

---

## Consequences

**If Accepted (adopt b for h2025v1):**

- h2025v1 becomes ~14 small, per-element files; each element's read and write
  sit side by side, making future stubs obvious in review.
- Classic stays frozen and untouched; the asymmetry between the two version
  packages' file layouts is an accepted, documented cost (classic is legacy).
- B3b is unblocked as a purely mechanical, test-guarded task; the decode
  extraction is the only non-trivial part and is fenced by the round-trip tests.
- No public API change, no behavior change, no new dependency.

**If Rejected / Deferred:**

- No code changes. The drift risk remains, mitigated only by the `COVERAGE.md`
  matrices (documentation, not enforcement). Strongly consider adopting at least
  the coverage-assertion test in that case, since it is the cheapest thing that
  actually *enforces* no-new-drift.

**Costs either way:**

- Option (b) touches an actively developed file set; any in-flight h2025 work
  should land or rebase around B3b to avoid churn collisions.
- The coverage-assertion test adds maintenance: it must be updated when an
  element's intended status changes (that is the point — the update is the
  review checkpoint).

---

## Options considered

### (a) Keep split files — status quo

`schema.go` / `decode.go` / `encode.go` per version.

- **Pros:** zero work; familiar; `go` tooling and existing diffs unaffected;
  no risk to frozen classic.
- **Cons:** it *is* the drift surface. Track A and the B2 classic stubs both grew
  here. An element's two halves stay far apart and can silently diverge.
- **Migration cost:** none.
- **Risk:** none to adopt, but leaves the demonstrated failure mode in place.

### (b) Co-locate per element (recommended, h2025v1 only)

One file per `<map>` child element inside the version package, holding that
element's decode + encode pair; shared helpers in `helpers.go`; `schema.go`
optionally left whole.

- **Pros:** directly fixes the co-location root cause; a stubbed/asymmetric half
  is visible next to its partner; small files; intra-package only (no API
  change); scoped to the active codec so frozen classic is untouched.
- **Cons:** decode must be extracted from its monolith first (small refactor, not
  pure paste); two version packages end up with different file layouts; touches
  an active file set.
- **Migration cost:** moderate-low for h2025v1 — ~14 files; encode moves are
  trivial; decode needs per-element extraction from one `Unmarshal`-based
  function. Estimate a focused day, dominated by the decode extraction and by
  running the test suite after each step.
- **Risk:** low **under the safety net** (pure motion, tests after each element,
  one commit per element). The single elevated spot is decode extraction; the
  round-trip + populated + coverage tests cover it.

### (c) Full wxconv-style 3-layer refactor

Adopt raw-XML vs. domain vs. all-string encode model + template, plus an
adapters layer, mirroring wxconv.

- **Pros:** clean separation of concerns; the all-string `tmap` layer makes
  formatting explicit; matches the sibling repo's mental model.
- **Cons:** a *rewrite*, not a reorganization — the template-based encoder is a
  different strategy than our buffer `encodeXxx` functions; we already have two
  of the three layers (`XMLSchema` + `Map_t`), so the net new value is mostly the
  string layer, which is not worth the churn; would have to be built for h2025
  and would tempt churn into frozen classic; does not kill drift any better than
  (b). Note wxconv itself only ever wired v1.73 and still keeps each direction in
  one large file — so it does not even demonstrate the per-element win.
- **Migration cost:** high — new packages, a template engine for encode, and
  re-validation of every element's on-disk bytes.
- **Risk:** high — large behavior-changing surface, weakest fit to the frozen
  constraint, highest chance of new bugs for the least additional drift
  protection.

---

## Decision the maintainer must now make

Pick one:

1. **Adopt (b) for h2025v1** — mark this ADR *Accepted*; B3b executes the
   mechanical, test-guarded per-element move on h2025v1 only. (Recommended.)
2. **Adopt (b) *and* add the coverage-assertion test** — strongest drift
   protection. (Recommended if the appetite exists.)
3. **Adopt (c)** — full 3-layer refactor. (Not recommended; over-scoped, worst
   fit to frozen classic.)
4. **Defer** — no code change; close at this ADR. Consider adding *only* the
   coverage-assertion test, which is the one option that can also guard frozen
   classic without editing it.

Whatever is chosen, `xmlio/h2017v1` should **not** be structurally refactored
while it remains frozen.
