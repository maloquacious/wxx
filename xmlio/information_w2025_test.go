// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
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

	// Each fixture nests Nation and Culture entries as details of an
	// "Information" entry, and Religion entries a level deeper. 2.07 is listed
	// because it is the first stable release of the W2025 schema. The 2.06
	// populated map's lore has five non-ASCII titles, which the app spells as
	// decimal character references (title="Fabi&#225;n"); comparing it byte for
	// byte pins that the encoder spells them the same way (issue #96).
	for _, fixture := range []string{
		"2025-2.06-13x11-941577-blank.wxx",
		"2025-2.06-13x11-941577-layers-beta.wxx",
		"2025-2.06-13x11-941577-populated.wxx",
		"2025-2.07-13x11-941577-blank.wxx",
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

	charRef := false // some source attribute spells a character as &#N;
	for name, load := range cases {
		t.Run(name, func(t *testing.T) {
			docs := load(t)
			in := startTagAttrs(docs.in, "information")
			out := startTagAttrs(docs.out, "information")
			compareStartTags(t, name, "information", in, out)
			for _, attrs := range in {
				for _, attr := range attrs {
					charRef = charRef || strings.Contains(attr[1], "&#")
				}
			}

			// The regression is pinned only if the fixture states both shapes.
			sawNone, sawEmpty := false, false
			sawNation, sawCulture := false, false
			for _, attrs := range in {
				stated := 0
				for _, attr := range attrs {
					switch {
					case attr[0] == "type" && attr[1] == "Nation":
						sawNation = true
					case attr[0] == "type" && attr[1] == "Culture":
						sawCulture = true
					}
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
			if !sawNation || !sawCulture {
				t.Errorf("%s: Nation entry present = %v, Culture entry present = %v; want both, so their lore attributes are under test", name, sawNation, sawCulture)
			}
		})
	}
	if !charRef {
		t.Errorf("no fixture states a character reference in an <information> attribute, so the non-ASCII spelling (issue #96) is untested here")
	}
}

// informationsElement returns the <informations> element of a UTF-8 document,
// start tag to end tag.
func informationsElement(t *testing.T, label string, doc []byte) []byte {
	t.Helper()
	start := bytes.Index(doc, []byte("<informations>"))
	end := bytes.Index(doc, []byte("</informations>"))
	if start < 0 || end < start {
		t.Fatalf("%s: no <informations> element", label)
	}
	return doc[start : end+len("</informations>")]
}

// encodeFixture decodes a tracked fixture, encodes it as "2.06", and returns
// the source document, the decoded map and the encoded document, all UTF-8.
func encodeFixture(t *testing.T, fixture string) (in []byte, m *wxx.Map_t, out []byte) {
	t.Helper()
	path := filepath.Join("..", "testdata", fixture)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	var dd xmlio.DecoderDiagnostics
	m, err = xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", fixture, err)
	}
	out, err = xmlio.MarshalXML(m, "2.06")
	if err != nil {
		t.Fatalf("encode %s: %v", fixture, err)
	}
	return dd.Converted, m, out
}

// TestW2025InformationsMatchSource decodes every tracked W2025 fixture,
// encodes it as "2.06", and asserts the encoded <informations> element is
// byte-identical to the source's (issue #107).
//
// The bug this pins: every lore body was written as escaped text
// (&lt;h1&gt;, newlines as &#10;, apostrophes as &#39;) where Worldographer
// writes a CDATA section, and the newlines between entries came out as
// &#10; too. A parser reads both spellings as the same characters, so a
// Map_t round trip, and every attribute-level test, passed.
//
// Byte identity also holds the app's layout: the body, a newline, each nested
// entry followed by a newline, then a newline and the end tag; and " >" to
// close a start tag that states any lore attribute.
func TestW2025InformationsMatchSource(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "testdata", "2025-*.wxx"))
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("glob W2025 fixtures: %d found, err %v", len(fixtures), err)
	}
	sawGuard := false
	for _, path := range fixtures {
		fixture := filepath.Base(path)
		t.Run(fixture, func(t *testing.T) {
			in, _, out := encodeFixture(t, fixture)
			src := informationsElement(t, fixture+" source", in)
			got := informationsElement(t, fixture+" encoded", out)
			if !bytes.Equal(got, src) {
				i := 0
				for i < len(got) && i < len(src) && got[i] == src[i] {
					i++
				}
				t.Errorf("%s: encoded <informations> differs from the source at byte %d of %d:\nsource:  %q\nencoded: %q",
					fixture, i, len(src), window(src, i), window(got, i))
			}
			// Vacuity: the element must hold CDATA bodies and nested entries,
			// or byte identity proves nothing about either.
			if n := bytes.Count(src, []byte("<![CDATA[")); n == 0 {
				t.Errorf("%s: source <informations> has no CDATA body, so the body spelling is untested here", fixture)
			}
			if !bytes.Contains(src, []byte("]]>\n<information ")) {
				t.Errorf("%s: no entry in the source nests another, so the nested layout is untested here", fixture)
			}
			if bytes.Contains(src, []byte("]]&gt;")) {
				sawGuard = true
			}
		})
	}
	if !sawGuard {
		t.Errorf("no fixture has a lore body holding the app's ]]&gt; spelling of a typed \"]]>\"; the cdata-guard fixture should")
	}
}

