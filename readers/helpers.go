// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package readers

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"
)

// utf16ToUtf8 converts big-endian UTF-16 data to UTF-8
func utf16ToUtf8(src []byte) ([]byte, error) {
	// convert the slice of byte to a slice of uint16
	chars := make([]uint16, len(src)/2)
	if err := binary.Read(bytes.NewReader(src), binary.BigEndian, &chars); err != nil {
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
