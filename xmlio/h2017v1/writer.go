// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package h2017v1

import (
	"bytes"
	"fmt"
	"sort"

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
	wb.WriteString(fmt.Sprintf(" continentFactor=%q", intf(w.ContinentFactor)))
	wb.WriteString(fmt.Sprintf(" kingdomFactor=%q", intf(w.KingdomFactor)))
	wb.WriteString(fmt.Sprintf(" provinceFactor=%q", intf(w.ProvinceFactor)))
	wb.WriteString(fmt.Sprintf(" worldToContinentHOffset=%q", floatf(w.WorldToContinentHOffset)))
	wb.WriteString(fmt.Sprintf(" continentToKingdomHOffset=%q", floatf(w.ContinentToKingdomHOffset)))
	wb.WriteString(fmt.Sprintf(" kingdomToProvinceHOffset=%q", floatf(w.KingdomToProvinceHOffset)))
	wb.WriteString(fmt.Sprintf(" worldToContinentVOffset=%q", floatf(w.WorldToContinentVOffset)))
	wb.WriteString(fmt.Sprintf(" continentToKingdomVOffset=%q", floatf(w.ContinentToKingdomVOffset)))
	wb.WriteString(fmt.Sprintf(" kingdomToProvinceVOffset=%q\n", floatf(w.KingdomToProvinceVOffset)))
	wb.WriteString(fmt.Sprintf("hexWidth=%q", floatf(w.HexWidth)))
	wb.WriteString(fmt.Sprintf(" hexHeight=%q", floatf(w.HexHeight)))
	wb.WriteString(fmt.Sprintf(" hexOrientation=%q", w.HexOrientation))
	wb.WriteString(fmt.Sprintf(" mapProjection=%q", w.MapProjection))
	wb.WriteString(fmt.Sprintf(" showNotes=%q", boolf(w.ShowNotes)))
	wb.WriteString(fmt.Sprintf(" showGMOnly=%q", boolf(w.ShowGMOnly)))
	wb.WriteString(fmt.Sprintf(" showGMOnlyGlow=%q", boolf(w.ShowGMOnlyGlow)))
	wb.WriteString(fmt.Sprintf(" showFeatureLabels=%q", boolf(w.ShowFeatureLabels)))
	wb.WriteString(fmt.Sprintf(" showGrid=%q", boolf(w.ShowGrid)))
	wb.WriteString(fmt.Sprintf(" showGridNumbers=%q", boolf(w.ShowGridNumbers)))
	wb.WriteString(fmt.Sprintf(" showShadows=%q", boolf(w.ShowShadows)))
	wb.WriteString(fmt.Sprintf("  triangleSize=%q", intf(w.TriangleSize)))
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
	wb.WriteString(fmt.Sprintf(` width0="1.0"`))
	wb.WriteString(fmt.Sprintf(` width1="2.0"`))
	wb.WriteString(fmt.Sprintf(` width2="3.0"`))
	wb.WriteString(fmt.Sprintf(` width3="4.0"`))
	wb.WriteString(fmt.Sprintf(` width4="1.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetContinentKingdomX="0.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetContinentKingdomY="0.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetWorldContinentX="0.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetWorldContinentY="0.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetWorldKingdomX="0.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetWorldKingdomY="0.0"`))
	wb.WriteString(fmt.Sprintf(` gridSquare="0"`))
	wb.WriteString(fmt.Sprintf(` gridSquareHeight="-1.0"`))
	wb.WriteString(fmt.Sprintf(` gridSquareWidth="-1.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetX="0.0"`))
	wb.WriteString(fmt.Sprintf(` gridOffsetY="0.0"`))
	wb.WriteString(fmt.Sprintf(` numberFont="Arial"`))
	wb.WriteString(fmt.Sprintf(` numberColor="0x000000ff"`))
	wb.WriteString(fmt.Sprintf(` numberSize="20"`))
	wb.WriteString(fmt.Sprintf(` numberStyle="PLAIN"`))
	wb.WriteString(fmt.Sprintf(` numberFirstCol="0"`))
	wb.WriteString(fmt.Sprintf(` numberFirstRow="0"`))
	wb.WriteString(fmt.Sprintf(` numberOrder="COL_ROW"`))
	wb.WriteString(fmt.Sprintf(` numberPosition="BOTTOM"`))
	wb.WriteString(fmt.Sprintf(` numberPrePad="DOUBLE_ZERO"`))
	wb.WriteString(fmt.Sprintf(` numberSeparator="."`))
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
	wb.WriteString(fmt.Sprintf(" isVisible=%q", boolf(mapLayer.IsVisible)))
	wb.WriteString("/>\n")
	return nil
}

