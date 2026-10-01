// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"encoding/json"
	"testing"

	"github.com/maloquacious/hexg"
)

// TestTileCoordsRoundTripJSON pins issue #52's acceptance that a Tile_t keeps
// its cube coordinates through JSON. The vendored wxx/hexg had unexported
// fields and no marshaller, so Coords marshalled as {} and came back as the
// origin; the standalone hexg.Hex marshals itself. Negative components are
// included because the concise "+q+r+s" form is parsed by sign.
func TestTileCoordsRoundTripJSON(t *testing.T) {
	for _, oc := range []hexg.OffsetCoord{
		hexg.NewOffsetCoord(0, 0),
		hexg.NewOffsetCoord(1, 0),
		hexg.NewOffsetCoord(12, 10),
		hexg.NewOffsetCoord(-3, 7),
		hexg.NewOffsetCoord(5, -9),
		hexg.NewOffsetCoord(-17, -4),
	} {
		in := Tile_t{Coords: oc.QOffsetToCube(false), Column: oc.Col, Row: oc.Row}
		data, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("%v: json.Marshal: %v", oc, err)
		}
		var out Tile_t
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("%v: json.Unmarshal(%s): %v", oc, data, err)
		}
		if out.Coords != in.Coords {
			t.Errorf("%v: Coords = %s after the round trip, want %s (json %s)", oc, out.Coords, in.Coords, data)
		}
	}
}
