// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// xmlHeaderSamples pairs a real fixture with the XML declaration the release it
// states opens its files with, and with the schema that routes it there.
//
// wantSchema is the exact map/@schema bytes; "" means the fixture states none,
// which is the implicit legacy (classic) schema.
var xmlHeaderSamples = []struct {
	name       string
	path       string
	wantSchema string
	wantHeader string
}{
	{"classic 1.77", classicFixture, "", "<?xml version='1.0' encoding='utf-16'?>\n"},
	{"w2025 2.06", sample2025_206, "1.06", "<?xml version='1.1' encoding='utf-16'?>\n"},
}

// TestEncodeXMLHeaderFollowsRelease asserts, at the byte level and from real
// fixtures, that the XML declaration an encode emits follows the target RELEASE:
// classic opens `<?xml version='1.0'`, W2025 opens `<?xml version='1.1'`.
//
// This is the property the deleted `switch target.Major { case 2017: ...; case
// 2025: ... }` used to guarantee. The declaration is now data on the registry
// entry (Release_t.XMLVersion), so nothing in the encoder knows a family year --
// but the bytes of every file written must not have moved an inch, and that is
// what this pins.
//
// The bytes are the real ones: each fixture is encoded through the full default
// pipeline (XML, header, UTF-16/BE, gzip) and the output is transported back
// through the decoder, whose Converted diagnostic is the file's text as it would
// be read from disk -- header included, since it is captured before the header is
// consumed.
func TestEncodeXMLHeaderFollowsRelease(t *testing.T) {
	// Guard against a vacuous pass: this test discriminates only if the cases
	// expect DIFFERENT declarations. Were they ever collapsed to one expectation,
	// every assertion below would still pass against an encoder that hard-coded a
	// single header -- which is precisely the regression worth catching.
	distinct := map[string]bool{}
	for _, tc := range xmlHeaderSamples {
		distinct[tc.wantHeader] = true
	}
	if len(distinct) < 2 {
		t.Fatalf("all %d samples expect the same XML declaration, so 'the header follows the release' is not under test", len(xmlHeaderSamples))
	}

	for _, tc := range xmlHeaderSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}

			// Sanity: the fixture really does state the schema that routes it, so
			// the two cases exercise two different releases rather than one.
			schema := m.MetaData.Version.Schema
			if tc.wantSchema == "" {
				if schema != nil {
					t.Fatalf("%s: Version.Schema = %+v, want nil; this case must exercise the implicit legacy schema", tc.path, *schema)
				}
			} else if schema == nil {
				t.Fatalf("%s: Version.Schema = nil, want %q; this case must exercise the W2025 schema", tc.path, tc.wantSchema)
			} else if schema.Raw != tc.wantSchema {
				t.Fatalf("%s: Version.Schema.Raw = %q, want %q", tc.path, schema.Raw, tc.wantSchema)
			}

			var buf bytes.Buffer
			if err := xmlio.NewEncoder().Encode(&buf, m); err != nil {
				t.Fatalf("public encode %s (%v): %v", tc.path, m.MetaData.Version, err)
			}

			// Read the encoded file back to its text, header and all.
			var d xmlio.DecoderDiagnostics
			if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&d)).Decode(&buf); err != nil {
				t.Fatalf("re-decode encoded %s: %v", tc.path, err)
			}
			if len(d.Converted) == 0 {
				t.Fatalf("re-decode encoded %s: diagnostics.Converted is empty", tc.path)
			}
			if !bytes.HasPrefix(d.Converted, []byte(tc.wantHeader)) {
				t.Errorf("encoded %s opens %q, want it to open %q: the XML declaration must follow the release",
					tc.path, head(d.Converted, len(tc.wantHeader)), tc.wantHeader)
			}
			// The other release's declaration must not appear in its place.
			for _, other := range xmlHeaderSamples {
				if other.wantHeader == tc.wantHeader {
					continue
				}
				if bytes.HasPrefix(d.Converted, []byte(other.wantHeader)) {
					t.Errorf("encoded %s opens with %q, the declaration of a different release", tc.path, other.wantHeader)
				}
			}
		})
	}
}

// TestEncodeUnsupportedTargetIsError asserts that a target no supported release
// states is an error and writes nothing (ADR 0004 Decision 5). A best-effort
// write here would hand a user a file claiming to be a release that does not
// exist, or one they are not licensed for.
//
// The target reaches the encoder two ways -- the caller names it, or the map
// states it -- and neither may fall back to a nearest match.
func TestEncodeUnsupportedTargetIsError(t *testing.T) {
	// Control: the fixture encodes cleanly as itself. Guard against a vacuous
	// pass -- if this map could not be encoded at all, every error below would be
	// the map's fault and would say nothing about target resolution.
	base, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}
	var control bytes.Buffer
	if err := xmlio.NewEncoder().Encode(&control, base); err != nil {
		t.Fatalf("public encode %s as itself: %v; the unsupported-target cases below would prove nothing", classicFixture, err)
	}
	if control.Len() == 0 {
		t.Fatalf("public encode %s as itself: empty output", classicFixture)
	}

	for _, tc := range []struct {
		name string
		// app, when non-empty, is the version the map states; the caller names
		// nothing and the encoder must fall back to it.
		app string
		// target, when non-empty, is the version the caller names.
		target string
	}{
		{"caller names a future version", "", "9.99"},
		{"caller names an unpadded 2.06", "", "2.6"},
		{"caller names a schema, not an app version", "", "1.06"},
		{"caller names an unreleased classic", "", "1.75"},
		{"map states a future version", "9.99", ""},
		{"map states an unpadded 2.06", "2.6", ""},
		{"map states nothing", "", ""}, // App.Raw == "": no release states an empty version
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, classicFixture)
			if err != nil {
				t.Fatalf("public decode %s: %v", classicFixture, err)
			}
			// Rewrite the identity the map states, keeping Raw authoritative.
			m.MetaData.Version.App = wxx.Dotted{Raw: tc.app}
			if d, err := wxx.ParseDotted(tc.app); err == nil {
				m.MetaData.Version.App = d
			}

			var opts []xmlio.EncoderOption
			if tc.target != "" {
				opts = append(opts, xmlio.WithTargetVersion(tc.target))
			}

			var buf bytes.Buffer
			err = xmlio.NewEncoder(opts...).Encode(&buf, m)
			if err == nil {
				t.Fatalf("encode targeting an unsupported release wrote %d bytes and returned nil, want an error: an unregistered target must never be a best-effort write", buf.Len())
			}
			if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
				t.Errorf("encode error = %v, want it to wrap %v", err, wxx.ErrUnsupportedMapVersion)
			}
			// Nothing may reach the writer: a rejected target is not a partial file.
			if buf.Len() != 0 {
				t.Errorf("encode wrote %d bytes before failing, want 0", buf.Len())
			}
		})
	}
}
