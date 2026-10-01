# Test Fixtures

Every file the test harness reads lives flat in this directory and is tracked,
so `go test ./...` runs from a clean clone. Scratch output, debug dumps and
terrain textures live in `scratch/`, which is git-ignored — never put a fixture a
test needs there.

The classic (Worldographer 1.x) fixtures were removed by #103; only
`2017-1.77-1.0-columns-blank.wxx` remains, to pin that a classic map is refused.

## File naming

    YEAR-VERSION-WIDTHxHEIGHT-SEED-TERRAIN.wxx

e.g. `2025-2.06-13x11-941577-blank.wxx`.

`VERSION` is the `version` attribute on the file's own `map` element — the
Worldographer build that wrote it. Read it out of the file rather than typing
what the application reported, so the name always matches the contents.

Recording width, height, seed and terrain in the name makes a sample
reproducible from its filename alone.

## Saving a fixture

Every recipe below ends with a save. Worldographer changes a map's contents
when you touch it again, so follow these rules for every fixture:

- Do not scroll or resize the map before you save it.
- Save only once. A second save, even Save As from the same window, changes
  the file: `mapkey/@viewlevel` goes from `"null"` to `"WORLD"`, every religion
  in `<informations>` gets a new `uuid`, and the `<extraTerrain>` layers are
  reordered.
- Never open the map file again. Doing so may change the contents.

Worldographer writes a `*-autosave.wxx` alongside the map and deletes it on a
clean exit. Autosaves are transient and are git-ignored; never commit one.

## Inspecting a fixture

To read a sample as UTF-8 XML (the output is a scratch artifact, not
committed):

    go run ./cmd/wxx export testdata/2025-2.06-13x11-941577-blank.wxx --utf-8 scratch/2025-2.06-13x11-941577-blank.utf8

## Blank

A quick way to view the metadata on new versions.

File > New World/Kingdom map

Hex Orientation: Columns Line Up
Map Projection: Flat
  Hexes Wide: 13
  Hexes High: 11

Initial View Level: WORLD

[x] Use suggested pixel sizes

Random Seed: 941577

All one terrain: Blank

Generate Map

Save as testdata/YEAR-VERSION-WIDTHxHEIGHT-SEED-blank.wxx

## Layers

File > New World/Kingdom map

Hex Orientation: Columns Line Up
Map Projection: Flat
  Hexes Wide: 13
  Hexes High: 11

Initial View Level: WORLD

[x] Use suggested pixel sizes

Random Seed: 941577

All one terrain: Blank

Generate Map

### Add Terrain
Open the Terrain tab and:

1. Select Below All from the dropdown, disable terrain fill, and add Mountain Volcano to (11,9), (12,9), (9,10), (10,10), (11,10) and (12,10).
2. Select Terrain Water from the dropdown, disable terrain fill, and add Water Sea to (11,9) and (12,9).
3. Select Terrain Land from the dropdown, enable terrain fill, and fill the layer with Farmland by clicking in (1,0).
4. Select Above Terrain from the dropdown, disable terrain fill, and add Flat Snowfields to (12,8) and (12,9).

### Add Features
Open the Features tab, select Building Cathedral and:

1. Select the Features layer, and click on (0,0).
2. Select the Below All layer, and click on (11,10).
3. Select the Terrain Water layer, and click on (11,9).
4. Select the Terrain Land layer, and click on (10,9).
5. Select the Above Terrain layer, and click on (12,9).

### Save

Before saving, ensure all layers are visible, GM Only: Show is checked, Grid: Show/Numbers/Shadows are checked.

Save as testdata/YEAR-VERSION-WIDTHxHEIGHT-SEED-layers.wxx

## Rows

File > New World/Kingdom map

Hex Orientation: Rows Line Up
Map Projection: Flat
  Hexes Wide: 13
  Hexes High: 11

Initial View Level: WORLD

[x] Use suggested pixel sizes

Random Seed: 941577

All one terrain: Blank

Generate Map

### Add Terrain
Open the Terrain tab and:

1. Select Terrain Water from the dropdown, disable terrain fill, and add Water Sea to (0,0), (0,1), (1,2), (1,3) and (2,3).

### Save

Before saving, ensure all layers are visible, GM Only: Show is checked, Grid: Show/Numbers/Shadows are checked.

Save as testdata/YEAR-VERSION-WIDTHxHEIGHT-SEED-rows.wxx

## Resources

Painting Farmland makes Worldographer write each tile's resources out in
full (11 fields per tile, where a Blank or Sea tile has 6), so this map
carries tiles with uncompressed resources.

