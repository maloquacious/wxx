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

// populatedFixtures are the populated recipe's maps, one per version saved.
// The 2.06 map confirms the 2.08 colour spellings, ringColor included (#100).
var populatedFixtures = []string{
	"2025-2.06-13x11-941577-populated.wxx",
	populatedFixture,
}

// TestW2025FeatureBlackColorMatchesSource decodes each populated map, encodes
// it as the version it states (see sameVersionTarget), and requires every <feature>'s @color to come out as the
// source spells it (issue #99). Each map's cathedral on hex (2,1) has
// Override Color set to Black, which the app writes as "0.0,0.0,0.0,1.0"; wxx
// used to fold that into the same nil as "null" and write it back as "null",
// so the feature lost its colour.
//
// The ring colour is compared by TestW2025FeatureRingColorMatchesSource.
func TestW2025FeatureBlackColorMatchesSource(t *testing.T) {
	for _, fixture := range populatedFixtures {
		t.Run(fixture, func(t *testing.T) {
			src, m := readNotesFixture(t, fixture)
			out, err := xmlio.MarshalXML(m, sameVersionTarget(t, fixture))
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
				t.Fatalf("no <feature> states color=%q; the fixture no longer exercises #99", opaqueBlack)
			}
		})
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
		{"feature", []string{"color", "ringColor"}},
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

// ringAttr returns the index, name and value of a <feature>'s ring colour,
// whichever way it is spelled.
func ringAttr(attrs [][2]string) (int, string, string) {
	for i, a := range attrs {
		if strings.EqualFold(a[0], "ringcolor") {
			return i, a[0], a[1]
		}
	}
	return -1, "", ""
}

// TestW2025FeatureRingColorMatchesSource decodes each W2025 fixture that has
// features, encodes it as the version it states (see
// sameVersionTarget), and requires every <feature>'s ring colour
// to come out with the source's name, value and position (issue #100). The
// app writes ringcolor="null" when no ring colour is set and ringColor when
// one is; wxx read only the first, so a set ring colour was lost. The
// populated 2.06 and 2.08 maps each have a white and a black ring; the test
// fails if no fixture states ringColor.
func TestW2025FeatureRingColorMatchesSource(t *testing.T) {
	fixtures := append(append([]string(nil), notesShapesFixtures...), populatedFixtures...)
	camel := false
	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			src, m := readNotesFixture(t, fixture)
			out, err := xmlio.MarshalXML(m, sameVersionTarget(t, fixture))
			if err != nil {
				t.Fatalf("MarshalXML: %v", err)
			}
			in := startTagAttrs(src, "feature")
			got := startTagAttrs(out, "feature")
			if len(got) != len(in) {
				t.Fatalf("wrote %d <feature>(s), source has %d", len(got), len(in))
			}
			for i := range in {
				wi, wn, wv := ringAttr(in[i])
				if wi < 0 {
					t.Fatalf("source <feature> %d states no ring colour", i)
				}
				camel = camel || wn == "ringColor"
				gi, gn, gv := ringAttr(got[i])
				if gn != wn || gv != wv {
					t.Errorf("<feature> %d: %s=%q, want %s=%q", i, gn, gv, wn, wv)
				}
				// Compared as the position after @color, since the encoder
				// writes attributes the source may lack (e.g. no @uuid).
				if ci, _ := indexOf(got[i], "color"); gi != ci+1 {
					t.Errorf("<feature> %d: ring colour is attribute %d, want it straight after @color (%d)", i, gi, ci)
				}
				if ci, _ := indexOf(in[i], "color"); wi != ci+1 {
					t.Errorf("source <feature> %d: ring colour is not straight after @color", i)
				}
			}
		})
	}
	if !camel {
		t.Fatalf("no fixture states ringColor; the test no longer exercises #100")
	}
}

// indexOf returns the position of the named attribute in one start tag.
func indexOf(attrs [][2]string, name string) (int, bool) {
	for i, a := range attrs {
		if a[0] == name {
			return i, true
		}
	}
	return -1, false
}

// TestW2025FeatureRingColorBothSpellingsRefused requires a <feature> that
// states both ringcolor and ringColor to be refused rather than resolved
// (issue #100): the app writes one or the other, never both, so there is no
// telling which one it meant.
func TestW2025FeatureRingColorBothSpellingsRefused(t *testing.T) {
	src, _ := readNotesFixture(t, populatedFixture)
	src = bytes.TrimLeft(stripXMLDecl(src), "\n")
	const white = `ringColor="1.0,1.0,1.0,1.0"`
	if n := bytes.Count(src, []byte(white)); n != 1 {
		t.Fatalf("fixture states %s %d time(s), want 1", white, n)
	}
	doc := bytes.Replace(src, []byte(white), []byte(`ringcolor="null" `+white), 1)
	_, err := decodeRawXML(t, doc, "1.1")
	if !errors.Is(err, wxx.ErrAttributeSpelledTwice) {
		t.Fatalf("decode: err = %v, want errors.Is(err, %v)", err, wxx.ErrAttributeSpelledTwice)
	}
	if !strings.Contains(err.Error(), "map/features/feature") {
		t.Errorf("decode: err = %q, want it to name the feature", err)
	}

	// The unmodified document decodes, so the refusal above is the edit's.
	if _, err := decodeRawXML(t, src, "1.1"); err != nil {
		t.Fatalf("decode unmodified: %v", err)
	}
}
