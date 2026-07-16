// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeShapes copies each <shape> (with its <p> points) into the domain map.
func decodeShapes(src Shapes_t, w *wxx.Map_t) {
	for _, shape := range src.Shapes {
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

		for _, point := range shape.Points {
			wPoint := &wxx.Point_t{
				Type: point.Type,
				X:    point.X,
				Y:    point.Y,
			}
			wShape.Points = append(wShape.Points, wPoint)
		}

		w.Shapes = append(w.Shapes, wShape)
	}
}

func encodeShapes(shapes []*wxx.Shape_t, wb *bytes.Buffer) error {
	wb.WriteString("<shapes>\n")
	for _, shape := range shapes {
		if err := encodeShape(shape, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</shapes>\n")
	return nil
}

func encodeShape(shape *wxx.Shape_t, wb *bytes.Buffer) error {
	wb.WriteString("<shape")
	wb.WriteString(fmt.Sprintf(" type=%q", shape.Type))
	wb.WriteString(fmt.Sprintf(" creationType=%q", shape.CreationType))
	wb.WriteString(fmt.Sprintf(" isWorld=%q", bools(shape.IsWorld)))
	wb.WriteString(fmt.Sprintf(" isContinent=%q", bools(shape.IsContinent)))
	wb.WriteString(fmt.Sprintf(" isKingdom=%q", bools(shape.IsKingdom)))
	wb.WriteString(fmt.Sprintf(" isProvince=%q", bools(shape.IsProvince)))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(shape.IsGMOnly)))
	wb.WriteString(fmt.Sprintf(" isCurve=%q", bools(shape.IsCurve)))
	wb.WriteString(fmt.Sprintf(" isSnapVertices=%q", bools(shape.IsSnapVertices)))
	wb.WriteString(fmt.Sprintf(" isMatchTileBorders=%q", bools(shape.IsMatchTileBorders)))
	wb.WriteString(fmt.Sprintf(" isBoxBlur=%q", bools(shape.IsBoxBlur)))
	wb.WriteString(fmt.Sprintf(" isDropShadow=%q", bools(shape.IsDropShadow)))
	wb.WriteString(fmt.Sprintf(" isInnerShadow=%q", bools(shape.IsInnerShadow)))
	wb.WriteString(fmt.Sprintf(" dsOffsetX=%q", floats(shape.DsOffsetX)))
	wb.WriteString(fmt.Sprintf(" dsOffsetY=%q", floats(shape.DsOffsetY)))
	wb.WriteString(fmt.Sprintf(" dsRadius=%q", floats(shape.DsRadius)))
	wb.WriteString(fmt.Sprintf(" dsSpread=%q", floats(shape.DsSpread)))
	wb.WriteString(fmt.Sprintf(" dsColor=%q", shape.DsColor))
	wb.WriteString(fmt.Sprintf(" insOffsetX=%q", floats(shape.InsOffsetX)))
	wb.WriteString(fmt.Sprintf(" insOffsetY=%q", floats(shape.InsOffsetY)))
	wb.WriteString(fmt.Sprintf(" insRadius=%q", floats(shape.InsRadius)))
	wb.WriteString(fmt.Sprintf(" insChoke=%q", floats(shape.InsChoke)))
	wb.WriteString(fmt.Sprintf(" insColor=%q", shape.InsColor))
	wb.WriteString(fmt.Sprintf(" bbWidth=%q", floats(shape.BbWidth)))
	wb.WriteString(fmt.Sprintf(" bbHeight=%q", floats(shape.BbHeight)))
	wb.WriteString(fmt.Sprintf(" bbIterations=%q", ints(shape.BbIterations)))
	wb.WriteString(fmt.Sprintf(" mapLayer=%q", shape.MapLayer))
	wb.WriteString(fmt.Sprintf(" fillRule=%q", shape.FillRule))
	wb.WriteString(fmt.Sprintf(" fillTexture=%q", shape.FillTexture))
	wb.WriteString(fmt.Sprintf(" strokeTexture=%q", shape.StrokeTexture))
	wb.WriteString(fmt.Sprintf(" strokeType=%q", shape.StrokeType))
	wb.WriteString(fmt.Sprintf(" highestViewLevel=%q", shape.HighestViewLevel))
	wb.WriteString(fmt.Sprintf(" currentShapeViewLevel=%q", shape.CurrentShapeViewLevel))
	wb.WriteString(fmt.Sprintf(" lineCap=%q", shape.LineCap))
	wb.WriteString(fmt.Sprintf(" lineJoin=%q", shape.LineJoin))
	wb.WriteString(fmt.Sprintf(" opacity=%q", floats(shape.Opacity)))
	wb.WriteString(fmt.Sprintf(" strokeColor=%q", shape.StrokeColor))
	wb.WriteString(fmt.Sprintf(" strokeWidth=%q", floats(shape.StrokeWidth)))
	wb.WriteString(fmt.Sprintf(" tags=%q", shape.Tags))
	wb.WriteString(">\n")
	for _, p := range shape.Points {
		wb.WriteString("<p")
		wb.WriteString(fmt.Sprintf(" type=%q", p.Type))
		wb.WriteString(fmt.Sprintf(" x=%q", floats(p.X)))
		wb.WriteString(fmt.Sprintf(" y=%q", floats(p.Y)))
		wb.WriteString("/>\n")
	}
	wb.WriteString("</shape>\n")
	return nil
}
