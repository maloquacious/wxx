// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// w2025Recode drives the in-memory XML codec once: encode m1 with v1_06.Encode
// then decode those bytes with v1_06.Decode, returning the round-tripped model.
// It is the mechanism the coverage assertions below use to prove that what decode
// read, encode wrote, and decode read back again.
//
// The map is encoded as the application version it already states, so this
// exercises the codec and not the app-version gate.
// TestCodecRejectsUnacceptedAppVersion covers the gate.
//
// Naming that version explicitly is the ONLY way to encode since issue #45: the
// codec derives every identity byte it writes from the app it is handed and reads
// none of them from the map, so the app is not optional and not inferable. It used
// to be inferable, and that was the bug -- Release_t.identify stamped the target's
// identity onto the map on every public path, which meant the codec emitting the
// map's identity looked correct from outside. Reading the source's own version to
// choose the target is fine HERE and is what a client does (every cmd/* tool does
// exactly this); what the ENCODER may not do is read it for us. That distinction
// is the whole of issue #45, and it is why the read is spelled out at the call
// site rather than hidden inside Encode.
//
// It reads MetaData.Version.App.Raw because that is where the file's stated
// application version lives now: the top-level Map_t.Version this used to read was
// a second copy of the same provenance, sitting among the fields an encoder reads,
// and it is deleted. Raw is verbatim -- "2.06", never a re-rendered "2.6" (ADR
// 0004 Decision 1).
func w2025Recode(t *testing.T, m1 *wxx.Map_t) *wxx.Map_t {
	t.Helper()
	xmlBytes, err := v1_06.Encode(m1, m1.MetaData.Version.App.Raw)
	if err != nil {
		t.Fatalf("v1_06.Encode: %v", err)
	}
	m2, err := v1_06.Decode(xmlBytes)
	if err != nil {
		t.Fatalf("v1_06.Decode(re-encoded): %v", err)
	}
	return m2
}

