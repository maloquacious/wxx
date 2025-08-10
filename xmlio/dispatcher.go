package xmlio

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx/models"
	"github.com/maloquacious/wxx/xmlio/h2017v1"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// ReadFile creates a Map from the data in the given file,
// which must have an extension of `.wxx`.
// Returns any errors reading or parsing the file contents.
// Uses ReadCompressXML to parse the file contents.
func ReadFile(path string) (*models.Map_t, error) {
	// file must have `.wxx` suffix
	if !strings.HasSuffix(path, ".wxx") {
		return nil, models.ErrMissingWxxExtension
	}
	// file must exist and be a regular file
	sb, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Join(models.ErrFSError, err)
		}
		return nil, models.ErrNotExists
	} else if sb.IsDir() || !sb.Mode().IsRegular() {
		return nil, models.ErrNotFile
	}
	// open the file for reading, returning any errors
	fp, err := os.Open(path)
	if err != nil {
		return nil, errors.Join(models.ErrFSError, err)
	}
	defer func(fp *os.File) {
		_ = fp.Close()
	}(fp)
	// read the compressed data and return a map or an error
	return ReadCompressedXML(fp)
}

// ReadCompressedXML creates a Map from the input or returns an error.
// The input must be compressed using GZip.
// Uses ReadUTF16XML to parse the uncompressed input.
func ReadCompressedXML(r io.Reader) (*models.Map_t, error) {
	// create a new gzip reader to process the source
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.Join(models.ErrInvalidGZip, err)
	}
	defer func(gzr *gzip.Reader) {
		_ = gzr.Close() // ignore errors
	}(gzr)
	data, err := io.ReadAll(gzr)
	if err != nil {
		return nil, errors.Join(models.ErrGUnZipFailed, err)
	}
	// read the data and return a Map or an error
	return ReadUTF16XML(bytes.NewReader(data))
}

// ReadUTF16XML creates a Map from the input or returns an error.
// The input must be UTF-16 encoded and will be decoded to UTF-8.
// Uses ReadUTF8XML to parse the decoded input.
func ReadUTF16XML(r io.Reader) (*models.Map_t, error) {
	// decode UTF-16 into UTF-8. we should verify that the input is actually UTF-16/BE,
	// but this package accepts both BE and LE. c'est la vie.
	utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
	data, err := io.ReadAll(transform.NewReader(r, utf16Encoding.NewDecoder()))
	if err != nil {
		return nil, errors.Join(models.ErrInvalidUTF16, err)
	}
	// read the data and return a Map or an error
	return ReadUTF8XML(bytes.NewReader(data))
}

// ReadUTF8XML creates a Map from the input or returns any error doing so.
// Verifies that XML header is version 1.0 or 1.1 and either UTF-16 or UTF8-8 encoded.
// We accept both encodings because we expect the caller to have already decoded UTF-16
// into UTF-8, and we don't require them to update the encoding in the header.
// We then extract metadata from the <map> element (the root element of the document).
// We use that metadata (version, release, and schema) to dispatch to the right XML
// unmarshaler.
func ReadUTF8XML(r io.Reader) (*models.Map_t, error) {
	// there should be a better way to get the header out of the input
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Join(models.ErrInvalidXML, err)
	}

	// verify the xml header. the encoding may be wrong, but we'll accept it.
	xmlHeaderIndex, xmlHeaders := -1, []struct {
		heading  string
		version  string
		encoding string
	}{
		{heading: "<?xml version='1.0' encoding='utf-8'?>\n", version: "1.0", encoding: "utf-8"},
		{heading: "<?xml version='1.0' encoding='utf-16'?>\n", version: "1.0", encoding: "utf-16"},
		{heading: "<?xml version='1.1' encoding='utf-8'?>\n", version: "1.1", encoding: "utf-8"},
		{heading: "<?xml version='1.1' encoding='utf-16'?>\n", version: "1.1", encoding: "utf-16"},
	}
	for i, header := range xmlHeaders {
		if bytes.HasPrefix(data, []byte(header.heading)) {
			xmlHeaderIndex = i
			break
		}
	}
	if xmlHeaderIndex == -1 {
		return nil, models.ErrMissingXMLHeader
	}
	// consume past the xml header since our unmarshal code expects only the XML
	data = data[len(xmlHeaders[xmlHeaderIndex].heading):]

	// quick sanity check on the input
	if !bytes.HasPrefix(data, []byte("<map ")) {
		return nil, errors.Join(models.ErrInvalidXML, models.ErrMissingMapElement)
	}

	// read the map metadata so that we'll know how to dispatch for parsing the data
	var xmlMetaData struct {
		Version string `xml:"version,attr"` // required
		Release string `xml:"release,attr"` // H2017 optional, W2025 required
		Schema  string `xml:"schema,attr"`  // H2017 optional, W2025 required
		buffer  []byte
	}
	// speed things up by extracting just the map attributes.
	// we have to make the map element self-closing for this to work.
	endOfMap := bytes.IndexByte(data, '>')
	if endOfMap == -1 {
		return nil, errors.Join(models.ErrInvalidXML, models.ErrMapNotClosed)
	}
	// initialize metadata with a copy of the source up to (but not including) the first closing '>'
	xmlMetaData.buffer = append(make([]byte, 0, endOfMap+1), data[:endOfMap]...)
	xmlMetaData.buffer = append(xmlMetaData.buffer, '/', '>')
	// now we can read the version from our copy of the xml data
	err = xml.Unmarshal(xmlMetaData.buffer, &xmlMetaData)
	if err != nil {
		return nil, errors.Join(models.ErrInvalidMapMetadata, err)
	}

	// use the metadata to call the correct unmarshaler for the XML
	switch xmlMetaData.Release + "/" + xmlMetaData.Version + "/" + xmlMetaData.Schema {
	case "/1.73/", "/1.74/", "/1.77/":
		return h2017v1.Read(data)
	case "2025/1.10/1.01":
		return nil, fmt.Errorf("2025/1.10/1.01: not yet implemented")
	}
	return nil, errors.Join(models.ErrUnsupportedMapMetadata, fmt.Errorf("map: release %q: schema %q: version %q", xmlMetaData.Release, xmlMetaData.Schema, xmlMetaData.Version))
}

