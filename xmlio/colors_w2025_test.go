// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// opaqueBlack is how Worldographer spells a colour set to Black.
const opaqueBlack = "0.0,0.0,0.0,1.0"

// attrValue returns the value of the named attribute in one start tag.
func attrValue(attrs [][2]string, name string) (string, bool) {
	for _, a := range attrs {
		if a[0] == name {
			return a[1], true
		}
	}
	return "", false
}

// TestW2025FeatureBlackColorMatchesSource decodes the populated 2.08 map,
// encodes it as "2.06", and requires every <feature>'s @color to come out as
// the source spells it (issue #99). The map's cathedral on hex (2,1) has
// Override Color set to Black, which the app writes as "0.0,0.0,0.0,1.0"; wxx
// used to fold that into the same nil as "null" and write it back as "null",
// so the feature lost its colour.
//
// @ringColor is not compared: the codec reads it as "ringcolor" and loses it
// (issue #100).
func TestW2025FeatureBlackColorMatchesSource(t *testing.T) {
	src, m := readNotesFixture(t, populatedFixture)
	out, err := xmlio.MarshalXML(m, "2.06")
	if err != nil {
		t.Fatalf("MarshalXML: %v", err)
	}
	in := startTagAttrs(src, "feature")
	got := startTagAttrs(out, "feature")
	if len(got) != len(in) {
		t.Fatalf("wrote %d <feature>(s), source has %d", len(got), len(in))
	}
	black := false
	for i := range in {
		want, ok := attrValue(in[i], "color")
		if !ok {
			t.Fatalf("source <feature> %d states no @color", i)
		}
		black = black || want == opaqueBlack
		if have, _ := attrValue(got[i], "color"); have != want {
			t.Errorf("<feature> %d: @color = %q, want %q", i, have, want)
		}
	}
	if !black {
		t.Fatalf("%s: no <feature> states color=%q; the fixture no longer exercises #99", populatedFixture, opaqueBlack)
	}
}

// TestW2025BlackIsNotNull covers the other nullable colours that folded opaque
// black into "null" or dropped it (issue #99): feature @ringColor, note @color,
// shapestyle @fillPaint @dscolor @insColor, and a tile's custom background,
// whose column the encoder omits when nil. No fixture carries a black in any
// of them, so the map is synthesized, as TestW2025LabelStyleBlackBackgroundIsNotNull
// does for the label style colours.
func TestW2025BlackIsNotNull(t *testing.T) {
	black := func() *wxx.RGBA_t { return &wxx.RGBA_t{A: 1} }

	m := newRowsMap()
	m.Features = []*wxx.Feature_t{{
		Type:      "Classic/Building Cathedral",
		Uuid:      "00000000-0000-0000-0000-000000000099",
		MapLayer:  "Features",
		Color:     black(),
		RingColor: black(),
		Location:  &wxx.FeatureLocation_t{ViewLevel: "WORLD", X: 150, Y: 150},
	}}
	m.Notes = []*wxx.Note_t{{
		Color:    black(),
		Title:    "Black note",
		Location: &wxx.NoteLocation_t{ViewLevel: "WORLD", X: 150, Y: 150},
	}}
	m.Configuration.ShapeConfig = &wxx.ShapeConfig_t{ShapeStyles: []*wxx.ShapeStyle_t{{
		Name:        "Black Style",
		StrokePaint: black(),
		FillPaint:   black(),
		DsColor:     black(),
		InsColor:    black(),
	}}}
	m.Tiles.Tiles[0][0].CustomBackgroundColor = black()

	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.06", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("encode: %v", err)
	}

	for _, c := range []struct {
		element string
		attrs   []string
	}{
		{"feature", []string{"color", "ringcolor"}},
		{"note", []string{"color"}},
		{"shapestyle", []string{"fillPaint", "dscolor", "insColor"}},
	} {
		tags := startTagAttrs(ed.Utf8Encoded, c.element)
		if len(tags) != 1 {
			t.Fatalf("wrote %d <%s> element(s), want 1", len(tags), c.element)
		}
		for _, name := range c.attrs {
			if have, _ := attrValue(tags[0], name); have != opaqueBlack {
				t.Errorf("<%s> @%s = %q, want %q -- an opaque black must not be written as \"null\"", c.element, name, have, opaqueBlack)
			}
		}
	}

	// And each must survive the trip back: written correctly but decoded to
	// nil would be the same loss one step later.
	back, err := xmlio.NewDecoder().Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if len(back.Features) != 1 || len(back.Notes) != 1 || back.Configuration.ShapeConfig == nil || len(back.Configuration.ShapeConfig.ShapeStyles) != 1 {
		t.Fatalf("re-decode: want one feature, note and shape style")
	}
	ss := back.Configuration.ShapeConfig.ShapeStyles[0]
	for name, got := range map[string]*wxx.RGBA_t{
		"feature color":        back.Features[0].Color,
		"feature ringColor":    back.Features[0].RingColor,
		"note color":           back.Notes[0].Color,
		"shapestyle fillPaint": ss.FillPaint,
		"shapestyle dscolor":   ss.DsColor,
		"shapestyle insColor":  ss.InsColor,
		"tile custom colour":   back.Tiles.Tiles[0][0].CustomBackgroundColor,
	} {
		if got == nil {
			t.Errorf("re-decode: %s is nil, want an opaque black -- nil means \"null\"", name)
		}
	}
}