// TestW2025CoverageMatrix mechanically locks in the current per-element codec
// behavior documented in xmlio/internal/v1_06/COVERAGE.md. For every element the matrix
// marks "implemented", it asserts that element counts and key field values
// survive a decode -> encode -> decode round-trip. Any future drift (a stubbed
// encoder half, a dropped field, a changed count) trips one of these assertions.
//
// As of #11 the six formerly "un-modeled" W2025-native fields (maplayer/@opacity,
// labelstyle dropShadow*, shapestyle lineCap/lineJoin, map hScrollbarPos/
// vScrollbarPos, blurTerrainBG, extraTerrain) are modeled and wired through both
// decode and encode; this test now asserts each survives the 2025 round trip.
func TestW2025CoverageMatrix(t *testing.T) {
	// ---- Populated fixture: exercises features/labels/shapes/notes ----
	p1 := decodeFixture(t, populatedFixture)
	p2 := w2025Recode(t, p1)

	// features / feature / location / inline label (COVERAGE: implemented)
	if got, want := len(p2.Features), 2; got != want {
		t.Fatalf("populated: len(Features) = %d, want %d", got, want)
	}
	if got, want := len(p1.Features), len(p2.Features); got != want {
		t.Errorf("populated: Features count drifted across round-trip: %d -> %d", got, want)
	}
	if got, want := p2.Features[0].Type, "City"; got != want {
		t.Errorf("populated: Features[0].Type = %q, want %q", got, want)
	}
	if p2.Features[0].Location == nil {
		t.Fatalf("populated: Features[0].Location = nil, want non-nil")
	}
	if got := p2.Features[0].Location; got.X != 100.0 || got.Y != 200.0 {
		t.Errorf("populated: Features[0].Location = (%v,%v), want (100,200)", got.X, got.Y)
	}
	// Feature[0] carries an inline <label>; Feature[1] is labelless.
	if p2.Features[0].Label == nil {
		t.Fatalf("populated: Features[0].Label = nil, want non-nil (inline label present)")
	}
	if got, want := p2.Features[0].Label.InnerText, "Rivertown"; got != want {
		t.Errorf("populated: Features[0].Label.InnerText = %q, want %q", got, want)
	}
	if p2.Features[0].Label.Location == nil {
		t.Fatalf("populated: Features[0].Label.Location = nil, want non-nil")
	}
	if got, want := p2.Features[0].Label.Location.Scale, 33.0; got != want {
		t.Errorf("populated: Features[0].Label.Location.Scale = %v, want %v", got, want)
	}
	if got, want := p2.Features[1].Type, "Forest"; got != want {
		t.Errorf("populated: Features[1].Type = %q, want %q", got, want)
	}
	if p2.Features[1].Label != nil {
		t.Errorf("populated: Features[1].Label = %+v, want nil (labelless feature)", p2.Features[1].Label)
	}
	// Forest color "0.13,0.55,0.13,1.0" is a real non-black color and must survive.
	if p2.Features[1].Color == nil {
		t.Fatalf("populated: Features[1].Color = nil, want non-nil (0.13,0.55,0.13,1.0)")
	}
	if got := p2.Features[1].Color; got.R != 0.13 || got.G != 0.55 || got.B != 0.13 || got.A != 1.0 {
		t.Errorf("populated: Features[1].Color = %+v, want (0.13,0.55,0.13,1.0)", got)
	}

	// labels (standalone): empty in this fixture; must stay empty (implemented).
	if got, want := len(p2.Labels), 0; got != want {
		t.Errorf("populated: len(Labels) = %d, want %d", got, want)
	}

	// shapes / shape / points (COVERAGE: implemented)
	if got, want := len(p2.Shapes), 2; got != want {
		t.Fatalf("populated: len(Shapes) = %d, want %d", got, want)
	}
	if got, want := len(p1.Shapes), len(p2.Shapes); got != want {
		t.Errorf("populated: Shapes count drifted across round-trip: %d -> %d", got, want)
	}
	if len(p2.Shapes[0].Points) == 0 {
		t.Fatalf("populated: Shapes[0].Points empty, want non-empty")
	}
	if got := p2.Shapes[0].Points[0]; got.X != 148.0 || got.Y != 149.0 {
		t.Errorf("populated: Shapes[0].Points[0] = (%v,%v), want (148,149)", got.X, got.Y)
	}

	// notes / note / notetext (COVERAGE: implemented)
	if got, want := len(p2.Notes), 2; got != want {
		t.Fatalf("populated: len(Notes) = %d, want %d", got, want)
	}
	if got, want := p2.Notes[0].Title, "Units"; got != want {
		t.Errorf("populated: Notes[0].Title = %q, want %q", got, want)
	}
	if got, want := p2.Notes[0].Key, "WORLD,2343.75,3112.5"; got != want {
		t.Errorf("populated: Notes[0].Key = %q, want %q", got, want)
	}
	if !strings.Contains(p2.Notes[0].NoteText, "First note paragraph.") {
		t.Errorf("populated: Notes[0].NoteText = %q, want it to contain %q", p2.Notes[0].NoteText, "First note paragraph.")
	}
	if !p2.Notes[1].IsGMOnly {
		t.Errorf("populated: Notes[1].IsGMOnly = false, want true")
	}
	if !strings.Contains(p2.Notes[1].NoteText, "Second note paragraph.") {
		t.Errorf("populated: Notes[1].NoteText = %q, want it to contain %q", p2.Notes[1].NoteText, "Second note paragraph.")
	}

	// ---- #11: the six formerly un-modeled W2025-native fields ----

	// maplayer/@opacity (now modeled): the first layer ("Labels") is opacity 1.0.
	if len(p2.MapLayers) == 0 {
		t.Fatalf("populated: len(MapLayers) = 0, want non-empty")
	}
	if got, want := len(p1.MapLayers), len(p2.MapLayers); got != want {
		t.Errorf("populated: MapLayers count drifted across round-trip: %d -> %d", got, want)
	}
	if got, want := p2.MapLayers[0].Opacity, 1.0; got != want {
		t.Errorf("populated: MapLayers[0].Opacity = %v, want %v", got, want)
	}

	// labelstyle dropShadow* (now modeled): Nation carries dropShadowColor="null"
	// (nullable string spelled "null") plus zero radius/spread.
	if p2.Configuration == nil || p2.Configuration.TextConfig == nil || len(p2.Configuration.TextConfig.LabelStyles) == 0 {
		t.Fatalf("populated: TextConfig.LabelStyles empty, want non-empty")
	}
	ls0 := p2.Configuration.TextConfig.LabelStyles[0]
	if got, want := ls0.DropShadowColor, "null"; got != want {
		t.Errorf("populated: LabelStyles[0].DropShadowColor = %q, want %q", got, want)
	}
	if got, want := ls0.DropShadowRadius, 0.0; got != want {
		t.Errorf("populated: LabelStyles[0].DropShadowRadius = %v, want %v", got, want)
	}
	if got, want := ls0.DropShadowSpread, 0.0; got != want {
		t.Errorf("populated: LabelStyles[0].DropShadowSpread = %v, want %v", got, want)
	}

	// shapestyle lineCap/lineJoin (now modeled): Trail is lineCap="SQUARE"
	// lineJoin="ROUND".
	if p2.Configuration.ShapeConfig == nil || len(p2.Configuration.ShapeConfig.ShapeStyles) == 0 {
		t.Fatalf("populated: ShapeConfig.ShapeStyles empty, want non-empty")
	}
	ss0 := p2.Configuration.ShapeConfig.ShapeStyles[0]
	if got, want := ss0.LineCap, "SQUARE"; got != want {
		t.Errorf("populated: ShapeStyles[0].LineCap = %q, want %q", got, want)
	}
	if got, want := ss0.LineJoin, "ROUND"; got != want {
		t.Errorf("populated: ShapeStyles[0].LineJoin = %q, want %q", got, want)
	}

	// map hScrollbarPos/vScrollbarPos (now modeled): 0.0 in the fixture, and must
	// not drift across the round-trip.
	if got, want := p2.HScrollbarPos, 0.0; got != want {
		t.Errorf("populated: HScrollbarPos = %v, want %v", got, want)
	}
	if got, want := p2.VScrollbarPos, 0.0; got != want {
		t.Errorf("populated: VScrollbarPos = %v, want %v", got, want)
	}
	if p1.HScrollbarPos != p2.HScrollbarPos || p1.VScrollbarPos != p2.VScrollbarPos {
		t.Errorf("populated: scrollbar positions drifted across round-trip: (%v,%v) -> (%v,%v)",
			p1.HScrollbarPos, p1.VScrollbarPos, p2.HScrollbarPos, p2.VScrollbarPos)
	}

	// blurTerrainBG (now modeled): present in the fixture, must be non-nil after
	// round-trip with its attributes preserved.
	if p2.BlurTerrainBG == nil {
		t.Fatalf("populated: BlurTerrainBG = nil, want non-nil (present in fixture)")
	}
	if got := p2.BlurTerrainBG; got.Blur != false ||
		got.TopBleed != 0.33 || got.BottomBleed != 0.65 || got.Randomness != 0.1 ||
		got.BlurStart != 0.4 || got.BlurEnd != 0.95 {
		t.Errorf("populated: BlurTerrainBG = %+v, want {Blur:false TopBleed:0.33 BottomBleed:0.65 Randomness:0.1 BlurStart:0.4 BlurEnd:0.95}", got)
	}

	// extraTerrain (now modeled): present-but-empty in the fixture, must be
	// non-nil after round-trip.
	if p2.ExtraTerrain == nil {
		t.Fatalf("populated: ExtraTerrain = nil, want non-nil (present in fixture)")
	}

	// config no-op(intentional) sub-sections must stay empty after round-trip.
	assertConfigSectionsEmpty(t, "populated", p2)

	// ---- Real sample: exercises the blank-but-rich elements ----
	s1, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_206, err)
	}
	s2 := w2025Recode(t, s1)

	// gridandnumbering (implemented)
	if s2.GridAndNumbering == nil {
		t.Fatalf("sample: GridAndNumbering = nil, want non-nil")
	}
	if got, want := s2.GridAndNumbering.NumberFont, s1.GridAndNumbering.NumberFont; got != want {
		t.Errorf("sample: GridAndNumbering.NumberFont = %q, want %q", got, want)
	}

	// terrainmap (implemented)
	if s2.TerrainMap == nil || len(s2.TerrainMap.List) == 0 {
		t.Fatalf("sample: TerrainMap empty, want non-empty")
	}
	if got, want := len(s2.TerrainMap.List), len(s1.TerrainMap.List); got != want {
		t.Errorf("sample: TerrainMap.List count drifted across round-trip: %d -> %d", got, want)
	}
	if got, want := len(s2.TerrainMap.Data), len(s1.TerrainMap.Data); got != want {
		t.Errorf("sample: TerrainMap.Data count drifted across round-trip: %d -> %d", got, want)
	}

	// maplayers (implemented; opacity un-modeled but name+isVisible survive)
	if got, want := len(s2.MapLayers), len(s1.MapLayers); got != want {
		t.Fatalf("sample: len(MapLayers) = %d, want %d", got, want)
	}
	for i := range s1.MapLayers {
		if got, want := s2.MapLayers[i].Name, s1.MapLayers[i].Name; got != want {
			t.Errorf("sample: MapLayers[%d].Name = %q, want %q", i, got, want)
		}
		if got, want := s2.MapLayers[i].IsVisible, s1.MapLayers[i].IsVisible; got != want {
			t.Errorf("sample: MapLayers[%d].IsVisible = %v, want %v", i, got, want)
		}
	}

	// tiles (implemented)
	if s2.Tiles == nil {
		t.Fatalf("sample: Tiles = nil, want non-nil")
	}
	if got, want := s2.Tiles.TilesWide, s1.Tiles.TilesWide; got != want {
		t.Errorf("sample: Tiles.TilesWide = %d, want %d", got, want)
	}
	if got, want := s2.Tiles.TilesHigh, s1.Tiles.TilesHigh; got != want {
		t.Errorf("sample: Tiles.TilesHigh = %d, want %d", got, want)
	}
	if got, want := len(s2.Tiles.Tiles), len(s1.Tiles.Tiles); got != want {
		t.Errorf("sample: len(Tiles.Tiles) = %d, want %d", got, want)
	}

	// mapkey (implemented)
	if s2.MapKey == nil {
		t.Fatalf("sample: MapKey = nil, want non-nil")
	}
	if got, want := s2.MapKey.Viewlevel, s1.MapKey.Viewlevel; got != want {
		t.Errorf("sample: MapKey.Viewlevel = %q, want %q", got, want)
	}

	// informations / information / nested detail (implemented)
	if s2.Informations == nil {
		t.Fatalf("sample: Informations = nil, want non-nil")
	}
	if got, want := len(s2.Informations.Informations), len(s1.Informations.Informations); got != want {
		t.Fatalf("sample: Informations count drifted across round-trip: %d -> %d", got, want)
	}
	if len(s1.Informations.Informations) == 0 {
		t.Fatalf("sample: expected a non-empty <informations> tree")
	}

	// text-config / labelstyle (implemented)
	if s2.Configuration == nil || s2.Configuration.TextConfig == nil {
		t.Fatalf("sample: Configuration.TextConfig = nil, want non-nil")
	}
	if got, want := len(s2.Configuration.TextConfig.LabelStyles), len(s1.Configuration.TextConfig.LabelStyles); got != want {
		t.Errorf("sample: labelstyle count drifted across round-trip: %d -> %d", got, want)
	}
	if len(s1.Configuration.TextConfig.LabelStyles) == 0 {
		t.Errorf("sample: expected non-empty labelstyles")
	}

	// shape-config / shapestyle (implemented)
	if s2.Configuration.ShapeConfig == nil {
		t.Fatalf("sample: Configuration.ShapeConfig = nil, want non-nil")
	}
	if got, want := len(s2.Configuration.ShapeConfig.ShapeStyles), len(s1.Configuration.ShapeConfig.ShapeStyles); got != want {
		t.Errorf("sample: shapestyle count drifted across round-trip: %d -> %d", got, want)
	}
	if len(s1.Configuration.ShapeConfig.ShapeStyles) == 0 {
		t.Errorf("sample: expected non-empty shapestyles")
	}

	// config no-op(intentional) sub-sections must stay empty after round-trip.
	assertConfigSectionsEmpty(t, "sample", s2)
}