File > New World/Kingdom map

Hex Orientation: Columns Line Up
Map Projection: Flat
  Hexes Wide: 13
  Hexes High: 11

Initial View Level: WORLD

[x] Use suggested pixel sizes

Random Seed: 941577

All one terrain: Blank

Generate Map

### Add Terrain
Open the Terrain tab and:

1. Select Terrain Land from the dropdown, enable terrain fill, and fill the layer with Farmland by clicking in (1,0).
2. Disable terrain fill, select Terrain Water from the dropdown, and add Water Sea to (11,9), (12,9), (9,10), (10,10), (11,10) and (12,10).

Keep this order. Painting Sea on Terrain Water after the Farmland fill trips a
Worldographer UI bug, but the saved data is consistent, and this is the order
that produced the 2.07 and 2.08 fixtures.

### Save

Before saving, ensure all layers are visible, GM Only: Show is checked, Grid: Show/Numbers/Shadows are checked.

Save as testdata/YEAR-VERSION-WIDTHxHEIGHT-SEED-resources.wxx

## Notes and Shapes

Verify the shape and contents of Notes.

File > New World/Kingdom map

Hex Orientation: Columns Line Up
Map Projection: Flat
  Hexes Wide: 13
  Hexes High: 11

Initial View Level: WORLD

[x] Use suggested pixel sizes

Random Seed: 941577

All one terrain: Blank

Generate Map

### Add Terrain, Shapes, Features and Notes

"Row N" means every hex in row N.

1. Desert Cold on Below All

   - Hex (0,0) Terrain: check Icy, check GM Only, Below All layer, Desert Cold
   - Hex (0,1) Terrain: check Icy, uncheck GM Only, Below All layer, Desert Cold
   - Hex (0,2) Terrain: uncheck Icy, check GM Only, Below All layer, Desert Cold
   - Hex (0,3) Terrain: uncheck Icy, uncheck GM Only, Below All layer, Desert Cold
   - Row 10 Terrain: uncheck Icy, uncheck GM Only, Below All layer, Desert Cold

2. Water Sea on Terrain Water

   - Hex (1,0) Terrain: check Icy, check GM Only, Terrain Water layer, Water Sea
   - Hex (1,1) Terrain: check Icy, uncheck GM Only, Terrain Water layer, Water Sea
   - Hex (1,2) Terrain: uncheck Icy, check GM Only, Terrain Water layer, Water Sea
   - Hex (1,3) Terrain: uncheck Icy, uncheck GM Only, Terrain Water layer, Water Sea
   - Hex (1,10) Terrain: uncheck Icy, uncheck GM Only, Terrain Water layer, Water Sea
   - Row 9 Terrain: uncheck Icy, uncheck GM Only, Terrain Water layer, Water Sea

3. Water Shoals on Above Water

   - Hex (2,0) Terrain: check Icy, check GM Only, Above Water layer, Water Shoals
   - Hex (2,1) Terrain: check Icy, uncheck GM Only, Above Water layer, Water Shoals
   - Hex (2,2) Terrain: uncheck Icy, check GM Only, Above Water layer, Water Shoals
   - Hex (2,3) Terrain: uncheck Icy, uncheck GM Only, Above Water layer, Water Shoals
   - Hex (2,9) Terrain: uncheck Icy, uncheck GM Only, Above Water layer, Water Shoals
   - Hex (2,10) Terrain: uncheck Icy, uncheck GM Only, Above Water layer, Water Shoals
   - Row 8 Terrain: uncheck Icy, uncheck GM Only, Above Water layer, Water Shoals

4. Flat Farmland on Terrain Land

   - Hex (3,0) Terrain: check Icy, check GM Only, Terrain Land layer, Flat Farmland
   - Hex (3,1) Terrain: check Icy, uncheck GM Only, Terrain Land layer, Flat Farmland
   - Hex (3,2) Terrain: uncheck Icy, check GM Only, Terrain Land layer, Flat Farmland
   - Hex (3,3) Terrain: uncheck Icy, uncheck GM Only, Terrain Land layer, Flat Farmland
   - Hex (3,8) Terrain: uncheck Icy, uncheck GM Only, Terrain Land layer, Flat Farmland
   - Hex (3,9) Terrain: uncheck Icy, uncheck GM Only, Terrain Land layer, Flat Farmland
   - Hex (3,10) Terrain: uncheck Icy, uncheck GM Only, Terrain Land layer, Flat Farmland
   - Row 7 Terrain: uncheck Icy, uncheck GM Only, Terrain Land layer, Flat Farmland

