// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// TribeNetCoord implements "even-q," an offset coordinate with flat top hexes and even columns pushed down.
// It implements a different Stringer, displaying coordinates as "AB 0102," where:
//   - "A"  is grid row        with a range of A  ... Z
//   - "B"  is grid column     with a range of A  ... Z
//   - "01" is sub-grid column with a range of 1 ... 30
//   - "02" is sub-grid row    with a range of 1 ... 21
//
// Note that the TribeNet map uses "AA 0101" for the origin, accepts "##" as an anonymous grid, and considers
// "N/A" to be the null coordinate.
type TribeNetCoord struct {
	id   string
	cube CubeCoord
}

const (
	columnsPerGrid = 30
	rowsPerGrid    = 21
)

// NewTribeNetCoord converts a grid id to coordinates, returning any errors.
// For historical reasons, we treat grid "##" as "QQ" and an id of "N/A" as
// cube coordinates (0,0,0).
//
// Note that we always convert the grid id to uppercase.
func NewTribeNetCoord(id string) (TribeNetCoord, error) {
	// force the grid id to uppercase before converting
	id = strings.ToUpper(id)

	if validGridId := id == "N/A" || (len(id) == 7 && id[2] == ' '); !validGridId {
		return TribeNetCoord{}, wxx.ErrInvalidGridCoordinates
	}

	if id == "N/A" {
		return TribeNetCoord{id: id}, nil
	}

	// extract and validate the grid row and column
	gridRow, gridColumn := int(id[0]), int(id[1])
	if gridRow == '#' && gridColumn == '#' {
		// we have to put obscured coordinates somewhere, so we will put them in "QQ"
		gridRow, gridColumn = 'Q', 'Q'
	} else if isValidGridRow := 'A' <= gridRow && gridRow < 'Z'; !isValidGridRow {
		return TribeNetCoord{}, wxx.ErrInvalidGridCoordinates
	} else if isValidGridColumn := 'A' <= gridColumn && gridColumn < 'Z'; !isValidGridColumn {
		return TribeNetCoord{}, wxx.ErrInvalidGridCoordinates
	}
	// convert from "A" ... "Z" to 0 ... 26
	gridRow, gridColumn = gridRow-'A', gridColumn-'A'

	// extract and validate the sub-grid column and row
	subGridColumn, err := strconv.Atoi(id[3:5])
	if err != nil {
		return TribeNetCoord{}, wxx.ErrInvalidGridCoordinates
	}
	subGridRow, err := strconv.Atoi(id[5:])
	if err != nil {
		return TribeNetCoord{}, wxx.ErrInvalidGridCoordinates
	}
	if isValidSubGridColumn := 1 <= subGridColumn && subGridColumn <= columnsPerGrid; !isValidSubGridColumn {
		return TribeNetCoord{}, wxx.ErrInvalidGridCoordinates
	} else if isValidSubGridRow := 1 <= subGridRow && subGridRow <= rowsPerGrid; !isValidSubGridRow {
		return TribeNetCoord{}, wxx.ErrInvalidGridCoordinates
	}
	// convert from 1 based to 0 based
	subGridColumn, subGridRow = subGridColumn-1, subGridRow-1

	return TribeNetCoord{
		id: id,
		cube: EvenQCoord{
			col: gridColumn*columnsPerGrid + subGridColumn,
			row: gridRow*rowsPerGrid + subGridRow,
		}.ToCube(),
	}, nil
}

// Equals returns true if the original grid ids of the two coordinates are
// the same. This is wonky because of "N/A" and obscured grids, but seems
// like the best compromise.
func (a TribeNetCoord) Equals(b TribeNetCoord) bool {
	return a.id == b.id
}

// GridID returns the internal coordinates converted to a grid id.
// "N/A" and obscured coordinates may cause some surprise.
func (a TribeNetCoord) GridID() string {
	evenq := a.cube.ToEvenQ()
	gridRow, gridColumn := evenq.row/rowsPerGrid, evenq.col/columnsPerGrid
	subGridColumn, subGridRow := evenq.col-gridColumn*columnsPerGrid, evenq.row-gridRow*rowsPerGrid
	return fmt.Sprintf("%c%c %02d%02d", 'A'+gridRow, 'A'+gridColumn, subGridColumn, subGridRow)
}

// IsNA returns true if the coordinates were "N/A" or empty.
func (a TribeNetCoord) IsNA() bool {
	return a.id == "" || a.id == "N/A"
}

// String implements the strings.Stringer interface and returns the original grid id converted to upper-cose.
func (a TribeNetCoord) String() string {
	if a.id == "" {
		return "N/A"
	}
	return a.id
}

func (a TribeNetCoord) ToCube() CubeCoord {
	return a.cube
}
