// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package gzutf16 implements readers and writers for compressed (gzip),
// UTF-16 (big-endian) data files. That's the format that Worldographer
// 2017 and 2025 files use.
//
// ReadCloser implements the io.ReadCloser interface and returns the file data
// as an uncompressed UTF-8 stream.
package gzutf16

import (
	"compress/gzip"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"os"
)

// Open returns a ReadCloser. If there are errors (the path can't be read,
// the file isn't a gzip file, or the uncompressed data isn't encoded as
// UTF-16 in big-endian format), we return the error.
func Open(path string) (rc *ReadCloser, err error) {
	// open the file for reading. return any errors.
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// use NewReadCloser to create the actual reader
	rc, err = NewReadCloser(fp)
	if err != nil {
		_ = fp.Close()
		return nil, err
	}

	return rc, nil
}

// NewReadCloser creates a ReadCloser from an existing io.ReadCloser.
// The input reader must contain gzip-compressed UTF-16 big-endian data.
// If there are errors (the data isn't a gzip file, or the uncompressed data 
// isn't encoded as UTF-16 in big-endian format), we return the error.
func NewReadCloser(r io.ReadCloser) (rc *ReadCloser, err error) {
	// ensure that any Readers we create get closed if we exit early due to any error.
	defer func() {
		if err != nil && rc != nil {
			_ = rc.Close()
		}
	}()

	rc = &ReadCloser{
		fp: r,
	}

	// the data must be compressed using Gzip. it's an error if it isn't.
	rc.gz, err = gzip.NewReader(rc.fp)
	if err != nil {
		return nil, err
	}

	// the data must be UTF-16 big-endian encoded. it's an error if it isn't.
	// Note: ExpectBOM allows automatic detection of byte order from BOM

	// create a reader that will transform the input stream to UTF-8.
	utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
	rc.utf8r = transform.NewReader(rc.gz, utf16Encoding.NewDecoder())

	return rc, nil
}

// ReadCloser implements the io.ReadCloser interface for compressed UTF-16 data files.
type ReadCloser struct {
	utf8r *transform.Reader // transforms UTF-16 big-endian to UTF-8
	gz    *gzip.Reader      // gzip decompression stream
	fp    io.ReadCloser     // input data stream
}

// Read reads up to len(p) bytes from the underlying UTF-16 data stream,
// converting it to UTF-8 and storing the result in p.
// It returns the number of bytes read and any error encountered.
func (rc *ReadCloser) Read(p []byte) (n int, err error) {
	if rc.utf8r == nil {
		return 0, io.ErrClosedPipe
	}
	return rc.utf8r.Read(p)
}

// Close closes the ReadCloser and releases associated resources.
// If multiple errors occur during closing, it returns the first error encountered.
func (rc *ReadCloser) Close() (err error) {
	// utf-8 transformer is not a ReadCloser, so we don't need to close it
	rc.utf8r = nil
	
	// close the gzip reader only if it is open
	if rc.gz != nil {
		if closeErr := rc.gz.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		rc.gz = nil
	}
	
	// close the file reader only if it is open
	if rc.fp != nil {
		if closeErr := rc.fp.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		rc.fp = nil
	}
	
	return err
}
