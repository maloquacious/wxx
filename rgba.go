// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
	"strconv"
	"strings"
)

// rgbaToXmlAttr converts an RGBA struct into an XML attribute string.
// RGBA struct contains four fields, each representing Red, Green, Blue and Alpha respectively.
// Each field is a float and the rgbaToXmlAttr function formats them as a string, separated by commas.
// If the provided rgba pointer is nil, it defaults to "0.0,0.0,0.0,1.0".
//
// It uses the FToXF function to format the float values into an XML-friendly format.
//
// Parameters:
// - rgba: a pointer to a RGBA struct. Can be nil.
//
// Returns:
// - A XML attribute string representing the rgba. If rgba is nil, returns "0.0,0.0,0.0,1.0"
func rgbaToXmlAttr(rgba *RGBA) string {
	if rgba == nil {
		return "0.0,0.0,0.0,1.0"
	}
	return fmt.Sprintf("%s,%s,%s,%s",
		FToXF(rgba.R),
		FToXF(rgba.G),
		FToXF(rgba.B),
		FToXF(rgba.A))
}

func rgbaToNullableXmlAttr(rgba *RGBA) string {
	s := rgbaToXmlAttr(rgba)
	if s == "0.0,0.0,0.0,1.0" {
		s = "null"
	}
	return s
}

func decodeRgba(s string) (rgba *RGBA, err error) {
	if s == "" || s == "null" || s == "0.0,0.0,0.0,1.0" {
		return nil, nil
	}
	rgba = &RGBA{}
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

func decodeZeroableRgba(s string) (rgba *RGBA, err error) {
	if s == "" || s == "null" {
		return nil, nil
	}
	rgba = &RGBA{}
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
