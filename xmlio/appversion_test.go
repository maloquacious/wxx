// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
	"github.com/maloquacious/wxx/xmlio/internal/v0_77"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// unacceptedApp is an application version no codec accepts: a hypothetical build
// after the 2.06 baseline. It is deliberately not "1.06" or any other schema
// version -- an encoder accepts application versions and never schema versions
// (issue #41 requirement 1), so feeding it a schema version would test the wrong
// axis.
//
// If a real 2.07 ever ships, this constant moves rather than the assertions: the
// tests below refuse to run against a version some codec has since claimed.
const unacceptedApp = "2.07"

// codecUnderTest pairs a codec's declaration with its encoder and a map that
// codec can really write, so the negative case below fails for the app version
// and not for the map.
type codecUnderTest struct {
	name    string
	set     appver.Set_t
	encode  func(m *wxx.Map_t, app string) ([]byte, error)
	fixture string
}

// codecsUnderTest is every codec, as the disjointness check sees them.
var codecsUnderTest = []codecUnderTest{
	{"v0_77", v0_77.AcceptedApps(), v0_77.Encode, classicFixture},
	{"v1_06", v1_06.AcceptedApps(), v1_06.Encode, sample2025_206},
}

// TestCodecRejectsUnacceptedAppVersion is issue #41 requirement 3 asserted on the
// encoders themselves rather than inferred from the registry: each codec writes
// exactly one schema, so it accepts a closed set of application versions and must
// refuse any other -- an encode that wrote one would emit a file claiming a build
// that never wrote that format.
//
// Each case carries its own positive control, because a gate that rejects
// everything is as broken as one that rejects nothing: every version the codec
// declares must still encode. The accepted counter guards against the vacuous
// pass where a codec declares nothing at all, in which case "it rejected 2.07"
// would prove only that its set was empty.
func TestCodecRejectsUnacceptedAppVersion(t *testing.T) {
	for _, tc := range codecsUnderTest {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.fixture)
			if err != nil {
				t.Fatalf("decode %s: %v", tc.fixture, err)
			}

			// Positive control: every declared version encodes.
			var accepted int
			for _, a := range tc.set.Apps {
				if _, err := tc.encode(m, a.Version); err != nil {
					t.Fatalf("%s.Encode(%s, %q): %v; the codec must accept every version it declares", tc.name, tc.fixture, a.Version, err)
				}
				accepted++
			}
			// Guard against a vacuous pass: a codec that declares no application
			// version rejects unacceptedApp for the wrong reason.
			if accepted == 0 {
				t.Fatalf("codec %s declares no application version, so the gate is not under test", tc.name)
			}
			// Guard against a vacuous pass: the rejection case must be outside
			// the declared set, or the assertion below asks the gate to reject
			// something it is supposed to accept.
			if tc.set.Accepts(unacceptedApp) {
				t.Fatalf("codec %s accepts %q, so it is not an unaccepted version and proves nothing; pick another", tc.name, unacceptedApp)
			}

			// The gate itself.
			data, err := tc.encode(m, unacceptedApp)
			if err == nil {
				t.Fatalf("%s.Encode(%s, %q) returned %d bytes and nil, want an error: this codec does not accept %q", tc.name, tc.fixture, unacceptedApp, len(data), unacceptedApp)
			}
			if !errors.Is(err, wxx.ErrUnacceptedAppVersion) {
				t.Errorf("%s.Encode(%q) error = %v, want it to wrap %v", tc.name, unacceptedApp, err, wxx.ErrUnacceptedAppVersion)
			}
			if data != nil {
				t.Errorf("%s.Encode(%q) returned %d bytes alongside its error, want none", tc.name, unacceptedApp, len(data))
			}
		})
	}
}

