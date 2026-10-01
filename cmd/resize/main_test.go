// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
	"github.com/maloquacious/wxx/xmlio"
)

// runAsResize is set in the environment of a child process to make the test
// binary run main() instead of the tests, so a test can run the command end to
// end -- flags, exit status and stderr -- without building a separate binary.
const runAsResize = "WXX_TEST_RUN_RESIZE"

func TestMain(m *testing.M) {
	if os.Getenv(runAsResize) == "1" {
		// os.Args[0] stays the test binary; the flags after "--" are resize's.
		for i, a := range os.Args {
			if a == "--" {
				os.Args = append([]string{"resize"}, os.Args[i+1:]...)
				break
			}
		}
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// resize runs the command with args and returns its exit code and stderr.
func resize(t *testing.T, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), runAsResize+"=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), stderr.String()
	} else if err != nil {
		t.Fatalf("run resize: %v", err)
	}
	return 0, stderr.String()
}

// TestOverCropFailsCleanly: cropping more than the map has must be a one-line
// error and a non-zero exit, not a panic (issue #61). cmd/crop, deleted by
// #61, panicked this way; resize, the command that replaces it for cropping,
// did too, because its size check ran after the allocation it was guarding.
func TestOverCropFailsCleanly(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	// The fixture is 13 x 11; removing 14 columns leaves -1.
	code, stderr := resize(t,
		"-input", filepath.Join("..", "..", "testdata", "2025-2.06-13x11-941577-blank.wxx"),
		"-output", out,
		"-left", "-2", "-right", "-12")
	if code == 0 {
		t.Fatalf("exit 0, want a failure; stderr:\n%s", stderr)
	}
	if strings.Contains(stderr, "panic:") {
		t.Fatalf("resize panicked:\n%s", stderr)
	}
	if !strings.Contains(stderr, "smaller than 2 x 2 (this resize gives -1 x 11)") {
		t.Errorf("stderr does not explain the failure:\n%s", stderr)
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("output file exists after a failed resize (stat err = %v)", err)
	}
}

// TestCrop: a crop within the map's bounds writes the smaller map. It pins the
// replacement for cmd/crop's job, which README now sends users to resize for.
func TestCrop(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	code, stderr := resize(t,
		"-input", filepath.Join("..", "..", "testdata", "2025-2.06-13x11-941577-blank.wxx"),
		"-output", out,
		"-left", "-2", "-bottom", "-1")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr:\n%s", code, stderr)
	}
	m, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if m.Tiles.TilesWide != 11 || m.Tiles.TilesHigh != 10 {
		t.Errorf("output is %d x %d, want 11 x 10 (13 x 11 less 2 columns and 1 row)", m.Tiles.TilesWide, m.Tiles.TilesHigh)
	}
}

