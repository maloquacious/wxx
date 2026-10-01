// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"testing"

	"github.com/maloquacious/wxx"
)

// TestTileCornersMatchRiverLines checks wxx's hex geometry (issue #153)
// against the river-lines fixture (issue #154): every point of both lines is
// the hex corner issue #154 names for it, and consecutive distinct points are
// the two ends of one hex edge.
func TestTileCornersMatchRiverLines(t *testing.T) {
	_, m := readNotesFixture(t, riverLinesFixture)
	e := func(col, row int) wxx.Vertex_t { return wxx.Vertex_t{Col: col, Row: row, Corner: wxx.CornerE} }
	se := func(col, row int) wxx.Vertex_t { return wxx.Vertex_t{Col: col, Row: row, Corner: wxx.CornerSE} }
	lines := [][]wxx.Vertex_t{
		{se(4, 1), e(4, 2), se(4, 2), e(4, 3), se(5, 2), e(5, 3),
			se(5, 3), e(4, 4), se(4, 3), e(3, 3), se(3, 3), e(2, 4)},
		{e(6, 2), se(6, 2), e(5, 2), se(5, 2)},
	}
	if len(m.Shapes) != len(lines) {
		t.Fatalf("got %d shapes, want %d", len(m.Shapes), len(lines))
	}
	edges := 0
	for i, line := range lines {
		points := m.Shapes[i].Points
		if len(points) != len(line) {
			t.Fatalf("shape %d: got %d points, want %d", i, len(points), len(line))
		}
		for j, v := range line {
			got, err := m.VertexPosition(v)
			if err != nil {
				t.Fatal(err)
			}
			if want := (wxx.Position_t{X: points[j].X, Y: points[j].Y}); got != want {
				t.Errorf("shape %d point %d: %s is at %v, the file has %v", i, j, v, got, want)
			}
			if j == 0 {
				continue
			}
			ok, err := m.SharesEdge(line[j-1], v)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Errorf("shape %d: %s and %s do not share an edge", i, line[j-1], v)
			}
			edges++
		}
	}
	if edges != 11+3 {
		t.Errorf("checked %d edges, want 14", edges)
	}
}

// TestTileCornersMatchTileBorders checks the geometry against the "Add Tile
// Border" polygons the notes-shapes fixtures draw on hexes (12,0), (10,0) and
// (8,0): each polygon's six points are that hex's six corners.
func TestTileCornersMatchTileBorders(t *testing.T) {
	for _, fixture := range notesShapesFixtures {
		t.Run(fixture, func(t *testing.T) {
			_, m := readNotesFixture(t, fixture)
			corners, err := m.Corners()
			if err != nil {
				t.Fatal(err)
			}
			want := map[int]bool{12: true, 10: true, 8: true} // columns, all in row 0
			for _, s := range m.Shapes {
				if !s.IsMatchTileBorders {
					continue
				}
				if len(s.Points) != 6 {
					t.Errorf("tile-border polygon has %d points, want 6", len(s.Points))
					continue
				}
				// the hex is the one whose center is the polygon's centroid.
				var cx, cy float64
				for _, p := range s.Points {
					cx, cy = cx+p.X/6, cy+p.Y/6
				}
				col := -1
				for c := range want {
					if center, _ := m.TileCenter(c, 0); center.X == cx && center.Y == cy {
						col = c
					}
				}
				if col < 0 {
					t.Errorf("tile-border polygon centered at (%v,%v) is not on (12,0), (10,0) or (8,0)", cx, cy)
					continue
				}
				delete(want, col)
				onPolygon := map[wxx.Position_t]bool{}
				for _, p := range s.Points {
					onPolygon[wxx.Position_t{X: p.X, Y: p.Y}] = true
				}
				for _, c := range corners {
					p, _ := m.TileCorner(col, 0, c)
					if !onPolygon[p] {
						t.Errorf("(%d,0) %s at %v is not a point of the polygon", col, c, p)
					}
				}
			}
			if len(want) != 0 {
				t.Errorf("no tile-border polygon on columns %v of row 0", want)
			}
		})
	}
}
