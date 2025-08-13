// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import (
	"log"
	"testing"
)

func complain(name string) {
	log.Printf("FAIL %s\n", name)
}

func equal_hex(name string, a, b CubeCoord) bool {
	if !a.Equals(b) {
		complain(name)
		return false
	}
	return true
}

func equal_offsetcoord(name string, a, b OffsetCoord) bool {
	if !a.Equals(b) {
		complain(name)
		return false
	}
	return true
}

func equal_doubledcoord(name string, a, b DoubledCoord) bool {
	if !a.Equals(b) {
		complain(name)
		return false
	}
	return true
}

func equal_int(name string, a, b int) bool {
	if !(a == b) {
		complain(name)
		return false
	}
	return true
}

func equal_hex_array(name string, a, b []CubeCoord) bool {
	if !equal_int(name, len(a), len(b)) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !equal_hex(name, a[i], b[i]) {
			return false
		}
	}
	return true
}

func Test_hex_arithmetic(t *testing.T) {
	if !equal_hex("cube_add", CubeCoord{4, -10, 6}, CubeCoord{1, -3, 2}.Add(CubeCoord{3, -7, 4})) {
		t.Fatalf("cube_add")
	}
	if !equal_hex("cube_subtract", CubeCoord{-2, 4, -2}, CubeCoord{1, -3, 2}.Subtract(CubeCoord{3, -7, 4})) {
		t.Fatalf("cube_subtract")
	}
}

func Test_hex_direction(t *testing.T) {
	if !equal_hex("cube_direction", CubeCoord{0, -1, 1}, cube_direction(2)) {
		t.Fatalf("cube_direction")
	}
}

func test_hex_neighbor(t *testing.T) {
	if !equal_hex("cube_neighbor", CubeCoord{1, -3, 2}, CubeCoord{1, -2, 1}.Neighbor(2)) {
		t.Fatalf("cube_neighbor")
	}
}

func Test_hex_diagonal(t *testing.T) {
	if !equal_hex("hex_diagonal", CubeCoord{-1, -1, 2}, CubeCoord{1, -2, 1}.DiagonalNeighbor(3)) {
		t.Fatalf("hex_diagonal")
	}
}

func Test_hex_distance(t *testing.T) {
	if !equal_int("cube_distance", 7, CubeCoord{3, -7, 4}.Distance(CubeCoord{0, 0, 0})) {
		t.Fatalf("cube_distance")
	}
}

func Test_hex_rotate_right(t *testing.T) {
	if !equal_hex("cube_rotate_right", CubeCoord{1, -3, 2}.RotateRight(), CubeCoord{3, -2, -1}) {
		t.Fatalf("cube_rotate_right")
	}
}

func Test_hex_rotate_left(t *testing.T) {
	if !equal_hex("cube_rotate_left", CubeCoord{1, -3, 2}.RotateLeft(), CubeCoord{-2, -1, 3}) {
		t.Fatalf("cube_rotate_left")
	}
}

func Test_hex_round(t *testing.T) {
	a := FractionalCubeCoord{0.0, 0.0, 0.0}
	b := FractionalCubeCoord{1.0, -1.0, 0.0}
	c := FractionalCubeCoord{0.0, -1.0, 1.0}
	equal_hex("cube_round 1", CubeCoord{5, -10, 5}, FractionalCubeCoord{0.0, 0.0, 0.0}.Lerp(FractionalCubeCoord{10.0, -20.0, 10.0}, 0.5).Round())
	equal_hex("cube_round 2", a.Round(), a.Lerp(b, 0.499).Round())
	equal_hex("cube_round 3", b.Round(), a.Lerp(b, 0.501).Round())
	equal_hex("cube_round 4", a.Round(), FractionalCubeCoord{a.q*0.4 + b.q*0.3 + c.q*0.3, a.r*0.4 + b.r*0.3 + c.r*0.3, a.s*0.4 + b.s*0.3 + c.s*0.3}.Round())
	equal_hex("cube_round 5", c.Round(), FractionalCubeCoord{a.q*0.3 + b.q*0.3 + c.q*0.4, a.r*0.3 + b.r*0.3 + c.r*0.4, a.s*0.3 + b.s*0.3 + c.s*0.4}.Round())
}

func Test_hex_linedraw(t *testing.T) {
	equal_hex_array("cube_linedraw", []CubeCoord{CubeCoord{0, 0, 0}, CubeCoord{0, -1, 1}, CubeCoord{0, -2, 2}, CubeCoord{1, -3, 2}, CubeCoord{1, -4, 3}, CubeCoord{1, -5, 4}}, CubeCoord{0, 0, 0}.Linedraw(CubeCoord{1, -5, 4}))
}

