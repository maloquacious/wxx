// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// TestW2025TileResourceOutOfRangeRefused is issue #122's reproduction.
//
// The encoder wrote a tile resource as whatever integer it was given.
// Worldographer 2.08 refuses to open a file with Brick 150 (Byte.parseByte) and
// saves 127 back as 100, and the decoder refuses anything above 100, so wxx
// wrote a file that neither the app nor wxx could read back. The encoder now
// refuses it before writing a byte.
func TestW2025TileResourceOutOfRangeRefused(t *testing.T) {
	m, err := xmlio.ReadFile(sample2025_208TileResources)
	if err != nil {
		t.Fatalf("read %s: %v", sample2025_208TileResources, err)
	}
	m.Tiles.Tiles[0][0].Resources.Brick = 150

	var buf bytes.Buffer
	err = xmlio.NewEncoder("2.08").Encode(&buf, m)
	if err == nil {
		t.Fatalf("Encode: want an error, got nil")
	}
	if !errors.Is(err, wxx.ErrInvalidTileResource) {
		t.Errorf("Encode: err = %v, want errors.Is(err, %v)", err, wxx.ErrInvalidTileResource)
	}
	if want := "Tiles_t.Tiles[0][0].Resources.Brick): 150"; !strings.Contains(err.Error(), want) {
		t.Errorf("Encode: err = %q, want it to name %q", err.Error(), want)
	}
	if buf.Len() != 0 {
		t.Errorf("Encode: wrote %d bytes to w, want 0", buf.Len())
	}
}
