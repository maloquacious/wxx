// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/maloquacious/hexg"
)

// TestMapOrientationRoundTripJSON pins issue #52's orientation zero-value
// hazard. hexg.LayoutOffset has no unset value -- its zero is OddR -- so a
// stored layout field made a map whose orientation was never set round-trip
// through JSON as OddR, and omitempty dropped a genuine OddR from the JSON
// entirely. The layout is now derived from HexOrientation, so:
//
//   - a map with no orientation reports none (ok=false), not OddR;
//   - a ROWS or COLUMNS map reports its layout after the trip, and the JSON
//     itself carries the orientation string, so the layout cannot survive
//     merely because the decoded zero value happens to be OddR.
func TestMapOrientationRoundTripJSON(t *testing.T) {
	roundTrip := func(t *testing.T, in *Map_t) (*Map_t, string) {
		t.Helper()
		data, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		var out Map_t
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("json.Unmarshal(%s): %v", data, err)
		}
		return &out, string(data)
	}

	t.Run("unset", func(t *testing.T) {
		out, data := roundTrip(t, &Map_t{})
		if out.HexOrientation != "" {
			t.Errorf("HexOrientation = %q after the round trip, want unset", out.HexOrientation)
		}
		if got, ok := out.GridOrientation(); ok {
			t.Errorf("GridOrientation() = (%d, true) after the round trip of a map whose orientation was never set, want ok=false (json %s)", got, data)
		}
	})

	for _, tc := range []struct {
		name        string
		orientation string
		want        hexg.LayoutOffset
	}{
		{name: "rows is odd-r", orientation: "ROWS", want: hexg.OddR},
		{name: "columns is odd-q", orientation: "COLUMNS", want: hexg.OddQ},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, data := roundTrip(t, &Map_t{HexOrientation: tc.orientation})
			if wantJSON := `"hexOrientation":"` + tc.orientation + `"`; !strings.Contains(data, wantJSON) {
				t.Errorf("json lacks %s: %s", wantJSON, data)
			}
			if out.HexOrientation != tc.orientation {
				t.Errorf("HexOrientation = %q after the round trip, want %q (json %s)", out.HexOrientation, tc.orientation, data)
			}
			if got, ok := out.GridOrientation(); got != tc.want || !ok {
				t.Errorf("GridOrientation() = (%d, %v) after the round trip, want (%d, true) (json %s)", got, ok, tc.want, data)
			}
		})
	}
}
