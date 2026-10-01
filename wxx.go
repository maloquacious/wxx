// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package wxx defines the major data types for decoding, manipulating,
// and encoding Worldographer data files. We support Worldographer 2025.
// Files from the original Worldographer, sometimes called "Worldographer
// classic," are refused (ErrClassicMap); Worldographer 2025 converts them.
package wxx

import "io"

type Decoder interface {
	Decode(io.Reader) (*Map_t, error)
}

type Encoder interface {
	Encode(io.Writer, *Map_t) error
}
