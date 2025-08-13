// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package xmlio

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/h2017v1"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type Encoder struct {
	opts encoderOpts
}

type EncoderOption func(*encoderOpts)

type encoderOpts struct {
	compressedOutput bool
	utf16BeOutput    bool
	xmlHeader        bool
	diagnostics      *EncoderDiagnostics
}

type EncoderDiagnostics struct {
	Utf8Encoded   []byte // output after marshaling to UTF-8 XML
	WithXmlHeader []byte // output after inserting XML header
	Utf16Encoded  []byte // output after converting UTF-8 to UTF-16
	Compressed    []byte // output after running gzip
	Schema        string
}

// NewEncoder returns an Encoder that implements the wxx.Encoder interface.
// Some features of the encoding pipeline can be configured with options.
func NewEncoder(opts ...EncoderOption) *Encoder {
	e := &Encoder{
		opts: encoderOpts{
			compressedOutput: true,
			utf16BeOutput:    true,
			xmlHeader:        true,
			diagnostics:      nil,
		},
	}
	for _, opt := range opts {
		opt(&e.opts)
	}
	return e
}

// WithEncoderDiagnostics captures data from each step of the decoding into buffers.
func WithEncoderDiagnostics(buf *EncoderDiagnostics) EncoderOption {
	return func(o *encoderOpts) {
		o.diagnostics = buf
	}
}

func WithGzipOutput(enabled bool) EncoderOption {
	return func(o *encoderOpts) {
		o.compressedOutput = enabled
	}
}

func WithUTF16BEOutput(enabled bool) EncoderOption {
	return func(o *encoderOpts) {
		o.utf16BeOutput = enabled
	}
}

func WithXMLHeader(enabled bool) EncoderOption {
	return func(o *encoderOpts) {
		o.xmlHeader = enabled
	}
}

func (e *Encoder) Encode(w io.Writer, worldographerTargetVersion semver.Version, m *wxx.Map_t) error {
	var err error

	// marshal the Map_t to UTF‑8 XML
	data, err := MarshalXML(m, worldographerTargetVersion)
	if err != nil {
		return err
	}
	if e.opts.diagnostics != nil {
		e.opts.diagnostics.Utf8Encoded = bdup(data)
	}

	if e.opts.xmlHeader {
		var xmlHeader []byte
		switch worldographerTargetVersion.Major {
		case 2017:
			xmlHeader = []byte("<?xml version='1.0' encoding='utf-16'?>\n")
		case 2025:
			xmlHeader = []byte("<?xml version='1.1' encoding='utf-16'?>\n")
		default:
			return fmt.Errorf("unsupported worldographer version")
		}
		buf := make([]byte, 0, len(xmlHeader)+len(data))
		buf = append(buf, xmlHeader...)
		buf = append(buf, data...)
		data = buf
		if e.opts.diagnostics != nil {
			e.opts.diagnostics.WithXmlHeader = bdup(data)
		}
	}

	if e.opts.utf16BeOutput {
		// encode as UTF-16/BE for Worldographer
		utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
		data, err = io.ReadAll(transform.NewReader(bytes.NewReader(data), utf16Encoding.NewEncoder()))
		if err != nil {
			return errors.Join(wxx.ErrInvalidUTF8, err)
		}
		if e.opts.diagnostics != nil {
			e.opts.diagnostics.Utf16Encoded = bdup(data)
		}
	}

	if e.opts.compressedOutput {
		// compress the encoded data, returning any errors
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(data); err != nil {
			return err
		} else if err = gz.Close(); err != nil {
			return err
		}
		data = bdup(buf.Bytes())
		if e.opts.diagnostics != nil {
			e.opts.diagnostics.Utf16Encoded = bdup(data)
		}
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Parse/serialize the XML form of Map_t without transport concerns.

// MarshalXML uses the target version to pick the right XML schema, then converts the Map_t to XML.
// Returns an error for unsupported versions or if there are errors during the conversion.
func MarshalXML(m *wxx.Map_t, worldographerTargetVersion semver.Version) ([]byte, error) {
	switch worldographerTargetVersion.Major {
	case 2017:
		switch worldographerTargetVersion.Minor {
		case 1:
			return h2017v1.Encode(m)
		}
	}
	return nil, errors.Join(wxx.ErrUnsupportedSchemaVersion, fmt.Errorf("schema version: %s", worldographerTargetVersion.Short()))
}

func MarshalXMLBytes(m *wxx.Map_t) ([]byte, error) { // handy for tests
	panic("example only")
}
