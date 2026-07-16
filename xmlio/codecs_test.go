// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
	"github.com/maloquacious/wxx/xmlio/internal/codec"
	"github.com/maloquacious/wxx/xmlio/internal/v0_77"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// codecsForTest mirrors the dispatcher's table (xmlio/codecs.go).
//
// Tests reach the codecs through xmlio/internal/... because xmlio exports no way
// to: a public symbol that hands back an encoder is exactly the reach issue #41
// requirement 5 removes. These tests may import it because Go's internal rule is
// directory-based -- package xmlio_test is an external test package that still
// lives inside xmlio/ -- which is requirement 5's "test units may choose the
// encoder" exception, granted without any escape hatch in the packages
// themselves.
//
// The declarations are read from the codecs rather than restated, so the only
// thing this can drift on is which codecs exist -- which is what it is here to
// state.
func codecsForTest() []codec.Codec {
	return []codec.Codec{
		v0_77.Codec_t{},
		v1_06.Codec_t{},
	}
}

// codecForAppOfTest resolves the codec that writes app, the way the registry does
// internally: by asking every codec what it accepts.
//
// It is a restatement of xmlio's unexported codecFor and cannot be otherwise --
// there is no exported resolver, by design. Disjointness (asserted below) is what
// makes "the codec that writes app" singular, and this fails loudly rather than
// picking one if that ever stops holding.
func codecForAppOfTest(t *testing.T, app string) (codec.Codec, bool) {
	t.Helper()
	var found []codec.Codec
	for _, c := range codecsForTest() {
		if c.AcceptedApps().Accepts(app) {
			found = append(found, c)
		}
	}
	switch len(found) {
	case 0:
		return nil, false
	case 1:
		return found[0], true
	default:
		t.Fatalf("%d codecs accept %q, want at most 1: the sets must be disjoint", len(found), app)
		return nil, false
	}
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

// registrySamples is the registry restated as expectations: the three classic
// builds, which one codec serves, and the W2025 2.06 baseline. These four
// application versions are the whole registry.
//
// wantRelease, wantSchema and wantXMLVersion are the exact bytes the codec that
// writes each application version emits. They are asserted here because they used
// to be REGISTRY data -- Release_t carried all three -- and issue #45 moved them
// onto the codec. The move is an ownership change and not a behavior change, and
// this is what says so.
var registrySamples = []struct {
	name           string
	app            string // the registry key: map/@version verbatim
	wantCodec      codec.Codec
	wantRelease    string // map/@release verbatim; "" means the codec writes none
	wantSchema     string // map/@schema verbatim; "" means the codec writes none
	wantXMLVersion string // the XML declaration its files open with
}{
	{"classic 1.73", "1.73", v0_77.Codec_t{}, "", "", "1.0"},
	{"classic 1.74", "1.74", v0_77.Codec_t{}, "", "", "1.0"},
	{"classic 1.77", "1.77", v0_77.Codec_t{}, "", "", "1.0"},
	{"w2025 2.06", "2.06", v1_06.Codec_t{}, "2025", "1.06", "1.1"},
}

// TestRegistryResolvesEveryApplicationVersion asserts the registry's whole job:
// every supported application version names exactly one codec, and that codec
// writes the identity that goes with it.
//
// This is what TestRegistryLookup used to assert against Release_t, read from its
// new owner. The registry used to carry the release, the schema and the XML
// version and hand them out through Lookup; issue #45 Decision 8 leaves it
// holding nothing but its own key, so the values come from the codec the key
// resolves to.
func TestRegistryResolvesEveryApplicationVersion(t *testing.T) {
	for _, tc := range registrySamples {
		t.Run(tc.name, func(t *testing.T) {
			c, ok := codecForAppOfTest(t, tc.app)
			if !ok {
				t.Fatalf("no codec accepts %q, which is a supported application version", tc.app)
			}
			if c != tc.wantCodec {
				t.Fatalf("%q resolves to codec %T, want %T", tc.app, c, tc.wantCodec)
			}

			decl := c.AcceptedApps()
			a, ok := decl.App(tc.app)
			if !ok {
				t.Fatalf("codec %s accepts %q but has no declaration for it", decl.Codec, tc.app)
			}
			// Verbatim, every one of them: "2.06" must never reach disk as "2.6"
			// (ADR 0004 Decision 1), and the string a file states is the string
			// declared here.
			if got := a.Version; got != tc.app {
				t.Errorf("%q declares Version = %q, want %q verbatim", tc.app, got, tc.app)
			}
			if got := a.Release; got != tc.wantRelease {
				t.Errorf("%q writes release %q, want %q", tc.app, got, tc.wantRelease)
			}
			if got := decl.Schema; got != tc.wantSchema {
				t.Errorf("%q writes schema %q, want %q", tc.app, got, tc.wantSchema)
			}
			if got := decl.XMLVersion; got != tc.wantXMLVersion {
				t.Errorf("%q opens its files with xml version %q, want %q", tc.app, got, tc.wantXMLVersion)
			}
		})
	}
}

// TestRegistryKeysOnRawNotComponents pins the keying ruling: the registry keys on
// the verbatim map/@version string, never on the parsed components.
//
// "2.06" and "2.6" parse to identical components ({2, 6}) and are therefore
// indistinguishable to Dotted.Compare -- but they are different strings, and a
// version string is what a file states. Keying on components would let a request
// for a version no build ever shipped resolve to 2.06 and write a file claiming
// to be something it is not.
//
// It is asserted through MarshalXML because that is the only way left to ask the
// registry a question: there is no exported Lookup any more (issue #46 tracks
// whether a client ever gets one back). The assertion is the same -- the padded
// string resolves, the unpadded one does not.
func TestRegistryKeysOnRawNotComponents(t *testing.T) {
	const (
		registered = "2.06" // a supported application version
		unpadded   = "2.6"  // no supported application version
	)

	// Guard against a vacuous pass. This test proves something only if the two
	// strings really do share components -- if they ever diverged, the miss below
	// would be an ordinary unknown-version miss and would say nothing about which
	// of Raw and the components is the key.
	a, b := mustDotted(t, registered), mustDotted(t, unpadded)
	if a.Raw == b.Raw {
		t.Fatalf("%q and %q are the same string, so Raw-vs-component keying is not under test", registered, unpadded)
	}
	if a.Compare(b) != 0 {
		t.Fatalf("%q and %q have different components (%d.%d vs %d.%d), so component keying would miss anyway and this test proves nothing",
			registered, unpadded, a.Major, a.Minor, b.Major, b.Minor)
	}

	m, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}

	// The padded string is registered.
	if _, err := xmlio.MarshalXML(m, registered); err != nil {
		t.Fatalf("MarshalXML(%q): %v", registered, err)
	}

	// The unpadded one is not, despite identical components.
	data, err := xmlio.MarshalXML(m, unpadded)
	if err == nil {
		t.Fatalf("MarshalXML(%q) returned %d bytes and nil, want a miss: it shares components with %q but is a different string, and the registry keys on the verbatim string",
			unpadded, len(data), registered)
	}
	if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
		t.Errorf("MarshalXML(%q) error = %v, want it to wrap %v", unpadded, err, wxx.ErrUnsupportedMapVersion)
	}
}

