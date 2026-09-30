// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// extraTerrainElement matches the whole <extraTerrain> element and the newline
// after it.
var extraTerrainElement = regexp.MustCompile(`(?s)<extraTerrain>.*?</extraTerrain>\n`)

// TestW2025ExtraTerrainMatchesSource asserts that the <extraTerrain> the encoder
// writes is byte-for-byte the one the source states (issue #34).
//
// Before #34 this held trivially: the element was carried as verbatim InnerXML.
// It is now rebuilt from the model, so this is the test that the model and the
// encoder's layout together lose nothing -- names, order, values, spelling
// ("225.0,150.0", "Z", "false") and whitespace. Byte equality is affordable here
// because every sample lays the element out the same way.
//
// Every sample is encoded as 2.06, including the 2.07 and 2.08 ones: 2.06 is
// the only W2025 target registered until issue #73 adds theirs. All three write
// <extraTerrain> the same way, so the comparison still holds.
func TestW2025ExtraTerrainMatchesSource(t *testing.T) {
	type pair struct{ in, out []byte }
	cases := map[string]func(t *testing.T) pair{}
	for _, fixture := range []string{
		"2025-2.06-13x11-941577-blank.wxx",
		"2025-2.06-13x11-941577-layers-beta.wxx",
		"2025-2.06-13x11-941577-layers.wxx",
		"2025-2.07-13x11-941577-layers.wxx",
		"2025-2.08-13x11-941577-layers.wxx",
	} {
		cases[fixture] = func(t *testing.T) pair {
			f, err := os.Open(filepath.Join("..", "testdata", fixture))
			if err != nil {
				t.Fatalf("open %s: %v", fixture, err)
			}
			defer f.Close()
			var dd xmlio.DecoderDiagnostics
			m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f)
			if err != nil {
				t.Fatalf("decode %s: %v", fixture, err)
			}
			var ed xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder("2.06", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode %s: %v", fixture, err)
			}
			return pair{dd.Converted, ed.Utf8Encoded}
		}
	}
	cases[filepath.Base(populatedFixture)] = func(t *testing.T) pair {
		raw, err := os.ReadFile(populatedFixture)
		if err != nil {
			t.Fatalf("read %s: %v", populatedFixture, err)
		}
		out, err := v1_06.Encode(decodeFixture(t, populatedFixture), "2.06")
		if err != nil {
			t.Fatalf("encode %s: %v", populatedFixture, err)
		}
		return pair{raw, out}
	}

	sawPlacement := false
	for name, load := range cases {
		t.Run(name, func(t *testing.T) {
			docs := load(t)
			in := extraTerrainElement.Find(docs.in)
			out := extraTerrainElement.Find(docs.out)
			if in == nil {
				t.Fatalf("%s: source states no <extraTerrain>; this fixture cannot evidence anything", name)
			}
			if !bytes.Equal(in, out) {
				t.Errorf("%s: <extraTerrain> differs from the source\n got: %q\nwant: %q", name, out, in)
			}
			if bytes.Contains(in, []byte("<terrainAndLocation")) {
				sawPlacement = true
			}
		})
	}
	// Byte equality on empty containers alone would prove only the layout of
	// "<extraTerrain>\n</extraTerrain>".
	if !sawPlacement {
		t.Errorf("no fixture places terrain in <extraTerrain>, so the model's content is untested here")
	}
}

