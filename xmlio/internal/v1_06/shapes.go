// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// decodeShapes copies each <shape> (with its <p> points) into the domain map.
//
// Each point's coordinates are parsed from the spelling the file uses, and
// whether both were spelled as integers is kept in Point_t.IntegerXY so the
// encoder can write them back the same way (issue #94).
func decodeShapes(src Shapes_t, w *wxx.Map_t) error {
	for i, shape := range src.Shapes {
		wShape := &wxx.Shape_t{
			BbHeight:              shape.BbHeight,
			BbIterations:          shape.BbIterations,
			BbWidth:               shape.BbWidth,
			CreationType:          shape.CreationType,
			CurrentShapeViewLevel: shape.CurrentShapeViewLevel,
			DsColor:               shape.DsColor,
			DsOffsetX:             shape.DsOffsetX,
			DsOffsetY:             shape.DsOffsetY,
			DsRadius:              shape.DsRadius,
			DsSpread:              shape.DsSpread,
			ExtraLineDistance:     shape.ExtraLineDistance,
			ExtraLineLength:       shape.ExtraLineLength,
			ExtraLineWidth:        shape.ExtraLineWidth,
			ExtraLineSeparation:   shape.ExtraLineSeparation,
			FillRule:              shape.FillRule,
			FillTexture:           shape.FillTexture,
			HighestViewLevel:      shape.HighestViewLevel,
			InsChoke:              shape.InsChoke,
			InsColor:              shape.InsColor,
			InsOffsetX:            shape.InsOffsetX,
			InsOffsetY:            shape.InsOffsetY,
			InsRadius:             shape.InsRadius,
			IsBoxBlur:             shape.IsBoxBlur,
			IsContinent:           shape.IsContinent,
			IsCurve:               shape.IsCurve,
			IsDropShadow:          shape.IsDropShadow,
			IsGMOnly:              shape.IsGMOnly,
			IsInnerShadow:         shape.IsInnerShadow,
			IsKingdom:             shape.IsKingdom,
			IsMatchTileBorders:    shape.IsMatchTileBorders,
			IsProvince:            shape.IsProvince,
			IsSnapVertices:        shape.IsSnapVertices,
			IsWorld:               shape.IsWorld,
			LineCap:               shape.LineCap,
			LineJoin:              shape.LineJoin,
			MapLayer:              shape.MapLayer,
			Opacity:               shape.Opacity,
			StrokeColor:           shape.StrokeColor,
			StrokeTexture:         shape.StrokeTexture,
			StrokeType:            shape.StrokeType,
			StrokeWidth:           shape.StrokeWidth,
			Tags:                  shape.Tags,
			Type:                  shape.Type,
		}

		for j, point := range shape.Points {
			path := fmt.Sprintf("map/shapes/shape[%d]/p[%d]", i+1, j+1)
			x, err := parseCoordinate(path+"/@x", point.X)
			if err != nil {
				return err
			}
			y, err := parseCoordinate(path+"/@y", point.Y)
			if err != nil {
				return err
			}
			wPoint := &wxx.Point_t{
				Type:      point.Type,
				X:         x,
				Y:         y,
				IntegerXY: integerSpelled(point.X) && integerSpelled(point.Y),
			}
			// The four control-point attributes come as a set. A point with
			// some but not all of them is refused, because the model holds
			// the set or nothing and would otherwise drop or invent values.
			switch n := countSet(point.CX1, point.CY1, point.CX2, point.CY2); n {
			case 0:
			case 4:
				wPoint.Control = &wxx.CurveControl_t{CX1: *point.CX1, CY1: *point.CY1, CX2: *point.CX2, CY2: *point.CY2}
			default:
				return fmt.Errorf("%s: states %d of @cx1 @cy1 @cx2 @cy2; a curve point states all four or none", path, n)
			}
			wShape.Points = append(wShape.Points, wPoint)
		}

		w.Shapes = append(w.Shapes, wShape)
	}
	return nil
}

// countSet counts the non-nil pointers.
func countSet(ps ...*float64) int {
	n := 0
	for _, p := range ps {
		if p != nil {
			n++
		}
	}
	return n
}

// parseCoordinate parses a <p> coordinate. Before issue #94 encoding/xml
// parsed these into float64 directly; this keeps that behaviour, including
// its treatment of surrounding space, and names the attribute on failure.
func parseCoordinate(path, s string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("%s = %q: %w", path, s, err)
	}
	return f, nil
}

