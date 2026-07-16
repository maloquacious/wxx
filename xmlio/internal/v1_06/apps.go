// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
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
// It forwards Encode but not Decode, because the dispatcher only ever asks a
// codec to encode: xmlio's decoder reads the file's own map/@release and calls
// this package's Decode function directly, so a forwarding method had no caller.
// See codec.Codec.
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

// acceptedApps declares what this codec accepts and what it writes.
//
// 2.06 is the supported W2025 baseline, the first post-beta build, and today it
// is the only application version on schema 1.06 that this codec accepts. The set
// is the codec's own knowledge and must be read here rather than inferred from
// the package path: the path is the CODEC version, which by convention matches
// the schema the file states, and it says nothing about which builds wrote that
// schema. A later build on the same schema is added here, not by adding a
// package.
//
// 2.06 writes release="2025". That is stated per application version rather than
// as a constant of the codec because map/@release is DERIVED from the application
// version (issue #45 Decision 5): a later build on schema 1.06 shipped under a
// different label writes a different map/@release and is added as another App_t
// here. Today there is exactly one W2025 build, which makes the mapping look
// constant. It is not one, and collapsing it into one would hard-code away the
// relabel scenario ADR 0004 exists to keep expressible.
//
// XMLVersion is "1.1", the declaration 2.06 opens its files with.
var acceptedApps = appver.Set_t{
	Codec:      "v1_06",
	Schema:     "1.06",
	XMLVersion: "1.1",
	Apps: []appver.App_t{
		{Version: "2.06", Release: "2025"},
	},
}

// AcceptedApps returns this codec's declaration of the application versions it
// accepts and the schema it writes. The returned set is the caller's own copy.
func AcceptedApps() appver.Set_t {
	return acceptedApps.Clone()
}
