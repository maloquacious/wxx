// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package readers

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"
)

// ReadUTF16 is a helper to extract UTF-16 data from the input.
func ReadUTF16(data []byte) ([]byte, error) {
	if len(data)%2 != 0 {
		// UTF-16 data must contain an even number of bytes
		return nil, ErrMissingFinalByte
	}
	// verify the BOM
	if bytes.HasPrefix(data, []byte{0xfe, 0xff}) {
		// as expected
	} else if bytes.HasPrefix(data, []byte{0xff, 0xfe}) {
		return nil, ErrNotBigEndianUTF16Encoded
	} else {
		return nil, ErrMissingBOM
	}
	// consume the BOM
	data = data[2:]
	// convert the slice of byte to a slice of uint16
	chars := make([]uint16, len(data)/2)
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, &chars); err != nil {
		return nil, err
	}
	// create a buffer for the results
	dst := bytes.Buffer{}
	// convert the UTF-16 to runes, then to UTF-8 bytes
	var utfBuffer [utf8.UTFMax]byte
	for _, r := range utf16.Decode(chars) {
		utf8Size := utf8.EncodeRune(utfBuffer[:], r)
		dst.Write(utfBuffer[:utf8Size])
	}
	// finally, return the results
	return dst.Bytes(), nil
}
