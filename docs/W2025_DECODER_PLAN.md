# W2025 Decoder Implementation Plan

> **Historical / superseded.** This is an early plan (Feb 2026) salvaged for
> the record. W2025 support was ultimately completed by a different,
> round-trip-test-driven route (see issue #4 / PR #5), which found the decoder
> was already comprehensive and the *encoder* was the incomplete side — the
> opposite of this plan's premise. Kept for its schema-analysis and
> definition-of-done notes; do not treat it as the current roadmap.

This document is the implementation plan for adding full W2025 (Worldographer 2025) decoding support to the wxx package. The W2025 app is out of beta (Inkwell is tagging it as v2) and the prior H2017 format (v1.77) is at end of life.

## Prerequisites

### P1. Obtain a W2025 sample file

We need at least one `.wxx` file created by Worldographer 2025 in `testdata/`. The file is gitignored (`testdata/.gitignore` blocks all files except itself), so it stays local. Ideally we want:

- A small map (e.g. 10x10) with a few features, labels, shapes, and notes populated.
- A larger map to test performance and edge cases.

The file must decode through the existing pipeline's first stages (gunzip, UTF-16/BE to UTF-8) - the dispatcher already recognizes `2025/1.10/1.01` and routes to `v1_06.Decode`. We just need the Decode function to do real work.

**Action:** Create or export a W2025 map from Worldographer 2025 and place it in `testdata/`.

### P2. Dump the W2025 XML schema

Use the existing `cmd/schema` tool to dump the XML hierarchy of the sample file. This tool already handles gunzip, UTF-16/BE decoding, metadata detection, and recursive schema inference.

```bash
go build -o dist/local/schema ./cmd/schema
dist/local/schema testdata/<sample>.wxx
```

This produces:
- Confirmation the file is W2025 (`W2025: version 1.10: schema 1.01`)
- The full XML element/attribute hierarchy

**Action:** Run the schema tool, save the output to `xmlio/internal/v1_06/SCHEMA_DUMP.md` for reference during development.

### P3. Diff the schema against H2017

Compare the W2025 hierarchy output against the H2017 schema types in `xmlio/internal/v0_77/schema.go`. Categorize differences:

| Category | What to look for |
|---|---|
| **New elements** | Elements present in W2025 but absent in H2017 |
| **Removed elements** | Elements in H2017 but absent in W2025 |
| **New attributes** | Attributes added to existing elements |
| **Removed attributes** | Attributes dropped from existing elements |
| **Changed structure** | Elements that moved, were renamed, or changed nesting |
| **Tile row format** | Whether the tab-delimited tilerow `InnerText` format changed |

Pay special attention to:
- The `<map>` element (new `release` and `schema` attributes are already known)
- The tile row encoding (the most complex parsing in H2017)
- Configuration sections (terrain-config, feature-config, etc.)
- Any entirely new top-level sections

**Action:** Document findings in `xmlio/internal/v1_06/SCHEMA_DIFF.md`.

---

## Implementation Steps

### Step 1. Define `internal/v1_06/schema.go` - XML unmarshal types

Create Go structs with `xml` tags that match the W2025 XML structure. Follow the exact pattern from `internal/v0_77/schema.go`:

- One top-level `XMLSchema` struct with `xml:"map"` name
- Map attributes as fields with `xml:"...,attr"` tags (including the new `release` and `schema` attrs)
- Nested structs for each child element (`gridandnumbering`, `terrainmap`, `tiles`, etc.)
- Wrapper types for element collections (e.g. `Features` wrapping `[]Feature`)
- `InnerText string \`xml:",chardata"\`` for elements with text content
- Helper functions (`decodeRgba`, etc.) - reuse from v0_77 if format is identical, or create W2025-specific versions

**Key differences to expect vs H2017:**
- `XMLSchema` will have `Release string \`xml:"release,attr"\`` and `Schema string \`xml:"schema,attr"\``
- W2025 uses XML 1.1 (but Go's `encoding/xml` handles this since the header is already stripped)
- Possible new elements/attributes identified in P3

**Guidance:**
- Start by copying `internal/v0_77/schema.go` as a template, then modify based on the schema diff
- Keep types unexported (package-private) except the `Decode` function
- Use the same naming conventions (`_t` suffix for types, PascalCase for fields)

**File:** `xmlio/internal/v1_06/schema.go`

### Step 2. Implement `internal/v1_06/decode.go` - XML to Map_t translation

Replace the current stub with a full implementation. Follow the pattern in `internal/v0_77/decode.go`:

```go
func Decode(input []byte) (*wxx.Map_t, error) {
    m := &XMLSchema{}
    err := xml.Unmarshal(input, &m)
    if err != nil {
        return nil, err
    }

    w := &wxx.Map_t{}
    w.MetaData.AppVersion = wxx.Version()
    w.MetaData.DataVersion = semver.Version{Major: 2025, Minor: 1}
    // ... translate all fields from m to w ...
    return w, nil
}
```

Translation sections (mirroring H2017 decode order):

1. **Metadata** - Set `DataVersion` to `{Major: 2025, Minor: 1}`, populate `Worldographer.Release`, `Worldographer.Schema`
2. **Map attributes** - Copy all `<map>` attributes to `Map_t` fields (factors, offsets, hex dimensions, orientation, projection, show flags, etc.)
3. **Grid orientation** - Parse `HexOrientation` to `hexg.OddQ`/`hexg.OddR` (verify W2025 uses same values)
4. **GridAndNumbering** - Straight field copy
5. **TerrainMap** - Parse the tab-delimited `InnerText` into `TerrainMap_t` (verify format matches H2017)
6. **MapLayers** - Copy layer name/visibility pairs
7. **Tiles** - Parse tile rows. This is the most complex section:
   - Verify the tilerow InnerText format (tab-delimited fields per line)
   - Determine if field count/order changed from H2017's `6/7/11/12` variants
   - Parse terrain index, elevation, flags, resources, optional RGBA
   - Assign hex coordinates via `hexg.NewOddQCoord`/`hexg.NewOddRCoord`
8. **MapKey** - Copy with RGBA decoding
9. **Features** - Copy with nested location and label, RGBA decoding
10. **Labels** - Copy with location and RGBA decoding
11. **Shapes** - Copy with points, RGBA fields
12. **Notes** - Copy InnerText
13. **Informations** - Copy with nested details
14. **Configuration** - Copy terrain/feature/texture/text/shape configs with label and shape styles

**Map_t changes:** If W2025 introduces new elements or attributes that Map_t doesn't have fields for:
- Add the new fields to `Map_t` in `map.go`
- Use `json:"...,omitempty"` tags to avoid breaking H2017 round-trips
- Document what was added and why in the commit message

**File:** `xmlio/internal/v1_06/decode.go`

### Step 3. Add helper functions

Depending on the schema diff, we may need:

- **`decodeRgba` / `decodeZeroableRgba`** - If W2025 uses the same RGBA string format (`"R,G,B,A"`), share or duplicate from v0_77. If the format changed, write W2025-specific versions.
- **Tile row parser** - If the tile format changed, write a dedicated parser. If it's the same, the logic from v0_77 can be adapted.
- **New type converters** - For any new W2025-specific data types.

**File:** `xmlio/internal/v1_06/schema.go` (alongside types, as in v0_77)

### Step 4. Update version dispatch (if needed)

The dispatcher in `xmlio/decoder.go:234` already handles `"2025/1.10/1.01"`. If Worldographer 2025 has shipped newer version strings since, add them:

```go
case "2025/1.10/1.01", "2025/1.11/1.01", ...:
    return v1_06.Decode(data)
```

**File:** `xmlio/decoder.go`

### Step 5. Write tests

Create `xmlio/internal/v1_06/decode_test.go` with:

1. **Round-trip test** - Decode a W2025 file, verify `Map_t` is populated (non-nil tiles, terrain map, features, etc.)
2. **Metadata test** - Verify `DataVersion.Major == 2025`, `Worldographer.Release == "2025"`, `Worldographer.Schema == "1.01"`
3. **Tile count test** - Decode and verify `Tiles.TilesWide` and `Tiles.TilesHigh` match expected values
4. **Terrain map test** - Verify terrain labels and indices are parsed correctly
5. **Feature/label/shape counts** - Verify expected counts from the sample file
6. **Integration test via xmlio.Decoder** - Test the full pipeline (gunzip -> UTF-16 -> header -> dispatch -> v1_06.Decode) using a W2025 `.wxx` file

Since test data files are gitignored, tests should skip gracefully if the sample file is missing:

```go
func TestDecode(t *testing.T) {
    path := filepath.Join("..", "..", "testdata", "sample-w2025.wxx")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Skip("test data not available: ", path)
    }
    // ... test logic ...
}
```

**File:** `xmlio/internal/v1_06/decode_test.go`

### Step 6. Validate with CLI tools

Use existing CLI tools to verify end-to-end:

```bash
# Should print W2025 schema info, tile dimensions, terrain count
go build -o dist/local/info ./cmd/info
dist/local/info testdata/<sample>.wxx

# Should print map dimensions
go build -o dist/local/bounds ./cmd/bounds
dist/local/bounds testdata/<sample>.wxx

# Should dump schema hierarchy
go build -o dist/local/schema ./cmd/schema
dist/local/schema testdata/<sample>.wxx
```

If `cmd/copy` is used with a W2025 file, it will attempt encode - that's a separate task (W2025 encoder). For now, verify decode-only.

---

## Map_t Impact Assessment

Based on what we know about the W2025 format:

**Already handled:**
- `MetaData.Worldographer.Release` - exists in Map_t, populated from `<map release="2025">`
- `MetaData.Worldographer.Schema` - exists in Map_t, populated from `<map schema="1.01">`

**Likely additions (TBD after schema dump):**
- New map-level attributes (W2025 may have additional configuration)
- New tile fields (W2025 may store additional per-tile data)
- New element types entirely (W2025 is a major version jump)

**Design principle:** Map_t is a superset. Add fields with `omitempty` - don't remove or rename existing fields. H2017 decode/encode must continue to work unchanged.

---

## Definition of Done

- [ ] W2025 sample file in `testdata/` (local only, gitignored)
- [ ] Schema dump documented in `xmlio/internal/v1_06/SCHEMA_DUMP.md`
- [ ] Schema diff documented in `xmlio/internal/v1_06/SCHEMA_DIFF.md`
- [ ] `xmlio/internal/v1_06/schema.go` with XML unmarshal types
- [ ] `xmlio/internal/v1_06/decode.go` with full Decode implementation
- [ ] `xmlio/internal/v1_06/decode_test.go` with tests (skip if no test data)
- [ ] Any Map_t additions in `map.go` (if needed)
- [ ] `cmd/info` successfully reads and displays W2025 file info
- [ ] `go test ./...` passes (including existing H2017 tests)
- [ ] `go vet ./...` clean

---

## Out of Scope

- **W2025 encoder** (`internal/v1_06/encode.go`) - separate task, depends on decode being done first
- **Map_t redesign** - only additive changes; structural redesign is a separate task
- **SQLite3 store** - depends on Map_t stability
- **DSL/scripting** - depends on store and Map_t

---

## Risk & Open Questions

1. **No sample file yet.** Everything after P1 depends on having a real W2025 file. If the file format has changed significantly from what the dispatcher expects (`2025/1.10/1.01`), we may need to update the dispatch table first.

2. **Tile row format uncertainty.** The H2017 tile row parser handles 4 variants (6/7/11/12 fields). W2025 may use a different encoding or additional fields. This is the highest-risk section.

3. **Go's `encoding/xml` and XML 1.1.** The XML header is stripped before parsing, so Go's XML 1.0 parser should work. But if W2025 uses XML 1.1-specific features in the body (e.g. additional character references), we may hit issues. Low risk but worth noting.

4. **Version string evolution.** If Worldographer 2025 has shipped updates since the `1.10/1.01` combination, we'll need to decide whether to handle them all in `v1_06` or create additional schema packages.
