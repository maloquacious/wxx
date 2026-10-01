// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// PathOption configures NewPath and NewEdgePath.
type PathOption func(*pathOptions)

type pathOptions struct {
	mapLayer    string
	strokeWidth float64
	strokeColor RGBA_t
	gmOnly      bool
}

// The defaults are the values Worldographer wrote for the lines in the
// river-lines fixture (issue #154): blue, 0.05 wide, on Above Terrain, the
// layer above the terrain layers.
const (
	DefaultPathMapLayer    = "Above Terrain"
	DefaultPathStrokeWidth = 0.05
)

// defaultPathStrokeColor is opaque blue. It is not exported because a Go
// variable could be changed by any importer.
var defaultPathStrokeColor = RGBA_t{R: 0, G: 0, B: 1, A: 1}

// WithMapLayer puts the path on the named layer, which must be one of the
// map's MapLayers. The default is DefaultPathMapLayer.
func WithMapLayer(name string) PathOption {
	return func(o *pathOptions) { o.mapLayer = name }
}

// WithStrokeWidth sets the line's width, which must be greater than zero. The
// default is DefaultPathStrokeWidth.
func WithStrokeWidth(width float64) PathOption {
	return func(o *pathOptions) { o.strokeWidth = width }
}

// WithStrokeColor sets the line's color; each component must be from 0 to 1.
// The default is opaque blue, RGBA_t{R: 0, G: 0, B: 1, A: 1}.
func WithStrokeColor(c RGBA_t) PathOption {
	return func(o *pathOptions) { o.strokeColor = c }
}

// WithGMOnly marks the path as shown to the GM only. The default is false.
func WithGMOnly(gmOnly bool) PathOption {
	return func(o *pathOptions) { o.gmOnly = gmOnly }
}

// NewPath returns a line through points, in shape coordinates, as a Path shape
// carrying what Worldographer writes for a line drawn with the Shapes tab's
// Line tool and no fill (issue #155). Append it to m.Shapes to add it to the
// map.
//
// There must be at least two points, each with finite coordinates. The first
// becomes the move-to (type "m") and the rest line-tos, as the app writes them.
// A point equal to the one before it is kept: the app writes one where a
// corner is clicked twice.
//
// NewPath is for lines that are not on the grid; for a line along hex edges,
// NewEdgePath checks the corners and snaps the line to the grid. A NewPath
// line is written with isSnapVertices="false".
//
// Every attribute the options do not set has the value Worldographer 2.06,
// 2.07 and 2.08 all write for such a line, so the shape does not depend on the
// application version it is written as. The view levels are WORLD.
func (m *Map_t) NewPath(points []Position_t, opts ...PathOption) (*Shape_t, error) {
	if len(points) < 2 {
		return nil, errors.Join(ErrInvalidShape, fmt.Errorf("new path: %d point(s), want at least 2", len(points)))
	}
	for i, p := range points {
		if math.IsNaN(p.X) || math.IsInf(p.X, 0) || math.IsNaN(p.Y) || math.IsInf(p.Y, 0) {
			return nil, errors.Join(ErrInvalidShape, fmt.Errorf("new path: point %d: (%v, %v) is not a finite position", i, p.X, p.Y))
		}
	}
	return m.newPath(points, false, opts)
}

// NewEdgePath returns a line along hex edges, through the hex corners
// vertices, as a Path shape (issue #155). It is NewPath for the grid: each
// vertex is converted with VertexPosition (see geometry.go), and the shape is
// written with isSnapVertices="true", as the app writes a line drawn with
// Snap Points to Grid. Append it to m.Shapes to add it to the map.
//
// Each pair of consecutive vertices must be the two ends of one hex edge
// (SharesEdge). Anything else, including the same corner twice, is an
// ErrInvalidShape error that names the pair: getting odd and even columns
// wrong is the easy mistake on this grid, and without the check it would show
// up only as a wrong line in Worldographer. One edge is a two-vertex path.
// Corners of hexes off the map are allowed, so a path can run along the map's
// edge. A corner that is not one of m's orientation is ErrInvalidCorner.
func (m *Map_t) NewEdgePath(vertices []Vertex_t, opts ...PathOption) (*Shape_t, error) {
	if len(vertices) < 2 {
		return nil, errors.Join(ErrInvalidShape, fmt.Errorf("new edge path: %d vertex(es), want at least 2", len(vertices)))
	}
	points := make([]Position_t, len(vertices))
	for i, v := range vertices {
		p, err := m.VertexPosition(v)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("new edge path: vertex %d", i))
		}
		points[i] = p
		if i == 0 {
			continue
		}
		ok, err := m.SharesEdge(vertices[i-1], v)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("new edge path: vertices %d and %d", i-1, i))
		}
		if !ok {
			return nil, errors.Join(ErrInvalidShape, fmt.Errorf(
				"new edge path: vertices %d and %d, %s and %s, are not the two ends of one hex edge", i-1, i, vertices[i-1], v))
		}
	}
	return m.newPath(points, true, opts)
}

