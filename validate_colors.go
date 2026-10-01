// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"fmt"
)

// validateColors reports a colour with a component outside 0..1 (issue #128).
//
// Worldographer builds every colour it reads with JavaFX's Color constructor,
// which throws on a component outside 0..1, and the file does not open:
//
//	java.lang.IllegalArgumentException: Color's red value (255.0) must be in the range 0.0-1.0
//
// That is an app check on a shape style's dsColor (#83) and on a tile record's
// background (#128); the other attributes are inferred to fail the same way,
// since both of those, read by different code, did. #83 refused out-of-range
// colours only where Map_t holds them as strings (xmlio's rgbaAttr). Every
// colour Map_t holds as an *RGBA_t went to disk unchecked. This is the check
// for those.
//
// A component must be a finite number from 0 to 1: NaN and the infinities are
// refused too. A nil colour is not a problem here; what it is written as
// ("null", or no attribute) is the encoder's business.
//
// Like the resource checks it names the FIRST offending colour, by on-disk path
// and Map_t field, and counts the rest.
func (m *Map_t) validateColors() []error {
	var first error
	count := 0
	// outOfRange guards the two loops that run once per tile or placement, so
	// they format a field name only for a colour that fails. Formatting it
	// unconditionally cost about 2 million Sprintf calls and 250 MB per
	// Validate on a 1920 x 1080 map (issue #142).
	outOfRange := func(c *RGBA_t) bool { return c != nil && !c.inRange() }
	check := func(path, field string, c *RGBA_t) {
		if !outOfRange(c) {
			return
		}
		count++
		if first == nil {
			first = fmt.Errorf("%s (%s): %g,%g,%g,%g: want each component from 0 to 1", path, field, c.R, c.G, c.B, c.A)
		}
	}

	if m.Tiles != nil {
		for x, column := range m.Tiles.Tiles {
			for y, tile := range column {
				if tile != nil && outOfRange(tile.CustomBackgroundColor) {
					check("map/tiles/tilerow", fmt.Sprintf("Tiles_t.Tiles[%d][%d].CustomBackgroundColor", x, y), tile.CustomBackgroundColor)
				}
			}
		}
	}
	if m.ExtraTerrain != nil {
		for i, layer := range m.ExtraTerrain.MapLayers {
			if layer == nil {
				continue
			}
			for j, tl := range layer.Terrain {
				if tl != nil && outOfRange(tl.CustomBackgroundColor) {
					check(fmt.Sprintf("map/extraTerrain/mapLayer[@name=%q]/terrainAndLocation[%d]/@bgColor", layer.Name, j),
						fmt.Sprintf("ExtraTerrain_t.MapLayers[%d].Terrain[%d].CustomBackgroundColor", i, j), tl.CustomBackgroundColor)
				}
			}
		}
	}
	if m.MapKey != nil {
		check("map/mapkey/@backgroundcolor", "MapKey_t.BackgroundColor", m.MapKey.BackgroundColor)
		check("map/mapkey/@titleFontColor", "MapKey_t.TitleFontColor", m.MapKey.TitleFontColor)
		check("map/mapkey/@scaleFontColor", "MapKey_t.ScaleFontColor", m.MapKey.ScaleFontColor)
		check("map/mapkey/@entryFontColor", "MapKey_t.EntryFontColor", m.MapKey.EntryFontColor)
	}
	checkLabel := func(path, field string, l *Label_t) {
		if l == nil {
			return
		}
		check(path+"/@color", field+".Color", l.Color)
		check(path+"/@backgroundColor", field+".BackgroundColor", l.BackgroundColor)
		check(path+"/@outlineColor", field+".OutlineColor", l.OutlineColor)
	}
	for i, f := range m.Features {
		if f == nil {
			continue
		}
		path, field := fmt.Sprintf("map/features/feature[%d]", i), fmt.Sprintf("Map_t.Features[%d]", i)
		check(path+"/@color", field+".Color", f.Color)
		check(path+"/@ringColor", field+".RingColor", f.RingColor)
		checkLabel(path+"/label", field+".Label", f.Label)
	}
	for i, l := range m.Labels {
		checkLabel(fmt.Sprintf("map/labels/label[%d]", i), fmt.Sprintf("Map_t.Labels[%d]", i), l)
	}
	for i, n := range m.Notes {
		if n != nil {
			check(fmt.Sprintf("map/notes/note[%d]/@color", i), fmt.Sprintf("Map_t.Notes[%d].Color", i), n.Color)
		}
	}
	if m.Configuration != nil && m.Configuration.TextConfig != nil {
		for i, s := range m.Configuration.TextConfig.LabelStyles {
			if s == nil {
				continue
			}
			path := fmt.Sprintf("map/configuration/text-config/labelstyle[@name=%q]", s.Name)
			field := fmt.Sprintf("TextConfig_t.LabelStyles[%d]", i)
			check(path+"/@color", field+".Color", s.Color)
			check(path+"/@backgroundColor", field+".BackgroundColor", s.BackgroundColor)
			check(path+"/@outlineColor", field+".OutlineColor", s.OutlineColor)
		}
	}
	if m.Configuration != nil && m.Configuration.ShapeConfig != nil {
		for i, s := range m.Configuration.ShapeConfig.ShapeStyles {
			if s == nil {
				continue
			}
			path := fmt.Sprintf("map/configuration/shape-config/shapestyle[@name=%q]", s.Name)
			field := fmt.Sprintf("ShapeConfig_t.ShapeStyles[%d]", i)
			check(path+"/@strokePaint", field+".StrokePaint", s.StrokePaint)
			check(path+"/@fillPaint", field+".FillPaint", s.FillPaint)
			check(path+"/@dscolor", field+".DsColor", s.DsColor)
			check(path+"/@insColor", field+".InsColor", s.InsColor)
		}
	}

	if first == nil {
		return nil
	}
	if count > 1 {
		first = fmt.Errorf("%w (and %d more out-of-range colours)", first, count-1)
	}
	return []error{errors.Join(ErrInvalidColorAttribute, first)}
}

// inRange reports whether every component is a finite number from 0 to 1.
// The comparison is written so that NaN, which compares false with everything,
// is out of range.
func (c *RGBA_t) inRange() bool {
	for _, v := range []float64{c.R, c.G, c.B, c.A} {
		if !(v >= 0 && v <= 1) {
			return false
		}
	}
	return true
}
