// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// Benchmarks for decode, encode and NewMap (issue #138). They run only with
// -bench, so plain go test is not slowed by them. docs/PERFORMANCE.md has the
// commands, the baseline and what it was measured on.
//
// Each reports ns/tile alongside ns/op, so the 13 x 11 and 1920 x 1080 sizes
// can be compared per tile: fixed costs show up as a higher ns/tile on the
// small map.

// benchFixture is a map the benchmarks read.
type benchFixture struct {
	name string // sub-benchmark name
	path string
}

// benchFixtures are the 13 x 11 blank map for the per-map fixed cost, the
// 1920 x 1080 blank map, which is almost all tile grid and measures the
// decoded map's size, and the 1920 x 1080 random map, a generated world with
// painted tiles and about 1,000 other elements, which measures real content.
var benchFixtures = []benchFixture{
	{"blank-13x11", "../testdata/2025-2.08-13x11-941577-blank.wxx"},
	{"blank-1920x1080", "../testdata/2025-2.08-1920x1080-941577-blank.wxx"},
	{"random-1920x1080", "../testdata/2025-2.08-1920x1080-941577-random.wxx"},
}

// reportPerTile adds an ns/tile metric for a map of tiles hexes.
func reportPerTile(b *testing.B, tiles int) {
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(tiles), "ns/tile")
}

func tileCount(m *wxx.Map_t) int {
	return m.Tiles.TilesWide * m.Tiles.TilesHigh
}

// BenchmarkDecode decodes a .wxx file from memory: gunzip, UTF-16 to UTF-8,
// XML and the model. Reading the file from disk is outside the timing.
func BenchmarkDecode(b *testing.B) {
	for _, f := range benchFixtures {
		b.Run(f.name, func(b *testing.B) {
			data, err := os.ReadFile(f.path)
			if err != nil {
				b.Fatal(err)
			}
			m, err := xmlio.NewDecoder().Decode(bytes.NewReader(data))
			if err != nil {
				b.Fatal(err)
			}
			tiles := tileCount(m)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := xmlio.NewDecoder().Decode(bytes.NewReader(data)); err != nil {
					b.Fatal(err)
				}
			}
			reportPerTile(b, tiles)
		})
	}
}

// BenchmarkEncode encodes a decoded map as 2.08 into memory: Validate, XML,
// UTF-8 to UTF-16 and gzip. Writing the file to disk is outside the timing.
func BenchmarkEncode(b *testing.B) {
	for _, f := range benchFixtures {
		b.Run(f.name, func(b *testing.B) {
			m, err := xmlio.ReadFile(f.path)
			if err != nil {
				b.Fatal(err)
			}
			tiles := tileCount(m)
			var buf bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				buf.Reset()
				if err := xmlio.NewEncoder("2.08").Encode(&buf, m); err != nil {
					b.Fatal(err)
				}
			}
			reportPerTile(b, tiles)
		})
	}
}

// BenchmarkNewMap builds a blank map: the model alone, with no I/O, so its
// bytes/op is close to what a map of that size costs to hold.
func BenchmarkNewMap(b *testing.B) {
	for _, size := range []struct{ columns, rows int }{{13, 11}, {1920, 1080}} {
		b.Run(fmt.Sprintf("%dx%d", size.columns, size.rows), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if _, err := xmlio.NewMap(size.columns, size.rows); err != nil {
					b.Fatal(err)
				}
			}
			reportPerTile(b, size.columns*size.rows)
		})
	}
}
