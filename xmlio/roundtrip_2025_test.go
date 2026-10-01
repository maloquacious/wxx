// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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
	// sample2025_206LayersBeta was written by a beta build of 2.06. It carries
	// labels, locations, map layers and terrain-and-location entries, including
	// the "Terrain Layer" map layer, which no production build ships.
	sample2025_206LayersBeta = "../testdata/2025-2.06-13x11-941577-layers-beta.wxx" // release=2025 version=2.06 schema=1.06
	// sample2025_207Blank is a blank 13x11 map saved by 2.07, the first stable
	// release of the W2025 schema.
	sample2025_207Blank = "../testdata/2025-2.07-13x11-941577-blank.wxx" // release=2025 version=2.07 schema=1.06
	// sample2025_207NotesShapes is the richest app-saved map (#93): features
	// with inline labels, shapes with points, notes, and <extraTerrain>
	// placements on three layers.
	sample2025_207NotesShapes = "../testdata/2025-2.07-13x11-941577-notes-shapes.wxx" // release=2025 version=2.07 schema=1.06
	// sample2025_207Layers, sample2025_207Resources and sample2025_207Rows are the
	// remaining 2.07 saves (#91): map layers, resources, and a rows-oriented map.
	sample2025_207Layers    = "../testdata/2025-2.07-13x11-941577-layers.wxx"    // release=2025 version=2.07 schema=1.06
	sample2025_207Resources = "../testdata/2025-2.07-13x11-941577-resources.wxx" // release=2025 version=2.07 schema=1.06
	sample2025_207Rows      = "../testdata/2025-2.07-13x11-941577-rows.wxx"      // release=2025 version=2.07 schema=1.06
	// The 2.08 saves (#91, #94), built from the same recipes as the 2.07 ones
	// plus the populated map, and tile-resources, which sets every field of one
	// tile record (#117).
	sample2025_208Blank         = "../testdata/2025-2.08-13x11-941577-blank.wxx"          // release=2025 version=2.08 schema=1.06
	sample2025_208Layers        = "../testdata/2025-2.08-13x11-941577-layers.wxx"         // release=2025 version=2.08 schema=1.06
	sample2025_208NotesShapes   = "../testdata/2025-2.08-13x11-941577-notes-shapes.wxx"   // release=2025 version=2.08 schema=1.06
	sample2025_208Populated     = "../testdata/2025-2.08-13x11-941577-populated.wxx"      // release=2025 version=2.08 schema=1.06
	sample2025_208Resources     = "../testdata/2025-2.08-13x11-941577-resources.wxx"      // release=2025 version=2.08 schema=1.06
	sample2025_208Rows          = "../testdata/2025-2.08-13x11-941577-rows.wxx"           // release=2025 version=2.08 schema=1.06
	sample2025_208TileResources = "../testdata/2025-2.08-13x11-941577-tile-resources.wxx" // release=2025 version=2.08 schema=1.06

	// w2025Target is the application version the tests encode 2.07 maps as: the
	// version they state, registered by issue #92.
	w2025Target = "2.07"
)

// sameVersionTarget returns the application version the *MatchSource tests
// encode fixture as: the version the fixture's file name states, so a byte
// comparison with the source is a same-version one. Every W2025 version with a
// tracked fixture is registered: 2.06, 2.07 (issue #92) and 2.08 (issue #73).
//
// It is keyed on the file name rather than on the decoded map, and the
// registered versions are listed rather than looked up, so that a version
// dropping out of the registry fails these tests instead of quietly falling
// back to 2.06.
func sameVersionTarget(t *testing.T, fixture string) string {
	t.Helper()
	base := filepath.Base(fixture)
	switch {
	case strings.HasPrefix(base, "2025-2.06-"):
		return "2.06"
	case strings.HasPrefix(base, "2025-2.07-"):
		return "2.07"
	case strings.HasPrefix(base, "2025-2.08-"):
		return "2.08"
	}
	t.Fatalf("%s: no target for this fixture's version", fixture)
	return ""
}

// fixtures207 is every tracked 2.07 save. Tests that loop over it assert they
// visited len(fixtures207) files, and TestRegisteredFixtureListsAreEveryTrackedFixture
// holds it to the testdata directory, so a fixture cannot be skipped unnoticed.
var fixtures207 = []string{
	sample2025_207Blank,
	sample2025_207Layers,
	sample2025_207NotesShapes,
	sample2025_207Resources,
	sample2025_207Rows,
}

// fixtures208 is every tracked 2.08 save, held to the testdata directory as
// fixtures207 is.
var fixtures208 = []string{
	sample2025_208Blank,
	sample2025_208Layers,
	sample2025_208NotesShapes,
	sample2025_208Populated,
	sample2025_208Resources,
	sample2025_208Rows,
	sample2025_208TileResources,
}

