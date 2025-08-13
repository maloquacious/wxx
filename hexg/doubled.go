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
