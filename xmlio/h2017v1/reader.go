// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2017v1

import (
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/h2017v1/rxml"
)

func Read(input []byte) (*wxx.Map_t, error) {
	return rxml.Read(input)
}
