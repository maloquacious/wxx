// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import "math"

// Layout_i defines the interface for layouts.
//
// Orientation is important for offset coordinates and every layout
// that implements this interface is expected to implement that per
// the Red Blob Games guide.
type Layout_i interface {
	// IsHorizontal returns true if the layout has horizontal rows.
	// Horizontal layouts have pointy-top hexes, staggered columns, and horizontal rows.
	IsHorizontal() bool

	// IsVertical returns true if the layout has vertical columns.
	// Vertical layouts have flat-top hexes, vertical columns, and staggered rows.
	IsVertical() bool

	// OffsetType returns the type of offset used for columns and rows.
	OffsetType() Orientation_e

	// DirectionToBearing returns the bearing of a direction in the layout
	DirectionToBearing(direction int) string

	// HexagonalGrid returns a grid centered about a hex.
	HexagonalGrid(center CubeCoord, radius int) []CubeCoord

	// HexCorner returns the screen coordinates of the hex corner.
	// We should define what "corner" means in this context.
	HexCorner(h CubeCoord, corner int) Point

	// HexPoints returns the screen coordinates for the center, then every corner starting somewhere and going counter-clockwise.
	HexPoints(h CubeCoord) [7]Point

	// HexToOffsetCoord returns the offset coordinates of the hex.
	// Uses the offset from the layout to shift rows and columns correctly.
	HexToOffsetCoord(h CubeCoord) OffsetCoord

	// HexToPixel returns the origin of the hex on the screen as a pixel.
	HexToPixel(h CubeCoord) Point

	// ColRowToHex returns a new hex using offset column and row coordinates.
	ColRowToHex(col, row int) CubeCoord

	// OffsetCoordToHex returns a new hex from the OffsetCoord.
	OffsetCoordToHex(oc OffsetCoord) CubeCoord

	// ParallelogramGrid returns a grid originating at (0,0,0).
	ParallelogramGrid(q1, r1, q2, r2 int) []CubeCoord

	// PixelToHexRounded turns a fractional hex into a regular hex coordinate:
	PixelToHexRounded(p Point) CubeCoord

	// PixelToFractionalHex returns the fractional hex that encloses the pixel.
	// In theory, the origin of that fractional hex will be the pixel.
	PixelToFractionalHex(p Point) FractionalCubeCoord

	// PolygonCornerOffset returns the offset from the center of a hex to a corner.
	// We should define what the parameter "corner" means. Which corner?
	PolygonCornerOffset(corner int) Point

	// PolygonCornerOffsets returns the offset for every corner of a hex.
	PolygonCornerOffsets() [6]Point

	// RectangularGrid returns a grid centered about a hex.
	RectangularGrid(center CubeCoord, left, right, top, bottom int) []CubeCoord

	// TriagonalGrid returns a grid originating at (0,0,0).
	TriagonalGrid(side_length int) []CubeCoord
}

type Layout struct {
	orientation Orientation
	size        Point
	origin      Point
}

func NewLayout(orientation_ Orientation, size_, origin_ Point) Layout {
	return Layout{orientation: orientation_, size: size_, origin: origin_}
}

var layout_pointy = Orientation{math.Sqrt(3.0), math.Sqrt(3.0) / 2.0, 0.0, 3.0 / 2.0, math.Sqrt(3.0) / 3.0, -1.0 / 3.0, 0.0, 2.0 / 3.0, 0.5}

var layout_flat = Orientation{3.0 / 2.0, 0.0, math.Sqrt(3.0) / 2.0, math.Sqrt(3.0), 2.0 / 3.0, 0.0, -1.0 / 3.0, math.Sqrt(3.0) / 3.0, 0.0}

func hex_corner_offset(layout Layout, corner int) Point {
	angle := 2.0 * math.Pi * (layout.orientation.start_angle - float64(corner)) / 6.0
	return Point{x: layout.size.x * math.Cos(angle), y: layout.size.y * math.Sin(angle)}
}