// TestW2025Decode_BothSamples documents that the public decoder accepts both
// shipped W2025 samples.
func TestW2025Decode_BothSamples(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{"2.06/1.06 blank", sample2025_206},
		{"2.06/1.06 layers beta", sample2025_206LayersBeta},
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

	xmlBytes, err := v1_06.Encode(m1, m1.MetaData.Version.App.Raw)
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

// TestW2025NotesShapesRoundTrip drives the in-memory XML codec over the
// richest app-saved 2.07 map: decode -> encode -> decode, then compares the two
// Map_t values group by group. The blank sample TestW2025RoundTrip reads has
// no features, labels, shapes or notes and an empty <extraTerrain>; this one
// has all of them, so an encoder that drops content from any of those groups
// surfaces here as a per-group mismatch naming the exact field.
//
// It is encoded as 2.07, the version it states (registered by issue #92), so
// the identity it writes back is the one it read and the MetaData group is
// compared with the rest.
func TestW2025NotesShapesRoundTrip(t *testing.T) {
	m1, err := decodeFile(t, sample2025_207NotesShapes)
	if err != nil {
		t.Fatalf("initial decode: %v", err)
	}
	requireNotesShapesContent(t, m1)

	xmlBytes, err := v1_06.Encode(m1, w2025Target)
	if err != nil {
		t.Fatalf("v1_06.Encode: %v", err)
	}

	m2, err := v1_06.Decode(xmlBytes)
	if err != nil {
		t.Fatalf("v1_06.Decode(re-encoded): %v\n---encoded xml (first 800 bytes)---\n%s", err, head(xmlBytes, 800))
	}

	normalizeVolatile(m1)
	normalizeVolatile(m2)
	sortTerrainList(m1)
	sortTerrainList(m2)

	compareGroups(t, m1, m2)
}

// TestW2025NotesShapesPublicRoundTrip drives the ENTIRE public pipeline over
// the same 2.07 map: encode through xmlio.NewEncoder().Encode (XML + header +
// UTF-16BE + gzip) and decode those bytes back with xmlio.NewDecoder().Decode.
// Unlike TestW2025NotesShapesRoundTrip (which drives only the in-memory XML
// codec), this proves the gzip/UTF-16/header transport layers round-trip
// shapes, notes, features and labels too.
func TestW2025NotesShapesPublicRoundTrip(t *testing.T) {
	m1, err := decodeFile(t, sample2025_207NotesShapes)
	if err != nil {
		t.Fatalf("initial decode: %v", err)
	}
	requireNotesShapesContent(t, m1)

	// The target is named by the caller (issue #45): 2.07, the version the
	// fixture states, so MetaData is compared with the rest.
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(w2025Target).Encode(&buf, m1); err != nil {
		t.Fatalf("public Encode: %v", err)
	}

	m2, err := xmlio.NewDecoder().Decode(&buf)
	if err != nil {
		t.Fatalf("public Decode(re-encoded): %v", err)
	}

	normalizeVolatile(m1)
	normalizeVolatile(m2)
	sortTerrainList(m1)
	sortTerrainList(m2)

	compareGroups(t, m1, m2)
}

// sortTerrainList puts TerrainMap.List in index order.
//
// Worldographer 2.07 saved the notes-shapes fixture's <terrainmap> out of
// index order ("Blank 0", "Classic/Water Sea 2", "Classic/Flat Farmland 1"),
// and the encoder writes the table in index order (issue #87), so the decoded
// List comes back reordered. The index each name maps to is unchanged, and
// that mapping is what these round trips compare. Whether the reordering
// matters to Worldographer has not been tested in the app.
func sortTerrainList(m *wxx.Map_t) {
	if m.TerrainMap == nil {
		return
	}
	slices.SortStableFunc(m.TerrainMap.List, func(a, b *wxx.Terrain_t) int {
		return a.Index - b.Index
	})
}

// requireNotesShapesContent fails the test when the notes-shapes fixture stops
// carrying the groups the round trips above exist to exercise. Without it, an
// empty group on both sides compares equal and the test passes having tested
// nothing.
func requireNotesShapesContent(t *testing.T, m *wxx.Map_t) {
	t.Helper()
	if len(m.Features) == 0 {
		t.Fatalf("%s: no features decoded", sample2025_207NotesShapes)
	}
	if m.Features[0].Label == nil {
		t.Fatalf("%s: Features[0] has no label", sample2025_207NotesShapes)
	}
	if len(m.Shapes) == 0 || len(m.Shapes[0].Points) == 0 {
		t.Fatalf("%s: no shape with points decoded", sample2025_207NotesShapes)
	}
	if len(m.Notes) == 0 {
		t.Fatalf("%s: no notes decoded", sample2025_207NotesShapes)
	}
	if m.ExtraTerrain == nil || len(m.ExtraTerrain.MapLayers) == 0 {
		t.Fatalf("%s: no <extraTerrain> layers decoded", sample2025_207NotesShapes)
	}
}

// TestW2025ConfigSectionsEmpty guards the intentional no-op encoders for the
// <terrain-config>, <feature-config>, and <texture-config> sections
// (encodeTerrainConfig / encodeFeatureConfig / encodeTextureConfig in
// xmlio/internal/v1_06/encode.go). Those encoders emit an empty wrapper and drop their
// decoded content; that is only lossless because real W2025 maps leave these
// sections empty. This test documents-in-code that invariant by asserting that
// every decoded config entry carries no non-whitespace content, for both the
// 2.06 blank sample and the 2.07 notes-shapes map. If a future fixture ever populates
// one of these sections, this test fails loudly, signaling that the encoders
// (and the corresponding schema.go `xml:",chardata"` fields) must be upgraded to
// preserve inner XML.
func TestW2025ConfigSectionsEmpty(t *testing.T) {
	sampleMap, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_206, err)
	}
	notesShapesMap, err := decodeFile(t, sample2025_207NotesShapes)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_207NotesShapes, err)
	}

	for _, tc := range []struct {
		name string
		m    *wxx.Map_t
	}{
		{"2.07-notes-shapes", notesShapesMap},
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
//
// The three identity attributes are not here, and they are not uncovered: they
// live in MetaData.Worldographer as the provenance they are (issue #45 Decision
// 9), which compareGroups compares as its own "MetaData" group. Map_t no longer
// carries a second copy of them among the fields an encoder reads, because a
// second copy among those fields is what issue #45 was.
func mapAttrs(m *wxx.Map_t) map[string]any {
	return map[string]any{
		"Type":           m.Type,
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
