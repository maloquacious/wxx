// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// DroppedFeature_t is one thing the source map carries that the target release
// cannot express (ADR 0004 Decision 7).
//
// It is a struct rather than a string so that a caller learns WHAT was lost and
// WHY without parsing prose: Path and Field name the thing on disk and in the
// model, Detail says what THIS map gives up in concrete values, and Reason says
// why the target cannot hold it. A report line is one formatting of that; a
// caller that wants to key off Path, or count layers, or surface only the
// Details to a user, can.
type DroppedFeature_t struct {
	// Path is the on-disk element/attribute path the loss occurs at, in the same
	// vocabulary the round-trip audit harness prints ("map/maplayer/@opacity").
	// It is the stable identifier: Detail varies per map, Path does not.
	Path string

	// Field is the Map_t field that holds the content being dropped
	// ("Map_t.MapLayers[].Opacity"). It is what a caller reads to recover the
	// data the file will not carry.
	Field string

	// Detail is what THIS map actually loses -- values and counts, not a
	// restatement of Path. It is the difference between "opacity is dropped" and
	// "opacity is dropped from 8 layers, all at 1".
	Detail string

	// Reason is why the target cannot express the content, cited to the format
	// rather than to the codec. A feature the target's FORMAT has no room for is
	// a downgrade loss; a feature the target's format has room for but our
	// encoder does not write yet is a codec gap and does not belong here (see
	// classicDowngradeLoss).
	Reason string
}

// String renders one dropped feature as a report line.
func (d DroppedFeature_t) String() string {
	return fmt.Sprintf("%s (%s): %s -- %s", d.Path, d.Field, d.Detail, d.Reason)
}

// downgradeLoss reports what m carries that target cannot express, and errors if
// the loss is one the encoder cannot honestly describe.
//
// The question is about the TARGET's expressiveness, not about which file m was
// read from: a map holding a W2025-native field loses it when written as classic
// however it was built. So there is no source-schema argument here, and a map
// that never held the field reports nothing -- which is why encoding classic as
// classic, or W2025 as W2025, reports no loss at all.
//
// THE LOSS CONTRACT (settled under #32; ADR 0004 Decision 7 left it open):
//
//   - A MODELED feature -- one Map_t understands well enough for the encoder to
//     enumerate precisely what is lost -- is reported through EncoderDiagnostics
//     and the encode SUCCEEDS.
//   - An UNMODELED STUB -- content Map_t carries only as verbatim InnerXML and
//     does not understand -- is a hard ERROR. The encoder cannot honestly say
//     what such a loss costs, only that something it never understood will not
//     survive, so it refuses rather than silently discarding it.
//
// The rule is that silence is acceptable only where the loss is fully enumerable
// and documented. Rejected alternatives: diagnostics-only for both (diagnostics
// are opt-in, and ADR 0004 flags silent loss as the worse failure -- a caller who
// never asks would lose stub content without a word), and erroring on any lossy
// encode unless the caller passes WithAllowLossy (too blunt: it makes the common,
// fully-enumerated downgrade as loud as the one we genuinely cannot describe).
//
// Consequence, and it is intended: when a feature moves from stub to modeled
// (#11 is doing exactly this), its hard error BECOMES a diagnostic. Modeling it
// is what earns the encoder the right to be quiet about it, because only then can
// it say what was lost.
func downgradeLoss(m *wxx.Map_t, target *Release_t) ([]DroppedFeature_t, error) {
	if target.Schema != nil {
		// Every non-classic supported schema is W2025 1.06, which expresses
		// everything Map_t models: Map_t is the superset of the two supported
		// schemas (ADR 0004 Decision 6) and every field classic-only content
		// occupies exists in W2025 too. Encoding to it is therefore not a
		// downgrade, which TestNoLossOnSameReleaseTargets pins for every
		// supported release.
		//
		// This is a claim about the schemas that exist today, not a law. A future
		// schema that cannot express something Map_t models needs its own arm
		// here; adding the registry entry alone would silently claim the target
		// is lossless.
		return nil, nil
	}
	return classicDowngradeLoss(m)
}

