// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package readers

import (
	"bytes"
	"encoding/xml"
	"github.com/maloquacious/wxx/adapters"
	"github.com/maloquacious/wxx/models/wxml173"
	"github.com/maloquacious/wxx/models/wxx"
)

// ReadWXML loads a Worldographer data file (which normally has an extension of ".wxx").
// The input is verified to be in the expected format for a Worldographer file
// (UTF-16, big-endian format). If it is, the version number is extracted. If we do not
// find a version, we return an error. If we don't know how to unmarshall that version,
// we return an error. Otherwise, we unmarshal the data to a wxx.Map and return it.
func ReadWXML(src []byte) (*wxx.Map, error) {
	// verify the xml header
	xmlHeader := []byte("<?xml version='1.0' encoding='utf-16'?>\n")
	if !bytes.HasPrefix(src, xmlHeader) {
		return nil, ErrMissingXMLHeader
	}

	// remove the xml header
	src = src[len(xmlHeader):]

	// read the version from the xml data
	var version struct {
		Version string `xml:"version,attr"`
	}
	if err := xml.Unmarshal(src, &version); err != nil {
		return nil, err
	}
	// log.Printf("read: read version %+v\n", version)

	switch version.Version {
	case "1.73":
		return unmarshalV173(src)
	}
	return nil, ErrUnsupportedVersion
}

// unmarshalV173 unmarshalls XML data into a new wxx.Map structure.
// It assumes that the input is UTF-8 data and is compatible with version 1.73.
// Returns the new wxx.Map structure or an error.
func unmarshalV173(src []byte) (*wxx.Map, error) {
	srcMap := &wxml173.Map{}

	// convert from xml to a structure that's built just for the conversion
	if err := xml.Unmarshal(src, &srcMap); err != nil {
		return nil, err
	}

	// process source into a WXX structure and return it or any errors
	return adapters.WXMLToWXX(srcMap)
}
