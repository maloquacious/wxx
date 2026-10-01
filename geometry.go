// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"fmt"
)

// Shape coordinates
//
// Shapes, labels, features, notes and <extraTerrain> placements are positioned
// in the map's own coordinate space, not by tile (issue #153). x grows to the
// right and y grows down, and the origin is the top left of the map.
//
// In that space a hex is 300 units wide and 300 high, whatever the map's
// hexWidth and hexHeight say. Every tracked fixture states the suggested hex
// size (46.18 x 40 for COLUMNS, 40 x 46.18 for ROWS), so whether the space
// scales with hexWidth and hexHeight has not been seen; 300 is what the
// fixtures show.
//
// COLUMNS (flat-top) hexes are 225 apart across and 300 apart down, and odd
// columns are 150 lower. Hex (col, row) has its center at
//
//	(150 + 225*col, 150 + 300*row + 150*odd(col))
//
// and its corners at (+-150, 0) and (+-75, +-150) from the center. This fits
// every COLUMNS fixture: the tile-border polygon at (12, 0) in
// testdata/2025-2.08-13x11-941577-notes-shapes.wxx, and every point of the two
// lines in testdata/2025-2.08-13x11-941577-river-lines.wxx (issue #154).
//
// ROWS (pointy-top) hexes are the same hex turned a quarter: 300 apart across,
// 225 apart down, and odd rows are 150 to the right. Hex (col, row) has its
// center at
//
//	(150 + 300*col + 150*odd(row), 150 + 225*row)
//
// and its corners at (0, +-150) and (+-150, +-75) from the center. The center
// was confirmed in Worldographer on a 2.08 ROWS map (#80). No tracked ROWS
// fixture has a shape, so the corners are the COLUMNS corners transposed, not
// yet seen in a file.
//
// Every center and corner is a whole multiple of 75, so it is exact in a
// float64 and can be compared with ==.

// Position_t is a point in shape coordinates.
type Position_t struct {
	X float64
	Y float64
}

// Corner_e names a corner of a hex. Each orientation uses six of the eight
// names: COLUMNS (flat-top) hexes have e, se, sw, w, nw and ne, and ROWS
// (pointy-top) hexes have n, ne, se, s, sw and nw. The names are compass
// directions from the hex's center, with north up the screen, so ne is up and
// to the right in both orientations, though it is a different point.
type Corner_e int

const (
	CornerE Corner_e = iota + 1
	CornerSE
	CornerS
	CornerSW
	CornerW
	CornerNW
	CornerN
	CornerNE
)

// String returns the corner's short name: "e", "se", "s", "sw", "w", "nw",
// "n" or "ne".
func (c Corner_e) String() string {
	switch c {
	case CornerE:
		return "e"
	case CornerSE:
		return "se"
	case CornerS:
		return "s"
	case CornerSW:
		return "sw"
	case CornerW:
		return "w"
	case CornerNW:
		return "nw"
	case CornerN:
		return "n"
	case CornerNE:
		return "ne"
	}
	return fmt.Sprintf("Corner_e(%d)", int(c))
}

// ParseCorner returns the corner a short name names, as String spells it.
func ParseCorner(name string) (Corner_e, error) {
	for c := CornerE; c <= CornerNE; c++ {
		if c.String() == name {
			return c, nil
		}
	}
	return 0, errors.Join(ErrInvalidCorner, fmt.Errorf("corner %q: want one of e, se, s, sw, w, nw, n, ne", name))
}

// Vertex_t is a hex corner named by one of the hexes it belongs to. Three
// hexes share most corners, and each names it differently: in a COLUMNS map,
// hex (4, 2)'s se is (5, 2)'s w and (4, 3)'s ne. CanonicalVertex picks one of
// the names.
type Vertex_t struct {
	Col    int
	Row    int
	Corner Corner_e
}

func (v Vertex_t) String() string {
	return fmt.Sprintf("(%d,%d) %s", v.Col, v.Row, v.Corner)
}

// hexGeometry_t is the geometry of one orientation.
type hexGeometry_t struct {
	orientation string
	// corners are the orientation's six corners in clockwise order, and
	// offsets their positions from the hex's center, in the same order.
	corners [6]Corner_e
	offsets [6]Position_t
	// canonical are the two corners CanonicalVertex names a corner by. Every
	// corner of the grid is exactly one hex's canonical[0] or canonical[1].
	canonical [2]Corner_e
	// center returns a hex's center.
	center func(col, row int) Position_t
	// neighbors returns the six hexes that share an edge with (col, row).
	neighbors func(col, row int) [6][2]int
}

