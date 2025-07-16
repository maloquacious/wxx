// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

import (
	"compress/gzip"
	"io"
)

func GZipToUTF16(r io.Reader) ([]byte, error) {
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
