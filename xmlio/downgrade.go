// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
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
	// Path is the on-disk element/attribute path the loss occurs at, as
	// local-names joined with '/' ("map/maplayer/@opacity").
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
	// encoder does not write yet is a codec gap and does not belong here.
	Reason string
}

// String renders one dropped feature as a report line.
func (d DroppedFeature_t) String() string {
	return fmt.Sprintf("%s (%s): %s -- %s", d.Path, d.Field, d.Detail, d.Reason)
}

// downgradeLoss reports what m carries that the target cannot express, and
// errors if the loss is one the encoder cannot honestly describe.
//
// The target is named twice, by targetSchema and targetApp, and both are the
// encoder's own bytes rather than anything read from m (issue #45).
// targetSchema is the schema the target codec writes, verbatim map/@schema
// ("1.06"), asked of the codec's declaration. targetApp is the application
// version being written, verbatim map/@version ("2.08"), already resolved by the
// registry to that codec.
//
// The SCHEMA picks the arm, and the APPLICATION VERSION decides what the arm
// reports. This used to be keyed on the schema alone, on the argument that what
// can be expressed is a property of the format, so two application versions
// sharing a schema lose exactly the same things. Issue #161 proved that false:
// Worldographer 2.09 writes schema 1.06, as 2.08 does, and adds label/@lineSpacing,
// label/@textAlignment and labelstyle/@lineSpacing, which 2.08 never writes. The
// schema number did not change, so it cannot tell the two targets apart, and an
// inventory keyed on it alone would report a 2.09 map written as 2.08 as
// lossless while the encoder drops all three. What a target writes is a property
// of the BUILD; the schema says only which codec writes it, and so which arm
// owns the question.
//
// The arm does not restate which builds write what. It asks the codec that
// writes the schema (v1_06.WritesLabelTextLayout), which is the same rule the
// encoder obeys when it decides to write or omit the attributes, so the report
// and the bytes cannot disagree. It asks the codec package directly, as the
// decoder does, rather than through codec.Codec: the fact is about schema 1.06's
// labels, and an interface method would make every codec answer a question that
// belongs to one.
//
// The question is about the TARGET's expressiveness, not about which file m was
// read from, so there is no source-version argument here, and a map that never
// held a field the target lacks reports nothing.
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
// Consequence, and it is intended: when a feature moves from stub to modeled,
// its hard error BECOMES a diagnostic. Modeling it is what earns the encoder the
// right to be quiet about it, because only then can it say what was lost. #34
// did exactly this for <extraTerrain>, the last stub, so today nothing produces
// wxx.ErrUnmodeledStubLoss. The rule stands for the next stub.
//
// Today there is one format. The classic schema, whose inventory this function
// used to hold, was removed with its codec (issue #103). The losses left are
// W2025-to-W2025: the 2.09 label text-layout attributes, written to a 1.06
// target that predates them (labelTextLayoutLoss). Issues #92, #73 and #162
// registered 2.07, 2.08 and 2.09 and found nothing else 2.06 lacks.
//
// A schema with no arm is an ERROR, not "no loss". Losslessness is a claim made
// per schema, by an arm that says so; a schema nobody has inventoried has made no
// such claim, and treating its silence as one is how a new codec would quietly
// report every encode as lossless. Every codec the registry holds writes a schema
// with an arm (TestNoLossOnSameReleaseTargets encodes through each of them), so
// the error is reachable only by registering a codec without inventorying it --
// which it then refuses on every encode, loudly, until someone does.
func downgradeLoss(m *wxx.Map_t, targetSchema, targetApp string) ([]DroppedFeature_t, error) {
	switch targetSchema {
	case "1.06":
		// W2025 schema 1.06 expresses everything Map_t models, Map_t being built
		// from it, except that the builds before 2.09 write no label
		// text-layout attributes.
		writes, err := v1_06.WritesLabelTextLayout(targetApp)
		if err != nil {
			return nil, err
		}
		if writes {
			return nil, nil
		}
		return labelTextLayoutLoss(m, targetApp), nil
	}
	return nil, errors.Join(wxx.ErrUnsupportedSchemaVersion, fmt.Errorf("schema %q: no loss inventory for this target, so the encoder cannot say what it would lose", targetSchema))
}

