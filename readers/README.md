# readers

Package `readers` provides readers for gzip-compressed UTF-16 big-endian data files. This is the format used by WXX map files.

## Overview

Worldographer is a map-generator application that stores data as XML in compressed (gzip) and UTF-16 big-endian encoded files. This package provides:

- **WxxReader**: Read and decompress WXX files, converting UTF-16 to UTF-8
- **OpenWxx**: Factory function for creating reader from a file

## Installation

```bash
go get github.com/maloquacious/wxx/readers
```

## Usage

### Reading WXX Files

```go
package main

import (
    "fmt"
    "io"
    "log"
    
    "github.com/maloquacious/wxx/readers"
)

func main() {
    // Open a Worldographer file
    rc, err := readers.OpenWxx("map.wxx")
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
    
    "github.com/maloquacious/wxx/readers"
)

func main() {
    // Open file manually
    file, err := os.Open("map.wxx")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    // Create Reader from existing reader
    rc, err := readers.NewWxxReader(file)
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

## API Reference

### WxxReader

```go
type WxxReader struct {
    // implements io.Reader
}

func OpenWxx(path string) (*WxxReader, error)
func NewWxxReader(r io.Reader) (*WxxReader, error)
```

The `WxxReader` type implements the `io.Reader` interface and automatically:
- Decompresses gzip data
- Converts UTF-16 big-endian to UTF-8
- Handles byte order marks (BOM) for both big-endian and little-endian UTF-16

## WXX Map File Format Details

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
