// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

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