// TestRegistryUnknownApplicationVersion asserts an unregistered version misses
// cleanly with a useful error rather than falling back to a best-effort nearest
// release (ADR 0004 Decision 5).
//
// It is the successor to TestRegistryLookupUnknown, asserted through the public
// entry point rather than through Lookup. The cases are unchanged, and three of
// them carry the load: "" is not a sentinel, a SCHEMA version is not an
// application version (issue #41 requirement 1), and neither is a CODEC version --
// "0.77" is on no disk and must resolve to nothing.
func TestRegistryUnknownApplicationVersion(t *testing.T) {
	m, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}

	for _, tc := range []struct {
		name string
		app  string
	}{
		{"empty", ""},
		{"unpadded 2.06", "2.6"},
		{"unreleased classic", "1.75"},
		{"future w2025", "9.99"},
		{"schema not app", "1.06"},
		{"codec version not app", "0.77"},
		{"trailing space", "2.06 "},
		{"malformed", "not-a-version"},
		{"three components", "2.06.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Guard against a vacuous pass: the version must genuinely be
			// unregistered, or this asks the registry to refuse something it should
			// accept.
			if _, ok := codecForAppOfTest(t, tc.app); ok {
				t.Fatalf("%q is accepted by a codec, so refusing it is not the contract under test", tc.app)
			}
			data, err := xmlio.MarshalXML(m, tc.app)
			if err == nil {
				t.Fatalf("MarshalXML(%q) returned %d bytes and nil, want an error", tc.app, len(data))
			}
			if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
				t.Errorf("MarshalXML(%q) error = %v, want it to wrap %v", tc.app, err, wxx.ErrUnsupportedMapVersion)
			}
			// The error must name the version the caller asked for, or it is not
			// actionable.
			if got := err.Error(); tc.app != "" && !strings.Contains(got, tc.app) {
				t.Errorf("MarshalXML(%q) error = %q, want it to name the version asked for", tc.app, got)
			}
			if len(data) != 0 {
				t.Errorf("MarshalXML(%q) returned %d bytes alongside its error, want none", tc.app, len(data))
			}
		})
	}
}

