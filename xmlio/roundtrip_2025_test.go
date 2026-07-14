// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/h2025v1"
)

// The W2025 samples we have on disk, with their true on-disk map metadata
// (release/version/schema) as opposed to the possibly-misleading file name.
const (
	// sample2025_206 decodes today: the dispatcher accepts 2025/2.06/1.06.
	sample2025_206 = "../data/2025-2.05.wxx" // internally release=2025 version=2.06 schema=1.06
	// sample2025_110 is a second, older W2025 build the dispatcher rejects today.
	sample2025_110 = "../testdata/input/blank-2025-1.10-1.01.wxx" // release=2025 version=1.10 schema=1.01
)

// TestW2025Decode_BothSamples documents which W2025 builds the public decoder
// accepts. Until the release-based dispatch lands, the 1.10/1.01 sample fails.
func TestW2025Decode_BothSamples(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{"2.06/1.06", sample2025_206},
		{"1.10/1.01", sample2025_110},
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

// TestW2025RoundTrip exercises the codec core independent of the (currently
// 2017-only) MarshalXML dispatch: decode a real file, h2025v1.Encode it back to
// XML, h2025v1.Decode that XML, and assert the two Map_t values are
// semantically equal. Any fidelity loss in encode/decode surfaces as a
// per-group mismatch.
func TestW2025RoundTrip(t *testing.T) {
	m1, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("initial decode: %v", err)
	}

	xmlBytes, err := h2025v1.Encode(m1)
	if err != nil {
		t.Fatalf("h2025v1.Encode: %v", err)
	}

	m2, err := h2025v1.Decode(xmlBytes)
	if err != nil {
		t.Fatalf("h2025v1.Decode(re-encoded): %v\n---encoded xml (first 800 bytes)---\n%s", err, head(xmlBytes, 800))
	}

	normalizeVolatile(m1)
	normalizeVolatile(m2)

	compareGroups(t, m1, m2)
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
