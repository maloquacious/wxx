// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package resize implements a command to resize a Worldographer map
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
	"github.com/maloquacious/wxx/xmlio"
)

func main() {
	var err error
	var inputFile, outputFile, debugUtf8File string
	var numberOfColumnsToAddToLeft int
	var numberOfColumnsToAddToRight int
	var numberOfRowsToAddToTop int
	var numberOfRowsToAddToBottom int
	var showBuildInfo, showSizing, showVersion bool
	var zoomLevel int

	flag.BoolVar(&showBuildInfo, "build-info", false, "show version with build info")
	flag.BoolVar(&showSizing, "debug-sizing", false, "show sizing and orientation")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.StringVar(&inputFile, "input", "", "name of Worldographer file to load and resize")
	flag.StringVar(&outputFile, "output", "", "name to write the resized file to")
	flag.StringVar(&debugUtf8File, "debug-utf8", "", "optional name to write debug data to")
	flag.IntVar(&numberOfRowsToAddToTop, "top", 0, "number of rows to add to top (negative to crop)")
	flag.IntVar(&numberOfRowsToAddToBottom, "bottom", 0, "number of rows to add to bottom (negative to crop)")
	flag.IntVar(&numberOfColumnsToAddToLeft, "left", 0, "number of columns to add to left (negative to crop)")
	flag.IntVar(&numberOfColumnsToAddToRight, "right", 0, "number of columns to add to right (negative to crop)")
	flag.IntVar(&zoomLevel, "zoom", 1, "zoom level in output file (default 1)")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s\n", wxx.Version().Short())
		return
	} else if showBuildInfo {
		fmt.Printf("%s\n", wxx.Version().String())
		return
	}

	foundErrors := false
	if inputFile == "" && outputFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing input and output file names\n")
		foundErrors = true
	} else if inputFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing input file name\n")
		foundErrors = true
	} else if outputFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing output file name\n")
		foundErrors = true
	}
	if zoomLevel < 1 || zoomLevel > 8 {
		_, _ = fmt.Fprintf(os.Stderr, "error: zoom level must be between 1 and 8")
	}
	if foundErrors {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "  -input      file   load   .wxx file                   (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -output     file   create .wxx file                   (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -debug-utf8 file   xcreate debug UTF-8 XML file       (optional)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -top        int    number of rows    to add to top    (negative to crop)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -bottom     int    number of rows    to add to bottom (negative to crop)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -left       int    number of columns to add to left   (negative to crop)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -right      int    number of columns to add to right  (negative to crop)\n")
		os.Exit(2)
	}

	// convert file names to absolute paths
	if inputFile, err = filepath.Abs(inputFile); err != nil {
		log.Fatalf("error: %v\n", err)
	} else if outputFile, err = filepath.Abs(outputFile); err != nil {
		log.Fatalf("error: %v\n", err)
	}
	if debugUtf8File != "" {
		if debugUtf8File, err = filepath.Abs(debugUtf8File); err != nil {
			log.Fatalf("error: %v\n", err)
		}
	}

	// it is an error if input and output have the same name.
	if inputFile == outputFile {
		log.Fatalf("error: cowardly refusing to overwrite input file")
	}
	log.Printf("input %q\n", inputFile)
	log.Printf("output %q\n", outputFile)
	if debugUtf8File != "" {
		log.Printf("debugUtf8 %q\n", debugUtf8File)
	}
	if showSizing {
		log.Printf("add %4d rows    to top\n", numberOfRowsToAddToTop)
		log.Printf("add %4d columns to left\n", numberOfColumnsToAddToLeft)
		log.Printf("add %4d rows    to bottom\n", numberOfRowsToAddToBottom)
		log.Printf("add %4d columns to right\n", numberOfColumnsToAddToRight)
	}

	// load the input file
	fp, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("error: opening file: %v\n", err)
	}
	defer func() {
		_ = fp.Close()
	}()

	inputMap, err := xmlio.NewDecoder().Decode(fp)
	if err != nil {
		log.Fatalf("error: loading Worldographer file: %v\n", err)
	}
	if showSizing {
		log.Printf("input  %6d      x %6d\n", len(inputMap.Tiles.Tiles), len(inputMap.Tiles.Tiles[0]))
	}

	// everything that depends on how the hexes are laid out comes from the
	// orientation the map states (issue #80).
	geometry, err := geometryFor(inputMap.HexOrientation)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	// the hex stagger is per column in COLUMNS and per row in ROWS, so adding
	// or removing an odd number at the start of that axis would shift every hex
	// half a step.
	if geometry.staggeredRows && numberOfRowsToAddToTop%2 != 0 {
		log.Fatalf("error: %s map: top rows must be even\n", inputMap.HexOrientation)
	} else if !geometry.staggeredRows && numberOfColumnsToAddToLeft%2 != 0 {
		log.Fatalf("error: %s map: left columns must be even\n", inputMap.HexOrientation)
	}

	// the hexes resize adds are Blank. A fully painted map does not list Blank,
	// because Worldographer lists only terrains in use, so it is added (#81).
	blankTerrainSlot := inputMap.TerrainMap.EnsureTerrain(wxx.BlankTerrain)

	inputMap.HexWidth, inputMap.HexHeight = geometry.hexWidth*float64(zoomLevel), geometry.hexHeight*float64(zoomLevel)

	// warning: map tiles are indexed [column][row], not [row][column]
	//
	// orientation "COLUMNS" (Hexes Wide: 5, Hexes High: 3) (Circle: 001,001)
	//   tiles       5 wide x       3 high
	//   rows        5      x       3
	// orientation "ROWS"    (Hexes Wide: 5, Hexes High: 3) (Circle: 001,001)
	//   tiles       5 wide x       3 high
	//   rows        5      x       3
	if showSizing {
		log.Printf("orientation %q\n", inputMap.HexOrientation)
		log.Printf("map    %6d wide x %6d high\n", inputMap.Tiles.TilesWide, inputMap.Tiles.TilesHigh)
		log.Printf("tiles  %6d      x %6d\n", len(inputMap.Tiles.Tiles), len(inputMap.Tiles.Tiles[0]))
	}

	// calculate the size of the resized map
	height := inputMap.Tiles.TilesHigh + numberOfRowsToAddToTop + numberOfRowsToAddToBottom
	width := inputMap.Tiles.TilesWide + numberOfColumnsToAddToLeft + numberOfColumnsToAddToRight

	// we can't make a tiny map. This must run before the allocation below: a
	// crop larger than the map makes width negative, and make() panics on it
	// (issue #61).
	if height < 2 || width < 2 {
		log.Fatalf("error: we can't create a map smaller than 2 x 2 (this resize gives %d x %d)\n", width, height)
	}

	// allocate a new Tiles_t to hold the resized map
	outputTiles := &wxx.Tiles_t{
		ViewLevel: inputMap.Tiles.ViewLevel,
		TilesHigh: height,
		TilesWide: width,
		Tiles:     make([][]*wxx.Tile_t, width),
	}

	// fill it with blank tiles that have the new coordinates
	for col := 0; col < width; col++ {
		outputTiles.Tiles[col] = make([]*wxx.Tile_t, height)
		for row := 0; row < height; row++ {
			tile := &wxx.Tile_t{
				Terrain: blankTerrainSlot,
				Coords:  geometry.coords(col, row),
			}
			outputTiles.Tiles[col][row] = tile
		}
	}
	if showSizing {
		log.Printf("output %6d      x %6d\n", len(outputTiles.Tiles), len(outputTiles.Tiles[0]))
	}

	// determine the source region to copy from the input map
	startCol := 0
	startRow := 0
	if numberOfColumnsToAddToLeft < 0 {
		startCol = -numberOfColumnsToAddToLeft // crop from left
	}
	if numberOfRowsToAddToTop < 0 {
		startRow = -numberOfRowsToAddToTop // crop from top
	}

	// determine where to stop copying (for right/bottom cropping)
	endCol := inputMap.Tiles.TilesWide
	endRow := inputMap.Tiles.TilesHigh
	if numberOfColumnsToAddToRight < 0 {
		endCol = inputMap.Tiles.TilesWide + numberOfColumnsToAddToRight // crop from right
	}
	if numberOfRowsToAddToBottom < 0 {
		endRow = inputMap.Tiles.TilesHigh + numberOfRowsToAddToBottom // crop from bottom
	}

	// copy tiles from input to output
	for col := startCol; col < endCol; col++ {
		inputColumn := inputMap.Tiles.Tiles[col]
		// calculate output column position: subtract start offset, add any left padding
		outputCol := col - startCol
		if numberOfColumnsToAddToLeft > 0 {
			outputCol += numberOfColumnsToAddToLeft
		}
		outputColumn := outputTiles.Tiles[outputCol]

		for row := startRow; row < endRow; row++ {
			inputTile := inputColumn[row]
			// calculate output row position: subtract start offset, add any top padding
			outputRow := row - startRow
			if numberOfRowsToAddToTop > 0 {
				outputRow += numberOfRowsToAddToTop
			}
			outputTile := outputColumn[outputRow]

			// copy tile attributes
			outputTile.Terrain = inputTile.Terrain
			outputTile.Elevation = inputTile.Elevation
			outputTile.IsIcy = inputTile.IsIcy
			outputTile.IsGMOnly = inputTile.IsGMOnly
			outputTile.Resources = inputTile.Resources
		}
	}

	// update the input map to use the new Tiles_t
	inputMap.RowsHigh = outputTiles.TilesHigh
	inputMap.ColumnsWide = outputTiles.TilesWide
	inputMap.Tiles = outputTiles

	// move everything that carries a map position by the same amount the tiles
	// moved, and drop what falls off the map (issue #79).
	shiftMapContent(inputMap, geometry, numberOfColumnsToAddToLeft, numberOfRowsToAddToTop)

	// Write to the output file, as the application version the INPUT states.
	//
	// Resizing changes the map's extent and nothing about its format, so this tool
	// names the input's own version as the target -- reading the provenance the
	// decoder recorded and choosing to write that. A CLIENT may do that; the
	// encoder may not do it for us, and has no default target (issue #45).
	var encoderDiagnostics xmlio.EncoderDiagnostics
	encoder := xmlio.NewEncoder(inputMap.MetaData.Version.App.Raw, xmlio.WithEncoderDiagnostics(&encoderDiagnostics))
	outputBuffer := &bytes.Buffer{}
	err = encoder.Encode(outputBuffer, inputMap)
	if err != nil {
		log.Fatalf("error: encoding %s: %v\n", outputFile, err)
	}
	if debugUtf8File != "" {
		err = os.WriteFile(debugUtf8File, encoderDiagnostics.Utf8Encoded, 0644)
		if err != nil {
			log.Fatalf("error: writing %s: %v\n", debugUtf8File, err)
		}
	}
	err = os.WriteFile(outputFile, outputBuffer.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error: writing %s: %v\n", outputFile, err)
	}

	log.Printf("%s: resized to %s\n", inputFile, outputFile)
}

