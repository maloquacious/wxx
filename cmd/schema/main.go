// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a command line tool for dumping an XML schema.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/maloquacious/wxx"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

// this is very hacky. alternatives to investigate are
//  https://github.com/clbanning/mxj

func main() {
	fmt.Printf("wxx: version %q\n", wxx.Version())

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
			continue
		}

		root, err := inferSchema(bytes.NewReader(input))
		if err != nil {
			log.Fatal(err)
		}

		//generateSQL(root, os.Stdout)

		fmt.Printf("\nXML Hierarchy\n")
		generateHierarchy(root, 1, os.Stdout)
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
		return mapMetaData{}, errors.Join(wxx.ErrInvalidMapMetadata, err)
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

type Element struct {
	Name       string
	Attributes map[string]struct{}
	Children   map[string]*Element
}

func inferSchema(input io.Reader) (*Element, error) {
	root := &Element{Children: map[string]*Element{}}
	stack := []*Element{root}

	decoder := xml.NewDecoder(input)
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			elem := &Element{
				Name:       tok.Name.Local,
				Attributes: map[string]struct{}{},
				Children:   map[string]*Element{},
			}
			for _, attr := range tok.Attr {
				elem.Attributes[attr.Name.Local] = struct{}{}
			}
			parent := stack[len(stack)-1]
			parent.Children[elem.Name] = elem
			stack = append(stack, elem)

		case xml.EndElement:
			stack = stack[:len(stack)-1]
		}
	}
	return root, nil
}

func generateSQL(root *Element, w io.Writer) {
	// sort the children before printing them so that we can have some consistency between versions
	var children []string
	for name := range root.Children {
		children = append(children, name)
	}
	sort.Strings(children)
	for _, name := range children {
		child := root.Children[name]
		_, _ = fmt.Fprintf(w, "CREATE"+" "+"TABLE %s (\n", child.Name)
		_, _ = fmt.Fprintf(w, "  %-42s INTEGER PRIMARY KEY,\n", "id")
		// sort the attributes before printing them so that we can have some consistency between versions
		var attributes []string
		for attr := range child.Attributes {
			attributes = append(attributes, attr)
		}
		sort.Strings(attributes)
		for _, attr := range attributes {
			_, _ = fmt.Fprintf(w, "  %-42s TEXT,\n", attr)
		}
		_, _ = fmt.Fprint(w, "  parent_id INTEGER\n")
		_, _ = fmt.Fprintln(w, ");\n")

		// Recursively process children
		generateSQL(child, w)
	}
}

func generateHierarchy(root *Element, level int, w io.Writer) {
	// sort the children before printing them so that we can have some consistency between versions
	var children []string
	for name := range root.Children {
		children = append(children, name)
	}
	sort.Strings(children)
	for _, name := range children {
		child := root.Children[name]
		_, _ = fmt.Fprintf(w, "%*s %-42s struct {\n", level*2, "", child.Name)

		// sort the attributes before printing them so that we can have some consistency between versions
		var attributes []string
		for attr := range child.Attributes {
			attributes = append(attributes, attr)
		}
		sort.Strings(attributes)
		for _, attr := range attributes {
			_, _ = fmt.Fprintf(w, "%*s %-42s string\n", (level+1)*2, "", attr)
		}

		// Recursively process children
		generateHierarchy(child, level+1, w)

		_, _ = fmt.Fprintf(w, "%*s } // %s\n", level*2, "", child.Name)
	}
}