5. Flat Snowfields on Above Terrain

   - Hex (4,0) Terrain: check Icy, check GM Only, Above Terrain layer, Flat Snowfields
   - Hex (4,1) Terrain: check Icy, uncheck GM Only, Above Terrain layer, Flat Snowfields
   - Hex (4,2) Terrain: uncheck Icy, check GM Only, Above Terrain layer, Flat Snowfields
   - Hex (4,3) Terrain: uncheck Icy, uncheck GM Only, Above Terrain layer, Flat Snowfields
   - Hex (4,7) Terrain: uncheck Icy, uncheck GM Only, Above Terrain layer, Flat Snowfields
   - Hex (4,8) Terrain: uncheck Icy, uncheck GM Only, Above Terrain layer, Flat Snowfields
   - Hex (4,9) Terrain: uncheck Icy, uncheck GM Only, Above Terrain layer, Flat Snowfields
   - Hex (4,10) Terrain: uncheck Icy, uncheck GM Only, Above Terrain layer, Flat Snowfields
   - Row 6 Terrain: uncheck Icy, uncheck GM Only, Above Terrain layer, Flat Snowfields

6. Shapes

   - Hex (12,0) Polygon: Above Terrain layer on Shapes tab, color White, check GM Only, check Add Tile Border, and add a polygon, click De-select
   - Hex (10,0) Polygon: Terrain Land layer on Shapes tab, color Green, uncheck GM Only, check Add Tile Border, and add a polygon, click De-select
   - Hex (8,0) Polygon: Below All layer on Shapes tab, color Magenta, uncheck GM Only, check Add Tile Border, and add a polygon, click De-select
   - Column 8 Line: Terrain Water layer on Shapes tab, color Blue, uncheck GM Only, check Snap Points to Grid, uncheck Add Tile Border, and draw a line from the centre of (8,1) to the centre of (8,4), click De-select

     The endpoints are best effort. The dark UI hides hex borders on a blank
     map, and Snap Points to Grid moves the points; the 2.06 sample's line runs
     from (1950,525) to (1950,1200).

7. Features

   - Hex (0,3) Features: click Building Cathedral on Features tab, select Features layer, uncheck GM Only, set Label to "(0,3)", add feature, click Select
   - Hex (1,3) Features: click Building Cathedral on Features tab, select Terrain Land layer, check GM Only, set Label to "(1,3) GM", add feature, click Select
   - Hex (12,6) Features: click Building Cathedral on Features tab, select Below All layer, check GM Only, set Label to "(12,6) GM", add feature, click Select

8. Notes

   - Hex (1,3) Features: click Select, select the feature, add note "Note on (1,3)", Save, click Select
   - Hex (12,6) Features: click Select, select the feature, add note "Note on (12,6)", Save, click Select

### Save

Before saving, ensure all layers are visible, GM Only: Show is checked, Grid: Show/Numbers/Shadows are checked.

Save as testdata/YEAR-VERSION-WIDTHxHEIGHT-SEED-notes-shapes.wxx

## Populated

Tests 
- a feature with a non-black colour
- a feature with an opaque black colour
- a feature with no label
- a note with a title
- a curved shape variants
- a standalone <labels> entry

  1. Flood fill the Terrain Land layer with Flat Farmland.
  2. Hex (0,0) Features: click Building Cathedral on Features tab, leave Label empty, add feature, click Select
  3. Hex (1,0) Features: click Building Cathedral on Features tab, check Override Color and Add Ring, set both to White, add feature, click Select
  4. Hex (2,1) Features: click Building Cathedral on Features tab, check Override Color and Add Ring, set both to Black, add feature, click Select
  5. Hex (3,1) Features: click Building Pyramid on Features tab, uncheck Override Color and Add Ring, add feature, click Select, select the feature, add a note with "Title (3,1)", color Magenta, body "Body (3,1)" and click Save, click Select
  6. Hex (4,2) Labels: Set Text "Label (4,2)", click New Label, click in the center of the hex, then click De-select
  7. Curve: Shapes tab, click Curve, color Red, then add a curve by clicking in (7,6), (10,6), (10,8), and (8,9), then click De-select

### Save

Before saving, ensure all layers are visible, GM Only: Show is checked, Grid: Show/Numbers/Shadows are checked.

Save as testdata/YEAR-VERSION-WIDTHxHEIGHT-SEED-populated.wxx
