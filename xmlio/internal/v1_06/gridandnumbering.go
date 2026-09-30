// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeGridAndNumbering copies the parsed <gridandnumbering> attributes into the
// domain map. All 30 attributes are modeled.
func decodeGridAndNumbering(src GridAndNumbering, w *wxx.Map_t) {
	w.GridAndNumbering = &wxx.GridAndNumbering_t{}
	w.GridAndNumbering.Color0 = src.Color0
	w.GridAndNumbering.Color1 = src.Color1
	w.GridAndNumbering.Color2 = src.Color2
	w.GridAndNumbering.Color3 = src.Color3
	w.GridAndNumbering.Color4 = src.Color4
	w.GridAndNumbering.Width0 = src.Width0
	w.GridAndNumbering.Width1 = src.Width1
	w.GridAndNumbering.Width2 = src.Width2
	w.GridAndNumbering.Width3 = src.Width3
	w.GridAndNumbering.Width4 = src.Width4
	w.GridAndNumbering.GridOffsetContinentKingdomX = src.GridOffsetContinentKingdomX
	w.GridAndNumbering.GridOffsetContinentKingdomY = src.GridOffsetContinentKingdomY
	w.GridAndNumbering.GridOffsetWorldContinentX = src.GridOffsetWorldContinentX
	w.GridAndNumbering.GridOffsetWorldContinentY = src.GridOffsetWorldContinentY
	w.GridAndNumbering.GridOffsetWorldKingdomX = src.GridOffsetWorldKingdomX
	w.GridAndNumbering.GridOffsetWorldKingdomY = src.GridOffsetWorldKingdomY
	w.GridAndNumbering.GridSquare = src.GridSquare
	w.GridAndNumbering.GridSquareHeight = src.GridSquareHeight
	w.GridAndNumbering.GridSquareWidth = src.GridSquareWidth
	w.GridAndNumbering.GridOffsetX = src.GridOffsetX
	w.GridAndNumbering.GridOffsetY = src.GridOffsetY
	w.GridAndNumbering.NumberFont = src.NumberFont
	w.GridAndNumbering.NumberColor = src.NumberColor
	w.GridAndNumbering.NumberSize = src.NumberSize
	w.GridAndNumbering.NumberStyle = src.NumberStyle
	w.GridAndNumbering.NumberFirstCol = src.NumberFirstCol
	w.GridAndNumbering.NumberFirstRow = src.NumberFirstRow
	w.GridAndNumbering.NumberOrder = src.NumberOrder
	w.GridAndNumbering.NumberPosition = src.NumberPosition
	w.GridAndNumbering.NumberPrePad = src.NumberPrePad
	w.GridAndNumbering.NumberSeparator = src.NumberSeparator
}

func encodeGridAndNumbering(gridAndNumbering *wxx.GridAndNumbering_t, wb *bytes.Buffer) error {
	// The five grid colors are strings written verbatim; check them before
	// anything is written (#83).
	var colors [5]string
	for i, c := range []string{gridAndNumbering.Color0, gridAndNumbering.Color1, gridAndNumbering.Color2, gridAndNumbering.Color3, gridAndNumbering.Color4} {
		var err error
		if colors[i], err = hexColorAttr(fmt.Sprintf("map/gridandnumbering/@color%d", i), c); err != nil {
			return err
		}
	}
	wb.WriteString(fmt.Sprintf(`<gridandnumbering`))
	wb.WriteString(fmt.Sprintf(" color0=%s", xmlAttr(colors[0])))
	wb.WriteString(fmt.Sprintf(" color1=%s", xmlAttr(colors[1])))
	wb.WriteString(fmt.Sprintf(" color2=%s", xmlAttr(colors[2])))
	wb.WriteString(fmt.Sprintf(" color3=%s", xmlAttr(colors[3])))
	wb.WriteString(fmt.Sprintf(" color4=%s", xmlAttr(colors[4])))
	wb.WriteString(fmt.Sprintf(" width0=%s", xmlAttr(floats(gridAndNumbering.Width0))))
	wb.WriteString(fmt.Sprintf(" width1=%s", xmlAttr(floats(gridAndNumbering.Width1))))
	wb.WriteString(fmt.Sprintf(" width2=%s", xmlAttr(floats(gridAndNumbering.Width2))))
	wb.WriteString(fmt.Sprintf(" width3=%s", xmlAttr(floats(gridAndNumbering.Width3))))
	wb.WriteString(fmt.Sprintf(" width4=%s", xmlAttr(floats(gridAndNumbering.Width4))))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomX=%s", xmlAttr(floats(gridAndNumbering.GridOffsetContinentKingdomX))))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomY=%s", xmlAttr(floats(gridAndNumbering.GridOffsetContinentKingdomY))))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentX=%s", xmlAttr(floats(gridAndNumbering.GridOffsetWorldContinentX))))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentY=%s", xmlAttr(floats(gridAndNumbering.GridOffsetWorldContinentY))))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomX=%s", xmlAttr(floats(gridAndNumbering.GridOffsetWorldKingdomX))))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomY=%s", xmlAttr(floats(gridAndNumbering.GridOffsetWorldKingdomY))))
	wb.WriteString(fmt.Sprintf(" gridSquare=%s", xmlAttr(ints(gridAndNumbering.GridSquare))))
	wb.WriteString(fmt.Sprintf(" gridSquareHeight=%s", xmlAttr(floats(gridAndNumbering.GridSquareHeight))))
	wb.WriteString(fmt.Sprintf(" gridSquareWidth=%s", xmlAttr(floats(gridAndNumbering.GridSquareWidth))))
	wb.WriteString(fmt.Sprintf(" gridOffsetX=%s", xmlAttr(floats(gridAndNumbering.GridOffsetX))))
	wb.WriteString(fmt.Sprintf(" gridOffsetY=%s", xmlAttr(floats(gridAndNumbering.GridOffsetY))))
	wb.WriteString(fmt.Sprintf(" numberFont=%s", xmlAttr(gridAndNumbering.NumberFont)))
	wb.WriteString(fmt.Sprintf(" numberColor=%s", xmlAttr(gridAndNumbering.NumberColor)))
	wb.WriteString(fmt.Sprintf(" numberSize=%s", xmlAttr(ints(gridAndNumbering.NumberSize))))
	wb.WriteString(fmt.Sprintf(" numberStyle=%s", xmlAttr(gridAndNumbering.NumberStyle)))
	wb.WriteString(fmt.Sprintf(" numberFirstCol=%s", xmlAttr(ints(gridAndNumbering.NumberFirstCol))))
	wb.WriteString(fmt.Sprintf(" numberFirstRow=%s", xmlAttr(ints(gridAndNumbering.NumberFirstRow))))
	wb.WriteString(fmt.Sprintf(" numberOrder=%s", xmlAttr(gridAndNumbering.NumberOrder)))
	wb.WriteString(fmt.Sprintf(" numberPosition=%s", xmlAttr(gridAndNumbering.NumberPosition)))
	wb.WriteString(fmt.Sprintf(" numberPrePad=%s", xmlAttr(gridAndNumbering.NumberPrePad)))
	wb.WriteString(fmt.Sprintf(" numberSeparator=%s", xmlAttr(gridAndNumbering.NumberSeparator)))
	wb.WriteString(fmt.Sprintf(" />\n"))
	return nil
}
