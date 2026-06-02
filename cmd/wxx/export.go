// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/maloquacious/wxx"
	"github.com/peterbourgon/ff/v4"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// newExportCommand returns the `wxx export` subcommand.
//
// `wxx export <wxx-file>` reads a Worldographer file and writes requested
// artifacts to disk. The input file argument is required.
//
// Optional flags select what to export:
//
//		--utf-8 <file>   write the UTF-8 bytes of the XML payload to <file>.
//		                 The XML declaration is preserved verbatim, including
//		                 its original encoding='utf-16' attribute, so the
//		                 output is intentionally "a lie" suitable for
//		                 diffing schemas across Worldographer versions.
//
//		--xml <file>     write the XML payload to <file> with original encoding
//	                  preserved.
func newExportCommand(parent *ff.FlagSet) *ff.Command {
	fs := ff.NewFlagSet("export").SetParent(parent)
	rawOut := fs.String('r', "raw", "", "write the uncompress payload to this file with original encoding")
	utf8Out := fs.String('u', "utf-8", "", "write the UTF-8 bytes of the XML payload to this file")

	return &ff.Command{
		Name:      "export",
		Usage:     "wxx export [flags] <wxx-file>",
		ShortHelp: "export content from a Worldographer WXX file",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			switch len(args) {
			case 0:
				return fmt.Errorf("export: missing required <wxx-file> argument")
			case 1:
				// ok
			default:
				return fmt.Errorf("export: expected exactly one <wxx-file> argument, got %d", len(args))
			}
			return runExport(args[0], *rawOut, *utf8Out)
		},
	}
}

func runExport(inputPath, rawContentOut, utf8Out string) error {
	// 1. Read the file and verify gzip magic.
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return errors.Join(wxx.ErrRawReadFailed, fmt.Errorf("read %s: %w", inputPath, err))
	}
	if !(len(data) >= 2 && data[0] == 0x1F && data[1] == 0x8B) {
		return fmt.Errorf("%s: %w", inputPath, wxx.ErrNotCompressed)
	}

	// 2. Uncompress.
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return errors.Join(wxx.ErrGZipNewReaderFailed, fmt.Errorf("%s: %w", inputPath, err))
	}
	defer func() { _ = gzr.Close() }()
	data, err = io.ReadAll(gzr)
	if err != nil {
		return errors.Join(wxx.ErrGUnZipFailed, fmt.Errorf("%s: %w", inputPath, err))
	}

	createdFiles := 0
	if rawContentOut != "" {
		if err := exportRawContent(data, rawContentOut); err != nil {
			return err
		}
		createdFiles++
	}

	if utf8Out != "" {
		if err := exportUTF8(data, utf8Out); err != nil {
			return fmt.Errorf("%s: %w", inputPath, err)
		}
		createdFiles++
	}

	if createdFiles == 0 {
		return fmt.Errorf("export: nothing to do; pass an output file format on the command line")
	}
	return nil
}

// exportRawContent writes the raw content from a Worldographer file to
// outputPath.
func exportRawContent(data []byte, outputPath string) error {
	// Write the content.
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("export: write %s: %w", outputPath, err)
	}
	fmt.Printf("export: wrote raw content to %s (%d bytes)\n", outputPath, len(data))
	return nil
}

// exportUTF8 converts the content of a Worldographer file from UTF16/BE
// to UTF8 and writes it to outputPath.
//
// Worldographer files are assumed to be UTF-16 big-endian XML.
//
//  1. Confirm a UTF-16 big-endian BOM (0xFE 0xFF).
//  2. Convert UTF-16/BE to UTF-8.
//  3. Write the UTF-8 bytes to outputPath.
func exportUTF8(data []byte, outputPath string) error {
	// 1. Verify the UTF-16/BE BOM.
	switch {
	case bytes.HasPrefix(data, []byte{0xfe, 0xff}):
		// UTF-16 big-endian, as expected.
	case bytes.HasPrefix(data, []byte{0xff, 0xfe}):
		return wxx.ErrNotBigEndianUTF16Encoded
	default:
		return wxx.ErrMissingBOM
	}

	// 2. Convert UTF-16/BE to UTF-8.
	utf16BE := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
	utf8, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), utf16BE.NewDecoder()))
	if err != nil {
		return wxx.ErrInvalidUTF16
	}

	// 3. Write the UTF-8 bytes.
	if err := os.WriteFile(outputPath, utf8, 0o644); err != nil {
		return fmt.Errorf("export: write %s: %w", outputPath, err)
	}
	fmt.Printf("export: wrote UTF-8 to %s (%d bytes)\n", outputPath, len(utf8))
	return nil
}
