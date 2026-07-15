// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

import (
	"time"

	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx/hexg"
)

// Map_t is the in-memory representation of the map data.
// We have created this to work with the known versions of Worldographer XML data.
// We are assuming that this will continue to work with future versions of the application.
type Map_t struct {
	MetaData struct {
		// AppVersion is this tool's own version -- a genuine semver, and the only
		// semver in this struct. The versions the file states are not semver and
		// live in Version.
		AppVersion semver.Version `json:"appVersion"`
		// Version is the file's on-disk version identity: map/@version and
		// map/@schema as the two independent axes they are (ADR 0004). It is
		// populated by every decoder, and its Schema selects the codec an encode
		// dispatches to.
		//
		// It replaces DataVersion, which held one semver whose Major was a family
		// year and whose Minor.Patch meant the application version for classic and
		// the schema version for W2025 -- a different axis per family, in one slot,
		// keyed by a label no file states (ADR 0004 Decisions 2 and 4).
		Version Version_t `json:"version"`
		// Worldographer defines the metadata for the WXX file
		Worldographer struct {
			Name    string    `json:"name"`    // name of input
			Created time.Time `json:"created"` // timestamp of input
			Release string    `json:"release"` // Worldographer release (e.g., 2025)
			Version string    `json:"version"` // Worldographer/Hexographer version (e.g. 2.06)
			Schema  string    `json:"schema"`  // Worldographer XML Schema version (e.g. 1.06)
		} `json:"worldographer"`
		Created string `json:"created"` // timestamp of this file
	} `json:"meta-data"`

	// attributes
	Type                      string             `json:"type,omitempty"`
	Release                   string             `json:"release,omitempty"`
	Version                   string             `json:"version,omitempty"`
	Schema                    string             `json:"schema,omitempty"`
	LastViewLevel             string             `json:"lastViewLevel,omitempty"`
	ContinentFactor           int                `json:"continentFactor,omitempty"`
	KingdomFactor             int                `json:"kingdomFactor,omitempty"`
	ProvinceFactor            int                `json:"provinceFactor,omitempty"`
	WorldToContinentHOffset   float64            `json:"worldToContinentHOffset,omitempty"`
	ContinentToKingdomHOffset float64            `json:"continentToKingdomHOffset,omitempty"`
	KingdomToProvinceHOffset  float64            `json:"kingdomToProvinceHOffset,omitempty"`
	WorldToContinentVOffset   float64            `json:"worldToContinentVOffset,omitempty"`
	ContinentToKingdomVOffset float64            `json:"continentToKingdomVOffset,omitempty"`
	KingdomToProvinceVOffset  float64            `json:"kingdomToProvinceVOffset,omitempty"`
	HexWidth                  float64            `json:"hexWidth,omitempty"`
	HexHeight                 float64            `json:"hexHeight,omitempty"`
	GridOrientation           hexg.Orientation_e `json:"gridOrientation,omitempty"` // orientation for hexg package
	HexOrientation            string             `json:"hexOrientation,omitempty"`  // "COLUMNS" or ??
	RowsHigh                  int                `json:"rowsHigh,omitempty"`        // number of rows (derived from TilesHigh based on orientation)
	ColumnsWide               int                `json:"columnsWide,omitempty"`     // number of columns (derived from TilesWide based on orientation)
	MapProjection             Projection_e       `json:"mapProjection,omitempty"`
	ShowNotes                 bool               `json:"showNotes,omitempty"`
	ShowGMOnly                bool               `json:"showGMOnly,omitempty"`
	ShowGMOnlyGlow            bool               `json:"showGMOnlyGlow,omitempty"`
	ShowFeatureLabels         bool               `json:"showFeatureLabels,omitempty"`
	ShowGrid                  bool               `json:"showGrid,omitempty"`
	ShowGridNumbers           bool               `json:"showGridNumbers,omitempty"`
	ShowShadows               bool               `json:"showShadows,omitempty"`
	TriangleSize              int                `json:"triangleSize,omitempty"`
	HScrollbarPos             float64            `json:"hScrollbarPos,omitempty"` // W2025 UI scroll position
	VScrollbarPos             float64            `json:"vScrollbarPos,omitempty"` // W2025 UI scroll position

	// elements
	GridAndNumbering *GridAndNumbering_t `json:"gridAndNumbering,omitempty"`

	// BlurTerrainBG is a W2025 top-level element; nil means absent from the file.
	BlurTerrainBG *BlurTerrainBG_t `json:"blurTerrainBG,omitempty"`

	// ExtraTerrain is a W2025 top-level element; nil means absent from the file.
	// Its content is opaque -- see ExtraTerrain_t.
	ExtraTerrain *ExtraTerrain_t `json:"extraTerrain,omitempty"`

	// TerrainMap assigns numbers to each terrain type.
	// The terrain type is used in the TileRow struct.
	TerrainMap *TerrainMap_t `json:"terrainMap,omitempty"`

	// MapLayers assigns a boolean "isVisible" to each layer.
	MapLayers []*MapLayer_t `json:"mapLayers,omitempty"`

	Tiles *Tiles_t `json:"tiles,omitempty"`

	MapKey *MapKey_t `json:"mapKey,omitempty"`

	Features []*Feature_t `json:"features,omitempty"`

	Labels []*Label_t `json:"labels,omitempty"`

	Shapes []*Shape_t `json:"shapes,omitempty"`

	Notes []*Note_t `json:"notes,omitempty"`

	Informations *Informations_t `json:"informations"`

	Configuration *Configuration_t `json:"configuration"`
}

