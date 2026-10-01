// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"reflect"
	"testing"

	"github.com/maloquacious/wxx"
)

// riverLinesFixture is the maintainer's river-lines 2.08 map (issue #154),
// whose recipe is in testdata/README.md: the blank map flood-filled with Flat
// Farmland, and two blue lines drawn along hex edges on Above Terrain with
// Snap Points to Grid checked, the way a river runs.
const riverLinesFixture = "2025-2.08-13x11-941577-river-lines.wxx"

// riverLine1 and riverLine2 are the fixture's two lines as the file states
// their points. Every point is a hex corner, each clicked once, so no point
// repeats; line 2 ends on line 1's (1350,1050), as a tributary does.
var (
	riverLine1 = [][2]float64{
		{1125, 600}, {1200, 750}, {1125, 900}, {1200, 1050}, {1350, 1050}, {1425, 1200},
		{1350, 1350}, {1200, 1350}, {1125, 1200}, {975, 1200}, {900, 1350}, {750, 1350},
	}
	riverLine2 = [][2]float64{{1650, 750}, {1575, 900}, {1425, 900}, {1350, 1050}}
)

// TestW2025RiverLinesFixture pins what the river-lines fixture holds, so the
// tests built on it (issues #153 and #155) can rely on it: two plain Path
// shapes, not the River shape style, with the attributes the app writes for an
// unfilled line, and points that are a move-to followed by line-tos.
//
// The attributes are compared as the decoder holds them. That the encoder
// writes them back as the file spells them is TestW2025ShapesMatchSource.
func TestW2025RiverLinesFixture(t *testing.T) {
	_, m := readNotesFixture(t, riverLinesFixture)
	if len(m.Shapes) != 2 {
		t.Fatalf("got %d shapes, want 2", len(m.Shapes))
	}
	for i, line := range [][][2]float64{riverLine1, riverLine2} {
		s := m.Shapes[i]
		want := wxx.Shape_t{
			Type: "Path", IsSnapVertices: true, CreationType: "BASIC",
			ExtraLineDistance: 30, ExtraLineLength: 30, ExtraLineWidth: 10, ExtraLineSeparation: 15,
			IsWorld: true, IsContinent: true, IsKingdom: true, IsProvince: true,
			DsSpread: 0.2, DsRadius: 50, InsChoke: 0.2, InsRadius: 50,
			BbWidth: 10, BbHeight: 10, BbIterations: 3,
			MapLayer: "Above Terrain", StrokeType: "SIMPLE",
			HighestViewLevel: "WORLD", CurrentShapeViewLevel: "WORLD",
			LineCap: "SQUARE", LineJoin: "ROUND", Opacity: 1, FillRule: "NON_ZERO",
			StrokeColor: "0.0,0.0,1.0,1.0", StrokeWidth: 0.05,
			DsColor:  "1.0,0.8941176533699036,0.7686274647712708,1.0",
			InsColor: "1.0,0.8941176533699036,0.7686274647712708,1.0",
		}
		got := *s
		got.Points = nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("shape %d:\n got %+v\nwant %+v", i, got, want)
		}
		if len(s.Points) != len(line) {
			t.Errorf("shape %d: got %d points, want %d", i, len(s.Points), len(line))
			continue
		}
		for j, p := range s.Points {
			wantType := ""
			if j == 0 {
				wantType = "m"
			}
			if p.Type != wantType || p.X != line[j][0] || p.Y != line[j][1] || p.IntegerXY || p.Control != nil {
				t.Errorf("shape %d point %d: got %+v, want type %q at (%v,%v), decimal-spelled, no control points",
					i, j, *p, wantType, line[j][0], line[j][1])
			}
		}
	}
}
