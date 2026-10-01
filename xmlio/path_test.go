// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// riverLineVertices are the river-lines fixture's two lines as hex corners
// (issue #154), in the order they were drawn.
func riverLineVertices() [][]wxx.Vertex_t {
	e := func(col, row int) wxx.Vertex_t { return wxx.Vertex_t{Col: col, Row: row, Corner: wxx.CornerE} }
	se := func(col, row int) wxx.Vertex_t { return wxx.Vertex_t{Col: col, Row: row, Corner: wxx.CornerSE} }
	return [][]wxx.Vertex_t{
		{se(4, 1), e(4, 2), se(4, 2), e(4, 3), se(5, 2), e(5, 3),
			se(5, 3), e(4, 4), se(4, 3), e(3, 3), se(3, 3), e(2, 4)},
		{e(6, 2), se(6, 2), e(5, 2), se(5, 2)},
	}
}

// TestNewEdgePathMatchesRiverLines is issue #155's acceptance test: a new map,
// with #154's two lines built by NewEdgePath with the defaults (blue, 0.05,
// Above Terrain), written as each registered application version, holds
// <shape> elements that match the river-lines fixture's attribute for
// attribute and point for point, as the strings the documents spell.
//
// Only 2.08 has a river-lines fixture so far, so every version is compared
// with it. For 2.06 and 2.07 the evidence that the app writes the same line
// is their notes-shapes fixtures: TestNewEdgePathMatchesNotesShapesLine checks
// each version's line there.
func TestNewEdgePathMatchesRiverLines(t *testing.T) {
	src, _ := readNotesFixture(t, riverLinesFixture)
	want := shapeElements(t, riverLinesFixture, src)
	if len(want) != 2 {
		t.Fatalf("%s: %d shapes, want 2", riverLinesFixture, len(want))
	}
	for _, app := range []string{"2.06", "2.07", "2.08"} {
		t.Run(app, func(t *testing.T) {
			m, err := xmlio.NewMap(13, 11, xmlio.WithApp(app))
			if err != nil {
				t.Fatal(err)
			}
			for i, line := range riverLineVertices() {
				s, err := m.NewEdgePath(line)
				if err != nil {
					t.Fatalf("line %d: %v", i+1, err)
				}
				m.Shapes = append(m.Shapes, s)
			}
			if err := m.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			out, err := xmlio.MarshalXML(m, app)
			if err != nil {
				t.Fatalf("MarshalXML: %v", err)
			}
			got := shapeElements(t, "output", out)
			if len(got) != len(want) {
				t.Fatalf("wrote %d shapes, want %d", len(got), len(want))
			}
			for i := range want {
				compareAttrSets(t, app, "shape", i, want[i].attrs, got[i].attrs)
				if len(got[i].points) != len(want[i].points) {
					t.Errorf("shape %d: wrote %d points, fixture has %d", i, len(got[i].points), len(want[i].points))
					continue
				}
				for j := range want[i].points {
					compareAttrSets(t, app, "shape p", j, want[i].points[j], got[i].points[j])
				}
			}
		})
	}
}

// TestNewEdgePathMatchesNotesShapesLine compares a NewEdgePath line's
// <shape> start tag with the line each notes-shapes fixture draws, for 2.06,
// 2.07 and 2.08. That line is drawn with Snap Points to Grid in blue at the
// default width, like the river lines, but on Terrain Water (and the 2.06
// fixture has a second one on Below All), so each line is built on its
// path's layer. Their points are hex centers, not corners, so only the start
// tag is compared.
func TestNewEdgePathMatchesNotesShapesLine(t *testing.T) {
	for _, fixture := range notesShapesFixtures {
		t.Run(fixture, func(t *testing.T) {
			src, m := readNotesFixture(t, fixture)
			var paths []shapeElement_t
			for _, s := range shapeElements(t, fixture, src) {
				if s.attrs["type"] == "Path" {
					paths = append(paths, s)
				}
			}
			if len(paths) == 0 {
				t.Fatalf("%s: no Path shape", fixture)
			}
			m.Shapes = nil
			for _, p := range paths {
				s, err := m.NewEdgePath(riverLineVertices()[1], wxx.WithMapLayer(p.attrs["mapLayer"]))
				if err != nil {
					t.Fatal(err)
				}
				m.Shapes = append(m.Shapes, s)
			}
			out, err := xmlio.MarshalXML(m, sameVersionTarget(t, fixture))
			if err != nil {
				t.Fatalf("MarshalXML: %v", err)
			}
			got := shapeElements(t, "output", out)
			for i, p := range paths {
				compareAttrSets(t, fixture, "Path shape", i, p.attrs, got[i].attrs)
			}
		})
	}
}

// TestNewEdgePathAlongMapEdge: corners named by hexes off the map are allowed,
// so a path can run down the map's left edge, and the map it is added to still
// validates and round-trips.
func TestNewEdgePathAlongMapEdge(t *testing.T) {
	m, err := xmlio.NewMap(13, 11)
	if err != nil {
		t.Fatal(err)
	}
	// (-1, 0) is odd, so it is half a hex lower than (0, 0): its ne is (0,0)'s
	// w, its e is (0,0)'s sw, and (-1, 1)'s ne is (0,1)'s w.
	line := []wxx.Vertex_t{{Col: -1, Row: 0, Corner: wxx.CornerNE}, {Col: -1, Row: 0, Corner: wxx.CornerE}, {Col: -1, Row: 1, Corner: wxx.CornerNE}}
	s, err := m.NewEdgePath(line)
	if err != nil {
		t.Fatalf("NewEdgePath: %v", err)
	}
	want := []wxx.Position_t{{X: 0, Y: 150}, {X: 75, Y: 300}, {X: 0, Y: 450}}
	for i, p := range s.Points {
		if (wxx.Position_t{X: p.X, Y: p.Y}) != want[i] {
			t.Errorf("point %d at (%v,%v), want %v", i, p.X, p.Y, want[i])
		}
	}
	m.Shapes = append(m.Shapes, s)
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(xmlio.CurrentApp()).Encode(&buf, m); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	back, err := xmlio.NewDecoder().Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(back.Shapes) != 1 || len(back.Shapes[0].Points) != 3 || back.Shapes[0].Points[2].Y != 450 {
		t.Errorf("read back %d shapes, want the one path of 3 points", len(back.Shapes))
	}
}

