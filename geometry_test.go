// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"math"
	"testing"
)

var (
	columnsMap = &Map_t{HexOrientation: "COLUMNS"}
	rowsMap    = &Map_t{HexOrientation: "ROWS"}
)

// TestTileCenterColumns checks the COLUMNS center formula on hexes whose
// centers the fixtures and issue #153 state, including one off the map.
func TestTileCenterColumns(t *testing.T) {
	for _, tc := range []struct {
		col, row int
		want     Position_t
	}{
		{0, 0, Position_t{150, 150}},
		{1, 0, Position_t{375, 300}},
		{8, 1, Position_t{1950, 450}},  // the notes-shapes line's first point
		{8, 4, Position_t{1950, 1350}}, // and its last
		{12, 0, Position_t{2850, 150}},
		{-1, 0, Position_t{-75, 300}}, // off the map: -1 is odd
		{0, -1, Position_t{150, -150}},
	} {
		got, err := columnsMap.TileCenter(tc.col, tc.row)
		if err != nil || got != tc.want {
			t.Errorf("TileCenter(%d, %d) = %v, %v; want %v", tc.col, tc.row, got, err, tc.want)
		}
	}
}

// TestTileCenterRows checks the ROWS center on the hex #80 confirmed in
// Worldographer: a feature at (1050,1050) sits on hex (3,4).
func TestTileCenterRows(t *testing.T) {
	for _, tc := range []struct {
		col, row int
		want     Position_t
	}{
		{0, 0, Position_t{150, 150}},
		{0, 1, Position_t{300, 375}},
		{3, 4, Position_t{1050, 1050}},
		{0, -1, Position_t{300, -75}},
	} {
		got, err := rowsMap.TileCenter(tc.col, tc.row)
		if err != nil || got != tc.want {
			t.Errorf("TileCenter(%d, %d) = %v, %v; want %v", tc.col, tc.row, got, err, tc.want)
		}
	}
}

// TestTileCornerColumns checks the corners of hex (12, 0) against the "Add
// Tile Border" polygon the notes-shapes fixtures draw there (issue #153), and
// one corner of line 1 in the river-lines fixture (issue #154).
func TestTileCornerColumns(t *testing.T) {
	for _, tc := range []struct {
		col, row int
		c        Corner_e
		want     Position_t
	}{
		{12, 0, CornerW, Position_t{2700, 150}},
		{12, 0, CornerSW, Position_t{2775, 300}},
		{12, 0, CornerSE, Position_t{2925, 300}},
		{12, 0, CornerE, Position_t{3000, 150}},
		{12, 0, CornerNE, Position_t{2925, 0}},
		{12, 0, CornerNW, Position_t{2775, 0}},
		{4, 1, CornerSE, Position_t{1125, 600}},
	} {
		got, err := columnsMap.TileCorner(tc.col, tc.row, tc.c)
		if err != nil || got != tc.want {
			t.Errorf("TileCorner(%d, %d, %s) = %v, %v; want %v", tc.col, tc.row, tc.c, got, err, tc.want)
		}
	}
}

// TestTileCornerRowsIsColumnsTransposed: a ROWS hex is a COLUMNS hex turned a
// quarter, so ROWS hex (a, b) is COLUMNS hex (b, a) with x and y swapped, and
// each corner maps to the corner its offset swaps to. No ROWS fixture has a
// shape, so this is the ROWS corner check there is.
func TestTileCornerRowsIsColumnsTransposed(t *testing.T) {
	transposed := map[Corner_e]Corner_e{
		CornerE: CornerS, CornerSE: CornerSE, CornerSW: CornerNE,
		CornerW: CornerN, CornerNW: CornerNW, CornerNE: CornerSW,
	}
	n := 0
	for a := -3; a <= 5; a++ {
		for b := -3; b <= 5; b++ {
			for cc, rc := range transposed {
				pc, err := columnsMap.TileCorner(b, a, cc)
				if err != nil {
					t.Fatal(err)
				}
				pr, err := rowsMap.TileCorner(a, b, rc)
				if err != nil {
					t.Fatal(err)
				}
				if pr.X != pc.Y || pr.Y != pc.X {
					t.Errorf("ROWS (%d,%d) %s = %v; COLUMNS (%d,%d) %s = %v, want it transposed", a, b, rc, pr, b, a, cc, pc)
				}
				n++
			}
		}
	}
	if n == 0 {
		t.Fatal("no corners compared")
	}
}