func Test_layout(t *testing.T) {
	h := CubeCoord{3, 4, -7}
	flat := Layout{layout_flat, Point{10.0, 15.0}, Point{35.0, 71.0}}
	equal_hex("layout", h, pixel_to_cube_rounded(flat, cube_to_pixel(flat, h)))
	pointy := Layout{layout_pointy, Point{10.0, 15.0}, Point{35.0, 71.0}}
	equal_hex("layout", h, pixel_to_cube_rounded(pointy, cube_to_pixel(pointy, h)))
}

func Test_offset_roundtrip(t *testing.T) {
	for q := -2; q < 3; q++ {
		for r := -2; r < 3; r++ {
			cube := CubeCoord{q, r, -q - r}
			equal_hex("conversion_roundtrip odd-q", cube, qoffset_to_cube(ODD, qoffset_from_cube(ODD, cube)))
			equal_hex("conversion_roundtrip odd-r", cube, roffset_to_cube(ODD, roffset_from_cube(ODD, cube)))
			equal_hex("conversion_roundtrip even-q", cube, qoffset_to_cube(EVEN, qoffset_from_cube(EVEN, cube)))
			equal_hex("conversion_roundtrip even-r", cube, roffset_to_cube(EVEN, roffset_from_cube(EVEN, cube)))
		}
	}
	for col := -2; col < 3; col++ {
		for row := -2; row < 3; row++ {
			offset := OffsetCoord{col, row}
			equal_offsetcoord("conversion_roundtrip odd-q", offset, qoffset_from_cube(ODD, qoffset_to_cube(ODD, offset)))
			equal_offsetcoord("conversion_roundtrip odd-r", offset, roffset_from_cube(ODD, roffset_to_cube(ODD, offset)))
			equal_offsetcoord("conversion_roundtrip even-q", offset, qoffset_from_cube(EVEN, qoffset_to_cube(EVEN, offset)))
			equal_offsetcoord("conversion_roundtrip even-r", offset, roffset_from_cube(EVEN, roffset_to_cube(EVEN, offset)))
		}
	}
}

func Test_offset_from_cube(t *testing.T) {
	equal_offsetcoord("offset_from_cube odd-r", OffsetCoord{-2, 2}, roffset_from_cube(ODD, CubeCoord{-3, 2, 1}))
	equal_offsetcoord("offset_from_cube odd-r", OffsetCoord{1, -1}, roffset_from_cube(ODD, CubeCoord{2, -1, -1}))
	equal_offsetcoord("offset_from_cube even-r", OffsetCoord{-2, 2}, roffset_from_cube(EVEN, CubeCoord{-3, 2, 1}))
	equal_offsetcoord("offset_from_cube even-r", OffsetCoord{2, -1}, roffset_from_cube(EVEN, CubeCoord{2, -1, -1}))
	equal_offsetcoord("offset_from_cube odd-q", OffsetCoord{-2, 2}, qoffset_from_cube(ODD, CubeCoord{-2, 3, -1}))
	equal_offsetcoord("offset_from_cube odd-q", OffsetCoord{-1, -2}, qoffset_from_cube(ODD, CubeCoord{-1, -1, 2}))
	equal_offsetcoord("offset_from_cube even-q", OffsetCoord{-2, 2}, qoffset_from_cube(EVEN, CubeCoord{-2, 3, -1}))
	equal_offsetcoord("offset_from_cube even-q", OffsetCoord{-1, -1}, qoffset_from_cube(EVEN, CubeCoord{-1, -1, 2}))
}

func Test_offset_to_cube(t *testing.T) {
	equal_hex("offset_to_cube odd-r", CubeCoord{-3, 2, 1}, roffset_to_cube(ODD, OffsetCoord{-2, 2}))
	equal_hex("offset_to_cube odd-r", CubeCoord{2, -1, -1}, roffset_to_cube(ODD, OffsetCoord{1, -1}))
	equal_hex("offset_to_cube even-r", CubeCoord{-3, 2, 1}, roffset_to_cube(EVEN, OffsetCoord{-2, 2}))
	equal_hex("offset_to_cube even-r", CubeCoord{2, -1, -1}, roffset_to_cube(EVEN, OffsetCoord{2, -1}))
	equal_hex("offset_to_cube odd-q", CubeCoord{-2, 3, -1}, qoffset_to_cube(ODD, OffsetCoord{-2, 2}))
	equal_hex("offset_to_cube odd-q", CubeCoord{-1, -1, 2}, qoffset_to_cube(ODD, OffsetCoord{-1, -2}))
	equal_hex("offset_to_cube even-q", CubeCoord{-2, 3, -1}, qoffset_to_cube(EVEN, OffsetCoord{-2, 2}))
	equal_hex("offset_to_cube even-q", CubeCoord{-1, -1, 2}, qoffset_to_cube(EVEN, OffsetCoord{-1, -1}))
}

