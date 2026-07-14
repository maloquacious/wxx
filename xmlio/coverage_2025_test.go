// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/h2025v1"
)

// w2025Recode drives the in-memory XML codec once: encode m1 with h2025v1.Encode
// then decode those bytes with h2025v1.Decode, returning the round-tripped model.
// It is the mechanism the coverage assertions below use to prove that what decode
// read, encode wrote, and decode read back again.
func w2025Recode(t *testing.T, m1 *wxx.Map_t) *wxx.Map_t {
	t.Helper()
	xmlBytes, err := h2025v1.Encode(m1)
	if err != nil {
		t.Fatalf("h2025v1.Encode: %v", err)
	}
	m2, err := h2025v1.Decode(xmlBytes)
	if err != nil {
		t.Fatalf("h2025v1.Decode(re-encoded): %v", err)
	}
	return m2
}

// TestW2025CoverageMatrix mechanically locks in the current per-element codec
// behavior documented in xmlio/h2025v1/COVERAGE.md. For every element the matrix
// marks "implemented", it asserts that element counts and key field values
// survive a decode -> encode -> decode round-trip. Any future drift (a stubbed
// encoder half, a dropped field, a changed count) trips one of these assertions.
//
// This test intentionally does NOT assert on the six "Known un-modeled fields"
// (maplayer/@opacity, dropShadow*, lineCap/lineJoin on shapestyle,
// hScrollbarPos/vScrollbarPos, blurTerrainBG, extraTerrain): those are symmetric
// drops by design and are covered by COVERAGE.md's prose, not by round-trip
// equality.
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
