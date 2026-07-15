// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package xmlio

import (
	"bytes"
	"compress/gzip"
	"errors"
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

// targetKind records HOW the caller named the target release, which is not the
// same question as which release they named.
//
// The distinction exists because "" is a legal string and a caller who passes
// one has named a target -- badly. Storing only the string would make "" mean
// "the caller said nothing", so WithTargetVersion("") would silently write the
// map's own release instead of the one the caller asked for. That is the
// best-effort write ADR 0004 Decision 5 forbids, dressed up as a default.
type targetKind int

const (
	// targetFromMap is the default: no caller named a target, so the encoder
	// targets the release the map itself states.
	targetFromMap targetKind = iota
	// targetByVersion: the caller named a verbatim application version, which the
	// registry must resolve to a release.
	targetByVersion
	// targetByRelease: the caller named a release the registry already resolved.
	targetByRelease
)

type encoderOpts struct {
	compressedOutput bool
	utf16BeOutput    bool
	xmlHeader        bool
	targetKind       targetKind
	targetVersion    string     // verbatim map/@version, when targetKind is targetByVersion
	targetRelease    *Release_t // registry entry, when targetKind is targetByRelease
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

// WithTargetVersion targets the supported release that states app as its
// verbatim map/@version ("1.73", "1.77", "2.06"). Without it the encoder targets
// the release the map itself states in m.MetaData.Version.App.
//
// The caller names a release, never a tuple: the registry resolves map/@release
// and map/@schema from app, so a combination no release states -- @version="1.77"
// on a modern schema, say -- cannot be asked for at all (ADR 0004 Decision 5).
// An app no supported release states is an error at Encode, never a nearest
// match and never a best-effort write. This is the licensing requirement: a user
// licensed for "2.06" targets "2.06" and cannot be handed a "2.07" file.
//
// "" is not a sentinel. It names no supported release, so it is the same error
// as any other unregistered version rather than a request for the default. An
// empty version reaching here is a caller bug -- an unset flag, an empty config
// field -- and quietly writing the map's own release instead would hand back a
// file in a version nobody asked for.
func WithTargetVersion(app string) EncoderOption {
	return func(o *encoderOpts) {
		o.targetKind = targetByVersion
		o.targetVersion = app
		o.targetRelease = nil
	}
}

// WithTargetRelease targets a release the registry has already resolved, as
// returned by Lookup or SupportedReleases. It is the typed form of
// WithTargetVersion, for a caller that has a *Release_t in hand and would
// otherwise round-trip it back through its own App.Raw.
//
// Only the registry's own entries are accepted. A Release_t is an ordinary
// exported struct, so a caller CAN assemble one naming @version="1.77" with the
// W2025 schema; what it cannot do is get that past Encode, which rejects any
// entry the registry did not produce. Without that check this option would be
// the "any other path" by which an unregistered (App, Schema) pair reaches the
// encoder -- and the codec follows the schema, so the file would be a W2025 one
// claiming to be classic 1.77.
//
// A nil release is an error for the same reason "" is: it names nothing, and
// defaulting it would write a release the caller never asked for.
func WithTargetRelease(r *Release_t) EncoderOption {
	return func(o *encoderOpts) {
		o.targetKind = targetByRelease
		o.targetRelease = r
		o.targetVersion = ""
	}
}

// resolveTarget resolves the release this encode targets.
//
// Every path ends at the registry: the caller's version string, the caller's
// release entry, and the map's own version are three ways of naming a release
// and none of them is a way of describing one. An unregistered target stops the
// encode here, before a byte is written.
func (o *encoderOpts) resolveTarget(m *wxx.Map_t) (*Release_t, error) {
	switch o.targetKind {
	case targetByVersion:
		return Lookup(o.targetVersion)
	case targetByRelease:
		return Resolve(o.targetRelease)
	default:
		return Lookup(m.MetaData.Version.App.Raw)
	}
}

func (e *Encoder) Encode(w io.Writer, m *wxx.Map_t) error {
	var err error

	// Resolve the target before anything is written. A target the registry does
	// not state stops the encode here: the caller gets an error and w gets
	// nothing, rather than a file in a release that does not exist or that they
	// are not licensed for (ADR 0004 Decision 5).
	target, err := e.opts.resolveTarget(m)
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
// concerns (no header, no UTF-16, no gzip). Resolve target with Lookup or
// SupportedReleases; an entry the registry did not produce is rejected, so an
// unregistered (App, Schema) pair cannot reach a codec down this path either.
//
// The target's SCHEMA picks the codec (ADR 0004 Decision 4): the schema is the
// format's identity, it is on disk, and it does not change when the product is
// relabelled. The application version does not pick the codec -- it is
// caller-chosen data, and two application versions sharing a schema marshal
// through one codec and differ only in the string written to map/@version.
//
// Writing that string is what makes the target a target rather than a hint: the
// bytes describe the release the caller asked for. See Release_t.identify.
//
// Returns an error for an unsupported target or if the conversion fails.
func MarshalXML(m *wxx.Map_t, target *Release_t) ([]byte, error) {
	target, err := Resolve(target)
	if err != nil {
		return nil, err
	}
	codec, err := CodecForSchema(target.Schema)
	if err != nil {
		return nil, err
	}
	return codec.Encode(target.identify(m))
}
