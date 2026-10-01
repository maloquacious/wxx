# ADR 0005 — Remove the classic (H2017 / Worldographer 1.x) format

- **Status:** **Accepted (2026-09-30)** — maintainer decision on
  [#103](https://github.com/maloquacious/wxx/issues/103).
- **Date:** 2026-09-30
- **Context tickets:** [#103](https://github.com/maloquacious/wxx/issues/103)
  (this decision), [#98](https://github.com/maloquacious/wxx/issues/98) (the
  classic encoder drops every shape and note; #103 is its resolution),
  [#92](https://github.com/maloquacious/wxx/issues/92) (registering 2.07, the
  first W2025-to-W2025 target that may lose content).
- **Affects:** ADR 0001, ADR 0003 and ADR 0004 (see *Effect on earlier ADRs*).
  Those ADRs are not edited; this one records how their decisions read now.

---

## Context

Until #103, wxx read and wrote two file-format families: classic (H2017,
Worldographer 1.x: `map/@version` `1.73`, `1.74` or `1.77`, no `@release`, no
`@schema`, XML 1.0) and W2025 (`release="2025"`, `@version` and `@schema`
stated, XML 1.1). Each had its own codec, `xmlio/internal/v0_77` and
`xmlio/internal/v1_06`.

Three things made keeping classic a poor trade:

- **Worldographer converts v1 files to v2 itself.** A user with a classic map
  can convert it in the application, which knows its own format, rather than
  rely on a wxx codec that only approximates it.
- **The classic codec was frozen.** After #98 (the classic encoder silently
  dropped every shape and note) was closed without a fix, the codec would
  receive no further work. Its known gaps — shapes, notes and `<informations>`
  dropped on write, a hard-coded `<mapkey>`, ROWS maps refused — were silent
  data loss that would never be fixed.
- **A frozen codec still costs.** `Map_t` had to stay a superset of both
  formats, so classic-only fields (such as `Note_t.InnerText`) lived beside the
  W2025 model. The downgrade-loss inventory existed only to describe what a
  W2025 map lost when written as classic. Every encoder, test and identity
  guarantee had to be stated twice.

## Decision

1. **wxx reads and writes W2025 only.** The classic codec
   `xmlio/internal/v0_77` (decoder, encoder, schema types and its coverage
   matrix) is deleted, and the classic application versions `1.73`, `1.74` and
   `1.77` are removed from the encoder registry. Asking to encode as one fails
   like any other unregistered version (`wxx.ErrUnsupportedMapVersion`). The
   only registered application version is `2.06`, on the `v1_06` codec.

2. **A classic file is refused on decode, before any of it is decoded.** A file
   with an empty `map/@release` and a `1.x` `map/@version` returns
   `wxx.ErrClassicMap` — `classic (Worldographer 1.x) map: convert it in
   Worldographer 2025 first` — joined with the version it found
   (`map: version "1.77": no release`). A file with no `@release` and any other
   version is refused with `wxx.ErrUnsupportedMapMetadata`. The message is
   deliberately generic: it names no Worldographer menu item, because that
   wording has not been checked in the application. Every command that decodes
   a map prints this refusal; `wxx export`, which only gunzips and transcodes,
   still exports any file.

3. **The loss-reporting channel is kept.** `DroppedFeature_t`,
   `EncoderDiagnostics.Dropped`, `wxx.ErrUnmodeledStubLoss` and the
   `downgradeLoss` contract stay, with the classic inventory removed. The
   principle that a write always tells the caller what it loses still applies
   to W2025-to-W2025 writes, the first of which is #92 (2.07). `downgradeLoss`
   reports no loss for schema `1.06` and returns `wxx.ErrUnsupportedSchemaVersion`
   for a schema with no loss inventory, so a future target cannot silently
   report "no loss".

4. **The classic RelaxNG grammar is kept as reference.** `schema/utf-8-xml.rnc`
   and `schema/utf-8-xml.rng` stay, labelled in `schema/README.md` as describing
   the classic format only, which wxx no longer reads. They still describe the
   elements W2025 shares with classic, and earlier coverage work cites them as
   evidence.

5. **One classic fixture is kept.** `testdata/2017-1.77-1.0-columns-blank.wxx`
   pins the refusal against bytes Worldographer actually wrote
   (`xmlio/classic_refusal_test.go`). Every other classic fixture is deleted.

6. **`Map_t` models W2025 only.** Fields and comments that existed only for
   classic are removed (`Note_t.InnerText`; the "classic codec ignores it"
   notes). The display label for a map stating no schema is now `schema none`,
   not `schema implicit (classic)`.

## Consequences

- A user with a classic map gets a clear error naming the fix, instead of a
  best-effort decode and a lossy write.
- The identity guarantee (a file's declared identity and its content format
  cannot disagree, #41/#45) is tested on W2025 alone: `xmlio/chimera_test.go`
  plants #41's classic identity (`release="" version="1.77" schema=""`) on a
  W2025 map and shows that it never reaches the bytes.
- Proof that W2025 behaviour did not change: for every W2025 fixture, the
  encoder's `2.06` output is byte-identical before and after the removal.
- The module version moves to `0.43.0-alpha`; the removal of the classic
  versions is a breaking API change, which a minor bump covers at `0.x`.
- Re-adding classic would mean restoring the codec from git history and
  re-establishing `Map_t` as a superset of both formats. Nothing here prevents
  that, but nothing is kept to make it cheap.

### Effect on earlier ADRs

The earlier ADRs are the historical record and are not edited. Where their
decisions mention classic, they now read as follows:

- **ADR 0001** (codec file organization) chose to leave the frozen classic
  codec (then `h2017v1`, later `v0_77`) alone and restructure only the W2025
  one (then `h2025v1`, now `v1_06`). The classic codec no longer exists, so the
  exemption is moot; the decision about `v1_06` is unaffected.
- **ADR 0003** (two independent version axes) used classic as its example of a
  file whose schema is implicit. The two-axis model stands; no supported file
  has an implicit schema any longer, so a decoded map always states both axes.
- **ADR 0004**
  - *Decision 2:* `Schema == nil` no longer identifies "the one implicit legacy
    schema". A decoded map never holds nil; only a map a caller built can, and
    it displays as `schema none`.
  - *Decision 3* (already amended by #45): its `""`-for-classic `Release`
    and nil-for-classic `Schema` describe entries that no longer exist.
  - *Decision 4:* the example of classic `1.73`/`1.74`/`1.77` sharing one codec
    no longer applies; the principle that application versions on one schema
    share a codec stands, and is how #92 adds 2.07.
  - *Decision 6* ("`Map_t` stays the superset"): with one format, `Map_t` is the
    union of the W2025 releases wxx supports, not of two families. "Encoding to
    an older target may drop" no longer has a classic target; loss can now
    arise only between W2025 releases.
  - *Decision 7* (downgrade loss must be reported) stands, through the kept
    channel (Decision 3 above). The classic loss inventory it produced, the
    *Terrain layers* open question and the 2026-09-29 amendment about a classic
    downgrade reporting the terrain-layers loss describe a target that no
    longer exists.
