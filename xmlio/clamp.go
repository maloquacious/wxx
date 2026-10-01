// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio

import (
	"fmt"

	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// ClampedValue_t is one value the decoder changed, on the way in, to bring it
// into the range Worldographer keeps (issue #124).
//
// A clamp is a LOSS: the decoded Map_t no longer holds what the file said, and
// an encode writes the clamped value back. The decoder does it instead of
// refusing the file, because the app is assumed to do the same (see Reason), and
// it reports every one so that a caller can tell the user what changed.
//
// Today one thing is clamped: a <terrainAndLocation> @resources value (a
// resource on an <extraTerrain> layer) outside 0..100. A tile record's
// resources outside 0..100 are refused on decode instead, and Map_t.Validate
// refuses either kind before an encode, so wxx never writes one.
type ClampedValue_t struct {
	// Path is the on-disk location, as local-names joined with '/'
	// (`map/extraTerrain/mapLayer[@name="Below All"]/terrainAndLocation[0]/@resources`).
	Path string

	// Field is the Map_t field that now holds Now
	// ("Map_t.ExtraTerrain.MapLayers[0].Terrain[0].Resources.Brick"). For an
	// <extraTerrain> resource, the resource NAME comes from the order the
	// decoder assumes for that attribute, which is not verified (issue #124).
	Field string

	// Was is the value the file stated; Now is the value the map holds.
	Was int
	Now int

	// Reason is why the value was changed rather than kept or refused.
	Reason string
}

// String renders one clamp as a report line.
func (c ClampedValue_t) String() string {
	return fmt.Sprintf("%s (%s): %d clamped to %d -- %s", c.Path, c.Field, c.Was, c.Now, c.Reason)
}

// extraTerrainResourceReason is the Reason on every clamped <extraTerrain>
// resource.
const extraTerrainResourceReason = "Worldographer keeps a resource only in 0..100: it will not open a file " +
	"with 128 or more (app check, #124) and saves a tile record's 101..127 as 100 (app check, #122); " +
	"wxx reads it as the app is assumed to (#124)"

// clampedValues converts the codec's record of clamps into the public report.
func clampedValues(clamps []v1_06.Clamp_t) []ClampedValue_t {
	if len(clamps) == 0 {
		return nil
	}
	out := make([]ClampedValue_t, 0, len(clamps))
	for _, c := range clamps {
		out = append(out, ClampedValue_t{Path: c.Path, Field: c.Field, Was: c.Was, Now: c.Now, Reason: extraTerrainResourceReason})
	}
	return out
}
