// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"fmt"
)

// Validate reports every way m is internally inconsistent (issue #20).
//
// Map_t is a fully-exported struct with no constructor: a caller assembles one
// field by field, and nothing until now told them when they had assembled
// something a Worldographer file cannot be. No decoder produces the states
// checked here -- a decoder that cannot make sense of a file errors instead --
// so every one of them is a map a CALLER built, and the write path met them by
// panicking. All eight nil substructures below were verified to crash at least
// one codec, and a grid one column shorter than its own @tilesWide header
// crashed both -- "index out of range [5] with length 5" through classic,
// "[13] with length 13" through W2025. A panic is not a
// diagnosis: it names the encoder's line, not the caller's mistake, and it is
// not recoverable by a caller who is encoding a map they were handed.
//
// It returns ALL of the problems, joined, rather than the first: a
// half-assembled map usually has several, and a caller fixing them one error at
// a time is being made to do the work this method exists to do. Each is wrapped
// in a constant from errors.go, so errors.Is answers "which KIND of malformed"
// while the message says which field.
//
// WHAT IT DOES NOT CHECK, deliberately:
//
//   - MetaData.Version. It can still hold a pair no release ever stated -- the
//     classic application version "1.77" alongside the W2025 schema "1.06" --
//     and that is issue #20's other half, left open on purpose. Checking it
//     means asking what is registered, the registry is xmlio's, and xmlio
//     cannot be imported here (import cycle). Discovery is #46's to design; a
//     hard-coded list of releases here would be a second registry to drift.
//     Nothing reads MetaData.Version on the encode path today (issue #45
//     deleted the reader), so the pair is unread rather than obeyed.
//
//   - RowsHigh and ColumnsWide. They are decode-side labels derived from the
//     orientation and the grid dimensions, and no encoder writes them -- neither
//     schema has an attribute for them. Requiring a caller to set a field
//     nothing emits would reject maps that encode perfectly.
//
//   - Whether a target can express what m carries. That is a question about the
//     target, not about m, and it is answered by the encoder (see xmlio's
//     downgradeLoss).
//
// A nil receiver is a problem in itself and is reported as one rather than
// panicking, because "the map I was given is nil" is exactly the sort of caller
// mistake this method exists to name.
func (m *Map_t) Validate() error {
	if m == nil {
		return ErrNilMap
	}

	var problems []error

	// The orientation, and the second copy of it.
	//
	// HexOrientation is the string the file states and the one both encoders
	// switch on; GridOrientation is the hexg coordinate convention the decoders
	// set from it in the same switch. Two fields holding one fact can disagree,
	// and a map that says COLUMNS in one and odd-r in the other describes no
	// hex grid that exists. The unset zero value (UnknownQR) is caught by the
	// same check, which is intended: a caller who set only the string has not
	// finished building the map.
	switch m.HexOrientation {
	case "COLUMNS":
		if !m.GridOrientation.IsColumns() {
			problems = append(problems, errors.Join(ErrMismatchedGridOrientation,
				fmt.Errorf("hexOrientation %q with gridOrientation %s: want even-q or odd-q", m.HexOrientation, m.GridOrientation)))
		}
	case "ROWS":
		if !m.GridOrientation.IsRows() {
			problems = append(problems, errors.Join(ErrMismatchedGridOrientation,
				fmt.Errorf("hexOrientation %q with gridOrientation %s: want even-r or odd-r", m.HexOrientation, m.GridOrientation)))
		}
	default:
		problems = append(problems, errors.Join(ErrInvalidHexOrientation,
			fmt.Errorf("hexOrientation %q: want \"COLUMNS\" or \"ROWS\"", m.HexOrientation)))
	}

	// Substructures an encoder dereferences without checking.
	//
	// Every entry was verified to panic a codec when nil, and the two that
	// panic only ONE of them are still required here: MapKey and Informations
	// crash v1_06 while classic writes a hard-coded <mapkey> and ignores
	// <informations> entirely. The invariant is a property of the model, not of
	// a target -- Map_t is the superset of the supported schemas (ADR 0004
	// Decision 6) -- so a map missing one is incomplete whoever is about to
	// write it, and making the rule depend on the target would mean a map that
	// validates for classic and crashes for W2025.
	//
	// Presence is all that is asked. An empty &GridAndNumbering_t{} passes, as
	// it must: what a caller puts in these is map CONTENT and none of this
	// method's business.
	for _, req := range []struct {
		path    string // where it lives on disk
		field   string // where it lives in the model
		present bool
	}{
		{"map/gridandnumbering", "Map_t.GridAndNumbering", m.GridAndNumbering != nil},
		{"map/terrainmap", "Map_t.TerrainMap", m.TerrainMap != nil},
		{"map/tiles", "Map_t.Tiles", m.Tiles != nil},
		{"map/mapkey", "Map_t.MapKey", m.MapKey != nil},
		{"map/informations", "Map_t.Informations", m.Informations != nil},
		{"map/configuration", "Map_t.Configuration", m.Configuration != nil},
	} {
		if !req.present {
			problems = append(problems, errors.Join(ErrIncompleteMap,
				fmt.Errorf("%s (%s): nil", req.path, req.field)))
		}
	}
	if m.Configuration != nil {
		if m.Configuration.TextConfig == nil {
			problems = append(problems, errors.Join(ErrIncompleteMap,
				fmt.Errorf("map/configuration/text-config (Map_t.Configuration.TextConfig): nil")))
		}
		if m.Configuration.ShapeConfig == nil {
			problems = append(problems, errors.Join(ErrIncompleteMap,
				fmt.Errorf("map/configuration/shape-config (Map_t.Configuration.ShapeConfig): nil")))
		}
	}

	problems = append(problems, m.Tiles.validate()...)

	// Join drops nils and returns nil for an empty slice, so a valid map returns
	// nil without a length check here.
	return errors.Join(problems...)
}

