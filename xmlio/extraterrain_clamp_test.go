// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"golang.org/x/text/encoding/unicode"
)

// layersVolcano is the first <terrainAndLocation> in the 2.08 layers fixture: a
// Mountain Volcano on Below All at hex (12,9). Its @resources is the value the
// #124 app check edited.
const layersVolcano = `resources="6,11,6,4,2,49,40" location="2700.0,2700.0"`

// doctoredLayers returns the 2.08 layers fixture as a .wxx file (gzip, UTF-16BE)
// with that volcano's @resources replaced by resources, and nothing else
// changed. It goes through the whole public decode pipeline, so diagnostics are
// exercised as a caller sees them.
func doctoredLayers(t *testing.T, resources string) []byte {
	t.Helper()
	f, err := os.Open(sample2025_208Layers)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	var dd xmlio.DecoderDiagnostics
	if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Contains(dd.Converted, []byte(layersVolcano)) {
		t.Fatalf("fixture does not contain %q; this test is doctoring the wrong text", layersVolcano)
	}
	doctored := bytes.Replace(dd.Converted, []byte(layersVolcano),
		[]byte(`resources="`+resources+`" location="2700.0,2700.0"`), 1)

	utf16, err := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewEncoder().Bytes(doctored)
	if err != nil {
		t.Fatalf("utf-16: %v", err)
	}
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write(utf16); err != nil {
		t.Fatalf("gzip: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip: %v", err)
	}
	return gz.Bytes()
}

// TestW2025ExtraTerrainResourceClampedOnDecode pins the decode half of issue
// #124, which is LOSSY by the maintainer's ruling: a <terrainAndLocation>
// resource outside 0..100 is clamped into it, not refused, and every clamp is
// reported in DecoderDiagnostics.Clamped so the caller can tell the user.
//
// 150 is the value the app refuses to open (app check, #124); 127 the largest
// it parses, which it is assumed to save back as 100 as it does in a tile
// record (#122); -5 the negative case, untested in the app and clamped to 0.
func TestW2025ExtraTerrainResourceClampedOnDecode(t *testing.T) {
	const field = "Map_t.ExtraTerrain.MapLayers[0].Terrain[0].Resources."
	for _, tc := range []struct {
		name      string
		resources string
		want      wxx.Resources_t
		clamped   []xmlio.ClampedValue_t // Path, Field, Was and Now; Reason is checked separately
	}{
		{
			name:      "150 in the second field",
			resources: "6,150,6,4,2,49,40",
			want:      wxx.Resources_t{Animal: 6, Brick: 100, Crops: 6, Gems: 4, Lumber: 2, Metals: 49, Rock: 40},
			clamped:   []xmlio.ClampedValue_t{{Field: field + "Brick", Was: 150, Now: 100}},
		},
		{
			name:      "127, the largest the app parses",
			resources: "6,11,6,4,2,49,127",
			want:      wxx.Resources_t{Animal: 6, Brick: 11, Crops: 6, Gems: 4, Lumber: 2, Metals: 49, Rock: 100},
			clamped:   []xmlio.ClampedValue_t{{Field: field + "Rock", Was: 127, Now: 100}},
		},
		{
			name:      "negative",
			resources: "-5,11,6,4,2,49,40",
			want:      wxx.Resources_t{Animal: 0, Brick: 11, Crops: 6, Gems: 4, Lumber: 2, Metals: 49, Rock: 40},
			clamped:   []xmlio.ClampedValue_t{{Field: field + "Animal", Was: -5, Now: 0}},
		},
		{
			name:      "several in one value, reported in field order",
			resources: "101,11,6,4,2,200,-1",
			want:      wxx.Resources_t{Animal: 100, Brick: 11, Crops: 6, Gems: 4, Lumber: 2, Metals: 100, Rock: 0},
			clamped: []xmlio.ClampedValue_t{
				{Field: field + "Animal", Was: 101, Now: 100},
				{Field: field + "Metals", Was: 200, Now: 100},
				{Field: field + "Rock", Was: -1, Now: 0},
			},
		},
		{
			// The bounds themselves are kept: an off-by-one would report a
			// loss that did not happen.
			name:      "0 and 100 are not clamped",
			resources: "0,100,0,100,0,100,0",
			want:      wxx.Resources_t{Animal: 0, Brick: 100, Crops: 0, Gems: 100, Lumber: 0, Metals: 100, Rock: 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dd xmlio.DecoderDiagnostics
			m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(bytes.NewReader(doctoredLayers(t, tc.resources)))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := m.ExtraTerrain.MapLayers[0].Terrain[0].Resources; got != tc.want {
				t.Errorf("resources = %+v, want %+v", got, tc.want)
			}
			if len(dd.Clamped) != len(tc.clamped) {
				t.Fatalf("Clamped = %v, want %d entries", dd.Clamped, len(tc.clamped))
			}
			for i, want := range tc.clamped {
				got := dd.Clamped[i]
				if got.Field != want.Field || got.Was != want.Was || got.Now != want.Now {
					t.Errorf("Clamped[%d] = %s, want Field %s, %d clamped to %d", i, got, want.Field, want.Was, want.Now)
				}
				if wantPath := `map/extraTerrain/mapLayer[@name="Below All"]/terrainAndLocation[0]/@resources`; got.Path != wantPath {
					t.Errorf("Clamped[%d].Path = %q, want %q", i, got.Path, wantPath)
				}
				if !strings.Contains(got.Reason, "#124") {
					t.Errorf("Clamped[%d].Reason = %q, want it to cite #124", i, got.Reason)
				}
			}
		})
	}
}

