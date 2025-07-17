// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements an application to import and export map data.
//
// This application shows all the steps to read and write the map data.
// Almost all the steps can (and should) be replaced by the reader and
// writer, but it is helpful to see what the results of each step look
// like.
//
// The following files will be created in the output folder:
//
//	input-utf-16.xml    -- the uncompressed data in the input .wxx file
//	input-utf-8.xml     -- the data converted to UTF-8 encoded XML
//	input.json          -- the data converted to JSON format
//	output.json         -- the data converted to JSON format
//	output-utf-8.xml    -- the data converted to UTF-8 encoded XML
//	output-utf-16.xml   -- the data converted to UTF-16 encoded XML
//	output.wxx          -- the data compressed and saved with .wxx extension
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/adapters"
	"github.com/maloquacious/wxx/readers"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	mv = semver.Version{
		Major: 0,
		Minor: 3,
		Patch: 0,
	}
)

func main() {
	log.SetFlags(log.LUTC | log.Ltime)

	// when the showVersion flag is true, the program will write the
	// application version (from the global 'mv' variable) and the
	// wxx package version and then quit.
	var showVersion bool
	flag.BoolVar(&showVersion, "version", showVersion, "show version and quit")

	// importFile is the name of the .wxx file to read from.
	var importFile string
	flag.StringVar(&importFile, "import", importFile, "file to load")

	// artifacts from each step will be written to the outputPath folder.
	var outputPath string
	flag.StringVar(&outputPath, "output-path", outputPath, "path to create debug files in")

	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(2)
	}

	if showVersion {
		fmt.Printf("wxxdemo version %s: wxx package version %s\n", mv.String(), wxx.Version())
		os.Exit(0)
	}

	// verify the parameters.
	if importFile == "" {
		log.Fatal("error: you must supply a file to import from\n")
	} else if outputPath == "" {
		log.Fatal("error: you must supply a path to write the output to\n")
	}

	if s, err := filepath.Abs(outputPath); err != nil {
		log.Fatalf("output: %v\n", err)
	} else {
		outputPath = s
	}

	log.Printf("importFile      == %s\n", importFile)
	log.Printf("outputPath      == %s\n", outputPath)
	if err := runDemo(importFile, outputPath); err != nil {
		log.Fatal(err)
	}
}

