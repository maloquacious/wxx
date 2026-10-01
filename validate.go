// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"fmt"

	"github.com/maloquacious/hexg"
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
	// HexOrientation is the string the file states and the one the encoder
	// switches on; GridOrientation is the hexg coordinate convention the decoder
	// sets from it in the same switch. Two fields holding one fact can disagree,
	// and a map that says COLUMNS in one and odd-r in the other describes no
	// hex grid that exists. hexg.LayoutOffset has no "unset" value -- its zero
	// is OddR -- so a caller who set only the string gets OddR, which this
	// check catches for COLUMNS but cannot catch for ROWS.
	switch m.HexOrientation {
	case "COLUMNS":
		if m.GridOrientation != hexg.EvenQ && m.GridOrientation != hexg.OddQ {
			problems = append(problems, errors.Join(ErrMismatchedGridOrientation,
				fmt.Errorf("hexOrientation %q with gridOrientation %s: want even-q or odd-q", m.HexOrientation, layoutOffsetName(m.GridOrientation))))
		}
	case "ROWS":
		if m.GridOrientation != hexg.EvenR && m.GridOrientation != hexg.OddR {
			problems = append(problems, errors.Join(ErrMismatchedGridOrientation,
				fmt.Errorf("hexOrientation %q with gridOrientation %s: want even-r or odd-r", m.HexOrientation, layoutOffsetName(m.GridOrientation))))
		}
	default:
		problems = append(problems, errors.Join(ErrInvalidHexOrientation,
			fmt.Errorf("hexOrientation %q: want \"COLUMNS\" or \"ROWS\"", m.HexOrientation)))
	}

	// Substructures an encoder dereferences without checking.
	//
	// Every entry was verified to panic a codec when nil. MapKey and
	// Informations panicked only v1_06 -- the classic codec, removed in issue
	// #103, wrote a hard-coded <mapkey> and ignored <informations> -- and were
	// required here anyway: the invariant is a property of the model, not of a
	// target (ADR 0004 Decision 6), so a map missing one is incomplete whoever
	// is about to write it.
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
	problems = append(problems, m.ExtraTerrain.validateResources()...)
	problems = append(problems, m.validateColors()...)

	// Join drops nils and returns nil for an empty slice, so a valid map returns
	// nil without a length check here.
	return errors.Join(problems...)
}

// validate reports the ways the tile grid contradicts its own header.
//
// The header is <tiles tilesWide= tilesHigh=> and the grid is Tiles[col][row];
// the encoder loops to the HEADER's dimensions and indexes the grid, so a header
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
	return append(problems, t.validateResources()...)
}

// validateResources reports a tile resource outside 0..100 (issue #122).
//
// The encoder writes a resource as the integer it is given, and Worldographer
// 2.08 does not keep a value above 100: it reads the field with
// Byte.parseByte, so 128 and above stop the file opening, and 101..127 open
// and are saved back as 100 (app checks on Brick, #122). The v1_06 decoder
// refuses anything outside 0..100, so before this check wxx wrote a file it
// could not read back. Negative values are refused for the same reason; the
// app was not tried with one.
//
// Like the nil-tile check, it names the FIRST offending tile and field, and
// counts the rest, rather than listing one problem per tile.
func (t *Tiles_t) validateResources() []error {
	var first error
	count := 0
	for x, column := range t.Tiles {
		for y, tile := range column {
			for _, f := range tile.Resources.outOfRange() {
				count++
				if first == nil {
					first = fmt.Errorf("map/tiles (Tiles_t.Tiles[%d][%d].Resources.%s): %d: want 0..100", x, y, f.name, f.value)
				}
			}
		}
	}
	return resourceProblem(ErrInvalidTileResource, first, count)
}

// validateResources reports an <extraTerrain> resource outside 0..100 (issue
// #124).
//
// The seven resources on a <terrainAndLocation> are the ones a tile record
// carries, for a terrain on another layer. Worldographer 2.08 reads them with
// Byte.parseByte too: a value of 150 stops the file opening (app check, #124).
// Whether it clamps 101..127 here, as it does in a tile record, is untested;
// the check refuses them anyway, so a map wxx writes never depends on that.
//
// The decoder clamps an out-of-range value on the way in and reports it (see
// xmlio.DecoderDiagnostics.Clamped), so a decoded map always passes this. Only a
// caller's own assignment can trip it.
func (e *ExtraTerrain_t) validateResources() []error {
	if e == nil {
		return nil
	}
	var first error
	count := 0
	for i, layer := range e.MapLayers {
		if layer == nil {
			continue
		}
		for j, tl := range layer.Terrain {
			if tl == nil {
				continue
			}
			for _, f := range tl.Resources.outOfRange() {
				count++
				if first == nil {
					first = fmt.Errorf("map/extraTerrain/mapLayer[@name=%q]/terrainAndLocation[%d]/@resources (ExtraTerrain_t.MapLayers[%d].Terrain[%d].Resources.%s): %d: want 0..100",
						layer.Name, j, i, j, f.name, f.value)
				}
			}
		}
	}
	return resourceProblem(ErrInvalidExtraTerrainResource, first, count)
}

// resourceField is one named resource value.
type resourceField struct {
	name  string
	value int
}

// outOfRange returns the resources outside 0..100, in field order.
func (r Resources_t) outOfRange() []resourceField {
	var bad []resourceField
	for _, f := range []resourceField{
		{"Animal", r.Animal}, {"Brick", r.Brick}, {"Crops", r.Crops}, {"Gems", r.Gems},
		{"Lumber", r.Lumber}, {"Metals", r.Metals}, {"Rock", r.Rock},
	} {
		if f.value < 0 || f.value > 100 {
			bad = append(bad, f)
		}
	}
	return bad
}

// resourceProblem wraps the first out-of-range resource in kind and counts the
// rest, or returns nil when there were none.
func resourceProblem(kind Error, first error, count int) []error {
	if first == nil {
		return nil
	}
	if count > 1 {
		first = fmt.Errorf("%w (and %d more out-of-range resources)", first, count-1)
	}
	return []error{errors.Join(kind, first)}
}

// layoutOffsetName names a hexg.LayoutOffset in the spelling the vendored
// hexg.Orientation_e used, because hexg v1.3.0 gives LayoutOffset no String
// method and the diagnostics above name the convention.
func layoutOffsetName(o hexg.LayoutOffset) string {
	switch o {
	case hexg.EvenQ:
		return "even-q"
	case hexg.OddQ:
		return "odd-q"
	case hexg.EvenR:
		return "even-r"
	case hexg.OddR:
		return "odd-r"
	}
	return fmt.Sprintf("hexg.LayoutOffset(%d)", int(o))
}
