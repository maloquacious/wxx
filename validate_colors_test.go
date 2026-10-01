// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// red255 is the colour #128's app check refused: Worldographer will not open
// a file with a component outside 0..1.
func red255() *RGBA_t { return &RGBA_t{R: 255, G: 0, B: 0, A: 1} }

// colorCases sets one *RGBA_t field to an out-of-range colour, keyed by
// "Type.Field". TestValidateColorsCoversEveryColorField holds the keys to every
// such field Map_t can reach, so a colour field added later fails that test
// until it is checked and listed here.
var colorCases = map[string]struct {
	set     func(m *Map_t)
	wantMsg string
}{
	"Tile_t.CustomBackgroundColor": {
		func(m *Map_t) { m.Tiles.Tiles[1][2].CustomBackgroundColor = red255() },
		"map/tiles/tilerow (Tiles_t.Tiles[1][2].CustomBackgroundColor): 255,0,0,1",
	},
	"TerrainAndLocation_t.CustomBackgroundColor": {
		func(m *Map_t) {
			m.ExtraTerrain = &ExtraTerrain_t{MapLayers: []*ExtraTerrainLayer_t{
				{Name: "Below All", Terrain: []*TerrainAndLocation_t{{CustomBackgroundColor: red255()}}},
			}}
		},
		`terrainAndLocation[0]/@bgColor (ExtraTerrain_t.MapLayers[0].Terrain[0].CustomBackgroundColor)`,
	},
	"MapKey_t.BackgroundColor": {func(m *Map_t) { m.MapKey.BackgroundColor = red255() }, "map/mapkey/@backgroundcolor"},
	"MapKey_t.TitleFontColor":  {func(m *Map_t) { m.MapKey.TitleFontColor = red255() }, "map/mapkey/@titleFontColor"},
	"MapKey_t.ScaleFontColor":  {func(m *Map_t) { m.MapKey.ScaleFontColor = red255() }, "map/mapkey/@scaleFontColor"},
	"MapKey_t.EntryFontColor":  {func(m *Map_t) { m.MapKey.EntryFontColor = red255() }, "map/mapkey/@entryFontColor"},
	"Feature_t.Color": {
		func(m *Map_t) { m.Features = []*Feature_t{{Color: red255()}} },
		"map/features/feature[0]/@color (Map_t.Features[0].Color)",
	},
	"Feature_t.RingColor": {
		func(m *Map_t) { m.Features = []*Feature_t{{RingColor: red255()}} },
		"map/features/feature[0]/@ringColor (Map_t.Features[0].RingColor)",
	},
	"Label_t.Color": {
		func(m *Map_t) { m.Labels = []*Label_t{{Color: red255()}} },
		"map/labels/label[0]/@color (Map_t.Labels[0].Color)",
	},
	"Label_t.BackgroundColor": {
		// The label inside a feature is checked too, not only standalone ones.
		func(m *Map_t) { m.Features = []*Feature_t{{Label: &Label_t{BackgroundColor: red255()}}} },
		"map/features/feature[0]/label/@backgroundColor (Map_t.Features[0].Label.BackgroundColor)",
	},
	"Label_t.OutlineColor": {
		func(m *Map_t) { m.Labels = []*Label_t{{OutlineColor: red255()}} },
		"map/labels/label[0]/@outlineColor (Map_t.Labels[0].OutlineColor)",
	},
	"Note_t.Color": {
		func(m *Map_t) { m.Notes = []*Note_t{{Color: red255()}} },
		"map/notes/note[0]/@color (Map_t.Notes[0].Color)",
	},
	"LabelStyle_t.Color": {
		func(m *Map_t) {
			m.Configuration.TextConfig.LabelStyles = []*LabelStyle_t{{Name: "City", Color: red255()}}
		},
		`labelstyle[@name="City"]/@color (TextConfig_t.LabelStyles[0].Color)`,
	},
	"LabelStyle_t.BackgroundColor": {
		func(m *Map_t) {
			m.Configuration.TextConfig.LabelStyles = []*LabelStyle_t{{Name: "City", BackgroundColor: red255()}}
		},
		"TextConfig_t.LabelStyles[0].BackgroundColor",
	},
	"LabelStyle_t.OutlineColor": {
		func(m *Map_t) {
			m.Configuration.TextConfig.LabelStyles = []*LabelStyle_t{{Name: "City", OutlineColor: red255()}}
		},
		"TextConfig_t.LabelStyles[0].OutlineColor",
	},
	"ShapeStyle_t.StrokePaint": {
		func(m *Map_t) {
			m.Configuration.ShapeConfig.ShapeStyles = []*ShapeStyle_t{{Name: "Road", StrokePaint: red255()}}
		},
		`shapestyle[@name="Road"]/@strokePaint (ShapeConfig_t.ShapeStyles[0].StrokePaint)`,
	},
	"ShapeStyle_t.FillPaint": {
		func(m *Map_t) {
			m.Configuration.ShapeConfig.ShapeStyles = []*ShapeStyle_t{{Name: "Road", FillPaint: red255()}}
		},
		"ShapeConfig_t.ShapeStyles[0].FillPaint",
	},
	"ShapeStyle_t.DsColor": {
		// dscolor is the attribute #83's app check refused.
		func(m *Map_t) {
			m.Configuration.ShapeConfig.ShapeStyles = []*ShapeStyle_t{{Name: "Road", DsColor: red255()}}
		},
		"ShapeConfig_t.ShapeStyles[0].DsColor",
	},
	"ShapeStyle_t.InsColor": {
		func(m *Map_t) {
			m.Configuration.ShapeConfig.ShapeStyles = []*ShapeStyle_t{{Name: "Road", InsColor: red255()}}
		},
		"ShapeConfig_t.ShapeStyles[0].InsColor",
	},
}

