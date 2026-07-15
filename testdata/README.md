# Test Fixtures

Every file the test harness reads lives flat in this directory and is tracked,
so `go test ./...` runs from a clean clone. Scratch output, debug dumps and
terrain textures live in `scratch/`, which is git-ignored — never put a fixture a
test needs there.

## File naming

    YEAR-VERSION-WIDTHxHEIGHT-SEED-TERRAIN.wxx

e.g. `2025-2.06-13x11-941577-blank.wxx`.

`VERSION` is the value of the `version` attribute on the file's own `map`
element — **not** the version Worldographer reports in its UI. Read it from the
file, never from the application; the earlier `2025-2.05.wxx` sample was named
from the application and declared `version="2.06"` on disk, which cost us a
round of confusion.

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

$ go run cmd/wxx export testdata/2025-2.06-13x11-941577-blank.wxx --utf-8 2025-2.06-13x11-941577-blank.utf8
