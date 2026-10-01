# ADR 0006 — gzip: the standard library, at level 6 by default

- **Status:** **Accepted (2026-10-01)**: maintainer decision, recorded after
  [#141](https://github.com/maloquacious/wxx/issues/141).
- **Date:** 2026-10-01
- **Context tickets:** [#138](https://github.com/maloquacious/wxx/issues/138)
  (the benchmarks and profile that found the cost),
  [#141](https://github.com/maloquacious/wxx/issues/141) (`WithGzipLevel`,
  merged in #147), [#142](https://github.com/maloquacious/wxx/issues/142) and
  [#143](https://github.com/maloquacious/wxx/issues/143) (the other encode
  costs).

---

## Context

A `.wxx` file is gzip-compressed UTF-16 XML, so every encode ends by gzipping
the whole document. The #138 profile showed this is most of the cost of
writing a large map. On the 1920 × 1080 random fixture, a 2.08 encode takes
3.19 s, and about 2.3 s of that is `compress/flate` at the default level.

#141 added `xmlio.WithGzipLevel` with a default of 6 (`DefaultGzipLevel`),
which is the level `gzip.DefaultCompression` stands for. The default
therefore writes exactly the bytes wxx always has, and `TestGzipLevel` pins
that. That left open whether the default level, or the compressor itself,
should change.

Worldographer 2.08 opened every file tried: the random and blank 1920 × 1080
fixtures written at levels 1, 6 and 9 with `compress/gzip`, and the random
fixture written with `github.com/klauspost/compress/gzip` at level 6. The app
reads any of them, so the choice is only one of speed, file size and
dependencies.

Measured on the random 1920 × 1080 fixture (full 2.08 encode, median of 3, on
the Apple M4 in `docs/PERFORMANCE.md`). Worldographer's own save of this map
is 9.56 MB.

| compressor | level | encode | size |
|---|---|---|---|
| `compress/gzip` | 1 | 1.18 s | 13.6 MB |
| `compress/gzip` | 4 | about 1.4 s (estimated) | 10.8 MB |
| `compress/gzip` | **6 (default)** | **3.19 s** | **9.39 MB** |
| `compress/gzip` | 9 | 15.6 s | 8.74 MB |
| klauspost | 1 | 1.08 s | 12.0 MB |
| klauspost | 6 | 1.14 s | 9.80 MB |
| klauspost | 9 | 10.5 s | 8.80 MB |

The level 4 time is estimated from a gzip-only measurement (0.40 s against
2.14 s at the default level, in #141) and was not measured as a full encode.

## Options considered

**A. Replace `compress/gzip` with `github.com/klauspost/compress/gzip`.** At
level 6, an encode drops from 3.19 s to 1.14 s for a file 4% larger. It is a
drop-in replacement, but:

- It is a new dependency, against the project rule to add none that can be
  avoided, and against its direction: wxx is to become a script host with
  fewer dependencies, dropping `ff/v4`.
- Its current release requires Go 1.25, so `go.mod` would move off 1.24.4.
- The default would no longer write the bytes wxx has always written.
  `TestGzipLevel` would have to compare the gunzipped document rather than
  the file.

**B. Keep `compress/gzip` and lower the default to 4.** The estimated encode
time, about 1.4 s, is close to A's with no dependency, but the file is about
15% larger, and every caller gets the larger file whether or not save time
matters to them.

**C. Keep `compress/gzip` at level 6, and leave the trade to the caller.**
Encode stays at 3.2 s for a 1920 × 1080 map. A caller who wants speed passes
`WithGzipLevel(1)` and gets 1.2 s.

## Decision

**Option C.** wxx compresses with the standard library's `compress/gzip`, at
level 6 by default (`xmlio.DefaultGzipLevel`), and a caller may choose another
level with `xmlio.WithGzipLevel`. No third-party compressor is used.

## Why

- **The gain does not pay for a dependency.** Over option B, klauspost saves
  about 0.3 s on a 2-million-hex map. Over C, it saves 2 s, and that time is
  available today to any caller who asks for level 1.
- **The default stays what wxx has always written,** and is about the size
  Worldographer writes itself (9.39 MB against 9.56 MB).
- **gzip is not the only cost.** #142 (`Validate` work per tile, run twice per
  encode) and #143 (per-tile `fmt` formatting and whole-document copies) can
  make encoding faster without trading file size or adding a dependency.

## Consequences

- `go.mod` stays at Go 1.24.4 and gains no dependency.
- Writing a 1920 × 1080 map at the default level takes about 3.2 s on the
  baseline machine. A caller who wants faster saves passes a lower level, and
  `docs/PERFORMANCE.md` has the table to choose from.
- `TestGzipLevel` continues to pin that the default writes the bytes
  `gzip.DefaultCompression` writes.

## Revisit when

- #142 and #143 are done and gzip is still most of the encode time on the
  maps people actually write, or
- the project's dependency rule changes, or
- the default size or speed becomes a user complaint.

Reopening this decision means a superseding ADR, not an edit to this one.
