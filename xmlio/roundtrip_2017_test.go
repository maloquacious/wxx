// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// This file is the first automated h2017 (classic) codec test. It is an AUDIT
// harness, not a fidelity check: for every classic fixture it decodes the file,
// re-encodes it, and diffs the ORIGINAL on-disk UTF-8 XML against the
// re-encoded UTF-8 XML at the element/attribute-set level. The point is to
// inventory exactly what the frozen classic codec drops or alters on a round
// trip -- losses that a Map_t-level comparison is structurally blind to, since
// decode and encode ignore the same fields symmetrically.
//
// The per-fixture loss set is asserted against a documented expectation (see
// classicRoundTripExpect below), mirroring how the h2025 coverage-matrix test
// asserts its matrix: any drift (a newly dropped/altered field, or a
// previously dropped field that starts surviving) trips the test so a
// maintainer must update the inventory in xmlio/h2017v1/COVERAGE.md
// deliberately. Run with `-v` to dump the full per-fixture loss set; the
// harness doubles as the report generator for that document.

const classicInputDir = "../testdata/input/"

// rowsFixture decodes but cannot be re-encoded: classic ROWS encode is a
// documented hard-error (encode.go encodeTiles asserts orientation != ROWS).
const rowsFixture = "2017-1.77-1.0-rows-blank.wxx"

// classicFixtures are the eight classic 2017 fixtures under testdata/input/.
var classicFixtures = []string{
	"blank-2017-1.73-1.0.wxx",
	"blank-2017-1.74-1.0.wxx",
	"blank-2017-1.77-1.0.wxx",
	"2017-1.77-1.0-columns-blank.wxx",
	"2017-1.77-1.0-import.wxx",
	"2017-1.77-1.0-merge-01.wxx",
	"2017-1.77-1.0-merge-02.wxx",
	rowsFixture,
}

// pathAgg aggregates everything seen at one element path across a document.
type pathAgg struct {
	count   int                        // number of occurrences of this element path
	attrs   map[string]map[string]bool // attr local-name -> set of normalized values
	hasText bool                       // any non-whitespace chardata under this path
}

// xmlAggregate tokenizes a UTF-8 XML document and returns an aggregate keyed by
// element path (local-names joined with '/', e.g. "map/features/feature"). For
// each path it records the occurrence count, the union of attribute local-names
// (with their normalized value sets), and whether the element carries text.
// Processing instructions (the <?xml ...?> header), comments, and directives
// are ignored, as is pure-whitespace chardata.
func xmlAggregate(data []byte) (map[string]*pathAgg, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity
	// The input's <?xml?> header declares encoding="utf-16", but diagnostics
	// have already converted the bytes to UTF-8. Pass the reader through
	// unchanged so the declared (stale) charset does not error the tokenizer.
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	agg := map[string]*pathAgg{}
	var stack []string
	get := func(path string) *pathAgg {
		a := agg[path]
		if a == nil {
			a = &pathAgg{attrs: map[string]map[string]bool{}}
			agg[path] = a
		}
		return a
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			path := strings.Join(stack, "/")
			a := get(path)
			a.count++
			for _, attr := range t.Attr {
				if attr.Name.Local == "xmlns" || attr.Name.Space == "xmlns" {
					continue
				}
				vs := a.attrs[attr.Name.Local]
				if vs == nil {
					vs = map[string]bool{}
					a.attrs[attr.Name.Local] = vs
				}
				vs[normVal(attr.Value)] = true
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			if strings.TrimSpace(string(t)) == "" {
				continue
			}
			get(strings.Join(stack, "/")).hasText = true
		}
	}
	return agg, nil
}

// normVal canonicalizes an attribute value so non-semantic numeric formatting
// (e.g. "0" vs "0.0", "50" vs "50.0") does not register as an alteration. Any
// value that parses as a float is reformatted canonically; everything else
// (RGBA tuples, enums, names) is compared verbatim after trimming.
func normVal(s string) string {
	s = strings.TrimSpace(s)
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return strconv.FormatFloat(f, 'g', -1, 64)
	}
	return s
}

