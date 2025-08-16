// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a simple merge command.
// Needs much work to be useable.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

func main() {
	var err error
	var outputFile, debugUtf8File string
	var showBuildInfo, showVersion bool

	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&showBuildInfo, "build-info", false, "show version with build info")
	flag.StringVar(&outputFile, "output", "", "name to write the resized file to")
	flag.StringVar(&debugUtf8File, "debug-utf8", "", "optional name to write debug data to")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s\n", wxx.Version().Short())
		return
	} else if showBuildInfo {
		fmt.Printf("%s\n", wxx.Version().String())
		return
	}

	// all remaining arguments are assumed to be input file names to merge
	inputFiles := flag.Args()
	log.Printf("input files %+v\n", inputFiles)

	foundErrors := false
	if len(inputFiles) == 0 && outputFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing input and output file names\n")
		foundErrors = true
	} else if len(inputFiles) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing input file name\n")
		foundErrors = true
	} else if outputFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing output file name\n")
		foundErrors = true
	}
	if foundErrors {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s [options] input-file-names\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "  -output     file   create .wxx file                   (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -debug-utf8 file   xcreate debug UTF-8 XML file       (optional)\n")
		os.Exit(2)
	}

	// convert file names to absolute paths
	for n, inputFile := range inputFiles {
		inputFile, err = filepath.Abs(inputFile)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		inputFiles[n] = inputFile
	}
	if outputFile, err = filepath.Abs(outputFile); err != nil {
		log.Fatalf("error: %v\n", err)
	}
	if debugUtf8File != "" {
		if debugUtf8File, err = filepath.Abs(debugUtf8File); err != nil {
			log.Fatalf("error: %v\n", err)
		}
	}

	// it is an error if any input has the same name as the output.
	for _, inputFile := range inputFiles {
		log.Printf("input %q\n", inputFile)
		if inputFile == outputFile {
			log.Fatalf("error: cowardly refusing to overwrite input file")
		}
	}
	log.Printf("output %q\n", outputFile)
	if debugUtf8File != "" {
		log.Printf("debugUtf8 %q\n", debugUtf8File)
	}

	// load the input files
	var inputMaps []*wxx.Map_t
	for _, inputFile := range inputFiles {
		log.Printf("%s: loading\n", inputFile)
		fp, err := os.Open(inputFile)
		if err != nil {
			log.Fatalf("error: opening file: %v\n", err)
		}
		inputMap, err := xmlio.NewDecoder().Decode(fp)
		if err != nil {
			log.Fatalf("error: loading Worldographer file: %v\n", err)
		}
		inputMaps = append(inputMaps, inputMap)
		_ = fp.Close()
	}

	// all maps must have the same orientation and size and terrain map slots
	orientation := inputMaps[0].HexOrientation
	height, width := inputMaps[0].RowsHigh, inputMaps[0].ColumnsWide
	terrainMap := inputMaps[0].TerrainMap.Data
	terrainSlot := map[int]string{}
	for name, slot := range terrainMap {
		terrainSlot[slot] = name
	}
	foundErrors = false
	for n, inputMap := range inputMaps {
		if inputMap.HexOrientation != orientation {
			log.Printf("%s: orientation is %q (should be %q)\n", inputFiles[n], inputMap.HexOrientation, orientation)
			foundErrors = true
		}
		if inputMap.RowsHigh != height {
			log.Printf("%s: height is %d (should be %d)\n", inputFiles[n], inputMap.RowsHigh, height)
			foundErrors = true
		}
		if inputMap.ColumnsWide != width {
			log.Printf("%s: width is %d (should be %d)\n", inputFiles[n], inputMap.ColumnsWide, width)
			foundErrors = true
		}
		// todo: verify the terrains
		for k, _ := range inputMap.TerrainMap.Data {
			_, ok := terrainMap[k]
			if !ok {
				log.Printf("%s: terrain %q: missing\n", inputFiles[n], k)
				foundErrors = true
				//} else if v != slot {
				//	log.Printf("%s: terrain %q: slot %d (should be %d)\n", inputFiles[n], k, v, slot)
				//	foundErrors = true
			}
		}
	}
	if foundErrors {
		log.Fatalf("error: can't merge maps unless they have the same size and orientation")
	}

	// merge the files
	var outputMap *wxx.Map_t
	for n, inputMap := range inputMaps {
		log.Printf("merging %q\n", inputFiles[n])
		if outputMap == nil {
			outputMap = inputMap
			continue
		}
		// merge terrain
		for col := 0; col < inputMap.Tiles.TilesWide; col++ {
			inputColumn := inputMap.Tiles.Tiles[col]
			outputColumn := outputMap.Tiles.Tiles[col]
			for row := 0; row < inputMap.Tiles.TilesHigh; row++ {
				outputTile := outputColumn[row]
				if terrainSlot[outputTile.Terrain] != "Blank" {
					// don't overwrite non-blank tiles
					log.Printf("%d, %d: skipping terrain %d\n", outputTile.Column, outputTile.Row, outputTile.Terrain)
					continue
				}
				inputTile := inputColumn[row]
				inputTerrainLabel := inputMap.TerrainMap.List[inputTile.Terrain].Label
				slot, ok := terrainMap[inputTerrainLabel]
				if !ok {
					panic("assert(slot is ok)")
				}
				outputTile.Terrain = slot
				outputTile.Elevation = inputTile.Elevation
				outputTile.IsIcy = inputTile.IsIcy
				outputTile.IsGMOnly = inputTile.IsGMOnly
				outputTile.Resources = inputTile.Resources
				outputTile.CustomBackgroundColor = inputTile.CustomBackgroundColor
			}
		}
		// merge features
		for _, feature := range inputMap.Features {
			outputMap.Features = append(outputMap.Features, feature)
		}
		// merge labels
		for _, label := range inputMap.Labels {
			outputMap.Labels = append(outputMap.Labels, label)
		}
	}
	if outputMap == nil {
		panic("assert(outputMap != nil)")
	}

	// Write to the output file
	var encoderDiagnostics xmlio.EncoderDiagnostics
	encoder := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&encoderDiagnostics))
	outputBuffer := &bytes.Buffer{}
	err = encoder.Encode(outputBuffer, outputMap.MetaData.DataVersion, outputMap)
	if err != nil {
		log.Fatalf("error: encoding %s: %v\n", outputFile, err)
	}
	if debugUtf8File != "" {
		err = os.WriteFile(debugUtf8File, encoderDiagnostics.Utf8Encoded, 0644)
		if err != nil {
			log.Fatalf("error: writing %s: %v\n", debugUtf8File, err)
		}
		log.Printf("created %q\n", debugUtf8File)
	}
	err = os.WriteFile(outputFile, outputBuffer.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error: writing %s: %v\n", outputFile, err)
	}
	log.Printf("created %q\n", outputFile)
}
