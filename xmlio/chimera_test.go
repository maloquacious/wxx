// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// The chimera, as issue #41 demonstrated it: a classic 1.77 map emitted through
// the W2025 codec, producing W2025 XML that declares the classic identity.
//
//	<map type="WORLD" release="" version="1.77" schema="" ...>
//
// The harm is not that the file is wrong. It is that NOTHING DETECTS IT.
// release="" plus version="1.77" routes to the classic decoder, which tolerates
// the W2025 elements it does not recognize, so the round trip reports success and
// wxx cannot tell the file is a chimera.
const (
	chimeraRelease = `release=""`
	chimeraVersion = `version="1.77"`
	chimeraSchema  = `schema=""`
)

// mapElement matches the <map> start tag, which is where a file's whole identity
// -- @release, @version, @schema -- lives.
var mapElement = regexp.MustCompile(`(?s)<map[^>]*>`)

// decodeRawXML runs marshaled XML back through the PUBLIC decoder.
//
// MarshalXML returns XML without transport (no header, no UTF-16, no gzip), so
// the two transport stages are switched off and the declaration is supplied here.
// The utf-8 declarations exist in the decoder's header table for exactly this --
// reading XML a tool has already transcoded. xmlVersion is transport rather than
// identity: what routes a file to a decoder is map/@release and map/@version, and
// that is the routing this test is about.
func decodeRawXML(t *testing.T, xml []byte, xmlVersion string) (*wxx.Map_t, error) {
	t.Helper()
	header := "<?xml version='" + xmlVersion + "' encoding='utf-8'?>\n"
	buf := append([]byte(header), xml...)
	return xmlio.NewDecoder(xmlio.WithSkipUncompress(), xmlio.WithUTF16BEInput(false)).Decode(bytes.NewReader(buf))
}