// TestCodecAppSetsAreDeclaredAndDisjoint pins what each codec declares, because
// the sets are the codec's own knowledge and nothing else states them: the
// package path names the CODEC version, not the application versions, so v0_77 is
// not the 1.77-only codec and no assertion may infer the set from the path.
//
// The map/@release each application version writes is pinned per app rather than
// per codec. Both codecs currently map every app they accept to one release
// ("2025" for v1_06, absent for v0_77), so an assertion written per codec would
// pass today and would silently stop being the thing under test the moment a
// relabelled build lands on an existing schema (ADR 0004, issue #45 Decision 5).
func TestCodecAppSetsAreDeclaredAndDisjoint(t *testing.T) {
	for _, tc := range []struct {
		name           string
		set            appver.Set_t
		wantSchema     string // the single schema the codec writes; "" means it writes none
		wantXMLVersion string // the XML declaration its files open with
		wantApps       []appver.App_t
	}{
		// Classic 1.73/1.74/1.77 share one element vocabulary, so one codec
		// serves all three, and classic states neither schema nor release.
		{"v0_77", v0_77.AcceptedApps(), "", "1.0", []appver.App_t{
			{Version: "1.73", Release: ""},
			{Version: "1.74", Release: ""},
			{Version: "1.77", Release: ""},
		}},
		{"v1_06", v1_06.AcceptedApps(), "1.06", "1.1", []appver.App_t{
			{Version: "2.06", Release: "2025"},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Guard against a vacuous pass: an empty want makes every loop below
			// a no-op and every assertion in it a tautology.
			if len(tc.wantApps) == 0 {
				t.Fatalf("the case for codec %s names no application version, so it asserts nothing", tc.name)
			}
			if got, want := formatApps(tc.set.Apps), formatApps(tc.wantApps); got != want {
				t.Errorf("codec %s declares [%s], want exactly [%s]", tc.name, got, want)
			}
			if got := tc.set.Schema; got != tc.wantSchema {
				t.Errorf("codec %s writes schema %q, want %q", tc.name, got, tc.wantSchema)
			}
			if got := tc.set.XMLVersion; got != tc.wantXMLVersion {
				t.Errorf("codec %s opens its files with xml version %q, want %q", tc.name, got, tc.wantXMLVersion)
			}
			// A codec version must never be mistaken for a schema version: "0.77"
			// is on no disk, so no codec may name it as the schema it writes.
			if tc.set.Schema == "0.77" {
				t.Errorf("codec %s declares schema %q, which is a codec version and appears in no file", tc.name, tc.set.Schema)
			}
			// The declaration must be valid on its own terms, which is what lets
			// the codec derive an identity from it rather than echo the source's.
			if err := tc.set.Verify(); err != nil {
				t.Errorf("codec %s declaration is invalid: %v", tc.name, err)
			}
			for _, want := range tc.wantApps {
				if !tc.set.Accepts(want.Version) {
					t.Errorf("codec %s does not accept %q, which it declares", tc.name, want.Version)
					continue
				}
				got, ok := tc.set.App(want.Version)
				if !ok {
					t.Errorf("codec %s: App(%q) not found though Accepts(%q) is true", tc.name, want.Version, want.Version)
					continue
				}
				if got.Release != want.Release {
					t.Errorf("codec %s: version %q writes release %q, want %q", tc.name, want.Version, got.Release, want.Release)
				}
			}
		})
	}

	// The property itself, over the real sets.
	//
	// "Every supported release is accepted by exactly one codec" used to be
	// asserted here too, by walking xmlio.SupportedReleases() and counting the
	// codecs that claimed each entry. It is gone rather than restated, and issue
	// #45 Decision 8 is why: the registry no longer has a table of releases to
	// disagree with the codecs -- it IS the union of what they declare -- so the
	// check would be asking whether the codecs agree with themselves. What
	// survives of it is disjointness, which is the half that was ever a property,
	// and TestRegistryIsExactlyTheSupportedApplicationVersions in codecs_test.go,
	// which pins what that union contains.
	if err := appver.VerifyDisjoint(codecAppSetsForTest()...); err != nil {
		t.Errorf("the compiled-in codec sets are not disjoint: %v", err)
	}
}

