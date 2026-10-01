// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// attrPattern matches one name="value" pair of an element's start tag. Values
// in these files never contain a quote, so the non-greedy character class is
// enough and there is no need to parse the document.
var attrPattern = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_]*)="([^"]*)"`)

// startTagAttrs extracts every start tag of the named element from a UTF-8 XML
// document, in document order, each as an ordered list of name/value pairs.
//
// It reads the raw text rather than decoding the document, and that is the
// point: encoding/xml would normalize away attribute ORDER, and cannot tell an
// absent attribute from one stated as "" -- both of which are part of what these
// comparisons are about. Go's decoder also refuses the version="1.1"
// declaration every W2025 file opens with.
//
// It matches tags, not lines, because the encoder writes nested elements
// back-to-back: every <information> detail shares a line with its parent.
func startTagAttrs(doc []byte, element string) [][][2]string {
	tag := regexp.MustCompile(`<` + regexp.QuoteMeta(element) + `(\s[^>]*)?/?>`)
	var out [][][2]string
	for _, m := range tag.FindAllSubmatch(doc, -1) {
		attrs := [][2]string{}
		for _, a := range attrPattern.FindAllStringSubmatch(string(m[1]), -1) {
			attrs = append(attrs, [2]string{a[1], a[2]})
		}
		out = append(out, attrs)
	}
	return out
}

// labelStyleAttrs extracts every <labelstyle> start tag; see startTagAttrs.
func labelStyleAttrs(doc []byte) [][][2]string {
	return startTagAttrs(doc, "labelstyle")
}

// compareStartTags reports, as test errors, every way the start tags the
// encoder wrote differ from the source's: a different number of elements, of
// attributes on an element, an attribute name or position, or a value. The
// only tolerated difference is a spellingExempt attribute whose two spellings
// are the same number.
//
// Inter-attribute whitespace is not compared, because it is not data; see
// TestW2025LabelStyleAttrsMatchSource for why that is a deliberate choice.
func compareStartTags(t *testing.T, fixture, element string, in, out [][][2]string) {
	t.Helper()
	if len(out) != len(in) {
		t.Errorf("%s: wrote %d <%s> element(s), source has %d", fixture, len(out), element, len(in))
		return
	}
	for i := range in {
		if len(out[i]) != len(in[i]) {
			t.Errorf("%s: <%s> %d has %d attribute(s) %v, source has %d %v", fixture, element, i, len(out[i]), names(out[i]), len(in[i]), names(in[i]))
			continue
		}
		for j, want := range in[i] {
			got := out[i][j]
			if got[0] != want[0] {
				t.Errorf("%s: <%s> %d attribute %d is %q, source has %q (order or name drift)", fixture, element, i, j, got[0], want[0])
				continue
			}
			if got[1] == want[1] {
				continue
			}
			if spellingExempt[want[0]] && sameNumber(want[1], got[1]) {
				continue
			}
			t.Errorf("%s: <%s> %d @%s = %q, source has %q", fixture, element, i, want[0], got[1], want[1])
		}
	}
}

// names lists the attribute names of one start tag, for error messages.
func names(attrs [][2]string) []string {
	var out []string
	for _, a := range attrs {
		out = append(out, a[0])
	}
	return out
}

// sameNumber reports whether two attribute values are one number spelled two
// ways ("0" and "0.0"). It is value equality, so exempting a difference it
// approves can never hide a changed value -- only a changed spelling.
func sameNumber(a, b string) bool {
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	return errA == nil && errB == nil && fa == fb
}

// spellingExempt is EMPTY, and its emptiness is the point (issue #64).
//
// It used to hold dropShadowRadius and dropShadowSpread, because Worldographer
// writes them "0" where this codec wrote "0.0" and nobody had established which
// spelling the application required. The experiment settled it: the decimal
// spelling is a hard load failure -- Integer.parseInt throws and the file does
// not open -- so the codec now states them as integers and there is nothing left
// to tolerate.
//
// It is kept rather than deleted because the exemption mechanism is the honest
// way to hold a test open over an unsettled question, and the next one will want
// it. An entry here means "the values are equal and we have not established
// which spelling is correct" -- never "this difference is acceptable". Anything
// proven wrong gets fixed instead, and anything proven harmless gets its
// justification written here beside it.
//
// TestW2025IntegerAttributeSpelling now covers the same ground far more widely:
// every attribute any tracked document spells integrally, in every element, not
// just the two named here.
var spellingExempt = map[string]bool{}

