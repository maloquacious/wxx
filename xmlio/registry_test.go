// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/h2017v1"
	"github.com/maloquacious/wxx/xmlio/h2025v1"
)

// funcPtr returns a function value's code pointer, which is how a test asserts
// "this is that codec": func values are not comparable with ==. Every codec in
// the registry is a package-level function, so the pointer identifies it.
func funcPtr(f any) uintptr {
	return reflect.ValueOf(f).Pointer()
}

// mustDotted parses an on-disk dotted version for a test table.
func mustDotted(t *testing.T, s string) wxx.Dotted {
	t.Helper()
	d, err := wxx.ParseDotted(s)
	if err != nil {
		t.Fatalf("ParseDotted(%q): %v", s, err)
	}
	return d
}

// dottedPtr is mustDotted for a schema slot, where nil means the implicit
// legacy schema and so a pointer is required.
func dottedPtr(t *testing.T, s string) *wxx.Dotted {
	t.Helper()
	d := mustDotted(t, s)
	return &d
}

// registrySamples is the supported-release table restated as expectations: the
// three classic builds on the implicit legacy schema, and the W2025 2.06
// baseline. These four are the whole registry (ADR 0004 Decision 3).
//
// wantSchema is the exact map/@schema bytes; "" means the release states none
// and must resolve to a nil Schema. wantXMLVersion is the version in the XML
// declaration the release's files open with.
var registrySamples = []struct {
	name           string
	app            string // the lookup key: map/@version verbatim
	wantRelease    string // map/@release verbatim
	wantSchema     string // map/@schema verbatim; "" means the release states none
	wantXMLVersion string // the release's XML declaration version
	wantDecode     any
	wantEncode     any
}{
	{"classic 1.73", "1.73", "", "", "1.0", h2017v1.Decode, h2017v1.Encode},
	{"classic 1.74", "1.74", "", "", "1.0", h2017v1.Decode, h2017v1.Encode},
	{"classic 1.77", "1.77", "", "", "1.0", h2017v1.Decode, h2017v1.Encode},
	{"w2025 2.06", "2.06", "2025", "1.06", "1.1", h2025v1.Decode, h2025v1.Encode},
}

// TestRegistryLookup asserts every supported release resolves to its full
// on-disk identity and its codec pair.
func TestRegistryLookup(t *testing.T) {
	for _, tc := range registrySamples {
		t.Run(tc.name, func(t *testing.T) {
			e, err := xmlio.Lookup(tc.app)
			if err != nil {
				t.Fatalf("Lookup(%q): %v", tc.app, err)
			}
			if got := e.App.Raw; got != tc.app {
				t.Errorf("Lookup(%q).App.Raw = %q, want %q verbatim", tc.app, got, tc.app)
			}
			if got := e.Release; got != tc.wantRelease {
				t.Errorf("Lookup(%q).Release = %q, want %q", tc.app, got, tc.wantRelease)
			}
			if tc.wantSchema == "" {
				if e.Schema != nil {
					t.Errorf("Lookup(%q).Schema = %+v, want nil (the release states no @schema)", tc.app, *e.Schema)
				}
			} else if e.Schema == nil {
				t.Errorf("Lookup(%q).Schema = nil, want %q", tc.app, tc.wantSchema)
			} else if got := e.Schema.Raw; got != tc.wantSchema {
				t.Errorf("Lookup(%q).Schema.Raw = %q, want %q verbatim", tc.app, got, tc.wantSchema)
			}
			if got := e.XMLVersion; got != tc.wantXMLVersion {
				t.Errorf("Lookup(%q).XMLVersion = %q, want %q", tc.app, got, tc.wantXMLVersion)
			}
			// The entry must be able to hand back the declaration its files open
			// with; the encoder writes exactly these bytes ahead of the XML.
			h, err := e.XMLHeader()
			if err != nil {
				t.Fatalf("Lookup(%q).XMLHeader(): %v", tc.app, err)
			}
			if want := "<?xml version='" + tc.wantXMLVersion + "' encoding='utf-16'?>\n"; string(h) != want {
				t.Errorf("Lookup(%q).XMLHeader() = %q, want %q", tc.app, h, want)
			}
			if got, want := funcPtr(e.Decode), funcPtr(tc.wantDecode); got != want {
				t.Errorf("Lookup(%q).Decode is not the expected codec", tc.app)
			}
			if got, want := funcPtr(e.Encode), funcPtr(tc.wantEncode); got != want {
				t.Errorf("Lookup(%q).Encode is not the expected codec", tc.app)
			}
		})
	}
}

