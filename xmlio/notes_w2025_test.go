// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// notesShapesFixtures are the maps Worldographer 2.07, 2.08 and 2.06 saved
// from one recipe (#93). Each carries two notes in the spelling all three
// builds write (issue #94).
//
// They are listed in order of evidence: 2.07 is the first stable release of
// the W2025 schema, and saves from earlier builds, 2.06 included, are slightly
// suspect. The single-fixture tests below use the first entry.
var notesShapesFixtures = []string{
	"2025-2.07-13x11-941577-notes-shapes.wxx",
	"2025-2.08-13x11-941577-notes-shapes.wxx",
	"2025-2.06-13x11-941577-notes-shapes.wxx",
}

// noteElement_t is one <note> as a document states it: every attribute on the
// start tag, the <notetext> body, and every attribute on <location>, or nil
// when there is no <location>.
type noteElement_t struct {
	attrs    map[string]string
	noteText string
	location map[string]string
}

// noteElements parses every <note> inside the document's <notes> element with
// encoding/xml. Only that element is parsed, so the XML 1.1 declaration a W2025
// document opens with never reaches encoding/xml.
func noteElements(t *testing.T, label string, doc []byte) []noteElement_t {
	t.Helper()
	start := bytes.Index(doc, []byte("<notes>"))
	end := bytes.Index(doc, []byte("</notes>"))
	if start < 0 || end < start {
		t.Fatalf("%s: no <notes>...</notes> element", label)
	}
	dec := xml.NewDecoder(bytes.NewReader(doc[start : end+len("</notes>")]))
	var notes []noteElement_t
	var cur *noteElement_t
	inText := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("%s: parse <notes>: %v", label, err)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "note":
				notes = append(notes, noteElement_t{attrs: map[string]string{}})
				cur = &notes[len(notes)-1]
				for _, a := range tok.Attr {
					cur.attrs[a.Name.Local] = a.Value
				}
			case "notetext":
				inText = true
			case "location":
				if cur == nil {
					t.Fatalf("%s: <location> outside a <note>", label)
				}
				if cur.location != nil {
					t.Errorf("%s: a <note> has more than one <location>", label)
				}
				cur.location = map[string]string{}
				for _, a := range tok.Attr {
					cur.location[a.Name.Local] = a.Value
				}
			case "notes":
			default:
				t.Errorf("%s: unexpected <%s> in <notes>", label, tok.Name.Local)
			}
		case xml.EndElement:
			if tok.Name.Local == "notetext" {
				inText = false
			}
		case xml.CharData:
			if inText && cur != nil {
				cur.noteText += string(tok)
			}
		}
	}
	return notes
}

// readNotesFixture returns a fixture's UTF-8 document and its decoded map.
func readNotesFixture(t *testing.T, fixture string) ([]byte, *wxx.Map_t) {
	t.Helper()
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
	return dd.Converted, m
}

// TestW2025NotesMatchSource decodes each notes-and-shapes fixture, encodes it
// as the version it states (see sameVersionTarget), and compares every <note> element the encoder wrote with the
// source's: the start tag's attributes, the <notetext> body and the
// <location>'s attributes (issue #94). Attribute order and whitespace are not
// compared; which attributes are present, and their values, are.
//
// A Map_t round trip cannot catch this bug: before #94 the codec read an older
// spelling, so decode lost the position and encode wrote viewLevel="" x="0.0"
// y="0.0", and a second decode agreed with the first.
//
// @key is compared as a string, so this is also the proof that the key the
// encoder derives from Location is spelled the way the app spells it.
func TestW2025NotesMatchSource(t *testing.T) {
	for _, fixture := range notesShapesFixtures {
		t.Run(fixture, func(t *testing.T) {
			src, m := readNotesFixture(t, fixture)
			out, err := xmlio.MarshalXML(m, sameVersionTarget(t, fixture))
			if err != nil {
				t.Fatalf("MarshalXML: %v", err)
			}
			in := noteElements(t, fixture+" source", src)
			got := noteElements(t, fixture+" output", out)
			// vacuous-pass guard: the loop below compares nothing if the
			// fixture has no notes.
			if len(in) != 2 {
				t.Fatalf("source has %d <note>(s), want the recipe's 2", len(in))
			}
			if len(got) != len(in) {
				t.Fatalf("wrote %d <note>(s), source has %d", len(got), len(in))
			}
			for i := range in {
				compareAttrSets(t, fixture, "note", i, in[i].attrs, got[i].attrs)
				if in[i].location == nil {
					t.Fatalf("source <note> %d has no <location>", i)
				}
				if got[i].location == nil {
					t.Errorf("<note> %d: wrote no <location>", i)
				} else {
					compareAttrSets(t, fixture, "note/location", i, in[i].location, got[i].location)
				}
				if got[i].noteText != in[i].noteText {
					t.Errorf("<note> %d notetext = %q, source has %q", i, got[i].noteText, in[i].noteText)
				}
				// The source's own key must agree with its location, or the
				// app would not be writing what #94 says it writes.
				loc := in[i].location
				if want := loc["viewLevel"] + "," + loc["x"] + "," + loc["y"]; in[i].attrs["key"] != want {
					t.Errorf("source <note> %d key = %q, its location spells %q", i, in[i].attrs["key"], want)
				}
			}
		})
	}
}

