// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package readers

import (
	"compress/gzip"
	"io"
)

// ReadCompressed is a helper to extract compressed data from the input.
func ReadCompressed(r io.Reader) ([]byte, error) {
	// create a new gzip reader to process the source
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func(gzr *gzip.Reader) {
		_ = gzr.Close() // ignore errors
	}(gzr)
	return io.ReadAll(gzr)
}
