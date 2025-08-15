// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

import "fmt"

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

// EvenQCoord implements "even-q," an offset coordinate with flat top hexes and even columns pushed down.
type EvenQCoord struct {
	col int
	row int
}

func NewEvenQCoord(col_, row_ int) EvenQCoord {
	return EvenQCoord{col: col_, row: row_}
}

func (a EvenQCoord) Equals(b EvenQCoord) bool {
	return a.col == b.col && a.row == b.row
}

// EvenRCoord implements "even-r," an offset coordinate with pointy top hexes and even rows pushed right.
type EvenRCoord struct {
	col int
	row int
}

func NewEvenRCoord(col_, row_ int) EvenRCoord {
	return EvenRCoord{col: col_, row: row_}
}

func (a EvenRCoord) Equals(b EvenRCoord) bool {
	return a.col == b.col && a.row == b.row
}

// OddQCoord implements "odd-q," an offset coordinate with flat top hexes and odd columns pushed down.
type OddQCoord struct {
	col int
	row int
}

func NewOddQCoord(col_, row_ int) OddQCoord {
	return OddQCoord{col: col_, row: row_}
}

func (a OddQCoord) Equals(b OddQCoord) bool {
	return a.col == b.col && a.row == b.row
}

// OddRCoord implements "odd-r," an offset coordinate with pointy top hexes and odd rows pushed right.
type OddRCoord struct {
	col int
	row int
}

func NewOddRCoord(col_, row_ int) OddRCoord {
	return OddRCoord{col: col_, row: row_}
}

func (a OddRCoord) Equals(b OddRCoord) bool {
	return a.col == b.col && a.row == b.row
}

// Orientation_e is orientation for offset coordinates
type Orientation_e int

const (
	UnknownQR Orientation_e = iota
	EvenQ                   // vertical layout shoves even columns down
	OddQ                    // vertical layout shoves odd columns down
	EvenR                   // horizontal layout shoves even rows right
	OddR                    // horizontal layout shoves odd rows right
)

// String implements the fmt.Stringer interface
func (o Orientation_e) String() string {
	switch o {
	case UnknownQR:
		return "unknown-qr"
	case EvenQ:
		return "even-q"
	case OddQ:
		return "odd-q"
	case EvenR:
		return "even-r"
	case OddR:
		return "odd-r"
	default:
		return fmt.Sprintf("Orientation_e(%d)", int(o))
	}
}

func (o Orientation_e) IsColumns() bool {
	return o == EvenQ || o == OddQ
}

func (o Orientation_e) IsHorizontal() bool {
	return o == EvenR || o == OddR
}

func (o Orientation_e) IsFlatTop() bool {
	return o == EvenQ || o == OddQ
}

func (o Orientation_e) IsPointyTop() bool {
	return o == EvenR || o == OddR
}

func (o Orientation_e) IsRows() bool {
	return o == EvenR || o == OddR
}

func (o Orientation_e) IsVertical() bool {
	return o == EvenQ || o == OddQ
}
