// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// Issue #143: Encode streams the header, UTF-16 and gzip stages into the
// writer instead of building each stage's whole output in memory. A stage
// whose bytes diagnostics want also copies them into a buffer, so an encode
// with diagnostics takes a different path through the stages than one
// without.
//
// TestEncode_DiagnosticsDoNotChangeOutput pins that the two paths write the
// same bytes, for every combination of stages, so asking to see the pipeline
// cannot change what it writes.
func TestEncode_DiagnosticsDoNotChangeOutput(t *testing.T) {
	m, err := xmlio.ReadFile(sample2025_208Blank)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, tc := range []struct {
		name string
		opts []xmlio.EncoderOption
	}{
		{"default", nil},
		{"gzip level 1", []xmlio.EncoderOption{xmlio.WithGzipLevel(1)}},
		{"gzip level 0", []xmlio.EncoderOption{xmlio.WithGzipLevel(0)}},
		{"no gzip", []xmlio.EncoderOption{xmlio.WithGzipOutput(false)}},
		{"no utf-16", []xmlio.EncoderOption{xmlio.WithUTF16BEOutput(false)}},
		{"no header", []xmlio.EncoderOption{xmlio.WithXMLHeader(false)}},
		{"no stages", []xmlio.EncoderOption{xmlio.WithXMLHeader(false), xmlio.WithUTF16BEOutput(false), xmlio.WithGzipOutput(false)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var plain, withDiag bytes.Buffer
			if err := xmlio.NewEncoder("2.08", tc.opts...).Encode(&plain, m); err != nil {
				t.Fatalf("encode: %v", err)
			}
			var diag xmlio.EncoderDiagnostics
			opts := append([]xmlio.EncoderOption{xmlio.WithEncoderDiagnostics(&diag)}, tc.opts...)
			if err := xmlio.NewEncoder("2.08", opts...).Encode(&withDiag, m); err != nil {
				t.Fatalf("encode with diagnostics: %v", err)
			}
			if plain.Len() == 0 {
				t.Fatal("encode wrote nothing")
			}
			if !bytes.Equal(plain.Bytes(), withDiag.Bytes()) {
				t.Errorf("with diagnostics wrote %d bytes, without wrote %d, and they differ", withDiag.Len(), plain.Len())
			}
		})
	}
}

// errWriter fails every write.
type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) { return 0, w.err }

// TestEncode_ReturnsWriterError pins that a failing writer fails the encode.
// Now that the stages stream into the writer, its errors arrive from inside
// the UTF-16 and gzip stages (issue #143), and each must still reach the
// caller.
func TestEncode_ReturnsWriterError(t *testing.T) {
	m, err := xmlio.ReadFile(sample2025_208Blank)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := errors.New("disk full")
	for _, tc := range []struct {
		name string
		opts []xmlio.EncoderOption
	}{
		{"default", nil},
		{"no gzip", []xmlio.EncoderOption{xmlio.WithGzipOutput(false)}},
		{"no stages", []xmlio.EncoderOption{xmlio.WithXMLHeader(false), xmlio.WithUTF16BEOutput(false), xmlio.WithGzipOutput(false)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := xmlio.NewEncoder("2.08", tc.opts...).Encode(errWriter{want}, m)
			if !errors.Is(err, want) {
				t.Errorf("encode returned %v, want %v", err, want)
			}
		})
	}
}
