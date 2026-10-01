# Performance and memory

How fast wxx reads and writes a map, and how much memory it uses (issue
[#138](https://github.com/maloquacious/wxx/issues/138)). The numbers below are
a baseline to compare later changes against. They are not a target.

## Running the benchmarks

The benchmarks are in `xmlio/bench_test.go`. Plain `go test ./...` does not
run them.

```sh
go test ./xmlio -run '^$' -bench . -benchmem
```

They read three fixtures, all written by 2.08:

| fixture | hexes | gzipped | UTF-16 XML | measures |
|---|---|---|---|---|
| `2025-2.08-13x11-941577-blank.wxx` | 143 | 8 KB | 80 KB | fixed cost per map |
| `2025-2.08-1920x1080-941577-blank.wxx` | 2,073,600 | 133 KB | 49.9 MB | the tile grid alone |
| `2025-2.08-1920x1080-941577-random.wxx` | 2,073,600 | 9.5 MB | 77.8 MB | a generated world: painted tiles and about 1,000 other elements |

- **Decode:** decodes from bytes already in memory.
- **Encode:** encodes a decoded map as 2.08 into a buffer.
- **NewMap:** builds a blank map, with no I/O.

Each benchmark also reports `ns/tile`, so the two sizes can be compared per
hex.

To measure a benchmark's peak memory, build the test binary and run that one
benchmark under `/usr/bin/time -l`:

```sh
go test -c -o dist/local/xmlio.test ./xmlio
cd xmlio && /usr/bin/time -l ../dist/local/xmlio.test -test.run '^$' -test.bench 'Encode/random' -test.benchtime 1x
```

The figure includes the benchmark's setup: Encode decodes the fixture first,
and Decode decodes it once before timing.

To profile one benchmark:

```sh
go test ./xmlio -run '^$' -bench 'Encode/random' -benchtime 3x \
  -cpuprofile cpu.out -memprofile mem.out -o xmlio.test
go tool pprof -top -cum xmlio.test cpu.out
```

## Baseline

Measured 2026-10-01 at 0.47.0-beta (`71fb31a`) on an Apple M4 (10 cores, 32 GB),
macOS, Go 1.26.4 (go.mod states 1.24.4). This is one run (`-count 1`); expect a few percent of variance
between runs.

| benchmark | time/op | ns/tile | bytes/op | allocs/op | peak RSS |
|---|---|---|---|---|---|
| Decode blank 13×11 | 0.63 ms | 4,395 | 0.7 MB | 5,691 | 9 MiB |
| Decode blank 1920×1080 | 0.41 s | 195 | 742 MB | 4.18 M | 424 MiB |
| Decode random 1920×1080 | 0.69 s | 332 | 942 MB | 4.28 M | 613 MiB |
| Encode blank 13×11 | 0.69 ms | 4,823 | 1.3 MB | 5,497 | — |
| Encode blank 1920×1080 | 0.85 s | 411 | 483 MB | 21.3 M | 646 MiB |
| Encode random 1920×1080 | 3.13 s | 1,511 | 717 MB | 26.9 M | 734 MiB |
| NewMap 13×11 | 6 µs | 42 | 27 KB | 237 | — |
| NewMap 1920×1080 | 73 ms | 35 | 284 MB | 2.08 M | 294 MiB |

## What the baseline shows

**A decoded map costs about 137 bytes per hex.** `NewMap(1920, 1080)`
allocates 284 MB, which is almost all tiles: each is a separate `*Tile_t`
holding cube coordinates, column, row, terrain, elevation, flags and seven
`int` resources. Decoding either 1920 × 1080 fixture leaves the same 280 MiB
in use. That is the model's size, and painted terrain does not change it.

**Encoding is slower than decoding, and gzip is most of it.** In a CPU profile
of Encode random, about 2.3 s of the 3.1 s is `compress/flate` at the default
level. The rest is mostly `encodeTiles` (one `fmt.Sprintf` per tile and per
resource) and `Validate`, which `Encoder.Encode` runs twice: once itself, and
again inside `MarshalXML`. These are filed as #141, #142 and #143.

**Fixed cost dominates a small map.** 13 × 11 costs about 4,400 ns per hex,
against 200–1,500 at 1920 × 1080. The configuration, map key and XML header
are the same at any size.

## Choosing a gzip level

`xmlio.WithGzipLevel(level)` sets the gzip level for an encode, from 1
(fastest) to 9 (smallest), or 0 for no compression (#141). The default is
`xmlio.DefaultGzipLevel`, 6, which writes the same bytes wxx always has.
Worldographer 2.08 opened the 1920 × 1080 maps written at levels 1, 6 and 9.

| map | level | encode | size |
|---|---|---|---|
| random 1920×1080 | 1 | 1.18 s | 13.6 MB |
| random 1920×1080 | 6 | 3.19 s | 9.39 MB |
| random 1920×1080 | 9 | 15.6 s | 8.74 MB |
| blank 1920×1080 | 1 | 0.81 s | 481 KB |
| blank 1920×1080 | 6 | 0.85 s | 133 KB |
| blank 1920×1080 | 9 | 0.86 s | 132 KB |

Times are the full encode, median of 3, on the baseline machine. Worldographer's
own saves of these maps are 9.56 MB and 133 KB.

## Is memory a practical limit?

Not at 1920 × 1080. Peak memory is under 750 MiB for any single operation.
It grows about linearly with the hex count:

- **The model:** about 137 bytes per hex.
- **Peak while decoding or encoding:** about 300–350 bytes per hex. The pipeline
  holds the whole document several times over: gzip output, UTF-16, UTF-8,
  and the XML tree or the encoder's buffer.

Extrapolating, which is untested:

| hexes | example | model | peak during read or write |
|---|---|---|---|
| 2 M | 1920 × 1080 | 0.3 GB | 0.75 GB |
| 8 M | 4000 × 2000 | 1.1 GB | about 3 GB |
| 16 M | 4000 × 4000 | 2.2 GB | about 6 GB |

So on a machine with 8–16 GB, the limit is somewhere around 10–30 million
hexes, mostly set by the copies the pipeline holds rather than by the model.
Streaming the pipeline stages and storing smaller tiles would raise it. Neither
is needed for the maps in hand.

## Follow-up issues

The profile findings are filed separately rather than fixed in #138:

- gzip level: #141 (`WithGzipLevel`, above)
- `Validate` allocates per tile and runs twice per encode: #142
- the encode pipeline's per-tile formatting and whole-document copies: #143