// runDemo is for development and testing
func runDemo(inputFile, outputPath string) error {
	started, step := time.Now(), time.Now()

	// input must exist and be a regular file
	step = time.Now()
	if sb, err := os.Stat(inputFile); err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	} else if sb.IsDir() {
		return fmt.Errorf("%s: is directory: %w", inputFile, os.ErrInvalid)
	} else if !sb.Mode().IsRegular() {
		return fmt.Errorf("%s: is not a file: %w", inputFile, os.ErrInvalid)
	}

	// output path is required and must be a folder
	if outputPath == "" {
		return fmt.Errorf("missing output path: %w", os.ErrInvalid)
	} else if sb, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("%s: %w", outputPath, err)
	} else if !sb.IsDir() {
		return fmt.Errorf("%s: is not a directory: %w", outputPath, os.ErrInvalid)
	}
	log.Printf("demo: completed setup checks      in %v\n", time.Now().Sub(step))

	// open a reader for the input
	step = time.Now()
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	defer func() {
		// readers must be closed to avoid leaking resources
		if input != nil {
			_ = input.Close()
			log.Printf("demo: closed  reader for %s\n", inputFile)
		}
	}()
	log.Printf("demo: created reader for %s\n", inputFile)
	log.Printf("demo: completed input reader      in %v\n", time.Now().Sub(step))

	// unzip the input
	step = time.Now()
	src, err := adapters.GZipToUTF16(input)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("demo: completed unzip             in %v\n", time.Now().Sub(step))

	// we are done with it, so cleanup the reader and free resources
	_ = input.Close()
	input = nil

	// write the uncompressed input to the output folder
	step = time.Now()
	filename := filepath.Join(outputPath, "input-utf-16.xml")
	if err = os.WriteFile(filename, src, 0644); err != nil {
		return err
	}
	log.Printf("demo: created %s\n", filename)
	log.Printf("demo: completed input-utf-16.xml  in %v\n", time.Now().Sub(step))

	// convert input from UTF-16 to UTF-8
	step = time.Now()
	src, err = readers.ReadUTF16(src)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("demo: completed utf-16 to utf-8   in %v\n", time.Now().Sub(step))

	// write the utf-8 data to the output folder
	step = time.Now()
	filename = filepath.Join(outputPath, "input-utf-8.xml")
	if err = os.WriteFile(filename, src, 0644); err != nil {
		return err
	}
	log.Printf("demo: created %s\n", filename)
	log.Printf("demo: completed input-utf-8.xml   in %v\n", time.Now().Sub(step))

	// read and convert the input from XML to Map data
	step = time.Now()
	inputMap, err := readers.ReadWXML(bytes.NewReader(src))
	if err != nil {
		log.Printf("src %q\n", src[:35])
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("demo: read map from %s %v\n", inputFile, inputMap.Version)
	log.Printf("demo: completed wxml conversion   in %v\n", time.Now().Sub(step))

	// convert the input Map data to JSON and write it to the output folder
	step = time.Now()
	filename = filepath.Join(outputPath, "input.json")
	if b, err := json.MarshalIndent(inputMap, "", "\t"); err != nil {
		return err
	} else if err = os.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	log.Printf("demo: created %s\n", filename)
	log.Printf("demo: completed input.json        in %v\n", time.Now().Sub(step))

	// pretend that we have manipulated the input data and created output data
	outputMap := inputMap

	// convert the output Map data to JSON data and write it to the output folder
	step = time.Now()
	filename = filepath.Join(outputPath, "output.json")
	if b, err := json.MarshalIndent(outputMap, "", "\t"); err != nil {
		return err
	} else if err = os.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	log.Printf("demo: completed output.json       in %v\n", time.Now().Sub(step))
	log.Printf("demo: created %s\n", filename)

	// convert the output Map data to TMap data
	step = time.Now()
	tmap, err := adapters.WMAPToTMAPv173(outputMap)
	if err != nil {
		return err
	}
	log.Printf("demo: completed wmap to tmap      in %v\n", time.Now().Sub(step))

	// convert the TMap data to UTF-8 encoded XML
	step = time.Now()
	data, err := tmap.Encode()
	if err != nil {
		return err
	}
	log.Printf("demo: completed tmap to xml       in %v %d\n", time.Now().Sub(step), len(data))

	// write the UTF-8 encode XML to the output folder
	filename = filepath.Join(outputPath, "output-utf-8.xml")
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	log.Printf("created %s\n", filename)
	log.Printf("demo: completed output-utf-8.xml  in %v\n", time.Now().Sub(started))

	// convert the UTF-8 encoded XML to UTF-16 encoded XML
	step = time.Now()
	data, err = adapters.UTF8ToUTF16(data)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("demo: completed utf-8 to utf-16   in %v\n", time.Now().Sub(step))

	// write the UTF-16 encoded XML to the output folder
	step = time.Now()
	filename = filepath.Join(outputPath, "output-utf-16.xml")
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	log.Printf("created %s\n", filename)
	log.Printf("demo: completed output-utf-16.xml in %v\n", time.Now().Sub(step))

	// compress the UTF-16 encoded XML data
	step = time.Now()
	data, err = adapters.UTF16ToGZip(data)
	if err != nil {
		return err
	}
	log.Printf("demo: completed compress xml      in %v\n", time.Now().Sub(step))

	// write the compressed data to the output folder using the .wxx extension
	step = time.Now()
	filename = filepath.Join(outputPath, "output.wxx")
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	log.Printf("created %s\n", filename)
	log.Printf("demo: completed output.wxx        in %v\n", time.Now().Sub(started))

	log.Printf("demo: completed                   in %v\n", time.Now().Sub(started))

	return nil
}