// TestChimeraIsUnreachableThroughThePublicAPI is issue #41's acceptance test:
// the chimera above cannot be produced through the public API.
//
// It proves that in two halves, because either alone is weak.
//
// The COMPILE-LEVEL half is TestPublicEncodePathsNameAnApplicationVersion below,
// plus what is absent: xmlio.CodecForSchema, xmlio.Codec_t, xmlio.Resolve,
// xmlio.WithTargetRelease, and Release_t's Decode/Encode fields and Codec()
// method are gone, while MarshalXML and WithTargetVersion take an application
// version STRING. A caller cannot name a schema or hold a codec, so the chimera
// cannot be expressed at all.
//
// The BEHAVIORAL half is here. A compile-level proof only shows the recipe cannot
// be written; it says nothing about whether the ingredients still exist. So this
// test builds the chimera FOR REAL through xmlio/internal/v1_06 -- which it may
// do, and only a test may: requirement 5's exception for test units, which works
// because Go's internal rule is directory-based and this external test package
// sits inside xmlio/. Building it proves the hazard is live rather than
// theoretical, and is the guard against a vacuous pass: if the chimera ever stops
// being constructible, "the public API does not produce it" would pass for the
// wrong reason, and the guard fatals instead.
//
// Note what does NOT stop the chimera: the codec's application-version gate. It
// passes here, because "2.06" is a version v1_06 accepts -- the gate checks the
// ARGUMENT, not the map. What stops it on every public path is Release_t.identify
// writing the target's identity onto the map before the codec sees it, so the
// bytes state the release whose schema selected the codec that wrote them.
func TestChimeraIsUnreachableThroughThePublicAPI(t *testing.T) {
	classic, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}

	// Guard against a vacuous pass: the source must really be classic. A W2025
	// source could not produce a classic-identity chimera, and everything below
	// would pass while testing nothing.
	if got := classic.MetaData.Version.App.Raw; got != "1.77" {
		t.Fatalf("%s states version %q, want %q: the chimera is a classic map through the W2025 codec", classicFixture, got, "1.77")
	}
	if classic.MetaData.Version.Schema != nil {
		t.Fatalf("%s states schema %+v, want nil: a classic source states no @schema", classicFixture, *classic.MetaData.Version.Schema)
	}

	// ---- The hazard is real: build the chimera through the internal codec ----
	//
	// v1_06.Encode with the accepted version "2.06", handed a map still carrying
	// its classic identity, and WITHOUT identify having run. This is #41's recipe
	// reached the only way left to reach it.
	chimera, err := v1_06.Encode(classic, "2.06")
	if err != nil {
		t.Fatalf("v1_06.Encode(classic map, %q): %v; the chimera must be constructible here or the public-API assertions below prove nothing", "2.06", err)
	}
	chimeraMap := mapElement.Find(chimera)
	if chimeraMap == nil {
		t.Fatalf("v1_06.Encode(classic map) emitted no <map> element")
	}
	// Guard against a vacuous pass: it must actually BE the chimera -- W2025
	// content wearing the classic identity.
	for _, want := range []string{chimeraRelease, chimeraVersion, chimeraSchema} {
		if !bytes.Contains(chimeraMap, []byte(want)) {
			t.Fatalf("the chimera's <map> element does not state %s, so it is not the chimera and this test proves nothing:\n%s", want, chimeraMap)
		}
	}
	// ...and the harm: it re-decodes SILENTLY as classic. Nothing reports that the
	// content is W2025, because @release and @version say it is not.
	back, err := decodeRawXML(t, chimera, "1.0")
	if err != nil {
		t.Fatalf("re-decoding the chimera failed with %v; #41's harm is that it SUCCEEDS silently, so this test no longer describes the hazard", err)
	}
	if got := back.MetaData.Version.App.Raw; got != "1.77" || back.MetaData.Version.Schema != nil {
		t.Fatalf("the chimera re-decoded as App=%q Schema=%v, want it to pass silently as classic 1.77: without the silent success there is no harm to prevent",
			got, back.MetaData.Version.Schema)
	}

	// ---- No public path produces it ----
	//
	// The public API cannot be asked for "the W2025 codec with a classic
	// identity", because naming the identity IS naming the codec: an application
	// version resolves to one release, and that one release supplies both the
	// identity written into the bytes and the schema that selects the codec
	// writing them. Every reachable request is checked for the property the
	// chimera violates -- that the two agree.
	reachedW2025Codec := false
	for _, tc := range []struct {
		name string
		app  string
	}{
		{"classic map as classic 1.77", "1.77"},
		{"classic map as classic 1.73", "1.73"},
		// The nearest a caller can get to the chimera: ask for the map to be
		// written as the application version whose schema selects the W2025 codec.
		// It SUCCEEDS -- a legitimate upgrade, not a chimera -- and the file states
		// the W2025 identity. That is exactly the difference.
		{"classic map as w2025 2.06", "2.06"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := xmlio.MarshalXML(classic, tc.app)
			if err != nil {
				t.Fatalf("MarshalXML(%s, %q): %v", classicFixture, tc.app, err)
			}
			if len(data) == 0 {
				t.Fatalf("MarshalXML(%q) returned no bytes and no error", tc.app)
			}
			target, err := xmlio.Lookup(tc.app)
			if err != nil {
				t.Fatalf("Lookup(%q): %v", tc.app, err)
			}
			if target.Schema != nil {
				reachedW2025Codec = true
			}

			el := mapElement.Find(data)
			if el == nil {
				t.Fatalf("MarshalXML(%q) emitted no <map> element", tc.app)
			}

			// The bytes must state the TARGET's identity, not the source's. This is
			// identify doing its job, and it is the whole guard.
			if !bytes.Contains(el, []byte(`version="`+target.App.Raw+`"`)) {
				t.Errorf("MarshalXML(%q) wrote a <map> that does not state version=%q:\n%s", tc.app, target.App.Raw, el)
			}
			if target.Release != "" && !bytes.Contains(el, []byte(`release="`+target.Release+`"`)) {
				t.Errorf("MarshalXML(%q) wrote a <map> that does not state release=%q:\n%s", tc.app, target.Release, el)
			}
			if target.Schema != nil && !bytes.Contains(el, []byte(`schema="`+target.Schema.Raw+`"`)) {
				t.Errorf("MarshalXML(%q) wrote a <map> that does not state schema=%q:\n%s", tc.app, target.Schema.Raw, el)
			}

			// It must never be the chimera itself.
			if bytes.Equal(data, chimera) {
				t.Errorf("MarshalXML(%q) returned the chimera byte-for-byte", tc.app)
			}

			// And what it says it is must be what it decodes back as -- the check
			// the chimera fails silently.
			m2, err := decodeRawXML(t, data, target.XMLVersion)
			if err != nil {
				t.Fatalf("re-decode MarshalXML(%q): %v", tc.app, err)
			}
			if got := m2.MetaData.Version.App.Raw; got != target.App.Raw {
				t.Errorf("MarshalXML(%q) re-decodes as App=%q, want %q", tc.app, got, target.App.Raw)
			}
			switch {
			case target.Schema == nil && m2.MetaData.Version.Schema != nil:
				t.Errorf("MarshalXML(%q) re-decodes as Schema=%+v, want nil", tc.app, *m2.MetaData.Version.Schema)
			case target.Schema != nil && m2.MetaData.Version.Schema == nil:
				t.Errorf("MarshalXML(%q) re-decodes as Schema=nil, want %q", tc.app, target.Schema.Raw)
			case target.Schema != nil && m2.MetaData.Version.Schema.Raw != target.Schema.Raw:
				t.Errorf("MarshalXML(%q) re-decodes as Schema=%q, want %q", tc.app, m2.MetaData.Version.Schema.Raw, target.Schema.Raw)
			}
		})
	}

	// Guard against a vacuous pass: at least one case must reach the W2025 codec.
	// The classic targets could never produce a classic-identity chimera -- that
	// IS their identity -- so if no case routed to v1_06, nothing above tested the
	// pairing the chimera is made of.
	if !reachedW2025Codec {
		t.Errorf("no case targeted a release whose schema selects the W2025 codec, so no case exercises the identity/codec pairing the chimera abuses")
	}

	// The 2.06 target is the one that reaches the W2025 codec, so pin it hard: the
	// public output and the chimera come from the SAME codec and differ only in
	// the identity they state. That difference is the entire ticket.
	public, err := xmlio.MarshalXML(classic, "2.06")
	if err != nil {
		t.Fatalf(`MarshalXML(classic, "2.06"): %v`, err)
	}
	el := mapElement.Find(public)
	for _, forbidden := range []string{chimeraRelease, chimeraVersion, chimeraSchema} {
		if bytes.Contains(el, []byte(forbidden)) {
			t.Errorf("the public W2025 encode states %s -- the classic identity on W2025 content, which is the chimera:\n%s", forbidden, el)
		}
	}
}

