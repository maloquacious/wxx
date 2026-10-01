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
	// app is the application version this encoder writes, verbatim map/@version.
	// It is required (see NewEncoder) and is not an option, because there is no
	// value it could default to that would not be a guess.
	app  string
	opts encoderOpts
}

type EncoderOption func(*encoderOpts)

type encoderOpts struct {
	compressedOutput bool
	gzipLevel        int
	utf16BeOutput    bool
	xmlHeader        bool
	diagnostics      *EncoderDiagnostics
}

type EncoderDiagnostics struct {
	Utf8Encoded   []byte // output after marshaling to UTF-8 XML
	WithXmlHeader []byte // output after inserting XML header
	Utf16Encoded  []byte // output after converting UTF-8 to UTF-16
	Compressed    []byte // output after running gzip

	// Dropped is the inventory of content the source map carried that the target
	// release cannot express (ADR 0004 Decision 7). It is empty when the encode
	// loses nothing, which today is every encode: the one supported schema
	// expresses everything Map_t models. It stays because the principle does --
	// always tell the user what they lose -- and W2025-to-W2025 loss, once one is
	// proven, is what it will report next (issues #92 and #73 found none between
	// 2.06, 2.07 and 2.08).
	//
	// Only MODELED losses appear here. A downgrade that would drop an unmodeled
	// stub does not report, it errors: the encoder can only stay quiet about a
	// loss it can enumerate. See downgradeLoss for the contract.
	//
	// This follows the project's diagnostics-over-logging convention, so it is
	// opt-in via WithEncoderDiagnostics. That is a real limit worth stating: a
	// caller who never asks does not hear about a modeled downgrade loss. It is
	// the enumerable, documented half of the loss -- the half a caller can
	// reconstruct from Map_t and the codec's COVERAGE.md after the fact -- and the
	// half that cannot be reconstructed is the half that errors.
	Dropped []DroppedFeature_t
}

