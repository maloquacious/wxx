// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// spyWriter is the io.Writer an encode that must write NOTHING is handed. It
// records every call rather than the bytes, because the assertion those tests
// make is not "the output is empty" but "the writer was never touched".
//
// A bytes.Buffer cannot tell the difference: an encoder that wrote a partial
// file and then returned an error would leave a non-empty buffer, but so would
// one that wrote a complete file, and neither is distinguishable from a caller
// who is about to os.WriteFile whatever arrived. A rejected target must not
// produce a file at all -- not a short one, not an empty one.
type spyWriter struct {
	calls int
	n     int
}

func (w *spyWriter) Write(p []byte) (int, error) {
	w.calls++
	w.n += len(p)
	return len(p), nil
}

// xmlHeaderSamples pairs a real fixture with the XML declaration the release it
// states opens its files with, and with the schema that routes it there.
//
// wantSchema is the exact map/@schema bytes; "" means the fixture states none,
// which is the implicit legacy (classic) schema.
var xmlHeaderSamples = []struct {
	name       string
	path       string
	app        string // the target, which each fixture states as its own
	wantSchema string
	wantHeader string
}{
	{"classic 1.77", classicFixture, "1.77", "", "<?xml version='1.0' encoding='utf-16'?>\n"},
	{"w2025 2.06", sample2025_206, "2.06", "1.06", "<?xml version='1.1' encoding='utf-16'?>\n"},
}

// TestEncodeXMLHeaderFollowsRelease asserts, at the byte level and from real
// fixtures, that the XML declaration an encode emits follows the target RELEASE:
// classic opens `<?xml version='1.0'`, W2025 opens `<?xml version='1.1'`.
//
// This is the property the deleted `switch target.Major { case 2017: ...; case
// 2025: ... }` used to guarantee. The declaration is now the CODEC's, declared
// alongside the schema and release it writes (appver.Set_t.XMLVersion), so nothing
// in the encoder knows a family year and the registry does not carry the answer
// either -- but the bytes of every file written must not have moved an inch, and
// that is what this pins.
//
// The bytes are the real ones: each fixture is encoded through the full default
// pipeline (XML, header, UTF-16/BE, gzip) and the output is transported back
// through the decoder, whose Converted diagnostic is the file's text as it would
// be read from disk -- header included, since it is captured before the header is
// consumed.
func TestEncodeXMLHeaderFollowsRelease(t *testing.T) {
	// Guard against a vacuous pass: this test discriminates only if the cases
	// expect DIFFERENT declarations. Were they ever collapsed to one expectation,
	// every assertion below would still pass against an encoder that hard-coded a
	// single header -- which is precisely the regression worth catching.
	distinct := map[string]bool{}
	for _, tc := range xmlHeaderSamples {
		distinct[tc.wantHeader] = true
	}
	if len(distinct) < 2 {
		t.Fatalf("all %d samples expect the same XML declaration, so 'the header follows the release' is not under test", len(xmlHeaderSamples))
	}

	for _, tc := range xmlHeaderSamples {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}

			// Sanity: the fixture really does state the schema that routes it, so
			// the two cases exercise two different releases rather than one.
			schema := m.MetaData.Version.Schema
			if tc.wantSchema == "" {
				if schema != nil {
					t.Fatalf("%s: Version.Schema = %+v, want nil; this case must exercise the implicit legacy schema", tc.path, *schema)
				}
			} else if schema == nil {
				t.Fatalf("%s: Version.Schema = nil, want %q; this case must exercise the W2025 schema", tc.path, tc.wantSchema)
			} else if schema.Raw != tc.wantSchema {
				t.Fatalf("%s: Version.Schema.Raw = %q, want %q", tc.path, schema.Raw, tc.wantSchema)
			}

			// The target is named explicitly -- there is no default (issue #45) --
			// and each case names the version its own fixture states, which is what
			// keeps this test about the HEADER rather than about re-targeting.
			if got := m.MetaData.Version.App.Raw; got != tc.app {
				t.Fatalf("%s states version %q, want %q: this case must target the version its fixture states", tc.path, got, tc.app)
			}
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(tc.app).Encode(&buf, m); err != nil {
				t.Fatalf("public encode %s (%v): %v", tc.path, m.MetaData.Version, err)
			}

			// Read the encoded file back to its text, header and all.
			var d xmlio.DecoderDiagnostics
			if _, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&d)).Decode(&buf); err != nil {
				t.Fatalf("re-decode encoded %s: %v", tc.path, err)
			}
			if len(d.Converted) == 0 {
				t.Fatalf("re-decode encoded %s: diagnostics.Converted is empty", tc.path)
			}
			if !bytes.HasPrefix(d.Converted, []byte(tc.wantHeader)) {
				t.Errorf("encoded %s opens %q, want it to open %q: the XML declaration must follow the release",
					tc.path, head(d.Converted, len(tc.wantHeader)), tc.wantHeader)
			}
			// The other release's declaration must not appear in its place.
			for _, other := range xmlHeaderSamples {
				if other.wantHeader == tc.wantHeader {
					continue
				}
				if bytes.HasPrefix(d.Converted, []byte(other.wantHeader)) {
					t.Errorf("encoded %s opens with %q, the declaration of a different release", tc.path, other.wantHeader)
				}
			}
		})
	}
}

