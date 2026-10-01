// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// TestGzipLevel pins WithGzipLevel (issue #141): the default is level 6 and
// writes the bytes the encoder wrote before the option existed, and every
// level writes the same document, only compressed differently.
func TestGzipLevel(t *testing.T) {
	m, err := xmlio.ReadFile(sample2025_208Blank)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	encode := func(t *testing.T, opts ...xmlio.EncoderOption) []byte {
		t.Helper()
		var buf bytes.Buffer
		if err := xmlio.NewEncoder("2.08", opts...).Encode(&buf, m); err != nil {
			t.Fatalf("encode: %v", err)
		}
		return buf.Bytes()
	}
	gunzip := func(t *testing.T, data []byte) []byte {
		t.Helper()
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("gzip reader: %v", err)
		}
		out, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("gunzip: %v", err)
		}
		return out
	}

	byDefault := encode(t)
	if got := encode(t, xmlio.WithGzipLevel(xmlio.DefaultGzipLevel)); !bytes.Equal(got, byDefault) {
		t.Errorf("WithGzipLevel(DefaultGzipLevel) differs from the default")
	}
	// The default is the level the encoder used before the option existed,
	// compress/gzip's DefaultCompression.
	var legacy bytes.Buffer
	gz := gzip.NewWriter(&legacy)
	if _, err := gz.Write(gunzip(t, byDefault)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(byDefault, legacy.Bytes()) {
		t.Errorf("the default level writes different bytes from gzip.DefaultCompression")
	}

	document := gunzip(t, byDefault)
	for _, level := range []int{gzip.NoCompression, gzip.BestSpeed, gzip.BestCompression} {
		data := encode(t, xmlio.WithGzipLevel(level))
		if !bytes.Equal(gunzip(t, data), document) {
			t.Errorf("level %d: the document differs from the default level's", level)
		}
		if _, err := xmlio.NewDecoder().Decode(bytes.NewReader(data)); err != nil {
			t.Errorf("level %d: read back: %v", level, err)
		}
	}

	for _, level := range []int{gzip.HuffmanOnly, gzip.DefaultCompression, 10} {
		var buf bytes.Buffer
		err := xmlio.NewEncoder("2.08", xmlio.WithGzipLevel(level)).Encode(&buf, m)
		if !errors.Is(err, wxx.ErrGZipFailed) {
			t.Errorf("level %d: error = %v, want %v", level, err, wxx.ErrGZipFailed)
		}
		if buf.Len() != 0 {
			t.Errorf("level %d: wrote %d bytes with the error", level, buf.Len())
		}
	}

	// The level is a gzip setting: with gzip off it is not consulted.
	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.08", xmlio.WithGzipOutput(false), xmlio.WithGzipLevel(10)).Encode(&buf, m); err != nil {
		t.Errorf("gzip off, level 10: %v", err)
	}
}
