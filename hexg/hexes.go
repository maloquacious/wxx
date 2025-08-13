// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import "math"

// Generated code -- CC0 -- No Rights Reserved -- http://www.redblobgames.com/grids/hexagons/

type Point struct {
	x float64
	y float64
}

func NewPoint(x_, y_ float64) Point {
	return Point{x: x_, y: y_}
}

type Hex struct {
	q int
	r int
	s int
}

func NewHex(q_, r_, s_ int) Hex {
	if q_+r_+s_ != 0 {
		panic("q + r + s must be 0")
	}
	return Hex{q: q_, r: r_, s: s_}
}

type FractionalHex struct {
	q float64
	r float64
	s float64
}

func NewFractionalHex(q_, r_, s_ float64) FractionalHex {
	if round(q_+r_+s_) != 0 {
		panic("q + r + s must be 0")
	}
	return FractionalHex{q: q_, r: r_, s: s_}
}

type OffsetCoord struct {
	col int
	row int
}

func NewOffsetCoord(col_, row_ int) OffsetCoord {
	return OffsetCoord{col: col_, row: row_}
}

type DoubledCoord struct {
	col int
	row int
}

func NewDoubledCoord(col_, row_ int) DoubledCoord {
	return DoubledCoord{col: col_, row: row_}
}

type Orientation struct {
	f0          float64
	f1          float64
	f2          float64
	f3          float64
	b0          float64
	b1          float64
	b2          float64
	b3          float64
	start_angle float64
}

func NewOrientation(f0_, f1_, f2_, f3_, b0_, b1_, b2_, b3_, start_angle_ float64) Orientation {
	return Orientation{f0: f0_, f1: f1_, f2: f2_, f3: f3_, b0: b0_, b1: b1_, b2: b2_, b3: b3_, start_angle: start_angle_}
}

type Layout struct {
	orientation Orientation
	size        Point
	origin      Point
}

func NewLayout(orientation_ Orientation, size_, origin_ Point) Layout {
	return Layout{orientation: orientation_, size: size_, origin: origin_}
}

func hex_add(a, b Hex) Hex {
	return Hex{q: a.q + b.q, r: a.r + b.r, s: a.s + b.s}
}

func hex_subtract(a, b Hex) Hex {
	return Hex{q: a.q - b.q, r: a.r - b.r, s: a.s - b.s}
}

func hex_scale(a Hex, k int) Hex {
	return Hex{q: a.q * k, r: a.r * k, s: a.s * k}
}

func hex_rotate_left(a Hex) Hex {
	return Hex{q: -a.s, r: -a.q, s: -a.r}
}

func hex_rotate_right(a Hex) Hex {
	return Hex{q: -a.r, r: -a.s, s: -a.q}
}

var hex_directions = []Hex{
	Hex{q: 1, r: 0, s: -1},
	Hex{q: 1, r: -1, s: 0},
	Hex{q: 0, r: -1, s: 1},
	Hex{q: -1, r: 0, s: 1},
	Hex{q: -1, r: 1, s: 0},
	Hex{q: 0, r: 1, s: -1},
}

func hex_direction(direction int) Hex {
	return hex_directions[(6+(direction%6))%6]
}

func hex_neighbor(hex Hex, direction int) Hex {
	return hex_add(hex, hex_direction(direction))
}

var hex_diagonals = []Hex{
	Hex{q: 2, r: -1, s: -1},
	Hex{q: 1, r: -2, s: 1},
	Hex{q: -1, r: -1, s: 2},
	Hex{q: -2, r: 1, s: 1},
	Hex{q: -1, r: 2, s: -1},
	Hex{q: 1, r: 1, s: -2},
}

func hex_diagonal_neighbor(hex Hex, direction int) Hex {
	return hex_add(hex, hex_diagonals[direction])
}

func hex_length(hex Hex) int {
	return int((abs(hex.q) + abs(hex.r) + abs(hex.s)) / 2)
}

func hex_distance(a, b Hex) int {
	return hex_length(hex_subtract(a, b))
}

func hex_round(h FractionalHex) Hex {
	qi, ri, si := int(round(h.q)), int(round(h.r)), int(round(h.s))
	q_diff, r_diff, s_diff := abs(float64(qi)-h.q), abs(float64(ri)-h.r), abs(float64(si)-h.s)
	if q_diff > r_diff && q_diff > s_diff {
		qi = -ri - si
	} else if r_diff > s_diff {
		ri = -qi - si
	} else {
		si = -qi - ri
	}
	return Hex{q: qi, r: ri, s: si}
}

func hex_lerp(a, b FractionalHex, t float64) FractionalHex {
	return FractionalHex{q: a.q*(1.0-t) + b.q*t, r: a.r*(1.0-t) + b.r*t, s: a.s*(1.0-t) + b.s*t}
}

func hex_linedraw(a, b Hex) []Hex {
	N := hex_distance(a, b)
	a_nudge := FractionalHex{q: float64(a.q) + 1e-06, r: float64(a.r) + 1e-06, s: float64(a.s) - 2e-06}
	b_nudge := FractionalHex{q: float64(b.q) + 1e-06, r: float64(b.r) + 1e-06, s: float64(b.s) - 2e-06}
	results := []Hex{}
	step := 1.0 / max(float64(N), 1.0)
	for i := 0; i <= N; i++ {
		results = append(results, hex_round(hex_lerp(a_nudge, b_nudge, step*float64(i))))
	}
	return results
}

