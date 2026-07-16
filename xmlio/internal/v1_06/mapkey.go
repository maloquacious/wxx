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
		Height:    src.Height,
	}
	if w.MapKey.BackgroundColor, err = decodeRgba(src.BackgroundColor); err != nil {
		return fmt.Errorf("mapkey.backgroundcolor: %w", err)
	}
	w.MapKey.BackgroundOpacity = src.BackgroundOpacity
	w.MapKey.TitleText = src.TitleText
	w.MapKey.TitleFontFace = src.TitleFontFace
	if w.MapKey.TitleFontColor, err = decodeRgba(src.TitleFontColor); err != nil {
		return fmt.Errorf("mapkey.titleFontColor: %w", err)
	}
	w.MapKey.TitleFontBold = src.TitleFontBold
	w.MapKey.TitleFontItalic = src.TitleFontItalic
	w.MapKey.TitleScale = src.TitleScale
	w.MapKey.ScaleText = src.ScaleText
	w.MapKey.ScaleFontFace = src.ScaleFontFace
	if w.MapKey.ScaleFontColor, err = decodeRgba(src.ScaleFontColor); err != nil {
		return fmt.Errorf("mapkey.scaleFontColor: %w", err)
	}
	w.MapKey.ScaleFontBold = src.ScaleFontBold
	w.MapKey.ScaleFontItalic = src.ScaleFontItalic
	w.MapKey.ScaleScale = src.ScaleScale
	w.MapKey.EntryFontFace = src.EntryFontFace
	if w.MapKey.EntryFontColor, err = decodeRgba(src.EntryFontColor); err != nil {
		return fmt.Errorf("mapkey.entryFontColor: %w", err)
	}
	w.MapKey.EntryFontBold = src.EntryFontBold
	w.MapKey.EntryFontItalic = src.EntryFontItalic
	w.MapKey.EntryScale = src.EntryScale
	return nil
}

func encodeMapKey(mapKey *wxx.MapKey_t, wb *bytes.Buffer) error {
	wb.WriteString("<mapkey")
	wb.WriteString(fmt.Sprintf(" positionx=%q", floats(mapKey.PositionX)))
	wb.WriteString(fmt.Sprintf(" positiony=%q", floats(mapKey.PositionY)))
	wb.WriteString(fmt.Sprintf(" viewlevel=%q", mapKey.Viewlevel))
	wb.WriteString(fmt.Sprintf(" height=%q", floats(mapKey.Height)))
	wb.WriteString(fmt.Sprintf(" backgroundcolor=%q", rgbas(mapKey.BackgroundColor))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" backgroundopacity=%q", floats(mapKey.BackgroundOpacity)))
	wb.WriteString(fmt.Sprintf(" titleText=%q", mapKey.TitleText))
	wb.WriteString(fmt.Sprintf(" titleFontFace=%q", mapKey.TitleFontFace))
	wb.WriteString(fmt.Sprintf(" titleFontColor=%q", rgbas(mapKey.TitleFontColor))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" titleFontBold=%q", bools(mapKey.TitleFontBold)))
	wb.WriteString(fmt.Sprintf(" titleFontItalic=%q", bools(mapKey.TitleFontItalic)))
	wb.WriteString(fmt.Sprintf(" titleScale=%q", floats(mapKey.TitleScale)))
	wb.WriteString(fmt.Sprintf(" scaleText=%q", mapKey.ScaleText))
	wb.WriteString(fmt.Sprintf(" scaleFontFace=%q", mapKey.ScaleFontFace))
	wb.WriteString(fmt.Sprintf(" scaleFontColor=%q", rgbas(mapKey.ScaleFontColor))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" scaleFontBold=%q", bools(mapKey.ScaleFontBold)))
	wb.WriteString(fmt.Sprintf(" scaleFontItalic=%q", bools(mapKey.ScaleFontItalic)))
	wb.WriteString(fmt.Sprintf(" scaleScale=%q", floats(mapKey.ScaleScale)))
	wb.WriteString(fmt.Sprintf(" entryFontFace=%q", mapKey.EntryFontFace))
	wb.WriteString(fmt.Sprintf(" entryFontColor=%q", rgbas(mapKey.EntryFontColor))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" entryFontBold=%q", bools(mapKey.EntryFontBold)))
	wb.WriteString(fmt.Sprintf(" entryFontItalic=%q", bools(mapKey.EntryFontItalic)))
	wb.WriteString(fmt.Sprintf(" entryScale=%q", floats(mapKey.EntryScale)))
	wb.WriteString(">\n")
	wb.WriteString("</mapkey>\n")
	return nil
}
