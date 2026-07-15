# ADR 0002 — Version identity: `DataVersion` = `{familyYear, on-disk dotted revision}`

- **Status:** Accepted (2026-07-14), **superseded by [ADR 0004](0004-version-struct-and-release-registry.md) (2026-07-15)** — outcome: **adopt Option 3 (family-dispatch), harmonized with the existing h2025 parse.** Classic decode parses its `version` attribute into `DataVersion.Minor.Patch` with `Major = 2017`; the public encoder dispatches on `DataVersion.Major` (family) only. Classic is being **unfrozen** for this reconciliation.
- **Superseded because:** this ADR defines `Minor.Patch` as parsed from "the single dotted value each format exposes." A W2025 file exposes two (`@version` and `@schema`), so the representation cannot hold the file's identity. ADR 0004 replaces `{familyYear, revision}` with `{App, Schema}` and a supported-release registry. **The verbatim-output guarantee (Decision 2) survives and is carried forward.**
- **Date:** 2026-07-14
- **Context tickets:** #12 (this decision), follow-up from #8 (Track B, B4 — version fidelity), PR #10.
- **Gates:** the code change (classic `decode.go`, `xmlio/encoder.go`, tests, docs) is authorized by this acceptance. **Implemented 2026-07-14** — see "Follow-up task" below (all steps done, suite green).

---

## Context

