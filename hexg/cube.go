// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import "fmt"

type CubeCoord struct {
	q int
	r int
	s int
}

func NewCubeCoord(q_, r_, s_ int) CubeCoord {
	if q_+r_+s_ != 0 {
		panic("q + r + s must be 0")
	}
	return CubeCoord{q: q_, r: r_, s: s_}
}

type FractionalCubeCoord struct {
	q float64
	r float64
	s float64
}

func NewFractionalCubeCoord(q_, r_, s_ float64) FractionalCubeCoord {
	if round(q_+r_+s_) != 0 {
		panic("q + r + s must be 0")
	}
	return FractionalCubeCoord{q: q_, r: r_, s: s_}
}

func (a CubeCoord) Add(b CubeCoord) CubeCoord {
	return CubeCoord{q: a.q + b.q, r: a.r + b.r, s: a.s + b.s}
}

func (a CubeCoord) Subtract(b CubeCoord) CubeCoord {
	return CubeCoord{q: a.q - b.q, r: a.r - b.r, s: a.s - b.s}
}

func (a CubeCoord) Equals(b CubeCoord) bool {
	return a.q == b.q && a.r == b.r && a.s == b.s
}

func (a CubeCoord) IsZero() bool {
	return a.q == 0 && a.r == 0 && a.s == 0
}

func (a CubeCoord) Scale(k int) CubeCoord {
	return CubeCoord{q: a.q * k, r: a.r * k, s: a.s * k}
}

func (a CubeCoord) RotateLeft() CubeCoord {
	return CubeCoord{q: -a.s, r: -a.q, s: -a.r}
}

func (a CubeCoord) RotateRight() CubeCoord {
	return CubeCoord{q: -a.r, r: -a.s, s: -a.q}
}

var cube_directions = []CubeCoord{
	{q: 1, r: 0, s: -1},
	{q: 1, r: -1, s: 0},
	{q: 0, r: -1, s: 1},
	{q: -1, r: 0, s: 1},
	{q: -1, r: 1, s: 0},
	{q: 0, r: 1, s: -1},
}

func cube_direction(direction int) CubeCoord {
	return cube_directions[(6+(direction%6))%6]
}

func (hex CubeCoord) Neighbor(direction int) CubeCoord {
	return hex.Add(cube_direction(direction))
}

var cube_diagonals = []CubeCoord{
	{q: 2, r: -1, s: -1},
	{q: 1, r: -2, s: 1},
	{q: -1, r: -1, s: 2},
	{q: -2, r: 1, s: 1},
	{q: -1, r: 2, s: -1},
	{q: 1, r: 1, s: -2},
}

func (hex CubeCoord) DiagonalNeighbor(direction int) CubeCoord {
	return hex.Add(cube_diagonals[(6+(direction%6))%6])
}

func (hex CubeCoord) Length() int {
	return (abs(hex.q) + abs(hex.r) + abs(hex.s)) / 2
}

func (a CubeCoord) Distance(b CubeCoord) int {
	return a.Subtract(b).Length()
}

func (h FractionalCubeCoord) Round() CubeCoord {
	qi, ri, si := int(round(h.q)), int(round(h.r)), int(round(h.s))
	q_diff, r_diff, s_diff := abs(float64(qi)-h.q), abs(float64(ri)-h.r), abs(float64(si)-h.s)
	if q_diff > r_diff && q_diff > s_diff {
		qi = -ri - si
	} else if r_diff > s_diff {
		ri = -qi - si
	} else {
		si = -qi - ri
	}
	return CubeCoord{q: qi, r: ri, s: si}
}

func (a FractionalCubeCoord) Lerp(b FractionalCubeCoord, t float64) FractionalCubeCoord {
	return FractionalCubeCoord{q: a.q*(1.0-t) + b.q*t, r: a.r*(1.0-t) + b.r*t, s: a.s*(1.0-t) + b.s*t}
}

func (a CubeCoord) Linedraw(b CubeCoord) []CubeCoord {
	N := a.Distance(b)
	a_nudge := FractionalCubeCoord{q: float64(a.q) + 1e-06, r: float64(a.r) + 1e-06, s: float64(a.s) - 2e-06}
	b_nudge := FractionalCubeCoord{q: float64(b.q) + 1e-06, r: float64(b.r) + 1e-06, s: float64(b.s) - 2e-06}
	results := []CubeCoord{}
	step := 1.0 / max(float64(N), 1.0)
	for i := 0; i <= N; i++ {
		results = append(results, a_nudge.Lerp(b_nudge, step*float64(i)).Round())
	}
	return results
}

// String returns the coordinates with signs.
// It returns the coordinates formatted as "+q+r+s".
func (h CubeCoord) String() string {
	return fmt.Sprintf("%+d%+d%+d", h.q, h.r, h.s)
}

//// String implements the Stringer interface.
//// It returns the coordinates formatted as (q,r,s).
//func (h CubeCoord) String() string {
//	return fmt.Sprintf("%d,%d,%d", h.q, h.r, h.s)
//}
