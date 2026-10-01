# WXX reference grammars

This directory holds the **wxx reference grammars** for Worldographer map files:
one RELAX NG compact (`.rnc`) grammar per Worldographer schema version
(`map/@schema`).

| grammar | describes | written by | evidence | status |
|---|---|---|---|---|
| [`1.06.rnc`](1.06.rnc) | `map/@schema="1.06"` | Worldographer 2025 2.06, 2.07, 2.08 | 2.08 saves in `testdata/` | **current** |

A grammar here is our best description of the files, built from a small sample.
It is **not the official schema**. The format belongs to Worldographer's author;
where a grammar and the app disagree, the app is right.

## For users: what is in a WXX file

A `.wxx` file is gzip-compressed, UTF-16 big-endian XML 1.1. To read one:

```sh
gunzip -c map.wxx | iconv -f UTF-16BE -t UTF-8 | less
```

The output starts with a byte-order mark and the declaration
`<?xml version='1.1' encoding='utf-16'?>`, then a single `<map>` element. The
grammar describes that `<map>` element and everything inside it. The `<map>`
attributes `release`, `version` and `schema` name the format: a Worldographer
2025 file states `release="2025"`, its application version (e.g. `"2.08"`), and
its schema version (`"1.06"`). The schema version picks the grammar.

## For developers

### How the grammar relates to the code

Three documents describe the same files from different sides. Each owns its
part; do not copy one into another, because copies drift.

- **The grammar** (`1.06.rnc`) describes what the files contain: elements,
  nesting, attributes, datatypes, where text appears.
- **`Map_t`** ([`map.go`](../map.go)) is how wxx models a map in memory. It is
  not a mirror of the file: it may hold a structure differently, or not at all.
- **The `v1_06` codec's coverage matrix**
  ([`xmlio/internal/v1_06/COVERAGE.md`](../xmlio/internal/v1_06/COVERAGE.md))
  records what the codec does with each element: implemented, stub, no-op or
  lossy.

So "is this in the file?" is a grammar question, and "does wxx keep it?" is a
coverage question.

### Running the test

```sh
go test ./schema/
go test -v -run TestGrammarCoversFixtures ./schema/   # per-fixture counts
```

`grammar_test.go` walks every `testdata/2025-2.08-*.wxx` and fails on any
element, attribute or non-whitespace text that the grammar does not allow at that
position. It checks one direction only, fixture to grammar. It does not check
values, child order or cardinality, and a rule that no fixture exercises passes.
That is why every rule carries a label (below).

The grammar uses a small subset of RELAX NG compact syntax so that a plain-Go
test can parse it without a dependency. The subset is listed in the header of
`1.06.rnc`, under *THE SUBSET*; the test rejects anything outside it.

### When a new file breaks the test

A failure reads like:

```
<fixture>: grammar does not allow attribute foo in pattern Map/Tiles (first at /map/tiles[1]; 1 occurrence(s))
```

The test reads only `testdata/2025-2.08-*.wxx`, so a new file is checked once
it is tracked there under that name.

1. Confirm the file is a genuine save from the app, and check the bytes
   yourself with the `gunzip | iconv` command above.
2. Add the element or attribute to the named pattern, in the position the file
   uses, with `?` (or `*` for an element that repeats).
3. Label it: `# observed: <fixture>` if a tracked fixture holds it, or
   `# observed: app check, #NN` if the maintainer tested it in the app. Use
   the fixture short names from the grammar's *EVIDENCE* list.
4. Pick the datatype by the header's *DATATYPES* rules. Do not add an
   enumeration from the values you see; an enumeration needs the full set from
   the app or the code.
5. If the file contradicts an existing `# inferred` rule, correct the rule and
   its label rather than working around the file.

A fixture the test reads must live in `testdata/` and be tracked.

## For agents

- **Cite the grammar** when reasoning about the format: name the pattern and
  its label (e.g. "`Feature`, observed: populated").
- **`# inferred` and `TBD` are hypotheses**, not facts. Do not repeat them as
  facts.
- **Never state a format fact that the grammar and a fixture don't support.**
  Before you write one in a comment, doc, issue or commit message, check the
  fixture bytes with the `gunzip | iconv` command above.
- **When a file contradicts an inferred rule, suspect the grammar first.** The
  grammar is built from six saves; the app wrote the file.
