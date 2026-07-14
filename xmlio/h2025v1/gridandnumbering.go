// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

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
	wb.WriteString(fmt.Sprintf(`<gridandnumbering`))
	wb.WriteString(fmt.Sprintf(" color0=%q", gridAndNumbering.Color0))
	wb.WriteString(fmt.Sprintf(" color1=%q", gridAndNumbering.Color1))
	wb.WriteString(fmt.Sprintf(" color2=%q", gridAndNumbering.Color2))
	wb.WriteString(fmt.Sprintf(" color3=%q", gridAndNumbering.Color3))
	wb.WriteString(fmt.Sprintf(" color4=%q", gridAndNumbering.Color4))
	wb.WriteString(fmt.Sprintf(" width0=%q", floats(gridAndNumbering.Width0)))
	wb.WriteString(fmt.Sprintf(" width1=%q", floats(gridAndNumbering.Width1)))
	wb.WriteString(fmt.Sprintf(" width2=%q", floats(gridAndNumbering.Width2)))
	wb.WriteString(fmt.Sprintf(" width3=%q", floats(gridAndNumbering.Width3)))
	wb.WriteString(fmt.Sprintf(" width4=%q", floats(gridAndNumbering.Width4)))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomX=%q", floats(gridAndNumbering.GridOffsetContinentKingdomX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomY=%q", floats(gridAndNumbering.GridOffsetContinentKingdomY)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentX=%q", floats(gridAndNumbering.GridOffsetWorldContinentX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentY=%q", floats(gridAndNumbering.GridOffsetWorldContinentY)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomX=%q", floats(gridAndNumbering.GridOffsetWorldKingdomX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomY=%q", floats(gridAndNumbering.GridOffsetWorldKingdomY)))
	wb.WriteString(fmt.Sprintf(" gridSquare=%q", ints(gridAndNumbering.GridSquare)))
	wb.WriteString(fmt.Sprintf(" gridSquareHeight=%q", floats(gridAndNumbering.GridSquareHeight)))
	wb.WriteString(fmt.Sprintf(" gridSquareWidth=%q", floats(gridAndNumbering.GridSquareWidth)))
	wb.WriteString(fmt.Sprintf(" gridOffsetX=%q", floats(gridAndNumbering.GridOffsetX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetY=%q", floats(gridAndNumbering.GridOffsetY)))
	wb.WriteString(fmt.Sprintf(" numberFont=%q", gridAndNumbering.NumberFont))
	wb.WriteString(fmt.Sprintf(" numberColor=%q", gridAndNumbering.NumberColor))
	wb.WriteString(fmt.Sprintf(" numberSize=%q", ints(gridAndNumbering.NumberSize)))
	wb.WriteString(fmt.Sprintf(" numberStyle=%q", gridAndNumbering.NumberStyle))
	wb.WriteString(fmt.Sprintf(" numberFirstCol=%q", ints(gridAndNumbering.NumberFirstCol)))
	wb.WriteString(fmt.Sprintf(" numberFirstRow=%q", ints(gridAndNumbering.NumberFirstRow)))
	wb.WriteString(fmt.Sprintf(" numberOrder=%q", gridAndNumbering.NumberOrder))
	wb.WriteString(fmt.Sprintf(" numberPosition=%q", gridAndNumbering.NumberPosition))
	wb.WriteString(fmt.Sprintf(" numberPrePad=%q", gridAndNumbering.NumberPrePad))
	wb.WriteString(fmt.Sprintf(" numberSeparator=%q", gridAndNumbering.NumberSeparator))
	wb.WriteString(fmt.Sprintf(" />\n"))
	return nil
}