// TestRegistryLookupKeysOnRawNotComponents pins the keying ruling: the registry
// keys on the verbatim map/@version string, never on the parsed components.
//
// "2.06" and "2.6" parse to identical components ({2, 6}) and are therefore
// indistinguishable to Dotted.Compare -- but they are different strings, and a
// version string is what a file states. Keying on components would let a lookup
// for a version that no supported release states resolve to 2.06 and write a
// file claiming to be something it is not.
func TestRegistryLookupKeysOnRawNotComponents(t *testing.T) {
	const (
		registered = "2.06" // a supported release states this
		unpadded   = "2.6"  // no supported release states this
	)

	// Guard against a vacuous pass. This test proves something only if the two
	// strings really do share components -- if they ever diverged, the miss
	// below would be an ordinary unknown-version miss and would say nothing
	// about which of Raw and the components is the key.
	a, b := mustDotted(t, registered), mustDotted(t, unpadded)
	if a.Raw == b.Raw {
		t.Fatalf("%q and %q are the same string, so Raw-vs-component keying is not under test", registered, unpadded)
	}
	if a.Compare(b) != 0 {
		t.Fatalf("%q and %q have different components (%d.%d vs %d.%d), so component keying would miss anyway and this test proves nothing",
			registered, unpadded, a.Major, a.Minor, b.Major, b.Minor)
	}

	// The padded string is registered.
	if _, err := xmlio.Lookup(registered); err != nil {
		t.Fatalf("Lookup(%q): %v", registered, err)
	}

	// The unpadded one is not, despite identical components.
	e, err := xmlio.Lookup(unpadded)
	if err == nil {
		t.Fatalf("Lookup(%q) resolved to release %q version %q, want a miss: it shares components with %q but is a different string, and the registry keys on the verbatim string",
			unpadded, e.Release, e.App.Raw, registered)
	}
	if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
		t.Errorf("Lookup(%q) error = %v, want it to wrap %v", unpadded, err, wxx.ErrUnsupportedMapVersion)
	}
}

// TestRegistryLookupUnknown asserts an unregistered version misses cleanly with
// a useful error rather than falling back to a best-effort nearest release (ADR
// 0004 Decision 5).
func TestRegistryLookupUnknown(t *testing.T) {
	for _, tc := range []struct {
		name string
		app  string
	}{
		{"empty", ""},
		{"unpadded 2.06", "2.6"},
		{"unreleased classic", "1.75"},
		{"future w2025", "9.99"},
		{"schema not app", "1.06"}, // the W2025 schema is not an application version
		{"trailing space", "2.06 "},
		{"malformed", "not-a-version"},
		{"three components", "2.06.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := xmlio.Lookup(tc.app)
			if err == nil {
				t.Fatalf("Lookup(%q) resolved to release %q version %q, want an error", tc.app, e.Release, e.App.Raw)
			}
			if e != nil {
				t.Errorf("Lookup(%q) returned entry %+v alongside an error, want nil", tc.app, e)
			}
			if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
				t.Errorf("Lookup(%q) error = %v, want it to wrap %v", tc.app, err, wxx.ErrUnsupportedMapVersion)
			}
			// The error must name the version the caller asked for, or it is not
			// actionable.
			if got := err.Error(); tc.app != "" && !strings.Contains(got, tc.app) {
				t.Errorf("Lookup(%q) error = %q, want it to name the version asked for", tc.app, got)
			}
		})
	}
}

