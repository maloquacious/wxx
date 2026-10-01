// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"errors"
	"fmt"
	"time"

	"github.com/maloquacious/hexg"
	"github.com/maloquacious/wxx"
)

// NewMap returns a map of columns x rows Blank hexes, carrying the defaults the
// application version app writes for File > New World/Kingdom map (issue #136).
//
// The defaults are read off the blank fixtures, testdata/2025-<app>-13x11-
// 941577-blank.wxx, whose recipe is the "Blank" section of testdata/README.md,
// and off the matching -rows fixtures for the ROWS hex size. They are this
// codec's knowledge because they are what the builds on schema 1.06 write; a
// later codec states its own.
//
// What the map does NOT carry is the lore: the app fills <informations> from the
// random seed, and NewMap leaves it empty.
//
// The show* flags are UI settings, saved as they stood when the file was saved,
// not defaults of the application version: the 2.06 blank fixture states
// showGrid="false" only because the grid was hidden when it was saved. NewMap
// shows the grid and its numbers for every version. Tiles[col][row] is filled in both
// orientations, each tile's Coords, Column and Row set as decodeTiles sets them.
//
// The caller has already checked columns, rows and hexOrientation (see
// xmlio.NewMap); NewMap checks only app, which is its own business.
func NewMap(app string, columns, rows int, hexOrientation string) (*wxx.Map_t, error) {
	if err := acceptedApps.VerifyApp(app); err != nil {
		return nil, err
	}
	version, err := wxx.ParseDotted(app)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("version %q", app))
	}
	schema, err := wxx.ParseDotted(acceptedApps.Schema)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("schema %q", acceptedApps.Schema))
	}

	w := &wxx.Map_t{}
	w.MetaData.AppVersion = wxx.Version()
	w.MetaData.Version = wxx.Version_t{App: version, Schema: &schema}
	w.MetaData.Created = time.Now().UTC().Format(time.RFC3339)

	w.Type = "WORLD"
	w.LastViewLevel = "WORLD"
	w.HexOrientation = hexOrientation
	switch hexOrientation {
	case "COLUMNS":
		w.HexWidth, w.HexHeight = 46.18, 40.0
	case "ROWS":
		w.HexWidth, w.HexHeight = 40.0, 46.18
	default:
		return nil, errors.Join(wxx.ErrInvalidHexOrientation, fmt.Errorf("hexOrientation %q: want \"COLUMNS\" or \"ROWS\"", hexOrientation))
	}
	w.MapProjection = wxx.FLAT
	w.ShowGMOnly = true
	w.ShowFeatureLabels = true
	w.ShowGrid = true
	w.ShowGridNumbers = true
	w.ShowShadows = true
	w.TriangleSize = 12

	w.GridAndNumbering = &wxx.GridAndNumbering_t{
		Color0:           "0x00000040",
		Color1:           "0x00000040",
		Color2:           "0x00000040",
		Color3:           "0x00000040",
		Color4:           "0x00000040",
		Width0:           1,
		Width1:           2,
		Width2:           3,
		Width3:           4,
		Width4:           1,
		GridSquareHeight: -1,
		GridSquareWidth:  -1,
		NumberFont:       "Arial",
		NumberColor:      "0x000000ff",
		NumberSize:       20,
		NumberStyle:      "PLAIN",
		NumberOrder:      "COL_ROW",
		NumberPosition:   "BOTTOM",
		NumberPrePad:     "DOUBLE_ZERO",
		NumberSeparator:  ".",
	}
	w.BlurTerrainBG = &wxx.BlurTerrainBG_t{
		TopBleed:    0.33,
		BottomBleed: 0.65,
		Randomness:  0.1,
		BlurStart:   0.4,
		BlurEnd:     0.95,
	}
	w.ExtraTerrain = &wxx.ExtraTerrain_t{}

	// The eight standard layers, in the order the app writes them.
	for _, name := range []string{"Labels", "Grid", "Features", "Above Terrain", "Terrain Land", "Above Water", "Terrain Water", "Below All"} {
		w.MapLayers = append(w.MapLayers, &wxx.MapLayer_t{Name: name, IsVisible: true, Opacity: 1})
	}

	// Blank goes in through EnsureTerrain, so it sits at index 0 of an empty
	// table (#81).
	w.TerrainMap = &wxx.TerrainMap_t{}
	blank := w.TerrainMap.EnsureTerrain(wxx.BlankTerrain)
	w.Tiles = blankTiles(columns, rows, blank, hexOrientation)
	// RowsHigh and ColumnsWide are the decoder's labels; see decodeTiles.
	if hexOrientation == "COLUMNS" {
		w.RowsHigh, w.ColumnsWide = rows, columns
	} else {
		w.RowsHigh, w.ColumnsWide = columns, rows
	}

	black := func() *wxx.RGBA_t { return &wxx.RGBA_t{A: 1} }
	w.MapKey = &wxx.MapKey_t{
		Viewlevel:         "null",
		Height:            -1,
		BackgroundColor:   &wxx.RGBA_t{R: 0.9803921580314636, G: 0.9215686321258545, B: 0.843137264251709, A: 1},
		BackgroundOpacity: 50,
		TitleText:         "Map Key",
		TitleFontFace:     "Arial",
		TitleFontColor:    black(),
		TitleFontBold:     true,
		TitleScale:        80,
		ScaleText:         "1 Hex = ? units",
		ScaleFontFace:     "Arial",
		ScaleFontColor:    black(),
		ScaleFontBold:     true,
		ScaleScale:        65,
		EntryFontFace:     "Arial",
		EntryFontColor:    black(),
		EntryFontBold:     true,
		EntryScale:        55,
	}

	w.Informations = &wxx.Informations_t{}
	w.Configuration = &wxx.Configuration_t{
		TextConfig:  &wxx.TextConfig_t{LabelStyles: defaultLabelStyles()},
		ShapeConfig: &wxx.ShapeConfig_t{ShapeStyles: defaultShapeStyles()},
	}
	return w, nil
}