func WriteFile(filename string, worldographerTargetVersion semver.Version, w *models.Map_t, utf8Filename string) error {
	fmt.Printf("debug: target version %s\n", worldographerTargetVersion.String())
	utf8XmlData, err := EncodeMapToXML(w, worldographerTargetVersion)
	if err != nil {
		return err
	}
	if utf8Filename != "" {
		var xmlHeader []byte
		switch worldographerTargetVersion.Major {
		case 2017:
			xmlHeader = []byte("<?xml version='1.0' encoding='utf-8'?>\n")
		case 2025:
			xmlHeader = []byte("<?xml version='1.1' encoding='utf-8'?>\n")
		default:
			return fmt.Errorf("unsupported worldographer version")
		}
		if err := os.WriteFile(utf8Filename, append(xmlHeader, utf8XmlData...), 0600); err != nil {
			return err
		}
	}

	// set the xml header and encode as utf-16/be for Worldographer
	var xmlHeader []byte
	switch worldographerTargetVersion.Major {
	case 2017:
		xmlHeader = []byte("<?xml version='1.0' encoding='utf-16'?>\n")
	case 2025:
		xmlHeader = []byte("<?xml version='1.1' encoding='utf-16'?>\n")
	default:
		return fmt.Errorf("unsupported worldographer version")
	}
	utf16XmlData, err := EncodeXMLToUTF16(append(xmlHeader, utf8XmlData...))
	if err != nil {
		return err
	}
	fmt.Printf("utf-16/be len %8d\n", len(utf16XmlData))

	// compress the encoded data, returning any errors
	gzipData, err := CompressUTF16(utf16XmlData)
	if err != nil {
		return err
	}

	// write the compressed data, returning any errors
	return os.WriteFile(filename, gzipData, 0600)
}

// EncodeMapToXML uses the target version to pick the right XML schema, then converts the Map_t to XML.
// Returns an error for unsupported versions or if there are errors during the conversion.
func EncodeMapToXML(w *models.Map_t, worldographerTargetVersion semver.Version) ([]byte, error) {
	switch worldographerTargetVersion.Major {
	case 2017:
		switch worldographerTargetVersion.Minor {
		case 1:
			return h2017v1.Encode(w)
		}
	}
	return nil, errors.Join(models.ErrUnsupportedSchemaVersion, fmt.Errorf("schema version: %s", worldographerTargetVersion.Short()))
}

// EncodeXMLToUTF16 adds the XML header and returns the data with UTF-16/BE encoding.
func EncodeXMLToUTF16(data []byte) ([]byte, error) {
	utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
	data, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), utf16Encoding.NewEncoder()))
	if err != nil {
		return nil, errors.Join(models.ErrInvalidUTF8, err)
	}
	return data, nil
}

// CompressUTF16 compresses the data.
func CompressUTF16(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	} else if err = gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