// positionedMap writes a 13 x 11 COLUMNS map (the 2.06 blank fixture) carrying
// one of each positioned element #79 is about, and returns its path:
//
//   - a layered-terrain placement at hex (4,2) and one at hex (0,0), the corner
//     a naive "x <= 0" test would wrongly drop;
//   - a triangle whose points sit around hex (4,2);
//   - a note at hex (4,2)'s center, with its key repeating the position.
//
// Hex (4,2)'s center is (150+225*4, 150+300*2) = (1050, 750); its corner, where
// a placement is stated, is (900, 600).
func positionedMap(t *testing.T) string {
	t.Helper()
	m, err := xmlio.ReadFile(filepath.Join("..", "..", "testdata", "2025-2.06-13x11-941577-blank.wxx"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	m.ExtraTerrain = &wxx.ExtraTerrain_t{MapLayers: []*wxx.ExtraTerrainLayer_t{{
		Name: "Below All",
		Terrain: []*wxx.TerrainAndLocation_t{
			{Terrain: "Classic/Flat Beach", Elevation: 1, X: 900, Y: 600},
			{Terrain: "Classic/Flat Beach", Elevation: 1, X: 0, Y: 0},
		},
	}}}
	m.Shapes = []*wxx.Shape_t{{
		Type: "Polygon", MapLayer: "Above Terrain", CreationType: "BASIC",
		IsWorld: true, HighestViewLevel: "WORLD", CurrentShapeViewLevel: "WORLD",
		StrokeColor: "1.0,1.0,1.0,1.0", DsColor: "null", InsColor: "null", StrokeType: "SIMPLE", Opacity: 1,
		Points: []*wxx.Point_t{{X: 1000, Y: 700, IntegerXY: true}, {X: 1100, Y: 700, IntegerXY: true,
			Type: "c", Control: &wxx.CurveControl_t{CX1: 1020, CY1: 650, CX2: 1080, CY2: 0}}, {X: 1050, Y: 800, IntegerXY: true}},
	}}
	m.Notes = []*wxx.Note_t{{
		Location: &wxx.NoteLocation_t{ViewLevel: "WORLD", X: 1050, Y: 750},
		Color:    &wxx.RGBA_t{R: 1, G: 1, A: 1}, Title: "Here",
	}}
	path := filepath.Join(t.TempDir(), "positioned.wxx")
	if err := xmlio.WriteFile(path, m, "2.06"); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return path
}

// TestResizeMovesEverything: growing the map at the left and top moves every
// positioned element with the tiles (issue #79). Before #79 only features and
// labels moved, so layered terrain, shapes and notes ended up on the wrong hexes.
func TestResizeMovesEverything(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	if code, stderr := resize(t, "-input", positionedMap(t), "-output", out, "-left", "2", "-top", "1"); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr)
	}
	m, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	// 2 columns and 1 row: +450, +300. Hex (4,2) becomes (6,3).
	const dx, dy = 450, 300

	placements := m.ExtraTerrain.MapLayers[0].Terrain
	if len(placements) != 2 {
		t.Fatalf("%d placement(s), want 2", len(placements))
	}
	for i, want := range [][2]float64{{900 + dx, 600 + dy}, {0 + dx, 0 + dy}} {
		if got := [2]float64{placements[i].X, placements[i].Y}; got != want {
			t.Errorf("placement %d at %v, want %v", i, got, want)
		}
	}

	if len(m.Shapes) != 1 {
		t.Fatalf("%d shape(s), want 1", len(m.Shapes))
	}
	for i, want := range [][2]float64{{1000 + dx, 700 + dy}, {1100 + dx, 700 + dy}, {1050 + dx, 800 + dy}} {
		p := m.Shapes[0].Points[i]
		if got := [2]float64{p.X, p.Y}; got != want {
			t.Errorf("shape point %d at %v, want %v", i, got, want)
		}
		// integer-spelled points move by whole hex steps and stay so (#94)
		if !p.IntegerXY {
			t.Errorf("shape point %d: IntegerXY lost in the move", i)
		}
	}
	// a curve point's control points move with it (#94)
	if c := m.Shapes[0].Points[1].Control; c == nil || *c != (wxx.CurveControl_t{CX1: 1020 + dx, CY1: 650 + dy, CX2: 1080 + dx, CY2: 0 + dy}) {
		t.Errorf("control points = %+v, want moved by (%v,%v)", c, dx, dy)
	}

	if len(m.Notes) != 1 {
		t.Fatalf("%d note(s), want 1", len(m.Notes))
	}
	// The on-disk @key repeats the position; ReadFile refuses a key that
	// disagrees with the <location>, so reading the output at all shows the
	// key moved with it.
	n := m.Notes[0]
	if n.Location == nil || n.Location.X != 1050+dx || n.Location.Y != 750+dy {
		t.Errorf("note at %+v, want (%v,%v)", n.Location, 1050+dx, 750+dy)
	}
}