// classicDowngradeLoss reports what m carries that the implicit legacy (classic)
// schema cannot express.
//
// EVERY ENTRY IS EVIDENCED. The inventory was built by running the round-trip
// audit harness (roundtrip_2017_test.go's xmlAggregate/computeLoss) over a
// decoded W2025 2.06 fixture encoded through the classic target and diffed
// against the W2025 original, NOT from memory -- ADR 0003's revision history
// records what an unverified claim costs. TestClassicDowngradeLossInventory
// re-runs that diff and holds this function to it.
//
// Two classes of loss the harness reports are deliberately NOT here, because
// neither is a downgrade:
//
//   - TARGET IDENTITY. map/@version changes 2.06 -> 1.77 and map/@release and
//     map/@schema disappear. That is Release_t.identify writing the release the
//     caller asked for; a classic file states no release and no schema. Nothing
//     is lost -- the file is being told what it now is.
//   - CLASSIC CODEC GAPS. map/mapkey/@viewlevel is altered and
//     map/informations/information and map/configuration/text-config/labelstyle
//     are dropped -- but the classic FORMAT has room for all three (RelaxNG
//     schema/utf-8-xml.rnc defines them, classic samples carry them). Our classic
//     ENCODER does not write them yet, which h2017v1/COVERAGE.md documents and
//     the classic round-trip harness proves by losing the same three on a
//     classic -> classic trip. They cost the caller data, but they are this
//     codec's gaps and would not be fixed by targeting differently. Reporting
//     them as downgrade loss would blame the format for our encoder.
//
// The residual -- what the W2025 -> classic diff shows that a classic -> classic
// and a W2025 -> W2025 diff do not -- is the inventory below.
func classicDowngradeLoss(m *wxx.Map_t) ([]DroppedFeature_t, error) {
	var dropped []DroppedFeature_t

	// map/@hScrollbarPos, map/@vScrollbarPos -- the classic <map> element states
	// no scrollbar position (RelaxNG defines 24 map attributes, none of them
	// these; h2017v1's XMLSchema struct has no field for either).
	//
	// Gated on non-zero because Map_t models these as plain float64: absent and
	// "0.0" decode identically, so a zero cannot be reported as a loss without
	// inventing one. Both tracked 2.06 fixtures carry 0.0, so this entry is
	// LATENT on them -- real by format, unexercised by the samples, in the same
	// sense h2017v1/COVERAGE.md means "latent-by-code". Its test synthesizes a
	// non-zero value rather than pretending a fixture proves it.
	const scrollbarReason = "the classic <map> element states no scrollbar position (schema/utf-8-xml.rnc defines 24 map attributes, none of them a scrollbar position)"
	if m.HScrollbarPos != 0 {
		dropped = append(dropped, DroppedFeature_t{
			Path:   "map/@hScrollbarPos",
			Field:  "Map_t.HScrollbarPos",
			Detail: fmt.Sprintf("horizontal scrollbar position %s is dropped", floatDetail(m.HScrollbarPos)),
			Reason: scrollbarReason,
		})
	}
	if m.VScrollbarPos != 0 {
		dropped = append(dropped, DroppedFeature_t{
			Path:   "map/@vScrollbarPos",
			Field:  "Map_t.VScrollbarPos",
			Detail: fmt.Sprintf("vertical scrollbar position %s is dropped", floatDetail(m.VScrollbarPos)),
			Reason: scrollbarReason,
		})
	}

	// map/maplayer/@opacity -- the classic <maplayer> element has only @name and
	// @isVisible (RelaxNG lines 63-66; h2017v1.MapLayer_t has the same two
	// fields). Note this is NOT "classic has no layers": both formats carry
	// <maplayer> elements, and classic re-emits every one of them. Only the
	// per-layer opacity is lost.
	//
	// Gated on non-zero for the same reason as the scrollbars: a classic-decoded
	// map has Opacity == 0 on every layer, so a bare "has layers" test would
	// report this loss on a classic -> classic encode, where nothing is lost.
	if names := layersWithOpacity(m); len(names) > 0 {
		dropped = append(dropped, DroppedFeature_t{
			Path:   "map/maplayer/@opacity",
			Field:  "Map_t.MapLayers[].Opacity",
			Detail: fmt.Sprintf("opacity is dropped from %d of %d map layer(s): %s", len(names), len(m.MapLayers), strings.Join(names, ", ")),
			Reason: "the classic <maplayer> element states only @name and @isVisible (schema/utf-8-xml.rnc); classic layers have no opacity",
		})
	}

	// map/configuration/shape-config/shapestyle/@lineCap and @lineJoin -- the
	// classic <shapestyle> element has 27 attributes and neither of these
	// (RelaxNG lines 194-222; h2017v1.ShapeStyle_t has no LineCap/LineJoin).
	//
	// The near-miss worth naming: classic DOES define @lineCap and @lineJoin --
	// on <shape>, a different element (RelaxNG lines 157-158), which is why a
	// grep for the attribute name in the classic schema finds it. The style does
	// not carry them, so a shapestyle's caps and joins have nowhere to go.
	//
	// Gated on non-empty: classic decode leaves both strings "".
	dropped = append(dropped, shapeStyleLineLoss(m)...)

	// map/blurTerrainBG -- a W2025 top-level element the classic format does not
	// define at all (absent from the RelaxNG schema; schema/README.md
	// independently flags it as a verified W2025 delta). Modeled as a pointer, so
	// non-nil is exactly "the source carried one".
	if m.BlurTerrainBG != nil {
		b := m.BlurTerrainBG
		dropped = append(dropped, DroppedFeature_t{
			Path:  "map/blurTerrainBG",
			Field: "Map_t.BlurTerrainBG",
			Detail: fmt.Sprintf("terrain-background blur settings are dropped (blur=%t topBleed=%s bottomBleed=%s randomness=%s blurStart=%s blurEnd=%s)",
				b.Blur, floatDetail(b.TopBleed), floatDetail(b.BottomBleed), floatDetail(b.Randomness), floatDetail(b.BlurStart), floatDetail(b.BlurEnd)),
			Reason: "the classic format defines no <blurTerrainBG> element",
		})
	}

	// map/extraTerrain -- the ADR 0004 terrain-layers loss, and the one entry that
	// ERRORS instead of reporting.
	//
	// W2025 binds terrain to a named layer per hex
	// (<extraTerrain><mapLayer name="..."><terrainAndLocation location="x,y"/>);
	// classic binds mapLayer to features, labels and shapes but never to tiles,
	// so all classic terrain sits on one hard-coded layer and a per-hex layer
	// assignment collapses. That much is a genuine downgrade loss.
	//
	// What makes it an error rather than a diagnostic is that Map_t models this
	// element ONLY as opaque InnerXML (#11 has not reached it): the bytes
	// round-trip 2025 -> 2025 intact, yet nothing in the model understands them.
	// The encoder therefore cannot enumerate what dropping them costs -- it cannot
	// say how many hexes, on which layers, with what terrain -- and under the loss
	// contract it must refuse rather than discard content it cannot describe.
	// When #11 models terrainAndLocation, this becomes a diagnostic like the rest.
	if m.ExtraTerrain != nil && !isEmptyInnerXML(m.ExtraTerrain.InnerXML) {
		inner := strings.TrimSpace(m.ExtraTerrain.InnerXML)
		return nil, errors.Join(wxx.ErrUnmodeledStubLoss, fmt.Errorf(
			"map/extraTerrain (Map_t.ExtraTerrain.InnerXML): the classic format defines no <extraTerrain> element, and this map carries %d bytes of it that the model holds only as an opaque stub, so the encoder cannot describe what dropping them would cost: %s",
			len(m.ExtraTerrain.InnerXML), stubExcerpt(inner)))
	}

	return dropped, nil
}

