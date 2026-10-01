// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// blankFixture is the blank fixture the application version app wrote, made
// by the "Blank" recipe in testdata/README.md.
func blankFixture(app string) string {
	return fmt.Sprintf("../testdata/2025-%s-13x11-941577-blank.wxx", app)
}

// TestNewMapMatchesBlankFixture is issue #136's main test: a 13 x 11 map built
// with NewMap and written as an application version is the blank fixture that
// version wrote, written the same way.
//
// Both sides go through the encoder, so the comparison is of what wxx writes,
// not of the encoder's whitespace against the app's. The documented
// differences are set on the fixture first:
//
//   - <informations>: the app fills it with lore generated from the random
//     seed, and NewMap leaves it empty.
//   - showGrid and showGridNumbers: UI settings, saved as the grid stood when
//     the fixture was saved (hidden in the 2.06 one), and NewMap always shows
//     the grid.
func TestNewMapMatchesBlankFixture(t *testing.T) {
	for _, app := range []string{"2.06", "2.07", "2.08"} {
		t.Run(app, func(t *testing.T) {
			fixture, err := xmlio.ReadFile(blankFixture(app))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			fixture.Informations = &wxx.Informations_t{}
			fixture.ShowGrid, fixture.ShowGridNumbers = true, true
			want, err := xmlio.MarshalXML(fixture, app)
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}

			m, err := xmlio.NewMap(13, 11, xmlio.WithApp(app))
			if err != nil {
				t.Fatalf("NewMap: %v", err)
			}
			if err := m.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			got, err := xmlio.MarshalXML(m, app)
			if err != nil {
				t.Fatalf("marshal NewMap: %v", err)
			}
			if !bytes.Equal(got, want) {
				reportFirstDifference(t, got, want)
			}
		})
	}
}

// TestNewMapMatchesFullSizeBlankFixture is TestNewMapMatchesBlankFixture at
// 1920 x 1080, against the blank map 2.08 wrote at that size: NewMap's
// defaults hold beyond 13 x 11. It takes about two seconds, so -short skips it.
func TestNewMapMatchesFullSizeBlankFixture(t *testing.T) {
	if testing.Short() {
		t.Skip("1920 x 1080: skipped with -short")
	}
	const app = "2.08"
	fixture, err := xmlio.ReadFile("../testdata/2025-2.08-1920x1080-941577-blank.wxx")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	fixture.Informations = &wxx.Informations_t{}
	fixture.ShowGrid, fixture.ShowGridNumbers = true, true
	want, err := xmlio.MarshalXML(fixture, app)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	m, err := xmlio.NewMap(1920, 1080, xmlio.WithApp(app))
	if err != nil {
		t.Fatalf("NewMap: %v", err)
	}
	got, err := xmlio.MarshalXML(m, app)
	if err != nil {
		t.Fatalf("marshal NewMap: %v", err)
	}
	if !bytes.Equal(got, want) {
		reportFirstDifference(t, got, want)
	}
}

// reportFirstDifference fails t, naming the first line where got and want part.
func reportFirstDifference(t *testing.T, got, want []byte) {
	t.Helper()
	g, w := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		var gl, wl string
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			t.Fatalf("line %d differs:\n got: %s\nwant: %s", i+1, gl, wl)
		}
	}
	t.Fatalf("output differs from the fixture")
}

