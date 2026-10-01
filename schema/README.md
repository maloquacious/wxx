# WXX RelaxNG Schema (reference only)

This directory holds a formal [RelaxNG](https://relaxng.org/) schema describing the
XML structure of a Worldographer WXX map file. It is **reference and validation
material** — it is not compiled, imported, or enforced anywhere in the build.

**It describes the classic (H2017, Worldographer 1.x) format only, which wxx no
longer reads or writes.** Issue #103 removed classic support
([ADR 0005](../docs/adr/0005-remove-classic-format.md)): the decoder refuses a
classic file with `wxx.ErrClassicMap`. The schema is kept as reference material,
for the elements W2025 shares with classic and because earlier coverage work cites
it as evidence.

- `v1.73.rnc` — RelaxNG in the compact (`.rnc`) syntax. This is the readable one.
- `v1.73.rng` — the same schema in the XML (`.rng`) syntax, consumable by
  RelaxNG validators such as [Jing](https://relaxng.org/jclark/jing.html) or
  `xmllint --relaxng`.

The two files are equivalent expressions of one schema; edit neither — they are an
upstream copy (see Provenance).

## What RelaxNG is

RelaxNG is a schema language for XML: it declares which elements and attributes are
allowed, how they nest, cardinality (`?` optional, `+` one-or-more), and each
attribute's datatype. Here that gives us an authoritative, machine-readable checklist
of every element and attribute the WXX format uses, plus their types — for example
that grid/number colors are `xsd:NMTOKEN`, most enum-like values are `xsd:NCName`,
and coordinates/offsets are `xsd:decimal`. It is the reference the codec-coverage
work (issue #8, task B2) checks completeness against.

## Provenance

Ported verbatim from the **tnwxx** project (`github.com/playbymail/tnwxx`), a TribeNet
WXX mapping tool. The files were copied byte-for-byte (unmodified) from that repo's
`testdata/` directory:

- Source: `~/Jetbrains/worldographer/tnwxx/testdata/utf-8-xml.rnc` (dated 2023-12-15)
- Source: `~/Jetbrains/worldographer/tnwxx/testdata/utf-8-xml.rng` (dated 2024-01-02)
- Copied: 2026-07-13

This repo renamed them `v1.73.rnc` and `v1.73.rng` in #72; the content is unchanged.

The schema was derived from a real Worldographer export in that repo
(`testdata/utf-8-xml.xml`), whose root element is:

```xml
<map type="WORLD" version="1.73" lastViewLevel="WORLD" ...>
```

### License / attribution

tnwxx is MIT licensed, Copyright (c) 2024 Michael D Henderson. These schema files are
redistributed here under those terms. The MIT license permits copying and
redistribution provided the copyright and permission notice are preserved; the
original notice lives in the tnwxx repository's `LICENSE`.

## Version scope — this is v1.73 (classic / H2017) only, a format wxx no longer reads

The `<map version="1.73">` on the source export pins this schema to the classic
**H2017**-era format. The schema itself types the attribute loosely as
`attribute version { xsd:decimal }`, so it does not hardcode `1.73`; the version claim
comes from the source file it was reverse-engineered from.

**Do not treat this as a complete schema for Worldographer 2025 (W2025).** It predates
the W2025 additions. Verified deltas that are *absent* from this schema and therefore
NOT covered:

- `map/@release` — the `release="2025"` attribute the decoder uses to dispatch W2025.
  Classic files (this schema) carry no `release` attribute at all.
- `maplayer/@opacity` — here `<maplayer>` has only `isVisible` and `name`.
  (Note: the `opacity` attributes that *do* appear in this schema are unrelated —
  `mapkey/@backgroundopacity`, `shape/@opacity`, `shapestyle/@opacity`.)
- `blurTerrainBG` — a W2025 tile-background control; not present.
- `extraTerrain` — a W2025 addition; not present.
- `shape/@extraLineDistance`, `@extraLineLength`, `@extraLineWidth`,
  `@extraLineSeparation` — observed on every `<shape>` in
  `testdata/2025-{2.06,2.07,2.08}-13x11-941577-notes-shapes.wxx` (issue #94);
  this schema's `<shape>` has none of them.

Two differences in `<shape>` and `<p>` that are not additions, also observed in
those three fixtures (issue #94):

- `shape/@fillRule` is required here, but W2025 tile-border polygons
  (`isMatchTileBorders="true"`) state none; the paths state
  `fillRule="NON_ZERO"`.
- `p/@x` and `p/@y` are `xsd:decimal` here, which admits both spellings the
  fixtures use: integers on tile-border polygon points (`x="2700"`) and decimals
  on path points (`x="1950.0"`). `p/@type` is optional here, and in the
  notes-shapes fixtures appears only as `type="m"` on a path's first point.
- Curve control points, observed in 2.08 in
  `testdata/2025-2.08-13x11-941577-populated.wxx`: the point after `type="m"`
  on a curved path (`isCurve="true"`) is `type="c"` and states `@cx1 @cy1 @cx2
  @cy2`, all decimal-spelled (`cx1="2390.6453009961024"`). This schema's `<p>`
  has none of them.

Anyone using this as a checklist for W2025 codec coverage must layer those known
additions on top. For the W2025 shape of the format, see the `wog` V2025 structs noted
in `docs/WORLDOGRAPHER_INVENTORY.md` and the W2025 decoder work in `xmlio/internal/v1_06/`.

## Not build-time enforced

Nothing in this Go module reads or validates against these files. Validating a WXX
XML payload against the schema would require a third-party RelaxNG library, and this
project deliberately keeps its dependency set minimal (only `semver` and
`golang.org/x/text`). No validation harness is provided. If you want to validate a
sample by hand, decode the WXX container to UTF-8 XML and run an external validator,
e.g.:

```sh
xmllint --relaxng schema/v1.73.rng path/to/decoded.xml --noout
```
