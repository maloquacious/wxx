// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// The W2025 samples we have on disk, with their true on-disk map metadata
// (release/version/schema). The file name records the map's own version
// attribute, which is not necessarily the version the application reports.
// 2.06 is the first supported W2025 build; earlier builds are out of scope.
const (
	// sample2025_206 is the baseline: a blank 13x11 map.
	sample2025_206 = "../testdata/2025-2.06-13x11-941577-blank.wxx" // release=2025 version=2.06 schema=1.06
	// sample2025_206Layers is the same build carrying labels, locations,
	// map layers and terrain-and-location entries.
	sample2025_206Layers = "../testdata/2025-2.06-13x11-941577-layers.wxx" // release=2025 version=2.06 schema=1.06
)

// TestW2025Decode_BothSamples documents that the public decoder accepts both
// shipped W2025 samples.
func TestW2025Decode_BothSamples(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{"2.06/1.06 blank", sample2025_206},
		{"2.06/1.06 layers", sample2025_206Layers},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("decode %s: %v", tc.path, err)
			}
			if m.Tiles == nil {
				t.Fatalf("decode %s: nil Tiles", tc.path)
			}
			if m.TerrainMap == nil || len(m.TerrainMap.List) == 0 {
				t.Fatalf("decode %s: empty TerrainMap", tc.path)
			}
		})
	}
}

// TestW2025RoundTrip exercises the codec core on its own, calling v1_06
// directly rather than through the public dispatch: decode a real file,
// v1_06.Encode it back to XML, v1_06.Decode that XML, and assert the two
// Map_t values are semantically equal. Any fidelity loss in encode/decode
// surfaces as a per-group mismatch, unmixed with transport concerns.
// TestW2025PublicRoundTrip covers the same ground through MarshalXML and the
// gzip/UTF-16/header layers.
func TestW2025RoundTrip(t *testing.T) {
	m1, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("initial decode: %v", err)
	}

	xmlBytes, err := v1_06.Encode(m1, m1.Version)
	if err != nil {
		t.Fatalf("v1_06.Encode: %v", err)
	}

	m2, err := v1_06.Decode(xmlBytes)
	if err != nil {
		t.Fatalf("v1_06.Decode(re-encoded): %v\n---encoded xml (first 800 bytes)---\n%s", err, head(xmlBytes, 800))
	}

	normalizeVolatile(m1)
	normalizeVolatile(m2)

	compareGroups(t, m1, m2)
}

// TestW2025PublicRoundTrip exercises the entire public pipeline end to end:
// decode a real .wxx file, encode it back through xmlio.NewEncoder(app).Encode
// (XML + header + UTF-16BE + gzip), then decode those bytes with
// xmlio.NewDecoder().Decode and assert semantic equality. Unlike
// TestW2025RoundTrip (which drives only the in-memory XML codec), this proves
// the gzip/UTF-16/header transport layers round-trip too.
func TestW2025PublicRoundTrip(t *testing.T) {
	m1, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("initial decode: %v", err)
	}

	// The target is the version the fixture states: a round trip writes back what
	// it read. Since issue #45 the caller says so rather than the encoder assuming
	// it -- reading provenance and choosing a target is a CLIENT's job.
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(m1.MetaData.Version.App.Raw).Encode(&buf, m1); err != nil {
		t.Fatalf("public Encode: %v", err)
	}

	m2, err := xmlio.NewDecoder().Decode(&buf)
	if err != nil {
		t.Fatalf("public Decode(re-encoded): %v", err)
	}

	normalizeVolatile(m1)
	normalizeVolatile(m2)

	compareGroups(t, m1, m2)
}

// populatedFixture is a UTF-8 W2025 map that fills the elements the blank
// sample leaves empty (features with and without labels, shapes with points,
// and notes). It is a raw .xml file, so it is decoded with v1_06.Decode
// directly rather than through the full gunzip/UTF-16 public pipeline.
const populatedFixture = "../testdata/w2025-populated.xml"

