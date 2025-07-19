// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package writers

import (
	"github.com/maloquacious/wxx/models"
	"io"
)

// Write serializes the WXX data to the given io.Writer.
func Write(w io.Writer, m *models.Map) error {
	// insert the xml version and encoding header
	if _, err := w.Write([]byte("<?xml version='1.0' encoding='utf-16'?>\n")); err != nil {
		return err
	}

	// convert source from UTF-8 to UTF-16

	// compress

	// write the compressed data to w

	return ErrNotImplemented
}
