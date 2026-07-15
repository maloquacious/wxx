# ADR 0004 — Version identity: `{App, Schema}` plus a supported-release registry

- **Status:** Proposed — gates issue #28. Records the model and the encoder
  contract; the code change lands separately under #28.
- **Date:** 2026-07-15
- **Supersedes:** ADR 0002 (`0002-version-identity.md`, #12). Builds on ADR 0003
  (`0003-version-axes.md`, the two-axis model).
- **Context tickets:** [#28](https://github.com/maloquacious/wxx/issues/28)
  (this decision), #13 (true 2025 release), #20 (version-identity invariant).

---

## Context

ADR 0003 established that a file carries two independent version axes. ADR 0002
models identity as a single `semver`: `DataVersion = {Major: familyYear,
Minor.Patch: on-disk dotted revision}`. Three things have made that untenable.

**0002's premise expired.** It defines `Minor.Patch` as parsed from "the single
dotted value each format exposes." A W2025 file exposes **two**: `@version`
(`"2.06"`) and `@schema` (`"1.06"`). One slot cannot hold both, so `@version` is
dropped from identity entirely and survives only as an unexamined string.

**The family year carries no information.** No Worldographer product was ever
labelled "2017" — that is a project coinage, derived from a release year, and it
appears nowhere in a classic file. `"2025"` *is* on disk, but it is a marketing
label with a shelf life: the next product may be labelled differently while the
data format is unchanged. Neither value is a fact about the format, so neither
belongs in a version identity.

**These values are not semver.** `"2.06"` parsed into a semver and rendered back
is `"2.6"`; `"1.06"` becomes `"1.6"`. The dotted components are zero-padded
ordinals, not semantic version fields, and `semver.Version` silently loses them.
Output is correct today only because both encoders write the verbatim strings
(`h2025v1/map.go`, `h2017v1/encode.go`) and never render identity back to disk.
That is an accident worth making a rule.

## Decision

1. **`Dotted` replaces `semver` for on-disk versions.** It keeps the verbatim
   string as the source of truth and parsed components for comparison only:

   ```go
   // Dotted is an on-disk dotted version. It is NOT semver: "2.06" != "2.6".
   // Raw is authoritative for output; the components exist to compare.
   type Dotted struct {
       Raw   string // verbatim, exactly as read or to be written
       Major int
       Minor int
   }
   ```

   Anything written to disk comes from `Raw`. Never re-render a `Dotted` from
   its components.

2. **`Version_t` models the two axes and nothing else.**

   ```go
   type Version_t struct {
       App    Dotted  // map/@version -- the build that wrote the file
       Schema *Dotted // map/@schema  -- nil when the file states none
   }
   ```

   No family year. `Schema == nil` is meaningful: it identifies the one implicit
   legacy schema, evidenced by classic `1.73`/`1.74`/`1.77` sharing an identical
   element vocabulary.

3. **A registry is the single source of truth for supported releases.** Each
   entry binds the full on-disk identity to a codec pair:

   ```go
   type Release_t struct {
       Release string  // map/@release verbatim ("2025"); "" for classic
       App     Dotted  // map/@version
       Schema  *Dotted // map/@schema; nil for classic
       Decode  ...     // input codec
       Encode  ...     // output codec
   }
   ```

   `Release` is carried because writing a file requires it — not because it means
   anything. It is marketing data preserved for fidelity.

4. **Schema selects the codec; the application version is data.** These are
   different questions and conflating them is what produced the "weird vibe":
   - *Which code path parses/emits this?* → the schema. It is the format's
     identity, it is on disk, and it does not change when the product is
     relabelled.
   - *What goes in `@version`?* → the caller's choice, constrained by the
     registry.

   When two application versions share one schema (e.g. a hypothetical `2.06`
   and `2.07` both on schema `1.06`), they use the same codec and differ only in
   the string written to `@version`. The encoder must be **told** which, because
   the schema cannot tell it.

5. **Callers target a release, not a tuple.** `WithTargetVersion` takes an
   application version; the registry resolves `Release` and `Schema` from it.
   Invalid combinations are therefore *unrepresentable* rather than merely
   rejected — a caller cannot ask for `@version="1.77"` with a modern schema
   because no such entry exists. An unregistered `(App, Schema)` pair reaching
   the encoder by any other path is an error, never a best-effort write.

   This is the licensing requirement: a user licensed for `2.06` targets `2.06`
   and cannot be handed a `2.07` file.

6. **`Map_t` stays the superset.** It models the union of features across
   supported releases. Decoding never drops (the model is a superset of any
   input); **encoding to an older target may**. Those are different operations
   and only the second is lossy.

7. **Downgrade loss must be reported, not silent.** Encoding a `Map_t` to a
   target that cannot express some of its content is a data-losing operation and
   the caller must be able to see exactly what was lost. The mechanism follows
   the project's existing convention — diagnostics, not log statements — by
   extending `EncoderDiagnostics` with a dropped-feature inventory. See *Open
   questions* on whether opt-in visibility is sufficient.

## Consequences

- Identity stops lying. Every field is a fact the file states, and `@version` —
  currently discarded for W2025 — is modeled for the first time.
- Dispatch stops depending on a marketing label. A product relabelled from
  "2025" changes `@release` and nothing else; the schema, and therefore the
  codec, is unaffected.
- Adding a release becomes a registry entry rather than a `switch` arm.
- **ADR 0002 is superseded.** Its family-dispatch decision and its
  `{familyYear, revision}` representation are replaced wholesale. Its
  **verbatim-output guarantee survives unchanged** and is reinforced by
  Decision 1: we never synthesize or re-render what was on input.
- The `2017` family key disappears from the model. It remains in git history and
  in ADR 0002 as the record of a superseded decision.

## Open questions

These are for #28 to settle; this ADR does not pre-commit them.

- **Is opt-in loss reporting enough?** Diagnostics are opt-in, so a caller who
  does not ask gets silent data loss on a downgrade. The alternative is to make
  a lossy encode an error unless the caller explicitly accepts it
  (`WithAllowLossy` or similar), which trades convenience for safety. Silent
  loss seems the worse failure.
- **What is actually lost, per target?** The inventory must be built from
  evidence, not from memory. `roundtrip_2017_test.go` already audits what the
  classic codec drops on a round trip and asserts it against
  `h2017v1/COVERAGE.md`; downgrade loss is the same problem and should reuse
  that harness. One entry is already established — see *Terrain layers* below.

### Terrain layers: the first evidenced downgrade loss

Both formats support multiple `<maplayer>` elements, so "classic has no layers"
is **not** the distinction. The real one is what a layer name can be attached to:

- **Classic** binds `mapLayer` to features, labels and shapes (`h2017v1/schema.go`),
  but tiles carry no layer. All terrain sits on a single hard-coded layer.
- **W2025** binds terrain to a named layer per hex:
  `<extraTerrain><mapLayer name="…"><terrainAndLocation location="x,y"/></mapLayer></extraTerrain>`.

Encoding a W2025 map to a classic target therefore cannot express per-hex layer
assignment; terrain collapses onto the one hard-coded layer. That is a genuine
downgrade loss and the caller must be told.

It also qualifies Decision 6. `Map_t` is the superset *by intent*, but
`terrainAndLocation` is modeled only as an opaque stub under `<extraTerrain>`
(#11): the bytes round-trip 2025→2025 unharmed, yet nothing in `Map_t`
understands them. Until a feature is modeled rather than stubbed, the encoder
cannot enumerate what a downgrade would cost — it can only report that an
unmodeled stub could not be carried. **Stub coverage is a precondition for
honest loss reporting**, which makes the `COVERAGE.md` matrices load-bearing for
this ADR rather than merely informational.
- **Registry keying.** Targeting by application version assumes it uniquely
  identifies a release. If that ever fails, callers target a release identifier
  instead.
