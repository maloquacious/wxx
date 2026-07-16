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
// map's identity looked correct from outside. Reading m1.Version to choose the
// target is fine HERE and is what a client does; what the encoder may not do is
// read it for us.
func w2025Recode(t *testing.T, m1 *wxx.Map_t) *wxx.Map_t {
	t.Helper()
	xmlBytes, err := v1_06.Encode(m1, m1.Version)
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

	blankBytes, err := v1_06.Encode(b1, b1.Version)
	if err != nil {
		t.Fatalf("v1_06.Encode(cleared): %v", err)
	}
	if bytes.Contains(blankBytes, []byte("dropShadowColor")) {
		t.Errorf("re-encode spuriously added dropShadowColor (source had none)")
	}

	// Populated fixture: source has dropShadowColor -> output must preserve it.
	p1 := decodeFixture(t, populatedFixture)
	popBytes, err := v1_06.Encode(p1, p1.Version)
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
