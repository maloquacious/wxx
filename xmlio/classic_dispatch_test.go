// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/h2017v1"
)

// classicSamples are H2017 ("classic") .wxx fixtures with their true on-disk
// <map> version attribute. Classic files carry no release/schema attributes, so
// the public dispatcher must route them by the "1.x" version shape.
//
// wantMajor/wantMinor are the components MetaData.Version.App parses the on-disk
// version into ("1.77" -> {1, 77}). Every classic fixture states no @schema, so
// the identity's Schema is nil throughout -- that absence is the implicit legacy
// schema, and it is what selects the classic codec on the way back out. All
// classic fixtures are COLUMNS orientation (verified), so they can be re-encoded
// without hitting the documented classic ROWS-encode gap.
var classicSamples = []struct {
	name      string
	path      string
	version   string // on-disk <map version=...>
	wantMajor int    // parsed MetaData.Version.App.Major
	wantMinor int    // parsed MetaData.Version.App.Minor
}{
	{"1.77-columns-blank", "../testdata/2017-1.77-1.0-columns-blank.wxx", "1.77", 1, 77},
	{"1.77-blank", "../testdata/blank-2017-1.77-1.0.wxx", "1.77", 1, 77},
	{"1.74-blank", "../testdata/blank-2017-1.74-1.0.wxx", "1.74", 1, 74},
	{"1.73-blank", "../testdata/blank-2017-1.73-1.0.wxx", "1.73", 1, 73},
}

// assertClassicIdentity asserts m carries the on-disk version identity of a
// classic file: the exact <map version> bytes, the components they parse to, and
// a nil Schema for the @schema the file does not state.
func assertClassicIdentity(t *testing.T, label string, v wxx.Version_t, wantApp string, wantMajor, wantMinor int) {
	t.Helper()
	if got := v.App.Raw; got != wantApp {
		t.Errorf("%s: Version.App.Raw = %q, want %q verbatim", label, got, wantApp)
	}
	if got := v.App.Major; got != wantMajor {
		t.Errorf("%s: Version.App.Major = %d, want %d", label, got, wantMajor)
	}
	if got := v.App.Minor; got != wantMinor {
		t.Errorf("%s: Version.App.Minor = %d, want %d", label, got, wantMinor)
	}
	if v.Schema != nil {
		t.Errorf("%s: Version.Schema = %+v, want nil (a classic file states no @schema)", label, *v.Schema)
	}
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
// MetaData.Worldographer.Version -- the string the encoder writes back -- and as
// the version identity in MetaData.Version (App "1.7x", nil Schema) per ADR 0004
// Decision 2.
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
			// The identity states the on-disk version on the App axis and nothing
			// on the schema axis, because the file states nothing there.
			assertClassicIdentity(t, "MetaData", m.MetaData.Version, tc.version, tc.wantMajor, tc.wantMinor)
		})
	}
}

// TestClassicEncodeDispatch asserts that the public encoder routes every classic
// fixture to the h2017v1 encoder -- the guarantee this test has always made,
// restated on the model that now carries it.
//
// What routes has changed: the encoder used to switch on a DataVersion.Major of
// 2017, a family year no classic file states, and now resolves the codec from
// the schema the file itself states -- for a classic file, the schema it pointedly
// does not state (ADR 0004 Decision 4). The proof is unchanged and end-to-end: a
// map decoded from each fixture re-encodes to bytes that decode back as a classic
// file carrying its own version="1.7x", which only the classic encoder produces.
func TestClassicEncodeDispatch(t *testing.T) {
	for _, tc := range classicSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			// sanity: the identity the encode dispatch reads is the real on-disk
			// one, so what follows exercises the classic route rather than a
			// zero-valued fallback.
			assertClassicIdentity(t, "decoded", m.MetaData.Version, tc.version, tc.wantMajor, tc.wantMinor)

			// The schema this fixture states selects the classic codec. This is
			// the dispatch decision itself, asserted directly rather than inferred
			// from the output.
			codec, err := xmlio.CodecForSchema(m.MetaData.Version.Schema)
			if err != nil {
				t.Fatalf("CodecForSchema(%v): %v", m.MetaData.Version, err)
			}
			if funcPtr(codec.Encode) != funcPtr(h2017v1.Encode) {
				t.Errorf("schema of %s does not select the h2017v1 encoder", tc.path)
			}

			var buf bytes.Buffer
			if err := xmlio.NewEncoder().Encode(&buf, m); err != nil {
				t.Fatalf("public encode %s (%v): %v", tc.path, m.MetaData.Version, err)
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
			assertClassicIdentity(t, "re-decoded", m2.MetaData.Version, tc.version, tc.wantMajor, tc.wantMinor)
		})
	}
}
