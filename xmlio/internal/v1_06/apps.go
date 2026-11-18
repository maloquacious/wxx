// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"fmt"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
)

// Codec_t is this package's codec as a VALUE, which is what the dispatcher holds.
//
// A Go package cannot implement an interface, and the codecs are packages, so
// "the v1_06 codec" is not something the dispatcher can hold without this. It is
// zero-size and stateless: every method forwards to the package-level function
// that is the real implementation, which stays exported because the test units
// requirement 5 admits call it directly (see xmlio/chimera_test.go).
//
// It forwards Encode and NewMap but not Decode, because the dispatcher only
// ever asks a codec to encode or to make a new map: xmlio's decoder reads the
// file's own map/@release and calls this package's Decode function directly, so
// a forwarding method had no caller. See codec.Codec.
//
// It carries the declaration alongside encode because the registry is built by
// ASKING each codec what it accepts (issue #45 Decision 8): the mapping
// application version -> encoder exists in exactly one place, here, and the
// dispatcher reads it rather than restating it.
type Codec_t struct{}

// Encode emits a Map_t as W2025 schema 1.06 XML, as the application version app.
func (Codec_t) Encode(m *wxx.Map_t, app string) ([]byte, error) { return Encode(m, app) }

// AcceptedApps returns this codec's declaration. See acceptedApps.
func (Codec_t) AcceptedApps() appver.Set_t { return AcceptedApps() }

// NewMap returns a new blank map with app's defaults. See NewMap.
func (Codec_t) NewMap(app string, columns, rows int, hexOrientation string) (*wxx.Map_t, error) {
	return NewMap(app, columns, rows, hexOrientation)
}

// acceptedApps declares what this codec accepts and what it writes.
//
// 2.06 is the first post-beta W2025 build, 2.07 the next build on the same
// schema (issue #92), 2.08 the one after it (issue #73), and 2.09 the one after
// that (issue #162); they are the application versions on schema 1.06 that this
// codec accepts. 2.09 is the newest, so it is CurrentApp and NewMap's default.
// The set is the codec's own knowledge and must be read here rather than
// inferred from the package path: the path is the CODEC version, which by
// convention matches the schema the file states, and it says nothing about
// which builds wrote that schema. 2.07, 2.08 and 2.09 show the rule working: a
// later build on the same schema is added here, not by adding a package. 2.09
// also writes the label text-layout attributes (issue #161), still under
// schema 1.06; which targets write them is WritesLabelTextLayout's business,
// not a property of the entries here.
//
// 2.06, 2.07, 2.08 and 2.09 all write release="2025". That is stated per
// application version rather than as a constant of the codec because
// map/@release is DERIVED from the application version (issue #45 Decision 5):
// a later build on schema 1.06 shipped under a different label writes a
// different map/@release and is added as another App_t here. Today every W2025
// build this codec accepts writes "2025", which makes the mapping look
// constant. It is not one, and collapsing it into one would hard-code away the
// relabel scenario ADR 0004 exists to keep expressible.
//
// XMLVersion is "1.1", the declaration 2.06, 2.07, 2.08 and 2.09 open their
// files with.
var acceptedApps = appver.Set_t{
	Codec:      "v1_06",
	Schema:     "1.06",
	XMLVersion: "1.1",
	Apps: []appver.App_t{
		{Version: "2.06", Release: "2025"},
		{Version: "2.07", Release: "2025"},
		{Version: "2.08", Release: "2025"},
		{Version: "2.09", Release: "2025"},
	},
}

// AcceptedApps returns this codec's declaration of the application versions it
// accepts and the schema it writes. The returned set is the caller's own copy.
func AcceptedApps() appver.Set_t {
	return acceptedApps.Clone()
}

// labelTextLayoutSince is the first application version that writes the label
// text-layout attributes: label/@lineSpacing, label/@textAlignment and
// labelstyle/@lineSpacing (issue #161; observed in
// testdata/2025-2.09-13x11-941577-labels.wxx). 2.06, 2.07 and 2.08 write none
// of them on any element.
var labelTextLayoutSince = mustParseDotted("2.09")

// WritesLabelTextLayout reports whether the application version app writes the
// label text-layout attributes (see labelTextLayoutSince). It is the ONE place
// that fact lives: the encoder asks it before writing the attributes, and
// xmlio's loss inventory asks it before reporting them as dropped, so the two
// can never disagree about which targets carry them.
//
// It is an ordered comparison against the first build that writes them rather
// than a property declared on each App_t, for two reasons. The attributes are a
// feature the application ADDED, and a later build on schema 1.06 is presumed
// to keep writing what 2.09 writes; a comparison states that once, where a
// per-app flag would have to be restated, correctly, on every entry
// acceptedApps ever gains, and an entry that forgot it would silently drop the
// attributes from its own files. And the answer is a function of the version
// alone, so it is also defined for a version this codec does not (yet) accept,
// which let the encoder and the loss inventory be written and tested for 2.09
// before 2.09 was registered (issue #162).
//
// It does not check that app is accepted: Encode does that, and the loss
// inventory receives only a target the registry resolved. app must be a dotted
// version (ParseDotted); one that is not is an error, never a guess.
func WritesLabelTextLayout(app string) (bool, error) {
	v, err := wxx.ParseDotted(app)
	if err != nil {
		return false, fmt.Errorf("version %q: %w", app, err)
	}
	c, err := v.Compare(labelTextLayoutSince)
	if err != nil {
		return false, err
	}
	return c >= 0, nil
}

// mustParseDotted parses a version this package states as a constant; a
// constant that does not parse is a programming error.
func mustParseDotted(s string) wxx.Dotted {
	v, err := wxx.ParseDotted(s)
	if err != nil {
		panic(fmt.Sprintf("v1_06: constant version %q: %v", s, err))
	}
	return v
}
