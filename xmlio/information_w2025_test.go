// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// loreAttrs are the eight optional attributes an <information> element may
// state. Which ones it states depends on the entry's type: an "Information"
// entry states none, a "Nation" rulers/government/cultures, a "Religion"
// religionType/culture/holySymbol/domains.
var loreAttrs = map[string]bool{
	"rulers": true, "government": true, "cultures": true, "language": true,
	"religionType": true, "culture": true, "holySymbol": true, "domains": true,
}

// TestW2025InformationAttrsMatchSource asserts that the W2025 encoder writes
// back every <information> attribute the source stated, and no other, with the
// same names in the same order and the same values (issue #66). It covers the
// nested detail elements too: they are <information> elements as well.
//
// The bug this pins: every <information> came back stating all eight lore
// attributes, as "" where the source stated none of them, because Map_t held
// them as plain strings and the encoder wrote them unconditionally.
//
// The fix has a mirror-image failure this test also guards. Gating on the
// VALUE -- write only non-empty -- would have made the reported case pass and
// dropped the domains="" Worldographer writes on every Religion entry. The
// fixtures state both shapes, and the vacuity checks at the bottom insist they
// keep doing so.
//
// Every <information> is compared, at every depth. Real files nest three deep,
// and until issue #69 the third level was lost on decode; the element-count
// check in compareStartTags is what holds that fixed.
func TestW2025InformationAttrsMatchSource(t *testing.T) {
	type pair struct{ in, out []byte }
	cases := map[string]func(t *testing.T) pair{}

	for _, fixture := range []string{
		"2025-2.06-13x11-941577-blank.wxx",
		"2025-2.06-13x11-941577-layers-beta.wxx",
	} {
		cases[fixture] = func(t *testing.T) pair {
			path := filepath.Join("..", "testdata", fixture)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("open %s: %v", path, err)
			}
			defer f.Close()
			var dd xmlio.DecoderDiagnostics
			m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f)
			if err != nil {
				t.Fatalf("decode %s: %v", fixture, err)
			}
			var ed xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder("2.06", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode %s: %v", fixture, err)
			}
			return pair{dd.Converted, ed.Utf8Encoded}
		}
	}

	// The populated sample is the only fixture with Nation and Culture entries
	// nested as details. It is raw XML, so it bypasses the transport stages.
	cases[filepath.Base(populatedFixture)] = func(t *testing.T) pair {
		raw, err := os.ReadFile(populatedFixture)
		if err != nil {
			t.Fatalf("read %s: %v", populatedFixture, err)
		}
		m := decodeFixture(t, populatedFixture)
		out, err := v1_06.Encode(m, "2.06")
		if err != nil {
			t.Fatalf("encode %s: %v", populatedFixture, err)
		}
		return pair{raw, out}
	}

	for name, load := range cases {
		t.Run(name, func(t *testing.T) {
			docs := load(t)
			in := startTagAttrs(docs.in, "information")
			out := startTagAttrs(docs.out, "information")
			compareStartTags(t, name, "information", in, out)

			// The regression is pinned only if the fixture states both shapes.
			sawNone, sawEmpty := false, false
			for _, attrs := range in {
				stated := 0
				for _, attr := range attrs {
					if loreAttrs[attr[0]] {
						stated++
						if attr[1] == "" {
							sawEmpty = true
						}
					}
				}
				if stated == 0 {
					sawNone = true
				}
			}
			if !sawNone {
				t.Errorf("%s: every <information> states a lore attribute, so the invented-attribute regression is untested here", name)
			}
			if !sawEmpty {
				t.Errorf("%s: no <information> states a lore attribute as \"\", so the value-gating regression is untested here", name)
			}
		})
	}
}