var columnsHexGeometry = hexGeometry_t{
	orientation: "COLUMNS",
	corners:     [6]Corner_e{CornerE, CornerSE, CornerSW, CornerW, CornerNW, CornerNE},
	offsets:     [6]Position_t{{150, 0}, {75, 150}, {-75, 150}, {-150, 0}, {-75, -150}, {75, -150}},
	canonical:   [2]Corner_e{CornerE, CornerSE},
	center: func(col, row int) Position_t {
		return Position_t{X: 150 + 225*float64(col), Y: 150 + 300*float64(row) + 150*float64(col&1)}
	},
	// odd-q: an odd column is half a hex lower than its neighbors.
	neighbors: func(col, row int) [6][2]int {
		p := col & 1
		return [6][2]int{
			{col, row - 1}, {col + 1, row - 1 + p}, {col + 1, row + p},
			{col, row + 1}, {col - 1, row + p}, {col - 1, row - 1 + p},
		}
	},
}

var rowsHexGeometry = hexGeometry_t{
	orientation: "ROWS",
	corners:     [6]Corner_e{CornerN, CornerNE, CornerSE, CornerS, CornerSW, CornerNW},
	offsets:     [6]Position_t{{0, -150}, {150, -75}, {150, 75}, {0, 150}, {-150, 75}, {-150, -75}},
	canonical:   [2]Corner_e{CornerS, CornerSE},
	center: func(col, row int) Position_t {
		return Position_t{X: 150 + 300*float64(col) + 150*float64(row&1), Y: 150 + 225*float64(row)}
	},
	// odd-r: an odd row is half a hex to the right of its neighbors.
	neighbors: func(col, row int) [6][2]int {
		p := row & 1
		return [6][2]int{
			{col + 1, row}, {col + p, row + 1}, {col - 1 + p, row + 1},
			{col - 1, row}, {col - 1 + p, row - 1}, {col + p, row - 1},
		}
	},
}

// hexGeometry returns the geometry of m's orientation.
func (m *Map_t) hexGeometry() (*hexGeometry_t, error) {
	switch m.HexOrientation {
	case "COLUMNS":
		return &columnsHexGeometry, nil
	case "ROWS":
		return &rowsHexGeometry, nil
	}
	return nil, errors.Join(ErrInvalidHexOrientation, fmt.Errorf("hexOrientation %q: want \"COLUMNS\" or \"ROWS\"", m.HexOrientation))
}

// cornerIndex returns c's index in g.corners, or an error if c is not a corner
// of g's orientation.
func (g *hexGeometry_t) cornerIndex(c Corner_e) (int, error) {
	for i, gc := range g.corners {
		if gc == c {
			return i, nil
		}
	}
	return 0, errors.Join(ErrInvalidCorner, fmt.Errorf("corner %s is not a corner of a %s hex: want one of %v", c, g.orientation, g.corners))
}

// corner returns the position of corner i (an index into g.corners) of hex
// (col, row).
func (g *hexGeometry_t) corner(col, row, i int) Position_t {
	c, o := g.center(col, row), g.offsets[i]
	return Position_t{X: c.X + o.X, Y: c.Y + o.Y}
}

// Corners returns the six corners of a hex in m's HexOrientation, clockwise
// (y down): e, se, sw, w, nw, ne for COLUMNS, and n, ne, se, s, sw, nw for
// ROWS. The error is ErrInvalidHexOrientation as for TileCenter.
func (m *Map_t) Corners() ([6]Corner_e, error) {
	g, err := m.hexGeometry()
	if err != nil {
		return [6]Corner_e{}, err
	}
	return g.corners, nil
}

// TileCenter returns the center of hex (col, row) in shape coordinates, for m's
// HexOrientation. col and row may be outside the map: a hex just off the map
// has a center too. The error is ErrInvalidHexOrientation when m's orientation
// is neither "COLUMNS" nor "ROWS".
func (m *Map_t) TileCenter(col, row int) (Position_t, error) {
	g, err := m.hexGeometry()
	if err != nil {
		return Position_t{}, err
	}
	return g.center(col, row), nil
}

