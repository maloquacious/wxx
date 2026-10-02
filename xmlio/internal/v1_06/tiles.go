// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/hexg"
	"github.com/maloquacious/wxx"
)

// decodeTiles parses the <tiles>/<tilerow> data into the domain map. It also
// decodes <mapkey> inside the tilerow loop (via decodeMapKey), preserving the
// original decoder's ordering in which the map key was materialized per tilerow
// after the tiles were parsed.
func decodeTiles(src Tiles_t, mapKeySrc MapKey_t, w *wxx.Map_t) error {
	var err error
	w.Tiles = &wxx.Tiles_t{
		ViewLevel: src.ViewLevel,
		TilesWide: src.TilesWide,
		TilesHigh: src.TilesHigh,
	}

	// Set RowsHigh and ColumnsWide based on the orientation. The layout is
	// derived from HexOrientation (issue #52); decodeMap has already refused a
	// map stating neither orientation.
	layout, _ := w.GridOrientation()
	switch layout {
	case hexg.OddQ:
		// Column orientation: TilesWide = columns, TilesHigh = rows
		w.RowsHigh = w.Tiles.TilesHigh
		w.ColumnsWide = w.Tiles.TilesWide
	case hexg.OddR:
		// Row orientation: TilesWide = rows, TilesHigh = columns
		w.RowsHigh = w.Tiles.TilesWide
		w.ColumnsWide = w.Tiles.TilesHigh
	}
	for _, tilerow := range src.TileRows {
		x, y := len(w.Tiles.Tiles), 0
		w.Tiles.Tiles = append(w.Tiles.Tiles, make([]*wxx.Tile_t, w.Tiles.TilesHigh))
		// The tilerow is walked with SplitSeq rather than split into a slice of
		// lines, and each line's fields go into a fixed array, so no []string
		// is built per tilerow or per hex (issue #152).
		for line := range strings.SplitSeq(tilerow.InnerText, "\n") {
			if len(line) == 0 { // ignore blank lines
				continue
			}
			// x is the <tilerow> index and y the entry within it. Each
			// <tilerow> is one column in both orientations (issue #85): every
			// sample has tilesWide tilerows of tilesHigh entries, and hexg's
			// offset coordinates take (col, row).
			t := &wxx.Tile_t{Column: x, Row: y}
			if layout == hexg.OddQ {
				t.Coords = hexg.NewOffsetCoord(x, y).QOffsetToCube(false)
			} else if layout == hexg.OddR {
				// Worldographer staggers odd rows right (#80), which is
				// odd-r: even=false (#130).
				t.Coords = hexg.NewOffsetCoord(x, y).ROffsetToCube(false)
			}
			w.Tiles.Tiles[x][y] = t
			y++
			// values are TerrainMapIndex Elevation IsIcy IsGMOnly Animals (Z|(Brick Crops Gems Lumber Metals Rock)) RGBA?
			var fields [12]string
			n := 0
			for field := range strings.SplitSeq(line, "\t") {
				if n < len(fields) {
					fields[n] = field
				}
				n++
			}
			switch n {
			case 6, 7, 11, 12: // allowed
			default:
				return fmt.Errorf("values: expected 6/7/11/12, got %d", n)
			}
			values := fields[:n]
			if t.Terrain, err = strconv.Atoi(values[0]); err != nil {
				return fmt.Errorf("value: terrainType: %w", err)
			}
			if t.Elevation, err = strconv.ParseFloat(values[1], 64); err != nil {
				return fmt.Errorf("value: elevation: %w", err)
			}
			t.IsIcy = values[2] == "1"
			t.IsGMOnly = values[3] == "1"
			if t.Resources.Animal, err = strconv.Atoi(values[4]); err != nil {
				return fmt.Errorf("value: animals: %w", err)
			} else if t.Resources.Animal < 0 {
				return fmt.Errorf("value: animals: %w", fmt.Errorf("invalid value"))
			} else if t.Resources.Animal > 100 {
				return fmt.Errorf("value: animals: %w", fmt.Errorf("invalid value"))
			}
			compressedResources := len(values) == 6 || len(values) == 7
			if compressedResources {
				// a with compressed resources should flag them with a Z
				if values[5] != "Z" {
					return fmt.Errorf("value: sentinel: %w", fmt.Errorf("invalid value"))
				}
			} else {
				if t.Resources.Brick, err = strconv.Atoi(values[5]); err != nil {
					// split afresh: passing values to fmt would move fields
					// to the heap for every tile, not just this one
					return fmt.Errorf("value: brick: %q: %w", strings.Split(line, "\t"), err)
				} else if t.Resources.Brick < 0 {
					return fmt.Errorf("value: brick: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Brick > 100 {
					return fmt.Errorf("value: brick: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Crops, err = strconv.Atoi(values[6]); err != nil {
					return fmt.Errorf("value: crops: %w", err)
				} else if t.Resources.Crops < 0 {
					return fmt.Errorf("value: crops: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Crops > 100 {
					return fmt.Errorf("value: crops: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Gems, err = strconv.Atoi(values[7]); err != nil {
					return fmt.Errorf("value: gems: %w", err)
				} else if t.Resources.Gems < 0 {
					return fmt.Errorf("value: gems: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Gems > 100 {
					return fmt.Errorf("value: gems: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Lumber, err = strconv.Atoi(values[8]); err != nil {
					return fmt.Errorf("value: lumber: %w", err)
				} else if t.Resources.Lumber < 0 {
					return fmt.Errorf("value: lumber: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Lumber > 100 {
					return fmt.Errorf("value: lumber: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Metals, err = strconv.Atoi(values[9]); err != nil {
					return fmt.Errorf("value: metals: %w", err)
				} else if t.Resources.Metals < 0 {
					return fmt.Errorf("value: metals: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Metals > 100 {
					return fmt.Errorf("value: metals: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Rock, err = strconv.Atoi(values[10]); err != nil {
					return fmt.Errorf("value: rock: %w", err)
				} else if t.Resources.Rock < 0 {
					return fmt.Errorf("value: rock: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Rock > 100 {
					return fmt.Errorf("value: rock: %w", fmt.Errorf("invalid value"))
				}
			}
			if len(values) == 7 || len(values) == 12 {
				// split rgba. The column is present only when the tile has a
				// custom background, so an opaque black here is a colour; nil
				// would drop the column on encode (issue #99).
				if t.CustomBackgroundColor, err = decodeZeroableRgba(values[len(values)-1]); err != nil {
					return fmt.Errorf("value: rgba: %w", err)
				}
			}
		}

		if err := decodeMapKey(mapKeySrc, w); err != nil {
			return err
		}
	}
	return nil
}

func encodeTiles(tiles *wxx.Tiles_t, hexOrientation string, terrainRemap map[int]int, wb *bytes.Buffer) error {
	// to: width is the number of columns, height is the number of rows. does that depend on the orientation?
	wb.WriteString(fmt.Sprintf("<tiles"))
	wb.WriteString(fmt.Sprintf(" viewLevel=%s", xmlAttr(tiles.ViewLevel)))
	wb.WriteString(fmt.Sprintf(" tilesWide=%s", xmlAttr(ints(tiles.TilesWide))))
	wb.WriteString(fmt.Sprintf(" tilesHigh=%s", xmlAttr(ints(tiles.TilesHigh))))
	wb.WriteString(fmt.Sprintf(">\n"))

	// generate the tile-row elements:
	// * each tile-row will have a tile.tilesHigh lines of tab delimited data
	// * each line of data has the following values: Terrain type, elevation, is it icy, is it GM only, and its resources
	// * terrainType is an index into the terrainmap element
	// * resources are Animals, Brick, Crops, Gems, Lumber, Metals, and Rock, in that order, but are "compressed"
	//
	// The physical <tilerow> emission is IDENTICAL for COLUMNS and ROWS: the file
	// is always tilesWide <tilerow> elements, each holding tilesHigh tab-delimited
	// lines, and decodeTiles stores tiles in file-physical Tiles[x][y] order (x in
	// 0..tilesWide, y in 0..tilesHigh) for BOTH orientations. Orientation only
	// affects (i) the OddQ vs OddR coordinate interpretation and (ii) the
	// RowsHigh/ColumnsWide labels — neither of which changes the bytes written
	// here. (Cross-check: ROWS == pointy-top hexes per tcfna's vertex-geometry
	// notes; that is a client rendering concern and does not alter this data grid.)
	if hexOrientation == "COLUMNS" || hexOrientation == "ROWS" {
		for x := 0; x < tiles.TilesWide; x++ {
			wb.WriteString("<tilerow>\n")
			for y := 0; y < tiles.TilesHigh; y++ {
				tile := tiles.Tiles[x][y]
				if err := encodeTile(tile, terrainRemap, wb); err != nil {
					return err
				}
			}
			wb.WriteString(fmt.Sprintf("</tilerow>\n"))
		}
	} else {
		// An orientation this codec has never seen. v1_06 writes both of the
		// two that exist, so reaching here means the map states something that
		// is not an orientation at all, and Map_t.Validate rejects that before
		// an encode begins (issue #20). Kept as a refusal for a test unit
		// calling the codec directly, and worded as one: it used to be
		// `assert(orientation != ...)`, which named a condition the codec
		// expected rather than the problem the caller has.
		return errors.Join(wxx.ErrInvalidHexOrientation,
			fmt.Errorf("map/@hexOrientation %q: want %q or %q", hexOrientation, "COLUMNS", "ROWS"))
	}
	wb.WriteString(fmt.Sprintf("</tiles>\n"))
	return nil
}

// some documentation is only in this discord chat - https://discord.com/channels/535205750532997160/877285895991095369/1187771984768151653
// summarizing that:
// * tilerow is tab-delimited data that looks like terrainMapSlot elevation isIcy isGMOnly animals 0 0 0 0 0 0
// * the web page has isIcy as a float, but it seems to be a boolean
// * resource.animals is int with range 0...100
// * field after resource.animal is "Z" if remaining resources are all 0
// * otherwise we have brick, crops, gems, lumber, metals, rock
// * customBackgroundColor is an RGBA that is optional
func encodeTile(tile *wxx.Tile_t, terrainRemap map[int]int, wb *bytes.Buffer) error {
	terrain, err := tileTerrain(terrainRemap, tile)
	if err != nil {
		return err
	}
	// Each field is appended straight into the buffer's spare capacity rather
	// than formatted with fmt.Sprintf, which allocated a string per field: 21-27
	// million allocations on a 1920 x 1080 encode (issue #143). The bytes are the
	// same ones %d wrote.
	b := wb.AvailableBuffer()
	b = strconv.AppendInt(b, int64(terrain), 10)
	b = append(b, '\t')
	b = strconv.AppendInt(b, int64(floatd(tile.Elevation)), 10)
	b = append(b, '\t')
	b = strconv.AppendInt(b, int64(boold(tile.IsIcy)), 10)
	b = append(b, '\t')
	b = strconv.AppendInt(b, int64(boold(tile.IsGMOnly)), 10)
	b = appendTileResources(b, tile.Resources)
	if tile.CustomBackgroundColor != nil {
		b = append(b, '\t')
		b = append(b, rgbas(tile.CustomBackgroundColor)...)
	}
	b = append(b, '\n')
	wb.Write(b)
	return nil
}

// appendTileResources appends a tile's resource columns to b.
//
// Every resource is in 0..100: Map_t.Validate refuses anything else before the
// encoder runs (issue #122), because the app does not keep a value above 100.
func appendTileResources(b []byte, resources wxx.Resources_t) []byte {
	b = append(b, '\t')
	b = strconv.AppendInt(b, int64(resources.Animal), 10)
	// compress if there are no resources other than Animal
	if resources.Brick == 0 && resources.Crops == 0 && resources.Gems == 0 && resources.Lumber == 0 && resources.Metals == 0 && resources.Rock == 0 {
		return append(b, "\tZ"...)
	}
	for _, r := range []int{resources.Brick, resources.Crops, resources.Gems, resources.Lumber, resources.Metals, resources.Rock} {
		b = append(b, '\t')
		b = strconv.AppendInt(b, int64(r), 10)
	}
	return b
}
