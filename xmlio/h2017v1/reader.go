// Package h2017v1 implements an XML reader for the H2017.V1 XML schema
package h2017v1

import (
	"encoding/xml"
	"github.com/maloquacious/wxx/adapters"
	"github.com/maloquacious/wxx/models"
)

// Read unmarshalls XML data using the H2017.V1 schema and returns the internal Map or an error.
func Read(input []byte) (*models.Map, error) {
	srcMap := &Map{}

	// unmarshal into a structure that's built just for the conversion
	if err := xml.Unmarshal(input, &srcMap); err != nil {
		return nil, err
	}

	// process source into a WXX structure and return it or any errors
	return adapters.WXMLToWXX(srcMap)
}
