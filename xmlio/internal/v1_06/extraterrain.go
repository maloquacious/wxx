// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// decodeExtraTerrain copies the optional top-level <extraTerrain> element into
// the domain map (issue #34). src is nil when the element is absent; an empty
// container decodes to a non-nil ExtraTerrain_t with no layers, so it is written
// back.
//
// It refuses content the schema does not model -- an unknown child element or
// attribute at any level -- rather than dropping it. Before #34 this element
// was carried verbatim, so anything in it survived a round trip; modeling it
// must not quietly change that for content nobody has seen yet.
//
// Resource values outside 0..100 are clamped, not refused, and each one is
// appended to clamps; see decodeExtraTerrainResources.
func decodeExtraTerrain(src *ExtraTerrain_t, w *wxx.Map_t, clamps *[]Clamp_t) error {
	if src == nil {
		return nil
	}
	if err := refuseUnmodeled("map/extraTerrain", nil, src.Other); err != nil {
		return err
	}
	w.ExtraTerrain = &wxx.ExtraTerrain_t{}
	for i, layer := range src.MapLayers {
		path := fmt.Sprintf("map/extraTerrain/mapLayer[@name=%q]", layer.Name)
		if err := refuseUnmodeled(path, layer.OtherAttrs, layer.Other); err != nil {
			return err
		}
		wLayer := &wxx.ExtraTerrainLayer_t{Name: layer.Name}
		for j, tl := range layer.Terrain {
			tlPath := path + "/terrainAndLocation"
			if err := refuseUnmodeled(tlPath, tl.OtherAttrs, tl.Other); err != nil {
				return err
			}
			resources, clamped, err := decodeExtraTerrainResources(tl.Resources)
			if err != nil {
				return fmt.Errorf("%s/@resources: %w", tlPath, err)
			}
			for _, c := range clamped {
				c.Path = fmt.Sprintf("%s[%d]/@resources", tlPath, j)
				c.Field = fmt.Sprintf("Map_t.ExtraTerrain.MapLayers[%d].Terrain[%d].Resources.%s", i, j, c.Field)
				*clamps = append(*clamps, c)
			}
			x, y, err := decodeLocation(tl.Location)
			if err != nil {
				return fmt.Errorf("%s/@location: %w", tlPath, err)
			}
			wLayer.Terrain = append(wLayer.Terrain, &wxx.TerrainAndLocation_t{
				Terrain:   tl.Name,
				Elevation: float64(tl.Elevation),
				IsIcy:     tl.Icy,
				IsGMOnly:  tl.GmOnly,
				Resources: resources,
				X:         x,
				Y:         y,
			})
		}
		w.ExtraTerrain.MapLayers = append(w.ExtraTerrain.MapLayers, wLayer)
	}
	return nil
}

// refuseUnmodeled reports an error naming the first unmodeled attribute or
// child element found at path.
func refuseUnmodeled(path string, attrs []xml.Attr, elements []unmodeledElement_t) error {
	if len(attrs) != 0 {
		return errors.Join(wxx.ErrInvalidXML, fmt.Errorf("%s: unmodeled attribute @%s=%q; wxx does not understand it and will not drop it silently (issue #34)",
			path, attrs[0].Name.Local, attrs[0].Value))
	}
	if len(elements) != 0 {
		return errors.Join(wxx.ErrInvalidXML, fmt.Errorf("%s: unmodeled child element <%s>; wxx does not understand it and will not drop it silently (issue #34)",
			path, elements[0].XMLName.Local))
	}
	return nil
}

// Clamp_t is one value the decoder changed to bring it into the range the app
// keeps (issue #124). xmlio reports each one to the caller as a
// ClampedValue_t; the codec only records it.
//
// Path is the on-disk location, Field the Map_t field now holding Now, and Was
// the value the file stated.
type Clamp_t struct {
	Path  string
	Field string
	Was   int
	Now   int
}

// resourceNames are the seven resources in the order this decoder reads a
// <terrainAndLocation> @resources value. The order is the tile record's,
// borrowed and NOT verified for this attribute (issue #124): a Clamp_t's Field
// names the resource by this order, so it is only as right as the order is.
var resourceNames = [7]string{"Animal", "Brick", "Crops", "Gems", "Lumber", "Metals", "Rock"}

