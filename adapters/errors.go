// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

// Error implements constant errors
type Error string

// Error implements the Errors interface
func (e Error) Error() string {
	return string(e)
}

const (
	ErrUnsupportedWXMLVersion = Error("unsupported wxml version")
)
