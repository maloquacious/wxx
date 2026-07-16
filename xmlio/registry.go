// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package xmlio implements the XML decoding and encoding pipeline for
// Worldographer .wxx files and hosts the registry of supported releases.
package xmlio

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/v0_77"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// DecodeFunc parses one schema's XML into the Map_t superset.
type DecodeFunc func(input []byte) (*wxx.Map_t, error)

// EncodeFunc emits a Map_t as one schema's XML.
type EncodeFunc func(m *wxx.Map_t) ([]byte, error)

// Codec_t is the parse/emit pair a schema selects (ADR 0004 Decision 4).
//
// It is deliberately keyed off the schema and not off the release: two
// application versions sharing one schema share this pair and differ only in the
// string written to map/@version, which is caller-chosen data the codec cannot
// infer.
type Codec_t struct {
	Decode DecodeFunc
	Encode EncodeFunc
}

// Release_t binds one supported release's full on-disk identity to its codecs
// (ADR 0004 Decision 3).
type Release_t struct {
	// Release is map/@release verbatim ("2025"); "" for classic, which states no
	// such attribute. It is carried because writing a file requires it, not
	// because it means anything: it is a marketing label preserved for fidelity
	// and it never selects a codec.
	Release string

	// App is map/@version, the application build that wrote the file. Raw is
	// authoritative: "2.06" and "2.6" are different files even though their
	// components agree, so this is the registry's key as a verbatim string.
	App wxx.Dotted

	// Schema is map/@schema, the on-disk data format. nil is meaningful and
	// identifies the one implicit legacy (classic) schema rather than an unknown
	// one.
	Schema *wxx.Dotted

	// XMLVersion is the version in the XML declaration this release's files
	// carry: "1.0" for classic, "1.1" for W2025. Like Release it is on-disk
	// identity data the encoder needs in order to write a file, and like Release
	// it never selects a codec.
	//
	// It is bound here as data rather than derived from Schema == nil, which
	// would only work for as long as there are exactly two schemas. NewRegistry
	// rejects an entry naming an XML version no header exists for.
	XMLVersion string

	// Decode and Encode are the codec pair the entry's Schema selects. Entries
	// sharing a schema must name the same pair; NewRegistry enforces it.
	Decode DecodeFunc
	Encode EncodeFunc
}

// XMLHeader returns the XML declaration to write ahead of this release's XML.
//
// The header follows the release. It used to be chosen by a switch on a family
// year (2017 -> 1.0, 2025 -> 1.1) -- the same coinage ADR 0004 deletes from the
// model -- which tied the bytes of every file to a label no classic file states
// and that a relabelled product would change without touching the format.
func (r *Release_t) XMLHeader() ([]byte, error) {
	h, ok := utf16XMLHeader(r.XMLVersion)
	if !ok {
		return nil, errors.Join(wxx.ErrUnknownXMLHeader, fmt.Errorf("version %q: xml version %q: no header", r.App.Raw, r.XMLVersion))
	}
	return []byte(h), nil
}

// Codec returns the parse/emit pair this entry's schema selects.
func (r *Release_t) Codec() Codec_t {
	return Codec_t{Decode: r.Decode, Encode: r.Encode}
}

// identify returns a shallow copy of m stating this release's on-disk identity:
// the map/@release, map/@version and map/@schema strings the codecs write into
// the <map> element.
//
// This is what targeting a release means at the byte level. The schema picks the
// codec, but the codec emits Map_t.Release/Version/Schema -- the strings the
// SOURCE file stated -- so without this a map read from 1.77 and targeted at
// 1.73 would route through the classic codec and still write version="1.77",
// handing back a file claiming a release the caller did not ask for. The target
// would be a codec hint rather than a target, and the licensing guarantee in ADR
// 0004 Decision 5 would be worth nothing.
//
// Encoding a map as the release it already states -- the default target -- writes
// exactly the values it already carried, so no byte moves.
//
// Every string is copied verbatim from Raw and none is re-rendered from a
// Dotted's components (ADR 0004 Decision 1): "2.06" must never go to disk as
// "2.6". The copy is shallow and m is never mutated: the target is a property of
// one encode, not a change to the caller's map.
func (r *Release_t) identify(m *wxx.Map_t) *wxx.Map_t {
	out := *m
	out.Release = r.Release
	out.Version = r.App.Raw
	out.Schema = ""
	if r.Schema != nil {
		out.Schema = r.Schema.Raw
	}
	return &out
}

