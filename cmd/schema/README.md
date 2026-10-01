# schema
`schema` is a command-line tool for analyzing Worldographer data files (WXX files) and generating Go structs that match their XML schema.

## Overview

Worldographer stores map data as GZip-compressed, UTF-16 encoded XML files. This tool:

1. **Reads WXX files** - Decompresses and converts UTF-16 to UTF-8
2. **Infers XML schema** - Analyzes the XML structure to understand elements and attributes  
3. **Generates Go structs** - Creates typed structs that can marshal/unmarshal the XML data

## Current Status

The tool currently:
- ✅ Reads and decompresses WXX files
- ✅ Detects XML schema version (1.0 vs 1.1)
- ✅ Infers element hierarchy and attributes
- ✅ Outputs XML hierarchy for analysis

**In Development:**
- 🔄 Generate Go structs with proper naming conventions
- 🔄 Add XML tags for marshaling/unmarshaling
- 🔄 Handle nested structs and collections

## Usage

```bash
go run ./cmd/schema testdata/2025-2.06-13x11-941577-blank.wxx
```

Takes one or more `.wxx` files as arguments and prints the inferred XML
hierarchy to the console. Pass `-sql` to emit `CREATE TABLE` statements for the
inferred schema instead.

## XML Schema Versions

The tool supports Worldographer 2025 files (XML 1.1), which state
`release="2025"`, the application version and the schema version in the `map`
element, e.g. `release="2025" version="2.06" schema="1.06"`. It reports them
before the hierarchy:

```
	W2025: version 2.06: schema 1.06
```

A classic (Worldographer 1.x, XML 1.0) file states no release and no schema.
wxx no longer reads classic files (issue #103), and `schema` refuses one as the
decoder does, after the transport summary and before inferring anything:

```bash
go run ./cmd/schema testdata/2017-1.77-1.0-columns-blank.wxx
```
```
	classic (Worldographer 1.x) map: convert it in Worldographer 2025 first
map: version "1.77": no release
```

## Generated Code Conventions

- **Structs**: Use `_t` suffix (e.g., `Map_t`, `Configuration_t`)
- **Interfaces**: Use `_i` suffix when needed
- **No anonymous structs**: All child elements are typed as named structs
- **XML tags**: Include proper `xml:"elementName"` tags

## Future Features

- Command-line flags for input/output files
- Type inference for attributes (string, int, bool)
- Support for repeated elements as slices
- SQLite schema generation
- Validation of generated structs against original XML
