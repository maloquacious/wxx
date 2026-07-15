// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2025v1

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// XMLSchema defines the structure for reading a Worldographer file with the H2025v1 XML Schema
type XMLSchema struct {
	XMLName xml.Name `xml:"map"`

	// attributes
	Type                      string  `xml:"type,attr"`
	Release                   string  `xml:"release,attr"` // 2025
	Version                   string  `xml:"version,attr"`
	Schema                    string  `xml:"schema,attr"`
	LastViewLevel             string  `xml:"lastViewLevel,attr"`
	ContinentFactor           int     `xml:"continentFactor,attr"`
	KingdomFactor             int     `xml:"kingdomFactor,attr"`
	ProvinceFactor            int     `xml:"provinceFactor,attr"`
	WorldToContinentHOffset   float64 `xml:"worldToContinentHOffset,attr"`
	ContinentToKingdomHOffset float64 `xml:"continentToKingdomHOffset,attr"`
	KingdomToProvinceHOffset  float64 `xml:"kingdomToProvinceHOffset,attr"`
	WorldToContinentVOffset   float64 `xml:"worldToContinentVOffset,attr"`
	ContinentToKingdomVOffset float64 `xml:"continentToKingdomVOffset,attr"`
	KingdomToProvinceVOffset  float64 `xml:"kingdomToProvinceVOffset,attr"`
	HexWidth                  float64 `xml:"hexWidth,attr"`
	HexHeight                 float64 `xml:"hexHeight,attr"`
	HexOrientation            string  `xml:"hexOrientation,attr"` // COLUMNS or ROWS
	MapProjection             string  `xml:"mapProjection,attr"`  // FLAT or ICOSAHEDRAL
	ShowNotes                 bool    `xml:"showNotes,attr"`
	ShowGMOnly                bool    `xml:"showGMOnly,attr"`
	ShowGMOnlyGlow            bool    `xml:"showGMOnlyGlow,attr"`
	ShowFeatureLabels         bool    `xml:"showFeatureLabels,attr"`
	ShowGrid                  bool    `xml:"showGrid,attr"`
	ShowGridNumbers           bool    `xml:"showGridNumbers,attr"`
	ShowShadows               bool    `xml:"showShadows,attr"`
	TriangleSize              int     `xml:"triangleSize,attr"`
	HScrollbarPos             float64 `xml:"hScrollbarPos,attr"`
	VScrollbarPos             float64 `xml:"vScrollbarPos,attr"`

	// elements
	GridAndNumbering GridAndNumbering `xml:"gridandnumbering"`
	// BlurTerrainBG / ExtraTerrain are optional top-level elements; a pointer
	// leaves them nil when the element is absent from the file.
	BlurTerrainBG *BlurTerrainBG_t `xml:"blurTerrainBG"`
	TerrainMap    TerrainMap_t     `xml:"terrainmap"`
	MapLayers     []MapLayer_t     `xml:"maplayer"`
	Tiles         Tiles_t          `xml:"tiles"`
	MapKey        MapKey_t         `xml:"mapkey"`
	Features      Features         `xml:"features"`
	ExtraTerrain  *ExtraTerrain_t  `xml:"extraTerrain"`
	Labels        Labels_t         `xml:"labels"`
	Shapes        Shapes_t         `xml:"shapes"`
	Notes         Notes_t          `xml:"notes"`
	Informations  Informations_t   `xml:"informations"`
	Configuration Configuration_t  `xml:"configuration"`
}

// BlurTerrainBG_t is the on-disk <blurTerrainBG> element (W2025-native).
type BlurTerrainBG_t struct {
	Blur        bool    `xml:"blur,attr"`
	TopBleed    float64 `xml:"topBleed,attr"`
	BottomBleed float64 `xml:"bottomBleed,attr"`
	Randomness  float64 `xml:"randomness,attr"`
	BlurStart   float64 `xml:"blurStart,attr"`
	BlurEnd     float64 `xml:"blurEnd,attr"`
}

// ExtraTerrain_t is the on-disk <extraTerrain> element (W2025-native). It is an
// empty container in observed samples; InnerXML preserves any inner content.
type ExtraTerrain_t struct {
	InnerXML string `xml:",innerxml"`
}

type Configuration_t struct {
	// elements
	TerrainConfig []TerrainConfig_t `xml:"terrain-config"`
	FeatureConfig []FeatureConfig   `xml:"feature-config"`
	TextureConfig []TextureConfig_t `xml:"texture-config"`
	TextConfig    []TextConfig_t    `xml:"text-config"`
	ShapeConfig   []ShapeConfig_t   `xml:"shape-config"`
}