// integerSpelled reports whether a coordinate is spelled as an integer: an
// optional sign and digits, with no decimal point and no exponent ("2700",
// not "2700.0" or "2.7e3").
func integerSpelled(s string) bool {
	s = strings.TrimSpace(s)
	if s != "" && (s[0] == '-' || s[0] == '+') {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// integral reports whether f is a whole number an integer spelling can state.
func integral(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0) && f == math.Trunc(f)
}

// coordinate spells a <p> coordinate: as an integer ("2700") when the point
// was spelled that way, and as every other float in the file otherwise
// ("1950.0"). encodeShape has already refused a fractional value under
// integer spelling.
func coordinate(f float64, integerXY bool) string {
	if integerXY {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return floats(f)
}

func encodeShapes(shapes []*wxx.Shape_t, wb *bytes.Buffer) error {
	wb.WriteString("<shapes>\n")
	for i, shape := range shapes {
		if err := encodeShape(i, shape, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</shapes>\n")
	return nil
}

// encodeShape writes one <shape> in the attribute order and layout
// Worldographer 2.07 writes (issue #94): two spaces before @type, and each
// <p> on its own line with one leading space. 2.06 and 2.08 write the same
// order. The one layout difference left is that the app spells a path
// point's y as `y = "..."`, with spaces around the '='; that is the same XML,
// and this writes `y="..."`.
//
// i is the shape's index in the map, used only to name a refused point.
func encodeShape(i int, shape *wxx.Shape_t, wb *bytes.Buffer) error {
	// The three colors are strings written verbatim, so a zero value or a typo
	// would reach the file; they are checked before anything is written (#83).
	dsColor, err := rgbaAttr("map/shapes/shape/@dsColor", shape.DsColor, true)
	if err != nil {
		return err
	}
	insColor, err := rgbaAttr("map/shapes/shape/@insColor", shape.InsColor, true)
	if err != nil {
		return err
	}
	// strokeColor "" means the shape has none, and the attribute is omitted:
	// Worldographer 2.08 reads strokeColor="null" as no stroke and saves such a
	// shape with no strokeColor at all, which decodes to "" (#83).
	strokeColor := shape.StrokeColor
	if strokeColor != "" {
		if strokeColor, err = rgbaAttr("map/shapes/shape/@strokeColor", shape.StrokeColor, true); err != nil {
			return err
		}
	}
	// A point marked as integer-spelled must hold whole numbers: writing
	// 2700.5 as an integer would move it, and writing it with a decimal point
	// would silently change the spelling the caller asked for (issue #94).
	for j, p := range shape.Points {
		if !p.IntegerXY {
			continue
		}
		for _, c := range []struct {
			name string
			v    float64
		}{{"x", p.X}, {"y", p.Y}} {
			if !integral(c.v) {
				return errors.Join(wxx.ErrInvalidIntegerAttribute, fmt.Errorf(
					"map/shapes/shape[%d]/p[%d]/@%s: %s is not a whole number, but the point's IntegerXY says both coordinates are spelled as integers (issue #94)",
					i+1, j+1, c.name, strconv.FormatFloat(c.v, 'g', -1, 64)))
			}
		}
	}
	wb.WriteString("<shape ")
	wb.WriteString(fmt.Sprintf(" type=%s", xmlAttr(shape.Type)))
	wb.WriteString(fmt.Sprintf(" isCurve=%s", xmlAttr(bools(shape.IsCurve))))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%s", xmlAttr(bools(shape.IsGMOnly))))
	wb.WriteString(fmt.Sprintf(" isSnapVertices=%s", xmlAttr(bools(shape.IsSnapVertices))))
	wb.WriteString(fmt.Sprintf(" isMatchTileBorders=%s", xmlAttr(bools(shape.IsMatchTileBorders))))
	wb.WriteString(fmt.Sprintf(" tags=%s", xmlAttr(shape.Tags)))
	wb.WriteString(fmt.Sprintf(" creationType=%s", xmlAttr(shape.CreationType)))
	wb.WriteString(fmt.Sprintf(" isDropShadow=%s", xmlAttr(bools(shape.IsDropShadow))))
	wb.WriteString(fmt.Sprintf(" isInnerShadow=%s", xmlAttr(bools(shape.IsInnerShadow))))
	wb.WriteString(fmt.Sprintf(" isBoxBlur=%s", xmlAttr(bools(shape.IsBoxBlur))))
	// The four @extraLine* are written together or not at all. A shape
	// decoded from a classic file had all four at zero (classic was removed in
	// issue #103), and this encoder wrote none of them before #94; whether Worldographer reads extraLineWidth="0.0"
	// the same as no attribute has not been tested in the app, so all-zero
	// keeps the old output.
	if shape.ExtraLineDistance != 0 || shape.ExtraLineLength != 0 || shape.ExtraLineWidth != 0 || shape.ExtraLineSeparation != 0 {
		wb.WriteString(fmt.Sprintf(" extraLineDistance=%s", xmlAttr(floats(shape.ExtraLineDistance))))
		wb.WriteString(fmt.Sprintf(" extraLineLength=%s", xmlAttr(floats(shape.ExtraLineLength))))
		wb.WriteString(fmt.Sprintf(" extraLineWidth=%s", xmlAttr(floats(shape.ExtraLineWidth))))
		wb.WriteString(fmt.Sprintf(" extraLineSeparation=%s", xmlAttr(floats(shape.ExtraLineSeparation))))
	}
	wb.WriteString(fmt.Sprintf(" isWorld=%s", xmlAttr(bools(shape.IsWorld))))
	wb.WriteString(fmt.Sprintf(" isContinent=%s", xmlAttr(bools(shape.IsContinent))))
	wb.WriteString(fmt.Sprintf(" isKingdom=%s", xmlAttr(bools(shape.IsKingdom))))
	wb.WriteString(fmt.Sprintf(" isProvince=%s", xmlAttr(bools(shape.IsProvince))))
	wb.WriteString(fmt.Sprintf(" dsSpread=%s", xmlAttr(floats(shape.DsSpread))))
	wb.WriteString(fmt.Sprintf(" dsRadius=%s", xmlAttr(floats(shape.DsRadius))))
	wb.WriteString(fmt.Sprintf(" dsOffsetX=%s", xmlAttr(floats(shape.DsOffsetX))))
	wb.WriteString(fmt.Sprintf(" dsOffsetY=%s", xmlAttr(floats(shape.DsOffsetY))))
	wb.WriteString(fmt.Sprintf(" insChoke=%s", xmlAttr(floats(shape.InsChoke))))
	wb.WriteString(fmt.Sprintf(" insRadius=%s", xmlAttr(floats(shape.InsRadius))))
	wb.WriteString(fmt.Sprintf(" insOffsetX=%s", xmlAttr(floats(shape.InsOffsetX))))
	wb.WriteString(fmt.Sprintf(" insOffsetY=%s", xmlAttr(floats(shape.InsOffsetY))))
	wb.WriteString(fmt.Sprintf(" bbWidth=%s", xmlAttr(floats(shape.BbWidth))))
	wb.WriteString(fmt.Sprintf(" bbHeight=%s", xmlAttr(floats(shape.BbHeight))))
	wb.WriteString(fmt.Sprintf(" bbIterations=%s", xmlAttr(ints(shape.BbIterations))))
	wb.WriteString(fmt.Sprintf(" mapLayer=%s", xmlAttr(shape.MapLayer)))
	wb.WriteString(fmt.Sprintf(" fillTexture=%s", xmlAttr(shape.FillTexture)))
	wb.WriteString(fmt.Sprintf(" strokeTexture=%s", xmlAttr(shape.StrokeTexture)))
	wb.WriteString(fmt.Sprintf(" strokeType=%s", xmlAttr(shape.StrokeType)))
	wb.WriteString(fmt.Sprintf(" highestViewLevel=%s", xmlAttr(shape.HighestViewLevel)))
	wb.WriteString(fmt.Sprintf(" currentShapeViewLevel=%s", xmlAttr(shape.CurrentShapeViewLevel)))
	wb.WriteString(fmt.Sprintf(" lineCap=%s", xmlAttr(shape.LineCap)))
	wb.WriteString(fmt.Sprintf(" lineJoin=%s", xmlAttr(shape.LineJoin)))
	wb.WriteString(fmt.Sprintf(" opacity=%s", xmlAttr(floats(shape.Opacity))))
	// @fillRule is written only when the shape states one. The samples'
	// tile-border polygons state none, and no 2.06, 2.07 or 2.08 sample writes
	// fillRule="" (issue #94).
	if shape.FillRule != "" {
		wb.WriteString(fmt.Sprintf(" fillRule=%s", xmlAttr(shape.FillRule)))
	}
	if strokeColor != "" {
		wb.WriteString(fmt.Sprintf(" strokeColor=%s", xmlAttr(strokeColor)))
	}
	wb.WriteString(fmt.Sprintf(" strokeWidth=%s", xmlAttr(floats(shape.StrokeWidth))))
	wb.WriteString(fmt.Sprintf(" dsColor=%s", xmlAttr(dsColor)))
	wb.WriteString(fmt.Sprintf(" insColor=%s", xmlAttr(insColor)))
	wb.WriteString(">\n")
	for _, p := range shape.Points {
		wb.WriteString(" <p")
		// Like @fillRule, @type is written only when stated: the samples put
		// it on a path's first point and on no other point.
		if p.Type != "" {
			wb.WriteString(fmt.Sprintf(" type=%s", xmlAttr(p.Type)))
		}
		wb.WriteString(fmt.Sprintf(" x=%s", xmlAttr(coordinate(p.X, p.IntegerXY))))
		wb.WriteString(fmt.Sprintf(" y=%s", xmlAttr(coordinate(p.Y, p.IntegerXY))))
		// Written only when decoded (or set): every sample spells these with
		// a decimal point, so they have no integer spelling to keep.
		if c := p.Control; c != nil {
			wb.WriteString(fmt.Sprintf(" cx1=%s", xmlAttr(floats(c.CX1))))
			wb.WriteString(fmt.Sprintf(" cy1=%s", xmlAttr(floats(c.CY1))))
			wb.WriteString(fmt.Sprintf(" cx2=%s", xmlAttr(floats(c.CX2))))
			wb.WriteString(fmt.Sprintf(" cy2=%s", xmlAttr(floats(c.CY2))))
		}
		wb.WriteString("/>\n")
	}
	wb.WriteString("</shape>\n")
	return nil
}
