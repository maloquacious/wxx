// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a command line tool for dumping an XML schema.
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/readers"
	"io"
	"log"
	"os"
)

func main() {
	log.Printf("wxx: package version %q\n", wxx.Version())

	// read the input
	file, err := os.Open("input/blank-30x21.wxx")
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
	if !bytes.HasPrefix(xmlInput, []byte("<?xml version='1.0' encoding='utf-16'?>\n")) {
		log.Fatal(readers.ErrMissingXMLHeader)
	} else {
		for n, ch := range []byte("<?xml version='1.0' encoding='utf-8'?> \n") {
			xmlInput[n] = ch
		}
	}
	// write it out for analysis
	err = os.WriteFile("output/blank-30x21.xml", xmlInput, 0600)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created 'output/blank-30x21.xml'\n")

	root, err := inferSchema(bytes.NewReader(xmlInput))
	if err != nil {
		log.Fatal(err)
	}

	generateSQL(root, os.Stdout)
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
		}
		if err != nil {
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
	for _, child := range root.Children {
		fmt.Fprintf(w, "CREATE TABLE %s (\n", child.Name)
		fmt.Fprint(w, "  id INTEGER PRIMARY KEY,\n")
		for attr := range child.Attributes {
			fmt.Fprintf(w, "  %s TEXT,\n", attr)
		}
		fmt.Fprint(w, "  parent_id INTEGER\n")
		fmt.Fprintln(w, ");\n")

		// Recursively process children
		generateSQL(child, w)
	}
}
