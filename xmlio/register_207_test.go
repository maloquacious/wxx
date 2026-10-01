// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// w2025XMLHeader is the declaration every W2025 fixture opens with, quoting
// included, as the decoder reads it back from disk.
const w2025XMLHeader = "<?xml version='1.1' encoding='utf-16'?>\n"

// TestFixtures207AreEveryTracked207Fixture holds fixtures207 to the testdata
// directory, so a 2.07 save added later is not silently left out of the tests
// that loop over the list.
func TestFixtures207AreEveryTracked207Fixture(t *testing.T) {
	onDisk, err := filepath.Glob(filepath.Join("..", "testdata", "2025-2.07-*.wxx"))
	if err != nil {
		t.Fatalf("glob 2.07 fixtures: %v", err)
	}
	var listed []string
	for _, p := range fixtures207 {
		listed = append(listed, filepath.Base(p))
	}
	for i := range onDisk {
		onDisk[i] = filepath.Base(onDisk[i])
	}
	slices.Sort(listed)
	slices.Sort(onDisk)
	if !slices.Equal(listed, onDisk) {
		t.Errorf("fixtures207 lists %v, testdata has %v", listed, onDisk)
	}
	if len(onDisk) != 5 {
		t.Errorf("testdata has %d 2.07 fixtures, want the 5 issue #92 was written against", len(onDisk))
	}
}

// TestEncode207FixturesAs207 is issue #92's proof. Every 2.07 fixture encodes
// as "2.07" through the full public pipeline, and the file it writes states
// exactly release="2025" version="2.07" schema="1.06" under the XML 1.1
// declaration the source opens with.
//
// It also pins that 2.07 is data and not a format: the XML written as "2.07"
// is the XML written as "2.06" with the one version attribute changed. No
// difference between what the two builds write is known, so none may appear
// here (issue #92 ruling: report only proven losses).
func TestEncode207FixturesAs207(t *testing.T) {
	const app = "2.07"
	want := map[string]string{"release": "2025", "version": "2.07", "schema": "1.06"}
	checked := 0
	for _, path := range fixtures207 {
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("open %s: %v", path, err)
			}
			defer f.Close()
			var src xmlio.DecoderDiagnostics
			m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&src)).Decode(f)
			if err != nil {
				t.Fatalf("decode %s: %v", path, err)
			}
			// Guard: the source states the identity asserted below, so the test
			// is about writing 2.07 as itself.
			if got := string(src.XMLHeader); got != w2025XMLHeader {
				t.Fatalf("%s: source opens with %q, want %q", path, got, w2025XMLHeader)
			}
			if got := mapIdentity(t, path+" source", src.Converted); !maps.Equal(got, want) {
				t.Fatalf("%s: source states %v, want %v", path, got, want)
			}

			var ed xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(app, xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode %s as %q: %v", path, app, err)
			}
			if len(ed.Dropped) != 0 {
				t.Errorf("%s: encoding as %q reported a loss: %v", path, app, ed.Dropped)
			}

			// Read the written file back to its text, declaration included.
			var out xmlio.DecoderDiagnostics
			back, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&out)).Decode(&buf)
			if err != nil {
				t.Fatalf("re-decode %s written as %q: %v", path, app, err)
			}
			if got := string(out.XMLHeader); got != w2025XMLHeader {
				t.Errorf("%s written as %q opens with %q, want %q", path, app, got, w2025XMLHeader)
			}
			if got := mapIdentity(t, path+" output", out.Converted); !maps.Equal(got, want) {
				t.Errorf("%s written as %q states %v, want exactly %v", path, app, got, want)
			}
			if v := back.MetaData.Version; v.App.Raw != app || v.Schema == nil || v.Schema.Raw != "1.06" {
				t.Errorf("%s written as %q re-decodes as %v", path, app, v)
			}

			// 2.07 and 2.06 differ in the version attribute and nothing else.
			x07, err := xmlio.MarshalXML(m, app)
			if err != nil {
				t.Fatalf("MarshalXML(%s, %q): %v", path, app, err)
			}
			x06, err := xmlio.MarshalXML(m, "2.06")
			if err != nil {
				t.Fatalf("MarshalXML(%s, %q): %v", path, "2.06", err)
			}
			if !bytes.Equal(x07, ed.Utf8Encoded) {
				t.Errorf("%s: MarshalXML and Encode disagree on the XML for %q", path, app)
			}
			if bytes.Equal(x06, x07) {
				t.Fatalf("%s: the XML written as 2.06 and as 2.07 is identical, so the version attribute was not written from the target", path)
			}
			swapped := bytes.Replace(x06, []byte(` version="2.06" `), []byte(` version="2.07" `), 1)
			if !bytes.Equal(swapped, x07) {
				i := 0
				for i < len(swapped) && i < len(x07) && swapped[i] == x07[i] {
					i++
				}
				t.Errorf("%s: the XML written as 2.07 differs from the 2.06 XML beyond map/@version, at byte %d:\n2.06: %q\n2.07: %q",
					path, i, window(swapped, i), window(x07, i))
			}
			checked++
		})
	}
	if checked != 5 {
		t.Errorf("checked %d 2.07 fixture(s), want 5", checked)
	}
}

// mapIdentity returns the identity attributes the document's <map> start tag
// states. An attribute the tag does not state is absent from the result.
func mapIdentity(t *testing.T, label string, doc []byte) map[string]string {
	t.Helper()
	tags := startTagAttrs(doc, "map")
	if len(tags) != 1 {
		t.Fatalf("%s: %d <map> start tags, want 1", label, len(tags))
	}
	got := map[string]string{}
	for _, a := range tags[0] {
		switch a[0] {
		case "release", "version", "schema":
			if _, dup := got[a[0]]; dup {
				t.Fatalf("%s: <map> states @%s twice", label, a[0])
			}
			got[a[0]] = a[1]
		}
	}
	return got
}
