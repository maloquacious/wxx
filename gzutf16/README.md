# gzutf16

Package `gzutf16` provides readers and writers for gzip-compressed UTF-16 big-endian data files. This is the format used by Worldographer 2017 and Worldographer 2025 map files (WXX format).

## Overview

Worldographer is a map-generator application that stores data as XML in compressed (gzip) and UTF-16 big-endian encoded files. This package provides:

- **ReadCloser**: Read and decompress WXX files, converting UTF-16 to UTF-8
- **Buffer**: Build XML content and convert it to the proper WXX format
- **Open/NewReadCloser**: Factory functions for creating readers
- **WriteFile**: Write UTF-8 data directly to WXX files

## Installation

```bash
go get github.com/maloquacious/wxx/gzutf16
```

## Usage

### Reading WXX Files

```go
package main

import (
    "fmt"
    "io"
    "log"
    
    "github.com/maloquacious/wxx/gzutf16"
)

func main() {
    // Open a Worldographer file
    rc, err := gzutf16.Open("map.wxx")
    if err != nil {
        log.Fatal(err)
    }
    defer rc.Close()
    
    // Read the XML content as UTF-8
    data, err := io.ReadAll(rc)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("XML content: %s\n", string(data))
}
```

### Reading from io.ReadCloser

```go
package main

import (
    "fmt"
    "io"
    "log"
    "os"
    
    "github.com/maloquacious/wxx/gzutf16"
)

func main() {
    // Open file manually
    file, err := os.Open("map.wxx")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    // Create ReadCloser from existing reader
    rc, err := gzutf16.NewReadCloser(file)
    if err != nil {
        log.Fatal(err)
    }
    defer rc.Close()
    
    // Read the content
    data, err := io.ReadAll(rc)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("XML content: %s\n", string(data))
}
```

### Writing WXX Files

```go
package main

import (
    "log"
    "os"
    
    "github.com/maloquacious/wxx/gzutf16"
)

func main() {
    var buf gzutf16.Buffer
    
    // Build XML content using Write and Printf
    buf.Write([]byte(`<?xml version="1.1" encoding="utf-16"?>`))
    buf.Printf(`<map type="WORLD" version="%s">`, "1.10")
    buf.Write([]byte(`<tiles viewLevel="WORLD">`))
    buf.Write([]byte(`</tiles>`))
    buf.Write([]byte(`</map>`))
    
    // Convert to gzip-compressed UTF-16 big-endian format
    data, err := buf.Bytes()
    if err != nil {
        log.Fatal(err)
    }
    
    // Write to file
    err = os.WriteFile("output.wxx", data, 0644)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Writing WXX Files (Alternative)

```go
package main

import (
    "log"
    
    "github.com/maloquacious/wxx/gzutf16"
)

func main() {
    // Build XML content as UTF-8
    xmlData := []byte(`<?xml version="1.1" encoding="utf-16"?>
<map type="WORLD" version="1.10">
    <tiles viewLevel="WORLD">
        <!-- tile data -->
    </tiles>
</map>`)
    
    // Write directly to WXX file
    err := gzutf16.WriteFile("output.wxx", xmlData, 0644)
    if err != nil {
        log.Fatal(err)
    }
}
```

## API Reference

### ReadCloser

```go
type ReadCloser struct {
    // implements io.ReadCloser
}

func Open(path string) (*ReadCloser, error)
func NewReadCloser(r io.ReadCloser) (*ReadCloser, error)
```

The `ReadCloser` type implements the `io.ReadCloser` interface and automatically:
- Decompresses gzip data
- Converts UTF-16 big-endian to UTF-8
- Handles byte order marks (BOM) for both big-endian and little-endian UTF-16

### Buffer

```go
type Buffer struct {
    // internal buffer
}

func (b *Buffer) Write(p []byte) (n int, err error)
func (b *Buffer) Printf(format string, args ...interface{}) (n int, err error)
func (b *Buffer) Bytes() ([]byte, error)
func (b *Buffer) Len() int
func (b *Buffer) Reset()
func (b *Buffer) String() string
```

The `Buffer` type acts like `bytes.Buffer` but with automatic conversion:
- Accumulates UTF-8 text via `Write` and `Printf`
- `Bytes()` converts to UTF-16 big-endian and compresses with gzip
- Suitable for building XML content for WXX files
- Provides `Len()`, `Reset()`, and `String()` methods for buffer management

### WriteFile

```go
func WriteFile(name string, data []byte, perm os.FileMode) error
```

The `WriteFile` function provides a convenient way to write UTF-8 data directly to WXX files:
- Converts UTF-8 data to UTF-16 big-endian and compresses with gzip
- API mirrors `os.WriteFile` for familiarity
- Handles all format conversion automatically

## File Format Details

Worldographer files use a specific format:
1. **Compression**: gzip
2. **Encoding**: UTF-16 big-endian with BOM
3. **Content**: XML data

The package handles the format conversion automatically, allowing you to work with standard UTF-8 strings in Go.

## Error Handling

The package returns specific errors for:
- **Non-gzip data**: When input is not gzip-compressed
- **Missing BOM**: When UTF-16 data lacks a byte order mark
- **Invalid UTF-16**: When data cannot be decoded as UTF-16

## Testing

Run the test suite:

```bash
go test -v
```

The tests use synthetic data to ensure reliable testing without external file dependencies.

## License

Copyright (c) 2025 Michael D Henderson. All rights reserved.