// TestEncodeUnsupportedTargetIsError asserts that a target no codec writes is an
// error and writes nothing (ADR 0004 Decision 5). A best-effort write here would
// hand a user a file claiming to be a release that does not exist, or one they
// are not licensed for.
//
// The target now reaches the encoder ONE way: the caller names it (issue #45
// Decision 1). It used to reach it two, and the second one -- the map states it,
// the caller names nothing -- is what this ticket deleted, so the cases that
// covered it are gone rather than restated:
//
//   - "map states a future version" and "map states an unpadded 2.06" fed the
//     encoder a map with a doctored MetaData.Version.App and watched the fallback
//     refuse it. There is no fallback to refuse anything: the encoder never reads
//     that field, which TestEncodeIgnoresTheMapsOwnVersion pins directly.
//   - "map states nothing" was the fallback's version of the empty target, and
//     TestEncodeEmptyTargetVersionIsError is the surviving statement of it.
func TestEncodeUnsupportedTargetIsError(t *testing.T) {
	// Control: the fixture encodes cleanly as itself. Guard against a vacuous
	// pass -- if this map could not be encoded at all, every error below would be
	// the map's fault and would say nothing about target resolution.
	base, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}
	var control bytes.Buffer
	if err := xmlio.NewEncoder(base.MetaData.Version.App.Raw).Encode(&control, base); err != nil {
		t.Fatalf("public encode %s as itself: %v; the unsupported-target cases below would prove nothing", classicFixture, err)
	}
	if control.Len() == 0 {
		t.Fatalf("public encode %s as itself: empty output", classicFixture)
	}

	for _, tc := range []struct {
		name   string
		target string
	}{
		{"a future version", "9.99"},
		{"an unpadded 2.06", "2.6"},
		{"a schema, not an app version", "1.06"},
		{"a codec version, not an app version", "0.77"},
		{"an unreleased classic", "1.75"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, classicFixture)
			if err != nil {
				t.Fatalf("public decode %s: %v", classicFixture, err)
			}

			var buf bytes.Buffer
			err = xmlio.NewEncoder(tc.target).Encode(&buf, m)
			if err == nil {
				t.Fatalf("encode targeting %q wrote %d bytes and returned nil, want an error: an unregistered target must never be a best-effort write", tc.target, buf.Len())
			}
			if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
				t.Errorf("encode targeting %q error = %v, want it to wrap %v", tc.target, err, wxx.ErrUnsupportedMapVersion)
			}
			// Nothing may reach the writer: a rejected target is not a partial file.
			if buf.Len() != 0 {
				t.Errorf("encode targeting %q wrote %d bytes before failing, want 0", tc.target, buf.Len())
			}
		})
	}
}