// TestRegistrySelfConsistency asserts the compiled-in table's invariants: it is
// exactly the four supported releases, each states a schema if and only if it
// states a release, each names a codec pair, and no application version is
// claimed twice.
func TestRegistrySelfConsistency(t *testing.T) {
	got := xmlio.SupportedReleases()
	if len(got) != len(registrySamples) {
		t.Fatalf("SupportedReleases() has %d entries, want %d: the registry is exactly the supported releases", len(got), len(registrySamples))
	}

	seen := map[string]string{} // App.Raw -> Release
	for i, e := range got {
		if e.App.Raw == "" {
			t.Errorf("entry %d: App.Raw is empty; every supported release states a version", i)
			continue
		}
		// Classic states neither @release nor @schema; W2025 states both. The
		// absence of a schema is what identifies the implicit legacy schema, so
		// the two must agree.
		if (e.Schema == nil) != (e.Release == "") {
			t.Errorf("entry %d (version %q): Release = %q but Schema == nil is %v; a release states a schema if and only if it states a release",
				i, e.App.Raw, e.Release, e.Schema == nil)
		}
		if e.Schema != nil && e.Schema.Raw == "" {
			t.Errorf("entry %d (version %q): Schema.Raw is empty", i, e.App.Raw)
		}
		if e.Decode == nil || e.Encode == nil {
			t.Errorf("entry %d (version %q): nil codec (Decode nil: %v, Encode nil: %v)", i, e.App.Raw, e.Decode == nil, e.Encode == nil)
		}
		// Every release must know the declaration its files open with, or it
		// cannot write one.
		if _, err := e.XMLHeader(); err != nil {
			t.Errorf("entry %d (version %q): XMLHeader(): %v", i, e.App.Raw, err)
		}
		// App.Raw must be the string it claims: the components are parsed from
		// it and never rendered back, so they must agree with it.
		if d, err := wxx.ParseDotted(e.App.Raw); err != nil {
			t.Errorf("entry %d: App.Raw = %q does not parse: %v", i, e.App.Raw, err)
		} else if d.Major != e.App.Major || d.Minor != e.App.Minor {
			t.Errorf("entry %d: App = %q parsed as {%d, %d}, want {%d, %d}", i, e.App.Raw, e.App.Major, e.App.Minor, d.Major, d.Minor)
		}
		if prev, ok := seen[e.App.Raw]; ok {
			t.Errorf("entry %d: version %q claimed twice (release %q and release %q)", i, e.App.Raw, prev, e.Release)
		}
		seen[e.App.Raw] = e.Release
	}

	for _, tc := range registrySamples {
		if _, ok := seen[tc.app]; !ok {
			t.Errorf("SupportedReleases() is missing version %q", tc.app)
		}
	}
}

// classicEntry builds a valid classic entry for a constructor test.
func classicEntry(t *testing.T, app string) *xmlio.Release_t {
	t.Helper()
	return &xmlio.Release_t{
		Release:    "",
		App:        mustDotted(t, app),
		Schema:     nil,
		XMLVersion: "1.0",
		Decode:     h2017v1.Decode,
		Encode:     h2017v1.Encode,
	}
}

// w2025Entry builds a valid W2025 entry for a constructor test.
func w2025Entry(t *testing.T, app, schema string) *xmlio.Release_t {
	t.Helper()
	return &xmlio.Release_t{
		Release:    "2025",
		App:        mustDotted(t, app),
		Schema:     dottedPtr(t, schema),
		XMLVersion: "1.1",
		Decode:     h2025v1.Decode,
		Encode:     h2025v1.Encode,
	}
}

// TestNewRegistryRejectsDuplicateAppVersion asserts the duplicate guard. The
// application version is the lookup key, so a table claiming one version twice
// is ambiguous; NewRegistry must reject it at construction rather than let a
// lookup silently pick whichever entry landed in the map last.
func TestNewRegistryRejectsDuplicateAppVersion(t *testing.T) {
	// Control: without the duplicate this table is valid. Guard against a
	// vacuous pass -- if the table were rejected for some other reason, the
	// assertion below would pass while proving nothing about duplicates.
	ok := []*xmlio.Release_t{
		classicEntry(t, "1.73"),
		classicEntry(t, "1.77"),
		w2025Entry(t, "2.06", "1.06"),
	}
	if _, err := xmlio.NewRegistry(ok...); err != nil {
		t.Fatalf("NewRegistry(valid table): %v; the duplicate case below would prove nothing", err)
	}

	for _, tc := range []struct {
		name    string
		entries []*xmlio.Release_t
	}{
		{
			// The same release listed twice.
			name:    "identical entries",
			entries: []*xmlio.Release_t{classicEntry(t, "1.73"), classicEntry(t, "1.77"), classicEntry(t, "1.77")},
		},
		{
			// The nastier case: one version claimed by two different releases,
			// so a lookup for "2.06" cannot say which file to write.
			name:    "same version, different releases",
			entries: []*xmlio.Release_t{w2025Entry(t, "2.06", "1.06"), w2025Entry(t, "2.06", "1.06")},
		},
		{
			// A duplicate hiding behind a different schema.
			name:    "same version across schemas",
			entries: []*xmlio.Release_t{classicEntry(t, "1.77"), w2025Entry(t, "1.77", "1.06")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := xmlio.NewRegistry(tc.entries...)
			if err == nil {
				t.Fatalf("NewRegistry(%s) succeeded with %d entries, want a duplicate-version error: an ambiguous key must fail at construction, never resolve arbitrarily at encode time",
					tc.name, len(r.Releases()))
			}
			if !errors.Is(err, wxx.ErrDuplicateAppVersion) {
				t.Errorf("NewRegistry(%s) error = %v, want it to wrap %v", tc.name, err, wxx.ErrDuplicateAppVersion)
			}
			if r != nil {
				t.Errorf("NewRegistry(%s) returned a registry alongside an error, want nil", tc.name)
			}
		})
	}
}

