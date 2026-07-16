// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package codec defines what a codec IS: the parse/emit pair for one schema
// (ADR 0004 Decision 4), plus the declaration it states about itself.
//
// It holds the INTERFACE and nothing else. It used to hold the schema -> codec
// selector as well, which issue #45 Decision 8 removes: the application version
// selects the encoder now, the registry that does it is xmlio's, and a schema
// selects nothing. What survives is the one question this package was always
// really answering -- what does the dispatcher get to hold? -- and the answer is
// this interface.
//
// It is internal to xmlio because of issue #41 requirement 5: the DISPATCHER
// picks the encoder, and a caller may name only an application version. A public
// symbol that hands back an encoder is exactly the reach requirement 5 denies --
// #41 documents what it buys, a classic map emitted through the W2025 codec as
// 18,006 bytes of W2025 XML declaring release="" version="1.77" schema="", which
// then re-decodes silently as classic. An EXPORTED interface here is not that
// reach: nothing outside the xmlio subtree can name it, so no public function can
// return one.
//
// Exporting it rather than unexporting the interface inside xmlio is deliberate,
// and it is the same reason the selector lived here: xmlio's tests are an
// EXTERNAL test package (package xmlio_test) and could not see an unexported
// xmlio symbol. Go's internal rule is directory-based rather than
// package-name-based, so those tests -- which sit physically inside xmlio/ -- may
// import this package, while nothing outside the xmlio subtree can. That is
// requirement 5's "test units may choose the encoder" exception, preserved by
// construction and with no escape hatch.
package codec

import (
	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
)

// Codec parses and emits exactly one schema, and declares what it accepts and
// what it writes.
//
// It is an interface rather than a struct of function fields because a Go
// PACKAGE cannot implement an interface: the codecs are packages, so each has to
// export a VALUE for the dispatcher to hold. The struct-of-funcs shape this
// replaces could hold half a codec -- a nil Encode was constructible and
// therefore had to be checked for at load -- whereas a type either has the method
// set or does not compile. The check that guarded against the half-built pair is
// not rescued anywhere, because it is now a compile error.
//
// AcceptedApps is on the interface, not alongside it, for the same reason Encode
// is: the declaration is the codec's own knowledge (see appver), and the
// dispatcher must be able to ask any codec for it without knowing which one it
// holds. That question -- "which application versions do you write?" -- is the
// whole of what the registry is built from (issue #45 Decision 8).
type Codec interface {
	// Decode parses one schema's XML into the Map_t superset.
	Decode(input []byte) (*wxx.Map_t, error)

	// Encode emits a Map_t as one schema's XML, as the application version app.
	//
	// app is verbatim map/@version. The codec verifies it against the set it
	// declares and rejects one it does not accept, so an encoder can never be
	// talked into writing a release that never existed (issue #41 requirement 3).
	// It is also the identity the codec WRITES: what it checks and what it emits
	// must be one input, which is issue #45. It has to be passed rather than
	// derived because a codec's schema cannot tell it which of the application
	// versions sharing that schema the caller meant.
	Encode(m *wxx.Map_t, app string) ([]byte, error)

	// AcceptedApps returns the codec's declaration: the application versions it
	// accepts, the map/@release each writes, the schema it writes, and the XML
	// declaration its files open with. The returned set is the caller's own copy.
	AcceptedApps() appver.Set_t
}
