// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
	"github.com/maloquacious/wxx/xmlio"
)

// decodeValidFixture decodes one .wxx from testdata/ through the public
// pipeline and fails the test if it will not decode. Every test in this file
// starts from a real file, because the states being asserted are ones a CALLER
// produces from a map that was fine when it arrived.
func decodeValidFixture(t *testing.T, name string) *wxx.Map_t {
	t.Helper()
	m, err := decodeFile(t, filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return m
}

// TestEveryFixtureDecodesToAValidMap is the guard's other half: Map_t.Validate
// must accept everything the decoders produce. A validator the pipeline's own
// output fails is a validator that has to be worked around, and the workaround
// would be to stop calling it.
//
// It runs over every .wxx in testdata/ rather than a list, so a fixture added
// later is covered without anyone remembering to add it here.
func TestEveryFixtureDecodesToAValidMap(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "testdata", "*.wxx"))
	if err != nil {
		t.Fatalf("glob testdata: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatalf("glob testdata: no .wxx fixtures found")
	}
	for _, path := range fixtures {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			if err := decodeValidFixture(t, name).Validate(); err != nil {
				t.Errorf("%s: decoded map fails Validate(): %v", name, err)
			}
		})
	}
}

// TestMalformedMapRefusedBeforeAnyBytes asserts that a map whose fields
// contradict each other is refused by the public write path, and that the
// io.Writer is untouched when it is.
//
// Every case here PANICKED before issue #20 -- "invalid memory address or nil
// pointer dereference" for the nil substructures, "index out of range [5] with
// length 5" for the over-long header -- from inside a codec, part-way through
// building the document. The panic is what makes this worth a test rather than
// a doc note: a caller could not recover from it, and it named the encoder's
// line rather than the caller's mistake.
//
// Both targets are exercised because the two codecs differ in what they
// dereference: classic writes a hard-coded <mapkey> and ignores <informations>,
// so a nil MapKey crashed only v1_06. The invariant is on the model, not on the
// target, and this is where that shows.
func TestMalformedMapRefusedBeforeAnyBytes(t *testing.T) {
	for _, target := range []struct {
		app     string
		fixture string
	}{
		{"1.77", "2017-1.77-1.0-columns-blank.wxx"},
		{"2.06", "2025-2.06-13x11-941577-blank.wxx"},
	} {
		for _, tc := range []struct {
			name    string
			break_  func(*wxx.Map_t)
			wantErr error
		}{
			{"nil Tiles", func(m *wxx.Map_t) { m.Tiles = nil }, wxx.ErrIncompleteMap},
			{"nil TerrainMap", func(m *wxx.Map_t) { m.TerrainMap = nil }, wxx.ErrIncompleteMap},
			{"nil GridAndNumbering", func(m *wxx.Map_t) { m.GridAndNumbering = nil }, wxx.ErrIncompleteMap},
			{"nil MapKey", func(m *wxx.Map_t) { m.MapKey = nil }, wxx.ErrIncompleteMap},
			{"nil Informations", func(m *wxx.Map_t) { m.Informations = nil }, wxx.ErrIncompleteMap},
			{"nil Configuration", func(m *wxx.Map_t) { m.Configuration = nil }, wxx.ErrIncompleteMap},
			{"header wider than the grid", func(m *wxx.Map_t) { m.Tiles.TilesWide++ }, wxx.ErrInvalidTileGrid},
			{"header higher than the grid", func(m *wxx.Map_t) { m.Tiles.TilesHigh++ }, wxx.ErrInvalidTileGrid},
			{"nil tile", func(m *wxx.Map_t) { m.Tiles.Tiles[0][0] = nil }, wxx.ErrInvalidTileGrid},
			{"orientation desync", func(m *wxx.Map_t) { m.GridOrientation = hexg.OddR }, wxx.ErrMismatchedGridOrientation},
		} {
			t.Run(target.app+"/"+tc.name, func(t *testing.T) {
				m := decodeValidFixture(t, target.fixture)
				tc.break_(m)

				var buf bytes.Buffer
				err := xmlio.NewEncoder(target.app).Encode(&buf, m)
				if err == nil {
					t.Fatalf("Encode: want %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("Encode: err = %v, want errors.Is(err, %v)", err, tc.wantErr)
				}
				if buf.Len() != 0 {
					t.Errorf("Encode: wrote %d bytes to w, want 0 -- a refused encode must not produce a partial file", buf.Len())
				}

				// MarshalXML is the other public way in, and it must refuse the
				// same map: a check on only one of them leaves a path to the
				// panic it was added to remove.
				if _, err := xmlio.MarshalXML(m, target.app); !errors.Is(err, tc.wantErr) {
					t.Errorf("MarshalXML: err = %v, want errors.Is(err, %v)", err, tc.wantErr)
				}
			})
		}
	}
}

// TestClassicRowsRefusedUpFront is issue #20's live instance.
//
// The classic codec has no ROWS emit branch and classic is frozen, so it will
// not get one (internal/v0_77/COVERAGE.md). What changed is the report: the
// refusal used to be `assert(orientation != "ROWS")`, returned from encodeTiles
// with a partly-built document in the buffer, and it named a condition the
// codec expected rather than the problem the caller has.
//
// The test asserts all three properties of the new refusal -- it is matchable
// (ErrUnsupportedHexOrientation), it names the caller's field, and it is not an
// assert -- and then verifies the remedy the message offers, because a message
// that recommends a target nobody checked is exactly the unverified claim this
// repository keeps paying for. The ROWS fixture is written to 2.06 and read
// back: orientation, dimensions and every tile survive.
func TestClassicRowsRefusedUpFront(t *testing.T) {
	m := decodeValidFixture(t, "2017-1.77-1.0-rows-blank.wxx")
	if m.HexOrientation != "ROWS" {
		t.Fatalf("fixture: HexOrientation = %q, want ROWS", m.HexOrientation)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("fixture: Validate() = %v, want nil -- a ROWS map is a VALID map, the classic codec just cannot write one", err)
	}

	var buf bytes.Buffer
	err := xmlio.NewEncoder("1.77").Encode(&buf, m)
	if err == nil {
		t.Fatalf("Encode(1.77): want an error, got nil")
	}
	if !errors.Is(err, wxx.ErrUnsupportedHexOrientation) {
		t.Errorf("Encode(1.77): err = %v, want errors.Is(err, %v)", err, wxx.ErrUnsupportedHexOrientation)
	}
	if got := err.Error(); !strings.Contains(got, "hexOrientation") || !strings.Contains(got, `"ROWS"`) {
		t.Errorf("Encode(1.77): err = %q, want it to name the caller's field and its value", got)
	}
	if strings.Contains(err.Error(), "assert(") {
		t.Errorf("Encode(1.77): err = %q, still an assert -- the caller is owed a statement about their map", err.Error())
	}
	if buf.Len() != 0 {
		t.Errorf("Encode(1.77): wrote %d bytes to w, want 0", buf.Len())
	}

	// The remedy the error names, verified rather than asserted in prose.
	var w2025 bytes.Buffer
	if err := xmlio.NewEncoder("2.06").Encode(&w2025, m); err != nil {
		t.Fatalf("Encode(2.06): %v -- the error message recommends this target", err)
	}
	back, err := xmlio.NewDecoder().Decode(bytes.NewReader(w2025.Bytes()))
	if err != nil {
		t.Fatalf("re-decode the 2.06 encode: %v", err)
	}
	if back.HexOrientation != "ROWS" {
		t.Errorf("re-decode: HexOrientation = %q, want ROWS", back.HexOrientation)
	}
	if back.Tiles.TilesWide != m.Tiles.TilesWide || back.Tiles.TilesHigh != m.Tiles.TilesHigh {
		t.Fatalf("re-decode: grid is %dx%d, want %dx%d",
			back.Tiles.TilesWide, back.Tiles.TilesHigh, m.Tiles.TilesWide, m.Tiles.TilesHigh)
	}
	for x := 0; x < m.Tiles.TilesWide; x++ {
		for y := 0; y < m.Tiles.TilesHigh; y++ {
			in, out := m.Tiles.Tiles[x][y], back.Tiles.Tiles[x][y]
			if in.Terrain != out.Terrain || in.Elevation != out.Elevation ||
				in.IsIcy != out.IsIcy || in.IsGMOnly != out.IsGMOnly || in.Resources != out.Resources {
				t.Fatalf("re-decode: tile [%d][%d] changed: %+v -> %+v", x, y, *in, *out)
			}
		}
	}
}
