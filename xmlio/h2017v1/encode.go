// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2017v1

import (
	"bytes"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/maloquacious/wxx"
)

// Encode the Map_t into a slice of UTF-8 bytes that matches this version's XML schema.
//
// Note: the style of this code is intentionally verbose to make it easier to find changes between
// versions of the Worldographer files.
func Encode(w *wxx.Map_t) ([]byte, error) {
	wb := &bytes.Buffer{}
	if err := encodeMap(w, wb); err != nil {
		return nil, err
	}
	return wb.Bytes(), nil
}

func encodeMap(w *wxx.Map_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<map"))
	wb.WriteString(fmt.Sprintf(" type=%q", w.Type))
	wb.WriteString(fmt.Sprintf(" version=%q", w.Version))
	wb.WriteString(fmt.Sprintf(" lastViewLevel=%q", w.LastViewLevel))
	wb.WriteString(fmt.Sprintf(" continentFactor=%q", ints(w.ContinentFactor)))
	wb.WriteString(fmt.Sprintf(" kingdomFactor=%q", ints(w.KingdomFactor)))
	wb.WriteString(fmt.Sprintf(" provinceFactor=%q", ints(w.ProvinceFactor)))
	wb.WriteString(fmt.Sprintf(" worldToContinentHOffset=%q", floats(w.WorldToContinentHOffset)))
	wb.WriteString(fmt.Sprintf(" continentToKingdomHOffset=%q", floats(w.ContinentToKingdomHOffset)))
	wb.WriteString(fmt.Sprintf(" kingdomToProvinceHOffset=%q", floats(w.KingdomToProvinceHOffset)))
	wb.WriteString(fmt.Sprintf(" worldToContinentVOffset=%q", floats(w.WorldToContinentVOffset)))
	wb.WriteString(fmt.Sprintf(" continentToKingdomVOffset=%q", floats(w.ContinentToKingdomVOffset)))
	wb.WriteString(fmt.Sprintf(" kingdomToProvinceVOffset=%q \n", floats(w.KingdomToProvinceVOffset)))
	wb.WriteString(fmt.Sprintf("hexWidth=%q", floats(w.HexWidth)))
	wb.WriteString(fmt.Sprintf(" hexHeight=%q", floats(w.HexHeight)))
	wb.WriteString(fmt.Sprintf(" hexOrientation=%q", w.HexOrientation))
	if w.MapProjection == wxx.FLAT {
		wb.WriteString(fmt.Sprintf(" mapProjection=%q", "FLAT"))
	} else if w.MapProjection == wxx.ICOSAHEDRAL {
		wb.WriteString(fmt.Sprintf(" mapProjection=%q", "ICOSAHEDRAL"))
	} else {
		return fmt.Errorf("assert(map.projection != %q)", w.MapProjection)
	}
	wb.WriteString(fmt.Sprintf(" showNotes=%q", bools(w.ShowNotes)))
	wb.WriteString(fmt.Sprintf(" showGMOnly=%q", bools(w.ShowGMOnly)))
	wb.WriteString(fmt.Sprintf(" showGMOnlyGlow=%q", bools(w.ShowGMOnlyGlow)))
	wb.WriteString(fmt.Sprintf(" showFeatureLabels=%q", bools(w.ShowFeatureLabels)))
	wb.WriteString(fmt.Sprintf(" showGrid=%q", bools(w.ShowGrid)))
	wb.WriteString(fmt.Sprintf(" showGridNumbers=%q", bools(w.ShowGridNumbers)))
	wb.WriteString(fmt.Sprintf(" showShadows=%q", bools(w.ShowShadows)))
	wb.WriteString(fmt.Sprintf("  triangleSize=%q", ints(w.TriangleSize)))
	wb.WriteString(fmt.Sprintf(">\n"))

	if err := encodeGridAndNumbering(w.GridAndNumbering, wb); err != nil {
		return err
	}

	if err := encodeTerrainMap(w.TerrainMap, wb); err != nil {
		return err
	}

	if err := encodeMapLayers(w.MapLayers, wb); err != nil {
		return err
	}

	if err := encodeTiles(w.Tiles, w.HexOrientation, wb); err != nil {
		return err
	}

	if err := encodeMapKey(w.MapKey, wb); err != nil {
		return err
	}

	if err := encodeFeatures(w.Features, wb); err != nil {
		return err
	}

	if err := encodeLabels(w.Labels, wb); err != nil {
		return err
	}

	if err := encodeShapes(w.Shapes, wb); err != nil {
		return err
	}

	if err := encodeNotes(w.Notes, wb); err != nil {
		return err
	}

	if err := encodeInformations(w.Informations, wb); err != nil {
		return err
	}

	if err := encodeConfiguration(w.Configuration, wb); err != nil {
		return err
	}

	wb.WriteString("</map>\n")

	return nil
}

