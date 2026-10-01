// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// The decoder splits the converted UTF-8 document into exactly two pieces: the
// XML declaration and everything after it. DecoderDiagnostics is the supported
// way to see that split (CLAUDE.md: diagnostics over debug logging), so the
// fields have to hold what they are named for.
//
// Issue #51: XMLHeader was assigned twice, and the second assignment overwrote
// the declaration with the whole document body, while XMLData was never
// assigned at all. Both fields were populated-but-wrong rather than absent,
// which is the failure mode that misleads a caller instead of merely not
// serving one. It cost a false "0 labels" reading while #35 was being worked.
//
// TestDecoderDiagnostics_HeaderAndDataSplit pins the split. The load-bearing
// assertion is the length identity: len(Converted) == len(XMLHeader) +
// len(XMLData). It is what the old code could not satisfy, because XMLHeader
// held the body (Converted minus the declaration) and XMLData held nothing.
func TestDecoderDiagnostics_HeaderAndDataSplit(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{"w2025 2.06 layers beta", "../testdata/2025-2.06-13x11-941577-layers-beta.wxx"},
		{"w2025 2.07 notes-shapes", sample2025_207NotesShapes},
		{"w2025 2.08 notes-shapes", sample2025_208NotesShapes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open(tc.path)
			if err != nil {
				t.Fatalf("open %s: %v", tc.path, err)
			}
			defer f.Close()

			var diag xmlio.DecoderDiagnostics
			if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&diag)).Decode(f); err != nil {
				t.Fatalf("decode %s: %v", tc.path, err)
			}

			// Non-zero lengths first, so nothing below can pass vacuously on
			// an empty field.
			if len(diag.Converted) == 0 {
				t.Fatal("Converted is empty")
			}
			if len(diag.XMLHeader) == 0 {
				t.Fatal("XMLHeader is empty")
			}
			if len(diag.XMLData) == 0 {
				t.Fatal("XMLData is empty")
			}

			// XMLHeader is the declaration, and only the declaration.
			header := string(diag.XMLHeader)
			if !strings.HasPrefix(header, "<?xml ") || !strings.HasSuffix(header, "?>\n") {
				t.Errorf("XMLHeader is not an XML declaration: %q", truncate(header))
			}
			// A declaration is tens of bytes. The bug put tens of thousands
			// here, so bound it well above any real declaration and well below
			// any real document.
			if len(diag.XMLHeader) > 64 {
				t.Errorf("XMLHeader = %d bytes, want an XML declaration (<= 64); it is holding the document body: %q",
					len(diag.XMLHeader), truncate(header))
			}
			if strings.Contains(header, "<map ") {
				t.Errorf("XMLHeader contains the map element, so it is not just the declaration: %q", truncate(header))
			}

			// XMLData is everything after the declaration: what the codec was
			// handed, starting at the root element.
			if !bytes.HasPrefix(diag.XMLData, []byte("<map ")) {
				t.Errorf("XMLData does not start at the map element: %q", truncate(string(diag.XMLData)))
			}
			if bytes.HasPrefix(diag.XMLData, []byte("<?xml")) {
				t.Error("XMLData still carries the XML declaration; it should have been consumed")
			}

			// The split is exhaustive and non-overlapping: the two fields
			// partition Converted. This is the assertion #51 fails.
			if got, want := len(diag.XMLHeader)+len(diag.XMLData), len(diag.Converted); got != want {
				t.Errorf("XMLHeader(%d) + XMLData(%d) = %d, want len(Converted) = %d; the fields do not partition the document",
					len(diag.XMLHeader), len(diag.XMLData), got, want)
			}
			if !bytes.Equal(append(bdupTest(diag.XMLHeader), diag.XMLData...), diag.Converted) {
				t.Error("XMLHeader concatenated with XMLData does not reproduce Converted")
			}

			// The diagnostics are copies, not aliases into the live buffer: a
			// caller mutating one must not corrupt another.
			if len(diag.XMLData) > 0 && len(diag.Converted) > 0 {
				if &diag.XMLData[0] == &diag.Converted[0] {
					t.Error("XMLData aliases Converted instead of being a copy")
				}
			}
		})
	}
}

