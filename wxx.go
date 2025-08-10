// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package wxx defines the major data types for decoding, manipulating,
// and encoding Worldographer data files. We support two versions,
// Worldographer and Worldographer 2025. The original Worldographer is
// sometimes called "Worldographer classic." We call them H2017 and H2025
// to reduce confusion.
package wxx

import "io"

type Decoder interface {
	Decode(io.Reader) (*Map_t, error)
}

type Encoder interface {
	Encode(io.Writer, *Map_t) error
}
