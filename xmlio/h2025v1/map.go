// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
)

// Decode the XML data using the H2025.V1 schema and return a Map_t or an error.
//
// Decode does a single xml.Unmarshal into the version-specific XMLSchema structs
// (schema.go) then copies the <map> root attributes here and dispatches each
// child element to its co-located decodeXxx helper (see the per-element files).
func Decode(input []byte) (*wxx.Map_t, error) {
	m := &XMLSchema{}

	// unmarshal into a structure that's built just for the conversion
	err := xml.Unmarshal(input, &m)
	if err != nil {
		log.Printf("h2025v1: %v\n", err)
		return nil, err
	}
	if m.Release == "" {
		return nil, fmt.Errorf("missing map.Release")
	} else if m.Version == "" {
		return nil, fmt.Errorf("missing map.Version")
	} else if m.Schema == "" {
		return nil, fmt.Errorf("missing map.Schema")
	}
	if m.Release != "2025" {
		return nil, fmt.Errorf("%s/%s/%s: unsupported release", m.Release, m.Version, m.Schema)
	}
	// derive the data version from the schema attribute: Major is the release
	// year (2025) and Minor/Patch are the dotted components of the schema.
	// e.g. schema "1.06" -> {2025,1,6}; schema "1.01" -> {2025,1,1}.
	dataVersion := semver.Version{Major: 2025}
	schemaParts := strings.Split(m.Schema, ".")
	if len(schemaParts) != 2 {
		return nil, fmt.Errorf("%s/%s/%s: malformed schema", m.Release, m.Version, m.Schema)
	}
	if dataVersion.Minor, err = strconv.Atoi(schemaParts[0]); err != nil {
		return nil, fmt.Errorf("%s/%s/%s: schema minor: %w", m.Release, m.Version, m.Schema, err)
	}
	if dataVersion.Patch, err = strconv.Atoi(schemaParts[1]); err != nil {
		return nil, fmt.Errorf("%s/%s/%s: schema patch: %w", m.Release, m.Version, m.Schema, err)
	}

	// process source into a WXX structure and return it or any errors
	w := &wxx.Map_t{}
	w.MetaData.AppVersion = wxx.Version()
	w.MetaData.DataVersion = dataVersion
	w.MetaData.Created = time.Now().UTC().Format(time.RFC3339)
	w.MetaData.Worldographer.Name = "unknown"
	w.MetaData.Worldographer.Release = m.Release
	w.MetaData.Worldographer.Version = m.Version
	w.MetaData.Worldographer.Schema = m.Schema
	// w.MetaData.Worldographer.Created = time.Time{}

	w.ContinentFactor = m.ContinentFactor
	w.ContinentToKingdomHOffset = m.ContinentToKingdomHOffset
	w.ContinentToKingdomVOffset = m.ContinentToKingdomVOffset
	w.HexHeight = m.HexHeight
	w.HexOrientation = m.HexOrientation
	switch m.HexOrientation {
	case "COLUMNS":
		w.GridOrientation = hexg.OddQ
	case "ROWS":
		w.GridOrientation = hexg.OddR
	default:
		return nil, fmt.Errorf("%q: unknown orientation", m.HexOrientation)
	}
	w.HexWidth = m.HexWidth
	w.KingdomFactor = m.KingdomFactor
	w.KingdomToProvinceHOffset = m.KingdomToProvinceHOffset
	w.KingdomToProvinceVOffset = m.KingdomToProvinceVOffset
	w.LastViewLevel = m.LastViewLevel
	switch m.MapProjection {
	case "FLAT":
		w.MapProjection = wxx.FLAT
	case "ICOSAHEDRAL":
		w.MapProjection = wxx.ICOSAHEDRAL
	default:
		return nil, fmt.Errorf("%q: unknown projection", m.MapProjection)
	}
	w.ProvinceFactor = m.ProvinceFactor
	w.Release = m.Release
	w.Schema = m.Schema
	w.ShowFeatureLabels = m.ShowFeatureLabels
	w.ShowGMOnly = m.ShowGMOnly
	w.ShowGMOnlyGlow = m.ShowGMOnlyGlow
	w.ShowGrid = m.ShowGrid
	w.ShowGridNumbers = m.ShowGridNumbers
	w.ShowNotes = m.ShowNotes
	w.ShowShadows = m.ShowShadows
	w.TriangleSize = m.TriangleSize
	w.Type = m.Type
	w.Version = m.Version
	w.WorldToContinentHOffset = m.WorldToContinentHOffset
	w.WorldToContinentVOffset = m.WorldToContinentVOffset
	w.HScrollbarPos = m.HScrollbarPos
	w.VScrollbarPos = m.VScrollbarPos

	decodeGridAndNumbering(m.GridAndNumbering, w)

	decodeBlurTerrainBG(m.BlurTerrainBG, w)

	if err := decodeTerrainMap(m.TerrainMap, w); err != nil {
		return w, err
	}

	decodeMapLayers(m.MapLayers, w)

	if err := decodeTiles(m.Tiles, m.MapKey, w); err != nil {
		return w, err
	}

	if err := decodeFeatures(m.Features, w); err != nil {
		return w, err
	}

	decodeExtraTerrain(m.ExtraTerrain, w)

	if err := decodeLabels(m.Labels, w); err != nil {
		return w, err
	}

	decodeShapes(m.Shapes, w)

	if err := decodeNotes(m.Notes, w); err != nil {
		return w, err
	}

	decodeInformations(m.Informations, w)

	if err := decodeConfiguration(m.Configuration, w); err != nil {
		return w, err
	}

	return w, nil
}

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
	wb.WriteString(fmt.Sprintf(" release=%q", w.Release))
	wb.WriteString(fmt.Sprintf(" version=%q", w.Version))
	wb.WriteString(fmt.Sprintf(" schema=%q", w.Schema))
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
	wb.WriteString(fmt.Sprintf("hScrollbarPos=%q vScrollbarPos=%q \n", floats(w.HScrollbarPos), floats(w.VScrollbarPos)))
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

	if err := encodeBlurTerrainBG(w.BlurTerrainBG, wb); err != nil {
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

	if err := encodeExtraTerrain(w.ExtraTerrain, wb); err != nil {
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