// TestValidateColorsRejects sets each colour field, one at a time, to 255 red
// on a map that otherwise passes (issue #128).
func TestValidateColorsRejects(t *testing.T) {
	for name, tc := range colorCases {
		t.Run(name, func(t *testing.T) {
			m := validMap()
			tc.set(m)
			err := m.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want %v", ErrInvalidColorAttribute)
			}
			if !errors.Is(err, ErrInvalidColorAttribute) {
				t.Errorf("Validate() = %v, want errors.Is(err, %v)", err, ErrInvalidColorAttribute)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("Validate() = %q, want a message containing %q", err.Error(), tc.wantMsg)
			}
		})
	}
}

// TestValidateColorsCoversEveryColorField walks Map_t's type and finds every
// *RGBA_t field a map can reach. Each must have a case in colorCases: a colour
// the check skips is a colour wxx can still write out of range.
func TestValidateColorsCoversEveryColorField(t *testing.T) {
	rgba := reflect.TypeOf((*RGBA_t)(nil))
	var found []string
	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type)
	walk = func(typ reflect.Type) {
		for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
			typ = typ.Elem()
		}
		if typ.Kind() != reflect.Struct || seen[typ] {
			return
		}
		seen[typ] = true
		for i := range typ.NumField() {
			f := typ.Field(i)
			if f.Type == rgba {
				found = append(found, typ.Name()+"."+f.Name)
				continue
			}
			walk(f.Type)
		}
	}
	walk(reflect.TypeOf(Map_t{}))

	if len(found) < 19 {
		t.Fatalf("found %d *RGBA_t fields, want at least the 19 #128 found: %v", len(found), found)
	}
	for _, name := range found {
		if _, ok := colorCases[name]; !ok {
			t.Errorf("%s is an *RGBA_t that colorCases does not test; add it to validateColors and to colorCases", name)
		}
	}
	for name := range colorCases {
		if !slices.Contains(found, name) {
			t.Errorf("colorCases has %s, which Map_t no longer reaches", name)
		}
	}
}

