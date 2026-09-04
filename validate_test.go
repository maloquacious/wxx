// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/hexg"
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
	m.GridOrientation = hexg.OddQ
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
	rows.HexOrientation, rows.GridOrientation = "ROWS", hexg.OddR
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
			// The desync issue #20 is named for: two fields holding one fact,
			// disagreeing. COLUMNS with a rows coordinate convention is not a
			// grid that can be drawn.
			name:    "grid orientation contradicts hex orientation",
			break_:  func(m *Map_t) { m.GridOrientation = hexg.OddR },
			wantErr: ErrMismatchedGridOrientation,
			wantMsg: "odd-r",
		},
		{
			// The unset zero value is the same desync, reached by omission
			// rather than by contradiction.
			name:    "grid orientation unset",
			break_:  func(m *Map_t) { m.GridOrientation = hexg.UnknownQR },
			wantErr: ErrMismatchedGridOrientation,
			wantMsg: "unknown-qr",
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
			// the grid does not hold, and both codecs loop to the header.
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
