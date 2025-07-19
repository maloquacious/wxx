// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package readers

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/maloquacious/wxx/adapters"
	"github.com/maloquacious/wxx/models"
	"github.com/maloquacious/wxx/models/wxml173"
	"io"
)

// ReadWXML reads the XML data from Worldographer file.
// We extract the <map.version> element. If we do not find it, we return an error.
// If we don't know how to unmarshall that version, we return an error. Otherwise,
// we unmarshal the data to a wxx.Map and return it.
func ReadWXML(r io.Reader) (*models.Map, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// read the metadata from the xml data
	xmlMetaData, err := readMapMetadata(src)
	if err != nil {
		return nil, err
	}
	// log.Printf("read: read version %+v\n", xmlMetaData)

	switch xmlMetaData.Version {
	case "1.73":
		return unmarshalV173(src)
	}
	return nil, ErrUnsupportedVersion
}

// unmarshalV173 unmarshalls XML data into a new wxx.Map structure.
// It assumes that the input is UTF-8 data and is compatible with version 1.73.
// Returns the new wxx.Map structure or an error.
func unmarshalV173(src []byte) (*models.Map, error) {
	srcMap := &wxml173.Map{}

	// convert from xml to a structure that's built just for the conversion
	if err := xml.Unmarshal(src, &srcMap); err != nil {
		return nil, err
	}

	// process source into a WXX structure and return it or any errors
	return adapters.WXMLToWXX(srcMap)
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
