// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/maloquacious/hexg"
	"github.com/maloquacious/wxx"
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
// later is covered without anyone remembering to add it here. The one exception
// is classicFixture, which is kept to be refused (issue #103) and is held to that
// by TestDecodeRefusesAClassicMap.
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
		if name == filepath.Base(classicFixture) {
			continue
		}
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
// The invariant is on the model, not on the target: the refusal comes from
// validating the map before any codec runs.
func TestMalformedMapRefusedBeforeAnyBytes(t *testing.T) {
	for _, target := range []struct {
		app     string
		fixture string
	}{
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