// TestGeometryIsExact: every center and corner is a whole multiple of 75, so
// the positions compare exactly.
func TestGeometryIsExact(t *testing.T) {
	for _, m := range []*Map_t{columnsMap, rowsMap} {
		corners, err := m.Corners()
		if err != nil {
			t.Fatal(err)
		}
		for col := -2; col < 15; col++ {
			for row := -2; row < 13; row++ {
				p, _ := m.TileCenter(col, row)
				ps := []Position_t{p}
				for _, c := range corners {
					p, err := m.TileCorner(col, row, c)
					if err != nil {
						t.Fatal(err)
					}
					ps = append(ps, p)
				}
				for _, p := range ps {
					if math.Mod(p.X, 75) != 0 || math.Mod(p.Y, 75) != 0 {
						t.Errorf("%s (%d,%d): %v is not a multiple of 75", m.HexOrientation, col, row, p)
					}
				}
			}
		}
	}
}

// TestTileCornerRefused: a corner of the other orientation, a corner that is
// not one, and a map with no orientation are errors.
func TestTileCornerRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		m    *Map_t
		c    Corner_e
		want error
	}{
		{"COLUMNS n", columnsMap, CornerN, ErrInvalidCorner},
		{"COLUMNS s", columnsMap, CornerS, ErrInvalidCorner},
		{"ROWS e", rowsMap, CornerE, ErrInvalidCorner},
		{"ROWS w", rowsMap, CornerW, ErrInvalidCorner},
		{"zero corner", columnsMap, 0, ErrInvalidCorner},
		{"out of range", rowsMap, CornerNE + 1, ErrInvalidCorner},
		{"no orientation", &Map_t{}, CornerE, ErrInvalidHexOrientation},
		{"lower case", &Map_t{HexOrientation: "columns"}, CornerE, ErrInvalidHexOrientation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.m.TileCorner(1, 2, tc.c)
			if !errors.Is(err, tc.want) {
				t.Errorf("TileCorner: err = %v, want %v", err, tc.want)
			}
		})
	}
	if _, err := (&Map_t{}).TileCenter(0, 0); !errors.Is(err, ErrInvalidHexOrientation) {
		t.Errorf("TileCenter: err = %v, want %v", err, ErrInvalidHexOrientation)
	}
}

// TestCanonicalVertexIsUniqueAndAgrees: every corner of the grid belongs to
// three hexes, exactly one of which names it by a canonical corner; every name
// of a corner canonicalizes to that one; and two corners at different
// positions never share a canonical name.
func TestCanonicalVertexIsUniqueAndAgrees(t *testing.T) {
	for _, m := range []*Map_t{columnsMap, rowsMap} {
		g, _ := m.hexGeometry()
		byPosition := map[Position_t]Vertex_t{}
		byCanonical := map[Vertex_t]Position_t{}
		for col := -3; col <= 6; col++ {
			for row := -3; row <= 6; row++ {
				for i, c := range g.corners {
					v := Vertex_t{col, row, c}
					p := g.corner(col, row, i)
					sharers := g.sharers(col, row, p)
					if len(sharers) != 3 {
						t.Fatalf("%s %s: %d hexes share it, want 3: %v", m.HexOrientation, v, len(sharers), sharers)
					}
					canonical := 0
					for _, s := range sharers {
						if s.Corner == g.canonical[0] || s.Corner == g.canonical[1] {
							canonical++
						}
					}
					if canonical != 1 {
						t.Fatalf("%s %s: %d of %v are canonical, want 1", m.HexOrientation, v, canonical, sharers)
					}
					cv, err := m.CanonicalVertex(v)
					if err != nil {
						t.Fatal(err)
					}
					if cp, _ := m.VertexPosition(cv); cp != p {
						t.Errorf("%s %s at %v: canonical %s is at %v", m.HexOrientation, v, p, cv, cp)
					}
					if prev, ok := byPosition[p]; ok && prev != cv {
						t.Errorf("%s %v: canonical %s and %s", m.HexOrientation, p, prev, cv)
					}
					byPosition[p] = cv
					if prev, ok := byCanonical[cv]; ok && prev != p {
						t.Errorf("%s %s: canonical for %v and %v", m.HexOrientation, cv, prev, p)
					}
					byCanonical[cv] = p
				}
			}
		}
		if len(byPosition) == 0 {
			t.Fatalf("%s: no corners checked", m.HexOrientation)
		}
	}
}