// TestW2025ExtraTerrainClampIsWrittenBack shows what the loss costs: the
// clamped value, not the file's, is what an encode writes. That is the point of
// reporting it.
func TestW2025ExtraTerrainClampIsWrittenBack(t *testing.T) {
	m, err := xmlio.NewDecoder().Decode(bytes.NewReader(doctoredLayers(t, "6,150,6,4,2,49,40")))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	out, err := xmlio.MarshalXML(m, "2.08")
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if want := `resources="6,100,6,4,2,49,40" location="2700.0,2700.0"`; !bytes.Contains(out, []byte(want)) {
		t.Errorf("encoded output does not contain %q", want)
	}
}

// TestW2025FixturesReportNoClamps holds the decode to its claim that a file the
// app saved is never clamped: the app does not write an out-of-range value, so
// a clamp on a real save would mean the range, or the field order, is wrong.
func TestW2025FixturesReportNoClamps(t *testing.T) {
	for _, path := range append(append([]string{}, fixtures207...), fixtures208...) {
		var dd xmlio.DecoderDiagnostics
		if _, err := xmlio.ReadFile(path, xmlio.WithDecoderDiagnostics(&dd)); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if len(dd.Clamped) != 0 {
			t.Errorf("%s: Clamped = %v, want none", path, dd.Clamped)
		}
	}
}

// TestW2025ExtraTerrainResourceOutOfRangeRefused is the encode half of issue
// #124: a caller's out-of-range <extraTerrain> resource is refused before a byte
// is written, because the app will not open a file holding 150 there.
func TestW2025ExtraTerrainResourceOutOfRangeRefused(t *testing.T) {
	m, err := xmlio.ReadFile(sample2025_208Layers)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	m.ExtraTerrain.MapLayers[0].Terrain[0].Resources.Brick = 150

	var buf bytes.Buffer
	err = xmlio.NewEncoder("2.08").Encode(&buf, m)
	if err == nil {
		t.Fatalf("Encode: want an error, got nil")
	}
	if !errors.Is(err, wxx.ErrInvalidExtraTerrainResource) {
		t.Errorf("Encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidExtraTerrainResource)
	}
	if want := "ExtraTerrain_t.MapLayers[0].Terrain[0].Resources.Brick): 150"; !strings.Contains(err.Error(), want) {
		t.Errorf("Encode: err = %q, want it to name %q", err.Error(), want)
	}
	if buf.Len() != 0 {
		t.Errorf("Encode: wrote %d bytes to w, want 0", buf.Len())
	}
}