// TestRegistryIsExactlyTheSupportedApplicationVersions asserts the compiled-in
// registry is the four supported application versions and no others, and that no
// version is claimed twice.
//
// There is no second table to check it against any more. The registry IS the
// union of what the codecs declare (issue #45 Decision 8), so "the registry
// agrees with the codecs" -- which TestCodecAppSetsAreDeclaredAndDisjoint used to
// assert against SupportedReleases -- has become a tautology and is gone rather
// than restated. What is still worth asserting is that the union is what we think
// it is: a codec quietly gaining or losing an application version changes which
// files this package will write, and that must not happen unnoticed.
func TestRegistryIsExactlyTheSupportedApplicationVersions(t *testing.T) {
	got := map[string]string{} // app version -> the codec that accepts it
	for _, c := range codecsForTest() {
		decl := c.AcceptedApps()
		for _, a := range decl.Apps {
			if prev, ok := got[a.Version]; ok {
				t.Errorf("version %q is accepted by codec %s and codec %s; the registry cannot key on it", a.Version, prev, decl.Codec)
			}
			got[a.Version] = decl.Codec
		}
	}
	// Guard against a vacuous pass: an empty registry would satisfy every "is
	// missing" assertion below by having nothing to be wrong about.
	if len(got) == 0 {
		t.Fatalf("no codec declares an application version, so the registry is empty and this test asserts nothing")
	}
	if len(got) != len(registrySamples) {
		t.Errorf("the registry holds %d application versions (%s), want %d", len(got), strings.Join(sortedKeys(got), ", "), len(registrySamples))
	}
	for _, tc := range registrySamples {
		if _, ok := got[tc.app]; !ok {
			t.Errorf("the registry is missing version %q", tc.app)
		}
	}
}

// sortedKeys renders a registry map's keys for an assertion message, in a stable
// order so that a failure reads the same twice.
func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// registryFixtureSamples pairs every tracked .wxx fixture with the application
// version its bytes state and the identity the codec that writes it emits.
var registryFixtureSamples = []struct {
	name        string
	path        string
	wantApp     string
	wantCodec   codec.Codec
	wantRelease string
	wantSchema  string // "" means the codec writes no @schema
}{
	{"classic 1.73", "../testdata/blank-2017-1.73-1.0.wxx", "1.73", v0_77.Codec_t{}, "", ""},
	{"classic 1.74", "../testdata/blank-2017-1.74-1.0.wxx", "1.74", v0_77.Codec_t{}, "", ""},
	{"classic 1.77", "../testdata/blank-2017-1.77-1.0.wxx", "1.77", v0_77.Codec_t{}, "", ""},
	{"classic 1.77 columns", "../testdata/2017-1.77-1.0-columns-blank.wxx", "1.77", v0_77.Codec_t{}, "", ""},
	{"classic 1.77 rows", "../testdata/2017-1.77-1.0-rows-blank.wxx", "1.77", v0_77.Codec_t{}, "", ""},
	{"w2025 2.06 blank", sample2025_206, "2.06", v1_06.Codec_t{}, "2025", "1.06"},
	{"w2025 2.06 layers", sample2025_206Layers, "2.06", v1_06.Codec_t{}, "2025", "1.06"},
}

