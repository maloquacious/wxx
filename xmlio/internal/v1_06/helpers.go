// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"fmt"
	"html"
	"sort"
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

// rgbans converts an RGBA_t to a nullable string.
// It uses the rgbas function to format the RGBA_t
func rgbans(rgba *wxx.RGBA_t) string {
	s := rgbas(rgba)
	if s == "0.0,0.0,0.0,1.0" {
		s = "null"
	}
	return s
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

// terrainMapToSlice converts a map of terrain names and slot into a list
// of strings for the xml map.terrainmap element.
func terrainMapToSlice(data map[string]int) []string {
	type terrain_t struct {
		slot int
		name string
	}
	list := []*terrain_t{}
	for k, v := range data {
		list = append(list, &terrain_t{
			slot: v,
			name: k,
		})
	}
	// list must be sorted
	sort.Slice(list, func(i, j int) bool {
		return list[i].slot < list[j].slot
	})
	var s []string
	for _, v := range list {
		s = append(s, v.name)
	}
	return s
}

func encodeInnerText(input string) string {
	escaped := html.EscapeString(input) // Escapes < > & "
	return strings.ReplaceAll(escaped, "\n", "&#10;")
}
