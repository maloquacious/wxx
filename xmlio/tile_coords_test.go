// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"testing"

	"github.com/maloquacious/hexg"
	"github.com/maloquacious/wxx/xmlio"
)

// TestDecodedTileCoordinates: Tiles[c][r] is the hex in column c, row r, and
// its Column, Row and Coords must say so (issue #85).
//
// Each <tilerow> is one column in both orientations: every sample has
// tilesWide tilerows of tilesHigh entries, and a single-column resize of the
// maintainer's ROWS map adds its blank hexes as whole tilerows. The decoders
// used to set Row from the tilerow index and Column from the entry, and build
// Coords from (row, col), transposing every tile in memory. Nothing written
// depends on these fields, so the fixture outputs never showed it; cmd/server,
// since removed, labelled hexes with Coords and showed it.
//
// The fixtures are chosen non-square, so a transposition cannot agree by
// accident, and cover both orientations: W2025 COLUMNS 13 x 11 and W2025 ROWS
// 13 x 11.
func TestDecodedTileCoordinates(t *testing.T) {
	for _, fixture := range []string{
		"../testdata/2025-2.06-13x11-941577-blank.wxx",
		"../testdata/2025-2.06-13x11-941577-rows.wxx",
	} {
		t.Run(fixture, func(t *testing.T) {
			m, err := xmlio.ReadFile(fixture)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if len(m.Tiles.Tiles) != m.Tiles.TilesWide {
				t.Fatalf("%d columns of tiles, want tilesWide = %d", len(m.Tiles.Tiles), m.Tiles.TilesWide)
			}
			rows := m.HexOrientation == "ROWS"
			for c, column := range m.Tiles.Tiles {
				if len(column) != m.Tiles.TilesHigh {
					t.Fatalf("column %d has %d tiles, want tilesHigh = %d", c, len(column), m.Tiles.TilesHigh)
				}
				for r, tile := range column {
					if tile.Column != c || tile.Row != r {
						t.Fatalf("Tiles[%d][%d] says column %d, row %d", c, r, tile.Column, tile.Row)
					}
					var ok bool
					if rows {
						ok = tile.Coords.CubeToROffset(false) == hexg.NewOffsetCoord(c, r)
					} else {
						ok = tile.Coords.CubeToQOffset(false) == hexg.NewOffsetCoord(c, r)
					}
					if !ok {
						t.Fatalf("Tiles[%d][%d].Coords = %v, which is not (col %d, row %d)", c, r, tile.Coords, c, r)
					}
				}
			}
		})
	}
}
