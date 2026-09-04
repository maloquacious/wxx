// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"compress/gzip"
	"os"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Issue #44: DecoderDiagnostics.Schema was named for a schema but held a codec
// name, and the names it held -- "h2017v1" and "h2025v1" -- pointed at packages
// #41 had already deleted, carrying the family-year coinage ADR 0004 removed
// from the model. The two questions are now answered by two fields: Codec says
// which codec package decoded the file, Schema says what the file itself
// declared in map/@schema.
//
// TestDecoderDiagnostics_CodecAndSchema pins both, and pins that neither is
// answering the other's question.
func TestDecoderDiagnostics_CodecAndSchema(t *testing.T) {
	for _, tc := range []struct {
		name       string
		path       string
		wantCodec  string
		wantSchema string
	}{
		// A W2025 file states a schema, so both fields carry a value.
		{"w2025 2.06 blank", "../testdata/2025-2.06-13x11-941577-blank.wxx", "v1_06", "1.06"},
		{"w2025 2.06 layers", "../testdata/2025-2.06-13x11-941577-layers.wxx", "v1_06", "1.06"},
		// A classic file states no schema at all. Empty is the honest answer
		// here, not a missing one -- the file really does declare nothing.
		{"classic 1.77 blank", "../testdata/blank-2017-1.77-1.0.wxx", "v0_77", ""},
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

			if diag.Codec != tc.wantCodec {
				t.Errorf("Codec = %q, want %q", diag.Codec, tc.wantCodec)
			}
			if diag.Schema != tc.wantSchema {
				t.Errorf("Schema = %q, want %q", diag.Schema, tc.wantSchema)
			}

			// The codec field must name a package that exists. The old values
			// named deleted ones, which is what made them dangling references
			// rather than merely odd labels.
			if diag.Codec != "v0_77" && diag.Codec != "v1_06" {
				t.Errorf("Codec = %q, want a real codec package name", diag.Codec)
			}
			// Guard the specific regression: the retired coinage must not come
			// back in either field.
			for _, field := range []struct{ name, value string }{
				{"Codec", diag.Codec},
				{"Schema", diag.Schema},
			} {
				for _, dead := range []string{"h2017", "h2025"} {
					if strings.Contains(field.value, dead) {
						t.Errorf("%s = %q, which carries the retired %q coinage (ADR 0004, #44)",
							field.name, field.value, dead)
					}
				}
			}
			// Schema reports what the file stated, so it must never hold a
			// codec name -- the exact conflation #44 is about.
			if diag.Schema == "v0_77" || diag.Schema == "v1_06" {
				t.Errorf("Schema = %q, which is a codec name; Schema reports map/@schema", diag.Schema)
			}
		})
	}
}

// TestDecoderDiagnostics_SchemaSurvivesUnsupported pins the placement decision:
// Schema is assigned from the parsed metadata BEFORE dispatch, so a file that
// no codec accepts still reports what it claimed to be. That is the case where
// a diagnostic is most useful, and it is the one an assignment inside the
// dispatch arms would leave empty.
func TestDecoderDiagnostics_SchemaSurvivesUnsupported(t *testing.T) {
	// A well-formed document that no codec accepts: release "9999" matches no
	// dispatch arm.
	const doc = `<?xml version='1.1' encoding='utf-16'?>` + "\n" +
		`<map type="WORLD" release="9999" version="9.99" schema="7.77"></map>`

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	utf16BE := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	encoded, _, err := transform.Bytes(utf16BE.NewEncoder(), []byte(doc))
	if err != nil {
		t.Fatalf("encode utf-16: %v", err)
	}
	if _, err := zw.Write(encoded); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	var diag xmlio.DecoderDiagnostics
	_, err = xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&diag)).Decode(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("decode succeeded on an unsupported release; the fixture no longer tests what it means to")
	}

	// The file said 7.77, so that is what the diagnostic reports.
	if diag.Schema != "7.77" {
		t.Errorf("Schema = %q, want %q -- the stated schema must be reported even when dispatch rejects the file",
			diag.Schema, "7.77")
	}
	// No codec ran, so Codec has nothing honest to say and must stay empty.
	if diag.Codec != "" {
		t.Errorf("Codec = %q, want empty -- no codec decoded this file", diag.Codec)
	}
}
