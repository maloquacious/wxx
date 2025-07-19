// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements a command line tool that shows information
// on WXX data files.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/maloquacious/wxx/models"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"os"
	"strings"
)

func main() {
	for n, arg := range os.Args {
		if n == 0 {
			continue
		}
		fmt.Printf("%s\n", arg)
		if !strings.HasSuffix(arg, ".wxx") {
			fmt.Printf("\tnot a '.wxx' file\n")
			continue
		}
		sb, err := os.Stat(arg)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("\tdoes not exist\n")
			} else {
				fmt.Printf("\tunable to stat\n")
			}
			continue
		} else if sb.IsDir() {
			fmt.Printf("\tis a folder\n")
		} else if !sb.Mode().IsRegular() {
			fmt.Printf("\tis not a file\n")
		}
		fmt.Printf("\t%8d bytes on disk\n", sb.Size())
		input, err := os.ReadFile(arg)
		if err != nil {
			fmt.Printf("\tfailed to read\n")
		}

		// should be a gzip file
		input, err = unzip(input)
		if err != nil {
			fmt.Printf("\tnot gzip compressed\n")
		}
		fmt.Printf("\t%8d bytes compressed\n", sb.Size())
		fmt.Printf("\t%8d bytes uncompressed\n", len(input))

		// should be UTF-16/BE
		if len(input)%2 != 0 {
			fmt.Printf("\tnot utf-16/be encoded\n")
		}
		// verify the BOM
		if bytes.HasPrefix(input, []byte{0xfe, 0xff}) {
			fmt.Printf("\t%8d bytes utf-16/be encoded\n", len(input))
		} else if bytes.HasPrefix(input, []byte{0xff, 0xfe}) {
			fmt.Printf("\t%8d bytes utf-16/le encoded\n", len(input))
			continue
		} else {
			fmt.Printf("\tnot utf-16/be encoded\n")
			continue
		}

		// convert to UTF-8
		utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
		input, err = io.ReadAll(transform.NewReader(bytes.NewReader(input), utf16Encoding.NewDecoder()))
		fmt.Printf("\t%8d bytes utf-8 encoded\n", len(input))
		// verify the xml header. the encoding may be wrong, but we'll accept it.
		xmlHeaderIndex, xmlHeaders := -1, []string{
			"<?xml version='1.0' encoding='utf-8'?>\n",
			"<?xml version='1.0' encoding='utf-16'?>\n",
			"<?xml version='1.1' encoding='utf-8'?>\n",
			"<?xml version='1.1' encoding='utf-16'?>\n",
		}
		for i, xmlHeader := range xmlHeaders {
			if bytes.HasPrefix(input, []byte(xmlHeader)) {
				xmlHeaderIndex = i
				// trim the header so that we will be able to unmarshal the input
				input = input[len(xmlHeader):]
				break
			}
		}
		if xmlHeaderIndex == -1 {
			fmt.Printf("\tmissing xml header\n")
			continue
		}
		fmt.Printf("\txml header %q\n", xmlHeaders[xmlHeaderIndex])
		fmt.Printf("\t%8d bytes xml data\n", len(input))
		if !bytes.HasPrefix(input, []byte("<map ")) {
			fmt.Printf("\tmissing <map> element\n")
			continue
		}

		// read the map metadata
		xmlMetaData, err := readMapMetadata(input)
		if err != nil {
			fmt.Printf("\t%v\n", err)
			continue
		}
		if xmlMetaData.Release == "" && xmlMetaData.Version != "" && xmlMetaData.Schema == "" {
			// H2017 file
			fmt.Printf("\tH2017: version %s\n", xmlMetaData.Version)
		} else if xmlMetaData.Release == "2025" && xmlMetaData.Version != "" && xmlMetaData.Schema != "" {
			// W2025 file
			fmt.Printf("\tW2025: version %s: schema %s\n", xmlMetaData.Version, xmlMetaData.Schema)
		} else {
			fmt.Printf("\tunknown metadata: %q %q %q\n", xmlMetaData.Release, xmlMetaData.Version, xmlMetaData.Schema)
		}
	}
}

type mapMetaData struct {
	Version string `xml:"version,attr"` // required
	Release string `xml:"release,attr"` // H2017 optional, W2025 required
	Schema  string `xml:"schema,attr"`  // H2017 optional, W2025 required
}

// readMapMetadata
func readMapMetadata(input []byte) (mapMetaData, error) {
	// sanity check, sweet sanity checks
	if !bytes.HasPrefix(input, []byte(`<map `)) {
		return mapMetaData{}, fmt.Errorf("<map> element missing")
	}
	// speed up the remaining sanity checks by extracting the map attributes.
	// we have to make the map element self-closing for this to work.
	endOfMap := bytes.IndexByte(input, '>')
	if endOfMap == -1 {
		return mapMetaData{}, fmt.Errorf("<map> not closed")
	}
	// initialize metadata with a copy of the source up to (but not including) the first closing '>'
	metadata := append(make([]byte, 0, endOfMap+1), input[:endOfMap]...)
	metadata = append(metadata, '/', '>')
	// read the version from the xml data
	var results mapMetaData
	if err := xml.Unmarshal(metadata, &results); err != nil {
		return mapMetaData{}, errors.Join(models.ErrInvalidMapMetadata, err)
	}
	return results, nil
}

func unzip(input []byte) ([]byte, error) {
	// create a new gzip reader to process the source
	gzr, err := gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer func(gzr *gzip.Reader) {
		_ = gzr.Close() // ignore errors
	}(gzr)
	return io.ReadAll(gzr)
}