// TestW2025LabelStyleDropShadowGate guards the presence-gated emission of the
// W2025 <labelstyle> drop-shadow trio at the XML byte level (which the Map_t
// round-trip cannot catch, since absent decodes to the same zero values a
// symmetric drop would produce). The encoder keys the gate off DropShadowColor,
// which is "null" or an RGBA string when present and never empty, so clearing it
// models a source that carried no drop shadow and must suppress the trio; the
// populated fixture DOES carry them, so re-encoding it must preserve them.
//
// The negative case is synthesized rather than read from a fixture: every
// supported W2025 build (2.06 and later) writes the trio, so no sample supplies
// a drop-shadow-free source directly.
func TestW2025LabelStyleDropShadowGate(t *testing.T) {
	// Synthesized: decode the baseline, then clear the trio it carries.
	b1, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_206, err)
	}
	if b1.Configuration == nil || b1.Configuration.TextConfig == nil {
		t.Fatalf("decode %s: no TextConfig to clear", sample2025_206)
	}
	var cleared int
	for _, ls := range b1.Configuration.TextConfig.LabelStyles {
		if ls.DropShadowColor != "" {
			cleared++
		}
		ls.DropShadowColor, ls.DropShadowRadius, ls.DropShadowSpread = "", 0, 0
	}
	// Guard against a vacuous pass: if the baseline ever stops carrying the trio,
	// clearing is a no-op and the assertion below proves nothing.
	if cleared == 0 {
		t.Fatalf("%s: no label style carried dropShadowColor, so the gate is not under test", sample2025_206)
	}

	blankBytes, err := v1_06.Encode(b1, b1.MetaData.Version.App.Raw)
	if err != nil {
		t.Fatalf("v1_06.Encode(cleared): %v", err)
	}
	if bytes.Contains(blankBytes, []byte("dropShadowColor")) {
		t.Errorf("re-encode spuriously added dropShadowColor (source had none)")
	}

	// Populated fixture: source has dropShadowColor -> output must preserve it.
	p1 := decodeFixture(t, populatedFixture)
	popBytes, err := v1_06.Encode(p1, p1.MetaData.Version.App.Raw)
	if err != nil {
		t.Fatalf("v1_06.Encode(populated): %v", err)
	}
	if !bytes.Contains(popBytes, []byte("dropShadowColor")) {
		t.Errorf("populated re-encode dropped dropShadowColor (source had it); gate over-corrected")
	}
}