// labelTextLayoutLoss reports the 2.09 label text-layout attributes m carries,
// for a target app that does not write them (issue #161). The encoder omits
// them for such a target; this says what that costs.
//
// Only values that change how a label is drawn are reported, so that a real
// loss is not buried under the defaults every 2.09 map states:
//
//   - lineSpacing, on a label or a label style, when present and not 1.0. 2.09
//     omits it at 1.0, so a stated 1.0 is a default and its loss changes
//     nothing.
//   - textAlignment, on a label, when present and not CENTER. 2.09 writes it on
//     every label and writes CENTER by default; reporting every CENTER would
//     report a loss on every label of every 2.09 map.
//
// That CENTER and 1.0 are what the older builds draw is the maintainer's ruling
// (2026-11-18), not an app check.
//
// There is one entry per path: standalone labels and feature labels are
// separate paths and separate entries, and each entry's Detail counts the
// labels and gives their values.
func labelTextLayoutLoss(m *wxx.Map_t, app string) []DroppedFeature_t {
	var labelSpacing, labelAlign, featureSpacing, featureAlign []string
	for _, l := range m.Labels {
		if l == nil {
			continue
		}
		if l.LineSpacing != nil && *l.LineSpacing != 1.0 {
			labelSpacing = append(labelSpacing, doubleString(*l.LineSpacing))
		}
		if l.TextAlignment != "" && l.TextAlignment != "CENTER" {
			labelAlign = append(labelAlign, l.TextAlignment)
		}
	}
	for _, f := range m.Features {
		if f == nil || f.Label == nil {
			continue
		}
		if f.Label.LineSpacing != nil && *f.Label.LineSpacing != 1.0 {
			featureSpacing = append(featureSpacing, doubleString(*f.Label.LineSpacing))
		}
		if f.Label.TextAlignment != "" && f.Label.TextAlignment != "CENTER" {
			featureAlign = append(featureAlign, f.Label.TextAlignment)
		}
	}
	var styleSpacing []string
	if m.Configuration != nil && m.Configuration.TextConfig != nil {
		for _, s := range m.Configuration.TextConfig.LabelStyles {
			if s != nil && s.LineSpacing != nil && *s.LineSpacing != 1.0 {
				styleSpacing = append(styleSpacing, fmt.Sprintf("%s at %s", s.Name, doubleString(*s.LineSpacing)))
			}
		}
	}

	spacingReason := fmt.Sprintf("Worldographer %s writes no lineSpacing; 2.09 is the first build that does (issue #161)", app)
	alignReason := fmt.Sprintf("Worldographer %s writes no textAlignment; 2.09 is the first build that does (issue #161)", app)
	var out []DroppedFeature_t
	add := func(path, field, detail, reason string) {
		out = append(out, DroppedFeature_t{Path: path, Field: field, Detail: detail, Reason: reason})
	}
	if len(labelSpacing) > 0 {
		add("map/labels/label/@lineSpacing", "Map_t.Labels[].LineSpacing", countedDetail("label", "at", labelSpacing), spacingReason)
	}
	if len(labelAlign) > 0 {
		add("map/labels/label/@textAlignment", "Map_t.Labels[].TextAlignment", countedDetail("label", "aligned", labelAlign), alignReason)
	}
	if len(featureSpacing) > 0 {
		add("map/features/feature/label/@lineSpacing", "Map_t.Features[].Label.LineSpacing", countedDetail("feature label", "at", featureSpacing), spacingReason)
	}
	if len(featureAlign) > 0 {
		add("map/features/feature/label/@textAlignment", "Map_t.Features[].Label.TextAlignment", countedDetail("feature label", "aligned", featureAlign), alignReason)
	}
	if len(styleSpacing) > 0 {
		add("map/configuration/text-config/labelstyle/@lineSpacing", "Map_t.Configuration.TextConfig.LabelStyles[].LineSpacing", countedDetail("label style", "", styleSpacing), spacingReason)
	}
	return out
}

// countedDetail renders n values as a Detail: "1 label at 2.0", "3 labels at
// 2.0", or "3 labels: 2 at 2.0, 1 at 1.5" when they differ, values in the order
// they first appear. With an empty verb the values are self-describing and are
// listed rather than counted: "1 label style: City at 2.0".
func countedDetail(noun, verb string, values []string) string {
	nouns := noun
	if len(values) != 1 {
		nouns += "s"
	}
	if verb == "" {
		return fmt.Sprintf("%d %s: %s", len(values), nouns, strings.Join(values, ", "))
	}
	var order []string
	counts := map[string]int{}
	for _, v := range values {
		if counts[v] == 0 {
			order = append(order, v)
		}
		counts[v]++
	}
	if len(order) == 1 {
		return fmt.Sprintf("%d %s %s %s", len(values), nouns, verb, order[0])
	}
	parts := make([]string, 0, len(order))
	for _, v := range order {
		parts = append(parts, fmt.Sprintf("%d %s %s", counts[v], verb, v))
	}
	return fmt.Sprintf("%d %s: %s", len(values), nouns, strings.Join(parts, ", "))
}

// doubleString spells a value for a Detail as the file would ("2.0", "1.5"):
// the shortest exact decimal, with ".0" on an integral one. It is for reading,
// not for writing; the codec owns the on-disk spelling.
func doubleString(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eEnN") { // not 2.5, 1e+21, NaN, Inf
		s += ".0"
	}
	return s
}
