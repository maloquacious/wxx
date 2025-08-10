// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

import "io"

// Reader uncompresses the input, converts UTF-16/BE to UTF-8, updates the XML header encoding, and extracts the map data into a Map_t.
// If any step fails or the input isn't in the expected format and encoding, returns nil and an error.
type Reader interface {
	Read(r io.Reader) (*Map_t, error)
}

// BytesReader is intended for debugging and can return a reader with invalid data and errors.
//
// Read the entire input stream into memory.
// If there are errors, return nil with an error.
// Otherwise, use this data for the remaining steps.
//
// If the data is not gzip data, return a reader with the data and an error.
//
// Run gunzip on the data.
// If there are errors, return a reader with the data and an error.
// Otherwise, use the uncompressed data as the data for the remaining steps.
//
// If the input is not UTF-16/BE, return a reader with the data and an error.
//
// Convert the UTF-16/BE data to UTF-8.
// If the conversion fails, return the data and an error.
// Otherwise, use the converted data as the data for the remaining steps.
//
// Return a reader with the data and no error.
type BytesReader interface {
	ReadBytes(r io.Reader) (io.Reader, error)
}

// CompressedReader returns a reader for the uncompressed input or an error.
//
// Run gunzip on the input. If there are errors, return nil and an error.
//
// Return a reader with the uncompressed data and no error.
type CompressedReader interface {
	ReadCompressed(r io.Reader) (io.Reader, error)
}

// Utf8Reader uncompresses the input, converts to UTF-8, and returns a reader or any errors.
//
// Run gunzip on the input. If there are errors, return nil and an error.
//
// Convert the data from UTF-16/BE to UTF-8. If the conversion fails, return nil and an error.
//
// Return a reader with the UTF-8 data and no error.
type Utf8Reader interface {
	ReadUtf8Read(r io.Reader) (io.Reader, error)
}

// Utf8XmlReader uncompresses the input, converts to UTF-8, updates the XML header encoding, and returns a reader or an error.
//
// Run gunzip on the input. If there are errors, return nil and an error.
//
// Convert the data from UTF-16/BE to UTF-8. If the conversion fails, return nil and an error.
//
// If the data does not have an XML header with encoding, return nil and error.
//
// Update the header encoding to utf-8.
//
// Return a reader with the UTF-8 data and no error.
type Utf8XmlReader interface {
	Utf8Reader(r io.Reader) (io.Reader, error)
}