// computeLoss returns the sorted set of things present in the INPUT aggregate
// but missing or reduced in the OUTPUT aggregate. Each entry is a canonical
// tab-delimited string "<kind>\t<path>[\t<detail>]" so loss sets compare and
// print cleanly:
//   - element-dropped   path            (output count is 0)
//   - element-reduced   path   in=X out=Y
//   - attr-dropped      path   attr     (attr on this path in input, never in output)
//   - attr-altered      path   attr     (attr present in both, value set differs)
//   - text-dropped      path            (text under this path in input, not output)
//
// A fully dropped element implies its attrs/text are gone too, so those are not
// separately listed for it.
func computeLoss(in, out map[string]*pathAgg) []string {
	var loss []string
	for path, inA := range in {
		outA := out[path]
		if outA == nil || outA.count == 0 {
			loss = append(loss, fmt.Sprintf("element-dropped\t%s", path))
			continue
		}
		if outA.count < inA.count {
			loss = append(loss, fmt.Sprintf("element-reduced\t%s\tin=%d out=%d", path, inA.count, outA.count))
		}
		for attr, inVals := range inA.attrs {
			outVals, ok := outA.attrs[attr]
			if !ok {
				loss = append(loss, fmt.Sprintf("attr-dropped\t%s\t%s", path, attr))
				continue
			}
			if !equalStringSet(inVals, outVals) {
				loss = append(loss, fmt.Sprintf("attr-altered\t%s\t%s", path, attr))
			}
		}
		if inA.hasText && !outA.hasText {
			loss = append(loss, fmt.Sprintf("text-dropped\t%s", path))
		}
	}
	sort.Strings(loss)
	return loss
}

func equalStringSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// classicRoundTrip decodes a classic fixture (capturing the input UTF-8 XML in
// diagnostics), re-encodes it (capturing the output UTF-8 XML), and returns the
// element/attribute-set loss between input and output. If encode hard-errors
// (the ROWS case), it returns a nil loss set and the encode error.
func classicRoundTrip(t *testing.T, fixture string) (loss []string, encodeErr error) {
	t.Helper()
	path := classicInputDir + fixture
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var d xmlio.DecoderDiagnostics
	m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&d)).Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", fixture, err)
	}
	if len(d.Converted) == 0 {
		t.Fatalf("decode %s: diagnostics.Converted is empty", fixture)
	}

	var e xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&e)).Encode(&buf, m); err != nil {
		return nil, err
	}
	if len(e.Utf8Encoded) == 0 {
		t.Fatalf("encode %s: diagnostics.Utf8Encoded is empty", fixture)
	}

	inAgg, err := xmlAggregate(d.Converted)
	if err != nil {
		t.Fatalf("aggregate input %s: %v", fixture, err)
	}
	outAgg, err := xmlAggregate(e.Utf8Encoded)
	if err != nil {
		t.Fatalf("aggregate output %s: %v", fixture, err)
	}
	return computeLoss(inAgg, outAgg), nil
}

// TestRoundTrip2017LossInventory is the executable inventory. For each classic
// fixture it asserts the on-disk round-trip loss set against the documented
// expectation. Run with -v to dump the full per-fixture loss set.
func TestRoundTrip2017LossInventory(t *testing.T) {
	for _, fixture := range classicFixtures {
		t.Run(fixture, func(t *testing.T) {
			loss, encErr := classicRoundTrip(t, fixture)

			if fixture == rowsFixture {
				if encErr == nil {
					t.Fatalf("%s: expected encode hard-error (classic ROWS), got nil", fixture)
				}
				t.Logf("%s: round-trip not possible -- encode hard-errors: %v", fixture, encErr)
				return
			}
			if encErr != nil {
				t.Fatalf("%s: unexpected encode error: %v", fixture, encErr)
			}

			for _, l := range loss {
				t.Logf("LOSS %s :: %s", fixture, l)
			}

			want := classicRoundTripExpect[fixture]
			assertLossSet(t, fixture, want, loss)
		})
	}
}

