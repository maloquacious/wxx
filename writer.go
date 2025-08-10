// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

import "io"

// todo: i don't really know the right way to do the interfaces

// Note that implementations are expected to use an XML schema for
// a specific version of Worldographer when marshalling the Map_t to XML.

// Writer marshals the Map_t to XML, adds an XML header with the version and encoding,
// converts it to UTF-16/BE, compresses it with gzip, and returns a writer or an error.
type Writer interface {
	Write(*Map_t) (io.Writer, error)
}

// BytesWriter is intended for debugging and can return a writer with invalid data and errors.
//
// Create a slice of byte with an XML header with the version and encoding (which will always be UTF-8).
//
// Convert the Map_t to a slice of byte.
// If there are errors, return nil with an error.
// Append this slice to the XML header.
//
// Return a writer with the data and no error.
type BytesWriter interface {
	BytesWriter(*Map_t) (io.Writer, error)
}

// CompressedWriter returns a writer for the compressed output or an error.
//
// Converts the Map_t to XML, adds an XML header with the version and encoding,
// converts it to UTF-16/BE, compresses it with gzip, and returns a writer or an error.
//
// This is similar to the Writer interface, but may return different errors.
type CompressedWriter interface {
	WriteCompressed(*Map_t) (io.Writer, error)
}

// Utf8Writer converts the Map_t to XML and returns a writer or an error.
type Utf8Writer interface {
	WriteUtf8(*Map_t) (io.Writer, error)
}

// Utf8XmlWriter does
//
// Convert the Map_t to XML (actually a slice of bytes containing the XML structure) using
// the schema
type Utf8XmlWriter interface {
	WriteUtf8Xml(*Map_t) (io.Writer, error)
}

// Utf16Writer converts the Map_t to XML, converts to UTF-16/BE, and returns a writer or an error.
type Utf16Writer interface {
	WriteUtf16(*Map_t) (io.Writer, error)
}
