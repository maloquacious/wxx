// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
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

// informationTag matches an <information> start tag, an end tag, or an empty
// element, so shallowInformation can track nesting depth.
var informationTag = regexp.MustCompile(`<(/?)information(\s[^>]*)?(/?)>`)

// shallowInformation returns doc with every <information> element nested deeper
// than maxDepth removed, start tag through end tag.
//
// It exists only because of issue #69: Map_t models two levels of <information>
// and real files carry three, so the third level is lost on decode. This test
// is about #66, which is attributes, not depth; it compares the levels Map_t
// models and leaves the missing level to #69. Fixing #69 means deleting this
// function and its two calls, and the test must then still pass.
func shallowInformation(doc []byte, maxDepth int) []byte {
	var out []byte
	depth, last, cutFrom := 0, 0, -1
	for _, loc := range informationTag.FindAllSubmatchIndex(doc, -1) {
		closing := loc[3] > loc[2]
		empty := loc[7] > loc[6]
		switch {
		case closing:
			if depth == maxDepth+1 && cutFrom >= 0 {
				out = append(out, doc[last:cutFrom]...)
				last, cutFrom = loc[1], -1
			}
			depth--
		default:
			depth++
			if depth == maxDepth+1 && cutFrom < 0 {
				cutFrom = loc[0]
			}
			if empty {
				if depth == maxDepth+1 {
					out = append(out, doc[last:cutFrom]...)
					last, cutFrom = loc[1], -1
				}
				depth--
			}
		}
	}
	return append(out, doc[last:]...)
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
// Only the first two levels of nesting are compared, until issue #69 is fixed;
// see shallowInformation.
func TestW2025InformationAttrsMatchSource(t *testing.T) {
	type pair struct{ in, out []byte }
	cases := map[string]func(t *testing.T) pair{}

	for _, fixture := range []string{
		"2025-2.06-13x11-941577-blank.wxx",
		"2025-2.06-13x11-941577-layers.wxx",
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
			in := startTagAttrs(shallowInformation(docs.in, 2), "information")
			out := startTagAttrs(shallowInformation(docs.out, 2), "information")
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
