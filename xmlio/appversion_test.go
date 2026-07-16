// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
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
			for _, app := range tc.set.Apps {
				if _, err := tc.encode(m, app); err != nil {
					t.Fatalf("%s.Encode(%s, %q): %v; the codec must accept every version it declares", tc.name, tc.fixture, app, err)
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
func TestCodecAppSetsAreDeclaredAndDisjoint(t *testing.T) {
	for _, tc := range []struct {
		name       string
		set        appver.Set_t
		wantSchema string // the single schema the codec writes; "" means it writes none
		wantApps   []string
	}{
		// Classic 1.73/1.74/1.77 share one element vocabulary, so one codec
		// serves all three, and classic states no schema at all.
		{"v0_77", v0_77.AcceptedApps(), "", []string{"1.73", "1.74", "1.77"}},
		{"v1_06", v1_06.AcceptedApps(), "1.06", []string{"2.06"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := strings.Join(tc.set.Apps, ","), strings.Join(tc.wantApps, ","); got != want {
				t.Errorf("codec %s accepts [%s], want exactly [%s]", tc.name, got, want)
			}
			if got := tc.set.Schema; got != tc.wantSchema {
				t.Errorf("codec %s writes schema %q, want %q", tc.name, got, tc.wantSchema)
			}
			// A codec version must never be mistaken for a schema version: "0.77"
			// is on no disk, so no codec may name it as the schema it writes.
			if tc.set.Schema == "0.77" {
				t.Errorf("codec %s declares schema %q, which is a codec version and appears in no file", tc.name, tc.set.Schema)
			}
			for _, app := range tc.wantApps {
				if !tc.set.Accepts(app) {
					t.Errorf("codec %s does not accept %q, which it declares", tc.name, app)
				}
			}
		})
	}

	// The property itself, over the real sets.
	if err := appver.VerifyDisjoint(codecAppSetsForTest()...); err != nil {
		t.Errorf("the compiled-in codec sets are not disjoint: %v", err)
	}

	// Every supported release's application version must be accepted by the codec
	// bound to it -- exactly one codec, since the sets are disjoint. Without this
	// the declarations could be fiction that no registry entry ever reaches.
	for _, e := range xmlio.SupportedReleases() {
		var claims []string
		for _, s := range codecAppSetsForTest() {
			if s.Accepts(e.App.Raw) {
				claims = append(claims, s.Codec)
			}
		}
		if len(claims) != 1 {
			t.Errorf("supported release %q is accepted by %d codecs (%s), want exactly 1", e.App.Raw, len(claims), strings.Join(claims, ", "))
		}
	}
}

// TestVerifyDisjointRejectsOverlap deliberately hands the check an overlapping
// table, because a check that cannot fail is worthless and the real table cannot
// be made to overlap without editing the codecs.
//
// The check has to live outside init to be testable at all -- init panics, and a
// panic cannot be inspected -- which is why appver.VerifyDisjoint takes a table.
// This mirrors NewRegistry and the reason is the same.
func TestVerifyDisjointRejectsOverlap(t *testing.T) {
	// Control: without the overlap this table is accepted. Guard against a
	// vacuous pass -- if VerifyDisjoint rejected everything, the cases below
	// would pass while proving nothing about overlap.
	ok := []appver.Set_t{
		{Codec: "v0_77", Schema: "", Apps: []string{"1.73", "1.74", "1.77"}},
		{Codec: "v1_06", Schema: "1.06", Apps: []string{"2.06"}},
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
				{Codec: "v0_77", Schema: "", Apps: []string{"1.73", "1.74", "1.77"}},
				{Codec: "v1_06", Schema: "1.06", Apps: []string{"1.77", "2.06"}},
			},
			wantErr: wxx.ErrAmbiguousAppCodec,
			wantMsg: `"1.77"`,
		},
		{
			name: "one codec names a version twice",
			sets: []appver.Set_t{
				{Codec: "v1_06", Schema: "1.06", Apps: []string{"2.06", "2.06"}},
			},
			wantErr: wxx.ErrAmbiguousAppCodec,
			wantMsg: `"2.06"`,
		},
		{
			// An empty version is not a version: it would match a map that states
			// no @version at all and let it through the gate.
			name: "a codec names an empty application version",
			sets: []appver.Set_t{
				{Codec: "v1_06", Schema: "1.06", Apps: []string{""}},
			},
			wantErr: wxx.ErrMissingVersion,
			wantMsg: "v1_06",
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
	got.Apps[0] = unacceptedApp

	if again := v1_06.AcceptedApps(); again.Accepts(unacceptedApp) {
		t.Errorf("mutating the returned set changed what v1_06 accepts: it now accepts %q", unacceptedApp)
	}
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
