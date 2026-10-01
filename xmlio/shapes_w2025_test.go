// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// populatedFixture is the maintainer's populated 2.08 map (issue #94), whose
// recipe is in testdata/README.md. It has one shape, a curved path. Its
// features also show #99 and #100, so it is used here for <shapes> only.
const populatedFixture = "2025-2.08-13x11-941577-populated.wxx"

// shapeElement_t is one <shape> as a document states it: every attribute on
// the start tag, and every <p>'s attributes, in document order.
type shapeElement_t struct {
	attrs  map[string]string
	points []map[string]string
}

// shapeElements parses every <shape> inside the document's <shapes> element
// with encoding/xml. Only that element is parsed, so the XML 1.1 declaration a
// W2025 document opens with never reaches encoding/xml. Attribute values are
// the strings the document spells, so "2700" and "2700.0" differ.
func shapeElements(t *testing.T, label string, doc []byte) []shapeElement_t {
	t.Helper()
	start := bytes.Index(doc, []byte("<shapes>"))
	end := bytes.Index(doc, []byte("</shapes>"))
	if start < 0 || end < start {
		t.Fatalf("%s: no <shapes>...</shapes> element", label)
	}
	dec := xml.NewDecoder(bytes.NewReader(doc[start : end+len("</shapes>")]))
	var shapes []shapeElement_t
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("%s: parse <shapes>: %v", label, err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := map[string]string{}
		for _, a := range se.Attr {
			attrs[a.Name.Local] = a.Value
		}
		switch se.Name.Local {
		case "shapes":
		case "shape":
			shapes = append(shapes, shapeElement_t{attrs: attrs})
		case "p":
			if len(shapes) == 0 {
				t.Fatalf("%s: <p> outside a <shape>", label)
			}
			cur := &shapes[len(shapes)-1]
			cur.points = append(cur.points, attrs)
		default:
			t.Errorf("%s: unexpected <%s> in <shapes>", label, se.Name.Local)
		}
	}
	return shapes
}

// integerSpelledPoint reports whether a <p> spells both coordinates without a
// decimal point or an exponent.
func integerSpelledPoint(p map[string]string) bool {
	return !strings.ContainsAny(p["x"], ".eE") && !strings.ContainsAny(p["y"], ".eE")
}

// TestW2025ShapesMatchSource decodes each notes-and-shapes fixture, encodes it
// as the version it states (see sameVersionTarget), and compares every <shape> element the encoder wrote with the
// source's: the start tag's attributes and each <p>'s, as the strings the
// documents spell (issue #94). So @fillRule and p/@type must be present
// exactly where the source has them, the four @extraLine* must survive, and a
// coordinate the source spells "2700" must not come back as "2700.0".
//
// Attribute order and whitespace are not compared here; the encoder writes
// the app's order, which the COVERAGE.md entry for <shape> records.
//
// The populated 2.08 map (issue #94) adds a curved path, whose type="c" points
// state control points in @cx1 @cy1 @cx2 @cy2, and the first coordinates with
// long fractional parts (x="1715.028150714595"); both are compared as exact
// strings like every other attribute. The river-lines 2.08 map (issue #154)
// adds two lines drawn with Snap Points to Grid.
func TestW2025ShapesMatchSource(t *testing.T) {
	fixtures := append(append([]string(nil), notesShapesFixtures...), populatedFixture, riverLinesFixture)
	curve := false // a <p> with @cx1 in some fixture's source
	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			src, m := readNotesFixture(t, fixture)
			out, err := xmlio.MarshalXML(m, sameVersionTarget(t, fixture))
			if err != nil {
				t.Fatalf("MarshalXML: %v", err)
			}
			in := shapeElements(t, fixture+" source", src)
			got := shapeElements(t, fixture+" output", out)

			for _, sh := range in {
				for _, p := range sh.points {
					if _, ok := p["cx1"]; ok {
						curve = true
					}
				}
			}
			if len(got) != len(in) {
				t.Fatalf("wrote %d <shape>(s), source has %d", len(got), len(in))
			}
			for i := range in {
				compareAttrSets(t, fixture, "shape", i, in[i].attrs, got[i].attrs)
				if len(got[i].points) != len(in[i].points) {
					t.Errorf("<shape> %d: wrote %d <p>(s), source has %d", i, len(got[i].points), len(in[i].points))
					continue
				}
				for j := range in[i].points {
					compareAttrSets(t, fixture, fmt.Sprintf("shape %d/p", i), j, in[i].points[j], got[i].points[j])
				}
			}
			if fixture == populatedFixture || fixture == riverLinesFixture {
				return
			}

			// vacuous-pass guards: the comparison above proves nothing about a
			// spelling the fixture does not contain. The notes-shapes recipe
			// draws every one of these; the populated map draws one curve.
			var integerPolygon, typedPath, fillRulePath, extraLine bool
			for _, sh := range in {
				if _, ok := sh.attrs["extraLineWidth"]; ok {
					extraLine = true
				}
				switch sh.attrs["type"] {
				case "Polygon":
					all := len(sh.points) > 0
					for _, p := range sh.points {
						all = all && integerSpelledPoint(p)
					}
					integerPolygon = integerPolygon || all
				case "Path":
					if _, ok := sh.attrs["fillRule"]; ok {
						fillRulePath = true
					}
					for _, p := range sh.points {
						if _, ok := p["type"]; ok {
							typedPath = true
						}
					}
				}
			}
			if !integerPolygon {
				t.Fatalf("source has no Polygon whose points are all integer-spelled")
			}
			if !typedPath || !fillRulePath {
				t.Fatalf("source has no Path with a p/@type (%v) and a @fillRule (%v)", typedPath, fillRulePath)
			}
			if !extraLine {
				t.Fatalf("source has no shape with @extraLineWidth")
			}
			var untypedPolygon bool
			for _, sh := range in {
				if _, ok := sh.attrs["fillRule"]; !ok {
					untypedPolygon = true
				}
			}
			if !untypedPolygon {
				t.Fatalf("source has no shape without @fillRule")
			}
		})
	}
	if !curve {
		t.Fatalf("no fixture has a <p> with @cx1, so control points are not under test")
	}
}