// TestEncodeUnregisteredTargetProducesNoBytes closes the other half of the
// public encode contract for MarshalXML: an application version no release
// states is an error, and NOT a best-effort file.
//
// Encode's writer is already held to this by TestEncodeUnlicensedTargetWritesNothing.
// MarshalXML returns bytes rather than writing them, so the assertion is that it
// returns none: a caller who is about to os.WriteFile the result must not be
// handed a file for a release that does not exist.
func TestEncodeUnregisteredTargetProducesNoBytes(t *testing.T) {
	m, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}

	// Control: a registered target really does marshal. Guard against a vacuous
	// pass -- if this map could not be marshaled at all, the refusals below would
	// be the map's fault and would say nothing about target resolution.
	if data, err := xmlio.MarshalXML(m, m.MetaData.Version.App.Raw); err != nil {
		t.Fatalf("MarshalXML(%s, %q): %v; the refusals below would prove nothing", classicFixture, m.MetaData.Version.App.Raw, err)
	} else if len(data) == 0 {
		t.Fatalf("MarshalXML(%s, its own version) returned no bytes", classicFixture)
	}

	for _, tc := range []struct {
		name string
		app  string
	}{
		// The licensing case from ADR 0004 Decision 5: a build that does not exist.
		{"a future version", unlicensedTarget},
		// A schema version is not an application version. This is issue #41
		// requirement 1 stated as behavior: the one string that used to select a
		// codec here must now resolve to nothing.
		{"a schema, not an app version", "1.06"},
		{"the classic codec version, not an app version", "0.77"},
		{"unpadded 2.06", "2.6"},
		{"an unreleased classic", "1.75"},
		// "" is not a sentinel for "the map's own release".
		{"empty", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Guard against a vacuous pass: the version must genuinely be
			// unregistered, or this asks the encoder to refuse something it should
			// accept.
			if r, err := xmlio.Lookup(tc.app); err == nil {
				t.Fatalf("Lookup(%q) resolved to release %q: it is registered, so refusing it is not the contract under test", tc.app, r.Release)
			}
			data, err := xmlio.MarshalXML(m, tc.app)
			if err == nil {
				t.Fatalf("MarshalXML(%q) returned %d bytes and nil, want an error: an unregistered target must never be a best-effort write", tc.app, len(data))
			}
			if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
				t.Errorf("MarshalXML(%q) error = %v, want it to wrap %v", tc.app, err, wxx.ErrUnsupportedMapVersion)
			}
			if len(data) != 0 {
				t.Errorf("MarshalXML(%q) returned %d bytes alongside its error, want none", tc.app, len(data))
			}
		})
	}
}

