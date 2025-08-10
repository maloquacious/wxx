// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package gzutf16

//import (
//	"bytes"
//	"io"
//	"testing"
//)
//
//func TestBuffer(t *testing.T) {
//	// Create test data - Lorem Ipsum style text
//	originalText := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
//
//Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.
//
//At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident, similique sunt in culpa qui officia deserunt mollitia animi, id est laborum et dolorum fuga. Et harum quidem rerum facilis est et expedita distinctio.`
//
//	t.Run("write_and_bytes", func(t *testing.T) {
//		var buf Buffer
//
//		// Write data to buffer
//		n, err := buf.Write([]byte(originalText))
//		if err != nil {
//			t.Fatalf("Write() error = %v", err)
//		}
//		if n != len(originalText) {
//			t.Errorf("Write() returned %d, expected %d", n, len(originalText))
//		}
//
//		// Get compressed bytes
//		compressedData, err := buf.Bytes()
//		if err != nil {
//			t.Fatalf("Bytes() error = %v", err)
//		}
//
//		// Verify we got some compressed data
//		if len(compressedData) == 0 {
//			t.Error("Expected compressed data, got empty slice")
//		}
//
//		// Verify the data can be read back correctly using ReadCloser
//		compressedReader := io.NopCloser(bytes.NewReader(compressedData))
//		rc, err := NewReadCloser(compressedReader)
//		if err != nil {
//			t.Fatalf("NewReadCloser() error = %v", err)
//		}
//		defer rc.Close()
//
//		actualData, err := io.ReadAll(rc)
//		if err != nil {
//			t.Fatalf("ReadAll() error = %v", err)
//		}
//
//		if string(actualData) != originalText {
//			t.Errorf("Round-trip failed: expected %q, got %q", originalText, string(actualData))
//		}
//	})
//
//	t.Run("printf_and_bytes", func(t *testing.T) {
//		var buf Buffer
//
//		// Use Printf to write formatted data
//		n, err := buf.Printf("Hello %s! Number: %d", "World", 42)
//		if err != nil {
//			t.Fatalf("Printf() error = %v", err)
//		}
//		expectedText := "Hello World! Number: 42"
//		if n != len(expectedText) {
//			t.Errorf("Printf() returned %d, expected %d", n, len(expectedText))
//		}
//
//		// Get compressed bytes
//		compressedData, err := buf.Bytes()
//		if err != nil {
//			t.Fatalf("Bytes() error = %v", err)
//		}
//
//		// Verify the data can be read back correctly
//		compressedReader := io.NopCloser(bytes.NewReader(compressedData))
//		rc, err := NewReadCloser(compressedReader)
//		if err != nil {
//			t.Fatalf("NewReadCloser() error = %v", err)
//		}
//		defer rc.Close()
//
//		actualData, err := io.ReadAll(rc)
//		if err != nil {
//			t.Fatalf("ReadAll() error = %v", err)
//		}
//
//		if string(actualData) != expectedText {
//			t.Errorf("Round-trip failed: expected %q, got %q", expectedText, string(actualData))
//		}
//	})
//
//	t.Run("mixed_write_and_printf", func(t *testing.T) {
//		var buf Buffer
//
//		// Mix Write and Printf calls
//		_, err := buf.Write([]byte("Start: "))
//		if err != nil {
//			t.Fatalf("Write() error = %v", err)
//		}
//
//		_, err = buf.Printf("Value=%d", 123)
//		if err != nil {
//			t.Fatalf("Printf() error = %v", err)
//		}
//
//		_, err = buf.Write([]byte(" End"))
//		if err != nil {
//			t.Fatalf("Write() error = %v", err)
//		}
//
//		expectedText := "Start: Value=123 End"
//
//		// Get compressed bytes
//		compressedData, err := buf.Bytes()
//		if err != nil {
//			t.Fatalf("Bytes() error = %v", err)
//		}
//
//		// Verify the data can be read back correctly
//		compressedReader := io.NopCloser(bytes.NewReader(compressedData))
//		rc, err := NewReadCloser(compressedReader)
//		if err != nil {
//			t.Fatalf("NewReadCloser() error = %v", err)
//		}
//		defer rc.Close()
//
//		actualData, err := io.ReadAll(rc)
//		if err != nil {
//			t.Fatalf("ReadAll() error = %v", err)
//		}
//
//		if string(actualData) != expectedText {
//			t.Errorf("Round-trip failed: expected %q, got %q", expectedText, string(actualData))
//		}
//	})
//
//	t.Run("buffer_methods", func(t *testing.T) {
//		var buf Buffer
//
//		// Test Len
//		if buf.Len() != 0 {
//			t.Errorf("Expected empty buffer length 0, got %d", buf.Len())
//		}
//
//		// Write some data
//		buf.Write([]byte("Hello"))
//		if buf.Len() != 5 {
//			t.Errorf("Expected buffer length 5, got %d", buf.Len())
//		}
//
//		// Test String
//		if buf.String() != "Hello" {
//			t.Errorf("Expected 'Hello', got %q", buf.String())
//		}
//
//		// Test Reset
//		buf.Reset()
//		if buf.Len() != 0 {
//			t.Errorf("Expected empty buffer after reset, got length %d", buf.Len())
//		}
//		if buf.String() != "" {
//			t.Errorf("Expected empty string after reset, got %q", buf.String())
//		}
//	})
//
//	t.Run("nil_buffer_errors", func(t *testing.T) {
//		var buf *Buffer
//
//		// Test Write with nil buffer
//		_, err := buf.Write([]byte("test"))
//		if err == nil {
//			t.Error("Expected error writing to nil buffer")
//		}
//
//		// Test Printf with nil buffer
//		_, err = buf.Printf("test %d", 42)
//		if err == nil {
//			t.Error("Expected error with Printf on nil buffer")
//		}
//
//		// Test Bytes with nil buffer
//		_, err = buf.Bytes()
//		if err == nil {
//			t.Error("Expected error with Bytes on nil buffer")
//		}
//	})
//}