// assertConfigSectionsEmpty mirrors TestW2025ConfigSectionsEmpty's intent: the
// terrain-config / feature-config / texture-config encoders are documented
// no-op(intentional) sections that emit an empty wrapper. After a round-trip each
// decoded entry must carry no non-whitespace content, so an accidental change to
// that behavior trips.
func assertConfigSectionsEmpty(t *testing.T, label string, m *wxx.Map_t) {
	t.Helper()
	if m.Configuration == nil {
		t.Fatalf("%s: Configuration = nil, want non-nil", label)
	}
	for i, e := range m.Configuration.TerrainConfig {
		if got := strings.TrimSpace(e.InnerText); got != "" {
			t.Errorf("%s: TerrainConfig[%d] unexpected content after round-trip: %q", label, i, got)
		}
	}
	for i, e := range m.Configuration.FeatureConfig {
		if got := strings.TrimSpace(e.InnerText); got != "" {
			t.Errorf("%s: FeatureConfig[%d] unexpected content after round-trip: %q", label, i, got)
		}
	}
	for i, e := range m.Configuration.TextureConfig {
		if got := strings.TrimSpace(e.InnerText); got != "" {
			t.Errorf("%s: TextureConfig[%d] unexpected content after round-trip: %q", label, i, got)
		}
	}
}

