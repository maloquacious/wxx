// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"path/filepath"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// classicFixture is a classic H2017 (COLUMNS) blank map. It is safe to
// re-encode through the 2017 codec, so it drives the ReadFile/WriteFile
// round-trip tests.
const classicFixture = "../testdata/blank-2017-1.77-1.0.wxx"

// TestReadFile decodes a real .wxx fixture through the file-level convenience
// API and asserts the expected map metadata.
func TestReadFile(t *testing.T) {
	m, err := xmlio.ReadFile(classicFixture)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", classicFixture, err)
	}
	if m == nil {
		t.Fatalf("ReadFile(%s): nil map", classicFixture)
	}
	if got, want := m.MetaData.Worldographer.Version, "1.77"; got != want {
		t.Errorf("MetaData.Worldographer.Version = %q, want %q", got, want)
	}
}

// TestWriteFileReadFileRoundTrip reads a fixture, writes it back out with
// WriteFile, reads that file with ReadFile, and asserts the two maps agree on
// their key structural fields.
func TestWriteFileReadFileRoundTrip(t *testing.T) {
	m1, err := xmlio.ReadFile(classicFixture)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", classicFixture, err)
	}

	// The target is named explicitly: WriteFile has no default, for the reason
	// NewEncoder has none (issue #45). A round trip means writing the version the
	// file states, so this reads that provenance and says so -- which a CLIENT may
	// do and the encoder may not do for it.
	out := filepath.Join(t.TempDir(), "roundtrip.wxx")
	if err := xmlio.WriteFile(out, m1, m1.MetaData.Version.App.Raw); err != nil {
		t.Fatalf("WriteFile(%s): %v", out, err)
	}

	m2, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", out, err)
	}

	// The on-disk version identity must survive the write/read: same application
	// version, and a Schema that is still the absent (implicit legacy) one rather
	// than a schema the round-trip invented.
	v1, v2 := m1.MetaData.Version, m2.MetaData.Version
	if got, want := v2.App, v1.App; got != want {
		t.Errorf("MetaData.Version.App = %+v, want %+v", got, want)
	}
	switch {
	case v1.Schema == nil && v2.Schema != nil:
		t.Errorf("MetaData.Version.Schema = %+v, want nil", *v2.Schema)
	case v1.Schema != nil && v2.Schema == nil:
		t.Errorf("MetaData.Version.Schema = nil, want %+v", *v1.Schema)
	case v1.Schema != nil && v2.Schema != nil && *v1.Schema != *v2.Schema:
		t.Errorf("MetaData.Version.Schema = %+v, want %+v", *v2.Schema, *v1.Schema)
	}
	if got, want := m2.HexOrientation, m1.HexOrientation; got != want {
		t.Errorf("HexOrientation = %q, want %q", got, want)
	}
	if m1.Tiles == nil || m2.Tiles == nil {
		t.Fatalf("Tiles nil: m1=%v m2=%v", m1.Tiles == nil, m2.Tiles == nil)
	}
	if got, want := m2.Tiles.TilesWide, m1.Tiles.TilesWide; got != want {
		t.Errorf("Tiles.TilesWide = %d, want %d", got, want)
	}
	if got, want := m2.Tiles.TilesHigh, m1.Tiles.TilesHigh; got != want {
		t.Errorf("Tiles.TilesHigh = %d, want %d", got, want)
	}
	if got, want := len(m2.Tiles.Tiles), len(m1.Tiles.Tiles); got != want {
		t.Errorf("len(Tiles.Tiles) = %d, want %d", got, want)
	}
}

// TestReadFileMissing asserts ReadFile returns an error (not a panic) for a
// path that does not exist.
func TestReadFileMissing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.wxx")
	if _, err := xmlio.ReadFile(missing); err == nil {
		t.Fatalf("ReadFile(%s): expected error, got nil", missing)
	}
}