// BlurTerrainBG models the W2025 top-level <blurTerrainBG> element. It is
// referenced as a pointer on Map_t so that nil distinguishes "absent from the
// file" from "present with zero-valued attributes".
type BlurTerrainBG_t struct {
	Blur        bool    `json:"blur,omitempty"`
	TopBleed    float64 `json:"topBleed,omitempty"`
	BottomBleed float64 `json:"bottomBleed,omitempty"`
	Randomness  float64 `json:"randomness,omitempty"`
	BlurStart   float64 `json:"blurStart,omitempty"`
	BlurEnd     float64 `json:"blurEnd,omitempty"`
}

type Configuration_t struct {
	TerrainConfig []*TerrainConfig_t `json:"terrain-config,omitempty"`
	FeatureConfig []*FeatureConfig_t `json:"feature-config,omitempty"`
	TextureConfig []*TextureConfig_t `json:"texture-config,omitempty"`
	TextConfig    *TextConfig_t      `json:"text-config,omitempty"`
	ShapeConfig   *ShapeConfig_t     `json:"shape-config"`
	InnerText     string             `json:"InnerText,omitempty"`
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
	LabelDistance     int     `json:"labelDistance,omitempty"`
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

// ExtraTerrain models the W2025 top-level <extraTerrain> element. It is a
// pointer on Map_t so nil distinguishes absent from present-but-empty.
//
// The two tracked 2.06 fixtures show both shapes it takes:
// 2025-2.06-13x11-941577-blank.wxx carries an EMPTY container (InnerXML is "\n",
// the pretty-printer's newline), while 2025-2.06-13x11-941577-layers.wxx carries
// 183 bytes of real content -- a <mapLayer name="Terrain Layer"> holding a
// <terrainAndLocation> that binds one hex's terrain to that layer.
//
// InnerXML captures whatever is between the tags VERBATIM, and that is the
// element's entire modeling: nothing here understands a mapLayer or a
// terrainAndLocation. The bytes round-trip 2025 -> 2025 intact, but the model
// cannot answer a question about them, which is why encoding a populated
// <extraTerrain> to a target that has no such element is an error rather than a
// reported loss (#34 tracks modeling it; xmlio's downgradeLoss holds the
// contract).
type ExtraTerrain_t struct {
	InnerXML string `json:"innerXML,omitempty"`
}

type GridAndNumbering_t struct {
	Color0                      string  `json:"color0,omitempty"` // hex - "0x00000040"
	Color1                      string  `json:"color1,omitempty"` // hex - "0x00000040"
	Color2                      string  `json:"color2,omitempty"` // hex - "0x00000040"
	Color3                      string  `json:"color3,omitempty"` // hex - "0x00000040"
	Color4                      string  `json:"color4,omitempty"` // hex - "0x00000040"
	Width0                      float64 `json:"width0,omitempty"`
	Width1                      float64 `json:"width1,omitempty"`
	Width2                      float64 `json:"width2,omitempty"`
	Width3                      float64 `json:"width3,omitempty"`
	Width4                      float64 `json:"width4,omitempty"`
	GridOffsetContinentKingdomX float64 `json:"gridOffsetContinentKingdomX,omitempty"`
	GridOffsetContinentKingdomY float64 `json:"gridOffsetContinentKingdomY,omitempty"`
	GridOffsetWorldContinentX   float64 `json:"gridOffsetWorldContinentX,omitempty"`
	GridOffsetWorldContinentY   float64 `json:"gridOffsetWorldContinentY,omitempty"`
	GridOffsetWorldKingdomX     float64 `json:"gridOffsetWorldKingdomX,omitempty"`
	GridOffsetWorldKingdomY     float64 `json:"gridOffsetWorldKingdomY,omitempty"`
	GridSquare                  int     `json:"gridSquare,omitempty"`
	GridSquareHeight            float64 `json:"gridSquareHeight,omitempty"` // "-1.0" is special?
	GridSquareWidth             float64 `json:"gridSquareWidth,omitempty"`  // "-1.0" is special?
	GridOffsetX                 float64 `json:"gridOffsetX,omitempty"`
	GridOffsetY                 float64 `json:"gridOffsetY,omitempty"`
	NumberFont                  string  `json:"numberFont,omitempty"`
	NumberColor                 string  `json:"numberColor,omitempty"` // hex - "0x000000ff"
	NumberSize                  int     `json:"numberSize,omitempty"`
	NumberStyle                 string  `json:"numberStyle,omitempty"`     // "PLAIN" or ??
	NumberFirstCol              int     `json:"numberFirstCol,omitempty"`  // "0"
	NumberFirstRow              int     `json:"numberFirstRow,omitempty"`  // "0"
	NumberOrder                 string  `json:"numberOrder,omitempty"`     // "COL_ROW" or ??
	NumberPosition              string  `json:"numberPosition,omitempty"`  // "BOTTOM" or ??
	NumberPrePad                string  `json:"numberPrePad,omitempty"`    // "DOUBLE_ZERO" or ??
	NumberSeparator             string  `json:"numberSeparator,omitempty"` // "." or free text?
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

type Informations_t struct {
	Informations []*Information_t `json:"informations,omitempty"`
	InnerText    string           `json:"innerText,omitempty"`
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

	// W2025 drop-shadow attributes. DropShadowColor is a nullable string that
	// preserves the literal "null" spelling (mirrors Shape_t.DsColor).
	DropShadowColor  string  `json:"dropShadowColor,omitempty"`
	DropShadowRadius float64 `json:"dropShadowRadius,omitempty"`
	DropShadowSpread float64 `json:"dropShadowSpread,omitempty"`
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
	Name      string  `json:"name"`
	IsVisible bool    `json:"isVisible"`
	Opacity   float64 `json:"opacity,omitempty"`
}

type Note_t struct {
	InnerText string `json:"innerText,omitempty"`

	// attributes
	Key       string  `json:"key,omitempty"`
	ViewLevel string  `json:"viewLevel,omitempty"`
	X         float64 `json:"x,omitempty"`
	Y         float64 `json:"y,omitempty"`
	Filename  string  `json:"filename,omitempty"`
	Parent    string  `json:"parent,omitempty"`
	Color     *RGBA_t `json:"color,omitempty"`
	Title     string  `json:"title,omitempty"`
	IsGMOnly  bool    `json:"isGMOnly,omitempty"`

	// notetext CDATA body
	NoteText string `json:"notetext,omitempty"`
}

type Point_t struct {
	Type string  `json:"type,omitempty"`
	X    float64 `json:"x,omitempty"`
	Y    float64 `json:"y,omitempty"`
}

type Projection_e int

const (
	_ Projection_e = iota
	FLAT
	ICOSAHEDRAL
)

type Resources_t struct {
	Animal int
	Brick  int
	Crops  int
	Gems   int
	Lumber int
	Metals int
	Rock   int
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

	// W2025 line rendering attributes (mirrors Shape_t.LineCap/LineJoin).
	LineCap  string `json:"lineCap,omitempty"`
	LineJoin string `json:"lineJoin,omitempty"`
}

type Terrain_t struct {
	Index int    `json:"index"`
	Label string `json:"label"`
}

type TerrainConfig_t struct {
	InnerText string `json:"innerText,omitempty"`
}

type TerrainMap_t struct {
	Data map[string]int `json:"data,omitempty"`
	List []*Terrain_t   `json:"list,omitempty"`
}

type TextConfig_t struct {
	LabelStyles []*LabelStyle_t `json:"labelStyles,omitempty"`
	InnerText   string          `json:"innerText,omitempty"`
}

type TextureConfig_t struct {
	InnerText string `json:"innerText,omitempty"`
}

type Tile_t struct {
	Coords                hexg.CubeCoord
	Row                   int
	Column                int
	Terrain               int // lookup into TerrainMap
	Elevation             float64
	IsIcy                 bool
	IsGMOnly              bool
	Resources             Resources_t
	CustomBackgroundColor *RGBA_t
}

type Tiles_t struct {
	ViewLevel string `json:"viewLevel,omitempty"`
	TilesWide int    `json:"tilesWide,omitempty"` // number of columns of tiles (x)
	TilesHigh int    `json:"tilesHigh,omitempty"` // number of rows of tiles    (y)

	// Tiles is a two-dimensional array that is indexed by [col][row]
	Tiles [][]*Tile_t `json:"tiles,omitempty"`
}