func encodeTiles(w *models.Map_t, wb *bytes.Buffer) error {
	// to: width is the number of columns, height is the number of rows. does that depend on the orientation?
	wb.WriteString(fmt.Sprintf("<tiles"))
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", w.Tiles.ViewLevel))
	wb.WriteString(fmt.Sprintf(" tilesWide=%q", intf(w.Tiles.TilesWide)))
	wb.WriteString(fmt.Sprintf(" tilesHigh=%q", intf(w.Tiles.TilesHigh)))
	wb.WriteString(fmt.Sprintf(">\n"))
	// generate the tile-row elements. the order that we render them in depends on the orientation of the map.
	if w.HexOrientation == "COLUMNS" {
		// we're using COLUMNS orientation so we have to generate all the columns for a single row before we move on to the next row.
		for gridColumn := 0; gridColumn < w.Tiles.TilesWide; gridColumn++ {
			wb.WriteString(fmt.Sprintf("<tilerow>\n"))
			// generate all the tiles in this column, one tile per row
			for gridRow := 0; gridRow < w.Tiles.TilesHigh; gridRow++ {
				// todo: implement this
				wb.WriteString(fmt.Sprintf("0\t1\t0\t0\t0\tZ\n"))
			}
			wb.WriteString(fmt.Sprintf("</tilerow>\n"))
		}
	} else {
		return fmt.Errorf("assert(orientation != %q)", w.HexOrientation)
	}
	wb.WriteString(fmt.Sprintf("</tiles>\n"))
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
	//			printf(`<feature type="Three Dots" rotate="0.0" uuid="%s" mapLayer="Tribenet Origin" isFlipHorizontal="false" isFlipVertical="false" scale="-1.0" scaleHt="-1.0" tags="" color="0.800000011920929,0.800000011920929,0.800000011920929,1.0" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false">`, uuid.NewString())
	//			printf(`<location viewLevel="WORLD" x="%f" y="%f" />`, origin.X, origin.Y)
	//			printf(`<label  mapLayer="Tribenet Origin" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
	//			printf(`<location viewLevel="WORLD" x="%f" y="%f" scale="25.0" />`, origin.X, origin.Y)
	//			printf(`</label>`)
	//			printf("</feature>\n")
	return nil
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
	//			printf(`<label  mapLayer="Tribenet Coords" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
	//			printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.25" />`, points[0].X, points[0].Y)
	//			printf("0")
	//			printf("</label>\n")
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

// boolf formats a bool as a string
func boolf(b bool) string {
	return fmt.Sprintf("%v", b)
}

// floatf formats a float in the style that Worldographer expects.
// Zero values are rendered as 0.0.
func floatf(f float64) string {
	const epsilon = 1e-6
	if -epsilon < f && f <= epsilon {
		return "0.0"
	}
	return fmt.Sprintf("%g", f)
}

// floatg formats a float in the style that Worldographer expects.
func floatg(f float64) string {
	return fmt.Sprintf("%g", f)
}

// intf formats an int as a string
func intf(i int) string {
	return fmt.Sprintf("%d", i)
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
