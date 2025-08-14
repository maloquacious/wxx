// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

type DoubledCoord struct {
	col int
	row int
}

func NewDoubledCoord(col_, row_ int) DoubledCoord {
	return DoubledCoord{col: col_, row: row_}
}

func (a DoubledCoord) Equals(b DoubledCoord) bool {
	return a.col == b.col && a.row == b.row
}

// DoubleHeightCoord implements "double-height," a doubled coordinate with pointy top hexes
type DoubleHeightCoord struct {
	col int
	row int
}

func NewDoubleHeightCoord(col_, row_ int) DoubleHeightCoord {
	return DoubleHeightCoord{col: col_, row: row_}
}

func (a DoubleHeightCoord) Equals(b DoubleHeightCoord) bool {
	return a.col == b.col && a.row == b.row
}

// DoubleWidthCoord implements "double-width," a doubled coordinate with flat top hexes
type DoubleWidthCoord struct {
	col int
	row int
}

func NewDoubleWidthCoord(col_, row_ int) DoubleWidthCoord {
	return DoubleWidthCoord{col: col_, row: row_}
}

func (a DoubleWidthCoord) Equals(b DoubleWidthCoord) bool {
	return a.col == b.col && a.row == b.row
}
