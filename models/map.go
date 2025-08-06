// Copyright (c) 2024-2025 Michael D Henderson. All rights reserved.

package models

import (
	"time"

	"github.com/maloquacious/semver"
)

// Map_t is the in-memory representation of the map data.
// We have created this to work with the known versions of Worldographer XML data.
// We are assuming that this will continue to work with future versions of the application.
type Map_t struct {
	MetaData struct {
		AppVersion  semver.Version `json:"appVersion"`  // version of this application
		DataVersion semver.Version `json:"dataVersion"` // version of the data in the file
		// Worldographer defines the metadata for the WXX file
		Worldographer struct {
			Name    string    `json:"name"`    // name of input
			Created time.Time `json:"created"` // timestamp of input
			Release string    `json:"release"` // Worldographer release (eg, 2025)
			Version string    `json:"version"` // Worldographer/Hexographer version (eg 1.73)
			Schema  string    `json:"schema"`  // Worldographer XML Schema version
		} `json:"worldographer"`
		Created string `json:"created"` // timestamp of this file
	} `json:"meta-data"`

	// attributes
	Type                      string  `json:"type,omitempty"`                      // "WORLD"
	Version                   string  `json:"version,omitempty"`                   // "1.73"
	LastViewLevel             string  `json:"lastViewLevel,omitempty"`             // "WORLD"
	ContinentFactor           int     `json:"continentFactor,omitempty"`           // "-1"
	KingdomFactor             int     `json:"kingdomFactor,omitempty"`             // "-1"
	ProvinceFactor            int     `json:"provinceFactor,omitempty"`            // "-1"
	WorldToContinentHOffset   float64 `json:"worldToContinentHOffset,omitempty"`   // "0.0"
	ContinentToKingdomHOffset float64 `json:"continentToKingdomHOffset,omitempty"` // "0.0"
	KingdomToProvinceHOffset  float64 `json:"kingdomToProvinceHOffset,omitempty"`  // "0.0"
	WorldToContinentVOffset   float64 `json:"worldToContinentVOffset,omitempty"`   // "0.0"
	ContinentToKingdomVOffset float64 `json:"continentToKingdomVOffset,omitempty"` // "0.0"
	KingdomToProvinceVOffset  float64 `json:"kingdomToProvinceVOffset,omitempty"`  // "0.0"
	HexWidth                  float64 `json:"hexWidth,omitempty"`                  // "120.97791408032022"
	HexHeight                 float64 `json:"hexHeight,omitempty"`                 // "104.78814558711076"
	HexOrientation            string  `json:"hexOrientation,omitempty"`            // "COLUMNS"
	MapProjection             string  `json:"mapProjection,omitempty"`             // "FLAT"
	ShowNotes                 bool    `json:"showNotes,omitempty"`                 // "true"
	ShowGMOnly                bool    `json:"showGMOnly,omitempty"`                // "false"
	ShowGMOnlyGlow            bool    `json:"showGMOnlyGlow,omitempty"`            // "false"
	ShowFeatureLabels         bool    `json:"showFeatureLabels,omitempty"`         // "true"
	ShowGrid                  bool    `json:"showGrid,omitempty"`                  // "true"
	ShowGridNumbers           bool    `json:"showGridNumbers,omitempty"`           // "false"
	ShowShadows               bool    `json:"showShadows,omitempty"`               // "true"
	TriangleSize              int     `json:"triangleSize,omitempty"`              // "12"

	// elements
	GridAndNumbering struct {
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
	} `json:"gridAndNumbering,omitempty"`

	// TerrainMap assigns numbers to each terrain type.
	// The terrain type is used in the TileRow struct.
	TerrainMap struct {
		Data map[string]int `json:"data,omitempty"`
		List []*Terrain_t   `json:"list,omitempty"`
	} `json:"terrainMap,omitempty"`

	// MapLayer assigns a boolean "isVisible" to each layer.
	MapLayer []MapLayer_t `json:"mapLayer,omitempty"`

	Tiles Tiles_t `json:"tiles,omitempty"`

	MapKey MapKey_t `json:"mapKey,omitempty"`

	Features []*Feature_t `json:"features,omitempty"`

	Labels []*Label_t `json:"labels,omitempty"`

	Shapes []*Shape_t `json:"shapes,omitempty"`

	Notes []*Note_t `json:"notes,omitempty"`

	Informations struct {
		Informations []*Information_t `json:"informations,omitempty"`
		InnerText    string           `json:"innerText,omitempty"`
	} `json:"informations"`

	Configuration struct {
		TerrainConfig []*TerrainConfig_t `json:"terrain-config,omitempty"`
		FeatureConfig []*FeatureConfig_t `json:"feature-config,omitempty"`
		TextureConfig []*TextureConfig_t `json:"texture-config,omitempty"`
		TextConfig    struct {
			LabelStyles []*LabelStyle_t `json:"labelStyles,omitempty"`
			InnerText   string          `json:"innerText,omitempty"`
		} `json:"text-config,omitempty"`
		ShapeConfig struct {
			ShapeStyles []*ShapeStyle_t `json:"shapeStyles,omitempty"`
			InnerText   string          `json:"innerText,omitempty"`
		} `json:"shape-config"`
		InnerText string `json:"InnerText,omitempty"`
	} `json:"configuration"`
}