// TestNewMapTiles pins the grid in both orientations against the decoder: each
// tile's Coords, Column and Row are what decoding the 13 x 11 fixture of that
// orientation gives (#85, #130), every tile is Blank, and Blank is the whole
// terrain table, at index 0. The ROWS hex size is the rows fixture's.
func TestNewMapTiles(t *testing.T) {
	for _, tc := range []struct {
		orientation string
		fixture     string
	}{
		{"COLUMNS", "../testdata/2025-2.08-13x11-941577-blank.wxx"},
		{"ROWS", "../testdata/2025-2.08-13x11-941577-rows.wxx"},
	} {
		t.Run(tc.orientation, func(t *testing.T) {
			fixture, err := xmlio.ReadFile(tc.fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			m, err := xmlio.NewMap(13, 11, xmlio.WithHexOrientation(tc.orientation))
			if err != nil {
				t.Fatalf("NewMap: %v", err)
			}
			if err := m.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if m.HexOrientation != tc.orientation {
				t.Errorf("HexOrientation = %q, want %q", m.HexOrientation, tc.orientation)
			}
			if m.HexWidth != fixture.HexWidth || m.HexHeight != fixture.HexHeight {
				t.Errorf("hex size = %v x %v, want the fixture's %v x %v", m.HexWidth, m.HexHeight, fixture.HexWidth, fixture.HexHeight)
			}
			if m.RowsHigh != fixture.RowsHigh || m.ColumnsWide != fixture.ColumnsWide {
				t.Errorf("RowsHigh, ColumnsWide = %d, %d, want the decoder's %d, %d", m.RowsHigh, m.ColumnsWide, fixture.RowsHigh, fixture.ColumnsWide)
			}
			if m.Tiles.TilesWide != 13 || m.Tiles.TilesHigh != 11 {
				t.Errorf("tiles = %d x %d, want 13 x 11", m.Tiles.TilesWide, m.Tiles.TilesHigh)
			}
			if len(m.TerrainMap.Data) != 1 || m.TerrainMap.Data[wxx.BlankTerrain] != 0 {
				t.Errorf("terrain table = %v, want Blank at 0 alone", m.TerrainMap.Data)
			}
			for x, column := range m.Tiles.Tiles {
				for y, tile := range column {
					want := fixture.Tiles.Tiles[x][y]
					if tile.Coords != want.Coords || tile.Column != want.Column || tile.Row != want.Row {
						t.Fatalf("tile [%d][%d]: coords %v col %d row %d, want the decoder's %v col %d row %d",
							x, y, tile.Coords, tile.Column, tile.Row, want.Coords, want.Column, want.Row)
					}
					if tile.Terrain != 0 {
						t.Fatalf("tile [%d][%d]: terrain %d, want Blank (0)", x, y, tile.Terrain)
					}
				}
			}
		})
	}
}

// TestNewMapCurrentApp: with no WithApp, NewMap uses CurrentApp, which is a
// registered version and the newest one.
func TestNewMapCurrentApp(t *testing.T) {
	current := xmlio.CurrentApp()
	if current != "2.08" {
		// 2.08 is the newest registered version today (#73). Registering a
		// newer one moves CurrentApp, and this pin with it.
		t.Errorf("CurrentApp() = %q, want %q", current, "2.08")
	}
	m, err := xmlio.NewMap(13, 11)
	if err != nil {
		t.Fatalf("NewMap: %v", err)
	}
	if got := m.MetaData.Version.App.Raw; got != current {
		t.Errorf("MetaData.Version.App = %q, want CurrentApp() %q", got, current)
	}
	if _, err := xmlio.MarshalXML(m, current); err != nil {
		t.Errorf("MarshalXML(NewMap(), CurrentApp()): %v", err)
	}
}

// TestNewMapRefuses: an unregistered version is the error MarshalXML gives for
// it, and a bad orientation or size is refused rather than built.
func TestNewMapRefuses(t *testing.T) {
	for _, tc := range []struct {
		name          string
		columns, rows int
		opts          []xmlio.NewMapOption
		want          error
	}{
		{"unregistered version", 13, 11, []xmlio.NewMapOption{xmlio.WithApp("2.05")}, wxx.ErrUnsupportedMapVersion},
		{"unpadded version", 13, 11, []xmlio.NewMapOption{xmlio.WithApp("2.8")}, wxx.ErrUnsupportedMapVersion},
		{"empty version", 13, 11, []xmlio.NewMapOption{xmlio.WithApp("")}, wxx.ErrUnsupportedMapVersion},
		{"orientation", 13, 11, []xmlio.NewMapOption{xmlio.WithHexOrientation("columns")}, wxx.ErrInvalidHexOrientation},
		{"too narrow", 1, 11, nil, wxx.ErrInvalidTileGrid},
		{"too short", 13, 1, nil, wxx.ErrInvalidTileGrid},
		{"negative", -2, 11, nil, wxx.ErrInvalidTileGrid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := xmlio.NewMap(tc.columns, tc.rows, tc.opts...)
			if !errors.Is(err, tc.want) {
				t.Fatalf("NewMap error = %v, want %v", err, tc.want)
			}
			if m != nil {
				t.Errorf("NewMap returned a map with its error")
			}
		})
	}

	// The same error as MarshalXML's for the same version.
	m, err := xmlio.NewMap(13, 11)
	if err != nil {
		t.Fatalf("NewMap: %v", err)
	}
	_, marshalErr := xmlio.MarshalXML(m, "2.05")
	_, newErr := xmlio.NewMap(13, 11, xmlio.WithApp("2.05"))
	if marshalErr == nil || newErr == nil || marshalErr.Error() != newErr.Error() {
		t.Errorf("NewMap error %q, want MarshalXML's %q", newErr, marshalErr)
	}
}

// TestNewMapPanamaSize builds the size of the Panama map with its border, 115 x 230
// COLUMNS, and checks it survives a write and a read with every hex Blank.
// Whether the app opens it is an app check, not this test.
func TestNewMapPanamaSize(t *testing.T) {
	const columns, rows = 115, 230
	m, err := xmlio.NewMap(columns, rows)
	if err != nil {
		t.Fatalf("NewMap: %v", err)
	}
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(xmlio.CurrentApp()).Encode(&buf, m); err != nil {
		t.Fatalf("encode: %v", err)
	}
	back, err := xmlio.NewDecoder().Decode(&buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if back.Tiles.TilesWide != columns || back.Tiles.TilesHigh != rows {
		t.Fatalf("read back %d x %d, want %d x %d", back.Tiles.TilesWide, back.Tiles.TilesHigh, columns, rows)
	}
	blank, ok := back.TerrainMap.Data[wxx.BlankTerrain]
	if !ok || len(back.TerrainMap.Data) != 1 {
		t.Fatalf("terrain table = %v, want Blank alone", back.TerrainMap.Data)
	}
	for x, column := range back.Tiles.Tiles {
		for y, tile := range column {
			if tile.Terrain != blank {
				t.Fatalf("tile [%d][%d]: terrain %d, want Blank (%d)", x, y, tile.Terrain, blank)
			}
		}
	}
}
