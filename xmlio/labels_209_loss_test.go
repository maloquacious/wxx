// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// encodeWithDropped encodes m as app through the public encoder and returns the
// UTF-8 document it wrote and the losses it reported.
func encodeWithDropped(t *testing.T, m *wxx.Map_t, app string) ([]byte, []xmlio.DroppedFeature_t) {
	t.Helper()
	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(app, xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("encode as %s: %v", app, err)
	}
	return ed.Utf8Encoded, ed.Dropped
}

// wantDropped is the part of a DroppedFeature_t a test pins: the path, the field
// and the detail. Reason is prose and is checked only for being stated.
type wantDropped struct {
	path, field, detail string
}

// compareDropped reports every way got differs from want, in order.
func compareDropped(t *testing.T, got []xmlio.DroppedFeature_t, want []wantDropped) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("Dropped has %d entries, want %d:", len(got), len(want))
		for _, d := range got {
			t.Errorf("  got  %s", d)
		}
		for _, w := range want {
			t.Errorf("  want %s (%s): %s", w.path, w.field, w.detail)
		}
		return
	}
	for i, w := range want {
		d := got[i]
		if d.Path != w.path || d.Field != w.field || d.Detail != w.detail {
			t.Errorf("Dropped[%d] = %s (%s): %s; want %s (%s): %s", i, d.Path, d.Field, d.Detail, w.path, w.field, w.detail)
		}
		if d.Reason == "" {
			t.Errorf("Dropped[%d] (%s) states no Reason", i, d.Path)
		}
	}
}

// TestW2025Labels209LossReportedFor208: Worldographer 2.08 writes none of the
// label text-layout attributes 2.09 added, so writing the 2.09 labels fixture
// as 2.08 omits them and REPORTS what that drops (issue #161; the maintainer's
// rulings of 2026-11-18). Only values that change the drawing are reported:
// the one label at lineSpacing 2.0, the City style at 2.0, and the one RIGHT
// label. The two CENTER labels state the default every 2.09 label states, and
// reporting them would bury the real losses.
func TestW2025Labels209LossReportedFor208(t *testing.T) {
	m, err := xmlio.ReadFile(filepath.Join("..", "testdata", labels209Fixture))
	if err != nil {
		t.Fatalf("decode %s: %v", labels209Fixture, err)
	}
	// 2.06 and 2.07 write none of them either, and lose the same.
	for _, app := range []string{"2.08", "2.07", "2.06"} {
		t.Run(app, func(t *testing.T) {
			doc, dropped := encodeWithDropped(t, m, app)
			compareDropped(t, dropped, []wantDropped{
				{"map/labels/label/@lineSpacing", "Map_t.Labels[].LineSpacing", "1 label at 2.0"},
				{"map/labels/label/@textAlignment", "Map_t.Labels[].TextAlignment", "1 label aligned RIGHT"},
				{"map/configuration/text-config/labelstyle/@lineSpacing", "Map_t.Configuration.TextConfig.LabelStyles[].LineSpacing", "1 label style: City at 2.0"},
			})
			assertNoTextLayoutAttrs(t, doc)
		})
	}
}

// TestW2025Labels209LossDefaultsAndFeatureLabels covers what the fixture does
// not: a stated default (lineSpacing 1.0, textAlignment CENTER) is not
// reported, feature labels are reported on their own path, and differing values
// are counted per value. The map is the 2.09 labels fixture with values set
// on it; the 2.08 notes-shapes map supplies the feature labels.
func TestW2025Labels209LossDefaultsAndFeatureLabels(t *testing.T) {
	m, err := xmlio.ReadFile(filepath.Join("..", "testdata", labels209Fixture))
	if err != nil {
		t.Fatalf("decode %s: %v", labels209Fixture, err)
	}
	one, two, half := 1.0, 2.0, 1.5
	// label 1: an explicit default spacing, CENTER -- reports nothing.
	m.Labels[0].LineSpacing = &one
	// label 3: a second spacing and alignment, different values.
	m.Labels[2].LineSpacing = &half
	m.Labels[2].TextAlignment = "LEFT"
	// Every style at an explicit default -- reports nothing.
	for _, s := range m.Configuration.TextConfig.LabelStyles {
		s.LineSpacing = &one
	}

	fm, err := xmlio.ReadFile(filepath.Join("..", "testdata", "2025-2.08-13x11-941577-notes-shapes.wxx"))
	if err != nil {
		t.Fatalf("decode notes-shapes: %v", err)
	}
	var labelled []*wxx.Feature_t
	for _, f := range fm.Features {
		if f.Label != nil {
			labelled = append(labelled, f)
		}
	}
	if len(labelled) < 3 {
		t.Fatalf("notes-shapes has %d labelled features, want at least 3", len(labelled))
	}
	labelled[0].Label.LineSpacing = &two
	labelled[0].Label.TextAlignment = "JUSTIFY"
	labelled[1].Label.LineSpacing = &two
	labelled[1].Label.TextAlignment = "CENTER"
	labelled[2].Label.LineSpacing = &one
	m.Features = fm.Features

	doc, dropped := encodeWithDropped(t, m, "2.08")
	compareDropped(t, dropped, []wantDropped{
		{"map/labels/label/@lineSpacing", "Map_t.Labels[].LineSpacing", "2 labels: 1 at 2.0, 1 at 1.5"},
		{"map/labels/label/@textAlignment", "Map_t.Labels[].TextAlignment", "2 labels: 1 aligned RIGHT, 1 aligned LEFT"},
		{"map/features/feature/label/@lineSpacing", "Map_t.Features[].Label.LineSpacing", "2 feature labels at 2.0"},
		{"map/features/feature/label/@textAlignment", "Map_t.Features[].Label.TextAlignment", "1 feature label aligned JUSTIFY"},
	})
	assertNoTextLayoutAttrs(t, doc)
}

// TestW2025Labels209NoLossOn208Fixtures: a 2.08 map holds none of the 2.09
// attributes, so writing it as 2.08 reports nothing (the new inventory must not
// cry loss on content that was never there). notes-shapes has feature labels;
// populated has standalone labels.
func TestW2025Labels209NoLossOn208Fixtures(t *testing.T) {
	for _, fixture := range []string{
		"2025-2.08-13x11-941577-notes-shapes.wxx",
		"2025-2.08-13x11-941577-populated.wxx",
	} {
		t.Run(fixture, func(t *testing.T) {
			m, err := xmlio.ReadFile(filepath.Join("..", "testdata", fixture))
			if err != nil {
				t.Fatalf("decode %s: %v", fixture, err)
			}
			doc, dropped := encodeWithDropped(t, m, "2.08")
			compareDropped(t, dropped, nil)
			assertNoTextLayoutAttrs(t, doc)
		})
	}
}

// assertNoTextLayoutAttrs fails if doc states lineSpacing or textAlignment on
// any element: no 2.06-2.08 target writes either.
func assertNoTextLayoutAttrs(t *testing.T, doc []byte) {
	t.Helper()
	for _, attr := range []string{" lineSpacing=", " textAlignment="} {
		if n := bytes.Count(doc, []byte(attr)); n != 0 {
			t.Errorf("output states%s %d time(s); a 2.06-2.08 target writes it nowhere", attr, n)
		}
	}
}