// TileCorner returns corner c of hex (col, row) in shape coordinates, for m's
// HexOrientation. col and row may be outside the map, so a path can run along
// the map's edge. The error is ErrInvalidCorner when c is not a corner of m's
// orientation (e and w are not corners of a ROWS hex, n and s not of a COLUMNS
// one), and ErrInvalidHexOrientation as for TileCenter.
func (m *Map_t) TileCorner(col, row int, c Corner_e) (Position_t, error) {
	g, err := m.hexGeometry()
	if err != nil {
		return Position_t{}, err
	}
	i, err := g.cornerIndex(c)
	if err != nil {
		return Position_t{}, errors.Join(err, fmt.Errorf("hex (%d,%d)", col, row))
	}
	return g.corner(col, row, i), nil
}

// VertexPosition returns v's position: TileCorner(v.Col, v.Row, v.Corner).
func (m *Map_t) VertexPosition(v Vertex_t) (Position_t, error) {
	return m.TileCorner(v.Col, v.Row, v.Corner)
}

// sharers returns every (hex, corner index) that names the corner at p, which
// must be a corner of hex (col, row): three entries, since every corner of the
// grid belongs to three hexes. They are found among (col, row) and its
// neighbors by position, so they agree with TileCorner by construction.
func (g *hexGeometry_t) sharers(col, row int, p Position_t) []Vertex_t {
	neighbors := g.neighbors(col, row)
	hexes := append([][2]int{{col, row}}, neighbors[:]...)
	var out []Vertex_t
	for _, h := range hexes {
		for i, c := range g.corners {
			if g.corner(h[0], h[1], i) == p {
				out = append(out, Vertex_t{Col: h[0], Row: h[1], Corner: c})
			}
		}
	}
	return out
}

// CanonicalVertex returns the one name CanonicalVertex gives the corner v
// names, whichever of its three hexes v names it by. For a COLUMNS map that is
// a hex's e or se corner, and for a ROWS map a hex's s or se corner: every
// corner of the grid is exactly one hex's e or se (s or se). So two vertices
// are the same corner when their canonical vertices are equal; SameCorner says
// so directly. The errors are TileCorner's.
func (m *Map_t) CanonicalVertex(v Vertex_t) (Vertex_t, error) {
	g, err := m.hexGeometry()
	if err != nil {
		return Vertex_t{}, err
	}
	i, err := g.cornerIndex(v.Corner)
	if err != nil {
		return Vertex_t{}, errors.Join(err, fmt.Errorf("hex (%d,%d)", v.Col, v.Row))
	}
	for _, s := range g.sharers(v.Col, v.Row, g.corner(v.Col, v.Row, i)) {
		if s.Corner == g.canonical[0] || s.Corner == g.canonical[1] {
			return s, nil
		}
	}
	// unreachable: TestCanonicalVertexIsUniqueAndAgrees checks every corner.
	panic(fmt.Sprintf("wxx: %s corner %s has no canonical name", g.orientation, v))
}

// SameCorner reports whether a and b name the same corner of m's grid. The
// errors are TileCorner's.
func (m *Map_t) SameCorner(a, b Vertex_t) (bool, error) {
	pa, err := m.VertexPosition(a)
	if err != nil {
		return false, err
	}
	pb, err := m.VertexPosition(b)
	if err != nil {
		return false, err
	}
	return pa == pb, nil
}

// SharesEdge reports whether a and b are the two ends of one hex edge: some hex
// has both as consecutive corners. The same corner twice is not an edge. The
// errors are TileCorner's.
func (m *Map_t) SharesEdge(a, b Vertex_t) (bool, error) {
	g, err := m.hexGeometry()
	if err != nil {
		return false, err
	}
	pa, err := m.VertexPosition(a)
	if err != nil {
		return false, err
	}
	pb, err := m.VertexPosition(b)
	if err != nil {
		return false, err
	}
	for _, s := range g.sharers(a.Col, a.Row, pa) {
		i, _ := g.cornerIndex(s.Corner)
		if g.corner(s.Col, s.Row, (i+1)%6) == pb || g.corner(s.Col, s.Row, (i+5)%6) == pb {
			return true, nil
		}
	}
	return false, nil
}
