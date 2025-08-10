// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package readers_test

//import (
//	"bytes"
//	"compress/gzip"
//	"github.com/maloquacious/wxx/readers"
//	"golang.org/x/text/encoding/unicode"
//	"io"
//	"testing"
//)
//
//func TestNewWxxReader(t *testing.T) {
//	// Create test data - Lorem Ipsum style text
//	originalText := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
//Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.
//At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident, similique sunt in culpa qui officia deserunt mollitia animi, id est laborum et dolorum fuga. Et harum quidem rerum facilis est et expedita distinctio.`
//
//	t.Run("valid_gzip_utf16", func(t *testing.T) {
//		// Convert to UTF-16 big-endian
//		utf16Encoding := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
//		utf16Data, err := utf16Encoding.NewEncoder().Bytes([]byte(originalText))
//		if err != nil {
//			t.Fatalf("Failed to encode to UTF-16: %v", err)
//		}
//
//		// Compress with gzip
//		var gzipBuffer bytes.Buffer
//		gzipWriter := gzip.NewWriter(&gzipBuffer)
//		_, err = gzipWriter.Write(utf16Data)
//		if err != nil {
//			t.Fatalf("Failed to write to gzip: %v", err)
//		}
//		err = gzipWriter.Close()
//		if err != nil {
//			t.Fatalf("Failed to close gzip writer: %v", err)
//		}
//
//		// Create Reader from compressed data
//		compressedReader := bytes.NewReader(gzipBuffer.Bytes())
//		rc, err := readers.NewWxxReader(compressedReader)
//		if err != nil {
//			t.Fatalf("NewWxxReader() error = %v", err)
//		}
//
//		// Read all data from the reader
//		actualData, err := io.ReadAll(rc)
//		if err != nil {
//			t.Fatalf("ReadAll() error = %v", err)
//		}
//
//		// Compare the data
//		if string(actualData) != originalText {
//			t.Errorf("Data mismatch")
//			t.Errorf("Expected length: %d", len(originalText))
//			t.Errorf("Actual length: %d", len(actualData))
//
//			// Show first few characters for debugging
//			expectedStr := originalText
//			actualStr := string(actualData)
//
//			if len(expectedStr) > 200 {
//				expectedStr = expectedStr[:200]
//			}
//			if len(actualStr) > 200 {
//				actualStr = actualStr[:200]
//			}
//
//			t.Errorf("Expected first 200 chars: %q", expectedStr)
//			t.Errorf("Actual first 200 chars: %q", actualStr)
//		}
//	})
//
//	t.Run("non_gzip_input", func(t *testing.T) {
//		// Create non-gzip data (just plain text)
//		nonGzipReader := bytes.NewReader([]byte(originalText))
//
//		// This should fail because the data isn't gzip compressed
//		_, err := readers.NewWxxReader(nonGzipReader)
//		if err == nil {
//			t.Fatal("Expected error for non-gzip input, but got none")
//		}
//
//		// Verify it's a gzip-related error
//		if !bytes.Contains([]byte(err.Error()), []byte("gzip")) {
//			t.Errorf("Expected gzip-related error, got: %v", err)
//		}
//	})
//
//	t.Run("non_utf16_input", func(t *testing.T) {
//		// Create gzip-compressed UTF-8 data (not UTF-16)
//		var gzipBuffer bytes.Buffer
//		gzipWriter := gzip.NewWriter(&gzipBuffer)
//		_, err := gzipWriter.Write([]byte(originalText)) // UTF-8 data, not UTF-16
//		if err != nil {
//			t.Fatalf("Failed to write to gzip: %v", err)
//		}
//		err = gzipWriter.Close()
//		if err != nil {
//			t.Fatalf("Failed to close gzip writer: %v", err)
//		}
//
//		// Create Reader from compressed UTF-8 data
//		compressedReader := bytes.NewReader(gzipBuffer.Bytes())
//
//		// Creating the Reader should fail because UTF-8 data lacks the required UTF-16 BOM.
//		// This test verifies the reader properly handles non-UTF-16 data.
//		_, err = readers.NewWxxReader(compressedReader)
//		if err == nil {
//			t.Error("Expected error when reading UTF-8 data as UTF-16, but got none")
//		} else {
//			// Verify it's a BOM-related error
//			if !bytes.Contains([]byte(err.Error()), []byte("byte order mark")) {
//				t.Errorf("Expected BOM-related error, got: %v", err)
//			}
//		}
//	})
//
//	t.Run("utf16_little_endian_input", func(t *testing.T) {
//		// Create UTF-16 little-endian data (not big-endian as expected)
//		utf16LEEncoding := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
//		utf16LEData, err := utf16LEEncoding.NewEncoder().Bytes([]byte(originalText))
//		if err != nil {
//			t.Fatalf("Failed to encode to UTF-16 LE: %v", err)
//		}
//
//		// Compress with gzip
//		var gzipBuffer bytes.Buffer
//		gzipWriter := gzip.NewWriter(&gzipBuffer)
//		_, err = gzipWriter.Write(utf16LEData)
//		if err != nil {
//			t.Fatalf("Failed to write to gzip: %v", err)
//		}
//		err = gzipWriter.Close()
//		if err != nil {
//			t.Fatalf("Failed to close gzip writer: %v", err)
//		}
//
//		// Create Reader from compressed UTF-16 LE data
//		compressedReader := bytes.NewReader(gzipBuffer.Bytes())
//
//		// Create the Reader from the gzip reader
//		rc, err := readers.NewWxxReader(compressedReader)
//		if err != nil {
//			t.Fatalf("NewReadCloser() should not fail for gzip data: %v", err)
//		}
//
//		// Reading should not work, but we might have a bug - the BOM in UTF-16 LE data tells the decoder the byte order
//		// and the unicode.UTF16 decoder with ExpectBOM automatically handles both BE and LE encodings.
//		actualData, err := io.ReadAll(rc)
//		if err != nil {
//			t.Fatalf("ReadAll() error = %v", err)
//		}
//
//		// The data should be the same as original text because BOM detection works
//		if string(actualData) != originalText {
//			t.Error("Expected identical text when UTF-16 LE data is properly decoded via BOM detection")
//		}
//
//		// Verify we got the expected data
//		if len(actualData) == 0 {
//			t.Error("Expected decoded data, but got empty result")
//		}
//	})
//}