// TestResizeCropDropsWhatFallsOff: cropping removes positioned elements whose
// place is cropped away and moves the rest (issue #79). Cropping 2 columns from
// the left removes hex (0,0) and its placement, and moves hex (4,2) to (2,2).
// Cropping 8 more from the right leaves 3 columns, so the shape and note near
// hex (2,2)'s center survive only if they moved with it.
func TestResizeCropDropsWhatFallsOff(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	if code, stderr := resize(t, "-input", positionedMap(t), "-output", out, "-left", "-2", "-right", "-8"); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr)
	}
	m, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if m.Tiles.TilesWide != 3 {
		t.Fatalf("output is %d wide, want 3", m.Tiles.TilesWide)
	}
	placements := m.ExtraTerrain.MapLayers[0].Terrain
	if len(placements) != 1 || placements[0].X != 900-450 || placements[0].Y != 600 {
		t.Errorf("placements = %d, want 1 at (450,600): the (0,0) one is cropped away, the (4,2) one moves to (2,2)", len(placements))
	}
	if len(m.Shapes) != 1 || m.Shapes[0].Points[0].X != 1000-450 {
		t.Errorf("shape not kept and moved: %d shape(s)", len(m.Shapes))
	}
	if len(m.Notes) != 1 || m.Notes[0].Location.X != 1050-450 {
		t.Errorf("note not kept and moved: %d note(s)", len(m.Notes))
	}

	// Crop hex (2,2) itself away too: everything that sat on it goes.
	out2 := filepath.Join(t.TempDir(), "out2.wxx")
	if code, stderr := resize(t, "-input", out, "-output", out2, "-right", "-1"); code != 0 {
		t.Fatalf("second crop: exit %d; stderr:\n%s", code, stderr)
	}
	m2, err := xmlio.ReadFile(out2)
	if err != nil {
		t.Fatalf("read second output: %v", err)
	}
	if n := len(m2.ExtraTerrain.MapLayers[0].Terrain); n != 0 {
		t.Errorf("%d placement(s) survive on a map without their hex, want 0", n)
	}
	if len(m2.Shapes) != 0 || len(m2.Notes) != 0 {
		t.Errorf("%d shape(s) and %d note(s) survive off the map, want 0", len(m2.Shapes), len(m2.Notes))
	}
}

// TestResizeKeepsFirstColumnPlacement: a placement in column 0 is stated at
// x = 0, the hex's corner, so it must be kept by its hex and not by a point
// test that treats x = 0 as off the map (issue #79). Adding rows only leaves it
// at x = 0.
func TestResizeKeepsFirstColumnPlacement(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	if code, stderr := resize(t, "-input", positionedMap(t), "-output", out, "-top", "1"); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr)
	}
	m, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	placements := m.ExtraTerrain.MapLayers[0].Terrain
	if len(placements) != 2 {
		t.Fatalf("%d placement(s), want 2: the column-0 placement at x = 0 was dropped", len(placements))
	}
	if got := [2]float64{placements[1].X, placements[1].Y}; got != [2]float64{0, 300} {
		t.Errorf("column-0 placement at %v, want [0 300]", got)
	}
}

