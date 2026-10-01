// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"strings"
	"testing"
)

// validMap builds the smallest Map_t that Validate accepts: the orientation
// stated in both of the fields that hold it, every substructure an encoder
// dereferences without checking, and a 2x3 grid that matches its own header.
//
// It is deliberately EMPTY apart from that. Validate is about the shape of a
// map and not about its contents, so a map whose terrain map has no terrain in
// it and whose tiles are all zero must pass -- if this helper had to be
// populated to validate, the check would have grown past what it claims to do.
func validMap() *Map_t {
	const tilesWide, tilesHigh = 2, 3

	m := &Map_t{}
	m.HexOrientation = "COLUMNS"
	m.GridAndNumbering = &GridAndNumbering_t{}
	m.TerrainMap = &TerrainMap_t{}
	m.MapKey = &MapKey_t{}
	m.Informations = &Informations_t{}
	m.Configuration = &Configuration_t{
		TextConfig:  &TextConfig_t{},
		ShapeConfig: &ShapeConfig_t{},
	}
	m.Tiles = &Tiles_t{TilesWide: tilesWide, TilesHigh: tilesHigh}
	for x := 0; x < tilesWide; x++ {
		column := make([]*Tile_t, tilesHigh)
		for y := 0; y < tilesHigh; y++ {
			column[y] = &Tile_t{}
		}
		m.Tiles.Tiles = append(m.Tiles.Tiles, column)
	}
	return m
}

// TestValidateAcceptsAWellFormedMap is the half of the guard that is easy to
// get wrong in the other direction: a check that rejects everything passes
// every rejection test and is still useless. Both orientations are asserted,
// because the two arms of the orientation switch are different code.
func TestValidateAcceptsAWellFormedMap(t *testing.T) {
	if err := validMap().Validate(); err != nil {
		t.Errorf("COLUMNS: Validate() = %v, want nil", err)
	}

	rows := validMap()
	rows.HexOrientation = "ROWS"
	if err := rows.Validate(); err != nil {
		t.Errorf("ROWS: Validate() = %v, want nil", err)
	}
}

