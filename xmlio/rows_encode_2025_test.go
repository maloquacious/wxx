// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
	"github.com/maloquacious/wxx/xmlio/h2025v1"
)

// newRowsMap builds a minimal, fully-populated W2025 Map_t with ROWS
// orientation and a deliberately asymmetric 2x3 tile grid (TilesWide=2,
// TilesHigh=3). Every cell carries distinct Terrain/Elevation/IsIcy/IsGMOnly/
// Resources derived from its (x,y) position so that a transpose bug in the
// encoder surfaces as a per-cell mismatch on round-trip.
func newRowsMap() *wxx.Map_t {
	const tilesWide, tilesHigh = 2, 3

	m := &wxx.Map_t{}
	m.MetaData.AppVersion = wxx.Version()
	// The schema selects the encode codec: schema "1.06" routes through h2025v1.
	// The components are spelled out rather than rendered, because Raw is the
	// authority and "2.06" is not "2.6".
	schema := wxx.Dotted{Raw: "1.06", Major: 1, Minor: 6}
	m.MetaData.Version = wxx.Version_t{
		App:    wxx.Dotted{Raw: "2.06", Major: 2, Minor: 6},
		Schema: &schema,
	}
	m.MetaData.Worldographer.Release = "2025"
	m.MetaData.Worldographer.Version = "2.06"
	m.MetaData.Worldographer.Schema = "1.06"

	// <map> attributes the encoder emits and the decoder re-reads.
	m.Type = "WORLD"
	m.Release = "2025"
	m.Version = "2.06"
	m.Schema = "1.06"
	m.MapProjection = wxx.FLAT
	m.HexOrientation = "ROWS"
	m.GridOrientation = hexg.OddR

	// Required non-nil substructures so the encoder does not nil-deref.
	m.GridAndNumbering = &wxx.GridAndNumbering_t{}
	m.MapKey = &wxx.MapKey_t{}
	m.Informations = &wxx.Informations_t{}
	m.Configuration = &wxx.Configuration_t{
		TextConfig:  &wxx.TextConfig_t{},
		ShapeConfig: &wxx.ShapeConfig_t{},
	}
	m.TerrainMap = &wxx.TerrainMap_t{
		Data: map[string]int{"Blank": 0, "Water": 1},
		List: []*wxx.Terrain_t{{Index: 0, Label: "Blank"}, {Index: 1, Label: "Water"}},
	}

	m.Tiles = &wxx.Tiles_t{
		ViewLevel: "WORLD",
		TilesWide: tilesWide,
		TilesHigh: tilesHigh,
	}
	// Row orientation: RowsHigh = TilesWide, ColumnsWide = TilesHigh (mirrors decode).
	m.RowsHigh = tilesWide
	m.ColumnsWide = tilesHigh
	for x := 0; x < tilesWide; x++ {
		col := make([]*wxx.Tile_t, tilesHigh)
		for y := 0; y < tilesHigh; y++ {
			t := &wxx.Tile_t{
				Row:       x,
				Column:    y,
				Coords:    hexg.NewOddRCoord(y, x).ToCube(),
				Terrain:   10*x + y, // distinct, position-sensitive
				Elevation: float64(100*x + y),
				IsIcy:     (x+y)%2 == 0,
				IsGMOnly:  x == 1,
				Resources: wxx.Resources_t{Animal: 10*x + y},
			}
			// Give one cell uncompressed resources to exercise that encode path.
			if x == 1 && y == 2 {
				t.Resources.Brick = 3
				t.Resources.Crops = 4
			}
			col[y] = t
		}
		m.Tiles.Tiles = append(m.Tiles.Tiles, col)
	}
	return m
}

// TestW2025RowsRoundTrip drives the h2025v1 codec over a ROWS map: encode the
// in-memory ROWS map, decode the bytes back, and assert the orientation and
// every tile's data round-trips to the SAME grid position. Before the ROWS
// encode branch lands, h2025v1.Encode returns an error for ROWS and this fails.
func TestW2025RowsRoundTrip(t *testing.T) {
	m1 := newRowsMap()

	xmlBytes, err := h2025v1.Encode(m1)
	if err != nil {
		t.Fatalf("h2025v1.Encode(ROWS): %v", err)
	}

	m2, err := h2025v1.Decode(xmlBytes)
	if err != nil {
		t.Fatalf("h2025v1.Decode(re-encoded ROWS): %v\n---encoded xml---\n%s", err, head(xmlBytes, 1200))
	}

	// Orientation must round-trip.
	if got, want := m2.HexOrientation, "ROWS"; got != want {
		t.Errorf("HexOrientation = %q, want %q", got, want)
	}
	if got, want := m2.GridOrientation, hexg.OddR; got != want {
		t.Errorf("GridOrientation = %v, want %v", got, want)
	}

	// Grid dimensions must round-trip.
	if got, want := m2.Tiles.TilesWide, m1.Tiles.TilesWide; got != want {
		t.Fatalf("TilesWide = %d, want %d", got, want)
	}
	if got, want := m2.Tiles.TilesHigh, m1.Tiles.TilesHigh; got != want {
		t.Fatalf("TilesHigh = %d, want %d", got, want)
	}

	// Every cell's data must land at the SAME [x][y] position; a transpose bug
	// would swap values between cells and trip these checks.
	for x := 0; x < m1.Tiles.TilesWide; x++ {
		for y := 0; y < m1.Tiles.TilesHigh; y++ {
			a, b := m1.Tiles.Tiles[x][y], m2.Tiles.Tiles[x][y]
			if a.Terrain != b.Terrain {
				t.Errorf("Tiles[%d][%d].Terrain = %d, want %d", x, y, b.Terrain, a.Terrain)
			}
			if a.Elevation != b.Elevation {
				t.Errorf("Tiles[%d][%d].Elevation = %v, want %v", x, y, b.Elevation, a.Elevation)
			}
			if a.IsIcy != b.IsIcy {
				t.Errorf("Tiles[%d][%d].IsIcy = %v, want %v", x, y, b.IsIcy, a.IsIcy)
			}
			if a.IsGMOnly != b.IsGMOnly {
				t.Errorf("Tiles[%d][%d].IsGMOnly = %v, want %v", x, y, b.IsGMOnly, a.IsGMOnly)
			}
			if a.Resources != b.Resources {
				t.Errorf("Tiles[%d][%d].Resources = %+v, want %+v", x, y, b.Resources, a.Resources)
			}
		}
	}
}
