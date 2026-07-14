// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"testing"

	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx/xmlio"
)

// classicSamples are H2017 ("classic") .wxx fixtures with their true on-disk
// <map> version attribute. Classic files carry no release/schema attributes, so
// the public dispatcher must route them by the "1.x" version shape.
//
// wantDV is the parsed DataVersion (ADR 0002): Major=2017 (schema family),
// Minor.Patch = the on-disk dotted revision ("1.77" -> {2017,1,77}). All classic
// fixtures are COLUMNS orientation (verified), so they can be re-encoded without
// hitting the documented classic ROWS-encode gap.
var classicSamples = []struct {
	name    string
	path    string
	version string         // on-disk <map version=...>
	wantDV  semver.Version // parsed MetaData.DataVersion
}{
	{"1.77-columns-blank", "../testdata/input/2017-1.77-1.0-columns-blank.wxx", "1.77", semver.Version{Major: 2017, Minor: 1, Patch: 77}},
	{"1.77-blank", "../testdata/input/blank-2017-1.77-1.0.wxx", "1.77", semver.Version{Major: 2017, Minor: 1, Patch: 77}},
	{"1.74-blank", "../testdata/input/blank-2017-1.74-1.0.wxx", "1.74", semver.Version{Major: 2017, Minor: 1, Patch: 74}},
	{"1.73-blank", "../testdata/input/blank-2017-1.73-1.0.wxx", "1.73", semver.Version{Major: 2017, Minor: 1, Patch: 73}},
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
// real on-disk sub-revision (1.73/1.74/1.77) both verbatim in
// MetaData.Worldographer.Version and as a parsed semver in MetaData.DataVersion
// ({2017,1,7x}) per ADR 0002.
func TestClassicVersionFidelity(t *testing.T) {
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
			// DataVersion carries the family in Major and the on-disk revision in
			// Minor.Patch. Minor stays 1 for every classic file, so the encode
			// dispatch (which routes on Major) is unaffected.
			if got := m.MetaData.DataVersion; got != tc.wantDV {
				t.Errorf("MetaData.DataVersion = %v, want %v", got, tc.wantDV)
			}
		})
	}
}

// TestClassicEncodeDispatch asserts that the public encoder routes a classic map
// whose DataVersion is {2017,1,7x} to the h2017v1 encoder. This is the ADR 0002
// dispatch relaxation: before it, MarshalXML required Minor==1 AND Patch==0-style
// {2017,1} matching via the Minor switch; now it routes on Major (family) alone,
// so a real parsed classic DataVersion no longer trips ErrUnsupportedSchemaVersion.
func TestClassicEncodeDispatch(t *testing.T) {
	for _, tc := range classicSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			// sanity: this is the enriched DataVersion, not the old {2017,1,0}.
			if m.MetaData.DataVersion != tc.wantDV {
				t.Fatalf("DataVersion = %v, want %v", m.MetaData.DataVersion, tc.wantDV)
			}
			var buf bytes.Buffer
			if err := xmlio.NewEncoder().Encode(&buf, m); err != nil {
				t.Fatalf("public encode %s (DataVersion %v): %v", tc.path, m.MetaData.DataVersion, err)
			}
			if buf.Len() == 0 {
				t.Fatalf("public encode %s: empty output", tc.path)
			}
			// re-encoded output must re-decode as a classic file (round-trips the
			// version="1.7x" attribute the encoder emits verbatim).
			m2, err := xmlio.NewDecoder().Decode(&buf)
			if err != nil {
				t.Fatalf("re-decode %s: %v", tc.path, err)
			}
			if got := m2.MetaData.Worldographer.Version; got != tc.version {
				t.Errorf("re-decoded Worldographer.Version = %q, want %q", got, tc.version)
			}
			if got := m2.MetaData.DataVersion; got != tc.wantDV {
				t.Errorf("re-decoded DataVersion = %v, want %v", got, tc.wantDV)
			}
		})
	}
}
