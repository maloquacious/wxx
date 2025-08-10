// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package crop implements a command to crop a Worldographer file
// by reading the XML from the input and writing it out the new file.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

func main() {
	var showBuildInfo, showVersion, quiet bool
	var inputFile, outputFile, debugUtf8XmlFile string

	// TODO: Future sprint - add flag to let user specify the data version in the copied file
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&showBuildInfo, "build-info", false, "show version with build info")
	flag.StringVar(&inputFile, "input", "", "input .wxx file (required)")
	flag.StringVar(&outputFile, "output", "", "output .wxx file (required)")
	flag.StringVar(&debugUtf8XmlFile, "debug-utf8", "", "write debug UTF-8 XML file alongside compressed UTF-16 .wxx file")
	flag.BoolVar(&quiet, "quiet", false, "suppress output messages")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s\n", wxx.Version().Short())
		return
	} else if showBuildInfo {
		fmt.Printf("%s\n", wxx.Version().String())
		return
	}

	if inputFile == "" || outputFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "  -input file        input .wxx file (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -output file       output .wxx file (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -debug-utf8        write debug UTF-8 XML file alongside compressed UTF-16 .wxx file\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -quiet             suppress output messages\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -version           show version\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -build-info        show version with build info\n")
		os.Exit(2)
	}

	if !strings.HasSuffix(inputFile, ".wxx") {
		_, _ = fmt.Fprintf(os.Stderr, "error: input file must end with .wxx\n")
		os.Exit(2)
	}

	if !strings.HasSuffix(outputFile, ".wxx") {
		_, _ = fmt.Fprintf(os.Stderr, "error: output file must end with .wxx\n")
		os.Exit(2)
	}

	// Read the input file
	fp, err := os.Open(inputFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error opening %s: %v\n", inputFile, err)
		os.Exit(2)
	}
	defer fp.Close()
	var bif xmlio.Diagnostics
	joy := xmlio.NewDecoder(xmlio.WithDiagnostics(&bif))
	input, err := joy.Decode(fp)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading %s: %v\n", inputFile, err)
		os.Exit(2)
	}

	if !quiet {
		fmt.Printf("input: %s (data version %s)\n", inputFile, input.MetaData.DataVersion.String())
	}

	if input.Tiles.TilesWide < 8 || input.Tiles.TilesHigh < 8 {
		fmt.Printf("error: map too small to crop (want at least 8x8, got %dx%d)\n", input.Tiles.TilesWide, input.Tiles.TilesHigh)
	}

	fmt.Printf("input: %4d x %4d\n", input.Tiles.TilesWide, input.Tiles.TilesHigh)
	fmt.Printf("input: %q: %g x %g\n", input.HexOrientation, input.HexWidth, input.HexHeight)

	// crop the map by subtracting 2 from each edge
	lrow, rrow := 14, input.Tiles.TilesWide-14
	if lrow < 0 || rrow < 0 {
		fmt.Printf("error: map too small to crop (got %dx%d)\n", input.Tiles.TilesWide, input.Tiles.TilesHigh)
	}
	tcol, bcol := 14, input.Tiles.TilesHigh-14
	if tcol < 0 || bcol < 0 {
		fmt.Printf("error: map too small to crop (got %dx%d)\n", input.Tiles.TilesWide, input.Tiles.TilesHigh)
	}
	dstWidth, dstHeight := rrow-lrow+1, bcol-tcol+1
	fmt.Printf(" crop: %4d x %4d\n", dstWidth, dstHeight)
	// allocate cells for the destination map
	dst := make([][]*wxx.Tile_t, dstWidth)
	for i := range dst {
		dst[i] = make([]*wxx.Tile_t, dstHeight)
	}
	// copy tiles within the crop area
	for x := lrow; x <= rrow; x++ {
		for y := tcol; y <= bcol; y++ {
			dst[x-lrow][y-tcol] = input.Tiles.TileRows[x][y]
		}
	}
	// update the input
	input.Tiles.TilesWide, input.Tiles.TilesHigh = dstWidth, dstHeight
	input.Tiles.TileRows = dst
	fmt.Printf("input: %4d x %4d\n", input.Tiles.TilesWide, input.Tiles.TilesHigh)

	// Write to the output file
	err = xmlio.WriteFile(outputFile, input.MetaData.DataVersion, input, debugUtf8XmlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outputFile, err)
		os.Exit(1)
	}

	if !quiet {
		fmt.Printf("copied %s to %s\n", inputFile, outputFile)
	}
}