// isEmptyInnerXML reports whether a verbatim InnerXML stub holds nothing.
//
// InnerXML is the raw bytes between the element's tags, so an empty container is
// not "" but whatever the writer's pretty-printer put there: the blank 2.06
// fixture's <extraTerrain> carries "\n", one newline. A "" test would call that
// container populated and error on a file that loses nothing.
//
// TrimSpace is exactly the right test, and not an approximation. Whitespace
// BETWEEN elements is insignificant formatting, and no XML content can hide in
// it: an element needs a '<', an attribute lives inside a tag, and a text node
// made only of whitespace has no content to lose. So an all-whitespace InnerXML
// means the container has no children and no text -- dropping it costs nothing.
// Conversely a single non-whitespace byte means at least one child or text node
// the model never understood, which is precisely what the encoder must not
// silently discard.
func isEmptyInnerXML(inner string) bool {
	return strings.TrimSpace(inner) == ""
}

// stubExcerpt renders opaque stub content for an error message, truncated so a
// large stub cannot turn one error into a wall of XML. The content is echoed
// because the caller cannot ask the model what was in it -- that is what makes it
// a stub -- so the bytes are the only description available.
func stubExcerpt(inner string) string {
	const max = 160
	inner = strings.Join(strings.Fields(inner), " ")
	if len(inner) > max {
		return strconv.Quote(inner[:max]) + " (truncated)"
	}
	return strconv.Quote(inner)
}

