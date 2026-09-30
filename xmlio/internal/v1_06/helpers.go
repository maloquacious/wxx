// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"errors"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// boold formats a bool as an integer
func boold(b bool) int {
	if b {
		return 1
	}
	return 0
}

// bools formats a bool as a string
func bools(b bool) string {
	return fmt.Sprintf("%v", b)
}

// floatd formats a float as an integer.
func floatd(f float64) int {
	return int(f)
}

// floatf formats a float in the style that Worldographer expects.
// Zero values are rendered as 0.0.
// Note: floats is probably the right function to use.
func floatf(f float64) string {
	const epsilon = 1e-6
	if -epsilon < f && f <= epsilon {
		return "0.0"
	}
	return fmt.Sprintf("%g", f)
}

// floats converts a float64 number to a string representation adhering
// to certain Worldographer formatting rules.
//
// The function tries to represent the float in a manner that avoids scientific notation
// while preserving the fractional part of the float. It rounds off trailing zeros and
// ensures that there is always a digit after the decimal point.
//
// Parameters:
// - f: The float64 number to be converted.
//
// Returns:
//   - The string representation of the input float. If `f` is an integer, ".0" is appended to
//     signify that it is a float. For non-integer floats, trailing zeros after the decimal point are trimmed.
//
// Example:
//
//	floats(1234567.00) returns "1234567.0"
//	floats(0.120300) returns "0.1203"
func floats(f float64) string {
	s := fmt.Sprintf("%g", f)
	if strings.IndexByte(s, 'e') != -1 {
		s = fmt.Sprintf("%f", f)
	}
	if strings.IndexByte(s, '.') == -1 {
		return s + ".0"
	}
	s = strings.TrimRight(s, "0")
	if s[len(s)-1] == '.' {
		return s + "0"
	}
	return s
}

// floatg formats a float in the style that Worldographer expects.
func floatg(f float64) string {
	return fmt.Sprintf("%g", f)
}

// ints formats an int as a string
func ints(i int) string {
	return fmt.Sprintf("%d", i)
}

// rgbaOrNull renders a nullable colour: the literal "null" for nil, and the
// colour itself otherwise (issue #62).
//
// rgbas cannot do this: it renders nil as "0.0,0.0,0.0,1.0", so a file that
// said "null" came back claiming an opaque black -- that is the bug #62
// reports, and it was live on every <labelstyle> this codec wrote. The old
// rgbans helper went the other way: it rendered nil as "null", but it decided
// by comparing the FORMATTED STRING, so a genuine opaque black was laundered
// into "null" as well -- a feature coloured Black came back uncoloured (issue
// #99). This one is wrong for neither input, because it asks the pointer
// rather than the string.
//
// That only works where nil means "null" and nothing else, so the decode half
// has to reserve it: decodeZeroableRgba maps "" and "null" to nil and leaves
// black alone, while decodeRgba folds black into nil as well. A field decoded
// with decodeRgba must NOT be encoded with this -- the model has already lost
// which spelling the file used, and the honest rendering is whatever the old
// helper did. See the labelstyle decode in configuration.go, which was moved to
// decodeZeroableRgba for exactly this reason.
//
// Every nullable RGBA attribute now pairs decodeZeroableRgba with this
// helper (issue #99): feature @color and @ringColor, note @color, and
// shapestyle @fillPaint, @dscolor and @insColor, as well as the two labelstyle
// colours.
func rgbaOrNull(rgba *wxx.RGBA_t) string {
	if rgba == nil {
		return "null"
	}
	return rgbas(rgba)
}

// rgbas converts an RGBA_t struct into an XML attribute string.
// RGBA_t struct contains four fields, each representing Red, Green, Blue and Alpha respectively.
// Each field is a float. We format the struct as a comma separated string.
// If the provided rgba pointer is nil, it defaults to "0.0,0.0,0.0,1.0".
//
// We use the floats function to format the float values into an XML-friendly format.
//
// Parameters:
// - rgba: a pointer to an RGBA_t struct. Can be nil.
//
// Returns:
// - A XML attribute string representing the rgba. If rgba is nil, returns "0.0,0.0,0.0,1.0"
func rgbas(rgba *wxx.RGBA_t) string {
	if rgba == nil {
		return "0.0,0.0,0.0,1.0"
	}
	return fmt.Sprintf("%s,%s,%s,%s",
		floats(rgba.R),
		floats(rgba.G),
		floats(rgba.B),
		floats(rgba.A))
}

// terrainTable is the <terrainmap> the encoder writes and how tiles refer to
// it (issue #87).
//
// Worldographer numbers the table 0..n-1 and renumbers it on every save, so an
// index carries no meaning beyond linking a tile to a terrain. The table is
// written in index order, numbered by position, and remap takes each Map_t
// index to the position it is written at. Tiles are written through remap, so
// a table with gaps -- a map built or edited in code -- is renumbered together
// with every tile that uses it, and each tile keeps its terrain. For a decoded
// map the table is already 0..n-1 and remap is the identity.
//
// Two terrains sharing an index have no correct renumbering, so they are
// refused.
func terrainTable(data map[string]int) (names []string, remap map[int]int, err error) {
	type entry struct {
		index int
		name  string
	}
	var entries []entry
	for name, index := range data {
		entries = append(entries, entry{index, name})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].index != entries[j].index {
			return entries[i].index < entries[j].index
		}
		return entries[i].name < entries[j].name
	})
	remap = map[int]int{}
	for i, e := range entries {
		if i > 0 && entries[i-1].index == e.index {
			return nil, nil, errors.Join(wxx.ErrInvalidTerrainMap, fmt.Errorf(
				"<terrainmap>: %q and %q both have index %d; a tile using it could mean either (issue #87)", entries[i-1].name, e.name, e.index))
		}
		names = append(names, e.name)
		remap[e.index] = i
	}
	return names, remap, nil
}