type Feature_t struct {
	Type              string  `json:"type,omitempty"`
	Rotate            float64 `json:"rotate,omitempty"`
	Uuid              string  `json:"uuid,omitempty"`
	MapLayer          string  `json:"mapLayer,omitempty"`
	IsFlipHorizontal  bool    `json:"isFlipHorizontal,omitempty"`
	IsFlipVertical    bool    `json:"isFlipVertical,omitempty"`
	Scale             float64 `json:"scale,omitempty"`
	ScaleHt           float64 `json:"scaleHt,omitempty"`
	Tags              string  `json:"tags,omitempty"`
	Color             *RGBA_t `json:"color,omitempty"`
	RingColor         *RGBA_t `json:"ringcolor,omitempty"`
	IsGMOnly          bool    `json:"isGMOnly,omitempty"`
	IsPlaceFreely     bool    `json:"isPlaceFreely,omitempty"`
	LabelPosition     string  `json:"labelPosition,omitempty"`
	LabelDistance     float64 `json:"labelDistance,omitempty"`
	IsWorld           bool    `json:"isWorld,omitempty"`
	IsContinent       bool    `json:"isContinent,omitempty"`
	IsKingdom         bool    `json:"isKingdom,omitempty"`
	IsProvince        bool    `json:"isProvince,omitempty"`
	IsFillHexBottom   bool    `json:"isFillHexBottom,omitempty"`
	IsHideTerrainIcon bool    `json:"isHideTerrainIcon,omitempty"`

	Location *FeatureLocation_t `json:"location,omitempty"`
	Label    *Label_t           `json:"label,omitempty"`
}

type FeatureConfig_t struct {
	InnerText string `json:"innerText,omitempty"`
}

type FeatureLocation_t struct {
	ViewLevel string  `json:"viewLevel,omitempty"`
	X         float64 `json:"x,omitempty"`
	Y         float64 `json:"y,omitempty"`
}

type Information_t struct {
	Uuid         string `json:"uuid,omitempty"`
	Type         string `json:"type,omitempty"`
	Title        string `json:"title,omitempty"`
	Rulers       string `json:"rulers,omitempty"`
	Government   string `json:"government,omitempty"`
	Cultures     string `json:"cultures,omitempty"`
	Language     string `json:"language,omitempty"`
	ReligionType string `json:"religionType,omitempty"`
	Culture      string `json:"culture,omitempty"`
	HolySymbol   string `json:"holySymbol,omitempty"`
	Domains      string `json:"domains,omitempty"`

	Details   []*InformationDetail_t `json:"details,omitempty"`
	InnerText string                 `json:"innerText,omitempty"`
}

type InformationDetail_t struct {
	Uuid         string `json:"uuid,omitempty"`
	Type         string `json:"type,omitempty"`
	Title        string `json:"title,omitempty"`
	Rulers       string `json:"rulers,omitempty"`
	Government   string `json:"government,omitempty"`
	Cultures     string `json:"cultures,omitempty"`
	Language     string `json:"language,omitempty"`
	ReligionType string `json:"religionType,omitempty"`
	Culture      string `json:"culture,omitempty"`
	HolySymbol   string `json:"holySymbol,omitempty"`
	Domains      string `json:"domains,omitempty"`

	InnerText string `json:"innerText,omitempty"`
}

type Label_t struct {
	MapLayer        string  `json:"mapLayer,omitempty"`
	Style           string  `json:"style,omitempty"`
	FontFace        string  `json:"fontFace,omitempty"`
	Color           *RGBA_t `json:"color,omitempty"`
	OutlineColor    *RGBA_t `json:"outlineColor,omitempty"`
	OutlineSize     float64 `json:"outlineSize,omitempty"`
	Rotate          float64 `json:"rotate,omitempty"`
	IsBold          bool    `json:"isBold,omitempty"`
	IsItalic        bool    `json:"isItalic,omitempty"`
	IsWorld         bool    `json:"isWorld,omitempty"`
	IsContinent     bool    `json:"isContinent,omitempty"`
	IsKingdom       bool    `json:"isKingdom,omitempty"`
	IsProvince      bool    `json:"isProvince,omitempty"`
	IsGMOnly        bool    `json:"isGMOnly,omitempty"`
	Tags            string  `json:"tags,omitempty"`
	BackgroundColor *RGBA_t `json:"backgroundColor,omitempty"`

	Location  *LabelLocation_t `json:"location,omitempty"`
	InnerText string           `json:"innerText,omitempty"`
}

