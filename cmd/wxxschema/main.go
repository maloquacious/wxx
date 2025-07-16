// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a command line tool for dumping an XML schema.
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/readers"
	"io"
	"log"
	"os"
	"sort"
)

// this is very hacky. alternatives to investigate are
//  https://github.com/clbanning/mxj

var (
	Version = semver.Version{
		Major:      0,
		Minor:      1,
		Patch:      0,
		PreRelease: "alpha",
	}
)

func main() {
	log.Printf("wxxschema: %q: wxx %q\n", Version, wxx.Version())

	// read the input
	file, err := os.Open("input/blank-2025-1.10-1.01.wxx")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = file.Close()
	}()

	// uncompress the input. it should be UTF-16 encoded.
	utf16Data, err := readers.ReadCompressed(file)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("len(utf16Data) %d\n", len(utf16Data))

	// convert the UTF-16 to UTF-8
	xmlInput, err := readers.ReadUTF16(utf16Data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("len(xmlInput) %d\n", len(xmlInput))
	// verify the xml header was utf-16 and then force it to utf-8
	var xmlMeta struct{ release, version, schema string }
	xmlHeaderClassic := []byte("<?xml version='1.0' encoding='utf-16'?>\n")
	xmlHeader2025 := []byte("<?xml version='1.1' encoding='utf-16'?>\n")
	if bytes.HasPrefix(xmlInput, xmlHeaderClassic) {
		// NB: I'm making up the release, version, and schema for now
		xmlMeta.release, xmlMeta.version, xmlMeta.schema = "2017", "1.74", "1.0"
		xmlInput = append([]byte("<?xml version='1.0' encoding='utf-8'?>\n"), xmlInput[len(xmlHeaderClassic):]...)
	} else if bytes.HasPrefix(xmlInput, xmlHeader2025) {
		// NB: I'm making up the release, version, and schema for now
		xmlMeta.release, xmlMeta.version, xmlMeta.schema = "2025", "1.10", "1.01"
		xmlInput = append([]byte("<?xml version='1.0' encoding='utf-8'?>\n"), xmlInput[len(xmlHeader2025):]...)
	} else {
		if len(xmlInput) < 64 {
			log.Printf("xml: header %q\n", xmlInput)
		} else {
			log.Printf("xml: header %q\n", xmlInput[:64])
		}
		log.Fatal(readers.ErrMissingXMLHeader)
	}
	// write it out for analysis
	err = os.WriteFile("output/blank-2025-1.10-1.01.xml", xmlInput, 0600)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created 'output/blank-2025-1.10-1.01.xml'\n")

	root, err := inferSchema(bytes.NewReader(xmlInput))
	if err != nil {
		log.Fatal(err)
	}

	//generateSQL(root, os.Stdout)

	//fmt.Printf("\nXML Hierarchy\n")
	//generateHierarchy(root, false, 1, os.Stdout)

	fmt.Printf("\nXML Hierarchy with Attributes\n")
	generateHierarchy(root, true, 1, os.Stdout)
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
		_, _ = fmt.Fprintf(w, "CREATE TABLE %s (\n", child.Name)
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

func generateHierarchy(root *Element, withAttributes bool, level int, w io.Writer) {
	// sort the children before printing them so that we can have some consistency between versions
	var children []string
	for name := range root.Children {
		children = append(children, name)
	}
	sort.Strings(children)
	for _, name := range children {
		child := root.Children[name]
		_, _ = fmt.Fprintf(w, "%*s %-42s struct\n", level*2, "", child.Name)

		if withAttributes {
			// sort the attributes before printing them so that we can have some consistency between versions
			var attributes []string
			for attr := range child.Attributes {
				attributes = append(attributes, attr)
			}
			sort.Strings(attributes)
			for _, attr := range attributes {
				_, _ = fmt.Fprintf(w, "%*s %-42s string\n", (level+1)*2, "", "."+attr)
			}
		}
		// Recursively process children
		generateHierarchy(child, withAttributes, level+1, w)
	}
}
