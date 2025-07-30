// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package models

// Error implements constant errors
type Error string

// Error implements the Errors interface
func (e Error) Error() string {
	return string(e)
}

const (
	ErrFSError                     = Error("file-system")
	ErrGUnZipFailed                = Error("gunzip failed")
	ErrGZipFailed                  = Error("gzip failed")
	ErrInvalidEncodingHeader       = Error("invalid encoding header")
	ErrInvalidGZip                 = Error("invalid gzip")
	ErrInvalidHexOrientation       = Error("invalid hex orientation")
	ErrInvalidTerrainMapFieldCount = Error("invalid terrain map field count")
	ErrInvalidMapMetadata          = Error("invalid <map> metadata")
	ErrInvalidUTF16                = Error("invalid utf-16")
	ErrInvalidVersion              = Error("invalid version")
	ErrInvalidXML                  = Error("invalid xml")
	ErrMapNotClosed                = Error("<map> not closed")
	ErrMissingBOM                  = Error("missing bom")
	ErrMissingFinalByte            = Error("missing final byte")
	ErrMissingMapElement           = Error("missing map element")
	ErrMissingVersion              = Error("missing version")
	ErrMissingWxxExtension         = Error("missing .wxx extension")
	ErrMissingXMLHeader            = Error("missing xml header")
	ErrNotBigEndianUTF16Encoded    = Error("not big-endian utf-16 encoded")
	ErrNotCompressed               = Error("not compressed")
	ErrNotExists                   = Error("not exists")
	ErrNotFile                     = Error("not a file")
	ErrNotImplemented              = Error("not implemented")
	ErrUnknownVersion              = Error("unknown version")
	ErrUnsupportedMapMetadata      = Error("unsupported map metadata")
	ErrUnsupportedMapRelease       = Error("unsupported map release")
	ErrUnsupportedMapSchema        = Error("unsupported map schema")
	ErrUnsupportedMapVersion       = Error("unsupported map version")
	ErrUnsupportedSchemaVersion    = Error("unsupported schema version")
	ErrUnsupportedVersion          = Error("unsupported version")
	ErrUnsupportedWXMLVersion      = Error("unsupported wxml version")
)