type LabelLocation_t struct {
	ViewLevel string  `json:"viewLevel,omitempty"`
	X         float64 `json:"x,omitempty"`
	Y         float64 `json:"y,omitempty"`
	Scale     float64 `json:"scale,omitempty"`
}

type LabelStyle_t struct {
	Name            string  `json:"name,omitempty"`
	FontFace        string  `json:"fontFace,omitempty"`
	Scale           float64 `json:"scale,omitempty"`
	IsBold          bool    `json:"isBold,omitempty"`
	IsItalic        bool    `json:"isItalic,omitempty"`
	Color           *RGBA_t `json:"color,omitempty"`
	BackgroundColor *RGBA_t `json:"backgroundColor,omitempty"`
	OutlineSize     float64 `json:"outlineSize,omitempty"`
	OutlineColor    *RGBA_t `json:"outlineColor,omitempty"`
}

type MapKey_t struct {
	// attributes
	PositionX         float64 `json:"positionx,omitempty"`
	PositionY         float64 `json:"positiony,omitempty"`
	Viewlevel         string  `json:"viewlevel,omitempty"` // "null", "WORLD"
	Height            float64 `json:"height,omitempty"`
	BackgroundColor   *RGBA_t `json:"backgroundcolor,omitempty"`
	BackgroundOpacity float64 `json:"backgroundopacity,omitempty"`
	TitleText         string  `json:"titleText,omitempty"`
	TitleFontFace     string  `json:"titleFontFace,omitempty"`
	TitleFontColor    *RGBA_t `json:"titleFontColor,omitempty"`
	TitleFontBold     bool    `json:"titleFontBold,omitempty"`
	TitleFontItalic   bool    `json:"titleFontItalic,omitempty"`
	TitleScale        float64 `json:"titleScale,omitempty"`
	ScaleText         string  `json:"scaleText,omitempty"`
	ScaleFontFace     string  `json:"scaleFontFace,omitempty"`
	ScaleFontColor    *RGBA_t `json:"scaleFontColor,omitempty"`
	ScaleFontBold     bool    `json:"scaleFontBold,omitempty"`
	ScaleFontItalic   bool    `json:"scaleFontItalic,omitempty"`
	ScaleScale        float64 `json:"scaleScale,omitempty"`
	EntryFontFace     string  `json:"entryFontFace,omitempty"`
	EntryFontColor    *RGBA_t `json:"entryFontColor,omitempty"`
	EntryFontBold     bool    `json:"entryFontBold,omitempty"`
	EntryFontItalic   bool    `json:"entryFontItalic,omitempty"`
	EntryScale        float64 `json:"entryScale,omitempty"`
}

type MapLayer_t struct {
	Name      string `json:"name"`
	IsVisible bool   `json:"isVisible"`
}

type Note_t struct {
	InnerText string `json:"innerText,omitempty"`
}

type Point_t struct {
	Type string  `json:"type,omitempty"`
	X    float64 `json:"x,omitempty"`
	Y    float64 `json:"y,omitempty"`
}

type RGBA_t struct {
	R float64
	G float64
	B float64
	A float64
}

