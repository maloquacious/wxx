// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package hexg

// Error implements constant errors
type Error string

// Error implements the Errors interface
func (e Error) Error() string {
	return string(e)
}

const (
	ErrInvalidGridCoordinates = Error("invalid grid coordinates")
)
