# Codec coverage — index

Each WXX codec keeps a per-element read/write coverage matrix: a living
checklist that makes **stub-drift** visible (a stub encoder hiding behind a
passing round-trip test is what motivated this practice in issue #7).

The matrices are **load-bearing, not informational**. ADR 0004 makes honest stub
coverage a precondition for honest downgrade-loss reporting: the encoder refuses
to downgrade content it holds only as an unmodeled stub, because it cannot say
what dropping it would cost (`xmlio/downgrade.go`, #32). A matrix that overstates
coverage therefore weakens a runtime guarantee — it is not just documentation.

- **Worldographer 2025** — [`xmlio/internal/v1_06/COVERAGE.md`](../xmlio/internal/v1_06/COVERAGE.md)

There is one codec, so one matrix. The classic (H2017) codec and its matrix were
removed by #103 ([ADR 0005](adr/0005-remove-classic-format.md)); wxx no longer
reads or writes classic files.

The matrix uses one status vocabulary: **implemented** / **stub** /
**no-op(intentional)** / **lossy** / **unimplemented(dropped)**, mapping to the
ottomap `wog/FEATURES.md` `✅ / ⚠️ / ❌` legend as the matrix documents.

## Read the matrices, not this page

This index deliberately does **not** restate what each codec covers. It used to,
and every per-element claim it made had gone stale: it described the public
decoder as unable to route classic files (it routed them then — `xmlio/decoder.go`;
since #103 it refuses them), and reported six un-modeled W2025 fields that #11 had
already modeled. The per-element truth lives in the matrices and moves with the
code; duplicating it here only guarantees a second copy that drifts.

Two things are worth knowing before you open them:

- **The RelaxNG cross-check is partial.** The formal schema in `schema/` is
  **classic `version="1.73"` scope only** — an upstream copy that predates W2025,
  kept as reference for a format wxx no longer reads (`schema/README.md`). The
  W2025 matrix can only cross-check the elements W2025 *shares* with classic; the
  schema says nothing about W2025 additions such as `<extraTerrain>` or
  `<blurTerrainBG>`. Documenting the W2025 schema is #2.
- **A gap's failure mode matters as much as its existence.** The matrices
  distinguish an encoder that drops content silently from one that refuses loudly
  and from one that writes a plausible-looking constant.
