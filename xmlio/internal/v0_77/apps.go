// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v0_77

import "github.com/maloquacious/wxx/xmlio/internal/appver"

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
// may never reach a file.
var acceptedApps = appver.Set_t{
	Codec:  "v0_77",
	Schema: "",
	Apps:   []string{"1.73", "1.74", "1.77"},
}

// AcceptedApps returns this codec's declaration of the application versions it
// accepts and the schema it writes. The returned set is the caller's own copy.
func AcceptedApps() appver.Set_t {
	return acceptedApps.Clone()
}
