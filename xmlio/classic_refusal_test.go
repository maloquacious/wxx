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

// classicFixture is the one classic (Worldographer 1.x) file still tracked. wxx
// no longer reads classic maps (issue #103); this file is kept so the refusal is
// pinned against bytes Worldographer actually wrote, not only against a W2025
// document with its identity rewritten.
const classicFixture = "../testdata/2017-1.77-1.0-columns-blank.wxx"

// releaseAttr matches the map/@release attribute inside the <map> start tag.
var releaseAttr = regexp.MustCompile(`release="[^"]*"`)

// withMapRelease returns raw XML with map/@release replaced by want, rewriting
// only the <map> start tag (see withMapVersion for why).
func withMapRelease(t *testing.T, raw []byte, want string) []byte {
	t.Helper()
	tag := mapElement.Find(raw)
	if tag == nil {
		t.Fatalf("no <map> start tag in %d bytes of XML", len(raw))
	}
	rewritten := releaseAttr.ReplaceAll(tag, []byte(`release="`+want+`"`))
	if string(rewritten) == string(tag) {
		t.Fatalf("<map> tag was not rewritten; it states no @release: %s", head(tag, 200))
	}
	return append(append([]byte{}, rewritten...), raw[len(tag):]...)
}

// TestDecodeRefusesAClassicMap pins the refusal on a real classic file: the
// decode fails with wxx.ErrClassicMap, the message names the version the file
// states, and no Map_t comes back.
//
// "Refused" is precise here. The decoder reads the file's transport and its <map>
// start tag -- it has to, to know what the file is -- and stops at dispatch: no
// codec is chosen (DecoderDiagnostics.Codec stays empty) and none of the map body
// is decoded. A best-effort decode would have set Codec and returned a map.
func TestDecodeRefusesAClassicMap(t *testing.T) {
	var diag xmlio.DecoderDiagnostics
	m, err := xmlio.ReadFile(classicFixture, xmlio.WithDecoderDiagnostics(&diag))
	if err == nil {
		t.Fatalf("ReadFile(%s) = map, nil; want an error: classic maps are refused (issue #103)", classicFixture)
	}
	if !errors.Is(err, wxx.ErrClassicMap) {
		t.Errorf("ReadFile(%s) error = %v, want it to wrap %v", classicFixture, err, wxx.ErrClassicMap)
	}
	// The error must name what was found, or the user cannot tell why.
	if !strings.Contains(err.Error(), `"1.77"`) {
		t.Errorf("ReadFile(%s) error = %q, want it to name the version the file states, \"1.77\"", classicFixture, err)
	}
	// And it must say what to do about it.
	if !strings.Contains(err.Error(), "Worldographer 2025") {
		t.Errorf("ReadFile(%s) error = %q, want it to tell the user to convert the map in Worldographer 2025", classicFixture, err)
	}
	if m != nil {
		t.Errorf("ReadFile(%s) returned a map alongside its error, want nil: a refused file is not decoded", classicFixture)
	}

	// Guard against a vacuous pass: the decode must have got as far as reading
	// the <map> tag, or the error above came from somewhere other than the
	// dispatch under test.
	if !strings.Contains(string(diag.MapElement), `version="1.77"`) {
		t.Fatalf("diagnostics captured <map> tag %q, want one stating version=\"1.77\": the decode did not reach dispatch", diag.MapElement)
	}
	if diag.Codec != "" {
		t.Errorf("DecoderDiagnostics.Codec = %q, want empty: a refused classic file reaches no codec", diag.Codec)
	}
}

// TestDecodeRefusesAMapWithNoRelease covers every file that states no
// map/@release, which is how a classic file identifies itself. Each one is
// refused: a "1." version as a classic map, anything else as unsupported
// metadata -- never decoded best-effort, and never as W2025.
//
// The documents are a W2025 2.06 file with its @release blanked and its
// @version rewritten. Everything else about them is valid W2025, which is the
// point: only the identity decides, so content a codec could read must not get
// one through.
func TestDecodeRefusesAMapWithNoRelease(t *testing.T) {
	src, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("public decode %s: %v", sample2025_206, err)
	}
	raw, err := xmlio.MarshalXML(src, "2.06")
	if err != nil {
		t.Fatalf("MarshalXML(%s, %q): %v", sample2025_206, "2.06", err)
	}

	for _, tc := range []struct {
		name    string
		version string
		want    error
		notWant error
	}{
		{"classic 1.73", "1.73", wxx.ErrClassicMap, wxx.ErrUnsupportedMapMetadata},
		{"classic 1.74", "1.74", wxx.ErrClassicMap, wxx.ErrUnsupportedMapMetadata},
		{"classic 1.77", "1.77", wxx.ErrClassicMap, wxx.ErrUnsupportedMapMetadata},
		{"malformed 1.x", "1.x", wxx.ErrClassicMap, wxx.ErrUnsupportedMapMetadata},
		{"w2025 version, no release", "2.07", wxx.ErrUnsupportedMapMetadata, wxx.ErrClassicMap},
		{"1-prefixed but not 1.x", "10.0", wxx.ErrUnsupportedMapMetadata, wxx.ErrClassicMap},
		{"garbage", "garbage", wxx.ErrUnsupportedMapMetadata, wxx.ErrClassicMap},
		{"empty", "", wxx.ErrUnsupportedMapMetadata, wxx.ErrClassicMap},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := withMapVersion(t, withMapRelease(t, raw, ""), tc.version)

			// Guard against a vacuous pass: the tag must state exactly the identity
			// under test.
			tag := string(mapElement.Find(doc))
			if !strings.Contains(tag, `release=""`) || !strings.Contains(tag, `version="`+tc.version+`"`) {
				t.Fatalf("rewritten <map> tag does not state release=\"\" version=%q: %s", tc.version, head([]byte(tag), 200))
			}

			var diag xmlio.DecoderDiagnostics
			m, err := xmlio.NewDecoder(
				xmlio.WithSkipUncompress(),
				xmlio.WithUTF16BEInput(false),
				xmlio.WithDecoderDiagnostics(&diag),
			).Decode(strings.NewReader("<?xml version='1.1' encoding='utf-8'?>\n" + string(doc)))
			if err == nil {
				t.Fatalf("decode of release=\"\" version=%q = map, nil; want an error wrapping %v", tc.version, tc.want)
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("decode of release=\"\" version=%q error = %v, want it to wrap %v", tc.version, err, tc.want)
			}
			if errors.Is(err, tc.notWant) {
				t.Errorf("decode of release=\"\" version=%q error = %v, want it not to wrap %v", tc.version, err, tc.notWant)
			}
			if !strings.Contains(err.Error(), `"`+tc.version+`"`) {
				t.Errorf("decode of release=\"\" version=%q error = %q, want it to name the version", tc.version, err)
			}
			if m != nil {
				t.Errorf("decode of release=\"\" version=%q returned a map alongside its error, want nil", tc.version)
			}
			if diag.Codec != "" {
				t.Errorf("decode of release=\"\" version=%q reached codec %q, want none", tc.version, diag.Codec)
			}
		})
	}
}