type Feature struct {
	// attributes
	Type              string  `xml:"type,attr"`
	Rotate            float64 `xml:"rotate,attr"`
	Uuid              string  `xml:"uuid,attr"`
	MapLayer          string  `xml:"mapLayer,attr"`
	IsFlipHorizontal  bool    `xml:"isFlipHorizontal,attr"`
	IsFlipVertical    bool    `xml:"isFlipVertical,attr"`
	Scale             float64 `xml:"scale,attr"`
	ScaleHt           float64 `xml:"scaleHt,attr"`
	Tags              string  `xml:"tags,attr"`
	Color             string  `xml:"color,attr"`
	RingColor         string  `xml:"ringcolor,attr"`
	IsGMOnly          bool    `xml:"isGMOnly,attr"`
	IsPlaceFreely     bool    `xml:"isPlaceFreely,attr"`
	LabelPosition     string  `xml:"labelPosition,attr"`
	LabelDistance     int     `xml:"labelDistance,attr"`
	IsWorld           bool    `xml:"isWorld,attr"`
	IsContinent       bool    `xml:"isContinent,attr"`
	IsKingdom         bool    `xml:"isKingdom,attr"`
	IsProvince        bool    `xml:"isProvince,attr"`
	IsFillHexBottom   bool    `xml:"isFillHexBottom,attr"`
	IsHideTerrainIcon bool    `xml:"isHideTerrainIcon,attr"`

	// elements
	Location struct {
		// attributes
		ViewLevel string  `xml:"viewLevel,attr"`
		X         float64 `xml:"x,attr"`
		Y         float64 `xml:"y,attr"`
	} `xml:"location"`
	Label     *Label_t `xml:"label,omitempty"`
	InnerText string   `xml:",chardata"`
}

type FeatureConfig struct {
	// elements
	InnerText string `xml:",chardata"`
}

type Features struct {
	// elements
	Features []Feature `xml:"feature"`
}

type GridAndNumbering struct {
	// attributes
	Color0                      string  `xml:"color0,attr"`
	Color1                      string  `xml:"color1,attr"`
	Color2                      string  `xml:"color2,attr"`
	Color3                      string  `xml:"color3,attr"`
	Color4                      string  `xml:"color4,attr"`
	Width0                      float64 `xml:"width0,attr"`
	Width1                      float64 `xml:"width1,attr"`
	Width2                      float64 `xml:"width2,attr"`
	Width3                      float64 `xml:"width3,attr"`
	Width4                      float64 `xml:"width4,attr"`
	GridOffsetContinentKingdomX float64 `xml:"gridOffsetContinentKingdomX,attr"`
	GridOffsetContinentKingdomY float64 `xml:"gridOffsetContinentKingdomY,attr"`
	GridOffsetWorldContinentX   float64 `xml:"gridOffsetWorldContinentX,attr"`
	GridOffsetWorldContinentY   float64 `xml:"gridOffsetWorldContinentY,attr"`
	GridOffsetWorldKingdomX     float64 `xml:"gridOffsetWorldKingdomX,attr"`
	GridOffsetWorldKingdomY     float64 `xml:"gridOffsetWorldKingdomY,attr"`
	GridSquare                  int     `xml:"gridSquare,attr"`
	GridSquareHeight            float64 `xml:"gridSquareHeight,attr"`
	GridSquareWidth             float64 `xml:"gridSquareWidth,attr"`
	GridOffsetX                 float64 `xml:"gridOffsetX,attr"`
	GridOffsetY                 float64 `xml:"gridOffsetY,attr"`
	NumberFont                  string  `xml:"numberFont,attr"`
	NumberColor                 string  `xml:"numberColor,attr"`
	NumberSize                  int     `xml:"numberSize,attr"`
	NumberStyle                 string  `xml:"numberStyle,attr"`
	NumberFirstCol              int     `xml:"numberFirstCol,attr"`
	NumberFirstRow              int     `xml:"numberFirstRow,attr"`
	NumberOrder                 string  `xml:"numberOrder,attr"`
	NumberPosition              string  `xml:"numberPosition,attr"`
	NumberPrePad                string  `xml:"numberPrePad,attr"`
	NumberSeparator             string  `xml:"numberSeparator,attr"`
}

