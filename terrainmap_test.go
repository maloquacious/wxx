// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import "testing"

// TestEnsureTerrain: a listed terrain keeps its index; an unlisted one is added
// one past the highest index, in both Data and List, so a table numbered
// 0..n-1 stays contiguous (issue #81).
func TestEnsureTerrain(t *testing.T) {
	// The maintainer's fully painted ROWS map, as Worldographer 2.08 wrote it:
	// three terrains, no Blank.
	tm := &TerrainMap_t{
		Data: map[string]int{"Classic/Water Sea": 1, "Classic/Flat Beach": 0, "Classic/Flat Farmland": 2},
		List: []*Terrain_t{{Index: 1, Label: "Classic/Water Sea"}, {Index: 0, Label: "Classic/Flat Beach"}, {Index: 2, Label: "Classic/Flat Farmland"}},
	}
	if got := tm.EnsureTerrain("Classic/Flat Beach"); got != 0 {
		t.Errorf("EnsureTerrain(listed) = %d, want its index 0", got)
	}
	if got := tm.EnsureTerrain(BlankTerrain); got != 3 {
		t.Errorf("EnsureTerrain(Blank) = %d, want 3, one past the highest", got)
	}
	if got := tm.EnsureTerrain(BlankTerrain); got != 3 {
		t.Errorf("second EnsureTerrain(Blank) = %d, want the same 3", got)
	}
	if len(tm.Data) != 4 || len(tm.List) != 4 {
		t.Errorf("table has %d Data and %d List entries, want 4 of each: Blank added once, to both", len(tm.Data), len(tm.List))
	}
	if last := tm.List[len(tm.List)-1]; last.Index != 3 || last.Label != BlankTerrain {
		t.Errorf("List gained %+v, want {3 Blank}", *last)
	}

	var empty TerrainMap_t
	if got := empty.EnsureTerrain(BlankTerrain); got != 0 || empty.Data[BlankTerrain] != 0 {
		t.Errorf("EnsureTerrain on an empty table = %d, want 0", got)
	}
}