// TestValidateRejects walks every state Validate exists to reject, one at a
// time, starting from a map that passes. Each case names the error constant a
// caller would match with errors.Is and a fragment of the message that must
// name the field, because "invalid tile grid" alone does not tell a caller
// which of their fields to look at.
func TestValidateRejects(t *testing.T) {
	for _, tc := range []struct {
		name    string
		break_  func(*Map_t)
		wantErr error
		wantMsg string
	}{
		{
			name:    "orientation unset",
			break_:  func(m *Map_t) { m.HexOrientation = "" },
			wantErr: ErrInvalidHexOrientation,
			wantMsg: `hexOrientation ""`,
		},
		{
			name:    "orientation not an orientation",
			break_:  func(m *Map_t) { m.HexOrientation = "HEXES" },
			wantErr: ErrInvalidHexOrientation,
			wantMsg: `hexOrientation "HEXES"`,
		},
		{
			name:    "nil GridAndNumbering",
			break_:  func(m *Map_t) { m.GridAndNumbering = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.GridAndNumbering",
		},
		{
			name:    "nil TerrainMap",
			break_:  func(m *Map_t) { m.TerrainMap = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.TerrainMap",
		},
		{
			name:    "nil Tiles",
			break_:  func(m *Map_t) { m.Tiles = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.Tiles",
		},
		{
			name:    "nil MapKey",
			break_:  func(m *Map_t) { m.MapKey = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.MapKey",
		},
		{
			name:    "nil Informations",
			break_:  func(m *Map_t) { m.Informations = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.Informations",
		},
		{
			name:    "nil Configuration",
			break_:  func(m *Map_t) { m.Configuration = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.Configuration",
		},
		{
			name:    "nil Configuration.TextConfig",
			break_:  func(m *Map_t) { m.Configuration.TextConfig = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.Configuration.TextConfig",
		},
		{
			name:    "nil Configuration.ShapeConfig",
			break_:  func(m *Map_t) { m.Configuration.ShapeConfig = nil },
			wantErr: ErrIncompleteMap,
			wantMsg: "Map_t.Configuration.ShapeConfig",
		},
		{
			// The grid the encoders panicked on: the header promises a column
			// the grid does not hold, and the encoder loops to the header.
			name:    "header claims more columns than the grid holds",
			break_:  func(m *Map_t) { m.Tiles.TilesWide++ },
			wantErr: ErrInvalidTileGrid,
			wantMsg: "@tilesWide is 3 but the grid holds 2 columns",
		},
		{
			name:    "header claims more rows than a column holds",
			break_:  func(m *Map_t) { m.Tiles.TilesHigh++ },
			wantErr: ErrInvalidTileGrid,
			wantMsg: "@tilesHigh is 4 but column 0 holds 3 tiles",
		},
		{
			name:    "one short column",
			break_:  func(m *Map_t) { m.Tiles.Tiles[1] = m.Tiles.Tiles[1][:1] },
			wantErr: ErrInvalidTileGrid,
			wantMsg: "column 1 holds 1 tiles",
		},
		{
			name:    "negative tilesWide",
			break_:  func(m *Map_t) { m.Tiles.TilesWide = -1 },
			wantErr: ErrInvalidTileGrid,
			wantMsg: "negative",
		},
		{
			name:    "nil tile",
			break_:  func(m *Map_t) { m.Tiles.Tiles[1][2] = nil },
			wantErr: ErrInvalidTileGrid,
			wantMsg: "Tiles_t.Tiles[1][2]",
		},
		{
			// 128 and above stop Worldographer opening the file (app check
			// at 150, issue #122).
			name:    "resource the app cannot read",
			break_:  func(m *Map_t) { m.Tiles.Tiles[0][0].Resources.Brick = 150 },
			wantErr: ErrInvalidTileResource,
			wantMsg: "Tiles_t.Tiles[0][0].Resources.Brick): 150: want 0..100",
		},
		{
			// 101..127 open, and the app saves them back as 100 (app check at
			// 127, issue #122).
			name:    "resource the app clamps",
			break_:  func(m *Map_t) { m.Tiles.Tiles[1][2].Resources.Rock = 101 },
			wantErr: ErrInvalidTileResource,
			wantMsg: "Tiles_t.Tiles[1][2].Resources.Rock): 101: want 0..100",
		},
		{
			name:    "negative resource",
			break_:  func(m *Map_t) { m.Tiles.Tiles[0][1].Resources.Animal = -1 },
			wantErr: ErrInvalidTileResource,
			wantMsg: "Tiles_t.Tiles[0][1].Resources.Animal): -1: want 0..100",
		},
		{
			// The first is named and the rest counted, not listed per tile.
			name: "several resources out of range",
			break_: func(m *Map_t) {
				m.Tiles.Tiles[0][0].Resources.Gems = 200
				m.Tiles.Tiles[1][0].Resources.Lumber = 200
				m.Tiles.Tiles[1][1].Resources.Metals = 200
			},
			wantErr: ErrInvalidTileResource,
			wantMsg: "Resources.Gems): 200: want 0..100 (and 2 more out-of-range resources)",
		},
		{
			// The app will not open a file with 150 in an <extraTerrain>
			// resource either (app check, issue #124).
			name: "extraTerrain resource the app cannot read",
			break_: func(m *Map_t) {
				m.ExtraTerrain = &ExtraTerrain_t{MapLayers: []*ExtraTerrainLayer_t{
					{Name: "Below All", Terrain: []*TerrainAndLocation_t{{Resources: Resources_t{Brick: 150}}}},
				}}
			},
			wantErr: ErrInvalidExtraTerrainResource,
			wantMsg: `mapLayer[@name="Below All"]/terrainAndLocation[0]/@resources (ExtraTerrain_t.MapLayers[0].Terrain[0].Resources.Brick): 150: want 0..100`,
		},
		{
			name: "negative extraTerrain resource",
			break_: func(m *Map_t) {
				m.ExtraTerrain = &ExtraTerrain_t{MapLayers: []*ExtraTerrainLayer_t{
					{Name: "Below All", Terrain: []*TerrainAndLocation_t{{Resources: Resources_t{Rock: -1}}}},
				}}
			},
			wantErr: ErrInvalidExtraTerrainResource,
			wantMsg: "Resources.Rock): -1: want 0..100",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := validMap()
			tc.break_(m)

			err := m.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want %v", tc.wantErr)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Validate() = %v, want errors.Is(err, %v)", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("Validate() = %q, want a message containing %q", err.Error(), tc.wantMsg)
			}
		})
	}
}

// TestValidateAcceptsResourceBounds asserts both ends of 0..100 pass on every
// resource field (issue #122). An off-by-one in the range check would refuse a
// value the app reads as written.
func TestValidateAcceptsResourceBounds(t *testing.T) {
	for _, v := range []int{0, 100} {
		m := validMap()
		m.Tiles.Tiles[1][2].Resources = Resources_t{Animal: v, Brick: v, Crops: v, Gems: v, Lumber: v, Metals: v, Rock: v}
		if err := m.Validate(); err != nil {
			t.Errorf("every resource %d: Validate() = %v, want nil", v, err)
		}
	}
}

// TestValidateNilMap asserts the nil receiver is reported rather than panicking.
// A method that panics on the state it exists to detect would be the same
// failure as the one issue #20 is about, moved one call deeper.
func TestValidateNilMap(t *testing.T) {
	var m *Map_t
	err := m.Validate()
	if !errors.Is(err, ErrNilMap) {
		t.Errorf("(*Map_t)(nil).Validate() = %v, want %v", err, ErrNilMap)
	}
}

// TestValidateReportsEveryProblem asserts the joined result holds ALL of the
// problems, not the first. A caller fixing a half-built map one error per
// compile-run is doing the work this method exists to do for them.
func TestValidateReportsEveryProblem(t *testing.T) {
	m := validMap()
	m.HexOrientation = "HEXES"
	m.TerrainMap = nil
	m.Tiles.TilesWide++

	err := m.Validate()
	if err == nil {
		t.Fatalf("Validate() = nil, want three problems")
	}
	for _, want := range []error{ErrInvalidHexOrientation, ErrIncompleteMap, ErrInvalidTileGrid} {
		if !errors.Is(err, want) {
			t.Errorf("Validate() = %v, want errors.Is(err, %v)", err, want)
		}
	}
}