There are **two notions of "version"** for a classic (H2017) file and they do not
line up (full statement in issue #12):

- `Map_t.MetaData.DataVersion` is a `semver.Version` hardcoded to `{Major:2017,
  Minor:1}` by `xmlio/h2017v1/decode.go`. It is really a **schema-family /
  dispatch key**, not the on-disk version.
- The real on-disk value is `version="1.73"` / `"1.74"` / `"1.77"`. Since B4 it
  is preserved as the **string** `Map_t.Version` and additively in
  `MetaData.Worldographer.Version`, but **no semver** carries `1.73/1.74/1.77`.

`DataVersion` is load-bearing for **encode dispatch**: `xmlio/encoder.go`
`MarshalXML(m, targetVersion)` switches on `targetVersion.Major` (2017/2025) and
requires `Minor == 1`, and every CLI tool (`cmd/copy`, `crop`, `merge`, `resize`,
`import`) plus the round-trip tests pass `m.MetaData.DataVersion` as that target.
B4 could not simply write `1.73` into `DataVersion` (it would fail dispatch and
regress the tools), so B4 kept `{2017,1}` and stored the real value in
`Worldographer.Version` as the minimal, non-regressing fix. This ADR decides the
clean end-state now that the freeze is being lifted for exactly this purpose.

### The key observation: the two decoders are already asymmetric

The h2025 decoder **already does the right thing** and the classic decoder is the
outlier:

- **h2025** (`xmlio/h2025v1/map.go:43-56`) parses the on-disk **schema** string
  into `DataVersion`: schema `"1.06"` → `{Major:2025, Minor:1, Patch:6}`. So
  `Major` = family year, `Minor.Patch` = the real on-disk dotted revision. It
  keeps the string in `Worldographer.Schema`/`Worldographer.Version` too.
- **classic** (`xmlio/h2017v1/decode.go:33`) hardcodes `{2017,1}` and throws the
  real `"1.73"` into a string, purely because B4 was working around the freeze.

So the clean end-state is **not** a new field, a rename, or a family-enum. It is
making classic do exactly what h2025 already does: parse its single dotted
on-disk value (`version="1.73"`) into `DataVersion.Minor.Patch` with
`Major = 2017`.

### The load-bearing coincidence that makes this nearly free

The classic `version` attribute is `"1.73"` / `"1.74"` / `"1.77"` — the leading
component is **`1`** in every known classic file, exactly as the h2025 schema
(`"1.06"`) has leading component `1`. Parsing `"1.73"` the h2025 way yields
`{2017, 1, 73}`: **`Minor` stays `1`.** Therefore the existing encode dispatch
(`Major ∈ {2017,2025}`, `Minor == 1`) and all seven `cmd/*` callers keep working
**unchanged**, while the sub-revision becomes a genuine, recoverable
`DataVersion.Patch`.

## Decision

1. **Canonical representation.** `DataVersion = {Major: familyYear, Minor.Patch:
   on-disk dotted revision}` for **both** codecs. `Major` is the schema-family /
   dispatch key (2017 or 2025); `Minor.Patch` is the true on-disk sub-revision,
   parsed from the single dotted value each format exposes:
   - classic — the `map/@version` attribute (`"1.73"` → `{2017, 1, 73}`);
   - 2025 — the `map/@schema` attribute (`"1.06"` → `{2025, 1, 6}`), unchanged.

   | on-disk | `Worldographer.Version` (string, kept) | `DataVersion` (semver) |
   |---|---|---|
   | classic `1.73` | `1.73` | `{2017, 1, 73}` |
   | classic `1.74` | `1.74` | `{2017, 1, 74}` |
   | classic `1.77` | `1.77` | `{2017, 1, 77}` |
   | 2025 schema `1.01` | `2.06` | `{2025, 1, 1}` |
   | 2025 schema `1.06` | `2.06` | `{2025, 1, 6}` |

2. **The string mirror stays, and is authoritative for output.** `Map_t.Version`
   and `MetaData.Worldographer.Version`/`.Schema` continue to hold the **verbatim**
   on-disk string, and **the encoder emits those strings byte-for-byte** — never a
   value re-derived from `DataVersion`. The semver is the *parsed, comparable*
   form; the string is the *fidelity* form. Nothing is removed — this is still
   additive on the string side.

3. **Encoder dispatches on family (`Major`) only.** `MarshalXML` routes
   `Major == 2017 → h2017v1.Encode` and `Major == 2025 → h2025v1.Encode`,
   treating `Minor.Patch` as informational. This removes the latent `Minor == 1`
   fragility (a future 2025 schema `"2.00"` → `{2025,2,0}` currently fails encode
   silently) and is the "version-range aware" option 3 from the ticket. Because
   `Minor` is `1` for every current file, this is behavior-preserving for all
   existing callers and fixtures.

### On-disk versions are NOT semantic — treat `DataVersion` as an opaque identifier

Worldographer does not use semantic versioning, and the on-disk values carry
known quirks. `DataVersion` is a *parsed, comparable transcription* of the
on-disk identifier — **not** a claim about schema compatibility. Two consequences
this ADR binds:

- **Classic `1.73` → `1.77` is the same schema.** As far as is known the classic
  XML schema did **not** change across `1.73`/`1.74`/`1.77`; those are
  application/file version bumps. So `DataVersion.Patch` distinguishes the
  on-disk revision string, but callers must **not** infer a schema difference
  from it. All classic sub-revisions decode/encode through the one `h2017v1`
  codec.
- **2025 on-disk versions are buggy and must be emitted verbatim.** Worldographer
  2025 shipped from beta with a **buggy version number**: files write a `1.x`
  value (e.g. schema `"1.06"`) where a `2.x` number was intended, and the app
  does not do semver. `wxx` must **re-emit those buggy values exactly as read**
  (round-trip fidelity — a Worldographer-written `"1.06"` must go back out as
  `"1.06"`). This is *why* Decision point 2 makes the preserved string
  authoritative for output and forbids re-deriving the header from `DataVersion`.
  Mapping the buggy on-disk values to the *true* 2025 release version is a
  **separate, deferred** effort (**#13**), not part of this reconciliation.

### Why not the other options

- **Option 1 (keep as-is, document only).** Leaves `DataVersion` as a family
  marker that looks like a version but isn't — the exact overload #12 exists to
  remove. Rejected now that the freeze is lifted.
- **Option 2 (separate family key + true-semver field / rename).** Introduces a
  new dispatch field or renames `DataVersion`, forcing changes in
  `xmlio/encoder.go`, `xmlio/decoder.go`, **every** `cmd/*` caller, and the
  round-trip tests. Larger blast radius for no more clarity than Option 3 buys,
  because Option 3 already yields a proper semver *and* keeps the callers intact.
- **Option 3, but encoding the sub-revision in `Minor`** (`"1.73"` → `{2017,73}`).
  Breaks the `Minor == 1` invariant and diverges from how h2025 parses (where
  `Minor` is the schema-major). The chosen `Minor.Patch` mapping is symmetric
  with h2025 and keeps `Minor == 1`, so it is strictly better.

## Consequences

**Behavioral change (intended, the point of #12):**

- Classic `DataVersion` changes from `{2017,1,0}` to `{2017,1,73/74/77}`. The
  classic sub-revision is now recoverable as a proper semver, not just a string.
- `cmd/*` that print `DataVersion` now show `2017.1.73` instead of `2017.1` —
  strictly more informative, no format break.

**Non-changes (guaranteed by the `Minor == 1` coincidence):**

- Encode dispatch, all seven `cmd/*` callers, and the 2025 round-trip tests are
  untouched at the call site; they keep passing `m.MetaData.DataVersion` and keep
  selecting the same encoder.
- The two decoders become **symmetric** (both `{familyYear, parsed dotted
  revision}`), removing a real inconsistency between the packages.

**Follow-up task (authorized by this ADR) — implemented 2026-07-14, test-first:**

1. ✅ `xmlio/h2017v1/decode.go` — `DataVersion` now comes from a
   `classicDataVersion(m.Version)` helper that parses `"1.7x"` into `Minor.Patch`
   with `Major: 2017` (best-effort; falls back to `{2017,1}` on a malformed
   string). The `Worldographer.Version = m.Version` verbatim copy is retained.
2. ✅ `xmlio/encoder.go` — `MarshalXML` dispatches on `Major` only
   (`2017 → h2017v1.Encode`, `2025 → h2025v1.Encode`).
3. ✅ `xmlio/classic_dispatch_test.go` — `TestClassicVersionFidelity` now asserts
   `DataVersion == {2017,1,73/74/77}` per fixture, and a new
   `TestClassicEncodeDispatch` re-encodes each classic fixture through the relaxed
   public dispatch and re-decodes to confirm the `version` attribute round-trips.
4. ✅ Docs — this ADR, the "Version identity" section of
   `xmlio/h2017v1/COVERAGE.md`, and the README version note describe the unified
   scheme.

`go build ./...`, `go vet ./...`, and `go test ./...` are green.

**Costs:**

- Classic is unfrozen for this reconciliation. Scope is deliberately narrow:
  **version-identity metadata only.** The classic *encode gaps* (ROWS, mapkey,
  shapes, notes, informations, labelstyles — see `xmlio/h2017v1/COVERAGE.md`) are
  **out of scope** for #12 and remain documented gaps unless a separate ticket
  schedules them.

---

## Canonical family ⇄ on-disk mapping (for reference)

| schema family (`DataVersion.Major`) | on-disk identifier | on-disk values seen | `DataVersion` |
|---|---|---|---|
| `2017` (classic / Hexographer 2) | `map/@version` (no `release`/`schema`) | `1.73`, `1.74`, `1.77` | `{2017, 1, <nn>}` |
| `2025` (Worldographer 2025) | `map/@schema` (`release="2025"`) | `1.01`, `1.06` | `{2025, 1, <n>}` |

Family is the dispatch key on both decode (`release` attribute shape) and encode
(`DataVersion.Major`); the on-disk dotted value is the sub-revision carried in
`Minor.Patch` and mirrored verbatim as a string.

Notes on the values (non-semantic — see the caveats above):

- Classic `1.73`/`1.74`/`1.77` are **application** version bumps over the **same**
  schema; `Patch` distinguishes the file, not the schema.
- 2025 on-disk values (`version`, `schema`) are **buggy** (`1.x` where `2.x` was
  intended) and are emitted **verbatim**. Reconciling them to the true 2025
  release version is deferred to **#13**.