// decodeExtraTerrainResources parses a <terrainAndLocation> @resources value:
// "Z", or seven comma-separated integers in the order animal, brick, crops,
// gems, lumber, metals, rock.
//
// This is not the <tilerow> spelling. A tile record is tab-separated and writes
// its animal value BEFORE the Z. Here "Z" stands alone, so it is read as all
// seven zero; that reading is inferred from the samples, which carry only "Z"
// and fully-populated values. The field order is also inferred (#124).
//
// A value outside 0..100 is CLAMPED into it (0 below, 100 above) and returned
// in clamped, rather than refused. THIS IS LOSSY: the map no longer holds what
// the file said, and an encode writes the clamped value. It is the maintainer's
// ruling on #124, and the reasoning is:
//
//   - The app does not keep such a value either. Worldographer 2.08 reads these
//     fields with Byte.parseByte, so 128 and above stop the file opening (app
//     check, #124: 150 in the second field). In a tile record it clamps
//     101..127 to 100 on load and saves 100 (app check, #122); clamping here
//     ASSUMES the same, which is untested for this attribute.
//   - Refusing would make wxx unable to read a map that the app may well have
//     opened. Clamping reads it as the app is assumed to, and says so.
//   - Negative values are clamped to 0. The app was not tried with one; a
//     negative resource has no meaning, and 0 is the nearest the app keeps.
//
// Clamping happens only here. A tile record outside 0..100 is still refused
// on decode (tiles.go), and Map_t.Validate refuses either before an encode
// (issues #122, #124), so wxx never WRITES an out-of-range resource. A value
// that is not an integer is refused, not clamped: there is nothing to clamp.
//
// Each Clamp_t comes back with Field set to the resource's name only; the
// caller fills in Path and the full Field, which it knows and this does not.
func decodeExtraTerrainResources(s string) (wxx.Resources_t, []Clamp_t, error) {
	if s == "Z" {
		return wxx.Resources_t{}, nil, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) != 7 {
		return wxx.Resources_t{}, nil, fmt.Errorf("%q: want \"Z\" or seven comma-separated integers", s)
	}
	var v [7]int
	var clamped []Clamp_t
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return wxx.Resources_t{}, nil, fmt.Errorf("%q: value %d: %w", s, i+1, err)
		}
		v[i] = min(max(n, 0), 100)
		if v[i] != n {
			clamped = append(clamped, Clamp_t{Field: resourceNames[i], Was: n, Now: v[i]})
		}
	}
	return wxx.Resources_t{Animal: v[0], Brick: v[1], Crops: v[2], Gems: v[3], Lumber: v[4], Metals: v[5], Rock: v[6]}, clamped, nil
}

// decodeLocation parses a <terrainAndLocation> @location value, "x,y".
func decodeLocation(s string) (x, y float64, err error) {
	xs, ys, ok := strings.Cut(s, ",")
	if !ok {
		return 0, 0, fmt.Errorf("%q: want \"x,y\"", s)
	}
	if x, err = strconv.ParseFloat(xs, 64); err != nil {
		return 0, 0, fmt.Errorf("%q: x: %w", s, err)
	}
	if y, err = strconv.ParseFloat(ys, 64); err != nil {
		return 0, 0, fmt.Errorf("%q: y: %w", s, err)
	}
	return x, y, nil
}

// encodeExtraTerrain emits <extraTerrain> only when present (non-nil), in the
// layout Worldographer writes: a newline after the opening tag, each <mapLayer>
// indented one tab and each <terrainAndLocation> two, the latter closed with
// " />". Every tracked sample is laid out this way, and an empty container is
// "<extraTerrain>\n</extraTerrain>".
//
// Elevation is an integer on disk, like a tile's, and goes through toInt: a
// value with a fraction is refused rather than rounded (issue #64).
func encodeExtraTerrain(extraTerrain *wxx.ExtraTerrain_t, wb *bytes.Buffer) error {
	if extraTerrain == nil {
		return nil
	}
	wb.WriteString("<extraTerrain>\n")
	for _, layer := range extraTerrain.MapLayers {
		wb.WriteString(fmt.Sprintf("\t<mapLayer name=%s>\n", xmlAttr(layer.Name)))
		for _, tl := range layer.Terrain {
			elevation, err := toInt("map/extraTerrain/mapLayer/terrainAndLocation/@elevation", tl.Elevation)
			if err != nil {
				return err
			}
			wb.WriteString(fmt.Sprintf("\t\t<terrainAndLocation name=%s elevation=%s icy=%s gmOnly=%s resources=%s location=%s />\n",
				xmlAttr(tl.Terrain),
				xmlAttr(elevation.String()),
				xmlAttr(bools(tl.IsIcy)),
				xmlAttr(bools(tl.IsGMOnly)),
				xmlAttr(encodeExtraTerrainResources(tl.Resources)),
				xmlAttr(floats(tl.X)+","+floats(tl.Y))))
		}
		wb.WriteString("\t</mapLayer>\n")
	}
	wb.WriteString("</extraTerrain>\n")
	return nil
}

// encodeExtraTerrainResources is the inverse of decodeExtraTerrainResources:
// "Z" when all seven are zero, otherwise all seven, comma-separated.
func encodeExtraTerrainResources(r wxx.Resources_t) string {
	if r == (wxx.Resources_t{}) {
		return "Z"
	}
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d", r.Animal, r.Brick, r.Crops, r.Gems, r.Lumber, r.Metals, r.Rock)
}