// TestNewEdgePathRefused: each malformed path is refused with the kind of
// error and a message that names the problem.
func TestNewEdgePathRefused(t *testing.T) {
	m, err := xmlio.NewMap(13, 11)
	if err != nil {
		t.Fatal(err)
	}
	v := func(col, row int, c wxx.Corner_e) wxx.Vertex_t { return wxx.Vertex_t{Col: col, Row: row, Corner: c} }
	for _, tc := range []struct {
		name     string
		vertices []wxx.Vertex_t
		opts     []wxx.PathOption
		want     error
		message  string
	}{
		{"gap", []wxx.Vertex_t{v(4, 1, wxx.CornerSE), v(4, 2, wxx.CornerE), v(4, 3, wxx.CornerE)}, nil,
			wxx.ErrInvalidShape, "vertices 1 and 2, (4,2) e and (4,3) e, are not the two ends of one hex edge"},
		// the column-parity mistake: (5,1) is odd, so its w is not (4,1)'s se
		{"parity", []wxx.Vertex_t{v(4, 1, wxx.CornerSE), v(5, 1, wxx.CornerW)}, nil,
			wxx.ErrInvalidShape, "(4,1) se and (5,1) w"},
		{"repeated corner", []wxx.Vertex_t{v(4, 2, wxx.CornerSE), v(5, 2, wxx.CornerW)}, nil,
			wxx.ErrInvalidShape, "(4,2) se and (5,2) w"},
		{"one vertex", []wxx.Vertex_t{v(4, 2, wxx.CornerSE)}, nil,
			wxx.ErrInvalidShape, "1 vertex(es), want at least 2"},
		{"ROWS corner", []wxx.Vertex_t{v(4, 2, wxx.CornerN), v(4, 2, wxx.CornerNE)}, nil,
			wxx.ErrInvalidCorner, "vertex 0"},
		{"unknown layer", riverLineVertices()[1], []wxx.PathOption{wxx.WithMapLayer("Rivers")},
			wxx.ErrUnknownMapLayer, `"Rivers"`},
		{"zero width", riverLineVertices()[1], []wxx.PathOption{wxx.WithStrokeWidth(0)},
			wxx.ErrInvalidShape, "stroke width 0"},
		{"color out of range", riverLineVertices()[1], []wxx.PathOption{wxx.WithStrokeColor(wxx.RGBA_t{R: 255, A: 1})},
			wxx.ErrInvalidColorAttribute, "stroke color"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := m.NewEdgePath(tc.vertices, tc.opts...)
			if s != nil || !errors.Is(err, tc.want) {
				t.Fatalf("NewEdgePath = %v, %v; want nil, %v", s, err, tc.want)
			}
			if !strings.Contains(err.Error(), tc.message) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.message)
			}
		})
	}
}

// TestNewPath: any points, not snapped; the options reach the shape; a
// repeated point is kept, as the app writes one; and fewer than two points or
// a non-finite one is refused.
func TestNewPath(t *testing.T) {
	m, err := xmlio.NewMap(13, 11)
	if err != nil {
		t.Fatal(err)
	}
	points := []wxx.Position_t{{X: 10.5, Y: 20.25}, {X: 10.5, Y: 20.25}, {X: 1234.5, Y: 99}}
	s, err := m.NewPath(points,
		wxx.WithMapLayer("Terrain Land"),
		wxx.WithStrokeWidth(0.2),
		wxx.WithStrokeColor(wxx.RGBA_t{R: 1, G: 0.5, B: 0.001, A: 1}),
		wxx.WithGMOnly(true))
	if err != nil {
		t.Fatalf("NewPath: %v", err)
	}
	if s.IsSnapVertices || !s.IsGMOnly || s.MapLayer != "Terrain Land" || s.StrokeWidth != 0.2 || s.StrokeColor != "1.0,0.5,0.001,1.0" {
		t.Errorf("NewPath: got snap %v, GM only %v, layer %q, width %v, color %q",
			s.IsSnapVertices, s.IsGMOnly, s.MapLayer, s.StrokeWidth, s.StrokeColor)
	}
	if len(s.Points) != 3 || s.Points[0].Type != "m" || s.Points[1].Type != "" || s.Points[2].X != 1234.5 {
		t.Errorf("NewPath: points %+v %+v %+v", *s.Points[0], *s.Points[1], *s.Points[2])
	}
	m.Shapes = append(m.Shapes, s)
	if _, err := xmlio.MarshalXML(m, xmlio.CurrentApp()); err != nil {
		t.Errorf("MarshalXML: %v", err)
	}

	for _, bad := range [][]wxx.Position_t{nil, {{X: 1, Y: 1}}, {{X: 1, Y: 1}, {X: 0, Y: nan()}}} {
		if s, err := m.NewPath(bad); s != nil || !errors.Is(err, wxx.ErrInvalidShape) {
			t.Errorf("NewPath(%v) = %v, %v; want nil, %v", bad, s, err, wxx.ErrInvalidShape)
		}
	}
}

func nan() float64 {
	var zero float64
	return zero / zero
}