// rowsMap writes a 13 x 11 ROWS map carrying a feature on hex (3,4), a
// triangle and a note around it, and a layered-terrain placement on hex (2,1),
// and returns its path. It is the 2.06 blank fixture turned to ROWS, laid out
// as the maintainer's 2.08 ROWS sample is: hexes 40 wide and 46.18 high,
// columns 300 apart, rows 225 apart, odd rows 150 to the right.
//
// Hex (3,4) has its center at (150+300*3, 150+225*4) = (1050,1050), which is
// where the maintainer's sample states its cathedral. Hex (2,1)'s corner, where
// a placement is stated, is (300*2+150, 225*1) = (750,225).
func rowsMap(t *testing.T) string {
	t.Helper()
	m, err := xmlio.ReadFile(filepath.Join("..", "..", "testdata", "2025-2.06-13x11-941577-blank.wxx"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	layers, err := xmlio.ReadFile(filepath.Join("..", "..", "testdata", "2025-2.06-13x11-941577-layers-beta.wxx"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	m.HexOrientation, m.GridOrientation = "ROWS", hexg.OddR
	m.HexWidth, m.HexHeight = 40, 46.18
	feature := layers.Features[0]
	feature.Location.X, feature.Location.Y = 1050, 1050
	feature.Label = nil
	m.Features = []*wxx.Feature_t{feature}
	m.Shapes = []*wxx.Shape_t{{
		Type: "Polygon", MapLayer: "Above Terrain", CreationType: "BASIC",
		IsWorld: true, HighestViewLevel: "WORLD", CurrentShapeViewLevel: "WORLD",
		StrokeColor: "1.0,1.0,1.0,1.0", DsColor: "null", InsColor: "null", StrokeType: "SIMPLE", Opacity: 1,
		Points: []*wxx.Point_t{{X: 1000, Y: 1000}, {X: 1100, Y: 1000}, {X: 1050, Y: 1100}},
	}}
	m.Notes = []*wxx.Note_t{{Location: &wxx.NoteLocation_t{ViewLevel: "WORLD", X: 1050, Y: 1050},
		Color: &wxx.RGBA_t{R: 1, G: 1, A: 1}, Title: "Hex 3,4"}}
	m.ExtraTerrain = &wxx.ExtraTerrain_t{MapLayers: []*wxx.ExtraTerrainLayer_t{{
		Name:    "Below All",
		Terrain: []*wxx.TerrainAndLocation_t{{Terrain: "Classic/Flat Beach", Elevation: 1, X: 750, Y: 225}},
	}}}
	path := filepath.Join(t.TempDir(), "rows.wxx")
	if err := xmlio.WriteFile(path, m, "2.06"); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return path
}

// TestResizeRows: on a ROWS map, -left 2 -top 2 moves everything by ROWS
// distances, 300 per column and 225 per row, so each element stays on its hex
// two columns right and two rows down (issue #80). The maintainer confirmed
// this in Worldographer: the cathedral on (3,4) lands on (5,6), at
// (1650,1500). The old COLUMNS distances put it at (1500,1650), which the app
// showed on (4,7). The hex size must stay ROWS-shaped; resize used to write the
// COLUMNS size, which skewed every hex.
func TestResizeRows(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	if code, stderr := resize(t, "-input", rowsMap(t), "-output", out, "-left", "2", "-top", "2"); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr)
	}
	m, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	const dx, dy = 600, 450
	if m.HexOrientation != "ROWS" || m.HexWidth != 40 || m.HexHeight != 46.18 {
		t.Errorf("output is %s with hexes %v x %v, want ROWS 40 x 46.18", m.HexOrientation, m.HexWidth, m.HexHeight)
	}
	if got := [2]float64{m.Features[0].Location.X, m.Features[0].Location.Y}; got != [2]float64{1650, 1500} {
		t.Errorf("feature at %v, want [1650 1500], the center of hex (5,6)", got)
	}
	if p := m.Shapes[0].Points[0]; p.X != 1000+dx || p.Y != 1000+dy {
		t.Errorf("shape point at (%v,%v), want (%v,%v)", p.X, p.Y, 1000+dx, 1000+dy)
	}
	if loc := m.Notes[0].Location; loc == nil || loc.X != 1050+dx || loc.Y != 1050+dy {
		t.Errorf("note at %+v, want (1650,1500)", loc)
	}
	tl := m.ExtraTerrain.MapLayers[0].Terrain
	if len(tl) != 1 || tl[0].X != 750+dx || tl[0].Y != 225+dy {
		t.Errorf("placement not moved to hex (4,3)'s corner (1350,675): %d placement(s)", len(tl))
	}
}

// TestResizeRowsParity: the stagger that must be preserved is per row in ROWS
// and per column in COLUMNS, so an odd number of top rows is refused on a ROWS
// map and allowed on a COLUMNS one, and the reverse for left columns (#80).
func TestResizeRowsParity(t *testing.T) {
	columns := filepath.Join("..", "..", "testdata", "2025-2.06-13x11-941577-blank.wxx")
	rows := rowsMap(t)
	for _, tc := range []struct {
		name, input, flag string
		wantOK            bool
	}{
		{"ROWS, odd top", rows, "-top", false},
		{"ROWS, odd left", rows, "-left", true},
		{"COLUMNS, odd left", columns, "-left", false},
		{"COLUMNS, odd top", columns, "-top", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, stderr := resize(t, "-input", tc.input, "-output", filepath.Join(t.TempDir(), "out.wxx"), tc.flag, "1")
			if ok := code == 0; ok != tc.wantOK {
				t.Errorf("exit %d, want success=%v; stderr:\n%s", code, tc.wantOK, stderr)
			}
			if !tc.wantOK && !strings.Contains(stderr, "must be even") {
				t.Errorf("stderr does not explain the refusal:\n%s", stderr)
			}
		})
	}
}

// TestResizeFullyPaintedMap: a map with no blank hex has no "Blank" entry,
// because Worldographer lists only terrains in use. Resize used to refuse it;
// now it adds the entry and fills the new hexes with it, leaving the painted
// ones alone (issue #81). Worldographer 2.08 opened a map whose table gained
// Blank this way, and a resize of it.
func TestResizeFullyPaintedMap(t *testing.T) {
	m, err := xmlio.ReadFile(filepath.Join("..", "..", "testdata", "2025-2.06-13x11-941577-blank.wxx"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	const farmland = "Classic/Flat Farmland"
	m.TerrainMap.Data = map[string]int{farmland: 0}
	m.TerrainMap.List = []*wxx.Terrain_t{{Index: 0, Label: farmland}}
	for _, column := range m.Tiles.Tiles {
		for _, tile := range column {
			tile.Terrain = 0
		}
	}
	src := filepath.Join(t.TempDir(), "painted.wxx")
	if err := xmlio.WriteFile(src, m, "2.06"); err != nil {
		t.Fatalf("write source: %v", err)
	}
	// A plain decode and encode must leave the table as it is: only an
	// operation that writes a Blank tile adds the entry.
	if back, err := xmlio.ReadFile(src); err != nil {
		t.Fatalf("read source: %v", err)
	} else if _, ok := back.TerrainMap.Data[wxx.BlankTerrain]; ok || len(back.TerrainMap.Data) != 1 {
		t.Fatalf("a round trip changed the table to %v; want only %s", back.TerrainMap.Data, farmland)
	}

	out := filepath.Join(t.TempDir(), "out.wxx")
	if code, stderr := resize(t, "-input", src, "-output", out, "-left", "2", "-top", "1"); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr)
	}
	got, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	blank, ok := got.TerrainMap.Data[wxx.BlankTerrain]
	if !ok {
		t.Fatalf("output table %v has no Blank entry", got.TerrainMap.Data)
	}
	if got.TerrainMap.Data[farmland] != 0 || blank != 1 {
		t.Errorf("table = %v, want farmland 0 and Blank 1", got.TerrainMap.Data)
	}
	counts := map[int]int{}
	for c, column := range got.Tiles.Tiles {
		for r, tile := range column {
			counts[tile.Terrain]++
			added := c < 2 || r < 1
			if added && tile.Terrain != blank {
				t.Fatalf("added hex (%d,%d) is terrain %d, want Blank (%d)", c, r, tile.Terrain, blank)
			} else if !added && tile.Terrain != 0 {
				t.Fatalf("painted hex (%d,%d) is terrain %d, want farmland (0)", c, r, tile.Terrain)
			}
		}
	}
	// 15 x 12 = 180 hexes, 143 of them painted.
	if counts[0] != 143 || counts[blank] != 37 {
		t.Errorf("terrain counts = %v, want 143 farmland and 37 Blank", counts)
	}
}