- **Do not infer the format from Go structs.** `Map_t` and the codec's schema
  types show what wxx models, not what the file holds.

## Conventions

- **One file per schema version, named by it**: `1.06.rnc` for
  `map/@schema="1.06"`. Several application versions can write one schema; the
  file's header names them and names the application version whose saves are
  the evidence.
- **Every schema is kept.** A new schema version gets a new file next to the old
  ones. Older grammars may stop being updated once Worldographer moves on; the
  table at the top says which grammar is current.
- **Every rule is labelled** with where it came from: `# observed: <fixtures>`
  (seen in those fixtures, or in an app check), `# inferred: <why>` (a guess),
  or `TBD` (not known). The full rules, including what `?`, `*` and child order
  claim, are in the grammar's header under *LABELS*.
- **Syntax, not semantics.** RNC carries structure. Meaning goes in short RNC
  comments, or in this README when it needs more room.

## Format notes

Facts that need more room than an RNC comment. Each names its source.

**Each `<tilerow>` is one column, in COLUMNS and ROWS maps alike.** A map has
`tilesWide` tilerows, each holding `tilesHigh` tile records, one per line, top
row first (grammar: `TileRow`, observed: all). The `rows` fixture is a ROWS map
and still has 13 tilerows of 11 records. A 2.08 ROWS map 1 wide by 11 high
holds one tilerow of 11 records (#85). Orientation changes only how hexes are
drawn: COLUMNS staggers odd columns down, ROWS staggers odd rows right (#80,
recorded in a comment on #72). The tab-separated fields of a tile record,
each with its source, are listed in the grammar (`TileRow`, #117).

**Colors have two spellings** (grammar header, *COLORS*):

- `r,g,b,a`: four decimals, each 0 to 1. An out-of-range component stops the
  file opening (app check, #83: `dsColor="255,0,0,1"`), and so does a shape
  with `dsColor=""` and `insColor=""`. Many color attributes also hold the
  literal `null`; that it means "no color" is inferred.
- `0xRRGGBBAA`: `gridandnumbering/@color0..4` and `@numberColor` (observed:
  all). `color0=""` opens and is saved back as `null` (app check, #83, on
  `color0` only; `color1..4` are inferred to behave the same).

Two per-attribute details from the same checks: `shape/@strokeColor="null"`
opens, and the app then saves the shape with no `strokeColor` at all (grammar:
`Shape`, app check, #83). `dsColor`/`insColor` as `null` appears in no fixture
here and is TBD (grammar: `Shape`).

**A feature's ring color has two spellings.** Every feature carries exactly one
of them: `ringcolor="null"` when no ring color is set, `ringColor="r,g,b,a"`
when one is (grammar: `Feature`, observed: layers, notes-shapes, populated;
#100). A reader that looks for only one spelling loses the color.

**Some names change spelling between contexts:**

- `dscolor` on `<shapestyle>`, `dsColor` on `<shape>` (grammar: `ShapeStyle`,
  `Shape`).
- `<maplayer>` as a child of `<map>`, `<mapLayer>` as a child of
  `<extraTerrain>` (grammar: `MapLayer`, `ExtraTerrainLayer`).

**Religion `<information>` UUIDs are not stable across a save.** Re-saving a
2.06 map in 2.08 regenerated all 84 Religion UUIDs, while Information, Nation
and Culture UUIDs were kept (comment on #72, observed 2026-09-29). References
still resolve: `information/@culture` and `@cultures` hold a Culture UUID
(grammar: `Information`), and no reference to a Religion UUID has been seen.
Anything that matches lore entries across saves, such as a diff tool or a
script, must not use a Religion UUID as a key.

## Changes since the previous schema

None yet. `1.06.rnc` is the only grammar. When Worldographer writes a new schema
version, this section records what changed between them, since that difference
is what downgrade-loss reporting needs.

## History

Issue #72 removed the classic (Worldographer 1.x) grammar,
`schema/utf-8-xml.rnc`/`.rng`, when it introduced `1.06.rnc`: wxx no longer
reads the format it describes (#103). [ADR 0005](../docs/adr/0005-remove-classic-format.md)
records the removal in its 2026-09-30 amendment. The files remain in git
history:

```sh
git show 8e0cd55:schema/utf-8-xml.rnc
```