// decodeFixture reads a raw UTF-8 XML W2025 map and decodes it with the
// schema-specific decoder (v1_06.Decode), bypassing the gunzip/UTF-16
// transport that decodeFile applies to .wxx files.
func decodeFixture(t *testing.T, path string) *wxx.Map_t {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m, err := v1_06.Decode(raw)
	if err != nil {
		t.Fatalf("v1_06.Decode(%s): %v", path, err)
	}
	return m
}

// TestW2025DecodePopulated asserts that the decoder materializes the populated
// fixture's content into Map_t using only fields that exist today. It guards
// against decode-side loss that a symmetric round-trip cannot catch.
func TestW2025DecodePopulated(t *testing.T) {
	m := decodeFixture(t, populatedFixture)

	// Shapes: both fixture shapes and their points must survive decode.
	if got, want := len(m.Shapes), 2; got != want {
		t.Fatalf("len(Shapes) = %d, want %d", got, want)
	}
	if len(m.Shapes[0].Points) == 0 {
		t.Fatalf("Shapes[0].Points is empty, want non-empty")
	}
	if got := m.Shapes[0].Points[0]; got.X != 148.0 || got.Y != 149.0 {
		t.Errorf("Shapes[0].Points[0] = (%v,%v), want (148,149)", got.X, got.Y)
	}

	// Notes: both fixture notes and their attributes/content must survive decode.
	if got, want := len(m.Notes), 2; got != want {
		t.Fatalf("len(Notes) = %d, want %d", got, want)
	}
	if got, want := m.Notes[0].Key, "WORLD,2343.75,3112.5"; got != want {
		t.Errorf("Notes[0].Key = %q, want %q", got, want)
	}
	if got, want := m.Notes[0].Title, "Units"; got != want {
		t.Errorf("Notes[0].Title = %q, want %q", got, want)
	}
	if got, want := "First note paragraph.", m.Notes[0].NoteText; !strings.Contains(m.Notes[0].NoteText, got) {
		t.Errorf("Notes[0].NoteText = %q, want it to contain %q", want, got)
	}
	if got, want := m.Notes[1].Key, "WORLD,100.0,200.0"; got != want {
		t.Errorf("Notes[1].Key = %q, want %q", got, want)
	}
	if got, want := m.Notes[1].Title, "Landmark"; got != want {
		t.Errorf("Notes[1].Title = %q, want %q", got, want)
	}
	if !m.Notes[1].IsGMOnly {
		t.Errorf("Notes[1].IsGMOnly = false, want true")
	}
	if !strings.Contains(m.Notes[1].NoteText, "Second note paragraph.") {
		t.Errorf("Notes[1].NoteText = %q, want it to contain %q", m.Notes[1].NoteText, "Second note paragraph.")
	}

	// Features: fixture has exactly two, [0] labeled, [1] labelless.
	if got, want := len(m.Features), 2; got != want {
		t.Fatalf("len(Features) = %d, want %d", got, want)
	}
	// The opaque-black feature color folds to nil (decodeRgba contract).
	if m.Features[0].Color != nil {
		t.Errorf("Features[0].Color = %+v, want nil (opaque black folds to nil)", m.Features[0].Color)
	}
	// The labelless feature must decode with a nil Label.
	if m.Features[1].Label != nil {
		t.Errorf("Features[1].Label = %+v, want nil (feature has no <label> child)", m.Features[1].Label)
	}
}

// TestW2025PopulatedRoundTrip drives the in-memory XML codec over the populated
// fixture: decode -> encode -> decode, then compares the two Map_t values group
// by group. Any encoder that drops content (shapes, notes) surfaces as a
// per-group mismatch naming the exact field.
func TestW2025PopulatedRoundTrip(t *testing.T) {
	raw, err := os.ReadFile(populatedFixture)
	if err != nil {
		t.Fatalf("read %s: %v", populatedFixture, err)
	}
	m1, err := v1_06.Decode(raw)
	if err != nil {
		t.Fatalf("initial decode: %v", err)
	}

	xmlBytes, err := v1_06.Encode(m1, m1.Version)
	if err != nil {
		t.Fatalf("v1_06.Encode: %v", err)
	}

	m2, err := v1_06.Decode(xmlBytes)
	if err != nil {
		t.Fatalf("v1_06.Decode(re-encoded): %v\n---encoded xml (first 800 bytes)---\n%s", err, head(xmlBytes, 800))
	}

	normalizeVolatile(m1)
	normalizeVolatile(m2)

	compareGroups(t, m1, m2)
}

