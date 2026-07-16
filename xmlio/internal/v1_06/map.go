// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
)

// dottedOrRaw parses an on-disk dotted version, falling back to a Dotted that
// carries the verbatim bytes with zero components when the string does not fit
// the dotted grammar.
//
// The fallback is deliberate: modeling these values must not turn a file that
// decodes today into one that errors, so this never adds an error path of its
// own. Raw is authoritative for output and is preserved in every case; the
// components exist only to compare. Validating an identity against the set of
// supported releases is the registry's job, not the decoder's.
func dottedOrRaw(s string) wxx.Dotted {
	d, err := wxx.ParseDotted(s)
	if err != nil {
		return wxx.Dotted{Raw: s}
	}
	return d
}

// Decode the XML data using the H2025.V1 schema and return a Map_t or an error.
// It is a work in progress; see COVERAGE.md.
//
// Decode does a single xml.Unmarshal into the version-specific XMLSchema structs
// (schema.go) then copies the <map> root attributes here and dispatches each
// child element to its co-located decodeXxx helper (see the per-element files).
//
// This is the whole of decode dispatch: xmlio's decoder.go calls it directly,
// having read the map/@release that identifies the file. There is no Decode on
// codec.Codec to route through -- see that interface for why.
func Decode(input []byte) (*wxx.Map_t, error) {
	m := &XMLSchema{}

	// unmarshal into a structure that's built just for the conversion
	err := xml.Unmarshal(input, &m)
	if err != nil {
		log.Printf("v1_06: %v\n", err)
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
	// The schema must parse: it is what selects the codec on the way back out
	// (ADR 0004 Decision 4), so a file whose @schema is not a dotted version is
	// rejected here rather than carried as bytes nothing can dispatch on. This
	// is the same input the removed schema-to-semver conversion rejected.
	schema, err := wxx.ParseDotted(m.Schema)
	if err != nil {
		return nil, fmt.Errorf("%s/%s/%s: malformed schema: %w", m.Release, m.Version, m.Schema, err)
	}

	// process source into a WXX structure and return it or any errors
	w := &wxx.Map_t{}
	w.MetaData.AppVersion = wxx.Version()
	// Version is the on-disk identity (ADR 0004 Decision 2). App is the dotted
	// <map version> ("2.06"), which the superseded DataVersion had no slot for --
	// it spent its Minor.Patch on the schema, so until now @version survived only
	// as an unexamined string. Schema is the dotted <map schema> ("1.06"), always
	// non-nil for W2025 because the guards above reject a file that states none
	// (ADR 0003 Decision 2). The zero padding is preserved: Raw is "2.06", not
	// the "2.6" a semver round-trip would return.
	w.MetaData.Version = wxx.Version_t{App: dottedOrRaw(m.Version), Schema: &schema}
	w.MetaData.Created = time.Now().UTC().Format(time.RFC3339)
	w.MetaData.Worldographer.Name = "unknown"
	// The three identity attributes the source states, verbatim. This is the ONLY
	// place they are recorded: they are provenance -- what wrote the file that came
	// in -- and provenance lives with the file's other metadata (issue #45 Decision
	// 9). The decoder is right to keep them; nothing downstream may treat them as
	// what to write, and no encoder reads them.
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
	w.ShowFeatureLabels = m.ShowFeatureLabels
	w.ShowGMOnly = m.ShowGMOnly
	w.ShowGMOnlyGlow = m.ShowGMOnlyGlow
	w.ShowGrid = m.ShowGrid
	w.ShowGridNumbers = m.ShowGridNumbers
	w.ShowNotes = m.ShowNotes
	w.ShowShadows = m.ShowShadows
	w.TriangleSize = m.TriangleSize
	w.Type = m.Type
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
// app is the application version to write the map as, verbatim as map/@version
// states it ("2.06"). It is verified against this codec's declared set before
// anything is emitted (see acceptedApps): this codec writes one schema, so an
// application version outside that set names a release that never existed, and
// writing it would produce a file claiming a build that did not write this format
// (issue #41 requirement 3).
//
// The check comes first, before any encode error can, so that a caller asking for
// an impossible release is told that rather than told about the contents of a map
// it was never going to get.
//
// Resolving app through acceptedApps.App RATHER than merely verifying it is what
// closes issue #45. The gate and the write path must read the SAME input: the
// former shape verified the app ARGUMENT and then wrote w.Release/w.Version/
// w.Schema, the map's own fields, so the two read different inputs and the
// verified argument was never written. Encode(classicMap, "2.06") passed the gate
// -- 2.06 IS this codec's version -- and then emitted W2025 content under the
// classic identity the map still stated: release="" version="1.77" schema="",
// which re-decodes silently as classic. That chimera is now unconstructible
// rather than merely prevented, because the resolved App_t and this codec's
// declared schema are the only identity inputs encodeMap is given.
//
// Note: the style of this code is intentionally verbose to make it easier to find changes between
// versions of the Worldographer files.
func Encode(w *wxx.Map_t, app string) ([]byte, error) {
	target, ok := acceptedApps.App(app)
	if !ok {
		// VerifyApp is what names the accepted set in the error text; App reports
		// only that the lookup missed.
		return nil, acceptedApps.VerifyApp(app)
	}
	wb := &bytes.Buffer{}
	if err := encodeMap(w, target, acceptedApps.Schema, wb); err != nil {
		return nil, err
	}
	return wb.Bytes(), nil
}

// encodeMap writes the <map> element for the application version target, stating
// schema.
//
// target and schema carry every identity value this element states, and w carries
// none of them: the identity written is the one the caller asked for, never the
// one the decoded map happens to state (issue #45 Decisions 4 and 6). w is
// consulted for map CONTENT only.
//
// map/@release comes from target rather than being a constant of this codec
// because it is DERIVED from the application version (issue #45 Decision 5):
// 2.06 writes "2025", and a later build on schema 1.06 shipped under a different
// label writes a different one. map/@schema is this codec's single declared
// constant, threaded in rather than read here so that the signature names every
// identity byte this function writes.
//
// Both strings, and target.Version, are written verbatim and are never re-rendered
// from a parsed version's components, because "2.06" must never reach disk as
// "2.6" (ADR 0004 Decision 1). target.Version is the app argument itself: App
// matched it by string equality, so the two are the same string.
func encodeMap(w *wxx.Map_t, target appver.App_t, schema string, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<map"))
	wb.WriteString(fmt.Sprintf(" type=%q", w.Type))
	wb.WriteString(fmt.Sprintf(" release=%q", target.Release))
	wb.WriteString(fmt.Sprintf(" version=%q", target.Version))
	wb.WriteString(fmt.Sprintf(" schema=%q", schema))
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
