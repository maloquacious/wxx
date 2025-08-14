// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

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
	ErrGZipNewReaderFailed         = Error("gzip new reader failed")
	ErrInvalidGridCoordinates      = Error("invalid grid coordinates")
	ErrInvalidEncodingHeader       = Error("invalid encoding header")
	ErrInvalidGZip                 = Error("invalid gzip")
	ErrInvalidHexOrientation       = Error("invalid hex orientation")
	ErrInvalidMapMetadata          = Error("invalid <map> metadata")
	ErrInvalidTerrainMapFieldCount = Error("invalid terrain map field count")
	ErrInvalidUTF16                = Error("invalid utf-16")
	ErrInvalidUTF8                 = Error("invalid utf-8")
	ErrInvalidVersion              = Error("invalid version")
	ErrInvalidXML                  = Error("invalid xml")
	ErrInvalidXMLHeader            = Error("invalid xml header")
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
	ErrPipelineHalted              = Error("pipeline halted")
	ErrRawReadFailed               = Error("raw read failed")
	ErrUnknownVersion              = Error("unknown version")
	ErrUnknownXMLHeader            = Error("unknown xml header")
	ErrUnsupportedMapMetadata      = Error("unsupported map metadata")
	ErrUnsupportedMapRelease       = Error("unsupported map release")
	ErrUnsupportedMapSchema        = Error("unsupported map schema")
	ErrUnsupportedMapVersion       = Error("unsupported map version")
	ErrUnsupportedSchemaVersion    = Error("unsupported schema version")
	ErrUnsupportedVersion          = Error("unsupported version")
	ErrUnsupportedWXMLVersion      = Error("unsupported wxml version")
)
