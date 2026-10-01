// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"math"
	"strconv"
	"testing"
)

// TestFloats pins floats to Java's Double.toString, which is how Worldographer
// writes a double (issue #111).
//
// The fixture rows are spellings the tracked saves state. The others follow
// the Double.toString specification: the plain form for 1e-3 <= |f| < 1e7, the
// computerized scientific form outside it, and the names Java gives NaN and the
// infinities. No fixture states a value at those boundaries or a non-finite one.
func TestFloats(t *testing.T) {
	for _, tc := range []struct {
		f    float64
		want string
	}{
		// stated by tracked fixtures
		{4.263256414560601e-14, "4.263256414560601E-14"}, // map/@vScrollbarPos, 2.07 and 2.08 rows
		{0, "0.0"},
		{1950, "1950.0"},
		{0.5, "0.5"},
		{1.4142135623730951, "1.4142135623730951"},

		// Double.toString, plain form
		{math.Copysign(0, -1), "-0.0"},
		{1e-3, "0.001"},
		{0.1203, "0.1203"},
		{1234567, "1234567.0"},
		{1234567.123456789, "1234567.123456789"},
		{9999999.999999998, "9999999.999999998"},
		{-2.5, "-2.5"},

		// Double.toString, computerized scientific form
		{9.99e-4, "9.99E-4"},
		{5e-4, "5.0E-4"},
		{1e7, "1.0E7"},
		{-1.5e7, "-1.5E7"},
		{1.2345e100, "1.2345E100"},
		{math.MaxFloat64, "1.7976931348623157E308"},
		{math.SmallestNonzeroFloat64, "4.9E-324"},

		// Double.toString, non-finite
		{math.NaN(), "NaN"},
		{math.Inf(1), "Infinity"},
		{math.Inf(-1), "-Infinity"},
	} {
		got := floats(tc.f)
		if got != tc.want {
			t.Errorf("floats(%v) = %q, want %q", tc.f, got, tc.want)
		}
		// Every spelling must read back as the same value, or the file and the
		// model disagree.
		if back, err := strconv.ParseFloat(got, 64); err != nil {
			t.Errorf("floats(%v) = %q does not parse: %v", tc.f, got, err)
		} else if back != tc.f && !(math.IsNaN(back) && math.IsNaN(tc.f)) {
			t.Errorf("floats(%v) = %q reads back as %v", tc.f, got, back)
		}
	}
}
