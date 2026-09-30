# Test Fixtures

Every file the test harness reads lives flat in this directory and is tracked,
so `go test ./...` runs from a clean clone. Scratch output, debug dumps and
terrain textures live in `scratch/`, which is git-ignored — never put a fixture a
test needs there.

## File naming

    YEAR-VERSION-WIDTHxHEIGHT-SEED-TERRAIN.wxx

e.g. `2025-2.06-13x11-941577-blank.wxx`.

`VERSION` is the `version` attribute on the file's own `map` element — the
Worldographer build that wrote it. Read it out of the file rather than typing
what the application reported, so the name always matches the contents.

Recording width, height, seed and terrain in the name makes a sample
reproducible from its filename alone.

## New Worldographer Versions

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

Save as testdata/YEAR-VERSION-WIDTHxHEIGHT-SEED-TERRAIN.wxx

Do not scroll or resize the map before you save it!
Never open the map file again. Doing so may change the contents.

Worldographer writes a `*-autosave.wxx` alongside the map and deletes it on a
clean exit. Autosaves are transient and are git-ignored; never commit one.

To inspect a sample as UTF-8 XML (the output is a scratch artifact, not
committed):

$ go run ./cmd/wxx export testdata/2025-2.06-13x11-941577-blank.wxx --utf-8 2025-2.06-13x11-941577-blank.utf8

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

Do not scroll or resize the map before you save it!
Never open the map file again. Doing so may change the contents.
