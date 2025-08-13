// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package copy implements a command to copy a Worldographer file
// by reading the XML from the input and writing it out the new file.
package main

import (
	"bytes"
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
	var decoderDiagnostics xmlio.DecoderDiagnostics
	joy := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&decoderDiagnostics))
	inputMap, err := joy.Decode(fp)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading %s: %v\n", inputFile, err)
		os.Exit(2)
	}

	if !quiet {
		fmt.Printf("input: %s (data version %s)\n", inputFile, inputMap.MetaData.DataVersion.String())
	}

	// Write to the output file
	var encoderDiagnostics xmlio.EncoderDiagnostics
	bah := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&encoderDiagnostics))
	outputBuffer := &bytes.Buffer{}
	err = bah.Encode(outputBuffer, inputMap.MetaData.DataVersion, inputMap)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error encoding %s: %v\n", outputFile, err)
		os.Exit(1)
	}
	err = os.WriteFile(outputFile, outputBuffer.Bytes(), 0644)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outputFile, err)
		os.Exit(1)
	}

	if !quiet {
		fmt.Printf("copied %s to %s\n", inputFile, outputFile)
	}
}
