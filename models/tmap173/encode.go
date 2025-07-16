// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tmap173

import (
	"bytes"
	_ "embed"
	"text/template"
)

var (
	//go:embed "xml.gohtml"
	xmlTemplate string
)

// Encode marshals the Map to XML using custom templates.
func (m *Map) Encode() ([]byte, error) {
	t, err := template.New("xml-1.73").Parse(xmlTemplate)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	err = t.Execute(b, m)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