// geometry_t is how a map's drawing coordinates relate to its hexes.
//
// Worldographer stores positions in an "ideal" coordinate space and scales them
// to the actual hex size when rendering, so moving content by whole hexes means
// moving it by these ideal amounts.
type geometry_t struct {
	colStep       float64 // x distance between adjacent columns
	rowStep       float64 // y distance between adjacent rows
	stagger       float64 // offset of the staggered hexes: odd columns down, or odd rows right
	staggeredRows bool    // ROWS: odd rows are staggered; COLUMNS: odd columns
	extraX        float64 // how far the map reaches past colStep*width
	extraY        float64 // how far the map reaches past rowStep*height

	// the hex size a zoom-1 map states, which resize writes
	hexWidth, hexHeight float64
}

// columnsGeometry is COLUMNS (flat-top) hexes: 300 x 300 ideal, columns 225
// apart, rows 300 apart, odd columns 150 lower. Hex (col,row) has its center
// at (150 + 225*col, 150 + 300*row + 150*odd(col)), and an <extraTerrain>
// placement is stated at the hex's corner, (225*col, 300*row + 150*odd(col)).
// Both fit every sample. Worldographer states the hex as 46.18 wide, 40 high.
var columnsGeometry = geometry_t{
	colStep: 225, rowStep: 300, stagger: 150, extraX: 75, extraY: 150,
	hexWidth: 46.18, hexHeight: 40,
}

