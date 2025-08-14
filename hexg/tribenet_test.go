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
		{id: "AA 0102", input: "AA 0102", expectId: "AA 0102", expectCoords: "+0+1-1"},
		{id: "AA 0103", input: "AA 0103", expectId: "AA 0103", expectCoords: "+0+2-2"},
		{id: "AA 0201", input: "AA 0201", expectId: "AA 0201", expectCoords: "+1+0-1"},
		{id: "AA 0202", input: "AA 0202", expectId: "AA 0202", expectCoords: "+1+1-2"},
		{id: "AA 0203", input: "AA 0203", expectId: "AA 0203", expectCoords: "+1+2-3"},
		{id: "AA 0301", input: "AA 0301", expectId: "AA 0301", expectCoords: "+2-1-1"},
		{id: "AA 0302", input: "AA 0302", expectId: "AA 0302", expectCoords: "+2+0-2"},
		{id: "AA 0303", input: "AA 0303", expectId: "AA 0303", expectCoords: "+2+1-3"},
		{id: "BC 0814", input: "BC 0814", expectId: "BC 0814", expectCoords: "+67+1-68"},
		{id: "JK 0609", input: "JK 0609", expectId: "JK 0609", expectCoords: "+305+45-350"},
		{id: "ZZ 3021", input: "ZZ 3021", expectId: "ZZ 3021", expectCoords: "+779+156-935"},
	} {
		a, err := NewTribeNetCoord(tc.input)
		if err != nil {
			t.Errorf("%s: new: got %v, wanted nil", tc.id, err)
		} else if a.id != tc.expectId {
			t.Errorf("%s: id: got %q, wanted %q", tc.id, a.id, tc.expectId)
		} else if a.cube.String() != tc.expectCoords {
			t.Errorf("%s: cube: got %q, wanted %q (%q)", tc.id, a.cube.String(), tc.expectCoords, a.GridID())
		}
	}
}
