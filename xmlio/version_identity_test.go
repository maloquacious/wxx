// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// versionIdentitySamples pairs tracked W2025 2.06, 2.07 and 2.08 fixtures with the version
// identity their bytes state, observed end-to-end through the public decoder
// (ADR 0004 Decision 2).
//
// wantApp and wantSchema are the exact on-disk attribute values, byte for byte
// — these are dotted ordinals, not semver, so "2.06" is the whole point and
// "2.6" would be a different file.
var versionIdentitySamples = []struct {
	name       string
	path       string
	wantApp    string // exact map/@version bytes
	wantMajor  int    // parsed App.Major
	wantMinor  int    // parsed App.Minor
	wantSchema string // exact map/@schema bytes
}{
	{"w2025 2.06 blank", sample2025_206, "2.06", 2, 6, "1.06"},
	{"w2025 2.06 layers beta", sample2025_206LayersBeta, "2.06", 2, 6, "1.06"},
	{"w2025 2.07 blank", sample2025_207Blank, "2.07", 2, 7, "1.06"},
	{"w2025 2.07 layers", sample2025_207Layers, "2.07", 2, 7, "1.06"},
	{"w2025 2.07 notes-shapes", sample2025_207NotesShapes, "2.07", 2, 7, "1.06"},
	{"w2025 2.07 resources", sample2025_207Resources, "2.07", 2, 7, "1.06"},
	{"w2025 2.07 rows", sample2025_207Rows, "2.07", 2, 7, "1.06"},
	{"w2025 2.08 blank", sample2025_208Blank, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 layers", sample2025_208Layers, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 notes-shapes", sample2025_208NotesShapes, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 populated", sample2025_208Populated, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 resources", sample2025_208Resources, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 rows", sample2025_208Rows, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 tile-resources", sample2025_208TileResources, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 extraterrain-bgcolor", sample2025_208ExtraTerrainBgColor, "2.08", 2, 8, "1.06"},
	{"w2025 2.08 river-lines", sample2025_208RiverLines, "2.08", 2, 8, "1.06"},
}

// TestVersionIdentity asserts that decoding populates MetaData.Version with the
// two axes the file states: App from map/@version and Schema from map/@schema.
//
// For W2025 this is the first time @version reaches the model at all — the
// superseded DataVersion had no slot for it, spending its Minor.Patch on the
// schema — so the App assertions here are new ground rather than a restatement.
//
// Each sample is then encoded as the version it states and decoded again, and
// must come back stating the same identity: a registered version written as
// itself keeps the identity the file had (issues #92 and #73 registered 2.07
// and 2.08).
func TestVersionIdentity(t *testing.T) {
	for _, tc := range versionIdentitySamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			v := m.MetaData.Version

			// App.Raw must be the on-disk bytes, unaltered.
			if got := v.App.Raw; got != tc.wantApp {
				t.Errorf("MetaData.Version.App.Raw = %q, want %q", got, tc.wantApp)
			}
			if got := v.App.Major; got != tc.wantMajor {
				t.Errorf("MetaData.Version.App.Major = %d, want %d", got, tc.wantMajor)
			}
			if got := v.App.Minor; got != tc.wantMinor {
				t.Errorf("MetaData.Version.App.Minor = %d, want %d", got, tc.wantMinor)
			}
			// App must agree with the verbatim string the encoders write, since
			// both come from the same attribute.
			if got, want := v.App.Raw, m.MetaData.Worldographer.Version; got != want {
				t.Errorf("MetaData.Version.App.Raw = %q, want it to match Worldographer.Version %q", got, want)
			}

			if v.Schema == nil {
				t.Fatalf("MetaData.Version.Schema = nil, want %q (file states a @schema)", tc.wantSchema)
			}
			if got := v.Schema.Raw; got != tc.wantSchema {
				t.Errorf("MetaData.Version.Schema.Raw = %q, want %q", got, tc.wantSchema)
			}

			// Written as itself, the identity survives.
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(tc.wantApp).Encode(&buf, m); err != nil {
				t.Fatalf("encode %s as %q: %v", tc.path, tc.wantApp, err)
			}
			back, err := xmlio.NewDecoder().Decode(&buf)
			if err != nil {
				t.Fatalf("re-decode %s encoded as %q: %v", tc.path, tc.wantApp, err)
			}
			bv := back.MetaData.Version
			if bv.App.Raw != tc.wantApp || bv.Schema == nil || bv.Schema.Raw != tc.wantSchema {
				t.Errorf("%s encoded as %q re-decodes as %v, want app %s, schema %s", tc.path, tc.wantApp, bv, tc.wantApp, tc.wantSchema)
			}
		})
	}
}

// TestVersionIdentityPaddingSurvivesDecode is ADR 0004 Decision 1 observed
// end-to-end from a real file rather than from a constructed string: the W2025
// baseline writes zero-padded ordinals ("2.06", "1.06"), and a semver round-trip
// renders those back as "2.6" and "1.6" — a different string, and therefore a
// different file. Decoding must hand back the bytes it was given.
func TestVersionIdentityPaddingSurvivesDecode(t *testing.T) {
	const wantSchema = "1.06" // rendering the components would give "1.6"

	for _, tc := range []struct {
		name    string
		path    string
		wantApp string // rendering the components would give "2.6", "2.7" or "2.8"
	}{
		{"2.06/1.06 blank", sample2025_206, "2.06"},
		{"2.06/1.06 layers", sample2025_206LayersBeta, "2.06"},
		{"2.07/1.06 blank", sample2025_207Blank, "2.07"},
		{"2.08/1.06 blank", sample2025_208Blank, "2.08"},
	} {
		wantApp := tc.wantApp
		// Guard against a vacuous pass: this test only proves anything if the
		// values it expects are genuinely zero-padded, i.e. if each differs from
		// the unpadded rendering of its own parsed components. Were an
		// expectation ever relaxed to "2.6", every assertion below would still
		// pass while asserting nothing whatsoever about padding.
		for _, want := range []string{wantApp, wantSchema} {
			d, err := wxx.ParseDotted(want)
			if err != nil {
				t.Fatalf("ParseDotted(%q): %v", want, err)
			}
			if unpadded := fmt.Sprintf("%d.%d", d.Major, d.Minor); unpadded == want {
				t.Fatalf("expected value %q is not zero-padded (renders identically as %q), so padding preservation is not under test", want, unpadded)
			}
		}

		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			v := m.MetaData.Version
			if v.Schema == nil {
				t.Fatalf("MetaData.Version.Schema = nil, want non-nil for a W2025 file")
			}
			if got := v.App.Raw; got != wantApp {
				t.Errorf("MetaData.Version.App.Raw = %q, want %q verbatim (re-rendering the components would give %q)",
					got, wantApp, fmt.Sprintf("%d.%d", v.App.Major, v.App.Minor))
			}
			if got := v.Schema.Raw; got != wantSchema {
				t.Errorf("MetaData.Version.Schema.Raw = %q, want %q verbatim (re-rendering the components would give %q)",
					got, wantSchema, fmt.Sprintf("%d.%d", v.Schema.Major, v.Schema.Minor))
			}
		})
	}
}
