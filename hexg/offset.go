// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

type OffsetCoord struct {
	col int
	row int
}

func NewOffsetCoord(col_, row_ int) OffsetCoord {
	return OffsetCoord{col: col_, row: row_}
}

func (a OffsetCoord) Equals(b OffsetCoord) bool {
	return a.col == b.col && a.row == b.row
}
