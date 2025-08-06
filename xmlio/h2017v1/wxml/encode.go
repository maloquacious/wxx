// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxml

import (
	"bytes"
	_ "embed"
	"log"
	"sort"
	"text/template"

	"github.com/maloquacious/wxx/models"
)

var (
	//go:embed "schema.goxml"
	xmlSchemaTemplate string
)

// Encode marshals the Map_t to XML using custom templates.
// It does not add the xml header.
func Encode(w *models.Map_t) ([]byte, error) {
	// convert the map data into our schema for writing
	s := &Schema_t{}

	// map attributes
	s.Type = w.Type
	s.Version = w.Version
	s.LastViewLevel = w.LastViewLevel
	s.ContinentFactor = w.ContinentFactor
	s.KingdomFactor = w.KingdomFactor
	s.ProvinceFactor = w.ProvinceFactor
	s.ContinentToKingdomHOffset = w.ContinentToKingdomHOffset
	s.KingdomToProvinceHOffset = w.KingdomToProvinceHOffset
	s.WorldToContinentVOffset = w.WorldToContinentVOffset
	s.ContinentToKingdomVOffset = w.ContinentToKingdomVOffset
	s.KingdomToProvinceVOffset = w.KingdomToProvinceVOffset
	s.HexWidth = w.HexWidth
	s.HexHeight = w.HexHeight
	s.HexOrientation = w.HexOrientation
	s.MapProjection = w.MapProjection
	s.ShowNotes = w.ShowNotes
	s.ShowGMOnly = w.ShowGMOnly
	s.ShowGMOnlyGlow = w.ShowGMOnlyGlow
	s.ShowFeatureLabels = w.ShowFeatureLabels
	s.ShowGrid = w.ShowGrid
	s.ShowGridNumbers = w.ShowGridNumbers
	s.ShowShadows = w.ShowShadows
	s.TriangleSize = w.TriangleSize

	// map elements
	s.GridAndNumbering = &GridAndNumbering_t{
		Color0:                      w.GridAndNumbering.Color0,
		Color1:                      w.GridAndNumbering.Color1,
		Color2:                      w.GridAndNumbering.Color2,
		Color3:                      w.GridAndNumbering.Color3,
		Color4:                      w.GridAndNumbering.Color4,
		Width0:                      w.GridAndNumbering.Width0,
		Width1:                      w.GridAndNumbering.Width1,
		Width2:                      w.GridAndNumbering.Width2,
		Width3:                      w.GridAndNumbering.Width3,
		Width4:                      w.GridAndNumbering.Width4,
		GridOffsetContinentKingdomX: w.GridAndNumbering.GridOffsetContinentKingdomX,
		GridOffsetContinentKingdomY: w.GridAndNumbering.GridOffsetContinentKingdomY,
		GridOffsetWorldContinentX:   w.GridAndNumbering.GridOffsetWorldContinentX,
		GridOffsetWorldContinentY:   w.GridAndNumbering.GridOffsetWorldContinentY,
		GridOffsetWorldKingdomX:     w.GridAndNumbering.GridOffsetWorldKingdomX,
		GridOffsetWorldKingdomY:     w.GridAndNumbering.GridOffsetWorldKingdomY,
		GridSquare:                  w.GridAndNumbering.GridSquare,
		GridSquareHeight:            w.GridAndNumbering.GridSquareHeight,
		GridSquareWidth:             w.GridAndNumbering.GridSquareWidth,
		GridOffsetX:                 w.GridAndNumbering.GridOffsetX,
		GridOffsetY:                 w.GridAndNumbering.GridOffsetY,
		NumberFont:                  w.GridAndNumbering.NumberFont,
		NumberColor:                 w.GridAndNumbering.NumberColor,
		NumberSize:                  w.GridAndNumbering.NumberSize,
		NumberStyle:                 w.GridAndNumbering.NumberStyle,
		NumberFirstCol:              w.GridAndNumbering.NumberFirstCol,
		NumberFirstRow:              w.GridAndNumbering.NumberFirstRow,
		NumberOrder:                 w.GridAndNumbering.NumberOrder,
		NumberPosition:              w.GridAndNumbering.NumberPosition,
		NumberPrePad:                w.GridAndNumbering.NumberPrePad,
		NumberSeparator:             w.GridAndNumbering.NumberSeparator,
	}
	s.TerrainMap = terrainMapToSlice(w.TerrainMap.Data)
	log.Printf("terrainMap %+v\n", s.TerrainMap)
	for k, v := range s.TerrainMap {
		log.Printf("terrainMap %3d %3d %q\n", k, v.Index, v.Label)
	}

	// use our template to convert the schema into XML
	return encode(s)
}

func encode(s *Schema_t) ([]byte, error) {
	t, err := template.New("h2017v1").Parse(xmlSchemaTemplate)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	err = t.Execute(b, s)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func terrainMapToSlice(data map[string]int) []*Terrain_t {
	list := []*Terrain_t{}
	for k, v := range data {
		list = append(list, &Terrain_t{
			Index: v,
			Label: k,
		})
	}
	// list must be sorted
	sort.Slice(list, func(i, j int) bool {
		return list[i].Index < list[j].Index
	})
	return list
}