// blankTiles returns a columns x rows grid of tiles with terrain blank, indexed
// [col][row], with Coords, Column and Row set as decodeTiles sets them: odd-q
// for COLUMNS and odd-r for ROWS (#85, #130). A new tile's elevation is 1, as
// every Blank tile in the fixtures states it.
func blankTiles(columns, rows, blank int, hexOrientation string) *wxx.Tiles_t {
	tiles := &wxx.Tiles_t{
		ViewLevel: "WORLD",
		TilesWide: columns,
		TilesHigh: rows,
		Tiles:     make([][]*wxx.Tile_t, columns),
	}
	for x := range columns {
		tiles.Tiles[x] = make([]*wxx.Tile_t, rows)
		for y := range rows {
			t := &wxx.Tile_t{Column: x, Row: y, Terrain: blank, Elevation: 1}
			if hexOrientation == "COLUMNS" {
				t.Coords = hexg.NewOffsetCoord(x, y).QOffsetToCube(false)
			} else {
				t.Coords = hexg.NewOffsetCoord(x, y).ROffsetToCube(false)
			}
			tiles.Tiles[x][y] = t
		}
	}
	return tiles
}

// defaultLabelStyles returns the seven label styles a new map carries, in the
// order the app writes them.
func defaultLabelStyles() []*wxx.LabelStyle_t {
	black := func() *wxx.RGBA_t { return &wxx.RGBA_t{A: 1} }
	white := func() *wxx.RGBA_t { return &wxx.RGBA_t{R: 1, G: 1, B: 1, A: 1} }
	darkRed := func() *wxx.RGBA_t { return &wxx.RGBA_t{R: 0.545098066329956, A: 1} }
	style := func(name, fontFace string, scale float64, isBold bool, color *wxx.RGBA_t, outlineSize float64, outlineColor *wxx.RGBA_t) *wxx.LabelStyle_t {
		return &wxx.LabelStyle_t{
			Name:            name,
			FontFace:        fontFace,
			Scale:           scale,
			IsBold:          isBold,
			Color:           color,
			BackgroundColor: nil, // "null"
			OutlineSize:     outlineSize,
			OutlineColor:    outlineColor, // nil is "null"
			DropShadowColor: "null",
		}
	}
	return []*wxx.LabelStyle_t{
		style("Nation", "Times", 80, true, black(), 2, white()),
		style("Geography", "Arial", 70, false, darkRed(), 2, white()),
		style("Village", "Times", 28, false, black(), 0, nil),
		style("Geography Minor", "Arial", 50, false, darkRed(), 1, white()),
		style("Geography Major", "Arial", 80, false, darkRed(), 2, white()),
		style("City", "Times", 33, false, black(), 0, nil),
		style("Province", "Times", 70, false, black(), 1, white()),
	}
}