func encodeGridAndNumbering(gridAndNumbering *wxx.GridAndNumbering_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf(`<gridandnumbering`))
	wb.WriteString(fmt.Sprintf(" color0=%q", gridAndNumbering.Color0))
	wb.WriteString(fmt.Sprintf(" color1=%q", gridAndNumbering.Color1))
	wb.WriteString(fmt.Sprintf(" color2=%q", gridAndNumbering.Color2))
	wb.WriteString(fmt.Sprintf(" color3=%q", gridAndNumbering.Color3))
	wb.WriteString(fmt.Sprintf(" color4=%q", gridAndNumbering.Color4))
	wb.WriteString(fmt.Sprintf(" width0=%q", floats(gridAndNumbering.Width0)))
	wb.WriteString(fmt.Sprintf(" width1=%q", floats(gridAndNumbering.Width1)))
	wb.WriteString(fmt.Sprintf(" width2=%q", floats(gridAndNumbering.Width2)))
	wb.WriteString(fmt.Sprintf(" width3=%q", floats(gridAndNumbering.Width3)))
	wb.WriteString(fmt.Sprintf(" width4=%q", floats(gridAndNumbering.Width4)))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomX=%q", floats(gridAndNumbering.GridOffsetContinentKingdomX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomY=%q", floats(gridAndNumbering.GridOffsetContinentKingdomY)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentX=%q", floats(gridAndNumbering.GridOffsetWorldContinentX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentY=%q", floats(gridAndNumbering.GridOffsetWorldContinentY)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomX=%q", floats(gridAndNumbering.GridOffsetWorldKingdomX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomY=%q", floats(gridAndNumbering.GridOffsetWorldKingdomY)))
	wb.WriteString(fmt.Sprintf(" gridSquare=%q", ints(gridAndNumbering.GridSquare)))
	wb.WriteString(fmt.Sprintf(" gridSquareHeight=%q", floats(gridAndNumbering.GridSquareHeight)))
	wb.WriteString(fmt.Sprintf(" gridSquareWidth=%q", floats(gridAndNumbering.GridSquareWidth)))
	wb.WriteString(fmt.Sprintf(" gridOffsetX=%q", floats(gridAndNumbering.GridOffsetX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetY=%q", floats(gridAndNumbering.GridOffsetY)))
	wb.WriteString(fmt.Sprintf(" numberFont=%q", gridAndNumbering.NumberFont))
	wb.WriteString(fmt.Sprintf(" numberColor=%q", gridAndNumbering.NumberColor))
	wb.WriteString(fmt.Sprintf(" numberSize=%q", ints(gridAndNumbering.NumberSize)))
	wb.WriteString(fmt.Sprintf(" numberStyle=%q", gridAndNumbering.NumberStyle))
	wb.WriteString(fmt.Sprintf(" numberFirstCol=%q", ints(gridAndNumbering.NumberFirstCol)))
	wb.WriteString(fmt.Sprintf(" numberFirstRow=%q", ints(gridAndNumbering.NumberFirstRow)))
	wb.WriteString(fmt.Sprintf(" numberOrder=%q", gridAndNumbering.NumberOrder))
	wb.WriteString(fmt.Sprintf(" numberPosition=%q", gridAndNumbering.NumberPosition))
	wb.WriteString(fmt.Sprintf(" numberPrePad=%q", gridAndNumbering.NumberPrePad))
	wb.WriteString(fmt.Sprintf(" numberSeparator=%q", gridAndNumbering.NumberSeparator))
	wb.WriteString(fmt.Sprintf(" />\n"))
	return nil
}

func encodeTerrainMap(terrainMap *wxx.TerrainMap_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<terrainmap>"))
	for k, v := range terrainMapToSlice(terrainMap.Data) {
		if k == 0 {
			wb.WriteString(fmt.Sprintf("%s\t%d", v, k))
		} else {
			wb.WriteString(fmt.Sprintf("\t%s\t%d", v, k))
		}
	}
	wb.WriteString(fmt.Sprintf("</terrainmap>\n"))
	return nil
}

// order of layers is important; worldographer renders them from the bottom up.
func encodeMapLayers(mapLayers []*wxx.MapLayer_t, wb *bytes.Buffer) error {
	for _, mapLayer := range mapLayers {
		if err := encodeMapLayer(mapLayer, wb); err != nil {
			return err
		}
	}
	return nil
}

func encodeMapLayer(mapLayer *wxx.MapLayer_t, wb *bytes.Buffer) error {
	wb.WriteString("<maplayer")
	wb.WriteString(fmt.Sprintf(" name=%q", mapLayer.Name))
	wb.WriteString(fmt.Sprintf(" isVisible=%q", bools(mapLayer.IsVisible)))
	wb.WriteString("/>\n")
	return nil
}

func encodeTiles(tiles *wxx.Tiles_t, hexOrientation string, wb *bytes.Buffer) error {
	// to: width is the number of columns, height is the number of rows. does that depend on the orientation?
	wb.WriteString(fmt.Sprintf("<tiles"))
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", tiles.ViewLevel))
	wb.WriteString(fmt.Sprintf(" tilesWide=%q", ints(tiles.TilesWide)))
	wb.WriteString(fmt.Sprintf(" tilesHigh=%q", ints(tiles.TilesHigh)))
	wb.WriteString(fmt.Sprintf(">\n"))

	// generate the tile-row elements:
	// * each tilerow will have a tile.tilesHigh lines of tab delimited data
	// * each line of data has the following values: Terrain type, elevation, is it icy, is it GM only, and its resources
	// * terrainType is an index into the terrainmap element
	// * resources are Animals, Brick, Crops, Gems, Lumber, Metals, and Rock, in that order, but are "compressed"
	if hexOrientation == "COLUMNS" {
		for x := 0; x < tiles.TilesWide; x++ {
			wb.WriteString("<tilerow>\n")
			for y := 0; y < tiles.TilesHigh; y++ {
				tile := tiles.TileRows[x][y]
				if err := encodeTile(tile, wb); err != nil {
					return err
				}
			}
			wb.WriteString(fmt.Sprintf("</tilerow>\n"))
		}
	} else if hexOrientation == "ROWS" {
		return fmt.Errorf("assert(orientation != %q)", hexOrientation)
	} else {
		return fmt.Errorf("assert(orientation != %q)", hexOrientation)
	}
	wb.WriteString(fmt.Sprintf("</tiles>\n"))
	return nil
}

// some documentation is only in this discord chat - https://discord.com/channels/535205750532997160/877285895991095369/1187771984768151653
// summarizing that:
// * tilerow is tab-delimited data that looks like terrainMapSlot elevation isIcy isGMOnly animals 0 0 0 0 0 0
// * the web page has isIcy as a float, but it seems to be a boolean
// * resource.animals is int with range 0...100
// * field after resource.animal is "Z" if remaining resources are all 0
// * otherwise we have brick, crops, gems, lumber, metals, rock
// * customBackgroundColor is an RGBA that is optional
func encodeTile(tile *wxx.Tile_t, wb *bytes.Buffer) error {
	// todo: implement this
	wb.WriteString(fmt.Sprintf("%d", tile.Terrain))
	wb.WriteString(fmt.Sprintf("\t%d", floatd(tile.Elevation)))
	wb.WriteString(fmt.Sprintf("\t%d", boold(tile.IsIcy)))
	wb.WriteString(fmt.Sprintf("\t%d", boold(tile.IsGMOnly)))
	if err := encodeTileResources(tile.Resources, wb); err != nil {
		return err
	}
	if tile.CustomBackgroundColor != nil {
		wb.WriteString(fmt.Sprintf("\t%s", rgbas(tile.CustomBackgroundColor)))
	}
	wb.WriteString(fmt.Sprintf("\n"))
	return nil
}

// all resources are supposed to be in the range of 0...100, but we don't enforce
func encodeTileResources(resources wxx.Resources_t, wb *bytes.Buffer) error {
	// compress if there are no resources other than Animal
	if resources.Brick == 0 && resources.Crops == 0 && resources.Gems == 0 && resources.Lumber == 0 && resources.Metals == 0 && resources.Rock == 0 {
		wb.WriteString(fmt.Sprintf("\t%d", resources.Animal))
		wb.WriteString(fmt.Sprintf("\tZ"))
		return nil
	}
	wb.WriteString(fmt.Sprintf("\t%d", resources.Animal))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Brick))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Crops))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Gems))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Lumber))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Metals))
	wb.WriteString(fmt.Sprintf("\t%d", resources.Rock))
	return nil
}