// extraTerrainMap returns a W2025 map whose <extraTerrain> exercises what the
// tracked fixtures do not: several layers, one hex placed on two layers and
// twice on one layer, both @resources spellings, icy and GM-only placements, a
// negative elevation, and a layer name that needs escaping.
func extraTerrainMap() *wxx.Map_t {
	m := newRowsMap()
	m.ExtraTerrain = &wxx.ExtraTerrain_t{MapLayers: []*wxx.ExtraTerrainLayer_t{
		{Name: "Below All", Terrain: []*wxx.TerrainAndLocation_t{
			{Terrain: "Classic/Underdark Broken Lands", Elevation: -100, X: 450, Y: 0},
			{Terrain: "Classic/Mountain Volcano", Elevation: 5000, IsIcy: true,
				Resources: wxx.Resources_t{Animal: 9, Brick: 14, Crops: 7, Gems: 9, Lumber: 7, Metals: 54, Rock: 44}, X: 1125, Y: 1650},
			{Terrain: "Classic/Flat Beach", Elevation: 1, X: 1125, Y: 1650},
		}},
		{Name: "Above Water", Terrain: []*wxx.TerrainAndLocation_t{
			{Terrain: "Classic/Water Sea", Elevation: -1, IsGMOnly: true, X: 1125, Y: 1650},
		}},
		{Name: `Mine & "Tunnels"`},
	}}
	return m
}

// TestW2025ExtraTerrainRoundTrip writes a multi-layer <extraTerrain>, checks the
// exact text, and reads it back to the same model (issue #34). No tracked
// fixture places more than one terrain, so the shape is synthesized.
func TestW2025ExtraTerrainRoundTrip(t *testing.T) {
	m := extraTerrainMap()
	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder("2.06", xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("encode: %v", err)
	}

	const want = "<extraTerrain>\n" +
		"\t<mapLayer name=\"Below All\">\n" +
		"\t\t<terrainAndLocation name=\"Classic/Underdark Broken Lands\" elevation=\"-100\" icy=\"false\" gmOnly=\"false\" resources=\"Z\" location=\"450.0,0.0\" />\n" +
		"\t\t<terrainAndLocation name=\"Classic/Mountain Volcano\" elevation=\"5000\" icy=\"true\" gmOnly=\"false\" resources=\"9,14,7,9,7,54,44\" location=\"1125.0,1650.0\" />\n" +
		"\t\t<terrainAndLocation name=\"Classic/Flat Beach\" elevation=\"1\" icy=\"false\" gmOnly=\"false\" resources=\"Z\" location=\"1125.0,1650.0\" />\n" +
		"\t</mapLayer>\n" +
		"\t<mapLayer name=\"Above Water\">\n" +
		"\t\t<terrainAndLocation name=\"Classic/Water Sea\" elevation=\"-1\" icy=\"false\" gmOnly=\"true\" resources=\"Z\" location=\"1125.0,1650.0\" />\n" +
		"\t</mapLayer>\n" +
		"\t<mapLayer name=\"Mine &amp; &quot;Tunnels&quot;\">\n" +
		"\t</mapLayer>\n" +
		"</extraTerrain>\n"
	if got := string(extraTerrainElement.Find(ed.Utf8Encoded)); got != want {
		t.Errorf("<extraTerrain> =\n%s\nwant\n%s", got, want)
	}

	back, err := xmlio.NewDecoder().Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if !reflect.DeepEqual(back.ExtraTerrain, m.ExtraTerrain) {
		t.Errorf("re-decoded ExtraTerrain differs from the one encoded")
		for i, l := range back.ExtraTerrain.MapLayers {
			t.Logf("  layer %d %q: %d placement(s)", i, l.Name, len(l.Terrain))
			for _, tl := range l.Terrain {
				t.Logf("    %+v", *tl)
			}
		}
	}
}

