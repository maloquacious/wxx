// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

import (
	"bytes"
	"compress/gzip"
)

func UTF16ToGZip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	} else if err = gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