type Information_t struct {
	// attributes
	Uuid       string `xml:"uuid,attr"`
	Type       string `xml:"type,attr"`
	Title      string `xml:"title,attr"`
	Rulers     string `xml:"rulers,attr"`
	Government string `xml:"government,attr"`
	Cultures   string `xml:"cultures,attr"`

	Language string `xml:"language,attr"`

	ReligionType string `xml:"religionType,attr"`
	Culture      string `xml:"culture,attr"`
	HolySymbol   string `xml:"holySymbol,attr"`
	Domains      string `xml:"domains,attr"`

	// elements
	Details   []Information_t `xml:"information"`
	InnerText string          `xml:",chardata"`
}

type Informations_t struct {
	// elements
	Informations []Information_t `xml:"information"`
	InnerText    string          `xml:",chardata"`
}

type Label_t struct {
	// attributes
	MapLayer        string  `xml:"mapLayer,attr"`
	Style           string  `xml:"style,attr"`
	FontFace        string  `xml:"fontFace,attr"`
	Color           string  `xml:"color,attr"`
	OutlineColor    string  `xml:"outlineColor,attr"`
	OutlineSize     float64 `xml:"outlineSize,attr"`
	Rotate          float64 `xml:"rotate,attr"`
	IsBold          bool    `xml:"isBold,attr"`
	IsItalic        bool    `xml:"isItalic,attr"`
	IsWorld         bool    `xml:"isWorld,attr"`
	IsContinent     bool    `xml:"isContinent,attr"`
	IsKingdom       bool    `xml:"isKingdom,attr"`
	IsProvince      bool    `xml:"isProvince,attr"`
	IsGMOnly        bool    `xml:"isGMOnly,attr"`
	Tags            string  `xml:"tags,attr"`
	BackgroundColor string  `xml:"backgroundColor,attr"`

	// elements
	Location  Location_t `xml:"location"`
	InnerText string     `xml:",chardata"`
}

type Labels_t struct {
	// elements
	Labels []Label_t `xml:"label"`
}

type LabelStyle_t struct {
	// attributes
	Name            string  `xml:"name,attr"`
	FontFace        string  `xml:"fontFace,attr"`
	Scale           float64 `xml:"scale,attr"`
	IsBold          bool    `xml:"isBold,attr"`
	IsItalic        bool    `xml:"isItalic,attr"`
	Color           string  `xml:"color,attr"`
	BackgroundColor string  `xml:"backgroundColor,attr"`
	OutlineSize     float64 `xml:"outlineSize,attr"`
	OutlineColor    string  `xml:"outlineColor,attr"`

	// W2025 drop-shadow attributes. dropShadowColor can be the literal "null".
	DropShadowColor  string  `xml:"dropShadowColor,attr"`
	DropShadowRadius float64 `xml:"dropShadowRadius,attr"`
	DropShadowSpread float64 `xml:"dropShadowSpread,attr"`
}

type Location_t struct {
	// attributes
	ViewLevel string  `xml:"viewLevel,attr"`
	X         float64 `xml:"x,attr"`
	Y         float64 `xml:"y,attr"`
	Scale     float64 `xml:"scale,attr"`
}

type MapKey_t struct {
	// attributes
	PositionX         float64 `xml:"positionx,attr"`
	PositionY         float64 `xml:"positiony,attr"`
	Viewlevel         string  `xml:"viewlevel,attr"`
	Height            float64 `xml:"height,attr"`
	BackgroundColor   string  `xml:"backgroundcolor,attr"`
	BackgroundOpacity float64 `xml:"backgroundopacity,attr"`
	TitleText         string  `xml:"titleText,attr"`
	TitleFontFace     string  `xml:"titleFontFace,attr"`
	TitleFontColor    string  `xml:"titleFontColor,attr"`
	TitleFontBold     bool    `xml:"titleFontBold,attr"`
	TitleFontItalic   bool    `xml:"titleFontItalic,attr"`
	TitleScale        float64 `xml:"titleScale,attr"`
	ScaleText         string  `xml:"scaleText,attr"`
	ScaleFontFace     string  `xml:"scaleFontFace,attr"`
	ScaleFontColor    string  `xml:"scaleFontColor,attr"`
	ScaleFontBold     bool    `xml:"scaleFontBold,attr"`
	ScaleFontItalic   bool    `xml:"scaleFontItalic,attr"`
	ScaleScale        float64 `xml:"scaleScale,attr"`
	EntryFontFace     string  `xml:"entryFontFace,attr"`
	EntryFontColor    string  `xml:"entryFontColor,attr"`
	EntryFontBold     bool    `xml:"entryFontBold,attr"`
	EntryFontItalic   bool    `xml:"entryFontItalic,attr"`
	EntryScale        float64 `xml:"entryScale,attr"`
}

