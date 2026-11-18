// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// labels209Fixture is a first save in Worldographer 2025 2.09 (schema 1.06)
// with three standalone labels and the seven stock label styles, one of them
// edited (issue #161).
const labels209Fixture = "2025-2.09-13x11-941577-labels.wxx"

// TestW2025Labels209AttrsDecode: the attributes 2.09 adds to <label>
// (@lineSpacing, @textAlignment) and <labelstyle> (@lineSpacing) decode, with
// presence kept (issue #161). 2.09 omits lineSpacing at the default spacing, so
// an absent attribute must decode to nil, not to 1.0 or 0.
//
// The expected values are the fixture's bytes:
//
//	label 1  textAlignment="CENTER", no lineSpacing   ("Custom Style\nDefault")
//	label 2  lineSpacing="2.0" textAlignment="RIGHT"  ("Custom Style\nLine Spacing 2\nAlign Right")
//	label 3  textAlignment="CENTER", no lineSpacing   ("City Style\nDefault")
//	labelstyle City  lineSpacing="2.0"; the other six have none
func TestW2025Labels209AttrsDecode(t *testing.T) {
	m, err := xmlio.ReadFile(filepath.Join("..", "testdata", labels209Fixture))
	if err != nil {
		t.Fatalf("decode %s: %v", labels209Fixture, err)
	}

	two := 2.0
	wantLabels := []struct {
		text          string
		lineSpacing   *float64
		textAlignment string
	}{
		{"Custom Style\nDefault", nil, "CENTER"},
		{"Custom Style\nLine Spacing 2\nAlign Right", &two, "RIGHT"},
		{"City Style\nDefault", nil, "CENTER"},
	}
	if len(m.Labels) != len(wantLabels) {
		t.Fatalf("decoded %d labels, want %d", len(m.Labels), len(wantLabels))
	}
	for i, want := range wantLabels {
		got := m.Labels[i]
		if got.InnerText != want.text {
			t.Errorf("label %d: InnerText = %q, want %q", i+1, got.InnerText, want.text)
		}
		if msg := lineSpacingMismatch(got.LineSpacing, want.lineSpacing); msg != "" {
			t.Errorf("label %d (%q): LineSpacing %s", i+1, want.text, msg)
		}
		if got.TextAlignment != want.textAlignment {
			t.Errorf("label %d (%q): TextAlignment = %q, want %q", i+1, want.text, got.TextAlignment, want.textAlignment)
		}
	}

	wantStyles := map[string]*float64{
		"Nation":          nil,
		"Geography":       nil,
		"Village":         nil,
		"Geography Minor": nil,
		"Geography Major": nil,
		"City":            &two,
		"Province":        nil,
	}
	if m.Configuration == nil || m.Configuration.TextConfig == nil {
		t.Fatal("decoded no text-config")
	}
	styles := m.Configuration.TextConfig.LabelStyles
	if len(styles) != len(wantStyles) {
		t.Fatalf("decoded %d label styles, want %d", len(styles), len(wantStyles))
	}
	for _, style := range styles {
		want, ok := wantStyles[style.Name]
		if !ok {
			t.Errorf("unexpected label style %q", style.Name)
			continue
		}
		if msg := lineSpacingMismatch(style.LineSpacing, want); msg != "" {
			t.Errorf("labelstyle %q: LineSpacing %s", style.Name, msg)
		}
	}
}

// TestW2025FeatureLabel209AttrsDecode: a feature's inline <label> is the same
// label type, written by the same app writer, so it decodes both 2.09
// attributes too (issue #161). No fixture has a 2.09 feature label, so this
// splices the attributes, in the positions 2.09 writes them on a standalone
// label, into the first feature label of a real 2.08 save re-encoded by wxx.
// The values are deliberately ones the labels fixture does not hold: a
// non-integral spacing ("2.09", seen in a non-fixture 2.09 save) and JUSTIFY (a
// JavaFX TextAlignment no save has shown), which must be kept verbatim, not
// checked against a list.
func TestW2025FeatureLabel209AttrsDecode(t *testing.T) {
	m, err := xmlio.ReadFile(filepath.Join("..", "testdata", "2025-2.08-13x11-941577-notes-shapes.wxx"))
	if err != nil {
		t.Fatalf("decode notes-shapes: %v", err)
	}
	doc, err := xmlio.MarshalXML(m, "2.08")
	if err != nil {
		t.Fatalf("encode notes-shapes: %v", err)
	}

	// The first feature's label; the feature labels are the only <label> before
	// <labels> in the encoded document.
	start := bytes.Index(doc, []byte(`<label  mapLayer="Features"`))
	if start < 0 {
		t.Fatal(`encoded notes-shapes has no <label  mapLayer="Features"`)
	}
	end := start + bytes.IndexByte(doc[start:], '>')
	tag := doc[start : end+1]
	edited := tag
	for _, r := range []struct{ before, insert string }{
		{` dropShadowColor=`, ` lineSpacing="2.09"`},
		{` isGMOnly=`, ` textAlignment="JUSTIFY"`},
	} {
		if n := bytes.Count(edited, []byte(r.before)); n != 1 {
			t.Fatalf("first feature label states %q %d times, want 1: %s", r.before, n, tag)
		}
		edited = bytes.Replace(edited, []byte(r.before), []byte(r.insert+r.before), 1)
	}
	src := append(append(append([]byte{}, doc[:start]...), edited...), doc[end+1:]...)

	got, err := decodeXML(src)
	if err != nil {
		t.Fatalf("decode a 2.09-style feature label: %v", err)
	}
	var labelled []*wxx.Feature_t
	for _, f := range got.Features {
		if f.Label != nil {
			labelled = append(labelled, f)
		}
	}
	if len(labelled) < 2 {
		t.Fatalf("decoded %d labelled features, want at least 2", len(labelled))
	}
	want := 2.09
	if msg := lineSpacingMismatch(labelled[0].Label.LineSpacing, &want); msg != "" {
		t.Errorf("feature %s label: LineSpacing %s", labelled[0].Uuid, msg)
	}
	if got := labelled[0].Label.TextAlignment; got != "JUSTIFY" {
		t.Errorf("feature %s label: TextAlignment = %q, want %q", labelled[0].Uuid, got, "JUSTIFY")
	}
	// Control: the labels not edited still decode as absent.
	for _, f := range labelled[1:] {
		if msg := lineSpacingMismatch(f.Label.LineSpacing, nil); msg != "" {
			t.Errorf("feature %s label (not edited): LineSpacing %s", f.Uuid, msg)
		}
		if f.Label.TextAlignment != "" {
			t.Errorf("feature %s label (not edited): TextAlignment = %q, want absent", f.Uuid, f.Label.TextAlignment)
		}
	}
}

// lineSpacingMismatch describes how got differs from want, nil meaning the
// attribute is absent, or returns "" when they agree.
func lineSpacingMismatch(got, want *float64) string {
	switch {
	case got == nil && want == nil:
		return ""
	case got == nil:
		return "= nil (absent), want " + floatString(*want)
	case want == nil:
		return "= " + floatString(*got) + ", want nil (absent)"
	case *got != *want:
		return "= " + floatString(*got) + ", want " + floatString(*want)
	}
	return ""
}

func floatString(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
