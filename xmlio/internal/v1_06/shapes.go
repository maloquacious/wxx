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
	wb.WriteString(fmt.Sprintf(" type=%s", xmlAttr(shape.Type)))
	wb.WriteString(fmt.Sprintf(" creationType=%s", xmlAttr(shape.CreationType)))
	wb.WriteString(fmt.Sprintf(" isWorld=%s", xmlAttr(bools(shape.IsWorld))))
	wb.WriteString(fmt.Sprintf(" isContinent=%s", xmlAttr(bools(shape.IsContinent))))
	wb.WriteString(fmt.Sprintf(" isKingdom=%s", xmlAttr(bools(shape.IsKingdom))))
	wb.WriteString(fmt.Sprintf(" isProvince=%s", xmlAttr(bools(shape.IsProvince))))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%s", xmlAttr(bools(shape.IsGMOnly))))
	wb.WriteString(fmt.Sprintf(" isCurve=%s", xmlAttr(bools(shape.IsCurve))))
	wb.WriteString(fmt.Sprintf(" isSnapVertices=%s", xmlAttr(bools(shape.IsSnapVertices))))
	wb.WriteString(fmt.Sprintf(" isMatchTileBorders=%s", xmlAttr(bools(shape.IsMatchTileBorders))))
	wb.WriteString(fmt.Sprintf(" isBoxBlur=%s", xmlAttr(bools(shape.IsBoxBlur))))
	wb.WriteString(fmt.Sprintf(" isDropShadow=%s", xmlAttr(bools(shape.IsDropShadow))))
	wb.WriteString(fmt.Sprintf(" isInnerShadow=%s", xmlAttr(bools(shape.IsInnerShadow))))
	wb.WriteString(fmt.Sprintf(" dsOffsetX=%s", xmlAttr(floats(shape.DsOffsetX))))
	wb.WriteString(fmt.Sprintf(" dsOffsetY=%s", xmlAttr(floats(shape.DsOffsetY))))
	wb.WriteString(fmt.Sprintf(" dsRadius=%s", xmlAttr(floats(shape.DsRadius))))
	wb.WriteString(fmt.Sprintf(" dsSpread=%s", xmlAttr(floats(shape.DsSpread))))
	wb.WriteString(fmt.Sprintf(" dsColor=%s", xmlAttr(shape.DsColor)))
	wb.WriteString(fmt.Sprintf(" insOffsetX=%s", xmlAttr(floats(shape.InsOffsetX))))
	wb.WriteString(fmt.Sprintf(" insOffsetY=%s", xmlAttr(floats(shape.InsOffsetY))))
	wb.WriteString(fmt.Sprintf(" insRadius=%s", xmlAttr(floats(shape.InsRadius))))
	wb.WriteString(fmt.Sprintf(" insChoke=%s", xmlAttr(floats(shape.InsChoke))))
	wb.WriteString(fmt.Sprintf(" insColor=%s", xmlAttr(shape.InsColor)))
	wb.WriteString(fmt.Sprintf(" bbWidth=%s", xmlAttr(floats(shape.BbWidth))))
	wb.WriteString(fmt.Sprintf(" bbHeight=%s", xmlAttr(floats(shape.BbHeight))))
	wb.WriteString(fmt.Sprintf(" bbIterations=%s", xmlAttr(ints(shape.BbIterations))))
	wb.WriteString(fmt.Sprintf(" mapLayer=%s", xmlAttr(shape.MapLayer)))
	wb.WriteString(fmt.Sprintf(" fillRule=%s", xmlAttr(shape.FillRule)))
	wb.WriteString(fmt.Sprintf(" fillTexture=%s", xmlAttr(shape.FillTexture)))
	wb.WriteString(fmt.Sprintf(" strokeTexture=%s", xmlAttr(shape.StrokeTexture)))
	wb.WriteString(fmt.Sprintf(" strokeType=%s", xmlAttr(shape.StrokeType)))
	wb.WriteString(fmt.Sprintf(" highestViewLevel=%s", xmlAttr(shape.HighestViewLevel)))
	wb.WriteString(fmt.Sprintf(" currentShapeViewLevel=%s", xmlAttr(shape.CurrentShapeViewLevel)))
	wb.WriteString(fmt.Sprintf(" lineCap=%s", xmlAttr(shape.LineCap)))
	wb.WriteString(fmt.Sprintf(" lineJoin=%s", xmlAttr(shape.LineJoin)))
	wb.WriteString(fmt.Sprintf(" opacity=%s", xmlAttr(floats(shape.Opacity))))
	wb.WriteString(fmt.Sprintf(" strokeColor=%s", xmlAttr(shape.StrokeColor)))
	wb.WriteString(fmt.Sprintf(" strokeWidth=%s", xmlAttr(floats(shape.StrokeWidth))))
	wb.WriteString(fmt.Sprintf(" tags=%s", xmlAttr(shape.Tags)))
	wb.WriteString(">\n")
	for _, p := range shape.Points {
		wb.WriteString("<p")
		wb.WriteString(fmt.Sprintf(" type=%s", xmlAttr(p.Type)))
		wb.WriteString(fmt.Sprintf(" x=%s", xmlAttr(floats(p.X))))
		wb.WriteString(fmt.Sprintf(" y=%s", xmlAttr(floats(p.Y))))
		wb.WriteString("/>\n")
	}
	wb.WriteString("</shape>\n")
	return nil
}
