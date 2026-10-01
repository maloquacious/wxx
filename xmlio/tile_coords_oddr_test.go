// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// TestDecodedTileCoordinatesLayout pins the offset layout behind Tile_t.Coords
// with literal cube coordinates and hex distances (issue #130).
//
// Worldographer's ROWS layout staggers odd rows right and its COLUMNS layout
// staggers odd columns down (#80; schema/README.md). The maintainer confirmed
// the ROWS stagger in the app on 2026-10-01. In the redblobgames convention
// those are odd-r and odd-q. Before #130 the decoder built ROWS coordinates
// with the even-r formula. TestDecodedTileCoordinates passed under
// both formulas, because it converts Coords back through the same hexg call
// that made them. This test therefore checks values written out by hand, and
// checks adjacency, which does not depend on what hexg calls its layouts:
//
//   - odd-r: odd row 1 sits right of row 0, so Tiles[0][1] (column 0, row 1)
//     touches Tiles[0][0] and Tiles[1][0]. Under even-r it touches Tiles[0][0]
//     and not Tiles[1][0].
//   - odd-q: odd column 1 sits below column 0, so Tiles[1][0] touches
//     Tiles[0][0] and Tiles[0][1]. Under even-q it does not touch Tiles[0][1].
//
// Tiles is indexed [column][row] (issue #85).
func TestDecodedTileCoordinatesLayout(t *testing.T) {
	type cell struct{ col, row int }
	type cube struct{ q, r, s int }
	type distance struct {
		a, b cell
		want int
	}

	// odd-r: q = col - (row - row&1)/2, r = row.
	rowsCoords := map[cell]cube{
		{0, 0}: {0, 0, 0},
		{0, 1}: {0, 1, -1}, // even-r gives (-1, 1, 0)
		{1, 0}: {1, 0, -1},
		{4, 2}: {3, 2, -5},
		{5, 3}: {4, 3, -7}, // even-r gives (3, 3, -6)
		{5, 4}: {3, 4, -7},
	}
	rowsDistances := []distance{
		{cell{0, 1}, cell{0, 0}, 1},
		{cell{0, 1}, cell{1, 0}, 1}, // 2 under even-r
		// odd row 3 at interior column 5: its row 2 neighbours are columns 5 and 6.
		{cell{5, 3}, cell{5, 2}, 1},
		{cell{5, 3}, cell{6, 2}, 1}, // 2 under even-r
		{cell{5, 3}, cell{4, 2}, 2}, // 1 under even-r
		{cell{5, 3}, cell{5, 4}, 1},
		{cell{5, 3}, cell{6, 4}, 1}, // 2 under even-r
		{cell{5, 3}, cell{4, 3}, 1},
		{cell{5, 3}, cell{6, 3}, 1},
		// even row 4 at column 5: its row 3 neighbours are columns 4 and 5.
		{cell{5, 4}, cell{4, 3}, 1},
		{cell{5, 4}, cell{5, 3}, 1},
		{cell{5, 4}, cell{6, 3}, 2},
	}

	// odd-q: q = col, r = row - (col - col&1)/2.
	columnsCoords := map[cell]cube{
		{0, 0}: {0, 0, 0},
		{1, 0}: {1, 0, -1},
		{0, 1}: {0, 1, -1},
		{2, 0}: {2, -1, -1},
		{5, 3}: {5, 1, -6},
	}
	columnsDistances := []distance{
		{cell{1, 0}, cell{0, 0}, 1},
		{cell{1, 0}, cell{0, 1}, 1}, // 2 under even-q
		{cell{1, 0}, cell{2, 0}, 1},
		{cell{1, 0}, cell{2, 1}, 1}, // 2 under even-q
		{cell{1, 0}, cell{0, 2}, 2},
		// odd column 5 at interior row 3: its column 4 neighbours are rows 3 and 4.
		{cell{5, 3}, cell{4, 3}, 1},
		{cell{5, 3}, cell{4, 4}, 1},
		{cell{5, 3}, cell{4, 2}, 2},
	}

	for _, tc := range []struct {
		fixture     string
		orientation string
		coords      map[cell]cube
		distances   []distance
	}{
		{"../testdata/2025-2.06-13x11-941577-rows.wxx", "ROWS", rowsCoords, rowsDistances},
		{"../testdata/2025-2.07-13x11-941577-rows.wxx", "ROWS", rowsCoords, rowsDistances},
		{"../testdata/2025-2.08-13x11-941577-rows.wxx", "ROWS", rowsCoords, rowsDistances},
		{"../testdata/2025-2.06-13x11-941577-blank.wxx", "COLUMNS", columnsCoords, columnsDistances},
		{"../testdata/2025-2.07-13x11-941577-blank.wxx", "COLUMNS", columnsCoords, columnsDistances},
		{"../testdata/2025-2.08-13x11-941577-blank.wxx", "COLUMNS", columnsCoords, columnsDistances},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			m, err := xmlio.ReadFile(tc.fixture)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			// Guard against a vacuous pass: the fixture must be the
			// orientation the table was written for, and every cell the
			// table names must exist.
			if m.HexOrientation != tc.orientation {
				t.Fatalf("hexOrientation %q, want %q", m.HexOrientation, tc.orientation)
			}
			if m.Tiles == nil || len(m.Tiles.Tiles) == 0 {
				t.Fatalf("no tiles decoded")
			}
			tile := func(c cell) (q, r, s int, ok bool) {
				if c.col >= len(m.Tiles.Tiles) || c.row >= len(m.Tiles.Tiles[c.col]) || m.Tiles.Tiles[c.col][c.row] == nil {
					t.Errorf("Tiles[%d][%d] missing", c.col, c.row)
					return 0, 0, 0, false
				}
				q, r, s = m.Tiles.Tiles[c.col][c.row].Coords.QRS()
				return q, r, s, true
			}

			checked := 0
			for c, want := range tc.coords {
				q, r, s, ok := tile(c)
				if !ok {
					continue
				}
				checked++
				if got := (cube{q, r, s}); got != want {
					t.Errorf("%s Tiles[%d][%d] (column %d, row %d): Coords = (%d, %d, %d), want (%d, %d, %d)",
						tc.orientation, c.col, c.row, c.col, c.row, got.q, got.r, got.s, want.q, want.r, want.s)
				}
			}
			for _, d := range tc.distances {
				aq, ar, as, aok := tile(d.a)
				bq, br, bs, bok := tile(d.b)
				if !aok || !bok {
					continue
				}
				checked++
				got := max(abs(aq-bq), abs(ar-br), abs(as-bs))
				if got != d.want {
					t.Errorf("%s distance Tiles[%d][%d] (%d, %d, %d) to Tiles[%d][%d] (%d, %d, %d) = %d, want %d",
						tc.orientation, d.a.col, d.a.row, aq, ar, as, d.b.col, d.b.row, bq, br, bs, got, d.want)
				}
			}
			if want := len(tc.coords) + len(tc.distances); checked != want {
				t.Fatalf("checked %d of %d cells and pairs", checked, want)
			}
		})
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
