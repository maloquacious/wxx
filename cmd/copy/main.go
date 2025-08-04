// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package copy implements a command to copy a Worldographer file
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
	var showBuildInfo, showVersion, writeDebugUtf8, quiet bool
	var inputFile, outputFile string

	// TODO: Future sprint - add flag to let user specify the data version in the copied file
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&showBuildInfo, "build-info", false, "show version with build info")
	flag.StringVar(&inputFile, "input", "", "input .wxx file (required)")
	flag.StringVar(&outputFile, "output", "", "output .wxx file (required)")
	flag.BoolVar(&writeDebugUtf8, "debug-utf8", false, "write debug UTF-8 XML file alongside compressed UTF-16 .wxx file")
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  -input file        input .wxx file (required)\n")
		fmt.Fprintf(os.Stderr, "  -output file       output .wxx file (required)\n")
		fmt.Fprintf(os.Stderr, "  -debug-utf8        write debug UTF-8 XML file alongside compressed UTF-16 .wxx file\n")
		fmt.Fprintf(os.Stderr, "  -quiet             suppress output messages\n")
		fmt.Fprintf(os.Stderr, "  -version           show version\n")
		fmt.Fprintf(os.Stderr, "  -build-info        show version with build info\n")
		os.Exit(2)
	}

	if !strings.HasSuffix(inputFile, ".wxx") {
		fmt.Fprintf(os.Stderr, "error: input file must end with .wxx\n")
		os.Exit(2)
	}

	if !strings.HasSuffix(outputFile, ".wxx") {
		fmt.Fprintf(os.Stderr, "error: output file must end with .wxx\n")
		os.Exit(2)
	}

	// Read the input file
	data, err := xmlio.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	if !quiet {
		fmt.Printf("input: %s (data version %s)\n", inputFile, data.MetaData.DataVersion.String())
	}

	// Write to the output file
	err = xmlio.WriteFile(data.MetaData.DataVersion, data, writeDebugUtf8)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outputFile, err)
		os.Exit(1)
	}

	if !quiet {
		fmt.Printf("copied %s to %s\n", inputFile, outputFile)
	}
}
