// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package xmlio

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/models"
	"github.com/maloquacious/wxx/xmlio/h2017v1"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Decoder implements the wxx MapDecoder interface.
type Decoder struct {
	opts decoderOpts
}

type DecoderOption func(*decoderOpts)

type decoderOpts struct {
	compressedInput bool
	utf16BeInput    bool
	hasXmlHeader    bool
	fixXmlHeader    bool
	diagnostics     *Diagnostics
}

type Diagnostics struct {
	Raw          []byte // original input
	Uncompressed []byte // input after running gunzip
	Converted    []byte // input after converting UTF-16 to UTF-8
	XMLHeader    []byte // XML header that was removed
	XMLData      []byte
}

// NewDecoder with functional options
func NewDecoder(opts ...DecoderOption) *Decoder {
	d := &Decoder{
		opts: decoderOpts{
			compressedInput: true,
			utf16BeInput:    true,
			hasXmlHeader:    true,
			fixXmlHeader:    true,
			diagnostics:     nil,
		},
	}
	for _, opt := range opts {
		opt(&d.opts)
	}
	return d
}

// WithSkipUncompress skips the step for running gunzip on the input.
func WithSkipUncompress() DecoderOption { // expect gzip on input
	return func(o *decoderOpts) {
		o.compressedInput = false
	}
}

// WithDiagnostics captures data from each step of the decoding into buffers.
func WithDiagnostics(buf *Diagnostics) DecoderOption {
	return func(o *decoderOpts) {
		o.diagnostics = buf
	}
}

// WithUTF16BEInput sets the flag for running the UTF16/BE to UTF8 conversion on the input.
func WithUTF16BEInput(enabled bool) DecoderOption { // expect UTF-16/BE
	return func(o *decoderOpts) {
		o.utf16BeInput = enabled
	}
}

// WithFixXMLHeaderEncoding sets the flag for updating the encoding in the XML header.
func WithFixXMLHeaderEncoding(enabled bool) DecoderOption {
	return func(o *decoderOpts) {
		o.fixXmlHeader = enabled
	}
}

// Decode creates a Map_t from the input or returns an error.
func (d *Decoder) Decode(r io.Reader) (*models.Map_t, error) {
	// internal steps:
	// * ReadFile
	// * ReadCompressedXML
	// * ReadUTF16XML
	// * ReadUTF8XML
	// * * Verify XML Header
	// * * Verify root element is `map`
	// * * Consume XML Header
	// * * Read map metadata
	// * * xml.Unmarshal
	// * * Dispatch to version+schema specific Read
	// * * Return Map_t

	// read the entire input into memory
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Join(wxx.ErrRawReadFailed, err)
	}
	if d.opts.diagnostics != nil {
		d.opts.diagnostics.Raw = bdup(data)
	}

	if d.opts.compressedInput {
		// Uncompress the input by running gunzip on it.
		// Todo: verify that the input is actually gzip data.

		// Create a new gzip reader to process the source.
		// This will return an error if the input is not gzip data.
		gzr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, errors.Join(wxx.ErrGZipNewReaderFailed, err)
		}
		defer func(gzr *gzip.Reader) {
			_ = gzr.Close() // ignore errors closing this reader
		}(gzr)
		// Run gunzip on the input, returning any errors.
		data, err = io.ReadAll(gzr)
		if err != nil {
			return nil, errors.Join(wxx.ErrGUnZipFailed, err)
		}
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.Uncompressed = bdup(data)
		}
	}

	if d.opts.utf16BeInput {
		// decode UTF-16/BE into UTF-8

		// verify the BOM for UTF-16/BE
		if bytes.HasPrefix(data, []byte{0xfe, 0xff}) {
			// as expected
		} else if bytes.HasPrefix(data, []byte{0xff, 0xfe}) {
			return nil, wxx.ErrNotBigEndianUTF16Encoded
		} else {
			return nil, wxx.ErrMissingBOM
		}

		utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
		data, err = io.ReadAll(transform.NewReader(bytes.NewReader(data), utf16Encoding.NewDecoder()))
		if err != nil {
			return nil, errors.Join(wxx.ErrInvalidUTF16, err)
		}
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.Converted = bdup(data)
		}
	}

	// table of XML headers that we can accept
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

	// extract the XML header
	if d.opts.hasXmlHeader {
		// verify that we have an XML header before we extract it.
		// this will fail if the input is not UTF-8 encoded.
		if !bytes.HasPrefix(data, []byte("<?xml")) {
			return nil, wxx.ErrMissingXMLHeader
		}
		for i, header := range xmlHeaders {
			if bytes.HasPrefix(data, []byte(header.heading)) {
				xmlHeaderIndex = i
				break
			}
		}
		if xmlHeaderIndex == -1 {
			return nil, wxx.ErrInvalidXMLHeader
		}
		data = data[len(xmlHeaders[xmlHeaderIndex].heading):]
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.XMLHeader = bdup(data[:len(xmlHeaders[xmlHeaderIndex].heading)])
		}
		// consume the XML header since our unmarshal code expects only the XML data
		data = data[len(xmlHeaders[xmlHeaderIndex].heading):]
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.XMLHeader = bdup(data)
		}
	}

	// data is now clean UTF‑8 XML data with no header

	// quick sanity check on the input
	if !bytes.HasPrefix(data, []byte("<map ")) {
		return nil, errors.Join(wxx.ErrInvalidXML, wxx.ErrMissingMapElement)
	}

	// Read the map metadata so we know how to dispatch for parsing.
	var xmlMetaData struct {
		Version string `xml:"version,attr"` // required
		Release string `xml:"release,attr"` // H2017 optional, W2025 required
		Schema  string `xml:"schema,attr"`  // H2017 optional, W2025 required
		buffer  []byte
	}

	// Extract just the opening <map ...> tag and make it self-closing: <map .../>
	end := bytes.IndexByte(data, '>')
	if end == -1 {
		return nil, errors.Join(wxx.ErrInvalidXML, wxx.ErrMapNotClosed)
	}

	// If it’s already self-closing (`.../>`), keep it; otherwise append `/>`.
	if end > 0 && data[end-1] == '/' {
		xmlMetaData.buffer = append([]byte{}, data[:end+1]...)
	} else {
		xmlMetaData.buffer = append(append(make([]byte, 0, end+2), data[:end]...), '/', '>')
	}

	// Now xmlMetaData.buffer holds a self-contained <map .../> you can unmarshal.
	if err := xml.Unmarshal(xmlMetaData.buffer, &xmlMetaData); err != nil {
		return nil, errors.Join(wxx.ErrInvalidXML, err)
	}

	// read the version from our copy of the map attributes
	err = xml.Unmarshal(xmlMetaData.buffer, &xmlMetaData)
	if err != nil {
		return nil, errors.Join(wxx.ErrInvalidMapMetadata, err)
	}

	// use the metadata to call the correct unmarshaler for the XML
	switch xmlMetaData.Release + "/" + xmlMetaData.Version + "/" + xmlMetaData.Schema {
	case "/1.73/", "/1.74/", "/1.77/":
		return h2017v1.Read(data)
	case "2025/1.10/1.01":
		return nil, fmt.Errorf("2025/1.10/1.01: not yet implemented")
	}

	return nil, errors.Join(wxx.ErrUnsupportedMapMetadata, fmt.Errorf("map: release %q: schema %q: version %q", xmlMetaData.Release, xmlMetaData.Schema, xmlMetaData.Version))
}

// bdup returns a copy of the source
func bdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