// NewEncoder returns an Encoder that writes the supported application version
// app ("2.06") and implements the wxx.Encoder interface. Some
// features of the encoding pipeline can be configured with options.
//
// app is REQUIRED and is a parameter rather than an option (issue #45 Decision
// 1). You cannot create an encoder without saying what it writes. The encoder
// used to default to the release the map states -- Lookup(m.MetaData.Version.App)
// -- which made the SOURCE file's identity the target: a Map_t is the superset of
// every supported schema and it does not matter what wrote the file it was read
// from, so "what wrote the input" is provenance and is not an answer to "what
// should I write". A default here also meant the common call site never named a
// target at all, so nothing about it was checkable.
//
// app is an APPLICATION VERSION and nothing else: never a schema version, never a
// codec (issue #41 requirement 5). It names the one codec that writes it, and
// that codec supplies every byte of identity the file states -- map/@release,
// map/@version, map/@schema and the XML declaration -- so the identity a file
// claims and the format of its content cannot disagree.
//
// "" is not a sentinel. It names no supported application version, so it is the
// same error as any other unregistered one rather than a request for the old
// default. An empty version reaching here is a caller bug -- an unset flag, an
// empty config field -- and quietly writing the map's own release instead would
// hand back a file in a version nobody asked for.
//
// A caller that WANTS the map's own version is entitled to it and says so:
// NewEncoder(m.MetaData.Version.App.Raw). That is a client reading provenance and
// choosing a target, which is legitimate and is what cmd/copy does. What is
// forbidden is the encoder doing it silently on the caller's behalf.
//
// An app no codec accepts is not rejected here: an option-shaped constructor that
// cannot fail is the project's convention, and the error would arrive at Encode
// anyway. Encode resolves it before writing a byte, so a rejected target produces
// no file rather than a partial one.
func NewEncoder(app string, opts ...EncoderOption) *Encoder {
	e := &Encoder{
		app: app,
		opts: encoderOpts{
			compressedOutput: true,
			gzipLevel:        DefaultGzipLevel,
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

// DefaultGzipLevel is the gzip level Encode uses unless WithGzipLevel says
// otherwise. It is 6, the level compress/gzip's DefaultCompression stands for,
// so it writes the bytes wxx has always written.
const DefaultGzipLevel = 6

// WithGzipLevel sets the gzip compression level, from gzip.BestSpeed (1) to
// gzip.BestCompression (9); 0 stores the data uncompressed (issue #141).
// A lower level is faster and writes a larger file. Encode refuses a level
// outside 0..9 before writing anything.
func WithGzipLevel(level int) EncoderOption {
	return func(o *encoderOpts) {
		o.gzipLevel = level
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

func (e *Encoder) Encode(w io.Writer, m *wxx.Map_t) error {
	var err error

	// Resolve the codec before anything is written. An application version no
	// codec writes stops the encode here: the caller gets an error and w gets
	// nothing, rather than a file in a release that does not exist or that they
	// are not licensed for (ADR 0004 Decision 5).
	c, err := codecFor(e.app)
	if err != nil {
		return err
	}
	if e.opts.compressedOutput && (e.opts.gzipLevel < gzip.NoCompression || e.opts.gzipLevel > gzip.BestCompression) {
		return errors.Join(wxx.ErrGZipFailed, fmt.Errorf("gzip level %d: want %d..%d", e.opts.gzipLevel, gzip.NoCompression, gzip.BestCompression))
	}

	// Then the map. A Map_t whose fields contradict each other is not something
	// this pipeline can write, and until issue #20 the pipeline found that out by
	// panicking somewhere inside a codec -- a nil substructure dereferenced, or a
	// tile grid indexed past its end after part of the document had been emitted.
	// Validate is the whole of the check and it lives on Map_t, not here: it is a
	// statement about the model, so a caller can ask it BEFORE building an encoder
	// and get the same answer (issue #20).
	//
	// It runs before downgradeLoss, which walks the same map: the malformed map is
	// rejected by the check written to describe it rather than by whatever the
	// next reader trips over.
	if err := m.Validate(); err != nil {
		return err
	}

	// The codec's declaration is where every identity byte comes from, this file's
	// XML declaration included (issue #45). Nothing is read from m: what wrote the
	// input does not decide what we write.
	decl := c.AcceptedApps()

	// Inventory what this target cannot express, before anything is written. A
	// modeled loss is reported through diagnostics and the encode proceeds; a
	// loss the encoder cannot honestly describe -- an unmodeled stub -- stops it
	// here, so w gets nothing rather than a file quietly missing content
	// (ADR 0004 Decision 7; the contract is settled in downgradeLoss).
	dropped, err := downgradeLoss(m, decl.Schema)
	if err != nil {
		return err
	}
	if e.opts.diagnostics != nil {
		e.opts.diagnostics.Dropped = dropped
	}

	// marshal the Map_t to UTF‑8 XML. This calls the codec resolved above rather
	// than MarshalXML, which would resolve the same codec and run Validate and
	// downgradeLoss a second time: on a 1920 x 1080 map that repeat cost about
	// 0.2 s and 330 MB per encode (issue #142). Both checks have run above, so
	// MarshalXML's would add nothing.
	data, err := c.Encode(m, e.app)
	if err != nil {
		return err
	}
	if e.opts.diagnostics != nil {
		e.opts.diagnostics.Utf8Encoded = bdup(data)
	}

	var xmlHeader []byte
	if e.opts.xmlHeader {
		// The XML declaration follows the CODEC (W2025 opens 1.1). It is a byte the encoder writes, so the encoder owns it and
		// declares it: it is neither a switch on a family year nor registry data
		// (issue #45).
		if xmlHeader, err = xmlHeaderFor(decl.XMLVersion); err != nil {
			return err
		}
		if e.opts.diagnostics != nil {
			e.opts.diagnostics.WithXmlHeader = append(bdup(xmlHeader), data...)
		}
	}

	// The remaining stages stream into w: header and XML go through UTF-16 and
	// gzip as they are written, so the document is held once, as the codec's
	// UTF-8, rather than once per stage (issue #143). Every refusal has happened
	// by now, and neither the UTF-16 encoder nor gzip rejects its input, so the
	// only error left is one from w itself.
	//
	// The stages are built from the output end. A stage whose bytes diagnostics
	// want also writes them into a buffer, so the copies exist only when asked
	// for.
	out := w
	var compressed, utf16Encoded *bytes.Buffer
	var gz *gzip.Writer
	if e.opts.compressedOutput {
		if e.opts.diagnostics != nil {
			compressed = &bytes.Buffer{}
			out = io.MultiWriter(out, compressed)
		}
		if gz, err = gzip.NewWriterLevel(out, e.opts.gzipLevel); err != nil {
			return errors.Join(wxx.ErrGZipFailed, err)
		}
		out = gz
	}
	var utf16 io.WriteCloser
	if e.opts.utf16BeOutput {
		if e.opts.diagnostics != nil {
			utf16Encoded = &bytes.Buffer{}
			out = io.MultiWriter(out, utf16Encoded)
		}
		// encode as UTF-16/BE for Worldographer
		utf16 = transform.NewWriter(out, unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewEncoder())
		out = utf16
	}

	if _, err = out.Write(xmlHeader); err != nil {
		return err
	} else if _, err = out.Write(data); err != nil {
		return err
	}
	if utf16 != nil {
		if err = utf16.Close(); err != nil {
			return err
		}
	}
	if gz != nil {
		if err = gz.Close(); err != nil {
			return err
		}
	}

	if utf16Encoded != nil {
		e.opts.diagnostics.Utf16Encoded = utf16Encoded.Bytes()
	}
	if compressed != nil {
		e.opts.diagnostics.Compressed = compressed.Bytes()
	}

	return nil
}

// xmlHeaderFor returns the XML declaration to write ahead of a codec's XML.
//
// The version comes from the codec's declaration, and init has already refused to
// load a codec naming one no header exists for (see verifyXMLVersions), so the
// miss here is unreachable. It is still an error rather than a panic: a
// dispatcher that is wrong about which header to write must not write a file.
//
// It used to be Release_t.XMLHeader, and before that a switch on a family year
// (2017 -> 1.0, 2025 -> 1.1) -- the same coinage ADR 0004 deletes from the model
// -- which tied the bytes of every file to a label no classic file states and
// that a relabelled product would change without touching the format.
func xmlHeaderFor(xmlVersion string) ([]byte, error) {
	h, ok := utf16XMLHeader(xmlVersion)
	if !ok {
		return nil, errors.Join(wxx.ErrUnknownXMLHeader, fmt.Errorf("xml version %q: no header", xmlVersion))
	}
	return []byte(h), nil
}

// Parse/serialize the XML form of Map_t without transport concerns.

// MarshalXML converts a Map_t to the XML the supported application version app
// ("2.06") writes, without transport concerns (no header, no
// UTF-16, no gzip).
//
// app is an APPLICATION VERSION and nothing else. It is never a schema version
// and never a *Release_t, and no codec comes back: those were the three ways this
// function used to let a caller reach past the dispatcher and choose an encoder,
// which is what issue #41 requirement 5 denies. It resolves app exactly as Encode
// does -- both call codecFor -- so the two public encode paths cannot disagree
// about what a target is. "" is not a sentinel here either: it names no
// application version and is the same error as any other unregistered one.
//
// The APPLICATION VERSION picks the codec (issue #45 Decision 8). The schema does
// not, and no caller may name one: a caller who can pair a schema with an identity
// can build #41's chimera -- W2025 content declaring release="" version="1.77"
// schema="" -- and here that is unaskable, because naming the identity IS naming
// the codec. The one codec app resolves to both writes app and derives every other
// identity byte from it, so a file's declared identity and its content format come
// from a single source and cannot disagree.
//
// app is handed to the codec verbatim. It is never re-rendered from a Dotted's
// components, which would send "2.06" to disk as "2.6" (ADR 0004 Decision 1).
// The codec verifies it against the set it declares, which is redundant here --
// codecFor found the codec BY that set -- and is not redundant for the test units
// that call the codec directly (issue #41 requirement 3).
//
// The map is passed unaltered. It used to be Release_t.identify's shallow copy,
// with the target's identity stamped onto it, because the codec emitted the map's
// identity fields and a target that did not overwrite them first was a codec hint
// rather than a target. The codec now derives what it writes from app, so there is
// nothing left to stamp (issue #45).
//
// A downgrade that would drop an unmodeled stub is refused here too, and for the
// same reason it is refused in Encode: this is a public entry point, so leaving
// the check to Encode would leave a path that silently discards content the model
// never understood. The MODELED half of the loss is reported through
// EncoderDiagnostics, which this function has no access to -- a caller that needs
// the inventory calls Encode.
//
// Returns an error for an unsupported target or if the conversion fails.
func MarshalXML(m *wxx.Map_t, app string) ([]byte, error) {
	c, err := codecFor(app)
	if err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	if _, err := downgradeLoss(m, c.AcceptedApps().Schema); err != nil {
		return nil, err
	}
	return c.Encode(m, app)
}
