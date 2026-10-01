// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// integralSpelling reports whether a value is a number written without a decimal
// point or an exponent -- the spelling Java's Integer.parseInt accepts.
//
// "50" and "-1" qualify. "50.0" does not, and neither does "1e2": both are the
// same VALUE and neither is a thing Integer.parseInt will read.
func integralSpelling(v string) bool {
	if strings.ContainsAny(v, ".eE") {
		return false
	}
	_, err := strconv.Atoi(v)
	return err == nil
}

// integerAttribute reports whether a set of observed values identifies an
// attribute this schema states as an integer.
//
// The rule is: every value the SOURCE ever gives the attribute is spelled
// integrally, and there is at least one. That is deliberately conservative. One
// decimal spelling anywhere disqualifies the attribute, so a float attribute
// whose sampled values happen to be whole -- "0.0" would disqualify it, but a
// document stating "0" for a genuine float would not -- can still be
// misclassified. The consequence of a false positive is a test that demands the
// integer spelling for something that tolerates both, which is a failure a
// person reads and settles against the application; the consequence of a false
// negative is issue #64 again. The asymmetry justifies the direction.
func integerAttribute(values map[string]bool) bool {
	if len(values) == 0 {
		return false
	}
	for v := range values {
		if !integralSpelling(v) {
			return false
		}
	}
	return true
}