// TestReleaseDescriptorHandsOutNoCodec is issue #41's acceptance criterion --
// "no public symbol names a schema version or returns a codec" -- checked
// mechanically over the public surface that can reach a codec.
//
// Release_t and Registry_t are the two types that HOLD codec knowledge, so they
// are where a codec would leak back out. The check is by shape rather than by
// name: a field or result that is a func, or a struct containing one, is a codec
// escaping regardless of what it is called. That catches a re-added Codec()
// under any name, which an assertion on the literal names "Decode"/"Encode"
// would not.
//
// Release_t.Schema is deliberately not flagged. It is a read-only descriptor
// field -- it answers "what does 2.06 write" -- and naming a schema in data is
// not the same as ACCEPTING one as a parameter to pick an encoder, which is what
// requirement 1 forbids and what this test's parameter check covers.
func TestReleaseDescriptorHandsOutNoCodec(t *testing.T) {
	// dottedPtr is the schema type a selector would take; a public method taking
	// one is CodecForSchema returning under a new name.
	schemaType := reflect.TypeOf((*wxx.Dotted)(nil))

	for _, tc := range []struct {
		name string
		typ  reflect.Type
	}{
		{"Release_t", reflect.TypeOf(&xmlio.Release_t{})},
		{"Registry_t", reflect.TypeOf(&xmlio.Registry_t{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Guard against a vacuous pass: reflection over a type with no exported
			// methods proves nothing, and both of these have some. If a refactor
			// ever emptied them, this fatals rather than passing silently.
			if tc.typ.NumMethod() == 0 {
				t.Fatalf("%s exposes no exported methods, so this test inspected nothing", tc.name)
			}
			for i := 0; i < tc.typ.NumMethod(); i++ {
				m := tc.typ.Method(i)
				for j := 1; j < m.Type.NumIn(); j++ { // 0 is the receiver
					if in := m.Type.In(j); in == schemaType {
						t.Errorf("%s.%s takes a %s: an encoder accepts an application version and never a schema (issue #41 requirement 1)", tc.name, m.Name, in)
					}
				}
				for j := 0; j < m.Type.NumOut(); j++ {
					if path := funcInType(m.Type.Out(j), nil); path != "" {
						t.Errorf("%s.%s returns a codec (%s at %s): the dispatcher picks the encoder and a caller may not hold one (issue #41 requirement 5)",
							tc.name, m.Name, m.Type.Out(j), path)
					}
				}
			}
		})
	}

	// Release_t's exported FIELDS are the other half: Decode and Encode used to
	// live here, so every Lookup result handed the caller a codec.
	rt := reflect.TypeOf(xmlio.Release_t{})
	if rt.NumField() == 0 {
		t.Fatalf("Release_t has no fields, so this test inspected nothing")
	}
	inspected := 0
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		inspected++
		if path := funcInType(f.Type, nil); path != "" {
			t.Errorf("Release_t.%s is a codec (%s at %s): a release descriptor states an identity and hands out no encoder (issue #41 requirement 5)", f.Name, f.Type, path)
		}
	}
	// Guard against a vacuous pass, the same way: if Release_t ever exported no
	// fields there would be nothing to check and the loop above would be silent.
	if inspected == 0 {
		t.Fatalf("Release_t exports no fields, so the codec-field check inspected nothing")
	}
}

// funcInType reports where a func hides inside t, or "" if none does.
//
// It recurses through structs, pointers, slices, arrays and maps because a codec
// does not have to be a bare func field to escape: a Codec_t struct, a []Codec_t
// or a *Codec_t hands out exactly the same thing. The returned path names the
// route to it so a failure is actionable. seen breaks recursive types.
func funcInType(t reflect.Type, seen map[reflect.Type]bool) string {
	if t == nil {
		return ""
	}
	if seen == nil {
		seen = map[reflect.Type]bool{}
	}
	if seen[t] {
		return ""
	}
	seen[t] = true

	switch t.Kind() {
	case reflect.Func:
		return t.String()
	case reflect.Ptr, reflect.Slice, reflect.Array:
		if p := funcInType(t.Elem(), seen); p != "" {
			return t.Kind().String() + " -> " + p
		}
	case reflect.Map:
		if p := funcInType(t.Elem(), seen); p != "" {
			return "map value -> " + p
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if p := funcInType(f.Type, seen); p != "" {
				return t.Name() + "." + f.Name + " -> " + p
			}
		}
	}
	return ""
}

// TestPublicEncodePathsNameAnApplicationVersion is the compile-level half of
// issue #41's acceptance criterion, stated so that a regression is a BUILD
// failure rather than a test failure.
//
// Each assertion pins the exact signature of a public entry point that reaches a
// codec. Re-adding a schema parameter, a *Release_t parameter or a codec result
// to any of them stops this file compiling.
//
// What it cannot state is the ABSENCE of a symbol -- Go has no "package does not
// export X" assertion -- so the absent ones are enforced by
// TestReleaseDescriptorHandsOutNoCodec at run time, and by the fact that this
// package would not compile if a test still referenced them.
var (
	// MarshalXML takes an APPLICATION VERSION STRING. Not a schema, not a
	// *Release_t, and it returns bytes rather than a codec.
	_ func(*wxx.Map_t, string) ([]byte, error) = xmlio.MarshalXML

	// WithTargetVersion is the one way to name a target, and it names it by
	// application version string.
	_ func(string) xmlio.EncoderOption = xmlio.WithTargetVersion

	// Lookup takes an application version string and returns a descriptor.
	_ func(string) (*xmlio.Release_t, error) = xmlio.Lookup

	// SupportedReleases hands out descriptors, not codecs.
	_ func() []*xmlio.Release_t = xmlio.SupportedReleases
)
