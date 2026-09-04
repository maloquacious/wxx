// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
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
		// Both codecs, because #51 reproduced identically on each.
		{"w2025 2.06 layers", "../testdata/2025-2.06-13x11-941577-layers.wxx"},
		{"classic 1.77 blank", "../testdata/blank-2017-1.77-1.0.wxx"},
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
