// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

// Error implements constant errors
type Error string

// Error implements the Errors interface
func (e Error) Error() string {
	return string(e)
}

const (
	ErrInvalidUTF16             = Error("invalid utf-16")
	ErrInvalidXMLHeader         = Error("invalid xml header")
	ErrUnknownXMLHeader         = Error("unknown xml header")
	ErrInvalidXML               = Error("invalid xml")
	ErrMissingMapElement        = Error("missing map element")
	ErrInvalidMapMetadata       = Error("invalid <map> metadata")
	ErrMapNotClosed             = Error("<map> not closed")
	ErrUnsupportedMapMetadata   = Error("unsupported map metadata")
	ErrGZipNewReaderFailed      = Error("gzip new reader failed")
	ErrGUnZipFailed             = Error("gunzip failed")
	ErrGZipFailed               = Error("gzip failed")
	ErrInvalidEncodingHeader    = Error("invalid encoding header")
	ErrInvalidVersion           = Error("invalid version")
	ErrMissingBOM               = Error("missing bom")
	ErrMissingFinalByte         = Error("missing final byte")
	ErrMissingVersion           = Error("missing version")
	ErrMissingXMLHeader         = Error("missing xml header")
	ErrNotBigEndianUTF16Encoded = Error("not big-endian utf-16 encoded")
	ErrNotCompressed            = Error("not compressed")
	ErrNotImplemented           = Error("not implemented")
	ErrPipelineHalted           = Error("pipeline halted")
	ErrRawReadFailed            = Error("raw read failed")
	ErrUnknownVersion           = Error("unknown version")
	ErrUnsupportedVersion       = Error("unsupported version")
	ErrUnsupportedWXMLVersion   = Error("unsupported wxml version")
)
