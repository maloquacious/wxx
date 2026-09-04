// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// versionAttr matches the map/@version attribute inside the <map> start tag,
// so a test can state an on-disk version the encoders would never write.
var versionAttr = regexp.MustCompile(`version="[^"]*"`)

// withMapVersion returns raw XML with map/@version replaced by want.
//
// It rewrites only the <map> start tag: @version is the identity attribute, and
// the string appears elsewhere in a file (the XML declaration is already gone at
// this point, but <mapLayer> and friends are not). Rewriting the whole document
// would change more than the identity under test.
func withMapVersion(t *testing.T, raw []byte, want string) []byte {
	t.Helper()
	tag := mapElement.Find(raw)
	if tag == nil {
		t.Fatalf("no <map> start tag in %d bytes of XML", len(raw))
	}
	rewritten := versionAttr.ReplaceAll(tag, []byte(`version="`+want+`"`))
	if string(rewritten) == string(tag) {
		t.Fatalf("<map> tag was not rewritten; it states no @version: %s", head(tag, 200))
	}
	out := append([]byte{}, raw...)
	return append(append(out[:0:0], rewritten...), out[len(tag):]...)
}

// TestMalformedOnDiskVersionDecodesUnparsed is issue #38 observed end-to-end
// through the public decoder, which is the only place the bug could ever have
// bitten: nothing constructs these values but a codec.
//
// Two claims, and they pull in opposite directions on purpose.
//
// FIRST, the file still decodes. Issue #32 fixed the constraint that modeling
// these values must not make decoding stricter, so a map/@version that does not
// fit the dotted grammar is carried, not rejected. Both codecs keep the bytes
// verbatim in Dotted.Raw, and that is what an encoder would write back.
//
// SECOND, the version it decodes to has NO components. Before #38 it had zero
// ones -- {Major: 0, Minor: 0} -- which Compare read as an ordinal, reporting the
// file EQUAL to "0.0" and LESS than every real version. The components were a
// lie the type had no way to disown. Now Parsed() is false and Compare refuses.
//
// The two malformed strings differ, and the difference is load-bearing. The
// dispatcher (decoder.go) routes a classic file only on a "1." prefix, so
// "garbage" never reaches the classic codec at all -- it misses dispatch and is
// an unsupported-metadata error. "1.x" carries the prefix and fails the dotted
// grammar, which is what makes v0_77's fallback reachable through the public
// API rather than only by calling it directly. release="2025" routes
// unconditionally, so W2025 takes anything non-empty.
func TestMalformedOnDiskVersionDecodesUnparsed(t *testing.T) {
	for _, tc := range []struct {
		name       string
		fixture    string
		app        string // application version to marshal the fixture as
		xmlVersion string // XML declaration to read it back under
		malformed  string // the map/@version the file will state
	}{
		{
			name:       "classic",
			fixture:    classicFixture,
			app:        "1.77",
			xmlVersion: "1.0",
			malformed:  "1.x", // keeps the "1." dispatch prefix, fails the grammar
		},
		{
			name:       "w2025",
			fixture:    sample2025_206,
			app:        "2.06",
			xmlVersion: "1.1",
			malformed:  "garbage", // release="2025" routes without inspecting @version
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.fixture)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.fixture, err)
			}
			raw, err := xmlio.MarshalXML(m, tc.app)
			if err != nil {
				t.Fatalf("MarshalXML(%s, %q): %v", tc.fixture, tc.app, err)
			}

			// Guard against a vacuous pass: the string this test calls malformed
			// must actually be malformed. If ParseDotted ever accepted it, every
			// assertion below would be asserting the opposite of what it says.
			if _, err := wxx.ParseDotted(tc.malformed); err == nil {
				t.Fatalf("ParseDotted(%q) succeeded, so this test states a well-formed version and proves nothing", tc.malformed)
			}

			got, err := xmlio.NewDecoder(
				xmlio.WithSkipUncompress(),
				xmlio.WithUTF16BEInput(false),
			).Decode(strings.NewReader(
				"<?xml version='" + tc.xmlVersion + "' encoding='utf-8'?>\n" +
					string(withMapVersion(t, raw, tc.malformed)),
			))

			// FIRST: it decodes. A malformed version is not a decode error.
			if err != nil {
				t.Fatalf("decode of a file stating version=%q failed: %v\ndecoding must not become stricter (issue #32); the codec is expected to carry the bytes, not reject them", tc.malformed, err)
			}

			app := got.MetaData.Version.App

			// Raw is authoritative and survives verbatim.
			if app.Raw != tc.malformed {
				t.Errorf("MetaData.Version.App.Raw = %q, want %q verbatim", app.Raw, tc.malformed)
			}
			if provenance := got.MetaData.Worldographer.Version; provenance != tc.malformed {
				t.Errorf("MetaData.Worldographer.Version = %q, want %q verbatim", provenance, tc.malformed)
			}

			// SECOND: it has no components, and says so.
			if app.Parsed() {
				t.Fatalf("MetaData.Version.App.Parsed() = true for version=%q, want false: the decoder could not parse it, so it has no components", tc.malformed)
			}

			// And therefore it cannot be ordered. These are the exact two
			// questions issue #38 named -- "is this equal to 0.0" and "is this
			// older than a real version" -- both of which used to be answered
			// yes, confidently and wrongly.
			for _, other := range []struct {
				label string
				d     wxx.Dotted
			}{
				{`"0.0"`, mustDotted(t, "0.0")},
				{`the real version this file was made from`, mustDotted(t, tc.app)},
			} {
				if c, err := app.Compare(other.d); err == nil {
					t.Errorf("version=%q Compare(%s) = %d, nil; want an error wrapping %v", tc.malformed, other.label, c, wxx.ErrUnparsedDottedVersion)
				} else if !errors.Is(err, wxx.ErrUnparsedDottedVersion) {
					t.Errorf("version=%q Compare(%s) error = %v, want it to wrap %v", tc.malformed, other.label, err, wxx.ErrUnparsedDottedVersion)
				}
				if less, err := app.Less(other.d); err == nil {
					t.Errorf("version=%q Less(%s) = %v, nil; want an error", tc.malformed, other.label, less)
				}
			}
		})
	}
}

// TestWellFormedOnDiskVersionDecodesParsed is the other half, and without it the
// test above passes against a Dotted that is never Parsed at all.
//
// Every tracked fixture states a well-formed version, so every one must decode
// to a version that IS comparable. This is what proves the flag tracks the parse
// rather than being stuck off.
func TestWellFormedOnDiskVersionDecodesParsed(t *testing.T) {
	for _, tc := range versionIdentitySamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			v := m.MetaData.Version
			if !v.App.Parsed() {
				t.Errorf("MetaData.Version.App.Parsed() = false for %q, want true: it is a well-formed dotted version", v.App.Raw)
			}
			if _, err := v.App.Compare(mustDotted(t, "0.0")); err != nil {
				t.Errorf("MetaData.Version.App.Compare(\"0.0\"): unexpected error: %v", err)
			}
			if v.Schema == nil {
				return // classic states none; ADR 0003 Decision 2
			}
			if !v.Schema.Parsed() {
				t.Errorf("MetaData.Version.Schema.Parsed() = false for %q, want true", v.Schema.Raw)
			}
		})
	}
}