type Shape_t struct {
	// attributes0
	BbHeight              float64 `json:"bbHeight,omitempty"`
	BbIterations          int     `json:"bbIterations,omitempty"`
	BbWidth               float64 `json:"bbWidth,omitempty"`
	CreationType          string  `json:"creationType,omitempty"`
	CurrentShapeViewLevel string  `json:"currentShapeViewLevel,omitempty"`
	DsColor               string  `json:"dsColor,omitempty"`
	DsOffsetX             float64 `json:"dsOffsetX,omitempty"`
	DsOffsetY             float64 `json:"dsOffsetY,omitempty"`
	DsRadius              float64 `json:"dsRadius,omitempty"`
	DsSpread              float64 `json:"dsSpread,omitempty"`
	FillRule              string  `json:"fillRule,omitempty"`
	FillTexture           string  `json:"fillTexture,omitempty"`
	HighestViewLevel      string  `json:"highestViewLevel,omitempty"`
	InsChoke              float64 `json:"insChoke,omitempty"`
	InsColor              string  `json:"insColor,omitempty"`
	InsOffsetX            float64 `json:"insOffsetX,omitempty"`
	InsOffsetY            float64 `json:"insOffsetY,omitempty"`
	InsRadius             float64 `json:"insRadius,omitempty"`
	IsBoxBlur             bool    `json:"isBoxBlur,omitempty"`
	IsContinent           bool    `json:"isContinent,omitempty"`
	IsCurve               bool    `json:"isCurve,omitempty"`
	IsDropShadow          bool    `json:"isDropShadow,omitempty"`
	IsGMOnly              bool    `json:"isGMOnly,omitempty"`
	IsInnerShadow         bool    `json:"isInnerShadow,omitempty"`
	IsKingdom             bool    `json:"isKingdom,omitempty"`
	IsMatchTileBorders    bool    `json:"isMatchTileBorders,omitempty"`
	IsProvince            bool    `json:"isProvince,omitempty"`
	IsSnapVertices        bool    `json:"isSnapVertices,omitempty"`
	IsWorld               bool    `json:"isWorld,omitempty"`
	LineCap               string  `json:"lineCap,omitempty"`
	LineJoin              string  `json:"lineJoin,omitempty"`
	MapLayer              string  `json:"mapLayer,omitempty"`
	Opacity               float64 `json:"opacity,omitempty"`
	StrokeColor           string  `json:"strokeColor,omitempty"`
	StrokeTexture         string  `json:"strokeTexture,omitempty"`
	StrokeType            string  `json:"strokeType,omitempty"`
	StrokeWidth           float64 `json:"strokeWidth,omitempty"`
	Tags                  string  `json:"tags,omitempty"`
	Type                  string  `json:"type,omitempty"`

	Points []*Point_t `json:"points,omitempty"`
}

type ShapeConfig_t struct {
	ShapeStyles []*ShapeStyle_t `json:"shapeStyles,omitempty"`
	InnerText   string          `json:"innerText,omitempty"`
}

type ShapeStyle_t struct {
	Name          string  `json:"name,omitempty"`
	StrokeType    string  `json:"strokeType,omitempty"`
	IsFractal     bool    `json:"isFractal,omitempty"`
	StrokeWidth   float64 `json:"strokeWidth,omitempty"`
	Opacity       float64 `json:"opacity,omitempty"`
	SnapVertices  bool    `json:"snapVertices,omitempty"`
	Tags          string  `json:"tags,omitempty"`
	DropShadow    bool    `json:"dropShadow,omitempty"`
	InnerShadow   bool    `json:"innerShadow,omitempty"`
	BoxBlur       bool    `json:"boxBlur,omitempty"`
	DsSpread      float64 `json:"dsSpread,omitempty"`
	DsRadius      float64 `json:"dsRadius,omitempty"`
	DsOffsetX     float64 `json:"dsOffsetX,omitempty"`
	DsOffsetY     float64 `json:"dsOffsetY,omitempty"`
	InsChoke      float64 `json:"insChoke,omitempty"`
	InsRadius     float64 `json:"insRadius,omitempty"`
	InsOffsetX    float64 `json:"insOffsetX,omitempty"`
	InsOffsetY    float64 `json:"insOffsetY,omitempty"`
	BbWidth       float64 `json:"bbWidth,omitempty"`
	BbHeight      float64 `json:"bbHeight,omitempty"`
	BbIterations  int     `json:"bbIterations,omitempty"`
	FillTexture   string  `json:"fillTexture,omitempty"`
	StrokeTexture string  `json:"strokeTexture,omitempty"`
	StrokePaint   *RGBA_t `json:"strokePaint,omitempty"`
	FillPaint     *RGBA_t `json:"fillPaint,omitempty"`
	DsColor       *RGBA_t `json:"dscolor,omitempty"`
	InsColor      *RGBA_t `json:"insColor,omitempty"`
}

type Terrain_t struct {
	Index int    `json:"index"`
	Label string `json:"label"`
}

type TerrainConfig_t struct {
	InnerText string `json:"innerText,omitempty"`
}

type TextConfig_t struct {
	LabelStyles []*LabelStyle_t `json:"labelStyles,omitempty"`
	InnerText   string          `json:"innerText,omitempty"`
}

type TextureConfig_t struct {
	InnerText string `json:"innerText,omitempty"`
}

type Tile_t struct {
	Row       int
	Column    int
	Terrain   int // lookup into TerrainMap
	Elevation float64
	IsIcy     bool
	IsGMOnly  bool
	Resources struct {
		Animal int
		Brick  int
		Crops  int
		Gems   int
		Lumber int
		Metals int
		Rock   int
	}
	CustomBackgroundColor *RGBA_t
}

type Tiles_t struct {
	ViewLevel string `json:"viewLevel,omitempty"`
	TilesWide int    `json:"tilesWide,omitempty"` // number of columns of tiles
	TilesHigh int    `json:"tilesHigh,omitempty"` // number of rows of tiles

	TileRows [][]*Tile_t `json:"tilerow,omitempty"`
}
