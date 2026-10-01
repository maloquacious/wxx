// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// TestW2025LabelBackgroundColor pins label/@backgroundColor (issue #115) in both
// label contexts, labels/label and feature/label.
//
// No saved fixture carries the attribute. An app check settled it: Worldographer
// 2025 2.08 draws backgroundColor, opaque black included, behind a free label
// or a feature label that has no preset style; a preset style overrides it. Its
// save keeps the attribute in both contexts, written straight after @color, and
// writes it on no label that had none. wxx keeps what the app keeps. So:
//
//   - a nil background writes no attribute;
//   - any set background is written after @color, opaque black included -- the
//     encoder used to drop black, folding it into nil;
//   - it survives the trip back in both contexts -- feature/label used to decode
//     black to nil.
func TestW2025LabelBackgroundColor(t *testing.T) {
	newLabel := func(text string, bg *wxx.RGBA_t) *wxx.Label_t {
		return &wxx.Label_t{
			MapLayer:        "Labels",
			Style:           "Use Custom Style",
			FontFace:        "Arial",
			Color:           &wxx.RGBA_t{R: 1, G: 1, B: 1, A: 1},
			BackgroundColor: bg,
			OutlineColor:    &wxx.RGBA_t{R: 1, G: 1, B: 1, A: 1},
			Location:        &wxx.LabelLocation_t{ViewLevel: "WORLD", X: 150, Y: 150, Scale: 25},
			InnerText:       text,
		}
	}

	m := newRowsMap()
	m.Labels = []*wxx.Label_t{
		newLabel("black", &wxx.RGBA_t{A: 1}),
		newLabel("red", &wxx.RGBA_t{R: 1, A: 1}),
		newLabel("none", nil),
	}
	m.Features = []*wxx.Feature_t{{
		Type:     "Classic/Building Cathedral",
		Uuid:     "00000000-0000-0000-0000-000000000115",
		MapLayer: "Features",
		Location: &wxx.FeatureLocation_t{ViewLevel: "WORLD", X: 150, Y: 150},
		Label:    newLabel("feature black", &wxx.RGBA_t{A: 1}),
	}}

	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.08", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Feature labels come first in the file, then labels/label.
	want := []string{opaqueBlack, opaqueBlack, "1.0,0.0,0.0,1.0", ""}
	tags := startTagAttrs(ed.Utf8Encoded, "label")
	if len(tags) != len(want) {
		t.Fatalf("wrote %d <label> element(s), want %d", len(tags), len(want))
	}
	for i, attrs := range tags {
		have, ok := attrValue(attrs, "backgroundColor")
		if want[i] == "" {
			if ok {
				t.Errorf("<label> %d: wrote backgroundColor=%q for a nil background, want no attribute", i, have)
			}
			continue
		}
		if have != want[i] {
			t.Errorf("<label> %d: @backgroundColor = %q, want %q", i, have, want[i])
		}
		for j, a := range attrs {
			if a[0] == "color" && (j+1 >= len(attrs) || attrs[j+1][0] != "backgroundColor") {
				t.Errorf("<label> %d: @backgroundColor does not follow @color, where the app writes it", i)
			}
		}
	}

	back, err := xmlio.NewDecoder().Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if len(back.Labels) != 3 || len(back.Features) != 1 || back.Features[0].Label == nil {
		t.Fatalf("re-decode: want three labels and one labelled feature")
	}
	for name, c := range map[string]struct {
		got, want *wxx.RGBA_t
	}{
		"feature label": {back.Features[0].Label.BackgroundColor, &wxx.RGBA_t{A: 1}},
		"black label":   {back.Labels[0].BackgroundColor, &wxx.RGBA_t{A: 1}},
		"red label":     {back.Labels[1].BackgroundColor, &wxx.RGBA_t{R: 1, A: 1}},
		"none label":    {back.Labels[2].BackgroundColor, nil},
	} {
		switch {
		case c.want == nil && c.got != nil:
			t.Errorf("re-decode: %s BackgroundColor = %+v, want nil", name, *c.got)
		case c.want != nil && c.got == nil:
			t.Errorf("re-decode: %s BackgroundColor is nil, want %+v", name, *c.want)
		case c.want != nil && *c.got != *c.want:
			t.Errorf("re-decode: %s BackgroundColor = %+v, want %+v", name, *c.got, *c.want)
		}
	}
}
