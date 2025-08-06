// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxml

// Schema_t defines the structure for writing a Worldographer file with the H2017v1 XML Schema
type Schema_t struct {
	// attributes
	Type                      string
	Version                   string
	LastViewLevel             string
	ContinentFactor           int
	KingdomFactor             int
	ProvinceFactor            int
	WorldToContinentHOffset   float64
	ContinentToKingdomHOffset float64
	KingdomToProvinceHOffset  float64
	WorldToContinentVOffset   float64
	ContinentToKingdomVOffset float64
	KingdomToProvinceVOffset  float64
	HexWidth                  float64
	HexHeight                 float64
	HexOrientation            string
	MapProjection             string
	ShowNotes                 bool
	ShowGMOnly                bool
	ShowGMOnlyGlow            bool
	ShowFeatureLabels         bool
	ShowGrid                  bool
	ShowGridNumbers           bool
	ShowShadows               bool
	TriangleSize              int

	// elements
	GridAndNumbering *GridAndNumbering_t
	TerrainMap       []*Terrain_t
}

type GridAndNumbering_t struct {
	Color0                      string  `json:"color0,omitempty"`                      // "0x00000040"
	Color1                      string  `json:"color1,omitempty"`                      // "0x00000040"
	Color2                      string  `json:"color2,omitempty"`                      // "0x00000040"
	Color3                      string  `json:"color3,omitempty"`                      // "0x00000040"
	Color4                      string  `json:"color4,omitempty"`                      // "0x00000040"
	Width0                      float64 `json:"width0,omitempty"`                      // "1.0"
	Width1                      float64 `json:"width1,omitempty"`                      // "2.0"
	Width2                      float64 `json:"width2,omitempty"`                      // "3.0"
	Width3                      float64 `json:"width3,omitempty"`                      // "4.0"
	Width4                      float64 `json:"width4,omitempty"`                      // "1.0"
	GridOffsetContinentKingdomX float64 `json:"gridOffsetContinentKingdomX,omitempty"` // "0.0"
	GridOffsetContinentKingdomY float64 `json:"gridOffsetContinentKingdomY,omitempty"` // "0.0"
	GridOffsetWorldContinentX   float64 `json:"gridOffsetWorldContinentX,omitempty"`   // "0.0"
	GridOffsetWorldContinentY   float64 `json:"gridOffsetWorldContinentY,omitempty"`   // "0.0"
	GridOffsetWorldKingdomX     float64 `json:"gridOffsetWorldKingdomX,omitempty"`     // "0.0"
	GridOffsetWorldKingdomY     float64 `json:"gridOffsetWorldKingdomY,omitempty"`     // "0.0"
	GridSquare                  int     `json:"gridSquare,omitempty"`                  // "0"
	GridSquareHeight            float64 `json:"gridSquareHeight,omitempty"`            // "-1.0"
	GridSquareWidth             float64 `json:"gridSquareWidth,omitempty"`             // "-1.0"
	GridOffsetX                 float64 `json:"gridOffsetX,omitempty"`                 // "0.0"
	GridOffsetY                 float64 `json:"gridOffsetY,omitempty"`                 // "0.0"
	NumberFont                  string  `json:"numberFont,omitempty"`                  // "Arial"
	NumberColor                 string  `json:"numberColor,omitempty"`                 // "0x000000ff"
	NumberSize                  int     `json:"numberSize,omitempty"`                  // "20"
	NumberStyle                 string  `json:"numberStyle,omitempty"`                 // "PLAIN"
	NumberFirstCol              int     `json:"numberFirstCol,omitempty"`              // "0"
	NumberFirstRow              int     `json:"numberFirstRow,omitempty"`              // "0"
	NumberOrder                 string  `json:"numberOrder,omitempty"`                 // "COL_ROW"
	NumberPosition              string  `json:"numberPosition,omitempty"`              // "BOTTOM"
	NumberPrePad                string  `json:"numberPrePad,omitempty"`                // "DOUBLE_ZERO"
	NumberSeparator             string  `json:"numberSeparator,omitempty"`             // "."
}

type Terrain_t struct {
	Index int
	Label string
}
