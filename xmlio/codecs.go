// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio

import (
	"fmt"

	"github.com/maloquacious/wxx/xmlio/internal/appver"
	"github.com/maloquacious/wxx/xmlio/internal/v0_77"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// codecAppSets returns every codec's declaration of the application versions it
// accepts, in codec-version order.
//
// The dispatcher is the only place that sees all of the codecs at once, so it is
// the only place that can check a property ACROSS them. The sets themselves are
// not restated here: each is read from the codec that owns it, so this table
// cannot drift from what the codecs actually enforce.
//
// Adding a codec means adding it here. A codec missing from this list is not
// checked against the others, which is the one way the disjointness guarantee
// below can be lost.
func codecAppSets() []appver.Set_t {
	return []appver.Set_t{
		v0_77.AcceptedApps(),
		v1_06.AcceptedApps(),
	}
}

// init verifies that every codec's declaration is valid on its own terms and that
// no application version is accepted by more than one codec, panicking if either
// fails. Both checks are appver.VerifyDisjoint's; see its doc for why the
// per-declaration one is there.
//
// This is issue #41 requirement 4, checked over what the CODECS declare. The
// registry's duplicate-application-version guard is not this check: it stops one
// version naming two releases, and it would happily pass a table in which two
// codecs both claimed "2.06" -- the table names the version once, so nothing
// there is duplicated. The ambiguity would surface as whichever codec the schema
// lookup happened to reach, i.e. as a silently wrong file.
//
// The codec set is a constant of the program, so an overlap in it is a
// programming error rather than a runtime condition a caller could handle, and
// failing at load makes it unmissable. The check itself lives in appver rather
// than here so that it stays testable: a test can hand appver.VerifyDisjoint a
// deliberately overlapping table and inspect the error, which it could not do
// with a panic in init. This mirrors NewRegistry, for the same reason.
func init() {
	if err := appver.VerifyDisjoint(codecAppSets()...); err != nil {
		panic(fmt.Sprintf("xmlio: codec application versions: %v", err))
	}
}
