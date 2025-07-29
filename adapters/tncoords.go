// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package adapters

import (
	"github.com/maloquacious/hexg"
	"log"
)

var (
	layout = hexg.NewTribeNetLayout()
)

// ConvertTNCoords translates a coordinate like "AB 0102" to a Hex.
func ConvertTNCoords(input string) (hex hexg.Hex, err error) {
	return layout.TribeNetCoordToHex(input)
}

type MapData struct {
	Tiles  map[uint64]*Tile_t
	Icons  []*Icon_t  // not implemented
	Labels []*Label_t // not implementd
	Notes  []*Note_t  // not implemnted

	// origin is the upper-left corner of the map.
	// we could calculate it on the fly, but that is expensive.
	Origin hexg.Hex
}

type Tile_t struct {
	Coords struct {
		TribeNet string
		Wxx      hexg.OffsetCoord // column + row in file
		Hex      hexg.Hex
	}
}

type Icon_t struct{}

type Label_t struct{}

type Note_t struct{}

// MergeMaps returns a MapData that bounds all the provided maps.
func MergeMaps(maps ...MapData) MapData {
	if len(maps) == 0 {
		return MapData{}
	}

	// collect the origins to get the top left hex
	var hexes []hexg.Hex
	for _, m := range maps {
		hexes = append(hexes, m.Origin)
	}

	// determine the top left hex
	topLeftHex := hexg.TopLeftHex(layout)
	log.Printf("merge: top left %q\n", topLeftHex)

	return MapData{}
}

// Bounds returns the top-left and bottom-right axial coordinates of the map.
// These are inclusive bounds.
func (m MapData) Bounds() (topLeft, bottomRight hexg.Hex) {
	topLeft = m.Origin
	bottomRight = hexg.Hex{} // m.Origin.Add(5, 6)
	return topLeft, bottomRight
}

func (m MapData) Corners() (topLeft, topRight, bottomLeft, bottomRight hexg.Hex) {
	panic("!implemented")
}
