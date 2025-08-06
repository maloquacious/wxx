// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2017v1

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/maloquacious/wxx/models"
)

// Encode the Map_t into a slice of UTF-8 bytes that matches this version's XML schema.
//
// Note: the style of this code is intentionally verbose to make it easier to find changes between
// versions of the Worldographer files.
func Encode(w *models.Map_t) ([]byte, error) {
	wb := &bytes.Buffer{}
	if err := encodeMap(w, wb); err != nil {
		return nil, err
	}
	return wb.Bytes(), nil
}

func encodeMap(w *models.Map_t, wb *bytes.Buffer) error {
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
	wb.WriteString(fmt.Sprintf(" kingdomToProvinceVOffset=%q\n", floats(w.KingdomToProvinceVOffset)))
	wb.WriteString(fmt.Sprintf("hexWidth=%q", floats(w.HexWidth)))
	wb.WriteString(fmt.Sprintf(" hexHeight=%q", floats(w.HexHeight)))
	wb.WriteString(fmt.Sprintf(" hexOrientation=%q", w.HexOrientation))
	if w.MapProjection == models.FLAT {
		wb.WriteString(fmt.Sprintf(" mapProjection=%q", "FLAT"))
	} else if w.MapProjection == models.ICOSAHEDRAL {
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

	if err := encodeGridAndNumbering(w, wb); err != nil {
		return err
	}

	if err := encodeTerrainMap(w, wb); err != nil {
		return err
	}

	if err := encodeMapLayers(w, wb); err != nil {
		return err
	}

	if err := encodeTiles(w, wb); err != nil {
		return err
	}

	if err := encodeMapKey(w, wb); err != nil {
		return err
	}

	if err := encodeFeatures(w, wb); err != nil {
		return err
	}

	if err := encodeLabels(w, wb); err != nil {
		return err
	}

	if err := encodeShapes(w, wb); err != nil {
		return err
	}

	if err := encodeNotes(w, wb); err != nil {
		return err
	}

	if err := encodeInformations(w, wb); err != nil {
		return err
	}

	if err := encodeConfiguration(w, wb); err != nil {
		return err
	}

	wb.WriteString("</map>\n\n")

	return nil
}

func encodeGridAndNumbering(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf(`<gridandnumbering `))
	wb.WriteString(fmt.Sprintf(" color0=%q", w.GridAndNumbering.Color0))
	wb.WriteString(fmt.Sprintf(" color1=%q", w.GridAndNumbering.Color1))
	wb.WriteString(fmt.Sprintf(" color2=%q", w.GridAndNumbering.Color2))
	wb.WriteString(fmt.Sprintf(" color3=%q", w.GridAndNumbering.Color3))
	wb.WriteString(fmt.Sprintf(" color4=%q", w.GridAndNumbering.Color4))
	wb.WriteString(fmt.Sprintf(" width0=%q", floats(w.GridAndNumbering.Width0)))
	wb.WriteString(fmt.Sprintf(" width1=%q", floats(w.GridAndNumbering.Width1)))
	wb.WriteString(fmt.Sprintf(" width2=%q", floats(w.GridAndNumbering.Width2)))
	wb.WriteString(fmt.Sprintf(" width3=%q", floats(w.GridAndNumbering.Width3)))
	wb.WriteString(fmt.Sprintf(" width4=%q", floats(w.GridAndNumbering.Width4)))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomX=%q", floats(w.GridAndNumbering.GridOffsetContinentKingdomX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetContinentKingdomY=%q", floats(w.GridAndNumbering.GridOffsetContinentKingdomY)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentX=%q", floats(w.GridAndNumbering.GridOffsetWorldContinentX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldContinentY=%q", floats(w.GridAndNumbering.GridOffsetWorldContinentY)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomX=%q", floats(w.GridAndNumbering.GridOffsetWorldKingdomX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetWorldKingdomY=%q", floats(w.GridAndNumbering.GridOffsetWorldKingdomY)))
	wb.WriteString(fmt.Sprintf(" gridSquare=%q", ints(w.GridAndNumbering.GridSquare)))
	wb.WriteString(fmt.Sprintf(" gridSquareHeight=%q", floats(w.GridAndNumbering.GridSquareHeight)))
	wb.WriteString(fmt.Sprintf(" gridSquareWidth=%q", floats(w.GridAndNumbering.GridSquareWidth)))
	wb.WriteString(fmt.Sprintf(" gridOffsetX=%q", floats(w.GridAndNumbering.GridOffsetX)))
	wb.WriteString(fmt.Sprintf(" gridOffsetY=%q", floats(w.GridAndNumbering.GridOffsetY)))
	wb.WriteString(fmt.Sprintf(" numberFont=%q", w.GridAndNumbering.NumberFont))
	wb.WriteString(fmt.Sprintf(" numberColor=%q", w.GridAndNumbering.NumberColor))
	wb.WriteString(fmt.Sprintf(" numberSize=%q", ints(w.GridAndNumbering.NumberSize)))
	wb.WriteString(fmt.Sprintf(" numberStyle=%q", w.GridAndNumbering.NumberStyle))
	wb.WriteString(fmt.Sprintf(" numberFirstCol=%q", ints(w.GridAndNumbering.NumberFirstCol)))
	wb.WriteString(fmt.Sprintf(" numberFirstRow=%q", ints(w.GridAndNumbering.NumberFirstRow)))
	wb.WriteString(fmt.Sprintf(" numberOrder=%q", w.GridAndNumbering.NumberOrder))
	wb.WriteString(fmt.Sprintf(" numberPosition=%q", w.GridAndNumbering.NumberPosition))
	wb.WriteString(fmt.Sprintf(" numberPrePad=%q", w.GridAndNumbering.NumberPrePad))
	wb.WriteString(fmt.Sprintf(" numberSeparator=%q", w.GridAndNumbering.NumberSeparator))
	wb.WriteString(fmt.Sprintf("/>\n"))
	return nil
}

