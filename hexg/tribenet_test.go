// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import "testing"

func Test_TribeNet_new(t *testing.T) {
	for _, tc := range []struct {
		id           string
		input        string
		expectId     string
		expectCoords string
	}{
		{id: "N/A", input: "n/a", expectId: "N/A", expectCoords: "+0+0+0"},
		{id: "AA 0101", input: "AA 0101", expectId: "AA 0101", expectCoords: "+0+0+0"},
	} {
		a, err := NewTribeNetCoord(tc.input)
		if err != nil {
			t.Errorf("%s: new: got error, wanted nil", tc.id)
		} else if a.id != tc.expectId {
			t.Errorf("%s: id: got %q, wanted %q", tc.id, a.id, tc.expectId)
		} else if a.cube.String() != tc.expectCoords {
			t.Errorf("%s: cube: got %q, wanted %q", tc.id, a.cube.String(), tc.expectCoords)
		}
	}
}