// layersWithOpacity returns "name"=opacity for every map layer carrying a
// non-zero opacity, in map order.
func layersWithOpacity(m *wxx.Map_t) []string {
	var names []string
	for _, l := range m.MapLayers {
		if l == nil || l.Opacity == 0 {
			continue
		}
		names = append(names, fmt.Sprintf("%q=%s", l.Name, floatDetail(l.Opacity)))
	}
	return names
}

// shapeStyleLineLoss returns the @lineCap/@lineJoin entries for m's shape styles,
// one per attribute, naming the styles that carry it.
func shapeStyleLineLoss(m *wxx.Map_t) []DroppedFeature_t {
	if m.Configuration == nil || m.Configuration.ShapeConfig == nil {
		return nil
	}
	var caps, joins []string
	for _, s := range m.Configuration.ShapeConfig.ShapeStyles {
		if s == nil {
			continue
		}
		if s.LineCap != "" {
			caps = append(caps, fmt.Sprintf("%q=%s", s.Name, s.LineCap))
		}
		if s.LineJoin != "" {
			joins = append(joins, fmt.Sprintf("%q=%s", s.Name, s.LineJoin))
		}
	}
	const reason = "the classic <shapestyle> element states neither @lineCap nor @lineJoin (schema/utf-8-xml.rnc defines 27 shapestyle attributes, without them); classic defines both on <shape>, a different element, so a style's caps and joins have nowhere to go"
	var out []DroppedFeature_t
	if len(caps) > 0 {
		out = append(out, DroppedFeature_t{
			Path:   "map/configuration/shape-config/shapestyle/@lineCap",
			Field:  "Map_t.Configuration.ShapeConfig.ShapeStyles[].LineCap",
			Detail: fmt.Sprintf("lineCap is dropped from %d shape style(s): %s", len(caps), strings.Join(caps, ", ")),
			Reason: reason,
		})
	}
	if len(joins) > 0 {
		out = append(out, DroppedFeature_t{
			Path:   "map/configuration/shape-config/shapestyle/@lineJoin",
			Field:  "Map_t.Configuration.ShapeConfig.ShapeStyles[].LineJoin",
			Detail: fmt.Sprintf("lineJoin is dropped from %d shape style(s): %s", len(joins), strings.Join(joins, ", ")),
			Reason: reason,
		})
	}
	return out
}

// floatDetail renders a float for a Detail string. It is display only -- no
// Dotted and nothing else bound for disk is ever rendered here (ADR 0004
// Decision 1).
func floatDetail(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
