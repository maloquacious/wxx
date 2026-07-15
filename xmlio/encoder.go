// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package xmlio

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/maloquacious/wxx"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Ensure *Encoder satisfies the wxx.Encoder interface contract at compile time.
var _ wxx.Encoder = (*Encoder)(nil)

type Encoder struct {
	opts encoderOpts
}

type EncoderOption func(*encoderOpts)

type encoderOpts struct {
	compressedOutput bool
	utf16BeOutput    bool
	xmlHeader        bool
	targetVersion    string // verbatim map/@version; "" means "the map's own"
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

// WithTargetVersion overrides the target Worldographer application version, as
// the verbatim map/@version string a supported release states ("1.77", "2.06").
// By default the encoder targets the map's own m.MetaData.Version.App.
//
// The registry resolves the rest of the target's identity from this string, so
// an unregistered version is an error rather than a best-effort write (ADR 0004
// Decision 5).
func WithTargetVersion(app string) EncoderOption {
	return func(o *encoderOpts) {
		o.targetVersion = app
	}
}

func (e *Encoder) Encode(w io.Writer, m *wxx.Map_t) error {
	var err error

	// Default the target to the application version the map states; callers may
	// override it with WithTargetVersion to re-target/convert the map. The
	// registry turns that string into a supported release -- an unknown one stops
	// here, before a single byte is written.
	targetVersion := m.MetaData.Version.App.Raw
	if e.opts.targetVersion != "" {
		targetVersion = e.opts.targetVersion
	}
	target, err := Lookup(targetVersion)
	if err != nil {
		return err
	}

	// marshal the Map_t to UTF‑8 XML
	data, err := MarshalXML(m, target)
	if err != nil {
		return err
	}
	if e.opts.diagnostics != nil {
		e.opts.diagnostics.Utf8Encoded = bdup(data)
	}

	if e.opts.xmlHeader {
		// The XML declaration follows the target release (classic opens 1.0,
		// W2025 opens 1.1). It is data on the registry entry, not a switch on a
		// family year.
		xmlHeader, err := target.XMLHeader()
		if err != nil {
			return err
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

// MarshalXML converts a Map_t to the XML of a target release, without transport
// concerns (no header, no UTF-16, no gzip). Resolve target with Lookup.
//
// The target's SCHEMA picks the codec (ADR 0004 Decision 4): the schema is the
// format's identity, it is on disk, and it does not change when the product is
// relabelled. The application version does not pick the codec -- it is
// caller-chosen data, and two application versions sharing a schema marshal
// through one codec and differ only in the string written to map/@version.
//
// Returns an error for an unsupported target or if the conversion fails.
func MarshalXML(m *wxx.Map_t, target *Release_t) ([]byte, error) {
	if target == nil {
		return nil, errors.Join(wxx.ErrUnsupportedMapVersion, fmt.Errorf("no target release"))
	}
	codec, err := CodecForSchema(target.Schema)
	if err != nil {
		return nil, err
	}
	return codec.Encode(m)
}
