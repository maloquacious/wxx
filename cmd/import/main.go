// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a command to import worldographer terrain layers.
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
	//var err error
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

	// the remaining argument must be the map file to import into
	foundErrors := false
	if len(flag.Args()) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing input file name\n")
		foundErrors = true
	} else if len(flag.Args()) != 1 {
		for _, arg := range flag.Args() {
			_, _ = fmt.Fprintf(os.Stderr, "error: unknown option %q\n", arg)
		}
		foundErrors = true
	} else if outputFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "error: missing output file name\n")
		foundErrors = true
	}
	if foundErrors {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s [options] input-file-name\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "  -output     file   create .wxx file                   (required)\n")
		_, _ = fmt.Fprintf(os.Stderr, "  -debug-utf8 file   xcreate debug UTF-8 XML file       (optional)\n")
		os.Exit(2)
	}

	// convert file names to absolute paths
	inputFile, err := filepath.Abs(flag.Args()[0])
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	if outputFile, err = filepath.Abs(outputFile); err != nil {
		log.Fatalf("error: %v\n", err)
	}
	if debugUtf8File != "" {
		if debugUtf8File, err = filepath.Abs(debugUtf8File); err != nil {
			log.Fatalf("error: %v\n", err)
		}
	}

	log.Printf("import: input  %q\n", inputFile)
	log.Printf("import: output %q\n", outputFile)
	if debugUtf8File != "" {
		log.Printf("import: debug  %q\n", debugUtf8File)
	}

	// it is an error if any input has the same name as the output.
	if inputFile == outputFile {
		log.Fatalf("error: cowardly refusing to overwrite input file")
	}

	// load the input file
	log.Printf("%s: loading\n", inputFile)
	fp, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("error: opening file: %v\n", err)
	}
	inputMap, err := xmlio.NewDecoder().Decode(fp)
	if err != nil {
		log.Fatalf("error: loading Worldographer file: %v\n", err)
	}
	_ = fp.Close()

	// import from the JSON files

	// Write to the output file
	outputMap := inputMap
	if outputMap == nil {
		panic("assert(outputMap != nil)")
	}
	// The target is the application version the INPUT states: importing adds
	// content to a map without changing what version it is. This tool reads that
	// provenance and names it as the target, which a CLIENT may do; the encoder may
	// not do it for us, and has no default target (issue #45).
	var encoderDiagnostics xmlio.EncoderDiagnostics
	encoder := xmlio.NewEncoder(outputMap.MetaData.Version.App.Raw, xmlio.WithEncoderDiagnostics(&encoderDiagnostics))
	outputBuffer := &bytes.Buffer{}
	err = encoder.Encode(outputBuffer, outputMap)
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
