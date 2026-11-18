// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

const labels209Fixture = "2025-2.09-13x11-941577-labels.wxx"

// decodeFixture decodes a tracked fixture through xmlio and returns the map and
// the source document as UTF-8.
func decodeFixture(t *testing.T, fixture string) (*wxx.Map_t, []byte) {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "..", "testdata", fixture))
	if err != nil {
		t.Fatalf("open %s: %v", fixture, err)
	}
	defer f.Close()
	var dd xmlio.DecoderDiagnostics
	m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", fixture, err)
	}
	return m, dd.Converted
}

// attrPattern matches one name="value" pair of a start tag. Values in these
// files never contain a quote.
var attrPattern = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_]*)="([^"]*)"`)

// startTags extracts every start tag of element from doc, in document order,
// each as its ordered name/value pairs. It reads the raw text, as
// xmlio's startTagAttrs does, because attribute order and absence are what is
// being compared; inter-attribute whitespace is not data and is not kept.
func startTags(doc []byte, element string) [][][2]string {
	tag := regexp.MustCompile(`<` + regexp.QuoteMeta(element) + `(\s[^>]*)?/?>`)
	var out [][][2]string
	for _, m := range tag.FindAllSubmatch(doc, -1) {
		attrs := [][2]string{}
		for _, a := range attrPattern.FindAllStringSubmatch(string(m[1]), -1) {
			attrs = append(attrs, [2]string{a[1], a[2]})
		}
		out = append(out, attrs)
	}
	return out
}

// compareTags reports every difference between the start tags written and the
// ones wanted: count, attribute names, order and values, all exact. It is the
// standard of xmlio's TestW2025LabelStyleAttrsMatchSource with no spelling
// exemptions.
func compareTags(t *testing.T, element string, want, got [][][2]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("wrote %d <%s> element(s), want %d", len(got), element, len(want))
		return
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Errorf("<%s> %d has %d attribute(s) %v, want %d %v", element, i, len(got[i]), attrNames(got[i]), len(want[i]), attrNames(want[i]))
			continue
		}
		for j, w := range want[i] {
			g := got[i][j]
			if g[0] != w[0] {
				t.Errorf("<%s> %d attribute %d is %q, want %q (order or name drift)", element, i, j, g[0], w[0])
				continue
			}
			if g[1] != w[1] {
				t.Errorf("<%s> %d @%s = %q, want %q", element, i, w[0], g[1], w[1])
			}
		}
	}
}

func attrNames(attrs [][2]string) []string {
	var out []string
	for _, a := range attrs {
		out = append(out, a[0])
	}
	return out
}

// TestWritesLabelTextLayout pins the one rule deciding which targets write the
// 2.09 label text-layout attributes: 2.09 and every later build, none before.
func TestWritesLabelTextLayout(t *testing.T) {
	for _, tc := range []struct {
		app  string
		want bool
	}{
		{"2.06", false}, {"2.07", false}, {"2.08", false},
		{"2.09", true}, {"2.10", true}, {"3.00", true},
	} {
		got, err := v1_06.WritesLabelTextLayout(tc.app)
		if err != nil {
			t.Errorf("WritesLabelTextLayout(%q): %v", tc.app, err)
			continue
		}
		if got != tc.want {
			t.Errorf("WritesLabelTextLayout(%q) = %v, want %v", tc.app, got, tc.want)
		}
	}
	for _, app := range []string{"", "2", "garbage", "2.09.1"} {
		if got, err := v1_06.WritesLabelTextLayout(app); err == nil {
			t.Errorf("WritesLabelTextLayout(%q) = %v, nil; want an error, not a guess", app, got)
		}
	}
	// Every version this codec accepts must have an answer.
	for _, a := range v1_06.AcceptedApps().Apps {
		if _, err := v1_06.WritesLabelTextLayout(a.Version); err != nil {
			t.Errorf("WritesLabelTextLayout(accepted %q): %v", a.Version, err)
		}
	}
}