// TestW2025LabelDropShadowRoundTrip pins issue #35: W2025 writes the drop-shadow
// trio on the <label> element itself, and the codec dropped it on a SAME-release
// 2.06 -> 2.06 round trip -- no downgrade involved. The layers fixture carries 3
// labels, all with the trio, so re-encoding it must reproduce all 3.
//
// Every label in this fixture is a FEATURE label (<feature><label>), not a
// standalone <labels><label> entry, so this exercises the features.go decode
// path; the encode side is shared (encodeFeatureLabel delegates to encodeLabel).
func TestW2025LabelDropShadowRoundTrip(t *testing.T) {
	m1, err := decodeFile(t, sample2025_206Layers)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_206Layers, err)
	}

	// Collect every Label_t the map carries, from both carriers.
	var labels []*wxx.Label_t
	labels = append(labels, m1.Labels...)
	for _, f := range m1.Features {
		if f.Label != nil {
			labels = append(labels, f.Label)
		}
	}

	// Guard against a vacuous pass: a loop over zero labels, or over labels that
	// carry nothing, would pass while proving nothing.
	if len(labels) == 0 {
		t.Fatalf("%s: no labels decoded, so the round trip is not under test", sample2025_206Layers)
	}
	var carrying int
	for _, l := range labels {
		if l.DropShadowColor != "" {
			carrying++
		}
	}
	if carrying == 0 {
		t.Fatalf("%s: decoded %d label(s) but none carried DropShadowColor; decode is dropping the trio",
			sample2025_206Layers, len(labels))
	}

	out, err := v1_06.Encode(m1, m1.MetaData.Version.App.Raw)
	if err != nil {
		t.Fatalf("v1_06.Encode: %v", err)
	}

	// The fixture's labels all spell the trio the same way; assert the emitted
	// bytes reproduce it once per carrying label. Scope the search to <label>
	// start tags: a whole-document search would also match <labelstyle>, which
	// carries a trio of its own (#11) and would mask a drop here.
	const want = `dropShadowColor="null" dropShadowRadius="0.0" dropShadowSpread="0.0"`
	tags := labelTags(out)
	if len(tags) != len(labels) {
		t.Fatalf("re-encode emitted %d <label> tags, want %d", len(tags), len(labels))
	}
	var emitted int
	for _, tag := range tags {
		if strings.Contains(tag, want) {
			emitted++
		}
	}
	if emitted != carrying {
		t.Errorf("re-encode emitted the trio on %d of %d <label> tags, want %d (one per carrying label)\nfirst tag: %s",
			emitted, len(tags), carrying, tags[0])
	}

	// And the trio must survive a decode of the re-encoded bytes, so the loss is
	// pinned at the model level too, not just in the byte stream.
	m2, err := v1_06.Decode(out)
	if err != nil {
		t.Fatalf("v1_06.Decode(re-encoded): %v", err)
	}
	var got2 int
	for _, f := range m2.Features {
		if f.Label != nil && f.Label.DropShadowColor != "" {
			got2++
		}
	}
	for _, l := range m2.Labels {
		if l.DropShadowColor != "" {
			got2++
		}
	}
	if got2 != carrying {
		t.Errorf("after round trip %d label(s) carry DropShadowColor, want %d", got2, carrying)
	}
}