// Registry_t is the single source of truth for supported releases. It is keyed
// by verbatim application version (see Lookup) and additionally indexes
// schema -> codec (see CodecForSchema).
//
// A Registry_t is read-only once built; NewRegistry validates every invariant up
// front so that a lookup can never resolve ambiguously at encode time.
type Registry_t struct {
	entries  []*Release_t
	byApp    map[string]*Release_t // key: App.Raw, verbatim
	bySchema map[string]Codec_t    // key: schemaKey(Schema)
}

// NewRegistry validates entries and returns a registry over them.
//
// Validation is exhaustive and up front because the alternative is an ambiguity
// that surfaces as a silently wrong codec at encode time. An entry must:
//   - state an application version;
//   - name a non-nil codec pair;
//   - name an XML version some header exists for, since every file written for
//     the release opens with that declaration;
//   - state a schema if and only if it states a release -- classic files carry
//     neither, W2025 files carry both (ADR 0003 Decision 2);
//   - not repeat another entry's verbatim application version, which is the
//     lookup key and must therefore identify exactly one release;
//   - agree with every other entry on the same schema about which codec that
//     schema selects (ADR 0004 Decision 4).
func NewRegistry(entries ...*Release_t) (*Registry_t, error) {
	r := &Registry_t{
		byApp:    make(map[string]*Release_t, len(entries)),
		bySchema: make(map[string]Codec_t, len(entries)),
	}
	for i, e := range entries {
		if e == nil {
			return nil, errors.Join(wxx.ErrInvalidReleaseEntry, fmt.Errorf("entry %d: nil", i))
		}
		if e.App.Raw == "" {
			return nil, errors.Join(wxx.ErrInvalidReleaseEntry, wxx.ErrMissingVersion, fmt.Errorf("entry %d: release %q", i, e.Release))
		}
		if e.Decode == nil || e.Encode == nil {
			return nil, errors.Join(wxx.ErrInvalidReleaseEntry, wxx.ErrMissingCodec, fmt.Errorf("entry %d: version %q", i, e.App.Raw))
		}
		if _, ok := utf16XMLHeader(e.XMLVersion); !ok {
			// Caught here rather than at encode time: an entry that cannot say
			// how its files open cannot write one, and finding that out mid-write
			// is finding it out too late.
			return nil, errors.Join(wxx.ErrInvalidReleaseEntry, wxx.ErrUnknownXMLHeader, fmt.Errorf("entry %d: version %q: xml version %q", i, e.App.Raw, e.XMLVersion))
		}
		if (e.Schema == nil) != (e.Release == "") {
			return nil, errors.Join(wxx.ErrInvalidReleaseEntry, fmt.Errorf("entry %d: version %q: release %q with schema %s: a release states a schema if and only if it states a release", i, e.App.Raw, e.Release, schemaLabel(e.Schema)))
		}
		if e.Schema != nil && e.Schema.Raw == "" {
			// An empty Raw would collide with the key reserved for the implicit
			// legacy schema, and ParseDotted rejects "" anyway.
			return nil, errors.Join(wxx.ErrInvalidReleaseEntry, wxx.ErrMissingVersion, fmt.Errorf("entry %d: version %q: empty schema", i, e.App.Raw))
		}
		if prev, ok := r.byApp[e.App.Raw]; ok {
			return nil, errors.Join(wxx.ErrDuplicateAppVersion, fmt.Errorf("version %q: claimed by release %q and release %q", e.App.Raw, prev.Release, e.Release))
		}
		r.byApp[e.App.Raw] = e

		key := schemaKey(e.Schema)
		if prev, ok := r.bySchema[key]; ok {
			if !sameCodec(prev, e.Codec()) {
				return nil, errors.Join(wxx.ErrAmbiguousSchemaCodec, fmt.Errorf("schema %s: version %q selects a different codec than an earlier entry on the same schema", schemaLabel(e.Schema), e.App.Raw))
			}
		} else {
			r.bySchema[key] = e.Codec()
		}
		r.entries = append(r.entries, e)
	}
	return r, nil
}

