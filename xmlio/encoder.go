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
)

type encoderOpts struct {
	compressedOutput bool
	utf16BeOutput    bool
	xmlHeader        bool
	targetKind       targetKind
	targetVersion    string // verbatim map/@version, when targetKind is targetByVersion
	diagnostics      *EncoderDiagnostics
}

type EncoderDiagnostics struct {
	Utf8Encoded   []byte // output after marshaling to UTF-8 XML
	WithXmlHeader []byte // output after inserting XML header
	Utf16Encoded  []byte // output after converting UTF-8 to UTF-16
	Compressed    []byte // output after running gzip
	Schema        string

	// Dropped is the inventory of content the source map carried that the target
	// release cannot express (ADR 0004 Decision 7). It is empty when the encode
	// loses nothing -- notably when the target is the release the map already
	// states, which is the default.
	//
	// Only MODELED losses appear here. A downgrade that would drop an unmodeled
	// stub does not report, it errors: the encoder can only stay quiet about a
	// loss it can enumerate. See downgradeLoss for the contract.
	//
	// This follows the project's diagnostics-over-logging convention, so it is
	// opt-in via WithEncoderDiagnostics. That is a real limit worth stating: a
	// caller who never asks does not hear about a modeled downgrade loss. It is
	// the enumerable, documented half of the loss -- the half a caller can
	// reconstruct from Map_t and internal/v0_77/COVERAGE.md after the fact -- and the
	// half that cannot be reconstructed is the half that errors.
	Dropped []DroppedFeature_t
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
// It is the ONE way to name a target, which is issue #41 requirement 5: a caller
// may request an application version and never an encoder. WithTargetRelease(r)
// used to be a second way; once Release_t stopped carrying a codec it was exactly
// WithTargetVersion(r.App.Raw), so it was redundancy rather than a typed
// alternative, and two ways to say one thing is two things to keep honest.
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
	}
}

// resolveTarget resolves the release this encode targets.
//
// Every path ends at the registry: the caller's version string and the map's own
// version are two ways of NAMING a release and neither is a way of DESCRIBING
// one. That is why Resolve is gone along with WithTargetRelease -- its
// pointer-identity check existed only to stop an assembled Release_t from
// reaching a codec, and a Release_t that carries no codec and cannot be passed to
// the encoder has nothing left to smuggle. An unregistered target stops the
// encode here, before a byte is written.
func (o *encoderOpts) resolveTarget(m *wxx.Map_t) (*Release_t, error) {
	switch o.targetKind {
	case targetByVersion:
		return Lookup(o.targetVersion)
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

	// Inventory what this target cannot express, before anything is written. A
	// modeled loss is reported through diagnostics and the encode proceeds; a
	// loss the encoder cannot honestly describe -- an unmodeled stub -- stops it
	// here, so w gets nothing rather than a file quietly missing content
	// (ADR 0004 Decision 7; the contract is settled in downgradeLoss).
	dropped, err := downgradeLoss(m, target)
	if err != nil {
		return err
	}
	if e.opts.diagnostics != nil {
		e.opts.diagnostics.Dropped = dropped
	}

	// marshal the Map_t to UTF‑8 XML. The target is named by its verbatim
	// application version, the only way to name one: MarshalXML resolves it back
	// to this same entry.
	data, err := MarshalXML(m, target.App.Raw)
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

// MarshalXML converts a Map_t to the XML of the supported release that states
// app as its verbatim map/@version ("1.73", "1.77", "2.06"), without transport
// concerns (no header, no UTF-16, no gzip).
//
// app is an APPLICATION VERSION and nothing else. It is never a schema version
// and never a *Release_t, and no codec comes back: those were the three ways this
// function used to let a caller reach past the dispatcher and choose an encoder,
// which is what issue #41 requirement 5 denies. The registry resolves app to a
// release exactly as Encode does, so the two public encode paths cannot disagree
// about what a target is. "" is not a sentinel here either -- it names no release
// and is the same error as any other unregistered version.
//
// The resolved release's SCHEMA picks the codec (ADR 0004 Decision 4): the schema
// is the format's identity, it is on disk, and it does not change when the
// product is relabelled. The application version does not pick the codec -- it is
// caller-chosen data, and two application versions sharing a schema marshal
// through one codec and differ only in the string written to map/@version.
//
// Because app resolves to the release whose schema then picks the codec, the
// identity in the bytes and the format of the content always come from one entry.
// That is what makes #41's chimera -- W2025 content declaring release=""
// version="1.77" schema="" -- unaskable here: naming the identity IS naming the
// codec. Writing that identity is Release_t.identify's job, and it is what makes
// the target a target rather than a codec hint.
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
	target, err := Lookup(app)
	if err != nil {
		return nil, err
	}
	if _, err := downgradeLoss(m, target); err != nil {
		return nil, err
	}
	c, err := codecForSchema(target.Schema)
	if err != nil {
		return nil, err
	}
	// The codec is handed the application version explicitly and verifies it
	// against the set it declares. identify has already written the same string
	// onto the map's identity, so the two agree by construction here; passing it
	// is what lets the codec state requirement 3 rather than trust its input.
	// The string is target.App.Raw verbatim -- never re-rendered from a Dotted's
	// components, which would send "2.06" to disk as "2.6" (ADR 0004 Decision 1).
	return c.Encode(target.identify(m), target.App.Raw)
}