// TestW2025LabelDropShadowCarriesValues proves the encoder READS the trio from
// the model rather than synthesizing it. Every label in the layers fixture spells
// the trio with the "no shadow" DEFAULTS ("null" / 0.0 / 0.0), so a fixture-only
// round trip would pass even against an encoder that hardcoded those defaults and
// consulted the model not at all. This drives NON-DEFAULT values -- a real RGBA
// colour and two distinct non-zero numerics -- so only a carried value can pass.
func TestW2025LabelDropShadowCarriesValues(t *testing.T) {
	label := &wxx.Label_t{
		MapLayer:        "Features",
		Style:           "City",
		FontFace:        "Arial",
		OutlineSize:     2.0,
		Rotate:          0.0,
		DropShadowColor: "0.25,0.5,0.75,1.0",
		// Deliberately distinct from each other and from the defaults.
		DropShadowRadius: 4.5,
		DropShadowSpread: 12.25,
		Location:         &wxx.LabelLocation_t{ViewLevel: "WORLD", X: 1.0, Y: 2.0, Scale: 1.0},
	}
	m := labelOnlyMap(t, label)

	out, err := v1_06.Encode(m, m.MetaData.Version.App.Raw)
	if err != nil {
		t.Fatalf("v1_06.Encode: %v", err)
	}

	// Scope every assertion to the <label> tag: the blank fixture's text-config
	// carries <labelstyle> elements whose own trio spells the defaults, so a
	// whole-document search would find "null" no matter what the encoder did.
	tag := onlyLabelTag(t, out)
	const want = `dropShadowColor="0.25,0.5,0.75,1.0" dropShadowRadius="4.5" dropShadowSpread="12.25"`
	if !strings.Contains(tag, want) {
		t.Errorf("encoded <label> does not carry the model's drop shadow.\nwant substring: %s\ngot <label>: %s",
			want, tag)
	}

	// The defaults must NOT appear on the label: their presence would mean the
	// encoder wrote a synthesized shadow instead of the one the model holds.
	if strings.Contains(tag, `dropShadowColor="null"`) {
		t.Errorf("encoded <label> contains a synthesized default dropShadowColor=\"null\"; the model said %q\ngot <label>: %s",
			label.DropShadowColor, tag)
	}
}