// TestRoundTrip2017RowsHardError is the focused ROWS subtest: the ROWS fixture
// must DECODE successfully but its re-encode must return a non-nil error
// (classic ROWS encode is intentionally unimplemented -- COVERAGE.md).
func TestRoundTrip2017RowsHardError(t *testing.T) {
	path := classicInputDir + rowsFixture
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	m, err := xmlio.NewDecoder().Decode(f)
	if err != nil {
		t.Fatalf("decode %s: want success, got %v", rowsFixture, err)
	}
	if m.Tiles == nil {
		t.Fatalf("decode %s: nil Tiles", rowsFixture)
	}
	if got := m.HexOrientation; got != "ROWS" {
		t.Fatalf("decode %s: HexOrientation = %q, want ROWS", rowsFixture, got)
	}

	var buf bytes.Buffer
	if err := xmlio.NewEncoder().Encode(&buf, m); err == nil {
		t.Fatalf("encode %s: want non-nil error (classic ROWS is a hard-error), got nil", rowsFixture)
	}
}

// assertLossSet compares the observed loss set to the expected set, failing
// with an explicit list of unexpected additions and removals so a maintainer
// can update the inventory deliberately.
func assertLossSet(t *testing.T, fixture string, want, got []string) {
	t.Helper()
	wantSet := map[string]bool{}
	for _, w := range want {
		wantSet[w] = true
	}
	gotSet := map[string]bool{}
	for _, g := range got {
		gotSet[g] = true
	}

	var added, removed []string
	for g := range gotSet {
		if !wantSet[g] {
			added = append(added, g)
		}
	}
	for w := range wantSet {
		if !gotSet[w] {
			removed = append(removed, w)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)

	if len(added) > 0 || len(removed) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%s: round-trip loss set drifted from the documented inventory.\n", fixture)
		if len(added) > 0 {
			b.WriteString("  UNEXPECTED (present now, not in inventory):\n")
			for _, a := range added {
				fmt.Fprintf(&b, "    + %s\n", a)
			}
		}
		if len(removed) > 0 {
			b.WriteString("  MISSING (in inventory, no longer observed):\n")
			for _, r := range removed {
				fmt.Fprintf(&b, "    - %s\n", r)
			}
		}
		b.WriteString("  Update classicRoundTripExpect and xmlio/h2017v1/COVERAGE.md together.")
		t.Error(b.String())
	}
}

// classicRoundTripExpect is the documented per-fixture loss inventory, derived
// from the harness's first real run. It is the machine-checkable twin of the
// "Round-trip loss inventory (executable)" section in
// xmlio/h2017v1/COVERAGE.md; keep the two in sync.
//
// Shared across every classic fixture (blank, columns, import, merge):
//   - <informations>/<information> lore tree dropped (encode.go:438-442 emits
//     an empty <informations> wrapper). The nesting depth listed per fixture is
//     the deepest <information> chain that fixture actually carries.
//   - configuration <text-config>/<labelstyle> dropped (encode.go:483-507:
//     encodeLabelStyle is a commented-out no-op; the <text-config> wrapper is
//     still emitted, empty).
//
// Observed on a subset only:
//   - map/mapkey @viewlevel altered "null" -> "WORLD": encode.go:253-258 emits
//     a hardcoded constant <mapkey> block. The samples happen to match that
//     block on every other attribute, so viewlevel is the only OBSERVED
//     alteration (the full constant-block override is latent-by-code). The
//     1.74 and columns-blank fixtures already carry viewlevel="WORLD", so they
//     show no mapkey drift at all.
var classicRoundTripExpect = map[string][]string{
	"blank-2017-1.73-1.0.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/configuration/text-config/labelstyle",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"blank-2017-1.74-1.0.wxx": {
		"element-dropped\tmap/configuration/text-config/labelstyle",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
	},
	"blank-2017-1.77-1.0.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/configuration/text-config/labelstyle",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"2017-1.77-1.0-columns-blank.wxx": {
		"element-dropped\tmap/configuration/text-config/labelstyle",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
	},
	"2017-1.77-1.0-import.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/configuration/text-config/labelstyle",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"2017-1.77-1.0-merge-01.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/configuration/text-config/labelstyle",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"2017-1.77-1.0-merge-02.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/configuration/text-config/labelstyle",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	// rowsFixture has no loss set: it decodes but re-encode hard-errors, so the
	// round trip is impossible. Asserted separately (see the ROWS handling in
	// TestRoundTrip2017LossInventory and TestRoundTrip2017RowsHardError).
}