// TestW2025Labels209EncodeMatchesSource: encoded for a 2.09 target, every
// <label> and <labelstyle> start tag of the 2.09 labels fixture comes back with
// the source's attribute names, in the source's order, with the source's values
// (issue #161). That covers both positions 2.09 writes on a label (lineSpacing
// after outlineSize, textAlignment after isProvince), lineSpacing last on a
// label style, the Java double spelling "2.0", and the omitted-at-default
// cases: two labels and six styles state no lineSpacing and must gain none.
// It encodes through the public API, as any caller targeting 2.09 does (issue
// #162 registered it).
func TestW2025Labels209EncodeMatchesSource(t *testing.T) {
	m, src := decodeFixture(t, labels209Fixture)
	out, err := xmlio.MarshalXML(m, "2.09")
	if err != nil {
		t.Fatalf("encode as 2.09: %v", err)
	}
	for _, element := range []string{"label", "labelstyle"} {
		in := startTags(src, element)
		if len(in) == 0 {
			t.Fatalf("source has no <%s>; this fixture cannot evidence anything", element)
		}
		compareTags(t, element, in, startTags(out, element))
	}
	// The fixture must actually exercise every case, or the comparison above
	// proves less than it claims.
	for _, want := range []string{` lineSpacing="2.0" dropShadowColor=`, ` isProvince="true" textAlignment="RIGHT"`, ` isProvince="true" textAlignment="CENTER"`} {
		if !bytes.Contains(src, []byte(want)) {
			t.Errorf("source no longer states %q", want)
		}
	}
}

// TestW2025Labels209NotEncodedFor208: the same map written as 2.08, which this
// codec accepts, states neither attribute anywhere, though the model holds them.
func TestW2025Labels209NotEncodedFor208(t *testing.T) {
	m, _ := decodeFixture(t, labels209Fixture)
	out, err := v1_06.Encode(m, "2.08")
	if err != nil {
		t.Fatalf("encode as 2.08: %v", err)
	}
	for _, attr := range []string{" lineSpacing=", " textAlignment="} {
		if n := bytes.Count(out, []byte(attr)); n != 0 {
			t.Errorf("2.08 output states%s %d time(s); 2.08 writes it nowhere", attr, n)
		}
	}
}

// TestW2025FeatureLabel209AttrsEncode: a feature's inline <label> is written
// by the same encoder, so for a 2.09 target it gains both attributes in the
// positions 2.09 writes on a standalone label. No fixture has a 2.09 feature
// label, so the values are set on a real 2.08 save's first feature label;
// they are ones the labels fixture does not hold (a non-integral spacing, and
// JUSTIFY, kept verbatim). The other feature labels have neither in the model
// and must gain neither: absent stays absent for 2.09 too.
func TestW2025FeatureLabel209AttrsEncode(t *testing.T) {
	m, src := decodeFixture(t, "2025-2.08-13x11-941577-notes-shapes.wxx")
	var labelled []*wxx.Feature_t
	for _, f := range m.Features {
		if f.Label != nil {
			labelled = append(labelled, f)
		}
	}
	if len(labelled) < 2 {
		t.Fatalf("notes-shapes has %d labelled features, want at least 2", len(labelled))
	}
	spacing := 2.09
	labelled[0].Label.LineSpacing = &spacing
	labelled[0].Label.TextAlignment = "JUSTIFY"

	out, err := xmlio.MarshalXML(m, "2.09")
	if err != nil {
		t.Fatalf("encode as 2.09: %v", err)
	}

	// The source's <label> tags, with the two attributes spliced into the first
	// feature label where 2.09 writes them. The feature labels precede <labels>
	// in the document, so the first <label> is the first feature's.
	want := startTags(src, "label")
	if len(want) == 0 {
		t.Fatal("notes-shapes source has no <label>")
	}
	first := want[0]
	insertBefore := func(attrs [][2]string, before string, attr [2]string) [][2]string {
		i := slices.IndexFunc(attrs, func(a [2]string) bool { return a[0] == before })
		if i < 0 {
			t.Fatalf("first feature label states no %s: %v", before, attrNames(attrs))
		}
		return slices.Insert(attrs, i, attr)
	}
	first = insertBefore(first, "dropShadowColor", [2]string{"lineSpacing", "2.09"})
	first = insertBefore(first, "isGMOnly", [2]string{"textAlignment", "JUSTIFY"})
	want[0] = first
	compareTags(t, "label", want, startTags(out, "label"))
}
