// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// Issue #152 made the decoder's transport stages -- the raw input, gunzip, and
// UTF-16 to UTF-8 -- a chain of readers instead of reading each one whole
// before the next began. In a chain, an error from an early stage surfaces from
// the read at the end of it, so these tests pin that each failure is still
// reported as the stage that failed, and not as the UTF-16 conversion that
// happened to be reading. They run with and without diagnostics, because
// diagnostics read the raw input whole and so take a different path into the
// chain.

// failingReader returns its data, then err.
type failingReader struct {
	data *bytes.Reader
	err  error
}

func (f *failingReader) Read(p []byte) (int, error) {
	if f.data.Len() == 0 {
		return 0, f.err
	}
	return f.data.Read(p)
}

func gzipped(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDecode_TransportErrorsNameTheirStage(t *testing.T) {
	fixture, err := os.ReadFile(sample2025_208Resources)
	if err != nil {
		t.Fatal(err)
	}
	// a gzip trailer is the CRC-32 then the length, so this breaks the CRC
	badCRC := bytes.Clone(fixture)
	badCRC[len(badCRC)-8] ^= 0xff
	readFailure := errors.New("disk on fire")

	for _, tc := range []struct {
		name    string
		input   func() io.Reader
		want    error
		notWant []error
	}{
		{"empty input", func() io.Reader { return bytes.NewReader(nil) }, wxx.ErrNotCompressed, nil},
		{"not gzip", func() io.Reader { return strings.NewReader("<?xml version") }, wxx.ErrNotCompressed, nil},
		{"truncated gzip", func() io.Reader { return bytes.NewReader(fixture[:len(fixture)/2]) },
			wxx.ErrGUnZipFailed, []error{wxx.ErrInvalidUTF16}},
		{"bad gzip checksum", func() io.Reader { return bytes.NewReader(badCRC) },
			wxx.ErrGUnZipFailed, []error{wxx.ErrInvalidUTF16}},
		{"raw read fails mid-stream", func() io.Reader {
			return &failingReader{data: bytes.NewReader(fixture[:len(fixture)/2]), err: readFailure}
		}, wxx.ErrRawReadFailed, []error{wxx.ErrInvalidUTF16}},
		{"raw read fails at once", func() io.Reader {
			return &failingReader{data: bytes.NewReader(nil), err: readFailure}
		}, wxx.ErrRawReadFailed, []error{wxx.ErrNotCompressed}},
		{"empty gunzip", func() io.Reader { return bytes.NewReader(gzipped(t, nil)) }, wxx.ErrMissingBOM, nil},
		{"no BOM", func() io.Reader { return bytes.NewReader(gzipped(t, []byte("<?xml version"))) }, wxx.ErrMissingBOM, nil},
		{"little-endian BOM", func() io.Reader { return bytes.NewReader(gzipped(t, []byte{0xff, 0xfe, '<', 0})) },
			wxx.ErrNotBigEndianUTF16Encoded, nil},
	} {
		for _, withDiag := range []bool{false, true} {
			name := tc.name
			if withDiag {
				name += " with diagnostics"
			}
			t.Run(name, func(t *testing.T) {
				var opts []xmlio.DecoderOption
				if withDiag {
					opts = append(opts, xmlio.WithDecoderDiagnostics(&xmlio.DecoderDiagnostics{}))
				}
				_, err := xmlio.NewDecoder(opts...).Decode(tc.input())
				if !errors.Is(err, tc.want) {
					t.Fatalf("Decode error = %v, want %v", err, tc.want)
				}
				for _, nw := range tc.notWant {
					if errors.Is(err, nw) {
						t.Errorf("Decode error = %v, which also claims %v: the failure is in an earlier stage", err, nw)
					}
				}
				if errors.Is(err, wxx.ErrRawReadFailed) && !errors.Is(err, readFailure) {
					t.Errorf("Decode error = %v, want it to carry the reader's own error", err)
				}
			})
		}
	}
}

// Diagnostics still hold every stage whole: Raw is the input and Uncompressed
// is all of it gunzipped, though the decode itself no longer holds them.
func TestDecoderDiagnostics_TransportStagesWhole(t *testing.T) {
	raw, err := os.ReadFile(sample2025_208Resources)
	if err != nil {
		t.Fatal(err)
	}
	gzr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	uncompressed, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatal(err)
	}

	var diag xmlio.DecoderDiagnostics
	if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&diag)).Decode(bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(diag.Raw, raw) {
		t.Errorf("Raw is %d bytes, want the %d bytes of the input", len(diag.Raw), len(raw))
	}
	if !bytes.Equal(diag.Uncompressed, uncompressed) {
		t.Errorf("Uncompressed is %d bytes, want the %d bytes of the gunzipped input", len(diag.Uncompressed), len(uncompressed))
	}
}

// Issue #152 replaced strings.Split in decodeTiles with a walk over each line's
// fields. The field-count check and the error messages must not change; the
// brick message prints every field, so it pins that the walk saw them all.
func TestDecode_TileLineErrors(t *testing.T) {
	var diag xmlio.DecoderDiagnostics
	f, err := os.Open(sample2025_208Resources)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&diag)).Decode(f); err != nil {
		t.Fatal(err)
	}
	const line = "<tilerow>\n0\t1000\t0\t0\t58\t3\t40\t0\t7\t1\t2\n"
	if !bytes.Contains(diag.Converted, []byte(line)) {
		t.Fatalf("%s: the first tile line is not %q", sample2025_208Resources, line)
	}

	for _, tc := range []struct {
		name string
		line string
		want string
	}{
		{"5 fields", "0\t1000\t0\t0\t58", "values: expected 6/7/11/12, got 5"},
		{"8 fields", "0\t1000\t0\t0\t58\t3\t40\t0", "values: expected 6/7/11/12, got 8"},
		{"13 fields", "0\t1000\t0\t0\t58\t3\t40\t0\t7\t1\t2\t0\t0", "values: expected 6/7/11/12, got 13"},
		{"20 fields", strings.Repeat("0\t", 19) + "0", "values: expected 6/7/11/12, got 20"},
		{"bad brick", "0\t1000\t0\t0\t58\tx\t40\t0\t7\t1\t2",
			`value: brick: ["0" "1000" "0" "0" "58" "x" "40" "0" "7" "1" "2"]: strconv.Atoi: parsing "x": invalid syntax`},
		{"bad sentinel", "0\t1000\t0\t0\t58\tY", "value: sentinel: invalid value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doctored := bytes.Replace(diag.Converted, []byte(line), []byte("<tilerow>\n"+tc.line+"\n"), 1)
			_, err := xmlio.NewDecoder(xmlio.WithSkipUncompress(), xmlio.WithUTF16BEInput(false)).Decode(bytes.NewReader(doctored))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Decode error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}