func Test_offset_to_float64d(t *testing.T) {
	for col := -2; col < 3; col++ {
		for row := -2; row < 3; row++ {
			offset := OffsetCoord{col, row}
			equal_doubledcoord("offset_to_float64d loop odd-q", qfloat64d_from_cube(qoffset_to_cube(ODD, offset)), qoffset_to_qfloat64d(ODD, offset))
			equal_doubledcoord("offset_to_float64d loop even-q", qfloat64d_from_cube(qoffset_to_cube(EVEN, offset)), qoffset_to_qfloat64d(EVEN, offset))
			equal_doubledcoord("offset_to_float64d loop odd-r", rfloat64d_from_cube(roffset_to_cube(ODD, offset)), roffset_to_rfloat64d(ODD, offset))
			equal_doubledcoord("offset_to_float64d loop even-r", rfloat64d_from_cube(roffset_to_cube(EVEN, offset)), roffset_to_rfloat64d(EVEN, offset))
			qfloat64d := DoubledCoord{col*2 + (row & 1), row}
			equal_offsetcoord("offset_from_float64d loop odd-q", qoffset_from_cube(ODD, qfloat64d_to_cube(qfloat64d)), qoffset_from_qfloat64d(ODD, qfloat64d))
			equal_offsetcoord("offset_from_float64d loop even-q", qoffset_from_cube(EVEN, qfloat64d_to_cube(qfloat64d)), qoffset_from_qfloat64d(EVEN, qfloat64d))
			rfloat64d := DoubledCoord{col, row*2 + (col & 1)}
			equal_offsetcoord("offset_from_float64d loop odd-r", roffset_from_cube(ODD, rfloat64d_to_cube(rfloat64d)), roffset_from_rfloat64d(ODD, rfloat64d))
			equal_offsetcoord("offset_from_float64d loop even-r", roffset_from_cube(EVEN, rfloat64d_to_cube(rfloat64d)), roffset_from_rfloat64d(EVEN, rfloat64d))
		}
	}
}

func Test_offset_from_float64d(t *testing.T) {
}

func Test_float64d_roundtrip(t *testing.T) {
	for q := -2; q < 3; q++ {
		for r := -2; r < 3; r++ {
			cube := CubeCoord{q, r, -q - r}
			equal_hex("conversion_roundtrip float64d-q", cube, qfloat64d_to_cube(qfloat64d_from_cube(cube)))
			equal_hex("conversion_roundtrip float64d-r", cube, rfloat64d_to_cube(rfloat64d_from_cube(cube)))
		}
	}
	for col := -2; col < 3; col++ {
		for row := -2; row < 3; row++ {
			qfloat64d := DoubledCoord{col*2 + (row & 1), row}
			equal_doubledcoord("conversion_roundtrip float64d-q", qfloat64d, qfloat64d_from_cube(qfloat64d_to_cube(qfloat64d)))
			rfloat64d := DoubledCoord{col, row*2 + (col & 1)}
			equal_doubledcoord("conversion_roundtrip float64d-r", rfloat64d, rfloat64d_from_cube(rfloat64d_to_cube(rfloat64d)))
		}
	}
}

func Test_float64d_from_cube(t *testing.T) {
	equal_doubledcoord("float64d_from_cube float64d-q", DoubledCoord{1, 5}, qfloat64d_from_cube(CubeCoord{1, 2, -3}))
	equal_doubledcoord("float64d_from_cube float64d-r", DoubledCoord{4, 2}, rfloat64d_from_cube(CubeCoord{1, 2, -3}))
}

func Test_float64d_to_cube(t *testing.T) {
	equal_hex("float64d_to_cube float64d-q", CubeCoord{1, 2, -3}, qfloat64d_to_cube(DoubledCoord{1, 5}))
	equal_hex("float64d_to_cube float64d-r", CubeCoord{1, 2, -3}, rfloat64d_to_cube(DoubledCoord{4, 2}))
}