// TestW2025ShapeExtraLineZeroOmitted: a shape with all four @extraLine* at
// zero, which is what a classic-decoded shape held before issue #103, is
// written with none of them, as before #94. Any one non-zero writes all four. Whether the app reads
// extraLineWidth="0.0" differently from no attribute has not been tested.
func TestW2025ShapeExtraLineZeroOmitted(t *testing.T) {
	names := []string{"extraLineDistance", "extraLineLength", "extraLineWidth", "extraLineSeparation"}
	_, m := readNotesFixture(t, notesShapesFixtures[0])
	if len(m.Shapes) < 3 {
		t.Fatalf("fixture has %d shape(s), want at least 3", len(m.Shapes))
	}
	zero, one := m.Shapes[0], m.Shapes[1]
	zero.ExtraLineDistance, zero.ExtraLineLength, zero.ExtraLineWidth, zero.ExtraLineSeparation = 0, 0, 0, 0
	one.ExtraLineDistance, one.ExtraLineLength, one.ExtraLineWidth, one.ExtraLineSeparation = 0, 0, 2.5, 0

	out, err := xmlio.MarshalXML(m, "2.06")
	if err != nil {
		t.Fatalf("MarshalXML: %v", err)
	}
	got := shapeElements(t, "output", out)
	for _, name := range names {
		if v, ok := got[0].attrs[name]; ok {
			t.Errorf("all-zero shape: wrote @%s = %q, want none of the four", name, v)
		}
	}
	want := map[string]string{"extraLineDistance": "0.0", "extraLineLength": "0.0", "extraLineWidth": "2.5", "extraLineSeparation": "0.0"}
	for _, name := range names {
		if v, ok := got[1].attrs[name]; !ok || v != want[name] {
			t.Errorf("one non-zero: @%s = %q (present %v), want %q", name, v, ok, want[name])
		}
	}
	// The untouched shape keeps its decoded values.
	for _, name := range names {
		if _, ok := got[2].attrs[name]; !ok {
			t.Errorf("untouched shape: @%s missing", name)
		}
	}
}

// TestW2025ShapeIntegerXYFractionalRefused: a point marked as integer-spelled
// with a fractional coordinate cannot be written either way without changing
// something the caller asked for, so the encoder refuses it before writing a
// byte and names the point (issue #94). Without the mark, the same value is
// written with its decimal point.
func TestW2025ShapeIntegerXYFractionalRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  func(p *wxx.Point_t)
		path string
	}{
		{"x", func(p *wxx.Point_t) { p.X += 0.5 }, "map/shapes/shape[1]/p[3]/@x"},
		{"y", func(p *wxx.Point_t) { p.Y -= 0.25 }, "map/shapes/shape[1]/p[3]/@y"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, m := readNotesFixture(t, notesShapesFixtures[0])
			p := m.Shapes[0].Points[2]
			if !p.IntegerXY {
				t.Fatalf("fixture's first shape's third point is not integer-spelled")
			}
			tc.set(p)

			out, err := xmlio.MarshalXML(m, "2.06")
			if !errors.Is(err, wxx.ErrInvalidIntegerAttribute) {
				t.Fatalf("MarshalXML: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidIntegerAttribute)
			}
			if !strings.Contains(err.Error(), tc.path) {
				t.Errorf("MarshalXML: err = %q, want it to name %s", err, tc.path)
			}
			if len(out) != 0 {
				t.Errorf("MarshalXML: returned %d bytes, want 0", len(out))
			}
			var buf bytes.Buffer
			if err := xmlio.NewEncoder("2.06").Encode(&buf, m); !errors.Is(err, wxx.ErrInvalidIntegerAttribute) {
				t.Fatalf("Encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidIntegerAttribute)
			}
			if buf.Len() != 0 {
				t.Errorf("Encode: wrote %d bytes, want 0", buf.Len())
			}

			// Cleared, the same value is written as a decimal.
			p.IntegerXY = false
			out, err = xmlio.MarshalXML(m, "2.06")
			if err != nil {
				t.Fatalf("MarshalXML with IntegerXY cleared: %v", err)
			}
			got := shapeElements(t, "output", out)[0].points[2]
			if strings.ContainsAny(got["x"]+got["y"], "eE") || !strings.Contains(got[tc.name], ".") {
				t.Errorf("IntegerXY cleared: wrote x=%q y=%q, want decimals", got["x"], got["y"])
			}
		})
	}
}

