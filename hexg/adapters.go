// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import "math"

// Generated code -- CC0 -- No Rights Reserved -- http://www.redblobgames.com/grids/hexagons/

const EVEN = 1

const ODD = -1

func qoffset_from_cube(offset int, h CubeCoord) OffsetCoord {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.q & 1
	col, row := h.q, h.r+int((h.q+offset*parity)/2)
	return OffsetCoord{col: col, row: row}
}

func (h CubeCoord) ToEvenQ() EvenQCoord {
	parity := h.q & 1
	col, row := h.q, h.r+((h.q+EVEN*parity)/2)
	return EvenQCoord{col: col, row: row}
}

func (h CubeCoord) ToOddQ() EvenQCoord {
	parity := h.q & 1
	col, row := h.q, h.r+((h.q+ODD*parity)/2)
	return EvenQCoord{col: col, row: row}
}

func qoffset_to_cube(offset int, h OffsetCoord) CubeCoord {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.col & 1
	q, r := h.col, h.row-((h.col+offset*parity)/2)
	return CubeCoord{q: q, r: r, s: -q - r}
}

func (h EvenQCoord) ToCube() CubeCoord {
	parity := h.col & 1
	q, r := h.col, h.row-((h.col+EVEN*parity)/2)
	return CubeCoord{q: q, r: r, s: -q - r}
}

func (h OddQCoord) ToCube() CubeCoord {
	parity := h.col & 1
	q, r := h.col, h.row-((h.col+ODD*parity)/2)
	return CubeCoord{q: q, r: r, s: -q - r}
}

func roffset_from_cube(offset int, h CubeCoord) OffsetCoord {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.r & 1
	col, row := h.q+((h.r+offset*parity)/2), h.r
	return OffsetCoord{col: col, row: row}
}

func (h CubeCoord) ToEvenR() EvenRCoord {
	parity := h.r & 1
	col, row := h.q+((h.r+EVEN*parity)/2), h.r
	return EvenRCoord{col: col, row: row}
}

func (h CubeCoord) ToOddR() OddRCoord {
	parity := h.r & 1
	col, row := h.q+((h.r+EVEN*parity)/2), h.r
	return OddRCoord{col: col, row: row}
}

func roffset_to_cube(offset int, h OffsetCoord) CubeCoord {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.row & 1
	q, r := h.col-((h.row+offset*parity)/2), h.row
	return CubeCoord{q: q, r: r, s: -q - r}
}

func (h EvenRCoord) ToCube() CubeCoord {
	parity := h.row & 1
	q, r := h.col-((h.row+EVEN*parity)/2), h.row
	return CubeCoord{q: q, r: r, s: -q - r}
}

func (h OddRCoord) ToCube() CubeCoord {
	parity := h.row & 1
	q, r := h.col-((h.row+EVEN*parity)/2), h.row
	return CubeCoord{q: q, r: r, s: -q - r}
}

func qoffset_from_qdoubled(offset int, h DoubledCoord) OffsetCoord {
	parity := h.col & 1
	return OffsetCoord{col: h.col, row: int((h.row + offset*parity) / 2)}
}

func qoffset_to_qdoubled(offset int, h OffsetCoord) DoubledCoord {
	parity := h.col & 1
	return DoubledCoord{col: h.col, row: 2*h.row - offset*parity}
}

func roffset_from_rdoubled(offset int, h DoubledCoord) OffsetCoord {
	parity := h.row & 1
	return OffsetCoord{col: int((h.col + offset*parity) / 2), row: h.row}
}

func roffset_to_rdoubled(offset int, h OffsetCoord) DoubledCoord {
	parity := h.row & 1
	return DoubledCoord{col: 2*h.col - offset*parity, row: h.row}
}

func qdoubled_from_cube(h CubeCoord) DoubledCoord {
	col, row := h.q, 2*h.r+h.q
	return DoubledCoord{col: col, row: row}
}

func (h CubeCoord) ToDoubleWidth() DoubleWidthCoord {
	col, row := h.q, 2*h.r+h.q
	return DoubleWidthCoord{col: col, row: row}
}

func qdoubled_to_cube(h DoubledCoord) CubeCoord {
	q, r := h.col, int((h.row-h.col)/2)
	return CubeCoord{q: q, r: r, s: -q - r}
}

func (h DoubleWidthCoord) ToCube() CubeCoord {
	q, r := h.col, int((h.row-h.col)/2)
	return CubeCoord{q: q, r: r, s: -q - r}
}

func rdoubled_from_cube(h CubeCoord) DoubledCoord {
	col, row := 2*h.q+h.r, h.r
	return DoubledCoord{col: col, row: row}
}

func (h CubeCoord) ToDoubleHeight() DoubleHeightCoord {
	col, row := 2*h.q+h.r, h.r
	return DoubleHeightCoord{col: col, row: row}
}

func rdoubled_to_cube(h DoubledCoord) CubeCoord {
	q, r := int((h.col-h.row)/2), h.row
	return CubeCoord{q: q, r: r, s: -q - r}
}

func (h DoubleHeightCoord) ToCube() CubeCoord {
	q, r := int((h.col-h.row)/2), h.row
	return CubeCoord{q: q, r: r, s: -q - r}
}

func cube_to_pixel(layout Layout, h CubeCoord) Point {
	x := (layout.orientation.f0*float64(h.q) + layout.orientation.f1*float64(h.r)) * layout.size.x
	y := (layout.orientation.f2*float64(h.q) + layout.orientation.f3*float64(h.r)) * layout.size.y
	return Point{x: x + layout.origin.x, y: y + layout.origin.y}
}

func pixel_to_fractional_cube(layout Layout, p Point) FractionalCubeCoord {
	pt := Point{x: (p.x - layout.origin.x) / layout.size.x, y: (p.y - layout.origin.y) / layout.size.y}
	q := layout.orientation.b0*pt.x + layout.orientation.b1*pt.y
	r := layout.orientation.b2*pt.x + layout.orientation.b3*pt.y
	return FractionalCubeCoord{q: q, r: r, s: -q - r}
}

func pixel_to_cube_rounded(layout Layout, p Point) CubeCoord {
	return pixel_to_fractional_cube(layout, p).Round()
}

func polygon_corners(layout Layout, h CubeCoord) []Point {
	corners := make([]Point, 6)
	center := cube_to_pixel(layout, h)
	for i := 0; i < 6; i++ {
		offset := hex_corner_offset(layout, i)
		corners[i] = Point{x: center.x + offset.x, y: center.y + offset.y}
	}
	return corners
}

// helpers for math
//

// number is a constraint that permits any int or float64.
type number interface {
	~int | ~float64
}

// abs returns the absolute value of x.
func abs[T number](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// lerp is a generic linear interpolation function.
// Accepts any integer or floating-point for a and b, always returns float64.
func lerp[T number](a, b T, t float64) float64 {
	return float64(a) + (float64(b)-float64(a))*t
}

func round(f float64) int {
	return int(math.Round(f))
}
