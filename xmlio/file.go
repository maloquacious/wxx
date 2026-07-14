// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio

import (
	"bytes"
	"fmt"
	"os"

	"github.com/maloquacious/wxx"
)

// ReadFile reads and decodes a Worldographer .wxx file from path.
// Decoder behavior may be tuned with DecoderOption values (see NewDecoder).
func ReadFile(path string, opts ...DecoderOption) (*wxx.Map_t, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return NewDecoder(opts...).Decode(f)
}

// WriteFile encodes m and writes it to path (0644). By default it targets
// m.MetaData.DataVersion; pass WithTargetVersion(...) (or other EncoderOption)
// to override.
//
// The map is encoded into memory first and only written to disk once the
// encode succeeds, so a failed encode never truncates or creates a corrupt
// file at path.
func WriteFile(path string, m *wxx.Map_t, opts ...EncoderOption) error {
	var buf bytes.Buffer
	if err := NewEncoder(opts...).Encode(&buf, m); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