// TestW2025PopulatedPublicRoundTrip drives the ENTIRE public pipeline over the
// populated fixture's content: decode the fixture, then encode through
// xmlio.NewEncoder().Encode (XML + header + UTF-16BE + gzip) and decode those
// bytes back with xmlio.NewDecoder().Decode. Unlike TestW2025PopulatedRoundTrip
// (which drives only the in-memory XML codec), this proves the gzip/UTF-16/header
// transport layers round-trip populated shapes/notes/features/labels too -- the
// "full public pipeline" half of the issue's definition of done.
func TestW2025PopulatedPublicRoundTrip(t *testing.T) {
	m1 := decodeFixture(t, populatedFixture)

	// The target is the version the fixture states: a round trip writes back what
	// it read. Since issue #45 the caller says so rather than the encoder assuming
	// it -- reading provenance and choosing a target is a CLIENT's job.
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(m1.MetaData.Version.App.Raw).Encode(&buf, m1); err != nil {
		t.Fatalf("public Encode: %v", err)
	}

	m2, err := xmlio.NewDecoder().Decode(&buf)
	if err != nil {
		t.Fatalf("public Decode(re-encoded): %v", err)
	}

	normalizeVolatile(m1)
	normalizeVolatile(m2)

	compareGroups(t, m1, m2)
}

// TestW2025ConfigSectionsEmpty guards the intentional no-op encoders for the
// <terrain-config>, <feature-config>, and <texture-config> sections
// (encodeTerrainConfig / encodeFeatureConfig / encodeTextureConfig in
// xmlio/internal/v1_06/encode.go). Those encoders emit an empty wrapper and drop their
// decoded content; that is only lossless because real W2025 maps leave these
// sections empty. This test documents-in-code that invariant by asserting that
// every decoded config entry carries no non-whitespace content, for BOTH the
// populated fixture and the real .wxx sample. If a future fixture ever populates
// one of these sections, this test fails loudly, signaling that the encoders
// (and the corresponding schema.go `xml:",chardata"` fields) must be upgraded to
// preserve inner XML.
func TestW2025ConfigSectionsEmpty(t *testing.T) {
	sampleMap, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_206, err)
	}

	for _, tc := range []struct {
		name string
		m    *wxx.Map_t
	}{
		{"populated-fixture", decodeFixture(t, populatedFixture)},
		{"real-sample", sampleMap},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.m.Configuration
			if cfg == nil {
				t.Fatalf("Configuration is nil")
			}
			for i, e := range cfg.TerrainConfig {
				if got := strings.TrimSpace(e.InnerText); got != "" {
					t.Errorf("TerrainConfig[%d] has unexpected non-whitespace content: %q", i, got)
				}
			}
			for i, e := range cfg.FeatureConfig {
				if got := strings.TrimSpace(e.InnerText); got != "" {
					t.Errorf("FeatureConfig[%d] has unexpected non-whitespace content: %q", i, got)
				}
			}
			for i, e := range cfg.TextureConfig {
				if got := strings.TrimSpace(e.InnerText); got != "" {
					t.Errorf("TextureConfig[%d] has unexpected non-whitespace content: %q", i, got)
				}
			}
		})
	}
}

// decodeFile runs the full public decode pipeline (gunzip -> UTF-16BE -> XML)
// on a .wxx file.
func decodeFile(t *testing.T, path string) (*wxx.Map_t, error) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	return xmlio.NewDecoder().Decode(f)
}

// normalizeVolatile zeroes fields that legitimately differ between two decodes
// of the same content (wall-clock timestamps), so they don't mask real diffs.
func normalizeVolatile(m *wxx.Map_t) {
	m.MetaData.Created = ""
	m.MetaData.Worldographer.Created = m.MetaData.Worldographer.Created.UTC()
}

