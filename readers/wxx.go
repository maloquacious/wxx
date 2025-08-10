// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package readers

//import (
//	"bytes"
//	"compress/gzip"
//	"golang.org/x/text/encoding/unicode"
//	"golang.org/x/text/transform"
//	"io"
//	"os"
//)
//
//// NewWxxReader creates a new WxxReader from an existing io.Reader.
//// Uncompresses and transforms the input from UTF-16/BE encoding to UTF-8.
//// Returns any errors from uncompressing or decoding the input.
//// Loads the entire input into memory.
//func NewWxxReader(input io.Reader) (*WxxReader, error) {
//	// the input must be compressed using Gzip. it's an error if it isn't.
//	gz, err := gzip.NewReader(input)
//	if err != nil {
//		return nil, err
//	}
//	// ensure we close the gzip reader to avoid leaking resources.
//	defer func() {
//		_ = gz.Close()
//	}()
//
//	// return an error if the uncompressed data is not UTF-16/BE encoded.
//	// Note: ExpectBOM allows automatic detection of byte order from BOM. This might be a bug.
//	utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
//
//	// create a reader that will transform the input stream from UTF-16/BE to UTF-8.
//	utf8r := transform.NewReader(gz, utf16Encoding.NewDecoder())
//
//	// transform the entire input, saving it to a buffer. return if there are errors.
//	data, err := io.ReadAll(utf8r)
//	if err != nil {
//		return nil, err
//	}
//
//	// return the new reader
//	return &WxxReader{
//		buf: bytes.NewBuffer(data),
//	}, nil
//}
//
//// OpenWxx is a convenience function for creating a new
//// WxxReader from a file. Returns any errors opening or
//// reading the file.
//func OpenWxx(path string) (*WxxReader, error) {
//	// open the file for reading. return any errors.
//	fp, err := os.Open(path)
//	if err != nil {
//		return nil, err
//	}
//	// ensure we close the file reader to avoid leaking resources.
//	defer func() {
//		_ = fp.Close()
//	}()
//	// return the new reader
//	return NewWxxReader(fp)
//}
//
//// WxxReader implements the io.Reader interface for WXX data files.
//// Note that the entire input stream is buffered in memory when the reader is created.
//type WxxReader struct {
//	buf *bytes.Buffer
//}
//
//// Read reads up to len(p) bytes from the underlying UTF-16 data stream,
//// converting it to UTF-8 and storing the result in p.
//// It returns the number of bytes read and any error encountered.
//func (r *WxxReader) Read(p []byte) (n int, err error) {
//	if r == nil || r.buf == nil {
//		return 0, io.ErrClosedPipe
//	}
//	return r.buf.Read(p)
//}
//
//// Close is a no-op
//func (r *WxxReader) Close() error {
//	if r != nil && r.buf != nil {
//		r.buf = nil
//	}
//	return nil
//}
