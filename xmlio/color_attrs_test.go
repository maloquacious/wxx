// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// colorShape is a shape whose colors are all valid, so a case can break one.
func colorShape() *wxx.Shape_t {
	return &wxx.Shape_t{
		Type: "Polygon", MapLayer: "Above Terrain", CreationType: "BASIC",
		IsWorld: true, HighestViewLevel: "WORLD", CurrentShapeViewLevel: "WORLD",
		StrokeColor: "1.0,1.0,1.0,1.0", DsColor: "null", InsColor: "null",
		StrokeType: "SIMPLE", Opacity: 1,
		Points: []*wxx.Point_t{{X: 10, Y: 10}, {X: 20, Y: 10}, {X: 15, Y: 20}},
	}
}

// TestColorAttributesRefused: a color attribute the model holds as a string is
// written verbatim, so the encoder checks it and refuses one Worldographer
// cannot read, before writing a byte and naming the attribute (issue #83).
//
// The case that started it is a shape built without colors: dsColor="" makes
// Worldographer 2.08 fail in LoadMapTask.readShapes with "empty String". The
// other attributes here are the same kind of field.
func TestColorAttributesRefused(t *testing.T) {
	for _, tc := range []struct {
		name   string
		app    string
		break_ func(*wxx.Map_t)
		path   string
	}{
		{"shape dsColor empty", "2.06", func(m *wxx.Map_t) { m.Shapes[0].DsColor = "" }, "map/shapes/shape/@dsColor"},
		{"shape insColor empty", "2.06", func(m *wxx.Map_t) { m.Shapes[0].InsColor = "" }, "map/shapes/shape/@insColor"},
		{"shape strokeColor malformed", "2.06", func(m *wxx.Map_t) { m.Shapes[0].StrokeColor = "white" }, "map/shapes/shape/@strokeColor"},
		{"shape dsColor three parts", "2.06", func(m *wxx.Map_t) { m.Shapes[0].DsColor = "1.0,1.0,1.0" }, "map/shapes/shape/@dsColor"},
		{"shape dsColor out of range", "2.06", func(m *wxx.Map_t) { m.Shapes[0].DsColor = "255,0,0,1" }, "map/shapes/shape/@dsColor"},
		{"label dropShadowColor malformed", "2.06", func(m *wxx.Map_t) {
			m.Labels = []*wxx.Label_t{{MapLayer: "Labels", Style: "Nation", FontFace: "Arial",
				Color: &wxx.RGBA_t{A: 1}, OutlineColor: &wxx.RGBA_t{A: 1}, DropShadowColor: "black",
				Location: &wxx.LabelLocation_t{ViewLevel: "WORLD", X: 10, Y: 10, Scale: 12.5}, InnerText: "x"}}
		}, "map/labels/label/@dropShadowColor"},
		{"labelstyle dropShadowColor malformed", "2.06", func(m *wxx.Map_t) {
			m.Configuration.TextConfig.LabelStyles = []*wxx.LabelStyle_t{{Name: "Nation", DropShadowColor: "black"}}
		}, "map/configuration/text-config/labelstyle/@dropShadowColor"},
		{"grid color malformed", "2.06", func(m *wxx.Map_t) { m.GridAndNumbering.Color2 = "0x123" }, "map/gridandnumbering/@color2"},
		{"grid color rgba", "2.06", func(m *wxx.Map_t) { m.GridAndNumbering.Color0 = "0.0,0.0,0.0,0.25" }, "map/gridandnumbering/@color0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newRowsMap()
			m.Shapes = []*wxx.Shape_t{colorShape()}
			tc.break_(m)
			var buf bytes.Buffer
			err := xmlio.NewEncoder(tc.app).Encode(&buf, m)
			if !errors.Is(err, wxx.ErrInvalidColorAttribute) {
				t.Fatalf("Encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidColorAttribute)
			}
			if !strings.Contains(err.Error(), tc.path) {
				t.Errorf("Encode: err = %q, want it to name %s", err, tc.path)
			}
			if buf.Len() != 0 {
				t.Errorf("Encode: wrote %d bytes, want 0", buf.Len())
			}
		})
	}
}

// TestColorAttributesAccepted: the spellings the samples use still go through
// unchanged -- "null" where a color is nullable, "r,g,b,a", "0xRRGGBBAA" -- and
// a label with no drop shadow ("") omits the trio rather than being refused.
func TestColorAttributesAccepted(t *testing.T) {
	m := newRowsMap()
	shape := colorShape()
	shape.DsColor = "1.0,0.8941176533699036,0.7686274647712708,1.0"
	m.Shapes = []*wxx.Shape_t{shape, colorShape()}
	m.Labels = []*wxx.Label_t{{MapLayer: "Labels", Style: "Nation", FontFace: "Arial",
		Color: &wxx.RGBA_t{A: 1}, OutlineColor: &wxx.RGBA_t{A: 1},
		Location: &wxx.LabelLocation_t{ViewLevel: "WORLD", X: 10, Y: 10, Scale: 12.5}, InnerText: "x"}}
	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.06", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	out := string(ed.Utf8Encoded)
	for _, want := range []string{
		`dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0"`,
		`dsColor="null"`, `insColor="null"`, `strokeColor="1.0,1.0,1.0,1.0"`,
		`color0="0x00000040"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %s", want)
		}
	}
	if strings.Contains(out, `<label mapLayer="Labels"`) && strings.Contains(out, "dropShadowColor=\"\"") {
		t.Errorf("a label without a drop shadow was written with dropShadowColor=\"\"")
	}
}

// TestColorAttributesAppSpellings pins what the maintainer's app check showed
// Worldographer 2.08 reads and writes, which no sample had stated (issue #83):
//
//   - gridandnumbering color0="" opens, and the app saves it as "null";
//   - shape strokeColor="null" opens, and the app saves the shape with no
//     strokeColor attribute at all, which decodes to "".
//
// Each spelling must round-trip: written as given, never refused.
func TestColorAttributesAppSpellings(t *testing.T) {
	m := newRowsMap()
	m.GridAndNumbering.Color0 = "null"
	m.GridAndNumbering.Color1 = ""
	nullStroke, noStroke := colorShape(), colorShape()
	nullStroke.StrokeColor = "null"
	noStroke.StrokeColor = ""
	m.Shapes = []*wxx.Shape_t{nullStroke, noStroke}

	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.06", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("Encode: %v -- a spelling Worldographer itself writes was refused", err)
	}
	out := string(ed.Utf8Encoded)
	for _, want := range []string{`color0="null"`, `color1=""`, `strokeColor="null"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %s", want)
		}
	}
	shapes := startTagAttrs(ed.Utf8Encoded, "shape")
	if len(shapes) != 2 {
		t.Fatalf("wrote %d shape(s), want 2", len(shapes))
	}
	for _, a := range shapes[1] {
		if a[0] == "strokeColor" {
			t.Errorf("a shape with no stroke color was written with strokeColor=%q; Worldographer omits the attribute", a[1])
		}
	}

	back, err := xmlio.NewDecoder().Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if got := [2]string{back.Shapes[0].StrokeColor, back.Shapes[1].StrokeColor}; got != [2]string{"null", ""} {
		t.Errorf("stroke colors read back as %q, want [null \"\"]", got)
	}
	if got := [2]string{back.GridAndNumbering.Color0, back.GridAndNumbering.Color1}; got != [2]string{"null", ""} {
		t.Errorf("grid colors read back as %q, want [null \"\"]", got)
	}
}