// TestEncodeIgnoresTheMapsOwnVersion is issue #45's bug, stated at the level the
// public API can see it: the encoder writes the version it was ASKED for, and the
// version the map states has no effect on the file at all.
//
// It is the successor to the deleted "map states a future version" cases, and it
// asserts the stronger thing they gestured at. Those cases proved the fallback
// refused a bad value; this proves there is no fallback -- a map whose stated
// identity is garbage encodes to a perfectly good file, byte-identical to the one
// the untouched map produces, because nothing ever read it.
//
// The doctored identity is deliberately something no codec accepts. If the
// encoder still read the map at all, this would fail loudly rather than quietly
// producing a slightly different file.
func TestEncodeIgnoresTheMapsOwnVersion(t *testing.T) {
	const target = "1.73"

	clean, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}
	// Guard against a vacuous pass: the fixture must state something OTHER than
	// the target, or "the map's version is ignored" is untested -- the two would
	// agree and either could be the source of the bytes.
	if got := clean.MetaData.Version.App.Raw; got == target {
		t.Fatalf("%s states version %q, the same as the target: the map's version must differ or it cannot be shown to be ignored", classicFixture, got)
	}
	var want bytes.Buffer
	if err := xmlio.NewEncoder(target).Encode(&want, clean); err != nil {
		t.Fatalf("public encode %s targeting %q: %v", classicFixture, target, err)
	}
	if want.Len() == 0 {
		t.Fatalf("public encode %s targeting %q: empty output", classicFixture, target)
	}

	for _, tc := range []struct {
		name string
		app  wxx.Dotted
	}{
		{"a version no codec writes", wxx.Dotted{Raw: "9.99", Major: 9, Minor: 99}},
		{"a schema version", wxx.Dotted{Raw: "1.06", Major: 1, Minor: 6}},
		{"nothing at all", wxx.Dotted{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, classicFixture)
			if err != nil {
				t.Fatalf("public decode %s: %v", classicFixture, err)
			}
			// Doctor the identity the map states. Every one of these would have
			// stopped the encode dead when the encoder defaulted to it.
			m.MetaData.Version.App = tc.app
			m.MetaData.Worldographer.Version = tc.app.Raw
			m.Version = tc.app.Raw

			var got bytes.Buffer
			if err := xmlio.NewEncoder(target).Encode(&got, m); err != nil {
				t.Fatalf("public encode targeting %q with the map stating %q: %v; the map's own version must not reach the encoder", target, tc.app.Raw, err)
			}
			if !bytes.Equal(got.Bytes(), want.Bytes()) {
				t.Errorf("a map stating %q encodes to %d bytes targeting %q, want the %d bytes the untouched map produces: the encoder must not read the map's identity (issue #45)",
					tc.app.Raw, got.Len(), target, want.Len())
			}
		})
	}
}

// unlicensedTarget is a hypothetical FUTURE Worldographer application version:
// one that does not exist, that the registry therefore does not state, and that
// no user can hold a license for. It is the exact example ADR 0004 Decision 5
// uses ("a user licensed for 2.06 cannot be handed a 2.07 file").
const unlicensedTarget = "2.07"

// TestEncodeUnlicensedTargetWritesNothing is the licensing test. Targeting a
// release the registry does not state must fail, and must fail before anything
// reaches the writer.
//
// Returning an error is not sufficient on its own. A caller that encodes to a
// file, a buffer or a network stream has already received whatever the encoder
// wrote by the time the error arrives; a best-effort write that also errors
// still hands the user a "2.07" file, which is the thing the licensing
// requirement forbids. So the assertion is on the WRITER: it must not be called.
//
// The control makes the pair meaningful: the same map, encoded by the same
// pipeline, succeeds for the licensed target and is refused for the unlicensed
// one, so the refusal is about the target and not about the map.
func TestEncodeUnlicensedTargetWritesNothing(t *testing.T) {
	// Guard against a vacuous pass: this test says nothing unless the target it
	// names is genuinely unregistered. Were 2.07 ever added to the registry, this
	// stops the test rather than letting it "pass" against a licensed target.
	if _, ok := codecForAppOfTest(t, unlicensedTarget); ok {
		t.Fatalf("%q is accepted by a codec: this test requires an UNREGISTERED version, so it is not testing the licensing refusal", unlicensedTarget)
	}

	m, err := decodeFile(t, sample2025_206)
	if err != nil {
		t.Fatalf("public decode %s: %v", sample2025_206, err)
	}
	licensed := m.MetaData.Version.App.Raw

	// Control: the licensed target encodes. If it did not, the refusal below
	// would be the map's fault and would prove nothing about the target.
	var ok spyWriter
	if err := xmlio.NewEncoder(licensed).Encode(&ok, m); err != nil {
		t.Fatalf("public encode %s targeting its licensed %q: %v; the refusal below would prove nothing", sample2025_206, licensed, err)
	}
	if ok.n == 0 {
		t.Fatalf("public encode %s targeting its licensed %q wrote 0 bytes", sample2025_206, licensed)
	}

	// The unlicensed target: refused, and the writer never touched.
	var spy spyWriter
	err = xmlio.NewEncoder(unlicensedTarget).Encode(&spy, m)
	if err == nil {
		t.Fatalf("encode targeting unlicensed %q returned nil after %d writes (%d bytes), want an error: an unregistered target must never be a best-effort write", unlicensedTarget, spy.calls, spy.n)
	}
	if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
		t.Errorf("encode error = %v, want it to wrap %v", err, wxx.ErrUnsupportedMapVersion)
	}
	if spy.calls != 0 || spy.n != 0 {
		t.Errorf("encode targeting unlicensed %q made %d writes totaling %d bytes, want 0 and 0: a user licensed for %q must not be handed a %q file, and an error they receive after the bytes have already been written does not prevent that",
			unlicensedTarget, spy.calls, spy.n, licensed, unlicensedTarget)
	}
}