// compareGroups checks each top-level element group independently so a failure
// names exactly which part of the model lost fidelity.
func compareGroups(t *testing.T, a, b *wxx.Map_t) {
	t.Helper()
	groups := []struct {
		name string
		x, y any
	}{
		{"MetaData", a.MetaData, b.MetaData},
		{"map-attributes", mapAttrs(a), mapAttrs(b)},
		{"GridAndNumbering", a.GridAndNumbering, b.GridAndNumbering},
		{"TerrainMap", a.TerrainMap, b.TerrainMap},
		{"MapLayers", a.MapLayers, b.MapLayers},
		{"Tiles", a.Tiles, b.Tiles},
		{"MapKey", a.MapKey, b.MapKey},
		{"Features", a.Features, b.Features},
		{"Labels", a.Labels, b.Labels},
		{"Shapes", a.Shapes, b.Shapes},
		{"Notes", a.Notes, b.Notes},
		{"Informations", a.Informations, b.Informations},
		{"Configuration", a.Configuration, b.Configuration},
	}
	for _, g := range groups {
		if !reflect.DeepEqual(g.x, g.y) {
			path, _ := firstDiff(g.name, reflect.ValueOf(g.x), reflect.ValueOf(g.y))
			t.Errorf("group %q differs after round-trip at %s", g.name, path)
		}
	}
}

// firstDiff walks two values in lockstep and returns a path string describing
// the first place they differ, so a round-trip failure names the exact field
// (e.g. "MapKey.Viewlevel: null vs WORLD") instead of dumping whole structs.
func firstDiff(path string, a, b reflect.Value) (string, bool) {
	if !a.IsValid() || !b.IsValid() {
		if a.IsValid() != b.IsValid() {
			return fmt.Sprintf("%s: valid %v vs %v", path, a.IsValid(), b.IsValid()), true
		}
		return "", false
	}
	if a.Type() != b.Type() {
		return fmt.Sprintf("%s: type %s vs %s", path, a.Type(), b.Type()), true
	}
	switch a.Kind() {
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			if a.IsNil() != b.IsNil() {
				return fmt.Sprintf("%s: nil %v vs %v", path, a.IsNil(), b.IsNil()), true
			}
			return "", false
		}
		return firstDiff(path, a.Elem(), b.Elem())
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if a.Type().Field(i).PkgPath != "" {
				continue // skip unexported fields
			}
			if d, ok := firstDiff(path+"."+a.Type().Field(i).Name, a.Field(i), b.Field(i)); ok {
				return d, true
			}
		}
		return "", false
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: len %d vs %d", path, a.Len(), b.Len()), true
		}
		for i := 0; i < a.Len(); i++ {
			if d, ok := firstDiff(fmt.Sprintf("%s[%d]", path, i), a.Index(i), b.Index(i)); ok {
				return d, true
			}
		}
		return "", false
	case reflect.Map:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: map len %d vs %d", path, a.Len(), b.Len()), true
		}
		for _, k := range a.MapKeys() {
			if d, ok := firstDiff(fmt.Sprintf("%s[%v]", path, k), a.MapIndex(k), b.MapIndex(k)); ok {
				return d, true
			}
		}
		return "", false
	default:
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			return fmt.Sprintf("%s: %v vs %v", path, a.Interface(), b.Interface()), true
		}
		return "", false
	}
}

// mapAttrs bundles the scalar <map> attributes for comparison.
func mapAttrs(m *wxx.Map_t) map[string]any {
	return map[string]any{
		"Type": m.Type, "Release": m.Release, "Version": m.Version, "Schema": m.Schema,
		"HexOrientation": m.HexOrientation, "MapProjection": m.MapProjection,
		"HexWidth": m.HexWidth, "HexHeight": m.HexHeight,
		"ContinentFactor": m.ContinentFactor, "KingdomFactor": m.KingdomFactor, "ProvinceFactor": m.ProvinceFactor,
		"ShowGrid": m.ShowGrid, "ShowGridNumbers": m.ShowGridNumbers, "ShowNotes": m.ShowNotes,
		"RowsHigh": m.RowsHigh, "ColumnsWide": m.ColumnsWide,
	}
}

func head(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