type MapLayer_t struct {
	// attributes
	Name      string  `xml:"name,attr"`
	IsVisible bool    `xml:"isVisible,attr"`
	Opacity   float64 `xml:"opacity,attr"`
}

type Note_t struct {
	// attributes
	Key       string  `xml:"key,attr"`
	ViewLevel string  `xml:"viewLevel,attr"`
	X         float64 `xml:"x,attr"`
	Y         float64 `xml:"y,attr"`
	Filename  string  `xml:"filename,attr"`
	Parent    string  `xml:"parent,attr"`
	Color     string  `xml:"color,attr"`
	Title     string  `xml:"title,attr"`
	IsGMOnly  bool    `xml:"isGMOnly,attr"`

	// elements
	NoteText string `xml:"notetext"`
}

type Notes_t struct {
	Notes []Note_t `xml:"note"`
}

type Point_t struct {
	// attributes
	Type string  `xml:"type,attr"`
	X    float64 `xml:"x,attr"`
	Y    float64 `xml:"y,attr"`
}

type Shape_t struct {
	// attributes
	BbHeight              float64 `xml:"bbHeight,attr"`
	BbIterations          int     `xml:"bbIterations,attr"`
	BbWidth               float64 `xml:"bbWidth,attr"`
	CreationType          string  `xml:"creationType,attr"`
	CurrentShapeViewLevel string  `xml:"currentShapeViewLevel,attr"`
	DsColor               string  `xml:"dsColor,attr"`
	DsOffsetX             float64 `xml:"dsOffsetX,attr"`
	DsOffsetY             float64 `xml:"dsOffsetY,attr"`
	DsRadius              float64 `xml:"dsRadius,attr"`
	DsSpread              float64 `xml:"dsSpread,attr"`
	FillRule              string  `xml:"fillRule,attr"`
	FillTexture           string  `xml:"fillTexture,attr"`
	HighestViewLevel      string  `xml:"highestViewLevel,attr"`
	InsChoke              float64 `xml:"insChoke,attr"`
	InsColor              string  `xml:"insColor,attr"`
	InsOffsetX            float64 `xml:"insOffsetX,attr"`
	InsOffsetY            float64 `xml:"insOffsetY,attr"`
	InsRadius             float64 `xml:"insRadius,attr"`
	IsBoxBlur             bool    `xml:"isBoxBlur,attr"`
	IsContinent           bool    `xml:"isContinent,attr"`
	IsCurve               bool    `xml:"isCurve,attr"`
	IsDropShadow          bool    `xml:"isDropShadow,attr"`
	IsGMOnly              bool    `xml:"isGMOnly,attr"`
	IsInnerShadow         bool    `xml:"isInnerShadow,attr"`
	IsKingdom             bool    `xml:"isKingdom,attr"`
	IsMatchTileBorders    bool    `xml:"isMatchTileBorders,attr"`
	IsProvince            bool    `xml:"isProvince,attr"`
	IsSnapVertices        bool    `xml:"isSnapVertices,attr"`
	IsWorld               bool    `xml:"isWorld,attr"`
	LineCap               string  `xml:"lineCap,attr"`
	LineJoin              string  `xml:"lineJoin,attr"`
	MapLayer              string  `xml:"mapLayer,attr"`
	Opacity               float64 `xml:"opacity,attr"`
	StrokeColor           string  `xml:"strokeColor,attr"`
	StrokeTexture         string  `xml:"strokeTexture,attr"`
	StrokeType            string  `xml:"strokeType,attr"`
	StrokeWidth           float64 `xml:"strokeWidth,attr"`
	Tags                  string  `xml:"tags,attr"`
	Type                  string  `xml:"type,attr"`

	// elements
	Points []Point_t `xml:"p"`
}

type ShapeConfig_t struct {
	// elements
	ShapeStyles []ShapeStyle_t `xml:"shapestyle"`
	InnerText   string         `xml:",chardata"`
}