// rawAttrValues aggregates a document as element path -> attribute -> the set of
// values it is spelled with, VERBATIM.
//
// It exists because the round-trip audit harness (xmlAggregate, since removed
// with the classic tests in issue #103) could not be used here, and the reason
// is worth keeping: it ran every value through a normalizer that reformatted
// anything parsing as a float so that "0" and "0.0" compared equal. That was
// exactly right for the loss inventory it served -- a re-spelled number is not
// lost data -- and it made it blind to the only thing this test is about. The
// first draft of this audit did use xmlAggregate and passed while the encoder
// emitted "-1.0"; it was caught by deliberately breaking the encoder and
// watching the test not fail.
func rawAttrValues(t *testing.T, label string, doc []byte) map[string]map[string]map[string]bool {
	t.Helper()
	out := map[string]map[string]map[string]bool{}

	dec := xml.NewDecoder(bytes.NewReader(stripXMLDecl(doc)))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity
	dec.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }

	var stack []string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("%s: tokenize: %v", label, err)
		}
		switch tv := tok.(type) {
		case xml.StartElement:
			stack = append(stack, tv.Name.Local)
			path := strings.Join(stack, "/")
			if out[path] == nil {
				out[path] = map[string]map[string]bool{}
			}
			for _, a := range tv.Attr {
				if a.Name.Local == "xmlns" || a.Name.Space == "xmlns" {
					continue
				}
				if out[path][a.Name.Local] == nil {
					out[path][a.Name.Local] = map[string]bool{}
				}
				out[path][a.Name.Local][a.Value] = true
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return out
}

// auditIntegerSpelling checks that every attribute the SOURCE spells integrally
// is emitted integrally, and returns how many attributes it checked.
func auditIntegerSpelling(t *testing.T, label string, source, output []byte) int {
	t.Helper()

	in := rawAttrValues(t, label+" (source)", source)
	out := rawAttrValues(t, label+" (output)", output)

	audited := 0
	for path, inAttrs := range in {
		outAttrs := out[path]
		if outAttrs == nil {
			continue // element not emitted; that is a coverage question, not this test's
		}
		for attr, inVals := range inAttrs {
			if !integerAttribute(inVals) {
				continue
			}
			outVals, ok := outAttrs[attr]
			if !ok {
				continue // attribute not emitted; likewise not this test's question
			}
			audited++
			for v := range outVals {
				if integralSpelling(v) {
					continue
				}
				t.Errorf("%s: %s/@%s emitted as %q; the source always spells it as an integer, and Worldographer reads it with Integer.parseInt -- a decimal point here is a refusal to open the file (issue #64)",
					label, path, attr, v)
			}
		}
	}
	return audited
}

// TestW2025IntegerAttributeSpelling is the guard that generalizes issue #64.
//
// #64 was one attribute -- map/mapkey/@height, emitted "-1.0" where the format
// says "-1" -- and it made every W2025 file this codec wrote impossible to open.
// Fixing that one attribute is not what stops it happening again: the schema
// states 22 integer attributes, 15 of which were already correct because their
// Map_t field happens to be an int, and nothing prevented the next one from
// being added with a fresh floats() call. That is exactly how #64 arrived, in
// b1efddf, alongside a change that was otherwise an improvement.
//
// So the guard is not a list of attributes to keep in step with the code. THE
// FIXTURES ARE THE LIST: any attribute a real Worldographer document always
// spells without a decimal point must come back without one. That covers all 22
// today, covers attributes nobody has thought of yet, and needs no maintenance
// when a new one is modeled. It fails on the day a regression lands rather than
// on the day someone runs an ad-hoc diff.
//
// It is a spelling test and not a value test. A value comparison treats "0" and
// "0.0" as the same attribute; this test is the one place that cares which of
// the two reached the file, because the application does.
func TestW2025IntegerAttributeSpelling(t *testing.T) {
	total := 0

	// Every fixture goes through the full public pipeline. The 2.07
	// notes-shapes map carries elements the blank samples do not -- features,
	// labels, shapes, notes -- so it is the only source for several integer
	// attributes. requiredIntegerAttrs names two of them; if the fixture ever
	// stops stating them as integers, the audit no longer covers them and the
	// test says so.
	requiredIntegerAttrs := map[string][]string{
		"2025-2.07-13x11-941577-notes-shapes.wxx": {
			"map/features/feature@labelDistance",
			"map/shapes/shape@bbIterations",
		},
	}
	fixtures := []string{
		"2025-2.06-13x11-941577-blank.wxx",
		"2025-2.06-13x11-941577-layers-beta.wxx",
		"2025-2.07-13x11-941577-blank.wxx",
		"2025-2.07-13x11-941577-layers.wxx",
		"2025-2.07-13x11-941577-notes-shapes.wxx",
		"2025-2.07-13x11-941577-resources.wxx",
		"2025-2.07-13x11-941577-rows.wxx",
		"2025-2.08-13x11-941577-blank.wxx",
		"2025-2.08-13x11-941577-layers.wxx",
		"2025-2.08-13x11-941577-notes-shapes.wxx",
		"2025-2.08-13x11-941577-populated.wxx",
		"2025-2.08-13x11-941577-resources.wxx",
		"2025-2.08-13x11-941577-rows.wxx",
	}
	for _, fixture := range fixtures {
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
			for _, required := range requiredIntegerAttrs[fixture] {
				elem, attr, _ := strings.Cut(required, "@")
				vals := rawAttrValues(t, fixture+" (source)", dd.Converted)[elem][attr]
				if len(vals) == 0 || !integerAttribute(vals) {
					t.Fatalf("%s: source does not state %s as an integer, so the audit does not cover it", fixture, required)
				}
			}
			var ed xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(sameVersionTarget(t, fixture), xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode %s: %v", fixture, err)
			}
			total += auditIntegerSpelling(t, fixture, dd.Converted, ed.Utf8Encoded)
		})
	}

	// A spelling audit that audited nothing would pass in silence, which is the
	// failure mode of every fixture-driven test. The floor is deliberately well
	// below the 22 attributes the survey found, so that removing a fixture is
	// noticed without the number becoming something to update after every change.
	const floor = 15
	if total < floor {
		t.Errorf("audited only %d integer attribute(s) across all fixtures, want at least %d -- the audit is not covering what it claims to", total, floor)
	}
	t.Logf("audited %d integer attribute occurrence(s) across %d document(s)", total, len(fixtures))
}