// TestW2025ExtraTerrainFractionalElevationRefused: elevation is an integer on
// disk, so a fractional one is refused rather than rounded (issue #64's rule,
// applied to the attribute #34 adds).
func TestW2025ExtraTerrainFractionalElevationRefused(t *testing.T) {
	m := extraTerrainMap()
	m.ExtraTerrain.MapLayers[0].Terrain[0].Elevation = 12.5
	var buf bytes.Buffer
	err := xmlio.NewEncoder("2.06").Encode(&buf, m)
	if !errors.Is(err, wxx.ErrInvalidIntegerAttribute) {
		t.Fatalf("encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidIntegerAttribute)
	}
	if !strings.Contains(err.Error(), "terrainAndLocation/@elevation") {
		t.Errorf("encode: err = %q, want it to name the attribute", err)
	}
	if buf.Len() != 0 {
		t.Errorf("encode: wrote %d bytes, want 0", buf.Len())
	}
}

// TestW2025ExtraTerrainDecodeRefusals: content the model does not understand is
// refused by name rather than dropped, since before #34 this element survived
// a round trip verbatim whatever it held. Malformed compound values are refused
// too. Each document is doctored from the layers fixture, so the only thing
// wrong with it is the change under test.
func TestW2025ExtraTerrainDecodeRefusals(t *testing.T) {
	f, err := os.Open(sample2025_206LayersBeta)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	var dd xmlio.DecoderDiagnostics
	if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f); err != nil {
		t.Fatalf("decode: %v", err)
	}
	source := stripXMLDecl(dd.Converted)

	const placement = `resources="Z" location="225.0,150.0" />`
	for _, tc := range []struct {
		name, from, to, wantMsg string
	}{
		{"unknown child of <extraTerrain>", "<extraTerrain>", "<extraTerrain><terrainPath/>", "<terrainPath>"},
		{"unknown child of <mapLayer>", `<mapLayer name="Terrain Layer">`, `<mapLayer name="Terrain Layer"><note/>`, "<note>"},
		{"unknown attribute on <mapLayer>", `<mapLayer name="Terrain Layer">`, `<mapLayer name="Terrain Layer" opacity="0.5">`, "@opacity"},
		{"unknown attribute on <terrainAndLocation>", placement, `resources="Z" location="225.0,150.0" rotate="90" />`, "@rotate"},
		{"resources neither Z nor seven values", placement, `resources="1,2,3" location="225.0,150.0" />`, "@resources"},
		{"location without a comma", placement, `resources="Z" location="225.0" />`, "@location"},
		{"fractional elevation", `elevation="1000"`, `elevation="1000.0"`, "elevation"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !bytes.Contains(source, []byte(tc.from)) {
				t.Fatalf("fixture does not contain %q; this case is doctoring the wrong text", tc.from)
			}
			doctored := bytes.Replace(source, []byte(tc.from), []byte(tc.to), 1)
			_, err := v1_06.Decode(doctored)
			if err == nil {
				t.Fatalf("decode: want an error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("decode: err = %q, want it to name %q", err, tc.wantMsg)
			}
		})
	}
}

// TestClassicDowngradeLabelDropShadowLatent covers the half of the label
// drop-shadow loss no tracked .wxx can demonstrate: top-level <labels>. Only
// the layers fixture has labels, and they belong to features. The source is
// synthesized, as TestClassicDowngradeScrollbarLatent's is.
func TestClassicDowngradeLabelDropShadowLatent(t *testing.T) {
	m := decodeW2025(t, sample2025_206)
	if len(m.Labels) != 0 {
		t.Fatalf("%s carries %d label(s); this test synthesizes the only one", sample2025_206, len(m.Labels))
	}
	m.Labels = []*wxx.Label_t{{
		MapLayer: "Labels", Style: "Nation", FontFace: "Arial",
		Color: &wxx.RGBA_t{A: 1}, OutlineColor: &wxx.RGBA_t{A: 1},
		DropShadowColor: "0.0,0.0,0.0,1.0", DropShadowRadius: 3, DropShadowSpread: 1,
		Location:  &wxx.LabelLocation_t{ViewLevel: "WORLD", X: 10, Y: 20, Scale: 12.5},
		InnerText: "Kingdom",
	}}
	var d xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(classicTarget, xmlio.WithEncoderDiagnostics(&d)).Encode(&buf, m); err != nil {
		t.Fatalf("encode -> classic: %v", err)
	}
	for _, e := range d.Dropped {
		if e.Path == "map/labels/label/@dropShadow*" {
			for _, want := range []string{`"Kingdom"`, "color=0.0,0.0,0.0,1.0", "radius=3", "spread=1"} {
				if !strings.Contains(e.Detail, want) {
					t.Errorf("Detail = %q, want it to contain %q", e.Detail, want)
				}
			}
			return
		}
	}
	t.Errorf("no map/labels/label/@dropShadow* entry reported; the label's drop shadow is lost silently")
}
