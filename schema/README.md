# WXX RelaxNG Schema (reference only)

This directory holds a formal [RelaxNG](https://relaxng.org/) schema describing the
XML structure of a Worldographer WXX map file. It is **reference and validation
material** — it is not compiled, imported, or enforced anywhere in the build.

- `utf-8-xml.rnc` — RelaxNG in the compact (`.rnc`) syntax. This is the readable one.
- `utf-8-xml.rng` — the same schema in the XML (`.rng`) syntax, consumable by
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

## Version scope — this is v1.73 (classic / H2017) only

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
xmllint --relaxng schema/utf-8-xml.rng path/to/decoded.xml --noout
```