// TestW2025NonIntegralValueRefused covers the state the fix leaves open by
// design (issue #64).
//
// Map_t models these attributes as float64 and always will: it is the superset
// of every supported format and does not know this schema's types. So nothing
// stops a caller assigning 50.5 to an attribute the file must state as an
// integer, and the encoder has three choices -- round it, truncate it, or refuse.
// The first two put a value on disk the caller never asked for. It refuses,
// before writing a byte, and the writer gets nothing rather than a file whose
// map key is silently a different size.
func TestW2025NonIntegralValueRefused(t *testing.T) {
	for _, tc := range []struct {
		name    string
		break_  func(*wxx.Map_t)
		wantMsg string
	}{
		{"mapkey height", func(m *wxx.Map_t) { m.MapKey.Height = 50.5 }, "map/mapkey/@height"},
		{"mapkey titleScale", func(m *wxx.Map_t) { m.MapKey.TitleScale = 79.5 }, "map/mapkey/@titleScale"},
		{
			"labelstyle dropShadowRadius",
			func(m *wxx.Map_t) {
				m.Configuration.TextConfig.LabelStyles = []*wxx.LabelStyle_t{
					{Name: "Nation", DropShadowColor: "null", DropShadowRadius: 1.5},
				}
			},
			"labelstyle/@dropShadowRadius",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newRowsMap()
			tc.break_(m)

			var buf bytes.Buffer
			err := xmlio.NewEncoder("2.06").Encode(&buf, m)
			if err == nil {
				t.Fatalf("Encode: want an error, got nil")
			}
			if !errors.Is(err, wxx.ErrInvalidIntegerAttribute) {
				t.Errorf("Encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidIntegerAttribute)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("Encode: err = %q, want it to name %q", err.Error(), tc.wantMsg)
			}
			if buf.Len() != 0 {
				t.Errorf("Encode: wrote %d bytes to w, want 0", buf.Len())
			}
		})
	}
}

// TestW2025DecodeRejectsDecimalInIntegerAttribute pins the decode half, which is
// strict by decision rather than by accident (issue #64).
//
// This codec wrote "-1.0" into map/mapkey/@height between b1efddf (2026-07-13)
// and the fix. An earlier draft accepted that spelling so such a file would
// repair itself on the next decode-encode cycle; the maintainer ruled the
// regression window irrelevant, since those files have no consumers, so the
// repair path bought nothing. What is left is the honest reading: a decimal
// point here is a file Worldographer itself refuses to open, and accepting it
// would mean carrying the value to a write that has to refuse it anyway.
//
// The document is doctored from a real fixture rather than hand-written, so the
// ONLY thing wrong with it is the spelling under test.
func TestW2025DecodeRejectsDecimalInIntegerAttribute(t *testing.T) {
	f, err := os.Open("../testdata/2025-2.06-13x11-941577-blank.wxx")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	var dd xmlio.DecoderDiagnostics
	if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// The codec is handed the document without its <?xml?> declaration, as the
	// public decoder hands it over -- Go's parser refuses version="1.1", which
	// every W2025 file states.
	source := stripXMLDecl(dd.Converted)

	const good, bad = `height="-1"`, `height="-1.0"`
	if !bytes.Contains(source, []byte(good)) {
		t.Fatalf("fixture does not state %s; this test is doctoring the wrong attribute", good)
	}
	doctored := bytes.Replace(source, []byte(good), []byte(bad), 1)

	_, _, err = v1_06.Decode(doctored)
	if err == nil {
		t.Fatalf("v1_06.Decode: want an error for %s, got nil", bad)
	}
	if !errors.Is(err, wxx.ErrInvalidIntegerAttribute) {
		t.Errorf("v1_06.Decode: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidIntegerAttribute)
	}
	if got := err.Error(); !strings.Contains(got, "height") || !strings.Contains(got, "-1.0") {
		t.Errorf("v1_06.Decode: err = %q, want it to name the attribute and the value", got)
	}
}
