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

// terrainNames returns each tile's terrain NAME, [col][row], which is what a
// round trip must preserve: indices are renumbered, names are not.
func terrainNames(t *testing.T, m *wxx.Map_t) [][]string {
	t.Helper()
	byIndex := map[int]string{}
	for name, index := range m.TerrainMap.Data {
		byIndex[index] = name
	}
	var out [][]string
	for _, column := range m.Tiles.Tiles {
		var names []string
		for _, tile := range column {
			name, ok := byIndex[tile.Terrain]
			if !ok {
				t.Fatalf("a tile uses terrain index %d, which the table %v does not list", tile.Terrain, m.TerrainMap.Data)
			}
			names = append(names, name)
		}
		out = append(out, names)
	}
	return out
}

// TestTerrainMapWithGaps: a table whose indices have a gap -- a map built or
// edited in code; Worldographer always writes 0..n-1 -- is written 0..n-1 with
// every tile renumbered to match, so each tile keeps its terrain (issue #87).
// Before #87 the table was renumbered by position and the tiles were not, so
// every tile past the gap pointed at the wrong terrain, or at none.
func TestTerrainMapWithGaps(t *testing.T) {
	for _, tc := range []struct{ fixture, app string }{
		{"../testdata/2025-2.06-13x11-941577-blank.wxx", "2.06"},
	} {
		t.Run(tc.app, func(t *testing.T) {
			m, err := xmlio.ReadFile(tc.fixture)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			const farmland, sea = "Classic/Flat Farmland", "Classic/Water Sea"
			blank := m.TerrainMap.Data[wxx.BlankTerrain]
			m.TerrainMap.Data = map[string]int{wxx.BlankTerrain: blank, farmland: 5, sea: 9}
			for c, column := range m.Tiles.Tiles {
				for r, tile := range column {
					switch {
					case c == 0:
						tile.Terrain = 5
					case r == 0:
						tile.Terrain = 9
					default:
						tile.Terrain = blank
					}
				}
			}
			want := terrainNames(t, m)

			var ed xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(tc.app, xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode: %v", err)
			}
			if !strings.Contains(string(ed.Utf8Encoded), "<terrainmap>Blank\t0\tClassic/Flat Farmland\t1\tClassic/Water Sea\t2</terrainmap>") {
				t.Errorf("table not written 0..n-1 in index order")
			}

			back, err := xmlio.NewDecoder().Decode(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("re-decode: %v", err)
			}
			got := terrainNames(t, back)
			for c := range want {
				for r := range want[c] {
					if got[c][r] != want[c][r] {
						t.Fatalf("hex (%d,%d) is %q after the round trip, want %q", c, r, got[c][r], want[c][r])
					}
				}
			}
		})
	}
}

// TestTerrainMapRefusals: two terrains sharing an index, and a tile whose
// index the table does not list, have no correct encoding, so each is refused
// before anything is written, naming the problem (issue #87).
func TestTerrainMapRefusals(t *testing.T) {
	for _, tc := range []struct {
		name    string
		break_  func(*wxx.Map_t)
		wantErr error
		wantMsg string
	}{
		{"shared index", func(m *wxx.Map_t) { m.TerrainMap.Data["Classic/Flat Farmland"] = m.TerrainMap.Data[wxx.BlankTerrain] },
			wxx.ErrInvalidTerrainMap, "both have index"},
		{"unlisted index", func(m *wxx.Map_t) {
			last := m.Tiles.Tiles[len(m.Tiles.Tiles)-1]
			last[len(last)-1].Terrain = 42
		},
			wxx.ErrInvalidTileGrid, "uses terrain index 42"},
	} {
		for _, app := range []string{"2.06"} {
			t.Run(tc.name+" "+app, func(t *testing.T) {
				fixture := map[string]string{"2.06": "../testdata/2025-2.06-13x11-941577-blank.wxx"}[app]
				m, err := xmlio.ReadFile(fixture)
				if err != nil {
					t.Fatalf("read: %v", err)
				}
				tc.break_(m)
				var buf bytes.Buffer
				err = xmlio.NewEncoder(app).Encode(&buf, m)
				if !errors.Is(err, tc.wantErr) || !strings.Contains(err.Error(), tc.wantMsg) {
					t.Fatalf("encode: err = %v, want %v naming %q", err, tc.wantErr, tc.wantMsg)
				}
				if buf.Len() != 0 {
					t.Errorf("wrote %d bytes, want 0", buf.Len())
				}
			})
		}
	}
}