// TestNewRegistryRejectsInvalidEntry asserts the remaining construction-time
// invariants.
func TestNewRegistryRejectsInvalidEntry(t *testing.T) {
	for _, tc := range []struct {
		name  string
		entry *xmlio.Release_t
	}{
		{"nil entry", nil},
		{
			"no application version",
			&xmlio.Release_t{Release: "", App: wxx.Dotted{}, Schema: nil, XMLVersion: "1.0", Decode: h2017v1.Decode, Encode: h2017v1.Encode},
		},
		{
			"nil decoder",
			&xmlio.Release_t{Release: "", App: mustDotted(t, "1.77"), Schema: nil, XMLVersion: "1.0", Decode: nil, Encode: h2017v1.Encode},
		},
		{
			"nil encoder",
			&xmlio.Release_t{Release: "", App: mustDotted(t, "1.77"), Schema: nil, XMLVersion: "1.0", Decode: h2017v1.Decode, Encode: nil},
		},
		{
			// A release with no schema: W2025 states both or the absence stops
			// identifying the implicit legacy schema.
			"release without schema",
			&xmlio.Release_t{Release: "2025", App: mustDotted(t, "2.06"), Schema: nil, XMLVersion: "1.1", Decode: h2025v1.Decode, Encode: h2025v1.Encode},
		},
		{
			// A schema with no release: classic states neither.
			"schema without release",
			&xmlio.Release_t{Release: "", App: mustDotted(t, "1.77"), Schema: dottedPtr(t, "1.06"), XMLVersion: "1.0", Decode: h2017v1.Decode, Encode: h2017v1.Encode},
		},
		{
			// An empty Raw would collide with the implicit legacy schema's key.
			"empty schema",
			&xmlio.Release_t{Release: "2025", App: mustDotted(t, "2.06"), Schema: &wxx.Dotted{}, XMLVersion: "1.1", Decode: h2025v1.Decode, Encode: h2025v1.Encode},
		},
		{
			// An entry that does not say how its files open cannot write one, and
			// encode time is too late to find that out.
			"no xml version",
			&xmlio.Release_t{Release: "", App: mustDotted(t, "1.77"), Schema: nil, XMLVersion: "", Decode: h2017v1.Decode, Encode: h2017v1.Encode},
		},
		{
			"unknown xml version",
			&xmlio.Release_t{Release: "", App: mustDotted(t, "1.77"), Schema: nil, XMLVersion: "1.2", Decode: h2017v1.Decode, Encode: h2017v1.Encode},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := xmlio.NewRegistry(tc.entry)
			if err == nil {
				t.Fatalf("NewRegistry(%s) succeeded, want an invalid-entry error", tc.name)
			}
			if !errors.Is(err, wxx.ErrInvalidReleaseEntry) {
				t.Errorf("NewRegistry(%s) error = %v, want it to wrap %v", tc.name, err, wxx.ErrInvalidReleaseEntry)
			}
			if r != nil {
				t.Errorf("NewRegistry(%s) returned a registry alongside an error, want nil", tc.name)
			}
		})
	}
}