func encodeTerrainMap(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<terrainmap>"))
	for k, v := range terrainMapToSlice(w.TerrainMap.Data) {
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
func encodeMapLayers(w *models.Map_t, wb *bytes.Buffer) error {
	for _, mapLayer := range w.MapLayer {
		if err := encodeMapLayer(w, wb, mapLayer); err != nil {
			return err
		}
	}
	return nil
}

func encodeMapLayer(w *models.Map_t, wb *bytes.Buffer, mapLayer models.MapLayer_t) error {
	wb.WriteString("<maplayer")
	wb.WriteString(fmt.Sprintf(" name=%q", mapLayer.Name))
	wb.WriteString(fmt.Sprintf(" isVisible=%q", bools(mapLayer.IsVisible)))
	wb.WriteString("/>\n")
	return nil
}

func encodeTiles(w *models.Map_t, wb *bytes.Buffer) error {
	// to: width is the number of columns, height is the number of rows. does that depend on the orientation?
	wb.WriteString(fmt.Sprintf("<tiles"))
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", w.Tiles.ViewLevel))
	wb.WriteString(fmt.Sprintf(" tilesWide=%q", ints(w.Tiles.TilesWide)))
	wb.WriteString(fmt.Sprintf(" tilesHigh=%q", ints(w.Tiles.TilesHigh)))
	wb.WriteString(fmt.Sprintf(">\n"))

	// generate the tile-row elements:
	// * each tilerow will have a tile.tilesHigh lines of tab delimited data
	// * each line of data has the following values: Terrain type, elevation, is it icy, is it GM only, and its resources
	// * terrainType is an index into the terrainmap element
	// * resources are Animals, Brick, Crops, Gems, Lumber, Metals, and Rock, in that order, but are "compressed"
	if w.HexOrientation == "COLUMNS" {
		for x := 0; x < w.Tiles.TilesWide; x++ {
			wb.WriteString("<tilerow>\n")
			for y := 0; y < w.Tiles.TilesHigh; y++ {
				tile := w.Tiles.TileRows[x][y]
				if err := encodeTile(w, wb, tile); err != nil {
					return err
				}
			}
			wb.WriteString(fmt.Sprintf("</tilerow>\n"))
		}
	} else if w.HexOrientation == "ROWS" {
	} else {
		return fmt.Errorf("assert(orientation != %q)", w.HexOrientation)
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
func encodeTile(w *models.Map_t, wb *bytes.Buffer, tile *models.Tile_t) error {
	// todo: implement this
	wb.WriteString(fmt.Sprintf("%d", tile.Terrain))
	wb.WriteString(fmt.Sprintf("\t%d", floatd(tile.Elevation)))
	wb.WriteString(fmt.Sprintf("\t%d", boold(tile.IsIcy)))
	wb.WriteString(fmt.Sprintf("\t%d", boold(tile.IsGMOnly)))
	if err := encodeTileResources(w, wb, tile.Resources); err != nil {
		return err
	}
	if tile.CustomBackgroundColor != nil {
		wb.WriteString(fmt.Sprintf("\t%s", rgbas(tile.CustomBackgroundColor)))
	}
	wb.WriteString(fmt.Sprintf("\n"))
	return nil
}

// all resources are supposed to be in the range of 0...100, but we don't enforce
func encodeTileResources(w *models.Map_t, wb *bytes.Buffer, resources models.Resources_t) error {
	// compress if there are no resources
	if resources.IsZero() {
		wb.WriteString(fmt.Sprintf("\t%d", 0))
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
	wb.WriteString(fmt.Sprintf("\tZ"))
	return nil
}

func encodeMapKey(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf(`<mapkey positionx="0.0" positiony="0.0" viewlevel="WORLD" height="-1" backgroundcolor="0.9803921580314636,0.9215686321258545,0.843137264251709,1.0" backgroundopacity="50" titleText="Map Key" titleFontFace="Arial"  titleFontColor="0.0,0.0,0.0,1.0" titleFontBold="true" titleFontItalic="false" titleScale="80" scaleText="1 Hex = ? units" scaleFontFace="Arial"  scaleFontColor="0.0,0.0,0.0,1.0" scaleFontBold="true" scaleFontItalic="false" scaleScale="65" entryFontFace="Arial"  entryFontColor="0.0,0.0,0.0,1.0" entryFontBold="true" entryFontItalic="false" entryScale="55"  >`))
	wb.WriteByte('\n')
	wb.WriteString("</mapkey>\n")
	return nil
}

// add features
func encodeFeatures(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("<features>\n")
	for _, feature := range w.Features {
		if err := encodeFeature(w, wb, feature); err != nil {
			return err
		}
	}
	wb.WriteString("</features>\n")
	return nil
}

func encodeFeature(w *models.Map_t, wb *bytes.Buffer, feature *models.Feature_t) error {
	wb.WriteString(fmt.Sprintf("<feature"))
	wb.WriteString(fmt.Sprintf(" type=%q", feature.Type))
	wb.WriteString(fmt.Sprintf(" rotate=%q", floats(feature.Rotate)))
	wb.WriteString(fmt.Sprintf(" uuid=%q", feature.Uuid))
	wb.WriteString(fmt.Sprintf(" mapLayer=%q", feature.MapLayer))
	wb.WriteString(fmt.Sprintf(" isFlipHorizontal=%q", bools(feature.IsFlipHorizontal)))
	wb.WriteString(fmt.Sprintf(" isFlipVertical=%q", bools(feature.IsFlipVertical)))
	wb.WriteString(fmt.Sprintf(" scale=%q", floats(feature.Scale)))
	wb.WriteString(fmt.Sprintf(" scaleHt=%q", floats(feature.Scale)))
	wb.WriteString(fmt.Sprintf(" tags=%q", feature.Tags))
	wb.WriteString(fmt.Sprintf(" color=%q", rgbas(feature.Color)))
	wb.WriteString(fmt.Sprintf(" ringcolor=%q", rgbas(feature.RingColor)))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(feature.IsGMOnly)))
	wb.WriteString(fmt.Sprintf(" isPlaceFreely=%q", bools(feature.IsPlaceFreely)))
	wb.WriteString(fmt.Sprintf(" labelPosition=%q", feature.LabelPosition))
	wb.WriteString(fmt.Sprintf(" labelDistance=%q", floats(feature.LabelDistance)))
	wb.WriteString(fmt.Sprintf(" isWorld=%q", bools(feature.IsWorld)))
	wb.WriteString(fmt.Sprintf(" isContinent=%q", bools(feature.IsContinent)))
	wb.WriteString(fmt.Sprintf(" isKingdom=%q", bools(feature.IsKingdom)))
	wb.WriteString(fmt.Sprintf(" isProvince=%q", bools(feature.IsProvince)))
	wb.WriteString(fmt.Sprintf(" isFillHexBottom=%q", bools(feature.IsFillHexBottom)))
	wb.WriteString(fmt.Sprintf(" isHideTerrainIcon=%q", bools(feature.IsHideTerrainIcon)))
	wb.WriteString(">")
	if feature.Location != nil {
		if err := encodeFeatureLocation(w, wb, feature.Location); err != nil {
			return err
		}
	}
	if feature.Label != nil {
		if err := encodeFeatureLabel(w, wb, feature.Label); err != nil {
			return err
		}
	}
	wb.WriteString("</feature>\n")
	return nil
}

func encodeFeatureLocation(w *models.Map_t, wb *bytes.Buffer, location *models.FeatureLocation_t) error {
	wb.WriteString("<location")
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", location.ViewLevel))
	wb.WriteString(fmt.Sprintf(" x=%q", floats(location.X)))
	wb.WriteString(fmt.Sprintf(" y=%q", floats(location.Y)))
	wb.WriteString("/>")
	return nil
}

func encodeFeatureLabel(w *models.Map_t, wb *bytes.Buffer, label *models.Label_t) error {
	return encodeLabel(w, wb, label)
}

func encodeLabels(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("<labels>\n")
	for _, label := range w.Labels {
		if err := encodeLabel(w, wb, label); err != nil {
			return err
		}
	}
	wb.WriteString("</labels>\n")
	return nil
}

func encodeLabel(w *models.Map_t, wb *bytes.Buffer, label *models.Label_t) error {
	wb.WriteString("<label")
	wb.WriteString(fmt.Sprintf(" mapLayer=%q", label.MapLayer))
	wb.WriteString(fmt.Sprintf(" style=%q", label.Style))       // can be null!
	wb.WriteString(fmt.Sprintf(" fontFace=%q", label.FontFace)) // can be null!
	wb.WriteString(fmt.Sprintf(" color=%q", rgbas(label.Color)))
	wb.WriteString(fmt.Sprintf(" outlineColor=%q", rgbas(label.OutlineColor)))
	wb.WriteString(fmt.Sprintf(" outlineSize=%q", floats(label.OutlineSize)))
	wb.WriteString(fmt.Sprintf(" isBold=%q", bools(label.IsBold)))
	wb.WriteString(fmt.Sprintf(" isItalic=%q", bools(label.IsItalic)))
	wb.WriteString(fmt.Sprintf(" isWorld=%q", bools(label.IsWorld)))
	wb.WriteString(fmt.Sprintf(" isContinent=%q", bools(label.IsContinent)))
	wb.WriteString(fmt.Sprintf(" isKingdom=%q", bools(label.IsKingdom)))
	wb.WriteString(fmt.Sprintf(" isProvince=%q", bools(label.IsProvince)))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(label.IsGMOnly)))
	wb.WriteString(fmt.Sprintf(" tags=%q", label.Tags))
	wb.WriteString(">")
	if err := encodeLabelLocation(w, wb, label.Location); err != nil {
		return err
	}
	if label.InnerText != "" {
		wb.WriteString(label.InnerText)
	}
	wb.WriteString("</label>\n")
	return nil
}

