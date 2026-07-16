# Codec coverage — index

The WXX codecs each keep a per-element read/write coverage matrix: a living
checklist that makes **stub-drift** visible (a stub encoder hiding behind a
passing round-trip test is what motivated this practice in issue #7).

The matrices are **load-bearing, not informational**. ADR 0004 makes honest stub
coverage a precondition for honest downgrade-loss reporting: the encoder refuses
to downgrade content it holds only as an unmodeled stub, because it cannot say
what dropping it would cost (`xmlio/downgrade.go`, #32). A matrix that overstates
coverage therefore weakens a runtime guarantee — it is not just documentation.

- **Classic H2017** — [`xmlio/h2017v1/COVERAGE.md`](../xmlio/h2017v1/COVERAGE.md)
- **Worldographer 2025** — [`xmlio/h2025v1/COVERAGE.md`](../xmlio/h2025v1/COVERAGE.md)

Both share one status vocabulary: **implemented** / **stub** /
**no-op(intentional)** / **lossy** / **unimplemented(dropped)**, mapping to the
ottomap `wog/FEATURES.md` `✅ / ⚠️ / ❌` legend as each matrix documents.

## Read the matrices, not this page

This index deliberately does **not** restate what each codec covers. It used to,
and every per-element claim it made had gone stale: it described the public
decoder as unable to route classic files (it routes them —
`xmlio/decoder.go`), and reported six un-modeled W2025 fields that #11 had
already modeled. The per-element truth lives in the matrices and moves with the
code; duplicating it here only guarantees a second copy that drifts.

Two things are worth knowing before you open them:

- **The RelaxNG cross-check is asymmetric.** The formal schema in `schema/` is
  **classic `version="1.73"` scope only** — an upstream copy that predates W2025
  (`schema/README.md`). So the classic matrix is cross-checked against a complete
  grammar, while the W2025 matrix can only cross-check the elements W2025 *shares*
  with classic; the schema says nothing about W2025 additions such as
  `<extraTerrain>` or `<blurTerrainBG>`. Documenting the W2025 schema is #2.
- **A gap's failure mode matters as much as its existence.** The matrices
  distinguish an encoder that drops content silently from one that refuses loudly
  and from one that writes a plausible-looking constant. Classic ROWS encode, for
  instance, is a **hard error by decision**, not a silent stub — the loudest gap in
  the codebase, and the easiest to mis-summarize as the quietest.
