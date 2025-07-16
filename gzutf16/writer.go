// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package gzutf16

import (
	"fmt"
	"os"
)

// WriteFile writes data to the named file in Worldographer format, creating it if necessary.
// The data is converted from UTF-8 to UTF-16 big-endian and compressed with gzip before writing.
// If the file does not exist, WriteFile creates it with permissions perm (before umask);
// otherwise WriteFile truncates it before writing, without changing permissions.
// Since WriteFile requires multiple system calls to complete, a failure mid-operation
// can leave the file in a partially written state.
func WriteFile(name string, data []byte, perm os.FileMode) error {
	if name == "" {
		return fmt.Errorf("empty filename")
	}
	
	// Use Buffer to convert the data to Worldographer format
	var buf Buffer
	
	// Write the UTF-8 data to the buffer
	_, err := buf.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to buffer: %w", err)
	}
	
	// Convert to gzip-compressed UTF-16 big-endian format
	convertedData, err := buf.Bytes()
	if err != nil {
		return fmt.Errorf("failed to convert data to Worldographer format: %w", err)
	}
	
	// Write the converted data to the file
	err = os.WriteFile(name, convertedData, perm)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", name, err)
	}
	
	return nil
}