// TestRegistrySchemaSelectsCodec asserts ADR 0004 Decision 4 against the
// compiled-in registry: the schema answers "which code path parses/emits this",
// and a nil schema asks for the implicit legacy one.
func TestRegistrySchemaSelectsCodec(t *testing.T) {
	t.Run("implicit legacy schema", func(t *testing.T) {
		c, err := xmlio.CodecForSchema(nil)
		if err != nil {
			t.Fatalf("CodecForSchema(nil): %v", err)
		}
		if funcPtr(c.Decode) != funcPtr(h2017v1.Decode) || funcPtr(c.Encode) != funcPtr(h2017v1.Encode) {
			t.Errorf("CodecForSchema(nil) is not the h2017v1 codec")
		}
	})

	t.Run("w2025 schema 1.06", func(t *testing.T) {
		c, err := xmlio.CodecForSchema(dottedPtr(t, "1.06"))
		if err != nil {
			t.Fatalf(`CodecForSchema("1.06"): %v`, err)
		}
		if funcPtr(c.Decode) != funcPtr(h2025v1.Decode) || funcPtr(c.Encode) != funcPtr(h2025v1.Encode) {
			t.Errorf(`CodecForSchema("1.06") is not the h2025v1 codec`)
		}
	})

	// Schemas key on Raw too, for the same reason application versions do.
	for _, tc := range []struct {
		name   string
		schema string
	}{
		{"unpadded 1.06", "1.6"},
		{"unknown schema", "9.99"},
		{"app not schema", "2.06"},
	} {
		t.Run("miss: "+tc.name, func(t *testing.T) {
			d := wxx.Dotted{Raw: tc.schema}
			if parsed, err := wxx.ParseDotted(tc.schema); err == nil {
				d = parsed
			}
			c, err := xmlio.CodecForSchema(&d)
			if err == nil {
				t.Fatalf("CodecForSchema(%q) resolved to a codec, want a miss", tc.schema)
			}
			if !errors.Is(err, wxx.ErrUnsupportedMapSchema) {
				t.Errorf("CodecForSchema(%q) error = %v, want it to wrap %v", tc.schema, err, wxx.ErrUnsupportedMapSchema)
			}
			if c.Decode != nil || c.Encode != nil {
				t.Errorf("CodecForSchema(%q) returned a codec alongside an error, want the zero Codec_t", tc.schema)
			}
		})
	}
}

// TestNewRegistrySharedSchema asserts that "two application versions share one
// schema" is representable and behaves as ADR 0004 Decision 4 requires: one
// codec, and the entries differ only in the string written to map/@version.
//
// The supported registry has no such collision today, so this is a table local
// to the test rather than a fabricated entry in the real one. It is here because
// the shape must hold when a release like that does appear.
func TestNewRegistrySharedSchema(t *testing.T) {
	// Two hypothetical application versions on one schema.
	first := w2025Entry(t, "2.06", "1.06")
	second := w2025Entry(t, "3.01", "1.06")

	r, err := xmlio.NewRegistry(first, second)
	if err != nil {
		t.Fatalf("NewRegistry(two versions on one schema): %v; sharing a schema must be representable", err)
	}

	// Both resolve, and to different application versions.
	a, err := r.Lookup("2.06")
	if err != nil {
		t.Fatalf(`Lookup("2.06"): %v`, err)
	}
	b, err := r.Lookup("3.01")
	if err != nil {
		t.Fatalf(`Lookup("3.01"): %v`, err)
	}
	if a.App.Raw == b.App.Raw {
		t.Fatalf("both versions resolved to App.Raw = %q, want the versions to differ", a.App.Raw)
	}

	// ...and to the same codec, because the schema -- not the version -- selects it.
	if funcPtr(a.Decode) != funcPtr(b.Decode) || funcPtr(a.Encode) != funcPtr(b.Encode) {
		t.Errorf("two versions on schema 1.06 resolved to different codecs; the schema selects the codec")
	}
	c, err := r.CodecForSchema(dottedPtr(t, "1.06"))
	if err != nil {
		t.Fatalf(`CodecForSchema("1.06"): %v`, err)
	}
	if funcPtr(c.Decode) != funcPtr(a.Decode) || funcPtr(c.Encode) != funcPtr(a.Encode) {
		t.Errorf("CodecForSchema does not agree with the entries on the same schema")
	}
}

// TestNewRegistryRejectsAmbiguousSchemaCodec is the converse of
// TestNewRegistrySharedSchema: entries may share a schema, but only if they
// agree on the codec it selects. Disagreement means the schema no longer answers
// "which code path emits this", so it must fail at construction.
func TestNewRegistryRejectsAmbiguousSchemaCodec(t *testing.T) {
	first := w2025Entry(t, "2.06", "1.06")
	second := w2025Entry(t, "3.01", "1.06")
	// Same schema, a different codec.
	second.Decode, second.Encode = h2017v1.Decode, h2017v1.Encode

	r, err := xmlio.NewRegistry(first, second)
	if err == nil {
		t.Fatalf("NewRegistry(one schema, two codecs) succeeded with %d entries, want an ambiguity error", len(r.Releases()))
	}
	if !errors.Is(err, wxx.ErrAmbiguousSchemaCodec) {
		t.Errorf("NewRegistry(one schema, two codecs) error = %v, want it to wrap %v", err, wxx.ErrAmbiguousSchemaCodec)
	}
}

