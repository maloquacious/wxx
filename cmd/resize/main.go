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
	if numberOfColumnsToAddToLeft%2 != 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: left columns must be even\n")
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

	blankTerrainSlot, ok := inputMap.TerrainMap.Data["Blank"]
	if !ok {
		log.Fatalf("error: file doesn't have a \"Blank\" terrain slot\n")
	}

	inputMap.HexWidth, inputMap.HexHeight = 46.18*float64(zoomLevel), 40.0*float64(zoomLevel)

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
	// allocate a new Tiles_t to hold the resized map
	outputTiles := &wxx.Tiles_t{
		ViewLevel: inputMap.Tiles.ViewLevel,
		TilesHigh: height,
		TilesWide: width,
		Tiles:     make([][]*wxx.Tile_t, width),
	}

	// we can't make a tiny map
	if height < 2 || width < 2 {
		log.Fatalf("error: we can't create a map smaller than 2 x 2\n")
	}

	// fill it with blank tiles that have the new coordinates
	for col := 0; col < width; col++ {
		outputTiles.Tiles[col] = make([]*wxx.Tile_t, height)
		for row := 0; row < height; row++ {
			tile := &wxx.Tile_t{
				Terrain: blankTerrainSlot,
				Coords:  hexg.NewOddQCoord(col, row).ToCube(),
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

	// translate feature and label coordinates. the constants (225 and 300)
	// are used because Worldographer uses an "ideal" hex size for coordinates
	// in the file; it translates them to the actual hex size when rendering.
	// we must use the ideal when translating the coordinates.
	translatedX := float64(numberOfColumnsToAddToLeft) * 225.0
	translatedY := float64(numberOfRowsToAddToTop) * 300.0

	// update feature locations, culling anything that is off the map
	var outputFeatures []*wxx.Feature_t
	for _, feature := range inputMap.Features {
		if feature.Location != nil {
			feature.Location.X += translatedX
			feature.Location.Y += translatedY
			if feature.Location.X <= 0 || feature.Location.Y <= 0 {
				continue // off the map, so ignore it
			}
		}
		// update feature label location if it exists
		if feature.Label != nil && feature.Label.Location != nil {
			feature.Label.Location.X += translatedX
			feature.Label.Location.Y += translatedY
			if feature.Label.Location.X <= 1.0 || feature.Label.Location.Y <= 1.0 {
				continue // off the map, so ignore it
			}
		}
		outputFeatures = append(outputFeatures, feature)
	}
	inputMap.Features = outputFeatures

	// update standalone label locations
	var outputLabels []*wxx.Label_t
	for _, label := range inputMap.Labels {
		if label.Location != nil {
			label.Location.X += translatedX
			label.Location.Y += translatedY
			if label.Location.X <= 1.0 || label.Location.Y <= 1.0 {
				continue // off the map, so ignore it
			}
		}
		outputLabels = append(outputLabels, label)
	}
	inputMap.Labels = outputLabels

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
