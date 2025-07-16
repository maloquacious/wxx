// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements an application to import and export map data.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx/adapters"
	"github.com/maloquacious/wxx/readers"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	mv = semver.Version{
		Major:      0,
		Minor:      1,
		Patch:      0,
		PreRelease: "alpha",
	}
)

func main() {
	log.SetFlags(log.LUTC | log.Ltime)

	var showVersion bool
	flag.BoolVar(&showVersion, "version", showVersion, "show version and quit")

	var importFile string
	flag.StringVar(&importFile, "import", importFile, "file to load")

	var exportFile string
	flag.StringVar(&exportFile, "export", exportFile, "file to create")

	var debugOutputPath string
	flag.StringVar(&debugOutputPath, "debug-output-path", debugOutputPath, "path to create debug files in")

	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(2)
	}

	if showVersion {
		fmt.Printf("%s\n", mv.String())
		os.Exit(0)
	}

	if importFile == "" {
		log.Fatal("error: you must supply a file to import from\n")
	} else if exportFile == "" && debugOutputPath == "" {
		log.Fatal("error: you must supply a file to export to\n")
	}

	if debugOutputPath != "" {
		if s, err := filepath.Abs(debugOutputPath); err != nil {
			log.Fatalf("output: %v\n", err)
		} else {
			debugOutputPath = s
		}
	}

	if debugOutputPath != "" {
		log.Printf("importFile      == %s\n", importFile)
		log.Printf("debugOutputPath == %s\n", debugOutputPath)
		if err := run(importFile, debugOutputPath); err != nil {
			log.Fatal(err)
		}
		log.Printf("debug output completed\n")
		os.Exit(0)
	}

	// input must exist and be a regular file
	if sb, err := os.Stat(importFile); err != nil {
		log.Fatalf("error: %v\n", err)
	} else if sb.IsDir() {
		log.Fatalf("error: %s: is a directory\n", importFile)
	} else if !sb.Mode().IsRegular() {
		log.Fatalf("error: %s: is not a file\n", importFile)
	}
}

// run is for development and testing
func run(inputFile, debugOutputPath string) error {
	started, step := time.Now(), time.Now()

	step = time.Now()
	if sb, err := os.Stat(inputFile); err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	} else if sb.IsDir() {
		return fmt.Errorf("%s: is directory: %w", inputFile, os.ErrInvalid)
	} else if !sb.Mode().IsRegular() {
		return fmt.Errorf("%s: is not a file: %w", inputFile, os.ErrInvalid)
	}
	if debugOutputPath == "" {
		return fmt.Errorf("missing debug output path: %w", os.ErrInvalid)
	}
	if sb, err := os.Stat(debugOutputPath); err != nil {
		return fmt.Errorf("%s: %w", debugOutputPath, err)
	} else if !sb.IsDir() {
		return fmt.Errorf("%s: is not a directory: %w", debugOutputPath, os.ErrInvalid)
	}
	log.Printf("debug: completed setup checks      in %v\n", time.Now().Sub(step))

	// open a reader for the input
	step = time.Now()
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	defer func() {
		if input != nil {
			_ = input.Close()
			log.Printf("debug: closed  reader for %s\n", inputFile)
		}
	}()
	log.Printf("debug: created reader for %s\n", inputFile)
	log.Printf("debug: completed input reader      in %v\n", time.Now().Sub(step))

	// unzip the input
	step = time.Now()
	src, err := adapters.GZipToUTF16(input)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("debug: completed unzip             in %v\n", time.Now().Sub(step))
	// cleanup the reader
	_ = input.Close()
	input = nil

	step = time.Now()
	filename := filepath.Join(debugOutputPath, "input-utf-16.xml")
	if err = os.WriteFile(filename, src, 0644); err != nil {
		return err
	}
	log.Printf("debug: created %s\n", filename)
	log.Printf("debug: completed input-utf-16.xml  in %v\n", time.Now().Sub(step))

	// convert input from UTF-16 to UTF-8
	step = time.Now()
	src, err = readers.ReadUTF16(src)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("debug: completed utf-16 to utf-8   in %v\n", time.Now().Sub(step))

	step = time.Now()
	filename = filepath.Join(debugOutputPath, "input-utf-8.xml")
	if err = os.WriteFile(filename, src, 0644); err != nil {
		return err
	}
	log.Printf("debug: created %s\n", filename)
	log.Printf("debug: completed input-utf-8.xml   in %v\n", time.Now().Sub(step))

	// read and convert the input
	step = time.Now()
	wmap, err := readers.ReadWXML(src)
	if err != nil {
		log.Printf("src %q\n", src[:35])
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("debug: read map from %s %v\n", inputFile, wmap.Version)
	log.Printf("debug: completed wxml conversion   in %v\n", time.Now().Sub(step))

	step = time.Now()
	filename = filepath.Join(debugOutputPath, "input.json")
	if b, err := json.MarshalIndent(wmap, "", "\t"); err != nil {
		return err
	} else if err = os.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	log.Printf("debug: created %s\n", filename)
	log.Printf("debug: completed input.json        in %v\n", time.Now().Sub(step))

	// todo: manipulate the input?

	step = time.Now()
	filename = filepath.Join(debugOutputPath, "output.json")
	if b, err := json.MarshalIndent(wmap, "", "\t"); err != nil {
		return err
	} else if err = os.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	log.Printf("debug: completed output.json       in %v\n", time.Now().Sub(step))
	log.Printf("debug: created %s\n", filename)

	step = time.Now()
	tmap, err := adapters.WMAPToTMAPv173(wmap)
	if err != nil {
		return err
	}
	log.Printf("debug: completed wmap to tmap      in %v\n", time.Now().Sub(step))

	step = time.Now()
	data, err := tmap.Encode()
	if err != nil {
		return err
	}
	log.Printf("debug: completed tmap to xml       in %v %d\n", time.Now().Sub(step), len(data))

	filename = filepath.Join(debugOutputPath, "output-utf-8.xml")
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	log.Printf("created %s\n", filename)
	log.Printf("debug: completed output-utf-8.xml  in %v\n", time.Now().Sub(started))

	step = time.Now()
	data, err = adapters.UTF8ToUTF16(data)
	if err != nil {
		return fmt.Errorf("%s: %w", inputFile, err)
	}
	log.Printf("debug: completed utf-8 to utf-16   in %v\n", time.Now().Sub(step))

	step = time.Now()
	filename = filepath.Join(debugOutputPath, "output-utf-16.xml")
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	log.Printf("created %s\n", filename)
	log.Printf("debug: completed output-utf-16.xml in %v\n", time.Now().Sub(step))

	step = time.Now()
	data, err = adapters.UTF16ToGZip(data)
	if err != nil {
		return err
	}
	log.Printf("debug: completed compress xml      in %v\n", time.Now().Sub(step))

	step = time.Now()
	filename = filepath.Join(debugOutputPath, "output.wxx")
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	log.Printf("created %s\n", filename)
	log.Printf("debug: completed output.wxx        in %v\n", time.Now().Sub(started))

	log.Printf("debug: completed                   in %v\n", time.Now().Sub(started))

	return nil
}