// defaultShapeStyles returns the seven shape styles a new map carries, in the
// order the app writes them.
func defaultShapeStyles() []*wxx.ShapeStyle_t {
	rgb := func(r, g, b float64) *wxx.RGBA_t { return &wxx.RGBA_t{R: r, G: g, B: b, A: 1} }
	shadow := func() *wxx.RGBA_t { return rgb(0.800000011920929, 0.8100000023841858, 0.7599999904632568) }
	line := func(name, strokeType, tags, lineCap string, strokePaint *wxx.RGBA_t) *wxx.ShapeStyle_t {
		return &wxx.ShapeStyle_t{
			Name:          name,
			StrokeType:    strokeType,
			StrokeWidth:   10,
			Opacity:       100,
			Tags:          tags,
			FillTexture:   "null",
			StrokeTexture: "null",
			StrokePaint:   strokePaint,
			LineCap:       lineCap,
			LineJoin:      "ROUND",
		}
	}
	riverIsometric := &wxx.ShapeStyle_t{
		Name:          "River Isometric",
		StrokeType:    "SIMPLE",
		StrokeWidth:   3,
		Opacity:       100,
		Tags:          "river",
		DropShadow:    true,
		BoxBlur:       true,
		DsSpread:      0.7,
		DsRadius:      14,
		BbWidth:       2,
		BbHeight:      2,
		BbIterations:  3,
		FillTexture:   "null",
		StrokeTexture: "null",
		StrokePaint:   rgb(0.3100000023841858, 0.5699999928474426, 0.6800000071525574),
		DsColor:       shadow(),
		LineCap:       "ROUND",
		LineJoin:      "ROUND",
	}
	coastIsometric := &wxx.ShapeStyle_t{
		Name:          "Coast Isometric",
		StrokeType:    "COAST",
		StrokeWidth:   0.5,
		Opacity:       100,
		Tags:          "coast",
		DropShadow:    true,
		BoxBlur:       true,
		DsSpread:      0.85,
		DsRadius:      10,
		BbWidth:       3,
		BbHeight:      3,
		BbIterations:  3,
		FillTexture:   "Sea",
		StrokeTexture: "null",
		StrokePaint:   rgb(0.9200000166893005, 0.9200000166893005, 0.9200000166893005),
		DsColor:       shadow(),
		LineCap:       "ROUND",
		LineJoin:      "ROUND",
	}
	border := line("Border", "SIMPLE", "border", "SQUARE", rgb(0.8100000023841858, 0, 0))
	border.SnapVertices = true
	return []*wxx.ShapeStyle_t{
		line("Trail", "SIMPLE", "trail", "SQUARE", rgb(0.8100000023841858, 0.7099999785423279, 0.30000001192092896)),
		riverIsometric,
		coastIsometric,
		line("Road", "SIMPLE", "road", "SQUARE", rgb(0, 0, 0)),
		line("River", "SIMPLE", "river", "SQUARE", rgb(0.550000011920929, 0.699999988079071, 0.8500000238418579)),
		line("Shipping", "DOTTED", "shipping", "ROUND", rgb(1, 1, 1)),
		border,
	}
}
