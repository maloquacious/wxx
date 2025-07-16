// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

// UTF8ToUTF16 converts a slice of UTF-8 encoded bytes into a slice of
// UTF-16 encoded bytes with a Big Endian Byte Order Mark (BOM).
//
// It takes as input a slice of bytes `src` which should be UTF-8 encoded data.
//
// Algorithm:
// - Creates a buffer and writes the Big Endian BOM (0xFE, 0xFF) to it.
// - Iterates over the input slice, decoding UTF-8 runes.
//   - If a rune is utf8.RuneError, it means the data is not a valid UTF-8 encoded rune and an error is returned.
//   - The decoded rune is then converted to UTF-16 format using the utf16.Encode function and written to the buffer.
//
// - Continues until all the input byte slice has been processed.
// - Returns a byte slice of UTF-16 encoded data from the buffer if no errors occurred during processing, else returns an error.
//
// Returns:
// - A slice of UTF-16 encoded bytes if successful.
// - An error in case of invalid UTF-8 data, or error when trying to write to a buffer.
//
// Example:
//
//	 utf8Data := []byte("Hello, World!")
//	 utf16Data, err := Stage97(utf8Data)
//		if err != nil {
//		    log.Fatal(err)
//		}
func UTF8ToUTF16(src []byte) ([]byte, error) {
	// create a buffer for the results, including the BOM
	dst := bytes.Buffer{}
	dst.Write([]byte{0xfe, 0xff})

	// convert the source
	for len(src) > 0 {
		// extract next rune from the source
		r, w := utf8.DecodeRune(src)
		if r == utf8.RuneError {
			return nil, fmt.Errorf("invalid utf8 data")
		}
		// consume that rune
		src = src[w:]
		// convert the rune to UTF-16 and write it to the results
		for _, v := range utf16.Encode([]rune{r}) {
			if err := binary.Write(&dst, binary.BigEndian, v); err != nil {
				return nil, err
			}
		}
	}

	// return the results
	return dst.Bytes(), nil
}
