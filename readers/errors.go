// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package readers

// Error implements constant errors
type Error string

// Error implements the Errors interface
func (e Error) Error() string {
	return string(e)
}

const (
	ErrMissingBOM               = Error("missing bom")
	ErrMissingFinalByte         = Error("missing final byte")
	ErrMissingXMLHeader         = Error("missing xml header")
	ErrNotBigEndianUTF16Encoded = Error("not big-endian utf-16 encoded")
	ErrUnsupportedVersion       = Error("unsupported version")
)