// TestSameCornerExample is issue #153's example: hex (4, 2)'s se is (5, 2)'s w
// and (4, 3)'s ne, and se is its canonical name.
func TestSameCornerExample(t *testing.T) {
	se, w, ne := Vertex_t{4, 2, CornerSE}, Vertex_t{5, 2, CornerW}, Vertex_t{4, 3, CornerNE}
	for _, v := range []Vertex_t{se, w, ne} {
		same, err := columnsMap.SameCorner(se, v)
		if err != nil || !same {
			t.Errorf("SameCorner(%s, %s) = %v, %v; want true", se, v, same, err)
		}
		cv, err := columnsMap.CanonicalVertex(v)
		if err != nil || cv != se {
			t.Errorf("CanonicalVertex(%s) = %s, %v; want %s", v, cv, err, se)
		}
	}
	if same, _ := columnsMap.SameCorner(se, Vertex_t{4, 2, CornerE}); same {
		t.Errorf("SameCorner(%s, (4,2) e) = true, want false", se)
	}
	if _, err := columnsMap.SameCorner(se, Vertex_t{4, 2, CornerN}); !errors.Is(err, ErrInvalidCorner) {
		t.Errorf("SameCorner with n on COLUMNS: err = %v, want %v", err, ErrInvalidCorner)
	}
}

// TestSharesEdge: consecutive corners of a hex share an edge, named from any of
// their hexes; a corner and itself, and corners two apart, do not.
func TestSharesEdge(t *testing.T) {
	for _, m := range []*Map_t{columnsMap, rowsMap} {
		corners, _ := m.Corners()
		for col := -1; col <= 2; col++ {
			for row := -1; row <= 2; row++ {
				for i, c := range corners {
					a := Vertex_t{col, row, c}
					next := Vertex_t{col, row, corners[(i+1)%6]}
					skip := Vertex_t{col, row, corners[(i+2)%6]}
					opposite := Vertex_t{col, row, corners[(i+3)%6]}
					if ok, err := m.SharesEdge(a, next); err != nil || !ok {
						t.Errorf("%s SharesEdge(%s, %s) = %v, %v; want true", m.HexOrientation, a, next, ok, err)
					}
					if ok, err := m.SharesEdge(next, a); err != nil || !ok {
						t.Errorf("%s SharesEdge(%s, %s) = %v, %v; want true", m.HexOrientation, next, a, ok, err)
					}
					for _, b := range []Vertex_t{a, skip, opposite} {
						if ok, err := m.SharesEdge(a, b); err != nil || ok {
							t.Errorf("%s SharesEdge(%s, %s) = %v, %v; want false", m.HexOrientation, a, b, ok, err)
						}
					}
				}
			}
		}
	}
	// named from other hexes: (4,2) se is (5,2) w; (4,2) e is (5,2) nw.
	if ok, err := columnsMap.SharesEdge(Vertex_t{5, 2, CornerW}, Vertex_t{4, 2, CornerE}); err != nil || !ok {
		t.Errorf("SharesEdge((5,2) w, (4,2) e) = %v, %v; want true", ok, err)
	}
	// the column parity mistake: (5,1) is odd, so it is half a hex lower than
	// (4,1), and its w is (4,1)'s se moved down 150, not (4,1)'s se. No
	// COLUMNS edge is vertical.
	if ok, err := columnsMap.SharesEdge(Vertex_t{4, 1, CornerSE}, Vertex_t{5, 1, CornerW}); err != nil || ok {
		t.Errorf("SharesEdge((4,1) se, (5,1) w) = %v, %v; want false", ok, err)
	}
}

func TestParseCorner(t *testing.T) {
	for c := CornerE; c <= CornerNE; c++ {
		got, err := ParseCorner(c.String())
		if err != nil || got != c {
			t.Errorf("ParseCorner(%q) = %v, %v; want %v", c.String(), got, err, c)
		}
	}
	for _, name := range []string{"", "E", "north", "x"} {
		if _, err := ParseCorner(name); !errors.Is(err, ErrInvalidCorner) {
			t.Errorf("ParseCorner(%q): err = %v, want %v", name, err, ErrInvalidCorner)
		}
	}
}