// bdupTest copies a slice so the concatenation above cannot scribble on the
// diagnostic it is checking.
func bdupTest(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// truncate keeps a failure message readable when the field under test is
// holding far more than it should.
func truncate(s string) string {
	const max = 60
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Issue #146: the gzip stage stored its output in Utf16Encoded, overwriting
// the UTF-16 stage's output, and never set Compressed. The written file was
// right; the diagnostics were populated-but-wrong, the same failure mode as #51.
//
// TestEncoderDiagnostics_StagesChain pins each field to the stage it is named
// for by chaining them: each stage's field, undone, gives the previous stage's
// field, and the last stage's field is what was written.
func TestEncoderDiagnostics_StagesChain(t *testing.T) {
	m, err := xmlio.ReadFile(sample2025_208Blank)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var diag xmlio.EncoderDiagnostics
	var out bytes.Buffer
	if err := xmlio.NewEncoder("2.08", xmlio.WithEncoderDiagnostics(&diag)).Encode(&out, m); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Non-zero lengths first, so nothing below can pass vacuously on an empty
	// field. Compressed is the field #146 never set.
	for _, f := range []struct {
		name string
		data []byte
	}{
		{"Utf8Encoded", diag.Utf8Encoded},
		{"WithXmlHeader", diag.WithXmlHeader},
		{"Utf16Encoded", diag.Utf16Encoded},
		{"Compressed", diag.Compressed},
	} {
		if len(f.data) == 0 {
			t.Fatalf("%s is empty", f.name)
		}
	}

	// WithXmlHeader is Utf8Encoded behind the declaration.
	if !bytes.HasSuffix(diag.WithXmlHeader, diag.Utf8Encoded) || !bytes.HasPrefix(diag.WithXmlHeader, []byte("<?xml ")) {
		t.Errorf("WithXmlHeader is not the XML declaration followed by Utf8Encoded: %q", truncate(string(diag.WithXmlHeader)))
	}

	// Utf16Encoded is UTF-16BE, BOM first, and decodes to WithXmlHeader. The
	// bug left gzip bytes here, which start 1f 8b.
	if !bytes.HasPrefix(diag.Utf16Encoded, []byte{0xfe, 0xff}) {
		t.Errorf("Utf16Encoded starts % x, want the UTF-16BE BOM fe ff", diag.Utf16Encoded[:min(4, len(diag.Utf16Encoded))])
	}
	utf8, err := io.ReadAll(transform.NewReader(bytes.NewReader(diag.Utf16Encoded),
		unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewDecoder()))
	if err != nil {
		t.Fatalf("decode Utf16Encoded: %v", err)
	}
	if !bytes.Equal(utf8, diag.WithXmlHeader) {
		t.Error("Utf16Encoded does not decode to WithXmlHeader")
	}

	// Compressed gunzips to Utf16Encoded.
	r, err := gzip.NewReader(bytes.NewReader(diag.Compressed))
	if err != nil {
		t.Fatalf("Compressed is not gzip: %v", err)
	}
	gunzipped, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("gunzip Compressed: %v", err)
	}
	if !bytes.Equal(gunzipped, diag.Utf16Encoded) {
		t.Error("gunzipping Compressed does not give Utf16Encoded")
	}

	// Compressed is exactly what was written, and a copy rather than an alias.
	if !bytes.Equal(diag.Compressed, out.Bytes()) {
		t.Errorf("Compressed (%d bytes) differs from the %d bytes written", len(diag.Compressed), out.Len())
	}
	if &diag.Compressed[0] == &out.Bytes()[0] {
		t.Error("Compressed aliases the written bytes instead of being a copy")
	}
}
