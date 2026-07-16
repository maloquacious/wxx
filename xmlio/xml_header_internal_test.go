// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// This is an INTERNAL test package (package xmlio, not xmlio_test), and it is the
// only one in this directory. It exists because verifyXMLVersions is unexported
// and must stay that way -- it is a load-time check over a compiled-in table, not
// API -- while a guard that cannot be watched to fail is worthless.
//
// Every other test here is external (package xmlio_test), which is deliberate:
// they hold the package to what a CALLER can see. This one is not testing the
// caller's view; it is testing a guard the caller cannot reach and should not be
// able to.
package xmlio

import (
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
)

// TestVerifyXMLVersionsRejectsAnUnknownDeclaration deliberately hands the check a
// codec declaring an XML version no header exists for, because a guard that
// cannot fail is worthless and the real declarations cannot be made wrong without
// editing the codecs.
//
// This is the guard NewRegistry ran over Release_t.XMLVersion, which issue #41
// called load-bearing and which issue #45 would have deleted along with the
// registry. It cannot live in appver with the declaration's other checks -- appver
// would have to import xmlio to see the header table, and xmlio imports appver --
// so it lives here, and this is where it is held to account.
//
// The check has to live outside init to be testable at all: init panics, and a
// panic cannot be inspected. This mirrors appver.VerifyDisjoint, and the reason is
// the same.
//
// This is one half of the guard's coverage and needs the other. It watches the
// CHECK reject a bad declaration; TestInitBuiltTheRegistry below watches init
// actually run it, and TestEncodeXMLHeaderFollowsRelease (encode_dispatch_test.go)
// pins the PROPERTY from the public side, at the byte level, over real fixtures.
// A guard can fail three ways -- wrong, never called, or right about the wrong
// thing -- and no one of these three tests catches more than one of them.
func TestVerifyXMLVersionsRejectsAnUnknownDeclaration(t *testing.T) {
	// Control: the two XML versions that exist are accepted. Guard against a
	// vacuous pass -- if verifyXMLVersions rejected everything, the cases below
	// would pass while proving nothing, and the real codecs would not load.
	ok := []appver.Set_t{
		{Codec: "classic-like", Schema: "", XMLVersion: "1.0", Apps: []appver.App_t{{Version: "1.77"}}},
		{Codec: "w2025-like", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}}},
	}
	if err := verifyXMLVersions(ok...); err != nil {
		t.Fatalf("verifyXMLVersions(real declarations): %v; the cases below would prove nothing", err)
	}

	for _, tc := range []struct {
		name    string
		sets    []appver.Set_t
		wantMsg string
	}{
		{
			// The failure the guard exists to stop: a codec that cannot say how its
			// files open. Encode time is too late to find that out -- by then a
			// caller is waiting for a file.
			name:    "an xml version no header exists for",
			sets:    []appver.Set_t{{Codec: "v0_77", Schema: "", XMLVersion: "1.2", Apps: []appver.App_t{{Version: "1.77"}}}},
			wantMsg: `"1.2"`,
		},
		{
			// "" is not a sentinel here, unlike Schema and Release where an empty
			// string is classic's meaningful identity. A file cannot open with no
			// declaration at all, so this is a codec that forgot to declare one.
			name:    "no xml version at all",
			sets:    []appver.Set_t{{Codec: "v0_77", Schema: "", XMLVersion: "", Apps: []appver.App_t{{Version: "1.77"}}}},
			wantMsg: "v0_77",
		},
		{
			// A schema version is not an XML version. The two are both dotted and
			// both on disk, which is exactly why a mix-up is plausible.
			name:    "a schema version, not an xml version",
			sets:    []appver.Set_t{{Codec: "v1_06", Schema: "1.06", XMLVersion: "1.06", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}}}},
			wantMsg: `"1.06"`,
		},
		{
			// The check is over EVERY set, not just the first: one bad codec in a
			// table of good ones must still stop the load.
			name: "one bad codec among good ones",
			sets: []appver.Set_t{
				{Codec: "classic-like", Schema: "", XMLVersion: "1.0", Apps: []appver.App_t{{Version: "1.77"}}},
				{Codec: "broken", Schema: "1.06", XMLVersion: "2.0", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}}},
			},
			wantMsg: "broken",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyXMLVersions(tc.sets...)
			if err == nil {
				t.Fatalf("verifyXMLVersions(%s) = nil, want an error", tc.name)
			}
			if !errors.Is(err, wxx.ErrUnknownXMLHeader) {
				t.Errorf("verifyXMLVersions(%s) error = %v, want it to wrap %v", tc.name, err, wxx.ErrUnknownXMLHeader)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("verifyXMLVersions(%s) error = %v, want it to name %s", tc.name, err, tc.wantMsg)
			}
		})
	}
}

// TestInitBuiltTheRegistry asserts that init ran the guards and built the index,
// rather than that the guards work -- which is the check above's job.
//
// It is the wiring assertion: verifyXMLVersions could be perfect and never
// called, which is precisely the "rescued check that lapsed silently" failure
// issue #45 warns about. A byApp that resolves proves init reached the end,
// past both guards.
func TestInitBuiltTheRegistry(t *testing.T) {
	if len(byApp) == 0 {
		t.Fatalf("byApp is empty: init did not build the registry")
	}
	for _, app := range []string{"1.73", "1.74", "1.77", "2.06"} {
		c, err := codecFor(app)
		if err != nil {
			t.Errorf("codecFor(%q): %v", app, err)
			continue
		}
		// The declaration must reach a header, which is what init verified.
		if _, err := xmlHeaderFor(c.AcceptedApps().XMLVersion); err != nil {
			t.Errorf("codecFor(%q) declares an xml version with no header: %v; init must not have loaded", app, err)
		}
	}
}