// TestValidateColorsComponents checks the range itself: each component, each
// end, and the values a float can hold that a colour cannot.
func TestValidateColorsComponents(t *testing.T) {
	for _, tc := range []struct {
		name string
		c    RGBA_t
		ok   bool
	}{
		{"all 0", RGBA_t{}, true},
		{"all 1", RGBA_t{R: 1, G: 1, B: 1, A: 1}, true},
		{"opaque black", RGBA_t{A: 1}, true},
		{"red above 1", RGBA_t{R: 1.0001, A: 1}, false},
		{"green negative", RGBA_t{G: -0.1, A: 1}, false},
		{"blue 255", RGBA_t{B: 255, A: 1}, false},
		{"alpha above 1", RGBA_t{A: 2}, false},
		{"NaN", RGBA_t{R: math.NaN(), A: 1}, false},
		{"+Inf", RGBA_t{G: math.Inf(1), A: 1}, false},
		{"-Inf", RGBA_t{B: math.Inf(-1), A: 1}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := validMap()
			c := tc.c
			m.Tiles.Tiles[0][0].CustomBackgroundColor = &c
			err := m.Validate()
			if tc.ok && err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
			if !tc.ok && !errors.Is(err, ErrInvalidColorAttribute) {
				t.Errorf("Validate() = %v, want errors.Is(err, %v)", err, ErrInvalidColorAttribute)
			}
		})
	}
}

// TestValidateColorsCountsTheRest names the first bad colour and counts the
// others rather than listing every one.
func TestValidateColorsCountsTheRest(t *testing.T) {
	m := validMap()
	m.Tiles.Tiles[0][0].CustomBackgroundColor = red255()
	m.Tiles.Tiles[1][1].CustomBackgroundColor = red255()
	m.MapKey.BackgroundColor = red255()
	err := m.Validate()
	if want := "Tiles_t.Tiles[0][0].CustomBackgroundColor): 255,0,0,1: want each component from 0 to 1 (and 2 more out-of-range colours)"; err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("Validate() = %v, want a message containing %q", err, want)
	}
}

// TestValidateDoesNotAllocatePerTile pins issue #142: a map whose colours are
// all in range costs Validate the same allocations at any size. validateColors
// used to format a field name for every tile and every <extraTerrain>
// placement before checking the colour, about 2 million Sprintf calls and
// 250 MB per Validate on a 1920 x 1080 map.
//
// Every tile and placement carries an in-range colour, so the check reaches
// each one rather than skipping a nil.
func TestValidateDoesNotAllocatePerTile(t *testing.T) {
	sized := func(tilesWide, tilesHigh, placements int) *Map_t {
		m := validMap()
		m.Tiles = &Tiles_t{TilesWide: tilesWide, TilesHigh: tilesHigh}
		for x := 0; x < tilesWide; x++ {
			column := make([]*Tile_t, tilesHigh)
			for y := range column {
				column[y] = &Tile_t{CustomBackgroundColor: &RGBA_t{R: 0.5, G: 0.5, B: 0.5, A: 1}}
			}
			m.Tiles.Tiles = append(m.Tiles.Tiles, column)
		}
		layer := &ExtraTerrainLayer_t{Name: "Below All"}
		for range placements {
			layer.Terrain = append(layer.Terrain, &TerrainAndLocation_t{CustomBackgroundColor: &RGBA_t{R: 0.5, G: 0.5, B: 0.5, A: 1}})
		}
		m.ExtraTerrain = &ExtraTerrain_t{MapLayers: []*ExtraTerrainLayer_t{layer}}
		return m
	}
	allocs := func(m *Map_t) float64 {
		t.Helper()
		if err := m.Validate(); err != nil {
			t.Fatalf("Validate: %v", err)
		}
		return testing.AllocsPerRun(10, func() { _ = m.Validate() })
	}

	small, large := allocs(sized(2, 3, 1)), allocs(sized(200, 100, 1000))
	if large != small {
		t.Errorf("Validate allocates %v times on a 200x100 map with 1000 placements and %v on a 2x3 map with 1; want the same: something allocates per tile or per placement",
			large, small)
	}
}