// TestVerifyDisjointRejectsOverlap deliberately hands the check an overlapping
// table, because a check that cannot fail is worthless and the real table cannot
// be made to overlap without editing the codecs.
//
// The check has to live outside init to be testable at all -- init panics, and a
// panic cannot be inspected -- which is why appver.VerifyDisjoint takes a table.
// xmlio.verifyXMLVersions is the other half of the same load-time contract and is
// shaped the same way, for the same reason.
//
// This is also the MERGED guard. Issue #41 kept the registry's
// ErrDuplicateAppVersion check (one version must not name two RELEASES) apart
// from this one (one version must not be accepted by two CODECS); with the
// registry collapsed to application version -> codec they are the same statement,
// and this is the survivor.
func TestVerifyDisjointRejectsOverlap(t *testing.T) {
	// Control: without the overlap this table is accepted. Guard against a
	// vacuous pass -- if VerifyDisjoint rejected everything, the cases below
	// would pass while proving nothing about overlap.
	ok := []appver.Set_t{
		{Codec: "v0_77", Schema: "", XMLVersion: "1.0", Apps: []appver.App_t{{Version: "1.73"}, {Version: "1.74"}, {Version: "1.77"}}},
		{Codec: "v1_06", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}}},
	}
	if err := appver.VerifyDisjoint(ok...); err != nil {
		t.Fatalf("VerifyDisjoint(disjoint table): %v; the overlap cases below would prove nothing", err)
	}

	for _, tc := range []struct {
		name    string
		sets    []appver.Set_t
		wantErr error
		wantMsg string
	}{
		{
			// The failure requirement 4 exists to stop: two codecs both claiming
			// to write one application version, so "which codec writes 1.77" has
			// two answers and the file you get depends on lookup order.
			name: "two codecs claim one application version",
			sets: []appver.Set_t{
				{Codec: "v0_77", Schema: "", XMLVersion: "1.0", Apps: []appver.App_t{{Version: "1.73"}, {Version: "1.74"}, {Version: "1.77"}}},
				{Codec: "v1_06", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "1.77", Release: "2025"}, {Version: "2.06", Release: "2025"}}},
			},
			wantErr: wxx.ErrAmbiguousAppCodec,
			wantMsg: `"1.77"`,
		},
		{
			name: "one codec names a version twice",
			sets: []appver.Set_t{
				{Codec: "v1_06", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}, {Version: "2.06", Release: "2025"}}},
			},
			wantErr: wxx.ErrAmbiguousAppCodec,
			wantMsg: `"2.06"`,
		},
		{
			// An empty version is not a version: it would match a map that states
			// no @version at all and let it through the gate. VerifyDisjoint
			// reaches this through Set_t.Verify -- see
			// TestVerifyRejectsSchemaAndReleaseDisagreeing for that check's own
			// coverage.
			name: "a codec names an empty application version",
			sets: []appver.Set_t{
				{Codec: "v1_06", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "", Release: "2025"}}},
			},
			wantErr: wxx.ErrMissingVersion,
			wantMsg: "v1_06",
		},
		{
			// A declaration that is invalid on its own terms is not a table worth
			// asking a question across: VerifyDisjoint runs Set_t.Verify first.
			name: "a codec states no schema but its app states a release",
			sets: []appver.Set_t{
				{Codec: "v0_77", Schema: "", XMLVersion: "1.0", Apps: []appver.App_t{{Version: "1.77", Release: "2025"}}},
			},
			wantErr: wxx.ErrInvalidCodecDeclaration,
			wantMsg: "if and only if",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := appver.VerifyDisjoint(tc.sets...)
			if err == nil {
				t.Fatalf("VerifyDisjoint(%s) = nil, want an error", tc.name)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("VerifyDisjoint(%s) error = %v, want it to wrap %v", tc.name, err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("VerifyDisjoint(%s) error = %v, want it to name %s", tc.name, err, tc.wantMsg)
			}
		})
	}
}

// TestAcceptedAppsIsACopy asserts a codec's declaration cannot be edited through
// the accessor. The set is the codec's own knowledge; handing out the ability to
// rewrite it would make every check above conditional on nobody having done so.
func TestAcceptedAppsIsACopy(t *testing.T) {
	got := v1_06.AcceptedApps()
	if len(got.Apps) == 0 {
		t.Fatalf("v1_06 declares no application version, so there is nothing to mutate and this proves nothing")
	}
	was := got.Apps[0]
	got.Apps[0] = appver.App_t{Version: unacceptedApp, Release: "9999"}

	again := v1_06.AcceptedApps()
	if again.Accepts(unacceptedApp) {
		t.Errorf("mutating the returned set changed what v1_06 accepts: it now accepts %q", unacceptedApp)
	}
	// The release a version writes is as reachable through the slice as the
	// version itself, and it is what the encoder puts on disk (issue #45), so the
	// copy has to cover it too.
	if a, ok := again.App(was.Version); !ok {
		t.Errorf("mutating the returned set removed %q from what v1_06 accepts", was.Version)
	} else if a.Release != was.Release {
		t.Errorf("mutating the returned set changed the release %q writes: got %q, want %q", was.Version, a.Release, was.Release)
	}
}