// compareAttrSets reports every attribute missing, added or changed.
func compareAttrSets(t *testing.T, fixture, element string, i int, in, out map[string]string) {
	t.Helper()
	for name, want := range in {
		got, ok := out[name]
		if !ok {
			t.Errorf("%s: <%s> %d: @%s missing, source has %q", fixture, element, i, name, want)
		} else if got != want {
			t.Errorf("%s: <%s> %d: @%s = %q, source has %q", fixture, element, i, name, got, want)
		}
	}
	for name, got := range out {
		if _, ok := in[name]; !ok {
			t.Errorf("%s: <%s> %d: wrote @%s = %q, which the source does not state", fixture, element, i, name, got)
		}
	}
}

// TestW2025NotesDecodeLocation asserts where each fixture's notes decode to.
// The recipe puts one note on hex (1,3) and one on hex (12,6); the three
// builds save them in different orders, so they are matched by title text.
func TestW2025NotesDecodeLocation(t *testing.T) {
	want := map[string]wxx.NoteLocation_t{
		"Note on (1,3)":  {ViewLevel: "WORLD", X: 375, Y: 1200},
		"Note on (12,6)": {ViewLevel: "WORLD", X: 2850, Y: 1950},
	}
	for _, fixture := range notesShapesFixtures {
		t.Run(fixture, func(t *testing.T) {
			_, m := readNotesFixture(t, fixture)
			if len(m.Notes) != len(want) {
				t.Fatalf("decoded %d note(s), want %d", len(m.Notes), len(want))
			}
			seen := map[string]bool{}
			for i, note := range m.Notes {
				var name string
				for text := range want {
					if strings.Contains(note.NoteText, text) {
						name = text
					}
				}
				if name == "" {
					t.Errorf("note %d: text %q names neither recipe note", i, note.NoteText)
					continue
				}
				seen[name] = true
				if note.Location == nil {
					t.Errorf("%s: Location = nil, want %+v", name, want[name])
				} else if *note.Location != want[name] {
					t.Errorf("%s: Location = %+v, want %+v", name, *note.Location, want[name])
				}
				if note.OriginalViewLevel != "WORLD" || !note.IsWorld || !note.IsContinent || !note.IsKingdom || !note.IsProvince {
					t.Errorf("%s: originalViewLevel %q, isWorld/Continent/Kingdom/Province %v/%v/%v/%v, want WORLD and all true",
						name, note.OriginalViewLevel, note.IsWorld, note.IsContinent, note.IsKingdom, note.IsProvince)
				}
			}
			if len(seen) != len(want) {
				t.Errorf("matched %d of the %d recipe notes", len(seen), len(want))
			}
		})
	}
}

// TestW2025NoteWithoutLocationRefused: a note with no Location has no position
// to write. The encoder refuses it before writing a byte, naming the note,
// rather than putting it at 0,0 in the map's corner (issue #94).
func TestW2025NoteWithoutLocationRefused(t *testing.T) {
	_, m := readNotesFixture(t, notesShapesFixtures[0])
	if len(m.Notes) < 2 {
		t.Fatalf("fixture has %d note(s), want at least 2", len(m.Notes))
	}
	m.Notes[1].Location = nil
	m.Notes[1].Title = "the lost one"

	out, err := xmlio.MarshalXML(m, "2.06")
	if !errors.Is(err, wxx.ErrNoteWithoutLocation) {
		t.Fatalf("MarshalXML: err = %v, want errors.Is(err, %v)", err, wxx.ErrNoteWithoutLocation)
	}
	if !strings.Contains(err.Error(), "map/notes/note[2]") || !strings.Contains(err.Error(), `"the lost one"`) {
		t.Errorf("MarshalXML: err = %q, want it to name map/notes/note[2] and its title", err)
	}
	if len(out) != 0 {
		t.Errorf("MarshalXML: returned %d bytes, want 0", len(out))
	}

	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.06").Encode(&buf, m); !errors.Is(err, wxx.ErrNoteWithoutLocation) {
		t.Fatalf("Encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrNoteWithoutLocation)
	}
	if buf.Len() != 0 {
		t.Errorf("Encode: wrote %d bytes, want 0", buf.Len())
	}
}

// TestW2025NoteKeyMismatchRefused: @key states a note's position a second
// time, and the model does not keep it. A key that names a different place
// from the <location> is refused on decode, naming the note, because dropping
// it would lose one of the two positions without a word (issue #94).
func TestW2025NoteKeyMismatchRefused(t *testing.T) {
	src, _ := readNotesFixture(t, notesShapesFixtures[0])
	src = bytes.TrimLeft(stripXMLDecl(src), "\n")
	const key, moved = `key="WORLD,375.0,1200.0"`, `key="WORLD,375.0,1275.0"`
	if n := bytes.Count(src, []byte(key)); n != 1 {
		t.Fatalf("fixture states %s %d time(s), want 1", key, n)
	}
	for _, tc := range []struct {
		name, from, to string
	}{
		{"y differs", key, moved},
		{"view level differs", key, `key="CONTINENT,375.0,1200.0"`},
		{"not a position", key, `key="custom"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := bytes.Replace(src, []byte(tc.from), []byte(tc.to), 1)
			_, err := decodeRawXML(t, doc, "1.1")
			if !errors.Is(err, wxx.ErrNoteKeyMismatch) {
				t.Fatalf("decode: err = %v, want errors.Is(err, %v)", err, wxx.ErrNoteKeyMismatch)
			}
			if !strings.Contains(err.Error(), "map/notes/note[") || !strings.Contains(err.Error(), tc.to[len(`key=`):]) {
				t.Errorf("decode: err = %q, want it to name the note and its key %s", err, tc.to[len(`key=`):])
			}
		})
	}

	// The unmodified document decodes, so the refusals above are the edit's.
	if _, err := decodeRawXML(t, src, "1.1"); err != nil {
		t.Fatalf("decode unmodified: %v", err)
	}
}
