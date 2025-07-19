// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package models

// Error implements constant errors
type Error string

// Error implements the Errors interface
func (e Error) Error() string {
	return string(e)
}

const (
	ErrInvalidMapMetadata       = Error("invalid map metadata")
	ErrInvalidXML               = Error("invalid xml")
	ErrMissingBOM               = Error("missing bom")
	ErrMissingFinalByte         = Error("missing final byte")
	ErrMissingXMLHeader         = Error("missing xml header")
	ErrNotBigEndianUTF16Encoded = Error("not big-endian utf-16 encoded")
	ErrUnsupportedMapMetadata   = Error("unsupported map metadata")
	ErrUnsupportedMapRelease    = Error("unsupported map release")
	ErrUnsupportedMapSchema     = Error("unsupported map schema")
	ErrUnsupportedMapVersion    = Error("unsupported map version")
	ErrUnsupportedSchemaVersion = Error("unsupported schema version")
)