// TestW2025LabelStyleAttrsMatchSource asserts that the W2025 encoder writes back
// every <labelstyle> attribute the source file stated, with the same names in
// the same order and the same values (issue #62).
//
// The bug this pins: backgroundColor="null" came back as "0.0,0.0,0.0,1.0".
// decodeRgba mapped both "null" and black-opaque to nil and rgbas rendered nil
// as black-opaque, so every W2025 file this codec wrote claimed an opaque black
// label background where the original said "no background".
//
// Nothing caught it because the W2025 round-trip tests compare Map_t STRUCTURES
// (decode -> encode -> decode) and decodeRgba collapsed both spellings to the
// same nil, so the change was invisible to them by construction. A comparison of
// the input document against the output document would have reported it. This
// test gives W2025 that comparison for the one element where the drift was
// found.
//
// It compares parsed attributes rather than raw bytes, and both deviations from
// a byte comparison are deliberate:
//
//   - INTER-ATTRIBUTE WHITESPACE. The source writes two spaces ahead of color,
//     backgroundColor, outlineSize and dropShadowColor; this codec writes one.
//     Whitespace between attributes is not data in XML, no value depends on it,
//     and normalizing it is not a loss.
//   - NUMERIC SPELLING. See spellingExempt and issue #64.
//
// Everything else -- names, order, and values -- must match exactly.
func TestW2025LabelStyleAttrsMatchSource(t *testing.T) {
	for _, fixture := range []string{
		"2025-2.06-13x11-941577-blank.wxx",
		"2025-2.06-13x11-941577-layers-beta.wxx",
		"2025-2.07-13x11-941577-blank.wxx",
		"2025-2.07-13x11-941577-layers.wxx",
		"2025-2.07-13x11-941577-notes-shapes.wxx",
		"2025-2.07-13x11-941577-resources.wxx",
		"2025-2.07-13x11-941577-rows.wxx",
	} {
		t.Run(fixture, func(t *testing.T) {
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
			if err := xmlio.NewEncoder(sameVersionTarget(t, fixture), xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode %s: %v", fixture, err)
			}

			in := labelStyleAttrs(dd.Converted)
			out := labelStyleAttrs(ed.Utf8Encoded)
			if len(in) == 0 {
				t.Fatalf("%s: source carries no <labelstyle>; this fixture cannot evidence anything", fixture)
			}
			compareStartTags(t, fixture, "labelstyle", in, out)

			sawNull := false
			for _, style := range in {
				for _, attr := range style {
					if attr[0] == "backgroundColor" && attr[1] == "null" {
						sawNull = true
					}
				}
			}

			// The regression is only pinned if the fixture actually exercises it.
			// Both tracked W2025 fixtures state backgroundColor="null" on every
			// label style; if that ever stops being true, this test is passing
			// vacuously and should say so rather than look green.
			if !sawNull {
				t.Errorf("%s: no label style carries backgroundColor=\"null\", so the issue #62 regression is untested here", fixture)
			}
		})
	}
}

// TestW2025LabelStyleBlackBackgroundIsNotNull is the other half of the fix, and
// the reason decode moved to decodeZeroableRgba rather than the encoder simply
// switching to rgbans (issue #62).
//
// rgbans would have fixed the reported bug and introduced its mirror image: it
// decides by comparing the FORMATTED string, so an opaque black -- a colour a
// user can legitimately choose -- would have been laundered into "null". The
// round trip would then be stable and still wrong, in the way that is hardest to
// notice, because both spellings decode to something self-consistent.
//
// No fixture carries a black background, so the source is synthesized. A guard
// no fixture exercises is worth having only if something exercises it.
func TestW2025LabelStyleBlackBackgroundIsNotNull(t *testing.T) {
	m := newRowsMap()
	m.Configuration.TextConfig.LabelStyles = []*wxx.LabelStyle_t{
		{
			Name:            "Black Background",
			FontFace:        "Arial",
			Scale:           25,
			BackgroundColor: &wxx.RGBA_t{R: 0, G: 0, B: 0, A: 1},
			OutlineColor:    &wxx.RGBA_t{R: 0, G: 0, B: 0, A: 1},
		},
	}

	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.06", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("encode: %v", err)
	}

	styles := labelStyleAttrs(ed.Utf8Encoded)
	if len(styles) != 1 {
		t.Fatalf("wrote %d <labelstyle> element(s), want 1", len(styles))
	}
	for _, attr := range styles[0] {
		switch attr[0] {
		case "backgroundColor", "outlineColor":
			if attr[1] != "0.0,0.0,0.0,1.0" {
				t.Errorf("@%s = %q, want %q -- an opaque black must not be written as \"null\"", attr[0], attr[1], "0.0,0.0,0.0,1.0")
			}
		}
	}

	// And it must survive the trip: a colour written correctly but decoded back
	// to nil would be the same loss one step later.
	back, err := xmlio.NewDecoder().Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	got := back.Configuration.TextConfig.LabelStyles
	if len(got) != 1 {
		t.Fatalf("re-decode: %d label style(s), want 1", len(got))
	}
	if got[0].BackgroundColor == nil {
		t.Errorf("re-decode: BackgroundColor is nil, want an opaque black -- nil means \"null\"")
	}
}
