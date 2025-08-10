// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package gzutf16

//import (
//	"bytes"
//	"compress/gzip"
//	"fmt"
//	"golang.org/x/text/encoding/unicode"
//)
//
//// Buffer acts like bytes.Buffer but converts output to gzip-compressed UTF-16 big-endian format.
//// This is intended for building XML output that will be written to Worldographer WXX files.
//type Buffer struct {
//	buf bytes.Buffer
//}
//
//// Write appends the contents of p to the buffer. It always returns len(p), nil.
//func (b *Buffer) Write(p []byte) (n int, err error) {
//	if b == nil {
//		return 0, fmt.Errorf("write to nil buffer")
//	}
//	return b.buf.Write(p)
//}
//
//// Printf formats according to a format specifier and writes to the buffer.
//// It returns the number of bytes written and any write error encountered.
//func (b *Buffer) Printf(format string, args ...interface{}) (n int, err error) {
//	if b == nil {
//		return 0, fmt.Errorf("write to nil buffer")
//	}
//	return fmt.Fprintf(&b.buf, format, args...)
//}
//
//// Bytes converts the buffer contents to UTF-16 big-endian, compresses with gzip,
//// and returns the compressed data. If there are errors during encoding or compression,
//// it returns the error.
//func (b *Buffer) Bytes() ([]byte, error) {
//	if b == nil {
//		return nil, fmt.Errorf("bytes from nil buffer")
//	}
//
//	// Get the UTF-8 data from the buffer
//	utf8Data := b.buf.Bytes()
//
//	// Convert to UTF-16 big-endian
//	utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
//	utf16Data, err := utf16Encoding.NewEncoder().Bytes(utf8Data)
//	if err != nil {
//		return nil, fmt.Errorf("failed to encode to UTF-16: %w", err)
//	}
//
//	// Compress with gzip
//	var gzipBuffer bytes.Buffer
//	gzipWriter := gzip.NewWriter(&gzipBuffer)
//
//	_, err = gzipWriter.Write(utf16Data)
//	if err != nil {
//		return nil, fmt.Errorf("failed to write to gzip: %w", err)
//	}
//
//	err = gzipWriter.Close()
//	if err != nil {
//		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
//	}
//
//	return gzipBuffer.Bytes(), nil
//}
//
//// Len returns the number of bytes in the buffer.
//func (b *Buffer) Len() int {
//	return b.buf.Len()
//}
//
//// Reset resets the buffer to be empty.
//func (b *Buffer) Reset() {
//	b.buf.Reset()
//}
//
//// String returns the contents of the buffer as a string.
//func (b *Buffer) String() string {
//	return b.buf.String()
//}
