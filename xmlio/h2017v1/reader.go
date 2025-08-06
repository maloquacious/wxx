// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2017v1

import (
	"github.com/maloquacious/wxx/models"
	"github.com/maloquacious/wxx/xmlio/h2017v1/rxml"
)

func Read(input []byte) (*models.Map_t, error) {
	return rxml.Read(input)
}
