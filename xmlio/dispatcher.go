package xmlio

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx/models"
	"github.com/maloquacious/wxx/xmlio/h2017v1"
	"log"
)

func Read(input []byte) (*models.Map, error) {
	// verify the xml header. the encoding may be wrong, but we'll accept it.
	for _, xmlHeader := range []string{
		"<?xml version='1.0' encoding='utf-16'?>\n",
		"<?xml version='1.1' encoding='utf-16'?>\n",
		"<?xml version='1.0' encoding='utf-8'?>\n",
		"<?xml version='1.1' encoding='utf-8'?>\n",
	} {
		if bytes.HasPrefix(input, []byte(xmlHeader)) {
			// trim the header and unmarshal the input
			return unmarshalXML(input[len(xmlHeader):])
		}
	}
	return nil, models.ErrMissingXMLHeader
}

func unmarshalXML(input []byte) (*models.Map, error) {
	metadata, err := readMapMetadata(input)
	if err != nil {
		return nil, err
	}
	log.Printf("read: metadata %+v\n", metadata)

	// use the metadata to call the correct unmarshaler for the XML
	switch metadata.Release + "/" + metadata.Version + "/" + metadata.Schema {
	case "/1.73/":
		return h2017v1.Read(input)
	case "/1.74/":
		return h2017v1.Read(input)
	case "2025/1.10/1.01":
	}
	return nil, errors.Join(models.ErrUnsupportedMapMetadata, fmt.Errorf("map: release %q: schema %q: version %q", metadata.Release, metadata.Schema, metadata.Version))
}

func Write(version semver.Version, data *models.Map) ([]byte, error) {
	panic("!implemented")
	//switch version.Major {
	//case 1:
	//	return h2017v1.Write(data)
	//case 2:
	//	return v2_0.Write(data)
	//default:
	//	return nil, errors.Join(models.ErrUnsupportedSchemaVersion, fmt.Errorf("schema version: %s", version))
	//}
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
		return mapMetaData{}, errors.Join(models.ErrInvalidXML, fmt.Errorf("<map> element missing"))
	}
	// speed up the remaining sanity checks by extracting the map attributes.
	// we have to make the map element self-closing for this to work.
	endOfMap := bytes.IndexByte(input, '>')
	if endOfMap == -1 {
		return mapMetaData{}, errors.Join(models.ErrInvalidXML, fmt.Errorf("<map> not closed"))
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