func encodeLabelLocation(w *models.Map_t, wb *bytes.Buffer, location *models.LabelLocation_t) error {
	wb.WriteString("<location")
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", location.ViewLevel))
	wb.WriteString(fmt.Sprintf(" x=%q", floats(location.X)))
	wb.WriteString(fmt.Sprintf(" y=%q", floats(location.Y)))
	wb.WriteString(fmt.Sprintf(" scale=%q", floats(location.Scale)))
	wb.WriteString("/>")
	return nil
}

func encodeShapes(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("<shapes>\n")
	for _, shape := range w.Shapes {
		if err := encodeShape(w, wb, shape); err != nil {
			return err
		}
	}
	wb.WriteString("</shapes>\n")
	return nil
}

func encodeShape(w *models.Map_t, wb *bytes.Buffer, shape *models.Shape_t) error {
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

func encodeNotes(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("<notes>\n")
	for _, note := range w.Notes {
		if err := encodeNote(w, wb, note); err != nil {
			return err
		}
	}
	wb.WriteString("</notes>\n")
	return nil
}

func encodeNote(w *models.Map_t, wb *bytes.Buffer, note *models.Note_t) error {
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

func encodeInformations(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("<informations>\n")
	wb.WriteString("</informations>\n")
	return nil
}

func encodeConfiguration(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<configuration>\n"))
	if err := encodeTerrainConfig(w, wb); err != nil {
		return err
	}
	if err := encodeFeatureConfig(w, wb); err != nil {
		return err
	}
	if err := encodeTextureConfig(w, wb); err != nil {
		return err
	}
	if err := encodeTextConfig(w, wb); err != nil {
		return err
	}
	if err := encodeShapeConfig(w, wb); err != nil {
		return err
	}
	wb.WriteString(fmt.Sprintf("  </configuration>\n"))
	return nil
}

func encodeTerrainConfig(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("  <terrain-config>\n")
	wb.WriteString("  </terrain-config>\n")
	return nil
}

func encodeFeatureConfig(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("  <feature-config>\n")
	wb.WriteString("  </feature-config>\n")
	return nil
}

func encodeTextureConfig(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("  <texture-config>\n")
	wb.WriteString("  </texture-config>\n")
	return nil
}

func encodeTextConfig(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("  <text-config>\n")
	wb.WriteString("  </text-config>\n")
	return nil
}

func encodeShapeConfig(w *models.Map_t, wb *bytes.Buffer) error {
	wb.WriteString("  <shape-config>\n")
	wb.WriteString("  </shape-config>\n")
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
func rgbans(rgba *models.RGBA_t) string {
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
func rgbas(rgba *models.RGBA_t) string {
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
