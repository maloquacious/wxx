# ADR 0003 — Application version and schema version are independent axes

- **Status:** Proposed — gates issue #28. Records the canonical conceptual model;
  the code/fixture work it authorizes lands separately.
- **Date:** 2026-07-14
- **Amended:** 2026-07-15 — the original draft asserted that *early* W2025 files
  omit `@schema` (implicit schema). Every W2025 sample we have contradicts this;
  the claim is **retracted** and the doc edits it authorized are reverted. See
  *Evidence* and *Revision history*.
- **Context tickets:** #1 (schema now has a version),
  [#28](https://github.com/maloquacious/wxx/issues/28) (this decision's follow-up
  work). Builds on ADR 0002 (`0002-version-identity.md`, #12) and touches #13
  (true 2025 release) and #20 (version-identity invariant).

---

## Context

Worldographer files carry **two distinct notions of "version"**, and most of our
prose and one of our fields quietly collapse them into one.

1. **Application version** — the Worldographer build that wrote the file. It is
   on disk in **every** supported schema:
   - classic (H2017): `map/@version` (`"1.73"` / `"1.74"` / `"1.77"`);
   - W2025: `map/@release` (`"2025"`) plus `map/@version` (e.g. `"2.06"`).

2. **XML schema version** — the on-disk data format the file conforms to:
   - classic (H2017): **implicit** — there is no `@schema` attribute; the schema
     is identified by the *absence* of `release`/`schema` plus a `1.x` `version`;
   - W2025: **explicit** — `map/@schema` is present (e.g. `"1.06"`).

So a W2025 file states **both** axes outright: `@version` is the application
version, `@schema` is the schema version, and the two move independently
(`version="2.06"` alongside `schema="1.06"`). Classic states only the
application version and leaves the schema implicit.

A single schema version can back **multiple** application versions: classic
`1.73`/`1.74`/`1.77` all share one schema (already noted in `README.md`).

### Evidence

Every W2025 sample examined carries an explicit `@schema`:

| Sample | release | version | schema | |
|---|---|---|---|---|
| `blank-2025-1.10-1.01.wxx` | 2025 | 1.10 | 1.01 | beta; **never tracked** — maintainer's copy only |
| `testdata/2025-2.06-13x11-941577-blank.wxx` | 2025 | 2.06 | 1.06 | baseline |
| `testdata/2025-2.06-13x11-941577-layers.wxx` | 2025 | 2.06 | 1.06 | baseline |

The 1.10/1.01 sample is the **oldest** W2025 build examined and predates the
first post-beta release (2.06) — and it still writes `@schema`. There is at
present **no evidence that any W2025 build ever omitted `@schema`**, and hence
no implicit-schema case to model for W2025.

The beta sample is deliberately not in the repository (2.06 is the supported
baseline, Decision 3), so this row cannot be re-verified from a clean clone. It
is recorded here because it is the evidence the retraction rests on.

The two axes are nevertheless real and independently observable: the 2.06
samples carry `version="2.06"` alongside `schema="1.06"` — different values that
moved separately. That, not the implicit-schema claim, is what grounds this ADR.

### Where the code and docs stand today

- **Dispatch is axis-aware.** `xmlio/decoder.go` routes on `release` alone:
  `release="2025"` → `h2025v1`, empty `release` + `1.x` `version` → `h2017v1`.
  Its comment is explicit that "version/schema no longer gate the dispatch."
- **Decode still requires `@schema`.** `xmlio/h2025v1/map.go` rejects a W2025
  file with no `@schema` (`missing map.Schema`). Given the evidence above, this
  guard is **correct as written** and no known file trips it.
- **Version *identity* is not axis-aware.** ADR 0002 defines a single
  `MetaData.DataVersion` semver whose `Minor.Patch` is parsed from `map/@version`
  for classic but from `map/@schema` for 2025. One field, two meanings: the
  application axis for classic, the schema axis for 2025. This works only because
  classic exposes no separate schema version. **This conflation is real and
  remains #28's work** — it is unaffected by the retraction above.

## Decision

1. **Adopt the two-axis model as canonical.** Documentation and any future
   version representation treat **application version** and **schema version** as
   **independent axes**. Schema version is a property that may be **implicit**
   (absent on disk and inferred) or **explicit** (present on disk), independently
   of which application version wrote the file.

2. **Treat `@schema` as required for `release="2025"`, pending evidence.**
   Every W2025 sample we hold carries one. `h2025v1` rejecting a W2025 file with
   no `@schema` is the correct behavior and stays. If a W2025 file lacking
   `@schema` is ever produced, reopen this decision rather than inferring a
   schema silently.

3. **W2025 2.06 is the supported baseline** (maintainer, 2026-07-15). It is the
   first post-beta 2025 build and the version we implement against. When a newer
   release appears, target it first and backport; Inkwell has historically been
   backward compatible. Releases may be skipped.

4. **Defer the identity refactor to #28.** Whether the fix is splitting
   application version and schema version into separate fields, or something
   else, is decided under #28 — coordinated with #13 (true 2025 release) and #20
   (invariant guarding). This ADR does not pre-commit that shape; it commits only
   the two-axis model. Both axes are on disk for W2025 and the schema axis is
   derivable for classic, so the refactor has everything it needs: it is
   **unblocked**, and gated only on #28 being picked up.

## Consequences

- The two-axis model is grounded in observation (`version="2.06"` alongside
  `schema="1.06"`) rather than in the retracted implicit-schema claim.
- **#28 loses a deliverable.** Hunting an early-2025 no-`@schema` fixture is
  dropped: no such file is known to exist. What remains is the `DataVersion`
  conflation, and nothing blocks it — a W2025 file carries both axes, so
  separating them is a modeling decision, not an investigation.
- **Verbatim output stays inviolable** (ADR 0002, Decision 2): we never
  synthesize a `@schema` that was not on input.
- No code changes are authorized by this ADR; the identity work is gated to #28.

## Revision history

- **2026-07-15** — Retracted the claim that early W2025 files omit `@schema`.
  The original draft attributed it to maintainer knowledge and flagged it as
  unverified pending a fixture. The fixture hunt instead produced counter-
  evidence: the oldest W2025 sample examined (1.10/1.01, pre-2.06) carries an
  explicit `@schema`, as do both 2.06 samples. Decision 2 is inverted
  accordingly, the `notes.adoc` and `README.md` edits authorized by the original
  Decision 3 are reverted, and the baseline policy is recorded in its place. The
  two-axis model and the `DataVersion` conflation finding are unchanged.
- **2026-07-15** — Removed a note questioning whether `map/@version` is the
  application version, and unblocked Decision 4. The note rested on a pre-2.06
  sample that is out of scope under Decision 3; with the baseline settled, a
  W2025 file states both axes outright (`@version` application, `@schema`
  schema), so the identity refactor is a modeling decision with no open
  investigation in front of it.
