// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2017v1

import (
	"fmt"
	"github.com/maloquacious/wxx/models"
	"strconv"
	"strings"
)

func decodeRgba(s string) (rgba *models.RGBA, err error) {
	if s == "" || s == "null" || s == "0.0,0.0,0.0,1.0" {
		return nil, nil
	}
	rgba = &models.RGBA{}
	values := strings.Split(s, ",")
	if len(values) != 4 {
		return rgba, fmt.Errorf("invalid value")
	} else if rgba.R, err = strconv.ParseFloat(values[0], 64); err != nil {
		return rgba, err
	} else if rgba.G, err = strconv.ParseFloat(values[1], 64); err != nil {
		return rgba, err
	} else if rgba.B, err = strconv.ParseFloat(values[2], 64); err != nil {
		return rgba, err
	} else if rgba.A, err = strconv.ParseFloat(values[3], 64); err != nil {
		return rgba, err
	}
	if rgba.R == 0 && rgba.G == 0 && rgba.B == 0 && rgba.A == 1 {
		return nil, nil
	}
	return rgba, nil
}

func decodeZeroableRgba(s string) (rgba *models.RGBA, err error) {
	if s == "" || s == "null" {
		return nil, nil
	}
	rgba = &models.RGBA{}
	values := strings.Split(s, ",")
	if len(values) != 4 {
		return rgba, fmt.Errorf("invalid value")
	} else if rgba.R, err = strconv.ParseFloat(values[0], 64); err != nil {
		return rgba, err
	} else if rgba.G, err = strconv.ParseFloat(values[1], 64); err != nil {
		return rgba, err
	} else if rgba.B, err = strconv.ParseFloat(values[2], 64); err != nil {
		return rgba, err
	} else if rgba.A, err = strconv.ParseFloat(values[3], 64); err != nil {
		return rgba, err
	}
	return rgba, nil
}