type ShapeStyle_t struct {
	// attributes
	BbHeight      float64 `xml:"bbHeight,attr"`
	BbIterations  int     `xml:"bbIterations,attr"`
	BbWidth       float64 `xml:"bbWidth,attr"`
	BoxBlur       bool    `xml:"boxBlur,attr"`
	DropShadow    bool    `xml:"dropShadow,attr"`
	DsOffsetX     float64 `xml:"dsOffsetX,attr"`
	DsOffsetY     float64 `xml:"dsOffsetY,attr"`
	DsRadius      float64 `xml:"dsRadius,attr"`
	DsSpread      float64 `xml:"dsSpread,attr"`
	Dscolor       string  `xml:"dscolor,attr"`
	FillPaint     string  `xml:"fillPaint,attr"`
	FillTexture   string  `xml:"fillTexture,attr"`
	InnerShadow   bool    `xml:"innerShadow,attr"`
	InsChoke      float64 `xml:"insChoke,attr"`
	InsColor      string  `xml:"insColor,attr"`
	InsOffsetX    float64 `xml:"insOffsetX,attr"`
	InsOffsetY    float64 `xml:"insOffsetY,attr"`
	InsRadius     float64 `xml:"insRadius,attr"`
	IsFractal     bool    `xml:"isFractal,attr"`
	Name          string  `xml:"name,attr"`
	Opacity       float64 `xml:"opacity,attr"`
	SnapVertices  bool    `xml:"snapVertices,attr"`
	StrokePaint   string  `xml:"strokePaint,attr"`
	StrokeTexture string  `xml:"strokeTexture,attr"`
	StrokeType    string  `xml:"strokeType,attr"`
	StrokeWidth   float64 `xml:"strokeWidth,attr"`
	Tags          string  `xml:"tags,attr"`

	// W2025 line rendering attributes (already modeled on <shape>).
	LineCap  string `xml:"lineCap,attr"`
	LineJoin string `xml:"lineJoin,attr"`
}

type Shapes_t struct {
	// elements
	Shapes []Shape_t `xml:"shape"`
}

type TerrainConfig_t struct {
	// elements
	InnerText string `xml:",chardata"`
}

type TerrainMap_t struct {
	// elements
	InnerText string `xml:",chardata"`
}

type TextConfig_t struct {
	// elements
	LabelStyles []LabelStyle_t `xml:"labelstyle"`
	InnerText   string         `xml:",chardata"`
}

type TextureConfig_t struct {
	// elements
	InnerText string `xml:",chardata"`
}

type Tiles_t struct {
	// attributes
	ViewLevel string `xml:"viewLevel,attr"`
	TilesWide int    `xml:"tilesWide,attr"` // number of columns of tiles
	TilesHigh int    `xml:"tilesHigh,attr"` // number of rows of tiles

	// elements
	TileRows []TileRow_t `xml:"tilerow"`
}

type TileRow_t struct {
	// elements
	InnerText string `xml:",chardata"`
}

// decodeRgba parses a Worldographer float-RGBA attribute ("r,g,b,a"). It folds
// "", "null", and the literal opaque black "0.0,0.0,0.0,1.0" all to nil.
//
// Latent fidelity risk: this fold assumes opaque black always means "absent".
// That round-trips losslessly only for fields the encoder re-emits with rgbas
// (which writes nil back as "0.0,0.0,0.0,1.0"). For fields that instead OMIT
// when nil (tile CustomBackgroundColor) or emit "null" via rgbans (feature
// Color/RingColor, some shapeStyle colors), a genuinely opaque-black on-disk
// value changes byte form across a round-trip.
//
// This IS now exercised by testdata/w2025-populated.xml: Features[0].Color
// = "0.0,0.0,0.0,1.0" decodes here to nil, and encodeFeature re-emits it via
// rgbans as "null"; re-decoding "null" also yields nil. So at the Map_t level it
// round-trips losslessly (nil both passes), and only the on-disk byte form shifts
// "0.0,0.0,0.0,1.0" -> "null" — lossless under Worldographer's semantics where
// both spellings mean "no color". The remaining on-disk ambiguity (a map that
// truly needs opaque-black distinct from "null") is still unresolved, so the fold
// is left as-is.
func decodeRgba(s string) (rgba *wxx.RGBA_t, err error) {
	if s == "" || s == "null" || s == "0.0,0.0,0.0,1.0" {
		return nil, nil
	}
	rgba = &wxx.RGBA_t{}
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

func decodeZeroableRgba(s string) (rgba *wxx.RGBA_t, err error) {
	if s == "" || s == "null" {
		return nil, nil
	}
	rgba = &wxx.RGBA_t{}
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