// Lookup resolves a supported release by its verbatim application version, e.g.
// Lookup("2.06") or Lookup("1.77").
//
// The key is the exact map/@version string, never the parsed components.
// Dotted{Raw: "2.06"} and Dotted{Raw: "2.6"} carry identical components ({2, 6})
// but name different files, so keying on components would conflate them; a
// malformed version, which the decoders model as a Dotted with zero components,
// likewise matches nothing here, which is the correct outcome.
//
// An unregistered version is an error, never a best-effort nearest match (ADR
// 0004 Decision 5). The returned entry is shared and must not be mutated.
func (r *Registry_t) Lookup(app string) (*Release_t, error) {
	if e, ok := r.byApp[app]; ok {
		return e, nil
	}
	return nil, errors.Join(wxx.ErrUnsupportedMapVersion, fmt.Errorf("version %q: not a supported release", app))
}

// Resolve returns the registry's own entry for e, and is how a *Release_t from a
// caller is admitted as a target.
//
// It rejects anything this registry did not produce, which is what keeps an
// invalid target unrepresentable rather than merely rejected (ADR 0004 Decision
// 5). Release_t is an ordinary exported struct, so a caller can assemble one
// pairing @version="1.77" with the W2025 schema -- a release that has never
// existed. Resolving by App.Raw and then requiring the identical entry back is
// what stops it: the fields cannot disagree with the registry, because the only
// entries that get through are the registry's.
//
// Pointer identity is the test, not field equality. A copy of an entry is
// rejected on purpose: a value the caller owns is a value the caller can mutate
// into an invalid pair after this check has passed, and Lookup and Releases both
// document their entries as shared and not to be mutated.
func (r *Registry_t) Resolve(e *Release_t) (*Release_t, error) {
	if e == nil {
		return nil, errors.Join(wxx.ErrUnsupportedMapVersion, fmt.Errorf("no target release"))
	}
	own, err := r.Lookup(e.App.Raw)
	if err != nil {
		return nil, err
	}
	if own != e {
		return nil, errors.Join(wxx.ErrUnsupportedMapVersion, fmt.Errorf("version %q: not this registry's release entry: target a release with Lookup or SupportedReleases rather than assembling one", e.App.Raw))
	}
	return own, nil
}

// CodecForSchema resolves the parse/emit pair a schema selects (ADR 0004
// Decision 4). A nil schema asks for the implicit legacy (classic) schema.
//
// The schema answers "which code path parses/emits this"; the application
// version does not, and is caller-chosen data. Two application versions sharing
// a schema therefore resolve to one Codec_t here.
func (r *Registry_t) CodecForSchema(schema *wxx.Dotted) (Codec_t, error) {
	if c, ok := r.bySchema[schemaKey(schema)]; ok {
		return c, nil
	}
	return Codec_t{}, errors.Join(wxx.ErrUnsupportedMapSchema, fmt.Errorf("schema %s: not a supported schema", schemaLabel(schema)))
}

// Releases returns the registry's entries in declaration order. The slice is a
// copy; the entries it points at are shared and must not be mutated.
func (r *Registry_t) Releases() []*Release_t {
	out := make([]*Release_t, len(r.entries))
	copy(out, r.entries)
	return out
}

// schemaKey renders a schema as a map key.
//
// nil -- the implicit legacy schema -- keys on "". Nothing on disk can collide
// with it: ParseDotted rejects an empty string and NewRegistry rejects an entry
// whose non-nil schema has an empty Raw.
func schemaKey(d *wxx.Dotted) string {
	if d == nil {
		return ""
	}
	return d.Raw
}

// schemaLabel renders a schema for an error message.
func schemaLabel(d *wxx.Dotted) string {
	if d == nil {
		return "implicit (classic)"
	}
	return fmt.Sprintf("%q", d.Raw)
}