// registryFixtureSamples pairs every tracked .wxx fixture with the registry
// entry its bytes must resolve to.
var registryFixtureSamples = []struct {
	name        string
	path        string
	wantApp     string
	wantRelease string
	wantSchema  string // "" means the file states no @schema
}{
	{"classic 1.73", "../testdata/blank-2017-1.73-1.0.wxx", "1.73", "", ""},
	{"classic 1.74", "../testdata/blank-2017-1.74-1.0.wxx", "1.74", "", ""},
	{"classic 1.77", "../testdata/blank-2017-1.77-1.0.wxx", "1.77", "", ""},
	{"classic 1.77 columns", "../testdata/2017-1.77-1.0-columns-blank.wxx", "1.77", "", ""},
	{"classic 1.77 rows", "../testdata/2017-1.77-1.0-rows-blank.wxx", "1.77", "", ""},
	{"w2025 2.06 blank", sample2025_206, "2.06", "2025", "1.06"},
	{"w2025 2.06 layers", sample2025_206Layers, "2.06", "2025", "1.06"},
}

// TestRegistryMatchesFixtures grounds the registry in the files on disk rather
// than in assertions about them: decode each tracked fixture and confirm the
// version identity its bytes state resolves to the expected entry, and that the
// entry's identity matches what the file actually says.
func TestRegistryMatchesFixtures(t *testing.T) {
	for _, tc := range registryFixtureSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			v := m.MetaData.Version

			// Guard against a vacuous pass: a registry lookup on an identity the
			// decoder never populated would prove nothing about the fixture.
			if v.App.Raw == "" {
				t.Fatalf("%s: MetaData.Version.App.Raw is empty; nothing to resolve", tc.path)
			}
			if got := v.App.Raw; got != tc.wantApp {
				t.Fatalf("%s: MetaData.Version.App.Raw = %q, want %q", tc.path, got, tc.wantApp)
			}

			// The identity the file states must resolve to a supported release.
			e, err := xmlio.Lookup(v.App.Raw)
			if err != nil {
				t.Fatalf("%s: Lookup(%q): %v; a tracked fixture must be a supported release", tc.path, v.App.Raw, err)
			}
			if got := e.Release; got != tc.wantRelease {
				t.Errorf("%s: entry Release = %q, want %q", tc.path, got, tc.wantRelease)
			}
			// The entry's release must be the one the file states.
			if got, want := e.Release, m.MetaData.Worldographer.Release; got != want {
				t.Errorf("%s: entry Release = %q, want it to match the file's @release %q", tc.path, got, want)
			}

			// The entry's schema must be the one the file states, absence included.
			if tc.wantSchema == "" {
				if v.Schema != nil {
					t.Fatalf("%s: MetaData.Version.Schema = %+v, want nil", tc.path, *v.Schema)
				}
				if e.Schema != nil {
					t.Errorf("%s: entry Schema = %+v, want nil to match the file", tc.path, *e.Schema)
				}
			} else {
				if v.Schema == nil {
					t.Fatalf("%s: MetaData.Version.Schema = nil, want %q", tc.path, tc.wantSchema)
				}
				if got := v.Schema.Raw; got != tc.wantSchema {
					t.Fatalf("%s: MetaData.Version.Schema.Raw = %q, want %q", tc.path, got, tc.wantSchema)
				}
				if e.Schema == nil {
					t.Fatalf("%s: entry Schema = nil, want %q to match the file", tc.path, tc.wantSchema)
				}
				if got := e.Schema.Raw; got != v.Schema.Raw {
					t.Errorf("%s: entry Schema.Raw = %q, want it to match the file's @schema %q", tc.path, got, v.Schema.Raw)
				}
			}

			// The codec the file's schema selects must be the entry's codec.
			c, err := xmlio.CodecForSchema(v.Schema)
			if err != nil {
				t.Fatalf("%s: CodecForSchema(%v): %v", tc.path, v.Schema, err)
			}
			if funcPtr(c.Decode) != funcPtr(e.Decode) || funcPtr(c.Encode) != funcPtr(e.Encode) {
				t.Errorf("%s: the codec its schema selects is not the codec its release entry names", tc.path)
			}
		})
	}
}
