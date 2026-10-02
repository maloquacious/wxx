// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package xmlio

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Ensure *Decoder satisfies the wxx.Decoder interface contract at compile time.
var _ wxx.Decoder = (*Decoder)(nil)

// Decoder implements the wxx Decoder interface.
type Decoder struct {
	opts decoderOpts
}

type DecoderOption func(*decoderOpts)

type decoderOpts struct {
	compressedInput bool
	utf16BeInput    bool
	hasXmlHeader    bool
	fixXmlHeader    bool
	diagnostics     *DecoderDiagnostics
}

type DecoderDiagnostics struct {
	Raw          []byte // original input
	Uncompressed []byte // input after running gunzip; on an error, as much as was read before it
	Converted    []byte // input after converting UTF-16 to UTF-8
	XMLHeader    []byte // the XML declaration that was removed, and only that
	XMLData      []byte // everything after the declaration: the XML handed to the codec
	MapElement   []byte

	// Codec names the codec package that decoded the file: "v1_06". It is a
	// diagnostic label, not identity -- see ADR 0004 and #44. It is set only
	// once dispatch has chosen, so it stays empty when the file's metadata
	// matches no codec, a refused classic file included.
	Codec string

	// Schema is the schema version the FILE stated in map/@schema, verbatim.
	// A classic file states no schema at all, so this is empty for one; that
	// absence is the honest answer, not a missing value. It is set from the
	// parsed metadata before dispatch, so it is populated even when Decode
	// goes on to reject the file as unsupported.
	Schema string

	// Clamped lists every value the decoder changed to bring it into range
	// (issue #124). Each is a LOSS: the returned Map_t does not hold what the
	// file said, and encoding it writes the clamped value. It is empty when the
	// decode changed nothing, which is every file the app itself saved, since
	// the app does not write an out-of-range value.
	//
	// Like everything here it is opt-in, via WithDecoderDiagnostics, and that is
	// a real limit: a caller who never asks is not told. A caller that writes a
	// decoded map back out should ask, and tell its user; cmd/copy, cmd/resize,
	// cmd/merge, cmd/import and cmd/info print each one to stderr. See
	// ClampedValue_t for what is clamped and why.
	Clamped []ClampedValue_t
}

// NewDecoder returns a Decoder that implements the wxx.Decoder interface.
// Some features of the decoding pipeline can be configured with options.
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