// window returns up to 60 bytes of b either side of i.
func window(b []byte, i int) []byte {
	lo, hi := max(i-60, 0), min(i+60, len(b))
	return b[lo:hi]
}

// TestW2025InformationsEscapedFormReadsBack: files wxx wrote before #107
// spell every body as escaped text. Decoding one gives the same map as the
// app's CDATA spelling, so re-encoding it writes the app's spelling. The
// test rewrites a fixture's CDATA bodies as escaped text and asserts the
// encoder still writes the source's <informations> byte for byte.
func TestW2025InformationsEscapedFormReadsBack(t *testing.T) {
	const fixture = "2025-2.06-13x11-941577-populated.wxx"
	in, _, _ := encodeFixture(t, fixture)
	cdata := regexp.MustCompile(`(?s)<!\[CDATA\[(.*?)\]\]>`)
	escaped := cdata.ReplaceAllFunc(in, func(sec []byte) []byte {
		body := cdata.FindSubmatch(sec)[1]
		return []byte(strings.ReplaceAll(html.EscapeString(string(body)), "\n", "&#10;"))
	})
	if bytes.Contains(informationsElement(t, "escaped", escaped), []byte("<![CDATA[")) {
		t.Fatalf("rewrite left a CDATA section in <informations>")
	}
	m, err := xmlio.NewDecoder(xmlio.WithSkipUncompress(), xmlio.WithUTF16BEInput(false)).Decode(bytes.NewReader(escaped))
	if err != nil {
		t.Fatalf("decode escaped form: %v", err)
	}
	out, err := xmlio.MarshalXML(m, "2.06")
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if got, want := informationsElement(t, "encoded", out), informationsElement(t, "source", in); !bytes.Equal(got, want) {
		t.Errorf("escaped form re-encoded as %d bytes of <informations>, want the source's %d bytes", len(got), len(want))
	}
}

// TestW2025InformationsOffLayoutRoundTrip: a map wxx did not decode can hold
// text the app's layout does not account for -- a body with none of the
// layout's trailing newlines, or non-whitespace in the wrapper. The encoder
// writes that text so a re-decode gives it back exactly.
func TestW2025InformationsOffLayoutRoundTrip(t *testing.T) {
	_, m, _ := encodeFixture(t, "2025-2.06-13x11-941577-populated.wxx")
	infos := m.Informations
	if len(infos.Informations) < 2 || len(infos.Informations[1].Details) == 0 {
		t.Fatalf("fixture lore has no nested entry to edit")
	}
	infos.InnerText = "wrapper text & <stuff>\n"
	infos.Informations[0].InnerText = "<h1>no layout</h1>"
	infos.Informations[1].InnerText = "<h1>one newline short</h1>\n"
	infos.Informations[1].Details[0].InnerText = ""

	out, err := xmlio.MarshalXML(m, "2.06")
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	back, err := decodeRawXML(t, out, "1.1")
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	for _, c := range []struct {
		name      string
		got, want string
	}{
		{"wrapper", back.Informations.InnerText, infos.InnerText},
		{"information[1]", back.Informations.Informations[0].InnerText, infos.Informations[0].InnerText},
		{"information[2]", back.Informations.Informations[1].InnerText, infos.Informations[1].InnerText},
		{"information[2]/information[1]", back.Informations.Informations[1].Details[0].InnerText, ""},
	} {
		if c.got != c.want {
			t.Errorf("%s: InnerText came back %q, want %q", c.name, c.got, c.want)
		}
	}
}

// TestW2025CDATATerminatorRefused: a CDATA section cannot hold "]]>", and the
// app never writes one there -- its lore and note editors store a typed
// "]]>" as "]]&gt;" (testdata/2025-2.06-13x11-941577-cdata-guard.wxx). A lore
// body or notetext holding it is refused before any output, naming the entry
// (issue #107), at any depth of lore.
func TestW2025CDATATerminatorRefused(t *testing.T) {
	for _, c := range []struct {
		name  string
		edit  func(m *wxx.Map_t)
		where string
	}{
		{"lore body, third level", func(m *wxx.Map_t) {
			god := m.Informations.Informations[3].Details[1].Details[0]
			god.InnerText = "<p>a]]>b</p>"
		}, "map/informations/information[4]/information[2]/information[1]"},
		{"notetext", func(m *wxx.Map_t) {
			m.Notes[0].NoteText = "<p>a]]>b</p>"
		}, "map/notes/note[1]"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, m, _ := encodeFixture(t, "2025-2.06-13x11-941577-populated.wxx")
			c.edit(m)
			out, err := xmlio.MarshalXML(m, "2.06")
			if !errors.Is(err, wxx.ErrCDATATerminator) {
				t.Fatalf("MarshalXML: err = %v, want errors.Is(err, %v)", err, wxx.ErrCDATATerminator)
			}
			if !strings.Contains(err.Error(), c.where+" ") {
				t.Errorf("MarshalXML: err = %q, want it to name %s", err, c.where)
			}
			if len(out) != 0 {
				t.Errorf("MarshalXML: returned %d bytes, want 0", len(out))
			}
			var buf bytes.Buffer
			if err := xmlio.NewEncoder("2.06").Encode(&buf, m); !errors.Is(err, wxx.ErrCDATATerminator) {
				t.Fatalf("Encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrCDATATerminator)
			}
			if buf.Len() != 0 {
				t.Errorf("Encode: wrote %d bytes, want 0", buf.Len())
			}
		})
	}
}