func encodeMapKey(mapKey *wxx.MapKey_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf(`<mapkey positionx="0.0" positiony="0.0" viewlevel="WORLD" height="-1" backgroundcolor="0.9803921580314636,0.9215686321258545,0.843137264251709,1.0" backgroundopacity="50" titleText="Map Key" titleFontFace="Arial"  titleFontColor="0.0,0.0,0.0,1.0" titleFontBold="true" titleFontItalic="false" titleScale="80" scaleText="1 Hex = ? units" scaleFontFace="Arial"  scaleFontColor="0.0,0.0,0.0,1.0" scaleFontBold="true" scaleFontItalic="false" scaleScale="65" entryFontFace="Arial"  entryFontColor="0.0,0.0,0.0,1.0" entryFontBold="true" entryFontItalic="false" entryScale="55"  >`))
	wb.WriteByte('\n')
	wb.WriteString("</mapkey>\n")
	return nil
}

// add features
func encodeFeatures(features []*wxx.Feature_t, wb *bytes.Buffer) error {
	wb.WriteString("<features>\n")
	for _, feature := range features {
		if err := encodeFeature(feature, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</features>\n")
	return nil
}

func encodeFeature(feature *wxx.Feature_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<feature"))
	wb.WriteString(fmt.Sprintf(" type=%q", feature.Type))
	wb.WriteString(fmt.Sprintf(" rotate=%q", floats(feature.Rotate)))
	wb.WriteString(fmt.Sprintf(" uuid=%q", feature.Uuid))
	wb.WriteString(fmt.Sprintf(" mapLayer=%q", feature.MapLayer))
	wb.WriteString(fmt.Sprintf(" isFlipHorizontal=%q", bools(feature.IsFlipHorizontal)))
	wb.WriteString(fmt.Sprintf(" isFlipVertical=%q", bools(feature.IsFlipVertical)))
	wb.WriteString(fmt.Sprintf(" scale=%q", floats(feature.Scale)))
	wb.WriteString(fmt.Sprintf(" scaleHt=%q", floats(feature.ScaleHt)))
	wb.WriteString(fmt.Sprintf(" tags=%q", feature.Tags))
	wb.WriteString(fmt.Sprintf(" color=%q", rgbans(feature.Color))) // nullable
	wb.WriteString(fmt.Sprintf(" ringcolor=%q", rgbans(feature.RingColor)))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(feature.IsGMOnly)))
	wb.WriteString(fmt.Sprintf(" isPlaceFreely=%q", bools(feature.IsPlaceFreely)))
	wb.WriteString(fmt.Sprintf(" labelPosition=%q", feature.LabelPosition))
	wb.WriteString(fmt.Sprintf(" labelDistance=%q", ints(feature.LabelDistance)))
	wb.WriteString(fmt.Sprintf(" isWorld=%q", bools(feature.IsWorld)))
	wb.WriteString(fmt.Sprintf(" isContinent=%q", bools(feature.IsContinent)))
	wb.WriteString(fmt.Sprintf(" isKingdom=%q", bools(feature.IsKingdom)))
	wb.WriteString(fmt.Sprintf(" isProvince=%q", bools(feature.IsProvince)))
	wb.WriteString(fmt.Sprintf(" isFillHexBottom=%q", bools(feature.IsFillHexBottom)))
	wb.WriteString(fmt.Sprintf(" isHideTerrainIcon=%q", bools(feature.IsHideTerrainIcon)))
	wb.WriteString(">")
	if feature.Location != nil {
		if err := encodeFeatureLocation(feature.Location, wb); err != nil {
			return err
		}
	}
	if feature.Label != nil {
		if err := encodeFeatureLabel(feature.Label, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</feature>\n")
	return nil
}

func encodeFeatureLocation(location *wxx.FeatureLocation_t, wb *bytes.Buffer) error {
	wb.WriteString("<location")
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", location.ViewLevel))
	wb.WriteString(fmt.Sprintf(" x=%q", floats(location.X)))
	wb.WriteString(fmt.Sprintf(" y=%q", floats(location.Y)))
	wb.WriteString(" />")
	return nil
}

func encodeFeatureLabel(label *wxx.Label_t, wb *bytes.Buffer) error {
	return encodeLabel(label, wb)
}

func encodeLabels(labels []*wxx.Label_t, wb *bytes.Buffer) error {
	wb.WriteString("<labels>\n")
	for _, label := range labels {
		if err := encodeLabel(label, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</labels>\n")
	return nil
}

func encodeLabel(label *wxx.Label_t, wb *bytes.Buffer) error {
	wb.WriteString("<label")
	wb.WriteString(fmt.Sprintf("  mapLayer=%q", label.MapLayer))
	wb.WriteString(fmt.Sprintf(" style=%q", label.Style))       // can be null!
	wb.WriteString(fmt.Sprintf(" fontFace=%q", label.FontFace)) // can be null!
	wb.WriteString(fmt.Sprintf(" color=%q", rgbas(label.Color)))
	// todo: backgroundColor is sometimes not displayed when its value is "0.0,0.0,0.0,1.0".
	// I may need to ask on the Inkwell Discord about this; I can't figure out the pattern.
	// Until then, seems to be no harm in excluding it (other than noise in the diff).
	if attr := rgbas(label.BackgroundColor); attr != "0.0,0.0,0.0,1.0" { // do not include if null
		wb.WriteString(fmt.Sprintf(" backgroundColor=%q", attr))
	}
	wb.WriteString(fmt.Sprintf(" outlineColor=%q", rgbas(label.OutlineColor)))
	wb.WriteString(fmt.Sprintf(" outlineSize=%q", floats(label.OutlineSize)))
	wb.WriteString(fmt.Sprintf(" rotate=%q", floats(label.Rotate)))
	wb.WriteString(fmt.Sprintf(" isBold=%q", bools(label.IsBold)))
	wb.WriteString(fmt.Sprintf(" isItalic=%q", bools(label.IsItalic)))
	wb.WriteString(fmt.Sprintf(" isWorld=%q", bools(label.IsWorld)))
	wb.WriteString(fmt.Sprintf(" isContinent=%q", bools(label.IsContinent)))
	wb.WriteString(fmt.Sprintf(" isKingdom=%q", bools(label.IsKingdom)))
	wb.WriteString(fmt.Sprintf(" isProvince=%q", bools(label.IsProvince)))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(label.IsGMOnly)))
	wb.WriteString(fmt.Sprintf(" tags=%q", label.Tags))
	wb.WriteString(">")
	if err := encodeLabelLocation(label.Location, wb); err != nil {
		return err
	}
	if label.InnerText != "" {
		wb.WriteString(encodeInnerText(label.InnerText))
	}
	wb.WriteString("</label>\n")
	return nil
}

func encodeLabelLocation(location *wxx.LabelLocation_t, wb *bytes.Buffer) error {
	wb.WriteString("<location")
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", location.ViewLevel))
	wb.WriteString(fmt.Sprintf(" x=%q", floats(location.X)))
	wb.WriteString(fmt.Sprintf(" y=%q", floats(location.Y)))
	wb.WriteString(fmt.Sprintf(" scale=%q", floats(location.Scale)))
	wb.WriteString(" />")
	return nil
}

func encodeShapes(shapes []*wxx.Shape_t, wb *bytes.Buffer) error {
	wb.WriteString("<shapes>\n")
	for _, shape := range shapes {
		if err := encodeShape(shape, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</shapes>\n")
	return nil
}

func encodeShape(shape *wxx.Shape_t, wb *bytes.Buffer) error {
	//<shape
	//    type="Arc"
	//    creationType="BASIC"
	//    isWorld="true" isContinent="true" isKingdom="true" isProvince="true"
	//    dsOffsetX="0.0" dsOffsetY="0.0" dsRadius="50.0" dsSpread="0.2" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0"
	//    insOffsetX="0.0" insOffsetY="0.0" insRadius="50.0" insChoke="0.2" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0"
	//    bbWidth="10.0" bbHeight="10.0" bbIterations="3"
	//    mapLayer="Above Terrain"
	//    strokeType="SIMPLE"
	//    highestViewLevel="WORLD"
	//    currentShapeViewLevel="WORLD"
	//    lineCap="SQUARE" lineJoin="ROUND"
	//    opacity="1.0"
	//    strokeColor="1.0,0.0,0.0,1.0" strokeWidth="0.05"
	//    length="360.0"
	//    startAngle="0.0"
	//    arcType="OPEN" >
	// <p x="148" y="149"/>
	// <p x="267" y="286"/>
	//</shape>
	return nil
}

func encodeNotes(notes []*wxx.Note_t, wb *bytes.Buffer) error {
	wb.WriteString("<notes>\n")
	for _, note := range notes {
		if err := encodeNote(note, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</notes>\n")
	return nil
}

func encodeNote(note *wxx.Note_t, wb *bytes.Buffer) error {
	///*
	//	<note key="WORLD,2343.75,3112.5" viewLevel="WORLD" x="2343.75" y="3112.5" filename="" parent="dde12f75-dcc9-4cb7-a96d-f18011601143" color="1.0,1.0,0.0,1.0" title="Units (Notes Title)">
	//	<notetext><![CDATA[<html dir="ltr"><head></head><body contenteditable="true">Paragraph (Notes Paragraph)</body></html>]]></notetext></note>
	//*/
	//	printf(`<note key="WORLD,%f,%f" viewLevel="WORLD" x="%f" y="%f" filename="" parent=%q color="1.0,1.0,0.0,1.0" title=%q>`, note.Origin.X, note.Origin.Y, note.Origin.X, note.Origin.Y, note.Id, note.Title)
	//	printf(`<notetext><![CDATA[<html dir="ltr"><head></head><body contenteditable="true">`)
	//	for _, line := range note.Text {
	//		printf(`%s<br/>`, line)
	//	}
	//	printfnl(`</body></html>]]></notetext></note>`)
	return nil
}

func encodeInformations(informations *wxx.Informations_t, wb *bytes.Buffer) error {
	wb.WriteString("<informations>\n")
	wb.WriteString("</informations>\n")
	return nil
}

func encodeConfiguration(configuration *wxx.Configuration_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<configuration>\n"))
	if err := encodeTerrainConfig(configuration.TerrainConfig, wb); err != nil {
		return err
	}
	if err := encodeFeatureConfig(configuration.FeatureConfig, wb); err != nil {
		return err
	}
	if err := encodeTextureConfig(configuration.TextureConfig, wb); err != nil {
		return err
	}
	if err := encodeTextConfig(configuration.TextConfig, wb); err != nil {
		return err
	}
	if err := encodeShapeConfig(configuration.ShapeConfig, wb); err != nil {
		return err
	}
	wb.WriteString(fmt.Sprintf("  </configuration>\n"))
	return nil
}

func encodeTerrainConfig(terrainConfig []*wxx.TerrainConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <terrain-config>\n")
	wb.WriteString("  </terrain-config>\n")
	return nil
}

func encodeFeatureConfig(featureConfig []*wxx.FeatureConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <feature-config>\n")
	wb.WriteString("  </feature-config>\n")
	return nil
}

func encodeTextureConfig(textureConfig []*wxx.TextureConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <texture-config>\n")
	wb.WriteString("  </texture-config>\n")
	return nil
}

func encodeTextConfig(textConfig *wxx.TextConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <text-config>\n")
	for _, labelStyle := range textConfig.LabelStyles {
		if err := encodeLabelStyle(labelStyle, wb); err != nil {
			return err
		}
	}
	wb.WriteString("  </text-config>\n")
	return nil
}

func encodeLabelStyle(labelStyle *wxx.LabelStyle_t, wb *bytes.Buffer) error {
	//wb.WriteString("<labelstyle")
	////name="Building"
	////fontFace="Arial"
	////scale="25.0"
	////isBold="false"
	////isItalic="false"
	////color="0.0,0.0,0.0,1.0"
	////backgroundColor="null"
	////outlineSize="0.0"
	////outlineColor="null"
	//wb.WriteString(" />\n\n")
	return nil
}

func encodeShapeConfig(shapeConfig *wxx.ShapeConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <shape-config>\n")
	for _, shapeStyle := range shapeConfig.ShapeStyles {
		if err := encodeShapeStyle(shapeStyle, wb); err != nil {
			return err
		}
	}
	wb.WriteString("  </shape-config>\n")
	return nil
}

func encodeShapeStyle(shapeStyle *wxx.ShapeStyle_t, wb *bytes.Buffer) error {
	wb.WriteString("<shapestyle")
	wb.WriteString(fmt.Sprintf(" name=%q", shapeStyle.Name))
	wb.WriteString(fmt.Sprintf(" strokeType=%q", shapeStyle.StrokeType))
	wb.WriteString(fmt.Sprintf(" isFractal=%q", bools(shapeStyle.IsFractal)))
	wb.WriteString(fmt.Sprintf(" strokeWidth=%q", floats(shapeStyle.StrokeWidth)))
	wb.WriteString(fmt.Sprintf(" opacity=%q", floats(shapeStyle.Opacity)))
	wb.WriteString(fmt.Sprintf(" snapVertices=%q", bools(shapeStyle.SnapVertices)))
	wb.WriteString(fmt.Sprintf(" tags=%q", shapeStyle.Tags))
	wb.WriteString(fmt.Sprintf(" dropShadow=%q", bools(shapeStyle.DropShadow)))
	wb.WriteString(fmt.Sprintf(" innerShadow=%q", bools(shapeStyle.InnerShadow)))
	wb.WriteString(fmt.Sprintf(" boxBlur=%q", bools(shapeStyle.BoxBlur)))
	wb.WriteString(fmt.Sprintf(" dsSpread=%q", floats(shapeStyle.DsSpread)))
	wb.WriteString(fmt.Sprintf(" dsRadius=%q", floats(shapeStyle.DsRadius)))
	wb.WriteString(fmt.Sprintf(" dsOffsetX=%q", floats(shapeStyle.DsOffsetX)))
	wb.WriteString(fmt.Sprintf(" dsOffsetY=%q", floats(shapeStyle.DsOffsetY)))
	wb.WriteString(fmt.Sprintf(" insChoke=%q", floats(shapeStyle.InsChoke)))
	wb.WriteString(fmt.Sprintf(" insRadius=%q", floats(shapeStyle.InsRadius)))
	wb.WriteString(fmt.Sprintf(" insOffsetX=%q", floats(shapeStyle.InsOffsetX)))
	wb.WriteString(fmt.Sprintf(" insOffsetY=%q", floats(shapeStyle.InsOffsetY)))
	wb.WriteString(fmt.Sprintf(" bbWidth=%q", floats(shapeStyle.BbWidth)))
	wb.WriteString(fmt.Sprintf(" bbHeight=%q", floats(shapeStyle.BbHeight)))
	wb.WriteString(fmt.Sprintf(" bbIterations=%q", ints(shapeStyle.BbIterations)))
	wb.WriteString(fmt.Sprintf(" fillTexture=%q", shapeStyle.FillTexture))         // nullable
	wb.WriteString(fmt.Sprintf(" strokeTexture=%q", shapeStyle.StrokeTexture))     // nullable
	wb.WriteString(fmt.Sprintf("  strokePaint=%q", rgbas(shapeStyle.StrokePaint))) // not nullable
	wb.WriteString(fmt.Sprintf("  fillPaint=%q", rgbans(shapeStyle.FillPaint)))    // nullable
	wb.WriteString(fmt.Sprintf("  dscolor=%q", rgbans(shapeStyle.DsColor)))        // nullable
	wb.WriteString(fmt.Sprintf("  insColor=%q", rgbans(shapeStyle.InsColor)))      // nullable
	wb.WriteString(" />\n")
	return nil
}

// boold formats a bool as an integer
func boold(b bool) int {
	if b {
		return 1
	}
	return 0
}

// bools formats a bool as a string
func bools(b bool) string {
	return fmt.Sprintf("%v", b)
}

// floatd formats a float as an integer.
func floatd(f float64) int {
	return int(f)
}

// floatf formats a float in the style that Worldographer expects.
// Zero values are rendered as 0.0.
// Note: floats is probably the right function to use.
func floatf(f float64) string {
	const epsilon = 1e-6
	if -epsilon < f && f <= epsilon {
		return "0.0"
	}
	return fmt.Sprintf("%g", f)
}

// floats converts a float64 number to a string representation adhering
// to certain Worldographer formatting rules.
//
// The function tries to represent the float in a manner that avoids scientific notation
// while preserving the fractional part of the float. It rounds off trailing zeros and
// ensures that there is always a digit after the decimal point.
//
// Parameters:
// - f: The float64 number to be converted.
//
// Returns:
//   - The string representation of the input float. If `f` is an integer, ".0" is appended to
//     signify that it is a float. For non-integer floats, trailing zeros after the decimal point are trimmed.
//
// Example:
//
//	floats(1234567.00) returns "1234567.0"
//	floats(0.120300) returns "0.1203"
func floats(f float64) string {
	s := fmt.Sprintf("%g", f)
	if strings.IndexByte(s, 'e') != -1 {
		s = fmt.Sprintf("%f", f)
	}
	if strings.IndexByte(s, '.') == -1 {
		return s + ".0"
	}
	s = strings.TrimRight(s, "0")
	if s[len(s)-1] == '.' {
		return s + "0"
	}
	return s
}

// floatg formats a float in the style that Worldographer expects.
func floatg(f float64) string {
	return fmt.Sprintf("%g", f)
}

// ints formats an int as a string
func ints(i int) string {
	return fmt.Sprintf("%d", i)
}

// rgbans converts an RGBA_t to a nullable string.
// It uses the rgbas function to format the RGBA_t
func rgbans(rgba *wxx.RGBA_t) string {
	s := rgbas(rgba)
	if s == "0.0,0.0,0.0,1.0" {
		s = "null"
	}
	return s
}

// rgbas converts an RGBA_t struct into an XML attribute string.
// RGBA_t struct contains four fields, each representing Red, Green, Blue and Alpha respectively.
// Each field is a float. We format the struct as a comma separated string.
// If the provided rgba pointer is nil, it defaults to "0.0,0.0,0.0,1.0".
//
// We use the floats function to format the float values into an XML-friendly format.
//
// Parameters:
// - rgba: a pointer to an RGBA_t struct. Can be nil.
//
// Returns:
// - A XML attribute string representing the rgba. If rgba is nil, returns "0.0,0.0,0.0,1.0"
func rgbas(rgba *wxx.RGBA_t) string {
	if rgba == nil {
		return "0.0,0.0,0.0,1.0"
	}
	return fmt.Sprintf("%s,%s,%s,%s",
		floats(rgba.R),
		floats(rgba.G),
		floats(rgba.B),
		floats(rgba.A))
}

// terrainMapToSlice converts a map of terrain names and slot into a list
// of strings for the xml map.terrainmap element.
func terrainMapToSlice(data map[string]int) []string {
	type terrain_t struct {
		slot int
		name string
	}
	list := []*terrain_t{}
	for k, v := range data {
		list = append(list, &terrain_t{
			slot: v,
			name: k,
		})
	}
	// list must be sorted
	sort.Slice(list, func(i, j int) bool {
		return list[i].slot < list[j].slot
	})
	var s []string
	for _, v := range list {
		s = append(s, v.name)
	}
	return s
}

func encodeInnerText(input string) string {
	escaped := html.EscapeString(input) // Escapes < > & "
	return strings.ReplaceAll(escaped, "\n", "&#10;")
}