// tileTerrain is the index a tile is written with: its terrain's position in
// the written table. A tile whose terrain the table does not list is refused;
// Worldographer would have no terrain to draw.
func tileTerrain(remap map[int]int, tile *wxx.Tile_t) (int, error) {
	index, ok := remap[tile.Terrain]
	if !ok {
		return 0, errors.Join(wxx.ErrInvalidTileGrid, fmt.Errorf(
			"map/tiles: hex (%d,%d) uses terrain index %d, which <terrainmap> does not list (issue #87)", tile.Column, tile.Row, tile.Terrain))
	}
	return index, nil
}

func encodeInnerText(input string) string {
	escaped := html.EscapeString(input) // Escapes < > & "
	return strings.ReplaceAll(escaped, "\n", "&#10;")
}

// xmlAttr renders s as a double-quoted XML attribute value (issue #71).
//
// Every attribute the encoder writes goes through this. It replaced fmt's %q,
// which produces a Go string literal, not XML: it wrote & and < raw, which is
// not well-formed, and " as \", which ends the attribute early, so any user
// text containing them made a file that was not XML. It also wrote tab and
// newline as the two characters \t and \n, silently changing the value.
//
// The five markup characters are written as entities. Tab, newline and
// carriage return are written as character references, because a parser
// normalizes those characters to a space when it reads them raw from an
// attribute value. Every other control character is also written as a
// character reference, which XML 1.1 -- the version every W2025 file declares
// -- permits. Anything else, non-ASCII included, is written as is: the document
// is UTF-8 until the transport stage converts it to UTF-16, and no W2025 sample
// shows Worldographer escaping non-ASCII.
func xmlAttr(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '&':
			b.WriteString("&amp;")
		case r == '<':
			b.WriteString("&lt;")
		case r == '>':
			b.WriteString("&gt;")
		case r == '"':
			b.WriteString("&quot;")
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, "&#%d;", r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// rgbaAttr checks a color attribute the model holds as a string and writes
// verbatim, and returns it unchanged if Worldographer can read it (issue #83).
//
// The spelling is "r,g,b,a": four decimal numbers from 0 to 1, as every sample
// writes them. With nullable, the literal "null" is accepted too, as the
// samples write it for an absent shadow color. Anything else is refused before
// a byte is written. An empty string, which is what a field left at its zero
// value holds, makes Worldographer fail to open the file:
//
//	java.lang.NumberFormatException: empty String
//	    at java.base/java.lang.Double.parseDouble(Unknown Source)
//	    at com.inkwellideas.ographer.task.LoadMapTask.readShapes(LoadMapTask.java:784)
//
// The 0-to-1 range is confirmed in the app: dsColor="255,0,0,1" fails with
// "Color's red value (255.0) must be in the range 0.0-1.0" from JavaFX's Color
// constructor (maintainer's app check, #83).
func rgbaAttr(path, s string, nullable bool) (string, error) {
	if nullable && s == "null" {
		return s, nil
	}
	parts := strings.Split(s, ",")
	ok := len(parts) == 4
	for _, p := range parts {
		if !ok {
			break
		}
		f, err := strconv.ParseFloat(p, 64)
		ok = err == nil && 0 <= f && f <= 1
	}
	if !ok {
		want := `four comma-separated numbers from 0 to 1, e.g. "1.0,1.0,1.0,1.0"`
		if nullable {
			want += `, or "null"`
		}
		return "", errors.Join(wxx.ErrInvalidColorAttribute, fmt.Errorf(
			"%s = %q: want %s; Worldographer will not open a file with anything else here (issue #83)", path, s, want))
	}
	return s, nil
}

// hexColorAttr checks a <gridandnumbering> color, which the samples spell as
// "0x" and eight hex digits (RRGGBBAA), and returns it unchanged if it has that
// form (issue #83).
//
// "null" and "" are accepted too. No sample states either, but Worldographer
// 2.08 opens a file with color0="" and saves it back as color0="null"
// (maintainer's app check, #83), so both are colors it reads as "none".
func hexColorAttr(path, s string) (string, error) {
	if s == "null" || s == "" {
		return s, nil
	}
	if len(s) == 10 && strings.HasPrefix(s, "0x") {
		if _, err := strconv.ParseUint(s[2:], 16, 32); err == nil {
			return s, nil
		}
	}
	return "", errors.Join(wxx.ErrInvalidColorAttribute, fmt.Errorf(
		`%s = %q: want "0x" and eight hex digits, e.g. "0x00000040", or "null" (issue #83)`, path, s))
}
