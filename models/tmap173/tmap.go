// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package tmap173 defines the data structures used to create XML output that Worldographer expects.
package tmap173

type Map struct {
	Type                      string // "WORLD"
	Version                   string // "1.73"
	LastViewLevel             string // "WORLD"
	ContinentFactor           string // "-1"
	KingdomFactor             string // "-1"
	ProvinceFactor            string // "-1"
	WorldToContinentHOffset   string // "0.0"
	ContinentToKingdomHOffset string // "0.0"
	KingdomToProvinceHOffset  string // "0.0"
	WorldToContinentVOffset   string // "0.0"
	ContinentToKingdomVOffset string // "0.0"
	KingdomToProvinceVOffset  string // "0.0"
	HexWidth                  string // "120.97791408032022"
	HexHeight                 string // "104.78814558711076"
	HexOrientation            string // "COLUMNS"
	MapProjection             string // "FLAT"
	ShowNotes                 string // "true"
	ShowGMOnly                string // "false"
	ShowGMOnlyGlow            string // "false"
	ShowFeatureLabels         string // "true"
	ShowGrid                  string // "true"
	ShowGridNumbers           string // "false"
	ShowShadows               string // "true"
	TriangleSize              string // "12"

	GridAndNumbering struct {
		Color0                      string //
		Color1                      string //
		Color2                      string //
		Color3                      string //
		Color4                      string //
		Width0                      string //
		Width1                      string //
		Width2                      string //
		Width3                      string //
		Width4                      string //
		GridOffsetContinentKingdomX string //
		GridOffsetContinentKingdomY string //
		GridOffsetWorldContinentX   string //
		GridOffsetWorldContinentY   string //
		GridOffsetWorldKingdomX     string //
		GridOffsetWorldKingdomY     string //
		GridSquare                  string //
		GridSquareHeight            string //
		GridSquareWidth             string //
		GridOffsetX                 string //
		GridOffsetY                 string //
		NumberFont                  string //
		NumberColor                 string //
		NumberSize                  string //
		NumberStyle                 string //
		NumberFirstCol              string //
		NumberFirstRow              string //
		NumberOrder                 string //
		NumberPosition              string //
		NumberPrePad                string //
		NumberSeparator             string //
	}

	TerrainMap string

	MapLayer []MapLayer

	Tiles struct {
		ViewLevel string
		TilesWide string // "30"
		TilesHigh string // "21"

		TileRows []string
	}

	MapKey struct {
		PositionX         string
		PositionY         string
		Viewlevel         string // "null", "WORLD"
		Height            string
		BackgroundColor   string
		BackgroundOpacity string
		TitleText         string
		TitleFontFace     string
		TitleFontColor    string
		TitleFontBold     string
		TitleFontItalic   string
		TitleScale        string
		ScaleText         string
		ScaleFontFace     string
		ScaleFontColor    string
		ScaleFontBold     string
		ScaleFontItalic   string
		ScaleScale        string
		EntryFontFace     string
		EntryFontColor    string
		EntryFontBold     string
		EntryFontItalic   string
		EntryScale        string
	}

	Features []*Feature

	Labels []*Label

	Shapes []*Shape

	Notes []*Note

	Information          []*Information
	InformationInnerText string

	Configuration struct {
		TerrainConfig TerrainConfig
		FeatureConfig FeatureConfig
		TextureConfig TextureConfig
		TextConfig    TextConfig
		ShapeConfig   ShapeConfig
		InnerText     string
	}
}

type Feature struct {
	Type              string
	Rotate            string
	Uuid              string
	MapLayer          string
	IsFlipHorizontal  string
	IsFlipVertical    string
	Scale             string
	ScaleHt           string
	Tags              string
	Color             string
	RingColor         string
	IsGMOnly          string
	IsPlaceFreely     string
	LabelPosition     string
	LabelDistance     string
	IsWorld           string
	IsContinent       string
	IsKingdom         string
	IsProvince        string
	IsFillHexBottom   string
	IsHideTerrainIcon string

	Location *FeatureLocation
	Label    *Label
}

type FeatureConfig struct {
	InnerText string `json:"innerText,omitempty"`
}

type FeatureLocation struct {
	ViewLevel string
	X         string
	Y         string
}

type Information struct {
	Uuid         string
	Type         string
	Title        string
	Rulers       string
	Government   string
	Cultures     string
	Language     string
	ReligionType string
	Culture      string
	HolySymbol   string
	Domains      string

	Details   []*InformationDetail
	InnerText string
}

type InformationDetail struct {
	Uuid         string
	Type         string
	Title        string
	Rulers       string
	Government   string
	Cultures     string
	Language     string
	ReligionType string
	Culture      string
	HolySymbol   string
	Domains      string
	InnerText    string
}

type Label struct {
	MapLayer        string
	Style           string
	FontFace        string
	Color           string
	OutlineColor    string
	OutlineSize     string
	Rotate          string
	IsBold          string
	IsItalic        string
	IsWorld         string
	IsContinent     string
	IsKingdom       string
	IsProvince      string
	IsGMOnly        string
	Tags            string
	BackgroundColor string

	Location  *LabelLocation
	InnerText string
}

type LabelLocation struct {
	ViewLevel string
	X         string
	Y         string
	Scale     string
}

type LabelStyle struct {
	Name            string
	FontFace        string
	Scale           string
	IsBold          string
	IsItalic        string
	Color           string
	BackgroundColor string
	OutlineSize     string
	OutlineColor    string
}

type MapLayer struct {
	// attributes
	Name      string
	IsVisible bool
}

type Note struct {
	InnerText string
}

type Point struct {
	Type string
	X    string
	Y    string
}

type Shape struct {
	BbHeight              string
	BbIterations          string
	BbWidth               string
	CreationType          string
	CurrentShapeViewLevel string
	DsColor               string
	DsOffsetX             string
	DsOffsetY             string
	DsRadius              string
	DsSpread              string
	FillRule              string
	FillTexture           string
	HighestViewLevel      string
	InsChoke              string
	InsColor              string
	InsOffsetX            string
	InsOffsetY            string
	InsRadius             string
	IsBoxBlur             string
	IsContinent           string
	IsCurve               string
	IsDropShadow          string
	IsGMOnly              string
	IsInnerShadow         string
	IsKingdom             string
	IsMatchTileBorders    string
	IsProvince            string
	IsSnapVertices        string
	IsWorld               string
	LineCap               string
	LineJoin              string
	MapLayer              string
	Opacity               string
	StrokeColor           string
	StrokeTexture         string
	StrokeType            string
	StrokeWidth           string
	Tags                  string
	Type                  string

	Points []*Point
}

type ShapeConfig struct {
	ShapeStyles []*ShapeStyle
	InnerText   string
}

type ShapeStyle struct {
	Name          string
	StrokeType    string
	IsFractal     string
	StrokeWidth   string
	Opacity       string
	SnapVertices  string
	Tags          string
	DropShadow    string
	InnerShadow   string
	BoxBlur       string
	DsSpread      string
	DsRadius      string
	DsOffsetX     string
	DsOffsetY     string
	InsChoke      string
	InsRadius     string
	InsOffsetX    string
	InsOffsetY    string
	BbWidth       string
	BbHeight      string
	BbIterations  string
	FillTexture   string
	StrokeTexture string
	StrokePaint   string
	FillPaint     string
	DsColor       string
	InsColor      string
}

type TerrainConfig struct {
	InnerText string
}

type TextConfig struct {
	LabelStyles []*LabelStyle
	InnerText   string
}

type TextureConfig struct {
	InnerText string
}
