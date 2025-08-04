// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a command to display the WXX package version.
package main

import (
	"fmt"
	"github.com/maloquacious/wxx"
)

func main() {
	fmt.Printf("%s\n", wxx.Version().Short())
}