// TestEncodeEmptyTargetVersionIsError pins the contract for NewEncoder(""): it
// names no application version, so it is an error, exactly as any other
// unregistered version is.
//
// "" used to be a sentinel meaning "target the map's own version". That made an
// unset flag or an empty config field silently produce a file in a release the
// caller never named, and produce it indistinguishably from the encoder honoring
// the request. An explicitly passed empty version is overwhelmingly a caller bug,
// and the correct answer to a caller bug is not to write a file anyway and hope
// the version it lands in was the one they meant.
//
// Issue #45 removed the sentinel's meaning entirely rather than only its use:
// there is no default target for "" to fall back TO. This test therefore reads
// differently than it did -- the control below can no longer show what "" used to
// buy, because no call can ask for it -- but the contract it pins is the same one,
// and "" being an ordinary miss is now a property of the registry rather than a
// special case in the option.
func TestEncodeEmptyTargetVersionIsError(t *testing.T) {
	m, err := decodeFile(t, classicFixture)
	if err != nil {
		t.Fatalf("public decode %s: %v", classicFixture, err)
	}

	// Guard against a vacuous pass: the map's own version must be REGISTERED, so
	// that the old sentinel behavior would have succeeded here and the error below
	// is the new contract firing rather than an unrelated failure that would
	// happen with or without it.
	own := m.MetaData.Version.App.Raw
	if _, ok := codecForAppOfTest(t, own); !ok {
		t.Fatalf("fixture states version %q, which no codec accepts: an empty target would then fail for the wrong reason and this test would assert nothing", own)
	}

	// Control: naming the map's own release really does write a file. That is the
	// behavior "" used to buy silently and must now be asked for out loud.
	var control bytes.Buffer
	if err := xmlio.NewEncoder(own).Encode(&control, m); err != nil {
		t.Fatalf("public encode %s targeting its own %q: %v", classicFixture, own, err)
	}
	if control.Len() == 0 {
		t.Fatalf("public encode %s targeting its own %q: empty output", classicFixture, own)
	}

	var spy spyWriter
	err = xmlio.NewEncoder("").Encode(&spy, m)
	if err == nil {
		t.Fatalf(`NewEncoder("") returned nil after %d writes (%d bytes), want an error: "" names no application version and must not silently fall back to the map's own %q`, spy.calls, spy.n, own)
	}
	if !errors.Is(err, wxx.ErrUnsupportedMapVersion) {
		t.Errorf(`NewEncoder("") error = %v, want it to wrap %v`, err, wxx.ErrUnsupportedMapVersion)
	}
	if spy.calls != 0 || spy.n != 0 {
		t.Errorf(`NewEncoder("") made %d writes totaling %d bytes, want 0 and 0: a target the encoder rejected is not a partial file`, spy.calls, spy.n)
	}
}

// retargetCases pair a source fixture with a registered application version to
// target it at, and are the positive half of the target contract: every
// registered release resolves, encodes, and writes ITS OWN version string.
//
// Cross-family re-targeting (classic <-> W2025) is deliberately absent. That is
// a question about what a target can express -- a downgrade -- and it is tracked
// separately; target RESOLUTION is what is under test here. Each case therefore
// stays within its source's schema, which is where a re-target is legitimate:
// classic 1.73, 1.74 and 1.77 share one element vocabulary and therefore one
// codec, and differ only in the string written to map/@version. That is ADR 0004
// Decision 4's "the application version is data" claim, stated as bytes.
var retargetCases = []struct {
	name   string
	path   string
	target string
}{
	{"classic 1.77 as itself", classicFixture, "1.77"},
	{"classic 1.77 -> 1.74", classicFixture, "1.74"},
	{"classic 1.77 -> 1.73", classicFixture, "1.73"},
	{"w2025 2.06 as itself", sample2025_206, "2.06"},
}

