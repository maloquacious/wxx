// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v0_77

import (
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
)

// Codec_t is this package's codec as a VALUE, which is what the dispatcher holds.
//
// A Go package cannot implement an interface, and the codecs are packages, so
// "the v0_77 codec" is not something the dispatcher can hold without this. It is
// zero-size and stateless: every method forwards to the package-level function
// that is the real implementation, which stays exported because the test units
// requirement 5 admits call it directly (see xmlio/chimera_test.go).
//
// It carries the declaration alongside decode and encode because the registry is
// built by ASKING each codec what it accepts (issue #45 Decision 8): the mapping
// application version -> encoder exists in exactly one place, here, and the
// dispatcher reads it rather than restating it.
type Codec_t struct{}

// Decode parses classic XML into the Map_t superset.
func (Codec_t) Decode(input []byte) (*wxx.Map_t, error) { return Decode(input) }

// Encode emits a Map_t as classic XML, as the application version app.
func (Codec_t) Encode(m *wxx.Map_t, app string) ([]byte, error) { return Encode(m, app) }

// AcceptedApps returns this codec's declaration. See acceptedApps.
func (Codec_t) AcceptedApps() appver.Set_t { return AcceptedApps() }

// acceptedApps declares what this codec accepts and what it writes.
//
// Despite the package path, this is NOT the 1.77-only codec: classic 1.73, 1.74
// and 1.77 share an identical element vocabulary, so one schema and therefore one
// codec serves all three. The path is the CODEC version ("0.77"), and ".77" is a
// mnemonic from the last classic application version -- it is not the set. The set
// is here, it is the codec's own knowledge, and it must be read here rather than
// inferred from the path.
//
// Schema is "" because classic files state no map/@schema at all. That absence is
// the identity (ADR 0004 Decision 2); the codec must never invent a schema string
// for it, and "0.77" in particular is a codec version that appears on no disk and
// may never reach a file. Every app's Release is "" for the same reason: a classic
// file states no map/@release either, and the declaration says so per application
// version because the mapping is derived rather than constant (issue #45 Decision
// 5) -- these three agreeing is a fact about classic, not a property of codecs.
//
// XMLVersion is "1.0": every classic build opens its files with that declaration.
var acceptedApps = appver.Set_t{
	Codec:      "v0_77",
	Schema:     "",
	XMLVersion: "1.0",
	Apps: []appver.App_t{
		{Version: "1.73", Release: ""},
		{Version: "1.74", Release: ""},
		{Version: "1.77", Release: ""},
	},
}

// AcceptedApps returns this codec's declaration of the application versions it
// accepts and the schema it writes. The returned set is the caller's own copy.
func AcceptedApps() appver.Set_t {
	return acceptedApps.Clone()
}
