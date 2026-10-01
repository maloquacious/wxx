// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio

import (
	"errors"
	"fmt"

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

// downgradeLoss reports what m carries that the target SCHEMA cannot express, and
// errors if the loss is one the encoder cannot honestly describe.
//
// targetSchema is the schema the target codec writes, verbatim map/@schema
// ("1.06"). It is asked of the codec rather than of a registry entry because the
// schema is a byte the encoder writes and the encoder owns it (issue #45).
//
// The schema is the right axis here even though it no longer selects the codec.
// What can be expressed is a property of the FORMAT, not of the build that wrote
// it: two application versions sharing a schema lose exactly the same things, so
// keying this on the application version would restate one fact per build and
// invite them to disagree.
//
// The question is about the TARGET's expressiveness, not about which file m was
// read from, so there is no source-schema argument here, and a map that never held
// a field the target lacks reports nothing.
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
// used to hold, was removed with its codec (issue #103), and the one schema left
// expresses everything Map_t models, so every supported target reports no loss.
// The principle outlives the downgrade it was written for -- always tell the user
// what they lose -- and the next entries are W2025-to-W2025: a release
// registered on a schema that cannot express something Map_t models gets its
// own arm here, reporting what that target drops. Issues #92 and #73 registered
// 2.07 and 2.08 on schema 1.06 and found no such loss, so they added none.
//
// A schema with no arm is an ERROR, not "no loss". Losslessness is a claim made
// per schema, by an arm that says so; a schema nobody has inventoried has made no
// such claim, and treating its silence as one is how a new codec would quietly
// report every encode as lossless. Every codec the registry holds writes a schema
// with an arm (TestNoLossOnSameReleaseTargets encodes through each of them), so
// the error is reachable only by registering a codec without inventorying it --
// which it then refuses on every encode, loudly, until someone does.
func downgradeLoss(m *wxx.Map_t, targetSchema string) ([]DroppedFeature_t, error) {
	switch targetSchema {
	case "1.06":
		// W2025 schema 1.06 expresses everything Map_t models: Map_t is built
		// from it. Encoding to it is therefore not a downgrade, which
		// TestNoLossOnSameReleaseTargets pins for every supported release.
		return nil, nil
	}
	return nil, errors.Join(wxx.ErrUnsupportedSchemaVersion, fmt.Errorf("schema %q: no loss inventory for this target, so the encoder cannot say what it would lose", targetSchema))
}
