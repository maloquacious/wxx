// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeMapKey copies the <mapkey> attributes into the domain map. Colors are
// folded through decodeRgba. It is invoked from decodeTiles (inside the tilerow
// loop) to preserve the original ordering, in which <mapkey> was decoded per
// tilerow after the tiles were parsed.
func decodeMapKey(src MapKey_t, w *wxx.Map_t) error {
	var err error
	w.MapKey = &wxx.MapKey_t{
		PositionX: src.PositionX,
		PositionY: src.PositionY,
		Viewlevel: src.Viewlevel,
		Height:    float64(src.Height),
	}
	if w.MapKey.BackgroundColor, err = decodeRgba(src.BackgroundColor); err != nil {
		return fmt.Errorf("mapkey.backgroundcolor: %w", err)
	}
	w.MapKey.BackgroundOpacity = float64(src.BackgroundOpacity)
	w.MapKey.TitleText = src.TitleText
	w.MapKey.TitleFontFace = src.TitleFontFace
	if w.MapKey.TitleFontColor, err = decodeRgba(src.TitleFontColor); err != nil {
		return fmt.Errorf("mapkey.titleFontColor: %w", err)
	}
	w.MapKey.TitleFontBold = src.TitleFontBold
	w.MapKey.TitleFontItalic = src.TitleFontItalic
	w.MapKey.TitleScale = float64(src.TitleScale)
	w.MapKey.ScaleText = src.ScaleText
	w.MapKey.ScaleFontFace = src.ScaleFontFace
	if w.MapKey.ScaleFontColor, err = decodeRgba(src.ScaleFontColor); err != nil {
		return fmt.Errorf("mapkey.scaleFontColor: %w", err)
	}
	w.MapKey.ScaleFontBold = src.ScaleFontBold
	w.MapKey.ScaleFontItalic = src.ScaleFontItalic
	w.MapKey.ScaleScale = float64(src.ScaleScale)
	w.MapKey.EntryFontFace = src.EntryFontFace
	if w.MapKey.EntryFontColor, err = decodeRgba(src.EntryFontColor); err != nil {
		return fmt.Errorf("mapkey.entryFontColor: %w", err)
	}
	w.MapKey.EntryFontBold = src.EntryFontBold
	w.MapKey.EntryFontItalic = src.EntryFontItalic
	w.MapKey.EntryScale = float64(src.EntryScale)
	return nil
}

// encodeMapKey writes the <mapkey> element.
//
// Five of its attributes are integers this schema states without a decimal
// point, and Worldographer refuses to open a file that spells them otherwise
// (issue #64; see Int_t). They are resolved BEFORE anything is written, so a
// map that cannot be expressed produces no document rather than a truncated
// one -- the same rule the codec follows for an unwritable orientation.
func encodeMapKey(mapKey *wxx.MapKey_t, wb *bytes.Buffer) error {
	height, err := toInt("map/mapkey/@height", mapKey.Height)
	if err != nil {
		return err
	}
	backgroundOpacity, err := toInt("map/mapkey/@backgroundopacity", mapKey.BackgroundOpacity)
	if err != nil {
		return err
	}
	titleScale, err := toInt("map/mapkey/@titleScale", mapKey.TitleScale)
	if err != nil {
		return err
	}
	scaleScale, err := toInt("map/mapkey/@scaleScale", mapKey.ScaleScale)
	if err != nil {
		return err
	}
	entryScale, err := toInt("map/mapkey/@entryScale", mapKey.EntryScale)
	if err != nil {
		return err
	}

	wb.WriteString("<mapkey")
	wb.WriteString(fmt.Sprintf(" positionx=%s", xmlAttr(floats(mapKey.PositionX))))
	wb.WriteString(fmt.Sprintf(" positiony=%s", xmlAttr(floats(mapKey.PositionY))))
	wb.WriteString(fmt.Sprintf(" viewlevel=%s", xmlAttr(mapKey.Viewlevel)))
	wb.WriteString(fmt.Sprintf(" height=%s", xmlAttr(height.String())))
	wb.WriteString(fmt.Sprintf(" backgroundcolor=%s", xmlAttr(rgbas(mapKey.BackgroundColor)))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" backgroundopacity=%s", xmlAttr(backgroundOpacity.String())))
	wb.WriteString(fmt.Sprintf(" titleText=%s", xmlAttr(mapKey.TitleText)))
	wb.WriteString(fmt.Sprintf(" titleFontFace=%s", xmlAttr(mapKey.TitleFontFace)))
	wb.WriteString(fmt.Sprintf(" titleFontColor=%s", xmlAttr(rgbas(mapKey.TitleFontColor)))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" titleFontBold=%s", xmlAttr(bools(mapKey.TitleFontBold))))
	wb.WriteString(fmt.Sprintf(" titleFontItalic=%s", xmlAttr(bools(mapKey.TitleFontItalic))))
	wb.WriteString(fmt.Sprintf(" titleScale=%s", xmlAttr(titleScale.String())))
	wb.WriteString(fmt.Sprintf(" scaleText=%s", xmlAttr(mapKey.ScaleText)))
	wb.WriteString(fmt.Sprintf(" scaleFontFace=%s", xmlAttr(mapKey.ScaleFontFace)))
	wb.WriteString(fmt.Sprintf(" scaleFontColor=%s", xmlAttr(rgbas(mapKey.ScaleFontColor)))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" scaleFontBold=%s", xmlAttr(bools(mapKey.ScaleFontBold))))
	wb.WriteString(fmt.Sprintf(" scaleFontItalic=%s", xmlAttr(bools(mapKey.ScaleFontItalic))))
	wb.WriteString(fmt.Sprintf(" scaleScale=%s", xmlAttr(scaleScale.String())))
	wb.WriteString(fmt.Sprintf(" entryFontFace=%s", xmlAttr(mapKey.EntryFontFace)))
	wb.WriteString(fmt.Sprintf(" entryFontColor=%s", xmlAttr(rgbas(mapKey.EntryFontColor)))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" entryFontBold=%s", xmlAttr(bools(mapKey.EntryFontBold))))
	wb.WriteString(fmt.Sprintf(" entryFontItalic=%s", xmlAttr(bools(mapKey.EntryFontItalic))))
	wb.WriteString(fmt.Sprintf(" entryScale=%s", xmlAttr(entryScale.String())))
	wb.WriteString(">\n")
	wb.WriteString("</mapkey>\n")
	return nil
}
