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
	// The fixture is 5 x 3; removing 6 columns leaves -1.
	code, stderr := resize(t,
		"-input", filepath.Join("..", "..", "testdata", "2017-1.77-1.0-columns-blank.wxx"),
		"-output", out,
		"-left", "-2", "-right", "-4")
	if code == 0 {
		t.Fatalf("exit 0, want a failure; stderr:\n%s", stderr)
	}
	if strings.Contains(stderr, "panic:") {
		t.Fatalf("resize panicked:\n%s", stderr)
	}
	if !strings.Contains(stderr, "smaller than 2 x 2 (this resize gives -1 x 3)") {
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
		"-input", filepath.Join("..", "..", "testdata", "2017-1.77-1.0-columns-blank.wxx"),
		"-output", out,
		"-left", "-2", "-bottom", "-1")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr:\n%s", code, stderr)
	}
	m, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if m.Tiles.TilesWide != 3 || m.Tiles.TilesHigh != 2 {
		t.Errorf("output is %d x %d, want 3 x 2 (5 x 3 less 2 columns and 1 row)", m.Tiles.TilesWide, m.Tiles.TilesHigh)
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
		Points: []*wxx.Point_t{{X: 1000, Y: 700}, {X: 1100, Y: 700}, {X: 1050, Y: 800}},
	}}
	m.Notes = []*wxx.Note_t{{
		Key: "WORLD,1050.0,750.0", ViewLevel: "WORLD", X: 1050, Y: 750,
		Color: &wxx.RGBA_t{R: 1, G: 1, A: 1}, Title: "Here",
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
	}

	if len(m.Notes) != 1 {
		t.Fatalf("%d note(s), want 1", len(m.Notes))
	}
	n := m.Notes[0]
	if n.X != 1050+dx || n.Y != 750+dy {
		t.Errorf("note at (%v,%v), want (%v,%v)", n.X, n.Y, 1050+dx, 750+dy)
	}
	if want := "WORLD,1500.0,1050.0"; n.Key != want {
		t.Errorf("note key = %q, want %q: the key repeats the position and must move with it", n.Key, want)
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
	if len(m.Notes) != 1 || m.Notes[0].X != 1050-450 {
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

// TestParseNoteKey: a key in Worldographer's "<level>,<x>,<y>" form is read as
// the note's position; anything else is not.
func TestParseNoteKey(t *testing.T) {
	for _, tc := range []struct {
		key   string
		ok    bool
		level string
		x, y  float64
	}{
		{"WORLD,100.0,200.0", true, "WORLD", 100, 200},
		{"WORLD,2343.75,3112.5", true, "WORLD", 2343.75, 3112.5},
		{"custom", false, "", 0, 0},
		{"WORLD,x,200.0", false, "", 0, 0},
	} {
		level, x, y, ok := parseNoteKey(tc.key)
		if ok != tc.ok || level != tc.level || x != tc.x || y != tc.y {
			t.Errorf("parseNoteKey(%q) = %q, %v, %v, %v; want %q, %v, %v, %v", tc.key, level, x, y, ok, tc.level, tc.x, tc.y, tc.ok)
		}
	}
}

// TestResizeMovesKeyOnlyNote: a 2.08 note states its position only in its
// key, so it decodes with X and Y zero. Resizing must move it by the key and
// keep it, not treat (0,0) as off the map and delete it.
func TestResizeMovesKeyOnlyNote(t *testing.T) {
	m, err := xmlio.ReadFile(positionedMap(t))
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	m.Notes[0].X, m.Notes[0].Y = 0, 0 // as a 2.08 note decodes
	src := filepath.Join(t.TempDir(), "keyonly.wxx")
	if err := xmlio.WriteFile(src, m, "2.06"); err != nil {
		t.Fatalf("write source: %v", err)
	}
	out := filepath.Join(t.TempDir(), "out.wxx")
	// A crop, because that is where zero X/Y goes wrong: cropping 2 columns
	// takes them to -450, off the map, while the key's position (1050,750),
	// hex (4,2), moves to (600,750), hex (2,2), still on it.
	if code, stderr := resize(t, "-input", src, "-output", out, "-left", "-2"); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr)
	}
	got, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(got.Notes) != 1 {
		t.Fatalf("%d note(s), want 1: the note was deleted because its X/Y are zero", len(got.Notes))
	}
	if want := "WORLD,600.0,750.0"; got.Notes[0].Key != want {
		t.Errorf("note key = %q, want %q", got.Notes[0].Key, want)
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