// validate reports the ways the tile grid contradicts its own header.
//
// The header is <tiles tilesWide= tilesHigh=> and the grid is Tiles[col][row];
// both encoders loop to the HEADER's dimensions and index the grid, so a header
// claiming more than the grid holds is an index-out-of-range panic mid-write --
// after part of the document has already been emitted. That is the state this
// rejects, and it is why the check is on the pair rather than on either alone:
// neither the header nor the grid is wrong by itself, they are wrong together.
//
// A nil *Tiles_t reports nothing here. It is already reported as an incomplete
// map by the caller, and saying it twice would only pad the joined error.
//
// It returns a slice rather than an error so the caller can hold one list of
// problems, and it reports the FIRST offending column and the FIRST nil tile
// rather than every one: a grid built wrong is usually wrong uniformly, and a
// 13x11 map would otherwise return 143 copies of one mistake.
func (t *Tiles_t) validate() []error {
	if t == nil {
		return nil
	}

	var problems []error
	if t.TilesWide < 0 {
		problems = append(problems, errors.Join(ErrInvalidTileGrid,
			fmt.Errorf("map/tiles/@tilesWide (Tiles_t.TilesWide): %d: negative", t.TilesWide)))
	}
	if t.TilesHigh < 0 {
		problems = append(problems, errors.Join(ErrInvalidTileGrid,
			fmt.Errorf("map/tiles/@tilesHigh (Tiles_t.TilesHigh): %d: negative", t.TilesHigh)))
	}
	if len(t.Tiles) != t.TilesWide {
		problems = append(problems, errors.Join(ErrInvalidTileGrid,
			fmt.Errorf("map/tiles (Tiles_t.Tiles): @tilesWide is %d but the grid holds %d columns", t.TilesWide, len(t.Tiles))))
	}
	for x, column := range t.Tiles {
		if len(column) != t.TilesHigh {
			problems = append(problems, errors.Join(ErrInvalidTileGrid,
				fmt.Errorf("map/tiles (Tiles_t.Tiles): @tilesHigh is %d but column %d holds %d tiles", t.TilesHigh, x, len(column))))
			break
		}
	}
	for x, column := range t.Tiles {
		for y, tile := range column {
			if tile == nil {
				return append(problems, errors.Join(ErrInvalidTileGrid,
					fmt.Errorf("map/tiles (Tiles_t.Tiles[%d][%d]): nil tile", x, y)))
			}
		}
	}
	return problems
}