const EVEN = 1

const ODD = -1

func qoffset_from_cube(offset int, h Hex) OffsetCoord {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.q & 1
	col, row := h.q, h.r+int((h.q+offset*parity)/2)
	return OffsetCoord{col: col, row: row}
}

func qoffset_to_cube(offset int, h OffsetCoord) Hex {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.col & 1
	q, r := h.col, h.row-int((h.col+offset*parity)/2)
	return Hex{q: q, r: r, s: -q - r}
}

func roffset_from_cube(offset int, h Hex) OffsetCoord {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.r & 1
	col, row := h.q+int((h.r+offset*parity)/2), h.r
	return OffsetCoord{col: col, row: row}
}

func roffset_to_cube(offset int, h OffsetCoord) Hex {
	if offset != EVEN && offset != ODD {
		panic("offset must be EVEN (+1) or ODD (-1)")
	}
	parity := h.row & 1
	q, r := h.col-int((h.row+offset*parity)/2), h.row
	return Hex{q: q, r: r, s: -q - r}
}

func qoffset_from_qfloat64d(offset int, h DoubledCoord) OffsetCoord {
	parity := h.col & 1
	return OffsetCoord{col: h.col, row: int((h.row + offset*parity) / 2)}
}

func qoffset_to_qfloat64d(offset int, h OffsetCoord) DoubledCoord {
	parity := h.col & 1
	return DoubledCoord{col: h.col, row: 2*h.row - offset*parity}
}

func roffset_from_rfloat64d(offset int, h DoubledCoord) OffsetCoord {
	parity := h.row & 1
	return OffsetCoord{col: int((h.col + offset*parity) / 2), row: h.row}
}

func roffset_to_rfloat64d(offset int, h OffsetCoord) DoubledCoord {
	parity := h.row & 1
	return DoubledCoord{col: 2*h.col - offset*parity, row: h.row}
}

func qfloat64d_from_cube(h Hex) DoubledCoord {
	col, row := h.q, 2*h.r+h.q
	return DoubledCoord{col: col, row: row}
}

func qfloat64d_to_cube(h DoubledCoord) Hex {
	q, r := h.col, int((h.row-h.col)/2)
	return Hex{q: q, r: r, s: -q - r}
}

func rfloat64d_from_cube(h Hex) DoubledCoord {
	col, row := 2*h.q+h.r, h.r
	return DoubledCoord{col: col, row: row}
}

func rfloat64d_to_cube(h DoubledCoord) Hex {
	q, r := int((h.col-h.row)/2), h.row
	return Hex{q: q, r: r, s: -q - r}
}

var layout_pointy = Orientation{math.Sqrt(3.0), math.Sqrt(3.0) / 2.0, 0.0, 3.0 / 2.0, math.Sqrt(3.0) / 3.0, -1.0 / 3.0, 0.0, 2.0 / 3.0, 0.5}

var layout_flat = Orientation{3.0 / 2.0, 0.0, math.Sqrt(3.0) / 2.0, math.Sqrt(3.0), 2.0 / 3.0, 0.0, -1.0 / 3.0, math.Sqrt(3.0) / 3.0, 0.0}

func hex_to_pixel(layout Layout, h Hex) Point {
	x := (layout.orientation.f0*float64(h.q) + layout.orientation.f1*float64(h.r)) * layout.size.x
	y := (layout.orientation.f2*float64(h.q) + layout.orientation.f3*float64(h.r)) * layout.size.y
	return Point{x: x + layout.origin.x, y: y + layout.origin.y}
}

func pixel_to_hex_fractional(layout Layout, p Point) FractionalHex {
	pt := Point{x: (p.x - layout.origin.x) / layout.size.x, y: (p.y - layout.origin.y) / layout.size.y}
	q := layout.orientation.b0*pt.x + layout.orientation.b1*pt.y
	r := layout.orientation.b2*pt.x + layout.orientation.b3*pt.y
	return FractionalHex{q: q, r: r, s: -q - r}
}

func pixel_to_hex_rounded(layout Layout, p Point) Hex {
	return hex_round(pixel_to_hex_fractional(layout, p))
}

func hex_corner_offset(layout Layout, corner int) Point {
	angle := 2.0 * math.Pi * (layout.orientation.start_angle - float64(corner)) / 6.0
	return Point{x: layout.size.x * math.Cos(angle), y: layout.size.y * math.Sin(angle)}
}

func polygon_corners(layout Layout, h Hex) []Point {
	corners := make([]Point, 6)
	center := hex_to_pixel(layout, h)
	for i := 0; i < 6; i++ {
		offset := hex_corner_offset(layout, i)
		corners[i] = Point{x: center.x + offset.x, y: center.y + offset.y}
	}
	return corners
}

// helpers for math
//

// number is a constraint that permits any integer or floating-point type.
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
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
