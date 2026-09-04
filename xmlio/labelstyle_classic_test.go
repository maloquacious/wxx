// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// elementLines returns every line of a UTF-8 XML document that opens the named
// element, with trailing whitespace trimmed.
//
// Line-oriented rather than parsed, deliberately: this is the one comparison in
// the suite that is about BYTES -- attribute order, spacing and value spelling
// -- and parsing would normalize away exactly what it is asserting. The classic
// encoder writes one element per line, and so does every classic file we have.
func elementLines(doc []byte, open string) []string {
	var out []string
	for _, line := range strings.Split(string(doc), "\n") {
		if strings.Contains(line, open) {
			out = append(out, strings.TrimRight(line, " \t\r"))
		}
	}
	return out
}

// TestClassicLabelStyleBytes asserts that the classic encoder re-emits every
// <labelstyle> byte-for-byte as the source file wrote it (issue #36).
//
// encodeLabelStyle was a commented-out no-op, so <text-config> went out with no
// children and all ten label styles were silently dropped from every classic
// encode -- the round-trip audit reported `element-dropped
// map/configuration/text-config/labelstyle` on all seven encodable fixtures, and
// the documented expectation carried that line as accepted loss.
//
// The audit compares element and attribute SETS, which is the right instrument
// for a loss inventory and is too weak here. It would pass on an element that
// re-spelled backgroundColor="null" as "0.0,0.0,0.0,1.0" only if the value set
// happened to match, and it says nothing about attribute order or the source's
// double spaces. This test demands the bytes, which is what "the encoder writes
// what Worldographer writes" actually means, and it is what caught the nullable
// colour needing rgbans rather than rgbas.
//
// It is scoped to the classic codec on purpose. v1_06 does NOT pass this test --
// it re-spells backgroundColor="null" as black-opaque, which is issue #62 --
// and asserting it here would fail for a reason this ticket did not cause.
func TestClassicLabelStyleBytes(t *testing.T) {
	// Every classic fixture that can be re-encoded. The ROWS fixture is absent
	// because classic ROWS encode is refused up front (issue #20), so there is
	// no output to compare.
	for _, fixture := range []string{
		"blank-2017-1.73-1.0.wxx",
		"blank-2017-1.74-1.0.wxx",
		"blank-2017-1.77-1.0.wxx",
		"2017-1.77-1.0-columns-blank.wxx",
		"2017-1.77-1.0-import.wxx",
		"2017-1.77-1.0-merge-01.wxx",
		"2017-1.77-1.0-merge-02.wxx",
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

			// The round trip writes the version the file states: a client reading
			// provenance and choosing it as the target (issue #45).
			var ed xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(m.MetaData.Version.App.Raw, xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode %s: %v", fixture, err)
			}

			in := elementLines(dd.Converted, "<labelstyle")
			out := elementLines(ed.Utf8Encoded, "<labelstyle")

			// A fixture with no label styles would make every assertion below
			// vacuously true, so the count is asserted first and against a
			// non-zero floor.
			if len(in) == 0 {
				t.Fatalf("%s: source carries no <labelstyle>; this fixture cannot evidence anything", fixture)
			}
			if len(out) != len(in) {
				t.Fatalf("%s: wrote %d <labelstyle> element(s), source has %d", fixture, len(out), len(in))
			}
			for i := range in {
				if out[i] != in[i] {
					t.Errorf("%s: labelstyle %d differs\n in : %s\n out: %s", fixture, i, in[i], out[i])
				}
			}
		})
	}
}
