// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

type Point struct {
	x float64
	y float64
}

func NewPoint(x_, y_ float64) Point {
	return Point{x: x_, y: y_}
}