// TestW2025ShapePointSpellingDecode: IntegerXY is set only when both
// coordinates are spelled as integers. A point spelled one each way decodes
// with it false and is written back with both decimal: the value is kept and
// the spelling of the integer coordinate changes. No 2.06, 2.07 or 2.08
// fixture has such a point.
func TestW2025ShapePointSpellingDecode(t *testing.T) {
	src, _ := readNotesFixture(t, notesShapesFixtures[0])
	src = bytes.TrimLeft(stripXMLDecl(src), "\n")
	const point = `<p x="2700" y="150"/>`
	if n := bytes.Count(src, []byte(point)); n != 1 {
		t.Fatalf("fixture states %s %d time(s), want 1", point, n)
	}
	for _, tc := range []struct {
		name, to     string
		wantInteger  bool
		wantX, wantY string
	}{
		{"both integer", point, true, "2700", "150"},
		{"y decimal", `<p x="2700" y="150.0"/>`, false, "2700.0", "150.0"},
		{"x decimal", `<p x="2700.0" y="150"/>`, false, "2700.0", "150.0"},
		{"exponent", `<p x="27e2" y="150"/>`, false, "2700.0", "150.0"},
		{"signed integer", `<p x="+2700" y="150"/>`, true, "2700", "150"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := bytes.Replace(src, []byte(point), []byte(tc.to), 1)
			m, err := decodeRawXML(t, doc, "1.1")
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			p := m.Shapes[0].Points[0]
			if p.X != 2700 || p.Y != 150 || p.IntegerXY != tc.wantInteger {
				t.Fatalf("decoded %+v, want X 2700 Y 150 IntegerXY %v", *p, tc.wantInteger)
			}
			out, err := xmlio.MarshalXML(m, "2.06")
			if err != nil {
				t.Fatalf("MarshalXML: %v", err)
			}
			got := shapeElements(t, "output", out)[0].points[0]
			if got["x"] != tc.wantX || got["y"] != tc.wantY {
				t.Errorf("wrote x=%q y=%q, want x=%q y=%q", got["x"], got["y"], tc.wantX, tc.wantY)
			}
		})
	}

	// A coordinate that is not a number is refused, naming the point.
	doc := bytes.Replace(src, []byte(point), []byte(`<p x="2700" y="north"/>`), 1)
	if _, err := decodeRawXML(t, doc, "1.1"); err == nil || !strings.Contains(err.Error(), "map/shapes/shape[1]/p[1]/@y") {
		t.Errorf("decode y=\"north\": err = %v, want it to name map/shapes/shape[1]/p[1]/@y", err)
	}
}

// TestW2025ShapeCurveControl: a curve point's control points are written only
// when the point has them, a control point at 0,0 is written rather than taken
// for absent, and a point stating some but not all four is refused on decode
// (issue #94).
func TestW2025ShapeCurveControl(t *testing.T) {
	src, m := readNotesFixture(t, populatedFixture)
	var curve *wxx.Point_t
	for _, sh := range m.Shapes {
		for _, p := range sh.Points {
			if p.Control != nil {
				curve = p
			}
		}
	}
	if curve == nil {
		t.Fatalf("%s decoded no point with control points", populatedFixture)
	}
	curve.Control = &wxx.CurveControl_t{}
	out, err := xmlio.MarshalXML(m, "2.06")
	if err != nil {
		t.Fatalf("MarshalXML: %v", err)
	}
	var found bool
	for _, sh := range shapeElements(t, "output", out) {
		for _, p := range sh.points {
			if p["type"] == "c" {
				found = true
				for _, name := range []string{"cx1", "cy1", "cx2", "cy2"} {
					if p[name] != "0.0" {
						t.Errorf("zero control point: @%s = %q, want \"0.0\"", name, p[name])
					}
				}
			} else if _, ok := p["cx1"]; ok {
				t.Errorf("point with no control points: wrote @cx1 = %q", p["cx1"])
			}
		}
	}
	if !found {
		t.Fatalf("output has no type=\"c\" point")
	}

	doc := bytes.TrimLeft(stripXMLDecl(src), "\n")
	const cy2 = ` cy2 = "2535.0"`
	if n := bytes.Count(doc, []byte(cy2)); n != 1 {
		t.Fatalf("fixture states %q %d time(s), want 1", cy2, n)
	}
	_, err = decodeRawXML(t, bytes.Replace(doc, []byte(cy2), nil, 1), "1.1")
	if err == nil || !strings.Contains(err.Error(), "map/shapes/shape[1]/p[2]") || !strings.Contains(err.Error(), "3 of @cx1") {
		t.Errorf("decode with @cy2 removed: err = %v, want it to name map/shapes/shape[1]/p[2] and 3 of the 4", err)
	}
}
