// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"testing"

	"github.com/maloquacious/semver"
)

// classicSamples are H2017 ("classic") .wxx fixtures with their true on-disk
// <map> version attribute. Classic files carry no release/schema attributes, so
// the public dispatcher must route them by the "1.x" version shape.
var classicSamples = []struct {
	name    string
	path    string
	version string // on-disk <map version=...>
}{
	{"1.77-columns-blank", "../testdata/input/2017-1.77-1.0-columns-blank.wxx", "1.77"},
	{"1.77-blank", "../testdata/input/blank-2017-1.77-1.0.wxx", "1.77"},
	{"1.74-blank", "../testdata/input/blank-2017-1.74-1.0.wxx", "1.74"},
	{"1.73-blank", "../testdata/input/blank-2017-1.73-1.0.wxx", "1.73"},
}

// TestClassicDispatch_Decode asserts that the PUBLIC decoder
// (xmlio.NewDecoder().Decode) accepts H2017 classic files. Before the classic
// dispatch backfill this fails with ErrUnsupportedMapMetadata because the
// dispatch switch only handled release=="2025".
func TestClassicDispatch_Decode(t *testing.T) {
	for _, tc := range classicSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			if m == nil {
				t.Fatalf("public decode %s: nil map", tc.path)
			}
			if m.Tiles == nil {
				t.Fatalf("public decode %s: nil Tiles", tc.path)
			}
			if m.TerrainMap == nil || len(m.TerrainMap.List) == 0 {
				t.Fatalf("public decode %s: empty TerrainMap", tc.path)
			}
		})
	}
}

// TestClassicVersionFidelity asserts that decoding a classic file preserves the
// real on-disk sub-revision (1.73/1.74/1.77) additively in
// MetaData.Worldographer.Version, while DataVersion stays {2017,1} so the encode
// dispatch key is unchanged.
func TestClassicVersionFidelity(t *testing.T) {
	want2017 := semver.Version{Major: 2017, Minor: 1}
	for _, tc := range classicSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			if got := m.MetaData.Worldographer.Version; got != tc.version {
				t.Errorf("MetaData.Worldographer.Version = %q, want %q", got, tc.version)
			}
			// Classic files carry no release/schema attrs; those stay empty.
			if got := m.MetaData.Worldographer.Release; got != "" {
				t.Errorf("MetaData.Worldographer.Release = %q, want empty", got)
			}
			if got := m.MetaData.Worldographer.Schema; got != "" {
				t.Errorf("MetaData.Worldographer.Schema = %q, want empty", got)
			}
			// DataVersion is the encode dispatch key; classic must stay {2017,1}.
			if got := m.MetaData.DataVersion; got != want2017 {
				t.Errorf("MetaData.DataVersion = %v, want %v", got, want2017)
			}
		})
	}
}