// TestW2025LabelDropShadowGate guards the presence gate at the XML byte level,
// which the Map_t round trip cannot catch: absent decodes to the same zero values
// a symmetric drop would produce. The encoder keys the gate off DropShadowColor,
// which is "null" or an RGBA string when present and never empty, so an empty one
// models a source that carried no drop shadow and must suppress the whole trio.
// Absent must stay absent -- never synthesize what was not on input.
//
// Note "" and "null" are different things: "" means the attribute was absent from
// the source; "null" means it was present, spelled with the four characters null.
func TestW2025LabelDropShadowGate(t *testing.T) {
	// A label with no drop shadow at all: emit none of the trio.
	label := &wxx.Label_t{
		MapLayer:        "Features",
		Style:           "City",
		FontFace:        "Arial",
		OutlineSize:     2.0,
		DropShadowColor: "", // absent from the source
		Location:        &wxx.LabelLocation_t{ViewLevel: "WORLD", X: 1.0, Y: 2.0, Scale: 1.0},
	}
	m := labelOnlyMap(t, label)

	out, err := v1_06.Encode(m, m.MetaData.Version.App.Raw)
	if err != nil {
		t.Fatalf("v1_06.Encode: %v", err)
	}
	tag := onlyLabelTag(t, out)
	for _, attr := range []string{"dropShadowColor", "dropShadowRadius", "dropShadowSpread"} {
		if strings.Contains(tag, attr) {
			t.Errorf("re-encode spuriously added %s to <label> (source had no drop shadow)\ngot <label>: %s",
				attr, tag)
		}
	}

	// The gate keys on the colour and never on the numerics: 0 is a legal radius
	// and spread, so a present-but-zero shadow must still be emitted.
	label.DropShadowColor, label.DropShadowRadius, label.DropShadowSpread = "null", 0, 0
	out, err = v1_06.Encode(m, m.MetaData.Version.App.Raw)
	if err != nil {
		t.Fatalf("v1_06.Encode(zero-valued): %v", err)
	}
	const want = `dropShadowColor="null" dropShadowRadius="0.0" dropShadowSpread="0.0"`
	if tag := onlyLabelTag(t, out); !strings.Contains(tag, want) {
		t.Errorf("gate over-corrected: a present shadow with zero numerics was suppressed.\nwant substring: %s\ngot <label>: %s",
			want, tag)
	}
}

// labelOnlyMap builds the smallest encodable W2025 map carrying exactly one
// standalone label, by starting from the blank fixture (so every unrelated
// required field is real) and replacing its labels.
func labelOnlyMap(t *testing.T, label *wxx.Label_t) *wxx.Map_t {
	t.Helper()
	m, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_206, err)
	}
	m.Labels = []*wxx.Label_t{label}
	m.Features = nil
	return m
}

// labelTags returns every <label ...> start tag in the encoded output.
//
// It anchors on the tag NAME being terminated by whitespace, so it matches
// <label ...> but never <labelstyle ...> -- which carries a drop-shadow trio of
// its own (#11). Searching the whole document for "dropShadowColor" would hit the
// labelstyle trio and pass regardless of what the <label> encoder did, so every
// assertion in these tests is scoped to what this returns.
func labelTags(out []byte) []string {
	var tags []string
	for rest := out; ; {
		i := bytes.Index(rest, []byte("<label"))
		if i == -1 {
			return tags
		}
		rest = rest[i:]
		// The tag name must end here; "<labelstyle" must not match.
		name := rest[len("<label"):]
		if len(name) == 0 {
			return tags
		}
		if c := name[0]; c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			rest = rest[len("<label"):]
			continue
		}
		j := bytes.IndexByte(rest, '>')
		if j == -1 {
			return append(tags, string(rest))
		}
		tags = append(tags, string(rest[:j+1]))
		rest = rest[j+1:]
	}
}

// onlyLabelTag returns the single <label> start tag the synthesized one-label maps
// are expected to produce, failing if there is not exactly one.
func onlyLabelTag(t *testing.T, out []byte) string {
	t.Helper()
	tags := labelTags(out)
	if len(tags) != 1 {
		t.Fatalf("encoded output has %d <label> tags, want exactly 1", len(tags))
	}
	return tags[0]
}
