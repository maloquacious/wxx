# Codec coverage — index

The WXX codecs each keep a per-element read/write coverage matrix, a living
checklist that makes **stub-drift** visible (a stub encoder hiding behind a
passing round-trip test is what motivated this practice in issue #7). The two
matrices share one status vocabulary and format, and each is cross-checked
against the formal RelaxNG schema in `schema/` (v1.73/classic scope):

- **Classic H2017** — [`xmlio/h2017v1/COVERAGE.md`](../xmlio/h2017v1/COVERAGE.md).
  Decode models 100% of the classic format; the encoder still has several silent
  stubs (ROWS write, `<mapkey>`, `<shape>`, `<note>`, `<informations>`,
  `<labelstyle>`), and the public decoder does not yet route classic files.
- **Worldographer 2025** — [`xmlio/h2025v1/COVERAGE.md`](../xmlio/h2025v1/COVERAGE.md).
  `Map_t`-level round-trip is proven by the tests in
  `xmlio/roundtrip_2025_test.go`; six W2025-only fields remain un-modeled and are
  listed there.

Statuses (both files): **implemented** / **stub** / **no-op(intentional)** /
**lossy** / **unimplemented(dropped)**, mapping to the ottomap `wog/FEATURES.md`
`✅ / ⚠️ / ❌` legend as each matrix documents.
