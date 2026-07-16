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

// WriteFile encodes m as the supported application version app ("1.73", "1.77",
// "2.06") and writes it to path (0644). Encoder behavior may be tuned with
// EncoderOption values (see NewEncoder).
//
// app is required, for the reason it is required on NewEncoder: writing the
// version the map happens to state would make the SOURCE file's identity the
// target, and a caller who wants that says so with m.MetaData.Version.App.Raw.
//
// The map is encoded into memory first and only written to disk once the
// encode succeeds, so a failed encode never truncates or creates a corrupt
// file at path.
func WriteFile(path string, m *wxx.Map_t, app string, opts ...EncoderOption) error {
	var buf bytes.Buffer
	if err := NewEncoder(app, opts...).Encode(&buf, m); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
