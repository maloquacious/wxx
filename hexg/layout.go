// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import "math"

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
