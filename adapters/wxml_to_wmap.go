// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

import (
	"fmt"
	"github.com/maloquacious/wxx/models/wxml173"
	"github.com/maloquacious/wxx/models/wxx"
	"log"
	"strconv"
	"strings"
	"time"
)

// WXMLToWXX translates any known WXML mapping to the current WXX mapping.
// It returns an error if the input is not a known WXML mapping or
// if there are errors translating between the two mappings.
// Panics is the input is not a WXML mapping.
func WXMLToWXX(wxml any) (*wxx.Map, error) {
	switch m := wxml.(type) {
	case *wxml173.Map:
		return wxmlV173ToWXX(m)
	}
	panic(fmt.Sprintf("assert(wxml.type != %T)", wxml))
}

func wxmlV173ToWXX(m *wxml173.Map) (*wxx.Map, error) {
	var err error

	w := &wxx.Map{}
	w.MetaData.Version = "0.0.1"
	w.MetaData.Created = time.Now().UTC().Format(time.RFC3339)
	w.MetaData.Source.Name = "unknown"
	w.MetaData.Source.Created = "0001-01-01T00:00:00Z"

	w.ContinentFactor = m.ContinentFactor
	w.ContinentToKingdomHOffset = m.ContinentToKingdomHOffset
	w.ContinentToKingdomVOffset = m.ContinentToKingdomVOffset
	w.HexHeight = m.HexHeight
	if m.HexOrientation != "COLUMNS" {
		log.Printf("warning: calculations for x,y coords do not work for %q\n", m.HexOrientation)
	}
	w.HexOrientation = m.HexOrientation
	w.HexWidth = m.HexWidth
	w.KingdomFactor = m.KingdomFactor
	w.KingdomToProvinceHOffset = m.KingdomToProvinceHOffset
	w.KingdomToProvinceVOffset = m.KingdomToProvinceVOffset
	w.LastViewLevel = m.LastViewLevel
	w.MapProjection = m.MapProjection
	w.ProvinceFactor = m.ProvinceFactor
	w.ShowFeatureLabels = m.ShowFeatureLabels
	w.ShowGMOnly = m.ShowGMOnly
	w.ShowGMOnlyGlow = m.ShowGMOnlyGlow
	w.ShowGrid = m.ShowGrid
	if m.ShowGridNumbers == false {
		log.Printf("todo: showGridNumber overriden to 'true'\n")
		w.ShowGridNumbers = true
	} else {
		w.ShowGridNumbers = m.ShowGridNumbers
	}
	w.ShowNotes = m.ShowNotes
	w.ShowShadows = m.ShowShadows
	w.TriangleSize = m.TriangleSize
	w.Type = m.Type
	w.Version = m.Version
	w.WorldToContinentHOffset = m.WorldToContinentHOffset
	w.WorldToContinentVOffset = m.WorldToContinentVOffset

	w.GridAndNumbering.Color0 = m.GridAndNumbering.Color0
	w.GridAndNumbering.Color1 = m.GridAndNumbering.Color1
	w.GridAndNumbering.Color2 = m.GridAndNumbering.Color2
	w.GridAndNumbering.Color3 = m.GridAndNumbering.Color3
	w.GridAndNumbering.Color4 = m.GridAndNumbering.Color4
	w.GridAndNumbering.Width0 = m.GridAndNumbering.Width0
	w.GridAndNumbering.Width1 = m.GridAndNumbering.Width1
	w.GridAndNumbering.Width2 = m.GridAndNumbering.Width2
	w.GridAndNumbering.Width3 = m.GridAndNumbering.Width3
	w.GridAndNumbering.Width4 = m.GridAndNumbering.Width4
	w.GridAndNumbering.GridOffsetContinentKingdomX = m.GridAndNumbering.GridOffsetContinentKingdomX
	w.GridAndNumbering.GridOffsetContinentKingdomY = m.GridAndNumbering.GridOffsetContinentKingdomY
	w.GridAndNumbering.GridOffsetWorldContinentX = m.GridAndNumbering.GridOffsetWorldContinentX
	w.GridAndNumbering.GridOffsetWorldContinentY = m.GridAndNumbering.GridOffsetWorldContinentY
	w.GridAndNumbering.GridOffsetWorldKingdomX = m.GridAndNumbering.GridOffsetWorldKingdomX
	w.GridAndNumbering.GridOffsetWorldKingdomY = m.GridAndNumbering.GridOffsetWorldKingdomY
	w.GridAndNumbering.GridSquare = m.GridAndNumbering.GridSquare
	w.GridAndNumbering.GridSquareHeight = m.GridAndNumbering.GridSquareHeight
	w.GridAndNumbering.GridSquareWidth = m.GridAndNumbering.GridSquareWidth
	w.GridAndNumbering.GridOffsetX = m.GridAndNumbering.GridOffsetX
	w.GridAndNumbering.GridOffsetY = m.GridAndNumbering.GridOffsetY
	w.GridAndNumbering.NumberFont = m.GridAndNumbering.NumberFont
	w.GridAndNumbering.NumberColor = m.GridAndNumbering.NumberColor
	w.GridAndNumbering.NumberSize = m.GridAndNumbering.NumberSize
	w.GridAndNumbering.NumberStyle = m.GridAndNumbering.NumberStyle
	w.GridAndNumbering.NumberFirstCol = m.GridAndNumbering.NumberFirstCol
	w.GridAndNumbering.NumberFirstRow = m.GridAndNumbering.NumberFirstRow
	w.GridAndNumbering.NumberOrder = m.GridAndNumbering.NumberOrder
	w.GridAndNumbering.NumberPosition = m.GridAndNumbering.NumberPosition
	w.GridAndNumbering.NumberPrePad = m.GridAndNumbering.NumberPrePad
	w.GridAndNumbering.NumberSeparator = m.GridAndNumbering.NumberSeparator

	// convert terrain map. in the source, the terrain key and values are
	// stored as tab delimited columns.
	w.TerrainMap.Data = map[string]int{}
	if fields := strings.Split(m.TerrainMap.InnerText, "\t"); len(fields)%2 != 0 {
		return w, fmt.Errorf("expected even number of fields, got odd")
	} else {
		for len(fields) != 0 {
			t := &wxx.Terrain{
				Label: fields[0],
			}
			t.Index, err = strconv.Atoi(fields[1])
			if err != nil {
				return w, fmt.Errorf("field: %s: invalid index: %w", fields[0], err)
			}
			w.TerrainMap.List = append(w.TerrainMap.List, t)
			w.TerrainMap.Data[t.Label] = t.Index
			fields = fields[2:]
		}
	}

	for _, layer := range m.MapLayers {
		w.MapLayer = append(w.MapLayer, wxx.MapLayer{Name: layer.Name, IsVisible: layer.IsVisible})
	}

	w.Tiles.ViewLevel = m.Tiles.ViewLevel
	w.Tiles.TilesWide = m.Tiles.TilesWide
	w.Tiles.TilesHigh = m.Tiles.TilesHigh
	log.Printf("hey, tilesHigh is %d tilesWide is %d\n", m.Tiles.TilesHigh, m.Tiles.TilesWide)
	isFirstTileRow := true
	for _, tilerow := range m.Tiles.TileRows {
		x, y := len(w.Tiles.TileRows), 0
		w.Tiles.TileRows = append(w.Tiles.TileRows, make([]*wxx.Tile, w.Tiles.TilesHigh))
		for _, line := range strings.Split(tilerow.InnerText, "\n") {
			if len(line) == 0 { // ignore blank lines
				continue
			}
			isXEdge, isYEdge := x == 0, y == 0
			t := &wxx.Tile{Row: x, Column: y}
			w.Tiles.TileRows[x][y] = t
			y++
			// values are TerrainMapIndex Elevation IsIcy IsGMOnly Animals (Z|(Brick Crops Gems Lumber Metals Rock)) RGBA?
			values := strings.Split(line, "\t")
			//fmt.Printf("tilerow: %d %d: len(inner) %d lines %d line %d values %d\n", r, i+1, len(element.InnerText), len(lines), len(line), len(values))
			switch len(values) {
			case 6, 7, 11, 12: // allowed
			default:
				return w, fmt.Errorf("values: expected 6/7/11/12, got %d", len(values))
			}
			if t.Terrain, err = strconv.Atoi(values[0]); err != nil {
				return w, fmt.Errorf("value: terrainType: %w", err)
			}
			if isFirstTileRow {
				// log.Printf("todo: overriding terrain for firstTileRow\n")
				t.Terrain = 1
			} else if x == y {
				t.Terrain = 1
			} else if isXEdge {
				//log.Printf("todo: overriding terrain for x edge\n")
				//t.Terrain = 1
			} else if isYEdge {
				//log.Printf("todo: overriding terrain for y edge\n")
				//t.Terrain = 7
			}
			if t.Elevation, err = strconv.ParseFloat(values[1], 64); err != nil {
				return w, fmt.Errorf("value: elevation: %w", err)
			}
			t.IsIcy = values[2] == "1"
			t.IsGMOnly = values[3] == "1"
			if t.Resources.Animal, err = strconv.Atoi(values[4]); err != nil {
				return w, fmt.Errorf("value: animals: %w", err)
			} else if t.Resources.Animal < 0 {
				return w, fmt.Errorf("value: animals: %w", fmt.Errorf("invalid value"))
			} else if t.Resources.Animal > 100 {
				return w, fmt.Errorf("value: animals: %w", fmt.Errorf("invalid value"))
			}
			if len(values) == 6 || len(values) == 7 {
				if values[5] != "Z" {
					return w, fmt.Errorf("value: sentinel: %w", fmt.Errorf("invalid value"))
				}
			} else {
				if t.Resources.Brick, err = strconv.Atoi(values[5]); err != nil {
					return w, fmt.Errorf("value: brick: %q: %w", values, err)
				} else if t.Resources.Brick < 0 {
					return w, fmt.Errorf("value: brick: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Brick > 100 {
					return w, fmt.Errorf("value: brick: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Crops, err = strconv.Atoi(values[6]); err != nil {
					return w, fmt.Errorf("value: crops: %w", err)
				} else if t.Resources.Crops < 0 {
					return w, fmt.Errorf("value: crops: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Crops > 100 {
					return w, fmt.Errorf("value: crops: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Gems, err = strconv.Atoi(values[7]); err != nil {
					return w, fmt.Errorf("value: gems: %w", err)
				} else if t.Resources.Gems < 0 {
					return w, fmt.Errorf("value: gems: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Gems > 100 {
					return w, fmt.Errorf("value: gems: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Lumber, err = strconv.Atoi(values[8]); err != nil {
					return w, fmt.Errorf("value: lumber: %w", err)
				} else if t.Resources.Lumber < 0 {
					return w, fmt.Errorf("value: lumber: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Lumber > 100 {
					return w, fmt.Errorf("value: lumber: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Metals, err = strconv.Atoi(values[9]); err != nil {
					return w, fmt.Errorf("value: metals: %w", err)
				} else if t.Resources.Metals < 0 {
					return w, fmt.Errorf("value: metals: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Metals > 100 {
					return w, fmt.Errorf("value: metals: %w", fmt.Errorf("invalid value"))
				}
				if t.Resources.Rock, err = strconv.Atoi(values[10]); err != nil {
					return w, fmt.Errorf("value: rock: %w", err)
				} else if t.Resources.Rock < 0 {
					return w, fmt.Errorf("value: rock: %w", fmt.Errorf("invalid value"))
				} else if t.Resources.Rock > 100 {
					return w, fmt.Errorf("value: rock: %w", fmt.Errorf("invalid value"))
				}
			}
			if len(values) == 7 || len(values) == 12 {
				// split rgba
				if t.CustomBackgroundColor, err = decodeRgba(values[len(values)-1]); err != nil {
					return w, fmt.Errorf("value: rgba: %w", err)
				}
			}
		}

		w.MapKey.PositionX = m.MapKey.PositionX
		w.MapKey.PositionY = m.MapKey.PositionY
		w.MapKey.Viewlevel = m.MapKey.Viewlevel
		w.MapKey.Height = m.MapKey.Height
		if w.MapKey.BackgroundColor, err = decodeRgba(m.MapKey.BackgroundColor); err != nil {
			return w, fmt.Errorf("mapkey.backgroundcolor: %w", err)
		}
		w.MapKey.BackgroundOpacity = m.MapKey.BackgroundOpacity
		w.MapKey.TitleText = m.MapKey.TitleText
		w.MapKey.TitleFontFace = m.MapKey.TitleFontFace
		if w.MapKey.TitleFontColor, err = decodeRgba(m.MapKey.TitleFontColor); err != nil {
			return w, fmt.Errorf("mapkey.titleFontColor: %w", err)
		}
		w.MapKey.TitleFontBold = m.MapKey.TitleFontBold
		w.MapKey.TitleFontItalic = m.MapKey.TitleFontItalic
		w.MapKey.TitleScale = m.MapKey.TitleScale
		w.MapKey.ScaleText = m.MapKey.ScaleText
		w.MapKey.ScaleFontFace = m.MapKey.ScaleFontFace
		if w.MapKey.ScaleFontColor, err = decodeRgba(m.MapKey.ScaleFontColor); err != nil {
			return w, fmt.Errorf("mapkey.scaleFontColor: %w", err)
		}
		w.MapKey.ScaleFontBold = m.MapKey.ScaleFontBold
		w.MapKey.ScaleFontItalic = m.MapKey.ScaleFontItalic
		w.MapKey.ScaleScale = m.MapKey.ScaleScale
		w.MapKey.EntryFontFace = m.MapKey.EntryFontFace
		if w.MapKey.EntryFontColor, err = decodeRgba(m.MapKey.EntryFontColor); err != nil {
			return w, fmt.Errorf("mapkey.entryFontColor: %w", err)
		}
		w.MapKey.EntryFontBold = m.MapKey.EntryFontBold
		w.MapKey.EntryFontItalic = m.MapKey.EntryFontItalic
		w.MapKey.EntryScale = m.MapKey.EntryScale

		isFirstTileRow = false
	}

	for _, mFeature := range m.Features.Features {
		f := &wxx.Feature{}
		f.Type = mFeature.Type
		f.Rotate = mFeature.Rotate
		f.Uuid = mFeature.Uuid
		f.MapLayer = mFeature.MapLayer
		f.IsFlipHorizontal = mFeature.IsFlipHorizontal
		f.IsFlipVertical = mFeature.IsFlipVertical
		f.Scale = mFeature.Scale
		f.ScaleHt = mFeature.ScaleHt
		f.Tags = mFeature.Tags
		if f.Color, err = decodeRgba(mFeature.Color); err != nil {
			return w, fmt.Errorf("feature.Color: %w", err)
		}
		if f.RingColor, err = decodeRgba(mFeature.RingColor); err != nil {
			return w, fmt.Errorf("feature.RingColor: %w", err)
		}
		f.IsGMOnly = mFeature.IsGMOnly
		f.IsPlaceFreely = mFeature.IsPlaceFreely
		f.LabelPosition = mFeature.LabelPosition
		f.LabelDistance = mFeature.LabelDistance
		f.IsWorld = mFeature.IsWorld
		f.IsContinent = mFeature.IsContinent
		f.IsKingdom = mFeature.IsKingdom
		f.IsProvince = mFeature.IsProvince
		f.IsFillHexBottom = mFeature.IsFillHexBottom
		f.IsHideTerrainIcon = mFeature.IsHideTerrainIcon
		f.Location = &wxx.FeatureLocation{
			ViewLevel: mFeature.Location.ViewLevel,
			X:         mFeature.Location.X,
			Y:         mFeature.Location.Y,
		}

		f.Label = &wxx.Label{
			MapLayer:    mFeature.Label.MapLayer,
			Style:       mFeature.Label.Style,
			FontFace:    mFeature.Label.FontFace,
			OutlineSize: mFeature.Label.OutlineSize,
			Rotate:      mFeature.Label.Rotate,
			IsBold:      mFeature.Label.IsBold,
			IsItalic:    mFeature.Label.IsItalic,
			IsWorld:     mFeature.Label.IsWorld,
			IsContinent: mFeature.Label.IsContinent,
			IsKingdom:   mFeature.Label.IsKingdom,
			IsProvince:  mFeature.Label.IsProvince,
			IsGMOnly:    mFeature.Label.IsGMOnly,
			Tags:        mFeature.Label.Tags,
		}
		if f.Label.Color, err = decodeRgba(mFeature.Label.Color); err != nil {
			return w, fmt.Errorf("feature.label.color: %w", err)
		}
		if f.Label.OutlineColor, err = decodeRgba(mFeature.Label.OutlineColor); err != nil {
			return w, fmt.Errorf("feature.label.outlineColor: %w", err)
		}
		if f.Label.BackgroundColor, err = decodeRgba(mFeature.Label.BackgroundColor); err != nil {
			return w, fmt.Errorf("feature.label.backgroundColor: %w", err)
		}
		f.Label.Location = &wxx.LabelLocation{
			ViewLevel: mFeature.Label.Location.ViewLevel,
			X:         mFeature.Label.Location.X,
			Y:         mFeature.Label.Location.Y,
			Scale:     mFeature.Label.Location.Scale,
		}
		w.Features = append(w.Features, f)
	}

	for _, mLabel := range m.Labels.Labels {
		wLabel := &wxx.Label{
			MapLayer:    mLabel.MapLayer,
			Style:       mLabel.Style,
			FontFace:    mLabel.FontFace,
			OutlineSize: mLabel.OutlineSize,
			Rotate:      mLabel.Rotate,
			IsBold:      mLabel.IsBold,
			IsItalic:    mLabel.IsItalic,
			IsWorld:     mLabel.IsWorld,
			IsContinent: mLabel.IsContinent,
			IsKingdom:   mLabel.IsKingdom,
			IsProvince:  mLabel.IsProvince,
			IsGMOnly:    mLabel.IsGMOnly,
			Tags:        mLabel.Tags,
		}
		if wLabel.Color, err = decodeRgba(mLabel.Color); err != nil {
			return w, fmt.Errorf("label.color: %w", err)
		}
		if wLabel.OutlineColor, err = decodeRgba(mLabel.OutlineColor); err != nil {
			return w, fmt.Errorf("label.outlineColor: %w", err)
		}
		if mLabel.BackgroundColor == "" {
			wLabel.BackgroundColor = nil
		} else if wLabel.BackgroundColor, err = decodeZeroableRgba(mLabel.BackgroundColor); err != nil {
			return w, fmt.Errorf("label.backgroundColor: %w", err)
		}
		wLabel.Location = &wxx.LabelLocation{
			ViewLevel: mLabel.Location.ViewLevel,
			X:         mLabel.Location.X,
			Y:         mLabel.Location.Y,
			Scale:     mLabel.Location.Scale,
		}
		wLabel.InnerText = mLabel.InnerText
		w.Labels = append(w.Labels, wLabel)
	}

	for _, shape := range m.Shapes.Shapes {
		wShape := &wxx.Shape{
			BbHeight:              shape.BbHeight,
			BbIterations:          shape.BbIterations,
			BbWidth:               shape.BbWidth,
			CreationType:          shape.CreationType,
			CurrentShapeViewLevel: shape.CurrentShapeViewLevel,
			DsColor:               shape.DsColor,
			DsOffsetX:             shape.DsOffsetX,
			DsOffsetY:             shape.DsOffsetY,
			DsRadius:              shape.DsRadius,
			DsSpread:              shape.DsSpread,
			FillRule:              shape.FillRule,
			FillTexture:           shape.FillTexture,
			HighestViewLevel:      shape.HighestViewLevel,
			InsChoke:              shape.InsChoke,
			InsColor:              shape.InsColor,
			InsOffsetX:            shape.InsOffsetX,
			InsOffsetY:            shape.InsOffsetY,
			InsRadius:             shape.InsRadius,
			IsBoxBlur:             shape.IsBoxBlur,
			IsContinent:           shape.IsContinent,
			IsCurve:               shape.IsCurve,
			IsDropShadow:          shape.IsDropShadow,
			IsGMOnly:              shape.IsGMOnly,
			IsInnerShadow:         shape.IsInnerShadow,
			IsKingdom:             shape.IsKingdom,
			IsMatchTileBorders:    shape.IsMatchTileBorders,
			IsProvince:            shape.IsProvince,
			IsSnapVertices:        shape.IsSnapVertices,
			IsWorld:               shape.IsWorld,
			LineCap:               shape.LineCap,
			LineJoin:              shape.LineJoin,
			MapLayer:              shape.MapLayer,
			Opacity:               shape.Opacity,
			StrokeColor:           shape.StrokeColor,
			StrokeTexture:         shape.StrokeTexture,
			StrokeType:            shape.StrokeType,
			StrokeWidth:           shape.StrokeWidth,
			Tags:                  shape.Tags,
			Type:                  shape.Type,
		}

		for _, point := range shape.Points {
			wPoint := &wxx.Point{
				Type: point.Type,
				X:    point.X,
				Y:    point.Y,
			}
			wShape.Points = append(wShape.Points, wPoint)
		}

		w.Shapes = append(w.Shapes, wShape)
	}

	for _, note := range m.Notes.Notes {
		wNote := &wxx.Note{
			InnerText: note.InnerText,
		}
		w.Notes = append(w.Notes, wNote)
	}

	for _, info := range m.Informations.Informations {
		wInfo := &wxx.Information{
			Uuid:         info.Uuid,
			Type:         info.Type,
			Title:        info.Title,
			Rulers:       info.Rulers,
			Government:   info.Government,
			Cultures:     info.Cultures,
			Language:     info.Language,
			ReligionType: info.ReligionType,
			Culture:      info.Culture,
			HolySymbol:   info.HolySymbol,
			Domains:      info.Domains,
			InnerText:    info.InnerText,
		}

		for _, detail := range info.Details {
			wDetail := &wxx.InformationDetail{
				Uuid:         detail.Uuid,
				Type:         detail.Type,
				Title:        detail.Title,
				Rulers:       detail.Rulers,
				Government:   detail.Government,
				Cultures:     detail.Cultures,
				Language:     detail.Language,
				ReligionType: detail.ReligionType,
				Culture:      detail.Culture,
				HolySymbol:   detail.HolySymbol,
				Domains:      detail.Domains,
				InnerText:    detail.InnerText,
			}
			wInfo.Details = append(wInfo.Details, wDetail)
		}

		w.Informations.Informations = append(w.Informations.Informations, wInfo)
	}
	w.Informations.InnerText = m.Informations.InnerText

	// convert m.Configuration to w.Configuration
	for _, mTerrainConfig := range m.Configuration.TerrainConfig {
		wTerrainConfig := &wxx.TerrainConfig{
			InnerText: mTerrainConfig.InnerText,
		}
		// append the terrain configuration
		w.Configuration.TerrainConfig = append(w.Configuration.TerrainConfig, wTerrainConfig)
	}
	for _, mFeatureConfig := range m.Configuration.FeatureConfig {
		wFeatureConfig := &wxx.FeatureConfig{
			InnerText: mFeatureConfig.InnerText,
		}
		w.Configuration.FeatureConfig = append(w.Configuration.FeatureConfig, wFeatureConfig)
	}
	for _, mTextureConfig := range m.Configuration.TextureConfig {
		wTextureConfig := &wxx.TextureConfig{
			InnerText: mTextureConfig.InnerText,
		}
		w.Configuration.TextureConfig = append(w.Configuration.TextureConfig, wTextureConfig)
	}
	for _, mTextConfig := range m.Configuration.TextConfig {
		for _, mLabelStyle := range mTextConfig.LabelStyles {
			wLabelStyle := &wxx.LabelStyle{
				Name:        mLabelStyle.Name,
				FontFace:    mLabelStyle.FontFace,
				Scale:       mLabelStyle.Scale,
				IsBold:      mLabelStyle.IsBold,
				IsItalic:    mLabelStyle.IsItalic,
				OutlineSize: mLabelStyle.OutlineSize,
			}
			if wLabelStyle.Color, err = decodeRgba(mLabelStyle.Color); err != nil {
				return w, fmt.Errorf("labelStyle.color: %w", err)
			}
			if wLabelStyle.BackgroundColor, err = decodeRgba(mLabelStyle.BackgroundColor); err != nil {
				return w, fmt.Errorf("labelStyle.backgroundColor: %w", err)
			}
			if mLabelStyle.OutlineColor == "null" {
				wLabelStyle.OutlineColor = nil
			} else if wLabelStyle.OutlineColor, err = decodeZeroableRgba(mLabelStyle.OutlineColor); err != nil {
				return w, fmt.Errorf("labelStyle.outlineColor: %w", err)
			}
			w.Configuration.TextConfig.LabelStyles = append(w.Configuration.TextConfig.LabelStyles, wLabelStyle)
		}
	}
	for _, mShapeConfig := range m.Configuration.ShapeConfig {
		for _, mShapeStyle := range mShapeConfig.ShapeStyles {
			wShapeStyle := &wxx.ShapeStyle{
				Name:          mShapeStyle.Name,
				StrokeType:    mShapeStyle.StrokeType,
				IsFractal:     mShapeStyle.IsFractal,
				StrokeWidth:   mShapeStyle.StrokeWidth,
				Opacity:       mShapeStyle.Opacity,
				SnapVertices:  mShapeStyle.SnapVertices,
				Tags:          mShapeStyle.Tags,
				DropShadow:    mShapeStyle.DropShadow,
				InnerShadow:   mShapeStyle.InnerShadow,
				BoxBlur:       mShapeStyle.BoxBlur,
				DsSpread:      mShapeStyle.DsSpread,
				DsRadius:      mShapeStyle.DsRadius,
				DsOffsetX:     mShapeStyle.DsOffsetX,
				DsOffsetY:     mShapeStyle.DsOffsetY,
				InsChoke:      mShapeStyle.InsChoke,
				InsRadius:     mShapeStyle.InsRadius,
				InsOffsetX:    mShapeStyle.InsOffsetX,
				InsOffsetY:    mShapeStyle.InsOffsetY,
				BbWidth:       mShapeStyle.BbWidth,
				BbHeight:      mShapeStyle.BbHeight,
				BbIterations:  mShapeStyle.BbIterations,
				FillTexture:   mShapeStyle.FillTexture,
				StrokeTexture: mShapeStyle.StrokeTexture,
			}
			if wShapeStyle.StrokePaint, err = decodeRgba(mShapeStyle.StrokePaint); err != nil {
				return w, fmt.Errorf("shapeStyle.strokePaint: %w", err)
			}
			if wShapeStyle.FillPaint, err = decodeRgba(mShapeStyle.FillPaint); err != nil {
				return w, fmt.Errorf("shapeStyle.fillPaint: %w", err)
			}
			if wShapeStyle.DsColor, err = decodeRgba(mShapeStyle.Dscolor); err != nil {
				return w, fmt.Errorf("shapeStyle.dsColor: %w", err)
			}
			if wShapeStyle.InsColor, err = decodeRgba(mShapeStyle.InsColor); err != nil {
				return w, fmt.Errorf("shapeStyle.insColor: %w", err)
			}
			w.Configuration.ShapeConfig.ShapeStyles = append(w.Configuration.ShapeConfig.ShapeStyles, wShapeStyle)
		}
	}

	return w, nil
}