// TestEncodeTargetsEveryRegisteredRelease asserts that each registered release
// can be targeted, and that targeting it puts THAT release's version on disk.
//
// The second half is the one with teeth. The target selects the codec by schema,
// but the codec writes the version string the Map_t carries -- the SOURCE file's
// -- so a target that did not also set the identity would route a 1.77 map
// through the classic codec, write version="1.77", and report success for a
// request to write 1.73. The caller would have been told they got the release
// they asked for while holding a file that says otherwise, which is the
// licensing guarantee failing open rather than closed.
func TestEncodeTargetsEveryRegisteredRelease(t *testing.T) {
	// Guard against a registry entry drifting out of test: "every registered
	// version resolves and encodes" is a claim about the registry, not about
	// whichever subset of it this table happens to list.
	targeted := map[string]bool{}
	for _, tc := range retargetCases {
		targeted[tc.target] = true
	}
	for _, c := range codecsForTest() {
		for _, a := range c.AcceptedApps().Apps {
			if !targeted[a.Version] {
				t.Errorf("version %q is registered but no case targets it: add one, or this test does not cover the registry", a.Version)
			}
		}
	}

	// Guard against a vacuous pass: at least one case must target a release the
	// source does NOT already state. If every case encoded a map as itself, the
	// assertions below would all pass against an encoder that ignored the target
	// entirely -- which is precisely the regression worth catching.
	retargets := 0

	for _, tc := range retargetCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := decodeFile(t, tc.path)
			if err != nil {
				t.Fatalf("public decode %s: %v", tc.path, err)
			}
			if m.MetaData.Version.App.Raw != tc.target {
				retargets++
			}

			var buf bytes.Buffer
			if err := xmlio.NewEncoder(tc.target).Encode(&buf, m); err != nil {
				t.Fatalf("public encode %s targeting %q: %v", tc.path, tc.target, err)
			}
			if buf.Len() == 0 {
				t.Fatalf("public encode %s targeting %q: empty output", tc.path, tc.target)
			}

			// Read the file back: what it SAYS it is must be what was asked for.
			m2, err := xmlio.NewDecoder().Decode(&buf)
			if err != nil {
				t.Fatalf("re-decode %s targeted at %q: %v", tc.path, tc.target, err)
			}
			if got := m2.MetaData.Worldographer.Version; got != tc.target {
				t.Errorf("%s targeted at %q wrote map/@version=%q, want %q verbatim: the file must state the release the caller targeted, not the one the source stated",
					tc.path, tc.target, got, tc.target)
			}
			if got := m2.MetaData.Version.App.Raw; got != tc.target {
				t.Errorf("%s targeted at %q re-decodes with Version.App.Raw = %q, want %q", tc.path, tc.target, got, tc.target)
			}
		})
	}

	if retargets == 0 {
		t.Errorf("every case targeted the release its source already states, so no case exercises re-targeting")
	}
}

// TestEncodeTargetRelease is gone, and its coverage moved rather than lapsed.
//
// It covered WithTargetRelease, which no longer exists: once Release_t stopped
// carrying a codec, WithTargetRelease(r) was exactly WithTargetVersion(r.App.Raw)
// and nothing more (issue #41). Each of its cases still has a home:
//
//   - "a registry entry targets its release" was Lookup("1.73") then targeting
//     it. TestEncodeTargetsEveryRegisteredRelease already targets 1.73 from the
//     1.77 fixture by version, and asserts the same thing about the same bytes.
//   - "a nil release is an error" was the option's form of "the caller named
//     nothing". NewEncoder("") is the surviving way to name nothing, and
//     TestEncodeEmptyTargetVersionIsError holds it to the same contract:
//     an error, and not one byte written.
//   - "an assembled release is rejected", "MarshalXML rejects an assembled
//     release" and "a copy of a registry entry is rejected" were the three faces
//     of one guarantee -- an unregistered (App, Schema) pair must never reach a
//     codec -- which Resolve enforced at run time by pointer identity. No public
//     entry point takes a *Release_t any more, so an assembled or copied one has
//     nowhere to go: the guarantee is now held by the type system rather than by
//     a check, and TestChimeraIsUnreachableThroughThePublicAPI in
//     chimera_test.go states it, including the compile-level half.