// rowsGeometry is ROWS (pointy-top) hexes, the same hex turned a quarter:
// columns 300 apart, rows 225 apart, odd rows 150 to the right. Hex (col,row)
// has its center at (150 + 300*col + 150*odd(row), 150 + 225*row): confirmed in
// Worldographer on the maintainer's 2.08 ROWS sample, where a feature at
// (1050,1050) sits on hex (3,4), and a hand-shifted -left 2 -top 2 resize put
// it on (5,6) and a polygon from (5,4) on (7,6) (#80). The hex is stated 40
// wide, 46.18 high. Inferred and not yet seen in a sample: that a ROWS
// <extraTerrain> placement is stated at the hex's corner, (300*col +
// 150*odd(row), 225*row), mirroring COLUMNS.
var rowsGeometry = geometry_t{
	colStep: 300, rowStep: 225, stagger: 150, staggeredRows: true, extraX: 150, extraY: 75,
	hexWidth: 40, hexHeight: 46.18,
}

// geometryFor returns the geometry of a map's stated orientation.
func geometryFor(hexOrientation string) (geometry_t, error) {
	switch hexOrientation {
	case "COLUMNS":
		return columnsGeometry, nil
	case "ROWS":
		return rowsGeometry, nil
	}
	return geometry_t{}, fmt.Errorf("map/@hexOrientation %q: want COLUMNS or ROWS", hexOrientation)
}

// extent is the map's size in drawing coordinates.
func (g geometry_t) extent(width, height int) (float64, float64) {
	return g.colStep*float64(width) + g.extraX, g.rowStep*float64(height) + g.extraY
}