// TestRegistryMatchesFixtures grounds the registry in the files on disk rather
// than in assertions about them: every tracked fixture states an application
// version the registry supports, and the codec that version resolves to declares
// exactly the identity the file itself carries.
//
// The second half is what keeps the declarations honest. App_t.Release and
// Set_t.Schema are what issue #45 moved off the registry and onto the codec, and
// nothing on disk forces them to be right -- they are strings in a Go file. This
// is the check that they match a real Worldographer file, for every fixture we
// have. The file's own @release and @schema are provenance, which the decoders
// record and no encoder reads (issue #45 Decision 9), so they are independent
// evidence rather than a restatement of the declaration.
func TestRegistryMatchesFixtures(t *testing.T) {
	for _, tc := range registryFixtureSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			v := m.MetaData.Version

			// Guard against a vacuous pass: a lookup on an identity the decoder
			// never populated would prove nothing about the fixture.
			if v.App.Raw == "" {
				t.Fatalf("%s: MetaData.Version.App.Raw is empty; nothing to resolve", tc.path)
			}
			if got := v.App.Raw; got != tc.wantApp {
				t.Fatalf("%s: MetaData.Version.App.Raw = %q, want %q", tc.path, got, tc.wantApp)
			}

			// The version the file states must resolve to a codec, and to the
			// expected one. A tracked fixture we cannot write is not supported.
			c, ok := codecForAppOfTest(t, v.App.Raw)
			if !ok {
				t.Fatalf("%s: no codec accepts %q; a tracked fixture must be a supported application version", tc.path, v.App.Raw)
			}
			if c != tc.wantCodec {
				t.Fatalf("%s: %q resolves to codec %T, want %T", tc.path, v.App.Raw, c, tc.wantCodec)
			}

			decl := c.AcceptedApps()
			a, ok := decl.App(v.App.Raw)
			if !ok {
				t.Fatalf("%s: codec %s accepts %q but has no declaration for it", tc.path, decl.Codec, v.App.Raw)
			}

			// What the codec DECLARES it writes must be what the FILE states.
			if got, want := a.Release, m.MetaData.Worldographer.Release; got != want {
				t.Errorf("%s: %q declares release %q, but the file states @release %q", tc.path, v.App.Raw, got, want)
			}
			if got := a.Release; got != tc.wantRelease {
				t.Errorf("%s: %q declares release %q, want %q", tc.path, v.App.Raw, got, tc.wantRelease)
			}
			if got, want := decl.Schema, m.MetaData.Worldographer.Schema; got != want {
				t.Errorf("%s: its codec declares schema %q, but the file states @schema %q", tc.path, got, want)
			}
			if got := decl.Schema; got != tc.wantSchema {
				t.Errorf("%s: its codec declares schema %q, want %q", tc.path, got, tc.wantSchema)
			}

			// The schema axis, cross-checked against the parsed identity: "" is the
			// implicit legacy schema and is an identity, not an unknown (ADR 0004
			// Decision 2).
			if tc.wantSchema == "" {
				if v.Schema != nil {
					t.Errorf("%s: MetaData.Version.Schema = %+v, want nil", tc.path, *v.Schema)
				}
			} else if v.Schema == nil {
				t.Errorf("%s: MetaData.Version.Schema = nil, want %q", tc.path, tc.wantSchema)
			} else if got := v.Schema.Raw; got != tc.wantSchema {
				t.Errorf("%s: MetaData.Version.Schema.Raw = %q, want %q", tc.path, got, tc.wantSchema)
			}
		})
	}
}

// TestCodecDeclarationsAreDisjointOverTheRealCodecs asserts, over the compiled-in
// codecs, the property the registry's key depends on: no application version is
// accepted by two codecs.
//
// It is the MERGED guard. Issue #41 kept two checks apart -- the registry's
// duplicate-application-version check (one version must not name two RELEASES)
// and codec disjointness (one version must not be accepted by two CODECS) --
// because the registry had releases in it to be ambiguous about. The registry is
// now application version -> codec, so the two are the same statement and
// appver.VerifyDisjoint is the survivor.
//
// xmlio's init runs this and panics, which is why this cannot be the only
// coverage: a panic cannot be inspected. TestVerifyDisjointRejectsOverlap in
// appversion_test.go is where the guard is watched to fail.
func TestCodecDeclarationsAreDisjointOverTheRealCodecs(t *testing.T) {
	var sets []appver.Set_t
	for _, c := range codecsForTest() {
		sets = append(sets, c.AcceptedApps())
	}
	// Guard against a vacuous pass: disjointness over one set is trivially true
	// and says nothing.
	if len(sets) < 2 {
		t.Fatalf("%d codec declaration(s) under test, want at least 2: disjointness needs two sets to be a property", len(sets))
	}
	if err := appver.VerifyDisjoint(sets...); err != nil {
		t.Errorf("the compiled-in codec sets are not disjoint: %v", err)
	}
}
