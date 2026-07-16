// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
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

	// Set RowsHigh and ColumnsWide based on GridOrientation
	switch w.GridOrientation {
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
		for _, line := range strings.Split(tilerow.InnerText, "\n") {
			if len(line) == 0 { // ignore blank lines
				continue
			}
			t := &wxx.Tile_t{Row: x, Column: y}
			if w.GridOrientation == hexg.OddQ {
				t.Coords = hexg.NewOddQCoord(y, x).ToCube()
			} else if w.GridOrientation == hexg.OddR {
				t.Coords = hexg.NewOddRCoord(y, x).ToCube()
			}
			w.Tiles.Tiles[x][y] = t
			y++
			// values are TerrainMapIndex Elevation IsIcy IsGMOnly Animals (Z|(Brick Crops Gems Lumber Metals Rock)) RGBA?
			values := strings.Split(line, "\t")
			//fmt.Printf("tilerow: %d %d: len(inner) %d lines %d line %d values %d\n", r, i+1, len(element.InnerText), len(lines), len(line), len(values))
			switch len(values) {
			case 6, 7, 11, 12: // allowed
			default:
				return fmt.Errorf("values: expected 6/7/11/12, got %d", len(values))
			}
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
					return fmt.Errorf("value: brick: %q: %w", values, err)
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
				// split rgba
				if t.CustomBackgroundColor, err = decodeRgba(values[len(values)-1]); err != nil {
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

func encodeTiles(tiles *wxx.Tiles_t, hexOrientation string, wb *bytes.Buffer) error {
	// to: width is the number of columns, height is the number of rows. does that depend on the orientation?
	wb.WriteString(fmt.Sprintf("<tiles"))
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", tiles.ViewLevel))
	wb.WriteString(fmt.Sprintf(" tilesWide=%q", ints(tiles.TilesWide)))
	wb.WriteString(fmt.Sprintf(" tilesHigh=%q", ints(tiles.TilesHigh)))
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
				if err := encodeTile(tile, wb); err != nil {
					return err
				}
			}
			wb.WriteString(fmt.Sprintf("</tilerow>\n"))
		}
	} else {
		return fmt.Errorf("assert(orientation != %q)", hexOrientation)
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
func encodeTile(tile *wxx.Tile_t, wb *bytes.Buffer) error {
	// todo: implement this
	wb.WriteString(fmt.Sprintf("%d", tile.Terrain))
	wb.WriteString(fmt.Sprintf("\t%d", floatd(tile.Elevation)))
	wb.WriteString(fmt.Sprintf("\t%d", boold(tile.IsIcy)))
	wb.WriteString(fmt.Sprintf("\t%d", boold(tile.IsGMOnly)))
	if err := encodeTileResources(tile.Resources, wb); err != nil {
		return err
	}
	if tile.CustomBackgroundColor != nil {
		wb.WriteString(fmt.Sprintf("\t%s", rgbas(tile.CustomBackgroundColor)))
	}
	wb.WriteString(fmt.Sprintf("\n"))
	return nil
}

// all resources are supposed to be in the range of 0...100, but we don't enforce
func encodeTileResources(resources wxx.Resources_t, wb *bytes.Buffer) error {
	// compress if there are no resources other than Animal
	if resources.Brick == 0 && resources.Crops == 0 && resources.Gems == 0 && resources.Lumber == 0 && resources.Metals == 0 && resources.Rock == 0 {
		wb.WriteString(fmt.Sprintf("\t%d", resources.Animal))
		wb.WriteString(fmt.Sprintf("\tZ"))
		return nil
	}
	wb.WriteString(fmt.Sprintf("\t%d", resources.Animal))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Brick))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Crops))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Gems))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Lumber))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Metals))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Rock))
	return nil
}