// TestCodecDeclarationsAreValid asserts every compiled-in codec declares a set
// that is valid on its own terms.
//
// The property is "a codec states a schema if and only if its apps state a
// release": classic files carry neither map/@schema nor map/@release, W2025 files
// carry both (ADR 0003 Decision 2). NewRegistry used to enforce this over registry
// entries, and issue #41 called it load-bearing; with the registry collapsed to
// application version -> codec (issue #45 Decision 8) the declaration is the only
// place left that knows both, so the check lives there and this asserts it over
// the real codecs.
//
// codecs.go's init runs the same check and panics, which is why this cannot be
// the only coverage: a panic cannot be inspected, and
// TestVerifyRejectsSchemaAndReleaseDisagreeing is where the guard is watched to
// fail.
func TestCodecDeclarationsAreValid(t *testing.T) {
	sets := codecAppSetsForTest()
	// Guard against a vacuous pass: a loop over an empty table proves nothing.
	if len(sets) == 0 {
		t.Fatalf("no codec declarations under test")
	}
	for _, s := range sets {
		if err := s.Verify(); err != nil {
			t.Errorf("codec %s: %v", s.Codec, err)
		}
	}
}

// TestVerifyRejectsSchemaAndReleaseDisagreeing deliberately hands Verify
// declarations that pair a schema with the wrong release, because a guard that
// cannot fail is worthless and the real declarations cannot be made to disagree
// without editing the codecs.
//
// Both directions are covered. A codec that states no schema but whose app writes
// release="2025" would emit classic content stating a W2025 release; one that
// states schema "1.06" but whose app writes no release would emit W2025 content
// with no release attribute. Either is a file that no build ever wrote, and
// either is the codec's own declaration being wrong rather than the caller's
// input -- so it must fail at load, not at encode.
func TestVerifyRejectsSchemaAndReleaseDisagreeing(t *testing.T) {
	// Control: without the disagreement these shapes verify. Guard against a
	// vacuous pass -- if Verify rejected everything, the cases below would pass
	// while proving nothing about the pairing.
	for _, s := range []appver.Set_t{
		{Codec: "classic-like", Schema: "", XMLVersion: "1.0", Apps: []appver.App_t{{Version: "1.77", Release: ""}}},
		{Codec: "w2025-like", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}}},
		// A relabelled build on an existing schema is valid: the guard ties the
		// PRESENCE of a release to the presence of a schema, never its value
		// (ADR 0004's relabel scenario).
		{Codec: "relabelled", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}, {Version: "2.14", Release: "2026"}}},
	} {
		if err := s.Verify(); err != nil {
			t.Fatalf("Verify(%s): %v; the cases below would prove nothing", s.Codec, err)
		}
	}

	for _, tc := range []struct {
		name    string
		set     appver.Set_t
		wantMsg string
	}{
		{
			name:    "no schema but a release",
			set:     appver.Set_t{Codec: "v0_77", Schema: "", XMLVersion: "1.0", Apps: []appver.App_t{{Version: "1.77", Release: "2025"}}},
			wantMsg: `"1.77"`,
		},
		{
			name:    "a schema but no release",
			set:     appver.Set_t{Codec: "v1_06", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "2.06", Release: ""}}},
			wantMsg: `"2.06"`,
		},
		{
			// The check is per app, so a set whose apps disagree with each other
			// is caught on the one that disagrees with the schema.
			name:    "apps disagree with each other",
			set:     appver.Set_t{Codec: "v1_06", Schema: "1.06", XMLVersion: "1.1", Apps: []appver.App_t{{Version: "2.06", Release: "2025"}, {Version: "2.14", Release: ""}}},
			wantMsg: `"2.14"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.set.Verify()
			if err == nil {
				t.Fatalf("Verify(%s) = nil, want an error", tc.name)
			}
			if !errors.Is(err, wxx.ErrInvalidCodecDeclaration) {
				t.Errorf("Verify(%s) error = %v, want it to wrap %v", tc.name, err, wxx.ErrInvalidCodecDeclaration)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("Verify(%s) error = %v, want it to name %s", tc.name, err, tc.wantMsg)
			}
		})
	}
}

// formatApps renders a declaration's apps for an assertion message, version and
// release both, so a diff shows which of the two moved.
func formatApps(apps []appver.App_t) string {
	out := make([]string, 0, len(apps))
	for _, a := range apps {
		out = append(out, fmt.Sprintf("%s->%q", a.Version, a.Release))
	}
	return strings.Join(out, ",")
}

// codecAppSetsForTest mirrors the dispatcher's table (xmlio/codecs.go). The
// entries are read from the codecs rather than restated, so the only thing it can
// drift on is which codecs exist -- which is what it is here to state.
func codecAppSetsForTest() []appver.Set_t {
	return []appver.Set_t{
		v0_77.AcceptedApps(),
		v1_06.AcceptedApps(),
	}
}
