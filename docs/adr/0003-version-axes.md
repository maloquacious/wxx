# ADR 0003 — Application version and schema version are independent axes

- **Status:** Proposed — gates issue #28. Records the canonical conceptual model and
  corrects two doc assertions; the code/fixture work it authorizes lands separately.
- **Date:** 2026-07-14
- **Context tickets:** #1 (schema now has a version), #28 (this decision's follow-up
  work). Builds on ADR 0002 (`0002-version-identity.md`, #12) and touches #13
  (true 2025 release) and #20 (version-identity invariant).

---

## Context

Worldographer files carry **two distinct notions of "version"**, and most of our
prose and one of our fields quietly collapse them into one.

1. **Application version** — the Worldographer build that wrote the file. It is
   present in **every** schema:
   - classic (H2017): `map/@version` (`"1.73"` / `"1.74"` / `"1.77"`);
   - W2025: `map/@release` (major, `"2025"`) plus `map/@version` (minor).
2. **XML schema version** — the on-disk data format the file conforms to. Its
   presence is **not** uniform:
   - classic (H2017): **implicit** — there is no `@schema` attribute; the schema
     is identified by the *absence* of `release`/`schema` plus a `1.x` `version`;
   - **early** W2025: **implicit** — `release="2025"` is present but there is
     **no `@schema` attribute** (reported by the maintainer; not yet captured by
     a fixture — see #28);
   - **later** W2025: **explicit** — `map/@schema` is present (e.g. `"1.06"`).

A single schema version can back **multiple** application versions: classic
`1.73`/`1.74`/`1.77` all share one schema (already noted in `README.md`).

### Where the code and docs stand today

- **Dispatch is already axis-aware and robust.** `xmlio/decoder.go` routes on
  `release` alone: `release="2025"` → `h2025v1`, empty `release` + `1.x`
  `version` → `h2017v1`. Its comment is explicit that "version/schema no longer
  gate the dispatch." So an early-2025 file with no `@schema` **decodes**.
- **Version *identity* is not axis-aware.** ADR 0002 defines a single
  `MetaData.DataVersion` semver whose `Minor.Patch` is parsed from `map/@version`
  for classic (the **application** version) but from `map/@schema` for 2025 (the
  **schema** version). One field, two meanings. This works only because classic
  exposes no separate schema version — and it has no defined behavior for the
  early-2025 case, where there is no `@schema` to parse.
- **The docs assert schema is always present in W2025.** `notes.adoc`
  ("map.schema is only in W2025 files") and the `README.md` version-identity
  table (2025 on-disk identifier listed *as* `map/@schema`) both read as
  *W2025 ⇒ has schema*, which is false for early 2025.

## Decision

1. **Adopt the two-axis model as canonical.** Documentation and any future
   version representation treat **application version** and **schema version** as
   **independent axes**. Schema version is a property that may be **implicit**
   (absent on disk and inferred) or **explicit** (present on disk), independently
   of which application version wrote the file.

2. **`@schema` is optional, not guaranteed, for `release="2025"`.** Presence of
   `release="2025"` establishes the **2025 family** for dispatch; it does **not**
   guarantee a `@schema` attribute. Code that derives schema identity must treat
   a missing `@schema` as "implicit/inferred," never as a parse error or a silent
   `{2025,1,0}`.

3. **Correct the two misleading doc assertions now** (this ADR ships with the
   edits): `notes.adoc` and `README.md` are amended to say `@schema` is present
   only in *later* W2025 files and may be absent (implicit) in early ones.

4. **Defer the identity refactor to #28.** Whether the fix is an inferred/optional
   schema value on the existing `DataVersion`, or splitting application version
   and schema version into separate fields, is decided under #28 — coordinated
   with #13 (true 2025 release) and #20 (invariant guarding). This ADR does not
   pre-commit that shape; it commits only the conceptual model and the
   optional-`@schema` rule.

## Consequences

- The maintainer's "implicit in early 2025" model is now written down and matches
  the decoder's actual (release-only) dispatch behavior.
- A concrete, testable gap is isolated: early-2025 (`release="2025"`, no
  `@schema`) has undefined version identity and no fixture. That is #28's first
  deliverable.
- **Verbatim output stays inviolable** (ADR 0002, Decision 2): we never
  synthesize a `@schema` that was not on input. An inferred schema version, if
  introduced, is metadata for callers to reason with — it must not change what the
  encoder emits.
- No code changes are authorized by this ADR beyond the doc edits; the identity
  work is explicitly gated to #28.