// sameCodec reports whether two codec pairs name the same functions.
//
// Func values are not comparable with ==, so this compares code pointers. Every
// codec in the registry is a package-level function rather than a closure, so a
// code pointer identifies it uniquely.
func sameCodec(a, b Codec_t) bool {
	return reflect.ValueOf(a.Decode).Pointer() == reflect.ValueOf(b.Decode).Pointer() &&
		reflect.ValueOf(a.Encode).Pointer() == reflect.ValueOf(b.Encode).Pointer()
}

// supportedReleases is the registry of releases this package supports. It is
// built and validated by init.
var supportedReleases *Registry_t

// init builds supportedReleases, panicking if the compiled-in table is invalid.
//
// The table is a constant of the program, so a duplicate application version or
// an ambiguous schema in it is a programming error, not a runtime condition a
// caller could handle -- and the failure it would otherwise produce is a
// silently wrong codec at encode time. Failing at load makes it unmissable. The
// validation itself lives in NewRegistry rather than here so that it stays
// testable: a test can hand NewRegistry a deliberately broken table and inspect
// the error, which it could not do with a panic in init.
func init() {
	entries, err := supportedReleaseEntries()
	if err != nil {
		panic(fmt.Sprintf("xmlio: supported release table: %v", err))
	}
	r, err := NewRegistry(entries...)
	if err != nil {
		panic(fmt.Sprintf("xmlio: supported release registry: %v", err))
	}
	supportedReleases = r
}

// supportedReleaseEntries returns the supported releases, and only these:
//
//   - classic 1.73, 1.74 and 1.77 on the implicit legacy schema, which share an
//     identical element vocabulary and therefore one codec;
//   - the W2025 2.06 baseline (release "2025", schema "1.06"), the first
//     post-beta 2025 build. Earlier 2025 builds are out of scope (ADR 0003
//     Decision 3), so no entry exists for one.
//
// Adding a release is an entry here rather than a new switch arm.
func supportedReleaseEntries() ([]*Release_t, error) {
	var entries []*Release_t

	// Classic: no release attribute, no schema attribute.
	for _, app := range []string{"1.73", "1.74", "1.77"} {
		a, err := wxx.ParseDotted(app)
		if err != nil {
			return nil, errors.Join(wxx.ErrInvalidReleaseEntry, fmt.Errorf("classic version %q", app), err)
		}
		entries = append(entries, &Release_t{
			Release:    "",
			App:        a,
			Schema:     nil,
			XMLVersion: "1.0",
			Decode:     v0_77.Decode,
			Encode:     v0_77.Encode,
		})
	}

	// W2025 baseline. The v1_06 decoder is a work in progress; the entry binds
	// the codec that exists today.
	app206, err := wxx.ParseDotted("2.06")
	if err != nil {
		return nil, errors.Join(wxx.ErrInvalidReleaseEntry, fmt.Errorf("w2025 version %q", "2.06"), err)
	}
	schema106, err := wxx.ParseDotted("1.06")
	if err != nil {
		return nil, errors.Join(wxx.ErrInvalidReleaseEntry, fmt.Errorf("w2025 schema %q", "1.06"), err)
	}
	entries = append(entries, &Release_t{
		Release:    "2025",
		App:        app206,
		Schema:     &schema106,
		XMLVersion: "1.1",
		Decode:     v1_06.Decode,
		Encode:     v1_06.Encode,
	})

	return entries, nil
}

// Lookup resolves a supported release by its verbatim application version. See
// Registry_t.Lookup.
func Lookup(app string) (*Release_t, error) {
	return supportedReleases.Lookup(app)
}

// Resolve admits a caller's *Release_t as a supported release. See
// Registry_t.Resolve.
func Resolve(e *Release_t) (*Release_t, error) {
	return supportedReleases.Resolve(e)
}

// CodecForSchema resolves the codec a schema selects. See
// Registry_t.CodecForSchema.
func CodecForSchema(schema *wxx.Dotted) (Codec_t, error) {
	return supportedReleases.CodecForSchema(schema)
}

// SupportedReleases returns the supported releases in declaration order. See
// Registry_t.Releases.
func SupportedReleases() []*Release_t {
	return supportedReleases.Releases()
}
