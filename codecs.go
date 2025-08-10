// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

import (
	"io"
)

type Decoder interface {
	Decode(io.Reader) (*Map_t, error)
}

type Encoder interface {
	Encode(io.Writer, *Map_t) error
}