// newPath builds the shape once the points are known to be good.
func (m *Map_t) newPath(points []Position_t, snap bool, opts []PathOption) (*Shape_t, error) {
	o := pathOptions{mapLayer: DefaultPathMapLayer, strokeWidth: DefaultPathStrokeWidth, strokeColor: defaultPathStrokeColor}
	for _, opt := range opts {
		opt(&o)
	}
	layer := false
	for _, ml := range m.MapLayers {
		layer = layer || (ml != nil && ml.Name == o.mapLayer)
	}
	if !layer {
		return nil, errors.Join(ErrUnknownMapLayer, fmt.Errorf("new path: map layer %q is not one of the map's layers", o.mapLayer))
	}
	if !(o.strokeWidth > 0) || math.IsInf(o.strokeWidth, 0) {
		return nil, errors.Join(ErrInvalidShape, fmt.Errorf("new path: stroke width %v: want a finite number greater than 0", o.strokeWidth))
	}
	if !o.strokeColor.inRange() {
		return nil, errors.Join(ErrInvalidColorAttribute, fmt.Errorf("new path: stroke color %+v: want each component from 0 to 1", o.strokeColor))
	}

	s := &Shape_t{
		Type:                  "Path",
		IsGMOnly:              o.gmOnly,
		IsSnapVertices:        snap,
		CreationType:          "BASIC",
		ExtraLineDistance:     30,
		ExtraLineLength:       30,
		ExtraLineWidth:        10,
		ExtraLineSeparation:   15,
		IsWorld:               true,
		IsContinent:           true,
		IsKingdom:             true,
		IsProvince:            true,
		DsSpread:              0.2,
		DsRadius:              50,
		InsChoke:              0.2,
		InsRadius:             50,
		BbWidth:               10,
		BbHeight:              10,
		BbIterations:          3,
		MapLayer:              o.mapLayer,
		StrokeType:            "SIMPLE",
		HighestViewLevel:      "WORLD",
		CurrentShapeViewLevel: "WORLD",
		LineCap:               "SQUARE",
		LineJoin:              "ROUND",
		Opacity:               1,
		FillRule:              "NON_ZERO",
		StrokeColor:           rgbaAttribute(o.strokeColor),
		StrokeWidth:           o.strokeWidth,
		// The app writes this color for the drop and inner shadows of every
		// shape in the fixtures, though neither is turned on.
		DsColor:  "1.0,0.8941176533699036,0.7686274647712708,1.0",
		InsColor: "1.0,0.8941176533699036,0.7686274647712708,1.0",
	}
	for i, p := range points {
		pt := &Point_t{X: p.X, Y: p.Y}
		if i == 0 {
			pt.Type = "m"
		}
		s.Points = append(s.Points, pt)
	}
	return s, nil
}

// rgbaAttribute spells a color as a shape's color attributes hold it,
// "r,g,b,a", with each component as Java's Double.toString writes it, as the
// W2025 codec's floats does: "0.0,0.0,1.0,1.0". The components are in 0..1.
func rgbaAttribute(c RGBA_t) string {
	parts := make([]string, 4)
	for i, v := range []float64{c.R, c.G, c.B, c.A} {
		parts[i] = javaDouble(v)
	}
	return strings.Join(parts, ",")
}

// javaDouble is Double.toString for a finite float64; see the W2025 codec's
// floats, which this repeats because the codec is internal to xmlio.
func javaDouble(f float64) string {
	if a := math.Abs(f); a == 0 || (1e-3 <= a && a < 1e7) {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	}
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(f, 'e', -1, 64), "e")
	if !strings.Contains(mantissa, ".") {
		mantissa, exponent, _ = strings.Cut(strconv.FormatFloat(f, 'e', 1, 64), "e")
	}
	e, _ := strconv.Atoi(exponent)
	return mantissa + "E" + strconv.Itoa(e)
}