// placementHex is the hex an <extraTerrain> placement's location names.
func (g geometry_t) placementHex(x, y float64) (col, row int) {
	if g.staggeredRows {
		row = int(math.Round(y / g.rowStep))
		if row%2 != 0 {
			x -= g.stagger
		}
		return int(math.Round(x / g.colStep)), row
	}
	col = int(math.Round(x / g.colStep))
	if col%2 != 0 {
		y -= g.stagger
	}
	return col, int(math.Round(y / g.rowStep))
}

// coords is a hex's cube coordinates in the hexg convention the decoders use
// for this orientation: odd-q for COLUMNS, odd-r for ROWS.
func (g geometry_t) coords(col, row int) hexg.CubeCoord {
	if g.staggeredRows {
		return hexg.NewOddRCoord(col, row).ToCube()
	}
	return hexg.NewOddQCoord(col, row).ToCube()
}

// shiftMapContent moves every positioned element of m by dCols columns and
// dRows rows, and drops what the resized map no longer covers. m.Tiles must
// already hold the resized tiles.
//
// Features, labels, notes and shapes are points in drawing coordinates. Each
// is kept if it lies inside the map; a shape is kept if any of its points
// does. <extraTerrain> placements name a hex, so each is kept if its hex is on
// the map: its location is the hex's corner, and the first column's corner is
// x = 0, which a point test would reject.
func shiftMapContent(m *wxx.Map_t, g geometry_t, dCols, dRows int) {
	dx, dy := float64(dCols)*g.colStep, float64(dRows)*g.rowStep
	maxX, maxY := g.extent(m.Tiles.TilesWide, m.Tiles.TilesHigh)
	onMap := func(x, y float64) bool {
		return 0 < x && x < maxX && 0 < y && y < maxY
	}

	var features []*wxx.Feature_t
	for _, feature := range m.Features {
		keep := true
		if feature.Location != nil {
			feature.Location.X += dx
			feature.Location.Y += dy
			keep = onMap(feature.Location.X, feature.Location.Y)
		}
		if feature.Label != nil && feature.Label.Location != nil {
			feature.Label.Location.X += dx
			feature.Label.Location.Y += dy
			keep = keep && onMap(feature.Label.Location.X, feature.Label.Location.Y)
		}
		if keep {
			features = append(features, feature)
		}
	}
	m.Features = features

	var labels []*wxx.Label_t
	for _, label := range m.Labels {
		if label.Location != nil {
			label.Location.X += dx
			label.Location.Y += dy
			if !onMap(label.Location.X, label.Location.Y) {
				continue
			}
		}
		labels = append(labels, label)
	}
	m.Labels = labels

	var notes []*wxx.Note_t
	for _, note := range m.Notes {
		// the key is the note's position when it has one: 2.08 states the
		// position only there, so its notes decode with X and Y zero.
		x, y := note.X, note.Y
		if level, kx, ky, ok := parseNoteKey(note.Key); ok {
			x, y = kx+dx, ky+dy
			note.Key = level + "," + keyFloat(x) + "," + keyFloat(y)
		} else {
			x, y = x+dx, y+dy
		}
		note.X += dx
		note.Y += dy
		if onMap(x, y) {
			notes = append(notes, note)
		}
	}
	m.Notes = notes

	var shapes []*wxx.Shape_t
	for _, shape := range m.Shapes {
		keep := len(shape.Points) == 0
		for _, p := range shape.Points {
			p.X += dx
			p.Y += dy
			keep = keep || onMap(p.X, p.Y)
		}
		if keep {
			shapes = append(shapes, shape)
		}
	}
	m.Shapes = shapes

	if m.ExtraTerrain != nil {
		for _, layer := range m.ExtraTerrain.MapLayers {
			var placements []*wxx.TerrainAndLocation_t
			for _, tl := range layer.Terrain {
				tl.X += dx
				tl.Y += dy
				col, row := g.placementHex(tl.X, tl.Y)
				if 0 <= col && col < m.Tiles.TilesWide && 0 <= row && row < m.Tiles.TilesHigh {
					placements = append(placements, tl)
				}
			}
			layer.Terrain = placements
		}
	}
}

// parseNoteKey reads a note key in Worldographer's "<viewLevel>,<x>,<y>" form.
//
// 2.06 writes the position twice, in the key and in @x/@y. 2.08 writes it only
// in the key (its <note> has no @x or @y), so the key is the position that
// counts. A key not in this form reports ok = false and is left alone.
func parseNoteKey(key string) (level string, x, y float64, ok bool) {
	parts := strings.Split(key, ",")
	if len(parts) != 3 {
		return "", 0, 0, false
	}
	x, errX := strconv.ParseFloat(parts[1], 64)
	y, errY := strconv.ParseFloat(parts[2], 64)
	if errX != nil || errY != nil {
		return "", 0, 0, false
	}
	return parts[0], x, y, true
}

// keyFloat spells a coordinate the way note keys do: "100.0", "2343.75".
func keyFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
