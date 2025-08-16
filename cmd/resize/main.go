// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package resize implements a command to resize a Worldographer map
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	var showBuildInfo, showVersion bool

	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&showBuildInfo, "build-info", false, "show version with build info")
	flag.StringVar(&inputFile, "input", "", "name of Worldographer file to load and resize")
	flag.StringVar(&outputFile, "output", "", "name to write the resized file to")
	flag.StringVar(&debugUtf8File, "debug-utf8", "", "optional name to write debug data to")
	flag.IntVar(&numberOfRowsToAddToTop, "top", 0, "optional number of rows to add to top")
	flag.IntVar(&numberOfRowsToAddToBottom, "bottom", 0, "optional number of rows to add to bottom")
	flag.IntVar(&numberOfColumnsToAddToLeft, "left", 0, "optional number of columns to add to left")
	flag.IntVar(&numberOfColumnsToAddToRight, "right", 0, "optional number of columns to add to right")
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
	if numberOfRowsToAddToTop < 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: top rows can't be less than zero\n")
		foundErrors = true
	}
	if numberOfRowsToAddToBottom < 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: bottom rows can't be less than zero\n")
		foundErrors = true
	}
	if numberOfColumnsToAddToLeft < 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: left columns can't be less than zero\n")
		foundErrors = true
	} else if numberOfColumnsToAddToLeft%2 != 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: left columns must be even\n")
		foundErrors = true
	}
	if numberOfColumnsToAddToRight < 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: right columns can't be less than zero\n")
		foundErrors = true
	}
	if foundErrors {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "  -input      file   load   .wxx file                   (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -output     file   create .wxx file                   (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -debug-utf8 file   xcreate debug UTF-8 XML file       (optional)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -top        int    number of rows    to add to top    (optional)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -bottom     int    number of rows    to add to bottom (optional)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -left       int    number of columns to add to left   (optional)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -right      int    number of columns to add to right  (optional)\n")
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
	log.Printf("add %4d rows    to top\n", numberOfRowsToAddToTop)
	log.Printf("add %4d columns to left\n", numberOfColumnsToAddToLeft)
	log.Printf("add %4d rows    to bottom\n", numberOfRowsToAddToBottom)
	log.Printf("add %4d columns to right\n", numberOfColumnsToAddToRight)

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
		os.Exit(1)
	}

	// warning: map tiles are indexed [column][row], not [row][column]
	//
	// orientation "COLUMNS" (Hexes Wide: 5, Hexes High: 3) (Circle: 001,001)
	//   tiles       5 wide x       3 high
	//   rows        5      x       3
	// orientation "ROWS"    (Hexes Wide: 5, Hexes High: 3) (Circle: 001,001)
	//   tiles       5 wide x       3 high
	//   rows        5      x       3
	log.Printf("orientation %q\n", inputMap.HexOrientation)
	log.Printf("map   %6d wide x %6d high\n", inputMap.Tiles.TilesWide, inputMap.Tiles.TilesHigh)
	log.Printf("tiles %6d      x %6d\n", len(inputMap.Tiles.Tiles), len(inputMap.Tiles.Tiles[0]))

	for _, feature := range inputMap.Features {
		if feature.Location != nil {
			log.Printf("feature x %g: y %g\n", feature.Location.X, feature.Location.Y)
		}
	}

	// calculate the size of the resized map
	height := inputMap.Tiles.TilesHigh + numberOfRowsToAddToTop + numberOfRowsToAddToBottom
	width := inputMap.Tiles.TilesWide + numberOfColumnsToAddToLeft + numberOfColumnsToAddToRight
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
				Terrain: 0, // by convention, slot 0 is the "blank" terrain tile
			}
			// todo: add logic to create coordinate based on the map orientation
			var cube hexg.CubeCoord
			if inputMap.GridOrientation == hexg.EvenQ {
				cube = hexg.NewOddQCoord(col, row).ToCube()
			} else {
				cube = hexg.NewOddQCoord(col, row).ToCube()
			}
			tile.Coords = cube
			outputTiles.Tiles[col][row] = tile
		}
	}
	log.Printf("sized %6d      x %6d\n", len(outputTiles.Tiles), len(outputTiles.Tiles[0]))
	// copy the input tiles to the resized map
	for col := 0; col < inputMap.Tiles.TilesWide; col++ {
		inputColumn := inputMap.Tiles.Tiles[col]
		outputColumn := outputTiles.Tiles[col+numberOfColumnsToAddToLeft]
		for row := 0; row < inputMap.Tiles.TilesHigh; row++ {
			inputTile := inputColumn[row]
			outputTile := outputColumn[row+numberOfRowsToAddToTop]
			// copy the remaining attributes
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

	// translate feature and label coordinates. the constants (225 and 300)
	// are used because Worldographer uses an "ideal" hex size for coordinates
	// in the file; it translates them to the actual hex size when rendering.
	// we must use the ideal when translating the coordinates.
	translatedX := float64(numberOfColumnsToAddToLeft) * 225.0
	translatedY := float64(numberOfRowsToAddToTop) * 300.0

	// update feature locations
	for _, feature := range inputMap.Features {
		if feature.Location != nil {
			feature.Location.X += translatedX
			feature.Location.Y += translatedY
		}
		// update feature label location if it exists
		if feature.Label != nil && feature.Label.Location != nil {
			feature.Label.Location.X += translatedX
			feature.Label.Location.Y += translatedY
		}
	}

	// update standalone label locations
	for _, label := range inputMap.Labels {
		if label.Location != nil {
			label.Location.X += translatedX
			label.Location.Y += translatedY
		}
	}

	// Write to the output file
	var encoderDiagnostics xmlio.EncoderDiagnostics
	encoder := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&encoderDiagnostics))
	outputBuffer := &bytes.Buffer{}
	err = encoder.Encode(outputBuffer, inputMap.MetaData.DataVersion, inputMap)
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

type Point struct {
	X, Y float64
}

// coordsToPoints returns the center point and vertices of a hexagon centered at the
// given column and row. It converts the column, row to the pixel at the center of the
// corresponding tile, then calculates the vertices based on that point.
// The center point is the first value in the returned slice.
func coordsToPoints(column, row int) [7]Point {
	const height, width = 300, 300
	const halfHeight, oneQuarterWidth, threeQuarterWidth = height / 2, width / 4, width * 3 / 4
	const leftMargin, topMargin = width / 2, halfHeight

	// points is the center plus the six vertices
	var points [7]Point

	points[0].X = float64(column)*threeQuarterWidth + leftMargin
	if column&1 == 1 { // shove odd rows down half the height of a tile
		points[0].Y = float64(row)*height + halfHeight + topMargin
	} else {
		points[0].Y = float64(row)*height + topMargin
	}

	// Calculate vertices based on offsets from center
	for i, offset := range flattenedHexOffsets {
		points[i+1] = Point{
			X: points[0].X + offset.X,
			Y: points[0].Y + offset.Y,
		}
	}

	return points
}

var (
	// Define the offsets based on the flattened hexagon dimensions
	flattenedHexOffsets = [6]Point{
		{-150, 0},   // left vertex
		{-75, -150}, // top-left vertex
		{75, -150},  // top-right vertex
		{150, 0},    // right vertex
		{75, 150},   // bottom-right vertex
		{-75, 150},  // bottom-left vertex
	}
)
