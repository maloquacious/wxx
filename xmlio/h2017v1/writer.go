// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2017v1

import (
	"bytes"
	_ "embed"

	"fmt"
	"github.com/maloquacious/wxx/models"
	"text/template"
)

var (
	//go:embed "writer.goxml"
	xmlTemplate string
)

// MarshalXML marshalls the Map_t to XML data using the H2017.V1 schema and returns the slice or an error.
func MarshalXML(w *models.Map_t) ([]byte, error) {
	return nil, fmt.Errorf("h2017v1: marshallXML: not implemented")
}

// Encode marshals the Map_t to XML using custom templates.
func Encode(w *models.Map_t) ([]byte, error) {
	t, err := template.New("h2017v1").Parse(xmlTemplate)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	err = t.Execute(b, w)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
