// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

import (
	"fmt"
	"github.com/maloquacious/wxx/models/tmap173"
	"github.com/maloquacious/wxx/models/wxx"
	"sort"
	"strings"
)

func WMAPToTMAPv173(w *wxx.Map) (*tmap173.Map, error) {
	t := &tmap173.Map{
		Version: "1.73",
	}
	t.Type = w.Type                                                  // "WORLD"
	t.LastViewLevel = w.LastViewLevel                                // "WORLD"
	t.ContinentFactor = fmt.Sprintf("%d", w.ContinentFactor)         // "-1", "0"
	t.KingdomFactor = fmt.Sprintf("%d", w.KingdomFactor)             // "-1", "0"
	t.ProvinceFactor = fmt.Sprintf("%d", w.ProvinceFactor)           // "-1", "0"
	t.WorldToContinentHOffset = FToXF(w.WorldToContinentHOffset)     // "0.0"
	t.ContinentToKingdomHOffset = FToXF(w.ContinentToKingdomHOffset) // "0.0"
	t.KingdomToProvinceHOffset = FToXF(w.KingdomToProvinceHOffset)   // "0.0"
	t.WorldToContinentVOffset = FToXF(w.WorldToContinentVOffset)     // "0.0"
	t.ContinentToKingdomVOffset = FToXF(w.ContinentToKingdomVOffset) // "0.0"
	t.KingdomToProvinceVOffset = FToXF(w.KingdomToProvinceVOffset)   // "0.0"
	t.HexWidth = FToXF(w.HexWidth)                                   // "46.18", "120.97791408032022"
	t.HexHeight = FToXF(w.HexHeight)                                 // "40.0", "104.78814558711076"
	t.HexOrientation = w.HexOrientation                              // "COLUMNS"
	t.MapProjection = w.MapProjection                                // "FLAT"
	t.ShowNotes = fmt.Sprintf("%v", w.ShowNotes)                     // "true"
	t.ShowGMOnly = fmt.Sprintf("%v", w.ShowGMOnly)                   // "false"
	t.ShowGMOnlyGlow = fmt.Sprintf("%v", w.ShowGMOnlyGlow)           // "false"
	t.ShowFeatureLabels = fmt.Sprintf("%v", w.ShowFeatureLabels)     // "true"
	t.ShowGrid = fmt.Sprintf("%v", w.ShowGrid)                       // "true"
	t.ShowGridNumbers = fmt.Sprintf("%v", w.ShowGridNumbers)         // "false"
	t.ShowShadows = fmt.Sprintf("%v", w.ShowShadows)                 // "true"
	t.TriangleSize = fmt.Sprintf("%d", w.TriangleSize)               // "12"

	t.GridAndNumbering.Color0 = w.GridAndNumbering.Color0                                                  // "0x00000040"
	t.GridAndNumbering.Color1 = w.GridAndNumbering.Color1                                                  // "0x00000040"
	t.GridAndNumbering.Color2 = w.GridAndNumbering.Color2                                                  // "0x00000040"
	t.GridAndNumbering.Color3 = w.GridAndNumbering.Color3                                                  // "0x00000040"
	t.GridAndNumbering.Color4 = w.GridAndNumbering.Color4                                                  // "0x00000040"
	t.GridAndNumbering.Width0 = FToXF(w.GridAndNumbering.Width0)                                           // "1.0"
	t.GridAndNumbering.Width1 = FToXF(w.GridAndNumbering.Width1)                                           // "2.0"
	t.GridAndNumbering.Width2 = FToXF(w.GridAndNumbering.Width2)                                           // "3.0"
	t.GridAndNumbering.Width3 = FToXF(w.GridAndNumbering.Width3)                                           // "4.0"
	t.GridAndNumbering.Width4 = FToXF(w.GridAndNumbering.Width4)                                           // "1.0"
	t.GridAndNumbering.GridOffsetContinentKingdomX = FToXF(w.GridAndNumbering.GridOffsetContinentKingdomX) // "0.0"
	t.GridAndNumbering.GridOffsetContinentKingdomY = FToXF(w.GridAndNumbering.GridOffsetContinentKingdomY) // "0.0"
	t.GridAndNumbering.GridOffsetWorldContinentX = FToXF(w.GridAndNumbering.GridOffsetWorldContinentX)     // "0.0"
	t.GridAndNumbering.GridOffsetWorldContinentY = FToXF(w.GridAndNumbering.GridOffsetWorldContinentY)     // "0.0"
	t.GridAndNumbering.GridOffsetWorldKingdomX = FToXF(w.GridAndNumbering.GridOffsetWorldKingdomX)         // "0.0"
	t.GridAndNumbering.GridOffsetWorldKingdomY = FToXF(w.GridAndNumbering.GridOffsetWorldKingdomY)         // "0.0"
	t.GridAndNumbering.GridSquare = fmt.Sprintf("%d", w.GridAndNumbering.GridSquare)                       // "0"
	t.GridAndNumbering.GridSquareHeight = FToXF(w.GridAndNumbering.GridSquareHeight)                       // "-1.0"
	t.GridAndNumbering.GridSquareWidth = FToXF(w.GridAndNumbering.GridSquareWidth)                         // "-1.0"
	t.GridAndNumbering.GridOffsetX = FToXF(w.GridAndNumbering.GridOffsetX)                                 // "0.0"
	t.GridAndNumbering.GridOffsetY = FToXF(w.GridAndNumbering.GridOffsetY)                                 // "0.0"
	t.GridAndNumbering.NumberFont = w.GridAndNumbering.NumberFont                                          // "Arial"
	t.GridAndNumbering.NumberColor = w.GridAndNumbering.NumberColor                                        // "0x000000ff"
	t.GridAndNumbering.NumberSize = fmt.Sprintf("%d", w.GridAndNumbering.NumberSize)                       // "20"
	t.GridAndNumbering.NumberStyle = w.GridAndNumbering.NumberStyle                                        // "PLAIN"
	t.GridAndNumbering.NumberFirstCol = fmt.Sprintf("%d", w.GridAndNumbering.NumberFirstCol)               // "0"
	t.GridAndNumbering.NumberFirstRow = fmt.Sprintf("%d", w.GridAndNumbering.NumberFirstRow)               // "0"
	t.GridAndNumbering.NumberOrder = w.GridAndNumbering.NumberOrder                                        // "COL_ROW"
	t.GridAndNumbering.NumberPosition = w.GridAndNumbering.NumberPosition                                  // "BOTTOM"
	t.GridAndNumbering.NumberPrePad = w.GridAndNumbering.NumberPrePad                                      // "DOUBLE_ZERO"
	t.GridAndNumbering.NumberSeparator = w.GridAndNumbering.NumberSeparator                                // "."

	// terrainLookup will be used later to link terrain names to terrain indexes
	terrainLookup := make(map[int]string)
	for i, terrain := range w.TerrainMap.List {
		if i == 0 {
			t.TerrainMap = fmt.Sprintf("%s\t%d", terrain.Label, terrain.Index)
		} else {
			t.TerrainMap = fmt.Sprintf("%s\t%s\t%d", t.TerrainMap, terrain.Label, terrain.Index)
		}
		terrainLookup[terrain.Index] = terrain.Label
	}

	for _, v := range w.MapLayer {
		t.MapLayer = append(t.MapLayer, tmap173.MapLayer{Name: v.Name, IsVisible: v.IsVisible})
	}

	t.Tiles.ViewLevel = w.Tiles.ViewLevel
	t.Tiles.TilesWide = fmt.Sprintf("%d", w.Tiles.TilesWide)
	t.Tiles.TilesHigh = fmt.Sprintf("%d", w.Tiles.TilesHigh)

	for _, tileRow := range w.Tiles.TileRows {
		sb := strings.Builder{}
		for _, tile := range tileRow {
			terrainIndex := tile.Terrain
			elevation := strings.TrimSuffix(FToXF(tile.Elevation), ".0")
			isIcy := "0"
			if tile.IsIcy {
				isIcy = "1"
			}
			isGMOnly := "0"
			if tile.IsGMOnly {
				isGMOnly = "1"
			}
			animal := tile.Resources.Animal
			// compressed resources is a "Z" if all the resources (except for animal) are zero.
			compressedResources := fmt.Sprintf("%d\t%d\t%d\t%d\t%d\t%d", tile.Resources.Brick, tile.Resources.Crops, tile.Resources.Gems, tile.Resources.Lumber, tile.Resources.Metals, tile.Resources.Rock)
			if compressedResources == "0\t0\t0\t0\t0\t0" {
				compressedResources = "Z"
			}
			// the custom background color is optional
			cbc := ""
			if tile.CustomBackgroundColor != nil {
				cbc = fmt.Sprintf("\t%s,%s,%s,%s",
					FToXF(tile.CustomBackgroundColor.R),
					FToXF(tile.CustomBackgroundColor.G),
					FToXF(tile.CustomBackgroundColor.B),
					FToXF(tile.CustomBackgroundColor.A))
			}
			sb.WriteString(fmt.Sprintf("%d\t%s\t%s\t%s\t%d\t%s%s\n", terrainIndex, elevation, isIcy, isGMOnly, animal, compressedResources, cbc))
		}
		t.Tiles.TileRows = append(t.Tiles.TileRows, sb.String())
	}

	t.MapKey.PositionX = FToXF(w.MapKey.PositionX)
	t.MapKey.PositionY = FToXF(w.MapKey.PositionY)
	t.MapKey.Viewlevel = w.MapKey.Viewlevel
	t.MapKey.Height = strings.TrimSuffix(FToXF(w.MapKey.Height), ".0")
	t.MapKey.BackgroundColor = rgbaToXmlAttr(w.MapKey.BackgroundColor)
	t.MapKey.BackgroundOpacity = strings.TrimSuffix(FToXF(w.MapKey.BackgroundOpacity), ".0")
	t.MapKey.TitleText = w.MapKey.TitleText
	t.MapKey.TitleFontFace = w.MapKey.TitleFontFace
	t.MapKey.TitleFontColor = rgbaToXmlAttr(w.MapKey.TitleFontColor)
	t.MapKey.TitleFontBold = fmt.Sprintf("%v", w.MapKey.TitleFontBold)
	t.MapKey.TitleFontItalic = fmt.Sprintf("%v", w.MapKey.TitleFontItalic)
	t.MapKey.TitleScale = strings.TrimSuffix(FToXF(w.MapKey.TitleScale), ".0")
	t.MapKey.ScaleText = w.MapKey.ScaleText
	t.MapKey.ScaleFontFace = w.MapKey.ScaleFontFace
	t.MapKey.ScaleFontColor = rgbaToXmlAttr(w.MapKey.ScaleFontColor)
	t.MapKey.ScaleFontBold = fmt.Sprintf("%v", w.MapKey.ScaleFontBold)
	t.MapKey.ScaleFontItalic = fmt.Sprintf("%v", w.MapKey.ScaleFontItalic)
	t.MapKey.ScaleScale = strings.TrimSuffix(FToXF(w.MapKey.ScaleScale), ".0")
	t.MapKey.EntryFontFace = w.MapKey.EntryFontFace
	t.MapKey.EntryFontColor = rgbaToXmlAttr(w.MapKey.EntryFontColor)
	t.MapKey.EntryFontBold = fmt.Sprintf("%v", w.MapKey.EntryFontBold)
	t.MapKey.EntryFontItalic = fmt.Sprintf("%v", w.MapKey.EntryFontItalic)
	t.MapKey.EntryScale = strings.TrimSuffix(FToXF(w.MapKey.EntryScale), ".0")

	for _, feature := range w.Features {
		tf := &tmap173.Feature{
			Type:              feature.Type,
			Rotate:            FToXF(feature.Rotate),
			Uuid:              feature.Uuid,
			MapLayer:          feature.MapLayer,
			IsFlipHorizontal:  fmt.Sprintf("%v", feature.IsFlipHorizontal),
			IsFlipVertical:    fmt.Sprintf("%v", feature.IsFlipVertical),
			Scale:             FToXF(feature.Scale),
			ScaleHt:           FToXF(feature.ScaleHt),
			Tags:              feature.Tags,
			Color:             rgbaToNullableXmlAttr(feature.Color),
			RingColor:         rgbaToNullableXmlAttr(feature.RingColor),
			IsGMOnly:          fmt.Sprintf("%v", feature.IsGMOnly),
			IsPlaceFreely:     fmt.Sprintf("%v", feature.IsPlaceFreely),
			LabelPosition:     feature.LabelPosition,
			LabelDistance:     strings.TrimSuffix(FToXF(feature.LabelDistance), ".0"),
			IsWorld:           fmt.Sprintf("%v", feature.IsWorld),
			IsContinent:       fmt.Sprintf("%v", feature.IsContinent),
			IsKingdom:         fmt.Sprintf("%v", feature.IsKingdom),
			IsProvince:        fmt.Sprintf("%v", feature.IsProvince),
			IsFillHexBottom:   fmt.Sprintf("%v", feature.IsFillHexBottom),
			IsHideTerrainIcon: fmt.Sprintf("%v", feature.IsHideTerrainIcon),
		}
		if feature.Location != nil {
			tf.Location = &tmap173.FeatureLocation{
				ViewLevel: feature.Location.ViewLevel,
				X:         FToXF(feature.Location.X),
				Y:         FToXF(feature.Location.Y),
			}
		}
		if feature.Label != nil {
			tf.Label = &tmap173.Label{
				MapLayer:        feature.Label.MapLayer,
				Style:           feature.Label.Style,
				FontFace:        feature.Label.FontFace,
				Color:           rgbaToXmlAttr(feature.Label.Color),
				OutlineColor:    rgbaToXmlAttr(feature.Label.OutlineColor),
				OutlineSize:     FToXF(feature.Label.OutlineSize),
				Rotate:          FToXF(feature.Label.Rotate),
				IsBold:          fmt.Sprintf("%v", feature.Label.IsBold),
				IsItalic:        fmt.Sprintf("%v", feature.Label.IsItalic),
				IsWorld:         fmt.Sprintf("%v", feature.Label.IsWorld),
				IsContinent:     fmt.Sprintf("%v", feature.Label.IsContinent),
				IsKingdom:       fmt.Sprintf("%v", feature.Label.IsKingdom),
				IsProvince:      fmt.Sprintf("%v", feature.Label.IsProvince),
				IsGMOnly:        fmt.Sprintf("%v", feature.Label.IsGMOnly),
				Tags:            feature.Label.Tags,
				BackgroundColor: rgbaToXmlAttr(feature.Label.BackgroundColor),
			}
			if feature.Label.Location != nil {
				tf.Label.Location = &tmap173.LabelLocation{
					ViewLevel: feature.Label.Location.ViewLevel,
					X:         FToXF(feature.Label.Location.X),
					Y:         FToXF(feature.Label.Location.Y),
					Scale:     FToXF(feature.Label.Location.Scale),
				}
			}
		}
		t.Features = append(t.Features, tf)
	}

	for _, wLabel := range w.Labels {
		tLabel := &tmap173.Label{
			MapLayer:     wLabel.MapLayer,
			Style:        wLabel.Style,
			FontFace:     wLabel.FontFace,
			Color:        rgbaToXmlAttr(wLabel.Color),
			OutlineColor: rgbaToXmlAttr(wLabel.OutlineColor),
			OutlineSize:  FToXF(wLabel.OutlineSize),
			Rotate:       FToXF(wLabel.Rotate),
			IsBold:       fmt.Sprintf("%v", wLabel.IsBold),
			IsItalic:     fmt.Sprintf("%v", wLabel.IsItalic),
			IsWorld:      fmt.Sprintf("%v", wLabel.IsWorld),
			IsContinent:  fmt.Sprintf("%v", wLabel.IsContinent),
			IsKingdom:    fmt.Sprintf("%v", wLabel.IsKingdom),
			IsProvince:   fmt.Sprintf("%v", wLabel.IsProvince),
			IsGMOnly:     fmt.Sprintf("%v", wLabel.IsGMOnly),
			Tags:         wLabel.Tags,
		}
		if wLabel.BackgroundColor != nil {
			tLabel.BackgroundColor = rgbaToXmlAttr(wLabel.BackgroundColor)
		}
		if wLabel.Location != nil {
			tLabel.Location = &tmap173.LabelLocation{
				ViewLevel: wLabel.Location.ViewLevel,
				X:         FToXF(wLabel.Location.X),
				Y:         FToXF(wLabel.Location.Y),
				Scale:     FToXF(wLabel.Location.Scale),
			}
		}
		tLabel.InnerText = strings.ReplaceAll(wLabel.InnerText, "\n", "&#10;")
		t.Labels = append(t.Labels, tLabel)
	}

	for _, wShape := range w.Shapes {
		tShape := &tmap173.Shape{
			BbHeight:              FToXF(wShape.BbHeight),
			BbIterations:          fmt.Sprintf("%d", wShape.BbIterations),
			BbWidth:               FToXF(wShape.BbWidth),
			CreationType:          wShape.CreationType,
			CurrentShapeViewLevel: wShape.CurrentShapeViewLevel,
			DsColor:               wShape.DsColor,
			DsOffsetX:             FToXF(wShape.DsOffsetX),
			DsOffsetY:             FToXF(wShape.DsOffsetY),
			DsRadius:              FToXF(wShape.DsRadius),
			DsSpread:              FToXF(wShape.DsSpread),
			FillRule:              wShape.FillRule,
			FillTexture:           wShape.FillTexture,
			HighestViewLevel:      wShape.HighestViewLevel,
			InsChoke:              FToXF(wShape.InsChoke),
			InsColor:              wShape.InsColor,
			InsOffsetX:            FToXF(wShape.InsOffsetX),
			InsOffsetY:            FToXF(wShape.InsOffsetY),
			InsRadius:             FToXF(wShape.InsRadius),
			IsBoxBlur:             fmt.Sprintf("%v", wShape.IsBoxBlur),
			IsContinent:           fmt.Sprintf("%v", wShape.IsContinent),
			IsCurve:               fmt.Sprintf("%v", wShape.IsCurve),
			IsDropShadow:          fmt.Sprintf("%v", wShape.IsDropShadow),
			IsGMOnly:              fmt.Sprintf("%v", wShape.IsGMOnly),
			IsInnerShadow:         fmt.Sprintf("%v", wShape.IsInnerShadow),
			IsKingdom:             fmt.Sprintf("%v", wShape.IsKingdom),
			IsMatchTileBorders:    fmt.Sprintf("%v", wShape.IsMatchTileBorders),
			IsProvince:            fmt.Sprintf("%v", wShape.IsProvince),
			IsSnapVertices:        fmt.Sprintf("%v", wShape.IsSnapVertices),
			IsWorld:               fmt.Sprintf("%v", wShape.IsWorld),
			LineCap:               wShape.LineCap,
			LineJoin:              wShape.LineJoin,
			MapLayer:              wShape.MapLayer,
			Opacity:               FToXF(wShape.Opacity),
			StrokeColor:           wShape.StrokeColor,
			StrokeTexture:         wShape.StrokeTexture,
			StrokeType:            wShape.StrokeType,
			StrokeWidth:           FToXF(wShape.StrokeWidth),
			Tags:                  wShape.Tags,
			Type:                  wShape.Type,
		}
		for _, wPoint := range wShape.Points {
			tShape.Points = append(tShape.Points, &tmap173.Point{
				Type: wPoint.Type,
				X:    FToXF(wPoint.X),
				Y:    FToXF(wPoint.Y),
			})
		}
		t.Shapes = append(t.Shapes, tShape)
	}

	for _, wNote := range w.Notes {
		tNote := &tmap173.Note{
			InnerText: wNote.InnerText,
		}
		t.Notes = append(t.Notes, tNote)
	}

	for _, wInformation := range w.Informations.Informations {
		tInformation := &tmap173.Information{
			Uuid:         wInformation.Uuid,
			Type:         wInformation.Type,
			Title:        wInformation.Title,
			Rulers:       wInformation.Rulers,
			Government:   wInformation.Government,
			Cultures:     wInformation.Cultures,
			Language:     wInformation.Language,
			ReligionType: wInformation.ReligionType,
			Culture:      wInformation.Culture,
			HolySymbol:   wInformation.HolySymbol,
			Domains:      wInformation.Domains,
			InnerText:    strings.TrimSpace(wInformation.InnerText),
		}
		for _, wDetail := range wInformation.Details {
			tDetail := &tmap173.InformationDetail{
				Uuid:         wDetail.Uuid,
				Type:         wDetail.Type,
				Title:        wDetail.Title,
				Rulers:       wDetail.Rulers,
				Government:   wDetail.Government,
				Cultures:     wDetail.Cultures,
				Language:     wDetail.Language,
				ReligionType: wDetail.ReligionType,
				Culture:      wDetail.Culture,
				HolySymbol:   wDetail.HolySymbol,
				Domains:      wDetail.Domains,
				InnerText:    strings.TrimSpace(wDetail.InnerText),
			}
			tInformation.Details = append(tInformation.Details, tDetail)
		}
		t.Information = append(t.Information, tInformation)
	}
	t.InformationInnerText = w.Informations.InnerText

	// copy over configuration
	// copy over configuration.terrain-config
	// copy over configuration.feature-config
	// copy over configuration.texture-config
	// copy over configuration.text-config
	for _, wLabelStyle := range w.Configuration.TextConfig.LabelStyles {
		tLabelStyle := &tmap173.LabelStyle{
			Name:            wLabelStyle.Name,
			FontFace:        wLabelStyle.FontFace,
			Scale:           FToXF(wLabelStyle.Scale),
			IsBold:          fmt.Sprintf("%v", wLabelStyle.IsBold),
			IsItalic:        fmt.Sprintf("%v", wLabelStyle.IsItalic),
			Color:           rgbaToXmlAttr(wLabelStyle.Color),
			BackgroundColor: rgbaToNullableXmlAttr(wLabelStyle.BackgroundColor),
			OutlineSize:     FToXF(wLabelStyle.OutlineSize),
		}
		if wLabelStyle.OutlineColor == nil {
			tLabelStyle.OutlineColor = "null"
		} else {
			tLabelStyle.OutlineColor = rgbaToXmlAttr(wLabelStyle.OutlineColor)
			//if tLabelStyle.OutlineSize == "null" {
			//	tLabelStyle.OutlineColor = "0.0,0.0,0.0,1.0"
			//}
		}
		t.Configuration.TextConfig.LabelStyles = append(t.Configuration.TextConfig.LabelStyles, tLabelStyle)
	}
	t.Configuration.TextConfig.InnerText = w.Configuration.TextConfig.InnerText
	// copy over configuration.shape-config
	for _, wShapeStyle := range w.Configuration.ShapeConfig.ShapeStyles {
		tShapeStyle := &tmap173.ShapeStyle{
			Name:          wShapeStyle.Name,
			StrokeType:    wShapeStyle.StrokeType,
			IsFractal:     fmt.Sprintf("%v", wShapeStyle.IsFractal),
			StrokeWidth:   FToXF(wShapeStyle.StrokeWidth),
			Opacity:       FToXF(wShapeStyle.Opacity),
			SnapVertices:  fmt.Sprintf("%v", wShapeStyle.SnapVertices),
			Tags:          wShapeStyle.Tags,
			DropShadow:    fmt.Sprintf("%v", wShapeStyle.DropShadow),
			InnerShadow:   fmt.Sprintf("%v", wShapeStyle.InnerShadow),
			BoxBlur:       fmt.Sprintf("%v", wShapeStyle.BoxBlur),
			DsSpread:      FToXF(wShapeStyle.DsSpread),
			DsRadius:      FToXF(wShapeStyle.DsRadius),
			DsOffsetX:     FToXF(wShapeStyle.DsOffsetX),
			DsOffsetY:     FToXF(wShapeStyle.DsOffsetY),
			InsChoke:      FToXF(wShapeStyle.InsChoke),
			InsRadius:     FToXF(wShapeStyle.InsRadius),
			InsOffsetX:    FToXF(wShapeStyle.InsOffsetX),
			InsOffsetY:    FToXF(wShapeStyle.InsOffsetY),
			BbWidth:       FToXF(wShapeStyle.BbWidth),
			BbHeight:      FToXF(wShapeStyle.BbHeight),
			BbIterations:  fmt.Sprintf("%d", wShapeStyle.BbIterations),
			FillTexture:   wShapeStyle.FillTexture,
			StrokeTexture: wShapeStyle.StrokeTexture,
			StrokePaint:   rgbaToXmlAttr(wShapeStyle.StrokePaint),
			FillPaint:     rgbaToNullableXmlAttr(wShapeStyle.FillPaint),
			DsColor:       rgbaToNullableXmlAttr(wShapeStyle.DsColor),
			InsColor:      rgbaToNullableXmlAttr(wShapeStyle.InsColor),
		}
		t.Configuration.ShapeConfig.ShapeStyles = append(t.Configuration.ShapeConfig.ShapeStyles, tShapeStyle)
	}
	t.Configuration.ShapeConfig.InnerText = w.Configuration.ShapeConfig.InnerText
	t.Configuration.InnerText = w.Configuration.InnerText

	return t, nil
}

// sortTerrainMap takes a map with string keys and int values, representing terrain data,
// and returns a slice of the map's keys sorted in lexicographical (alphabetical) order.
//
// Parameters:
// - tm : A map with string keys and integer values.
//
// Returns:
// - A slice of string, containing the keys from the input map sorted in lexicographical order.
//
// Note: This function makes no assumptions about the contents of the map,
// so it can be used to sort any map[string]int.
func sortTerrainMap(tm map[string]int) []string {
	var o []string
	for k := range tm {
		o = append(o, k)
	}
	sort.Strings(o)
	return o
}