// WithDecoderDiagnostics captures data from each step of the decoding into buffers.
func WithDecoderDiagnostics(buf *DecoderDiagnostics) DecoderOption {
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
func (d *Decoder) Decode(r io.Reader) (*wxx.Map_t, error) {
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

	// The transport stages -- gunzip, then UTF-16 to UTF-8 -- are a chain of
	// readers, and only the UTF-8 at the end of it is held in memory (issue
	// #152). Each stage used to read its whole output before the next began,
	// so a decode held the gzip input, the UTF-16 and the UTF-8 at once. The
	// magic-number and BOM checks look at the first bytes of a stage through a
	// bufio peek, before anything after them is read.
	//
	// Diagnostics are opt-in and cost what they always did: Raw is read whole
	// before the chain starts, and Uncompressed is copied as the chain reads it.
	src := r
	if d.opts.diagnostics != nil {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, errors.Join(wxx.ErrRawReadFailed, err)
		}
		d.opts.diagnostics.Raw = data
		src = bytes.NewReader(data)
	}
	src = stageReader{r: src, stage: wxx.ErrRawReadFailed}

	if d.opts.compressedInput {
		// Uncompress the input by running gunzip on it.

		// verify that the input is actually gzip data by looking for the magic number.
		br := bufio.NewReader(src)
		magic, err := br.Peek(2)
		if err != nil && err != io.EOF {
			return nil, err // the raw input failed; stageReader says so
		}
		if !(len(magic) == 2 && magic[0] == 0x1F && magic[1] == 0x8B) {
			return nil, wxx.ErrNotCompressed
		}

		// Create a new gzip reader to process the source.
		// This will return an error if the input is not gzip data.
		gzr, err := gzip.NewReader(br)
		if err != nil {
			return nil, errors.Join(wxx.ErrGZipNewReaderFailed, err)
		}
		defer func(gzr *gzip.Reader) {
			_ = gzr.Close() // ignore errors closing this reader
		}(gzr)
		src = stageReader{r: gzr, stage: wxx.ErrGUnZipFailed}
		if d.opts.diagnostics != nil {
			uncompressed := &bytes.Buffer{}
			src = io.TeeReader(src, uncompressed)
			// set when Decode returns, by which time the chain has read all
			// of it, or as much as it got through before failing
			defer func() { d.opts.diagnostics.Uncompressed = uncompressed.Bytes() }()
		}
	}

	var data []byte
	var err error
	if d.opts.utf16BeInput {
		// decode UTF-16/BE into UTF-8

		// verify the BOM for UTF-16/BE
		br := bufio.NewReader(src)
		bom, err := br.Peek(2)
		if err != nil && err != io.EOF {
			return nil, stageError(nil, err) // an earlier stage failed
		}
		if bytes.HasPrefix(bom, []byte{0xfe, 0xff}) {
			// as expected
		} else if bytes.HasPrefix(bom, []byte{0xff, 0xfe}) {
			return nil, wxx.ErrNotBigEndianUTF16Encoded
		} else {
			return nil, wxx.ErrMissingBOM
		}

		utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
		data, err = io.ReadAll(transform.NewReader(br, utf16Encoding.NewDecoder()))
		if err != nil {
			return nil, stageError(wxx.ErrInvalidUTF16, err)
		}
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.Converted = bdup(data)
		}
	} else {
		data, err = io.ReadAll(src)
		if err != nil {
			return nil, stageError(nil, err)
		}
	}

	// extract the XML header
	xmlHeaderIndex := -1 // sentinel value meaning no header found
	if d.opts.hasXmlHeader {
		// verify that we have an XML header before we extract it.
		// this will fail if the input is not UTF-8 encoded.
		if !bytes.HasPrefix(data, []byte("<?xml")) {
			return nil, wxx.ErrMissingXMLHeader
		}
		foundValidHeader := false
		for i, header := range xmlHeaders {
			if bytes.HasPrefix(data, []byte(header.heading)) {
				foundValidHeader, xmlHeaderIndex = true, i
				break
			}
		}
		if !foundValidHeader {
			return nil, wxx.ErrInvalidXMLHeader
		}
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.XMLHeader = bdup(data[:len(xmlHeaders[xmlHeaderIndex].heading)])
		}
		// consume the XML header since our unmarshal code expects only the XML data
		data = data[len(xmlHeaders[xmlHeaderIndex].heading):]
	}

	// data is now clean UTF‑8 XML data with no header
	if d.opts.diagnostics != nil {
		d.opts.diagnostics.XMLData = bdup(data)
	}

	// quick sanity check on the input
	if !bytes.HasPrefix(data, []byte("<map ")) {
		return nil, errors.Join(wxx.ErrInvalidXML, wxx.ErrMissingMapElement)
	}

	// Read the map metadata so we know how to dispatch for parsing.
	var xmlMetaData struct {
		Version string `xml:"version,attr"` // required
		Release string `xml:"release,attr"` // W2025 required; absent from a classic file
		Schema  string `xml:"schema,attr"`  // W2025 required; absent from a classic file
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
	// Now xmlMetaData.buffer holds a self-contained <map .../>.
	if d.opts.diagnostics != nil {
		d.opts.diagnostics.MapElement = bdup(xmlMetaData.buffer)
	}
	// unmarshal that metadata
	if err := xml.Unmarshal(xmlMetaData.buffer, &xmlMetaData); err != nil {
		return nil, errors.Join(wxx.ErrInvalidXML, err)
	}

	// read the version from our copy of the map attributes
	err = xml.Unmarshal(xmlMetaData.buffer, &xmlMetaData)
	if err != nil {
		return nil, errors.Join(wxx.ErrInvalidMapMetadata, err)
	}

	// Report the schema the file itself stated, before dispatch and whatever
	// dispatch decides. A classic file states none, which leaves this empty.
	if d.opts.diagnostics != nil {
		d.opts.diagnostics.Schema = xmlMetaData.Schema
	}

	// use the metadata to call the correct decoder for the XML
	switch xmlMetaData.Release {
	case "2025":
		// any W2025 build (release=2025) routes to the v1_06 decoder;
		// version/schema no longer gate the dispatch.
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.Codec = "v1_06"
		}
		m, clamps, err := v1_06.Decode(data)
		if d.opts.diagnostics != nil {
			d.opts.diagnostics.Clamped = clampedValues(clamps)
		}
		return m, err
	case "":
		// H2017 ("classic") files carry no release or schema attribute; they
		// are identified solely by a "1.x" version (e.g. 1.73/1.74/1.77).
		// wxx no longer reads them (issue #103): Worldographer 2025 converts
		// a classic map itself, and knows its own format, so the file is
		// refused before any of it is decoded rather than read best-effort.
		if strings.HasPrefix(xmlMetaData.Version, "1.") {
			return nil, errors.Join(wxx.ErrClassicMap, fmt.Errorf("map: version %q: no release", xmlMetaData.Version))
		}
	}

	// Anything else -- including no release with a version that is not "1.x"
	// -- is a format we do not know, and is refused rather than guessed at.
	return nil, errors.Join(wxx.ErrUnsupportedMapMetadata, fmt.Errorf("map: release %q: version %q: schema %q", xmlMetaData.Release, xmlMetaData.Version, xmlMetaData.Schema))
}

// bdup returns a copy of the source
func bdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// stageReader tags a read error with the transport stage it came from. The
// stages are chained readers, so an error from the raw input or from gunzip
// surfaces from whichever stage is reading at the end of the chain; the tag
// lets Decode report it as the stage that failed, as it did when each stage
// was read whole on its own (issue #152).
type stageReader struct {
	r     io.Reader
	stage wxx.Error
}

func (s stageReader) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	if err != nil && err != io.EOF {
		err = errors.Join(s.stage, err)
	}
	return n, err
}

// stageError returns err from the end of the transport chain. An error a
// stageReader tagged already names its stage; any other error is the last
// stage's own, and is joined with that stage's error, if it has one.
func stageError(stage error, err error) error {
	if errors.Is(err, wxx.ErrRawReadFailed) || errors.Is(err, wxx.ErrGUnZipFailed) || stage == nil {
		return err
	}
	return errors.Join(stage, err)
}
