// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package adapters

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type OffsetCoords struct {
	Row int
	Col int
}

// ConvertTNCoords parses a coordinate like "AB 0102" and returns the global row and column.
// Sub-maps are 21 rows x 30 columns. Origin (1,1) is top-left.
func ConvertTNCoords(input string) (coords OffsetCoords, err error) {
	input = strings.ToUpper(input)
	if len(input) != 7 || input[2] != ' ' {
		return coords, fmt.Errorf("invalid format: expected 'AB 0102'")
	}

	gridRow := input[0]
	gridCol := input[1]

	if !unicode.IsUpper(rune(gridRow)) || !unicode.IsUpper(rune(gridCol)) {
		return coords, fmt.Errorf("invalid grid letters: must be uppercase A-Z")
	}

	subColStr := input[3:5]
	subRowStr := input[5:7]

	subCol, err := strconv.Atoi(subColStr)
	if err != nil || subCol < 1 || subCol > 30 {
		return coords, fmt.Errorf("invalid sub-map column: %s", subColStr)
	}

	subRow, err := strconv.Atoi(subRowStr)
	if err != nil || subRow < 1 || subRow > 21 {
		return coords, fmt.Errorf("invalid sub-map row: %s", subRowStr)
	}

	// Convert letters to 0-based grid index (A=0, B=1, ...)
	gridRowIdx := int(gridRow - 'A')
	gridColIdx := int(gridCol - 'A')

	// Calculate global row and column (1-based)
	coords.Row = gridRowIdx*21 + subRow
	coords.Col = gridColIdx*30 + subCol

	return coords, nil
}

// OffsetToAxial converts from (row, col) with top-left origin (1-based) to (q, r) axial coordinates.
// Assumes even-q vertical layout (flat-topped hexes, even columns pushed down).
func OffsetToAxial(coords OffsetCoords) (axial AxialCoords) {
	col := coords.Col - 1
	row := coords.Row - 1

	axial.Q = col
	axial.R = row - (col / 2) // even-q layout
	return axial
}

type AxialCoords struct {
	Q, R int
}

// ToTNCoords converts axial (q, r) to a TribeNet coordinate string "AB 0102".
// Assumes:
// - Sub-map size: 21 rows x 30 columns
// - Top-left map origin is (1,1) in TribeNet grid
// - Even-q vertical layout (flat-topped hexes)
func (a AxialCoords) ToTNCoords() (string, error) {
	col := a.Q + 1
	row := a.R + (a.Q-(a.Q&1))/2 + 1 // even-q layout

	gridRow := (row - 1) / 21
	gridCol := (col - 1) / 30
	if gridRow >= 26 || gridCol >= 26 || gridRow < 0 || gridCol < 0 {
		return "", fmt.Errorf("coordinates out of bounds for TribeNet grid")
	}

	rowInSub := (row-1)%21 + 1
	colInSub := (col-1)%30 + 1

	gridRowChar := 'A' + rune(gridRow)
	gridColChar := 'A' + rune(gridCol)

	return fmt.Sprintf("%c%c %02d%02d", gridRowChar, gridColChar, colInSub, rowInSub), nil
}

type MapData struct {
	Origin AxialCoords
	Height int
	Width  int
}

// MergeMaps returns a MapData that bounds all the provided maps.
func MergeMaps(maps ...MapData) MapData {
	if len(maps) == 0 {
		return MapData{}
	}

	minQ := math.MaxInt
	minR := math.MaxInt
	maxQ := math.MinInt
	maxR := math.MinInt

	for _, m := range maps {
		startQ := m.Origin.Q
		startR := m.Origin.R
		endQ := startQ + m.Width - 1
		endR := startR + m.Height - 1

		if startQ < minQ {
			minQ = startQ
		}
		if startR < minR {
			minR = startR
		}
		if endQ > maxQ {
			maxQ = endQ
		}
		if endR > maxR {
			maxR = endR
		}
	}

	return MapData{
		Origin: AxialCoords{Q: minQ, R: minR},
		Width:  maxQ - minQ + 1,
		Height: maxR - minR + 1,
	}
}

// Bounds returns the top-left and bottom-right axial coordinates of the map.
// These are inclusive bounds.
func (m MapData) Bounds() (topLeft AxialCoords, bottomRight AxialCoords) {
	topLeft = m.Origin
	bottomRight = AxialCoords{
		Q: m.Origin.Q + m.Width - 1,
		R: m.Origin.R + m.Height - 1,
	}
	return
}

func (m MapData) Corners() (topLeft, topRight, bottomLeft, bottomRight AxialCoords) {
	topLeft = m.Origin
	br := m.Origin
	tr := m.Origin
	bl := m.Origin

	tr.Q += m.Width - 1
	bl.R += m.Height - 1
	br.Q += m.Width - 1

	// Adjust R for BR properly:
	//col := br.Q + 1
	row := m.Origin.R + (m.Origin.Q-(m.Origin.Q&1))/2 + 1 + m.Height - 1
	br.R = row - 1 - (br.Q-(br.Q&1))/2

	return topLeft, tr, bl, br
}
