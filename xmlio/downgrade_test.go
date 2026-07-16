// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// This file holds the downgrade-loss tests (#32, ADR 0004 Decision 7). The
// inventory under test was not written from memory: it was DERIVED by running
// the classic round-trip audit harness in roundtrip_2017_test.go (xmlAggregate /
// computeLoss) across three encodes and subtracting the controls --
//
//	experiment: W2025 2.06 -> classic target   (downgrade + identity + codec gaps)
//	control A:  W2025 2.06 -> W2025 2.06       (h2025 codec gaps alone)
//	control B:  classic    -> classic          (classic codec gaps alone;
//	                                            classicRoundTripExpect)
//
// -- and taking the residual. TestClassicDowngradeLossInventory below re-runs the
// experiment and holds the encoder to that residual, so the evidence is
// executable rather than a claim in a comment.
//
// The controls are what keep the inventory honest. Control A is the reason
// map/features/feature/label/@dropShadow* is NOT reported as a downgrade loss:
// those attributes are dropped on a 2.06 -> 2.06 round trip too (Map_t.Label_t
// models no drop shadow -- the trio lives on LabelStyle_t), so they are an h2025
// codec gap that targeting classic merely also exhibits. Control B is the reason
// mapkey/@viewlevel, <informations> and <labelstyle> are not reported: classic
// loses those to itself.

// classicTarget is the classic release every downgrade test targets. Any of
// 1.73/1.74/1.77 would do -- they share the one implicit legacy schema, and the
// schema is what determines expressiveness -- so this names the newest.
const classicTarget = "1.77"

// countingWriter records whether Encode wrote anything. The stub error must fire
// BEFORE any output reaches the caller: an error alongside a partial file still
// hands the user a file, so `err != nil` alone would not prove the refusal.
type countingWriter struct {
	n     int
	calls int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.calls++
	w.n += len(p)
	return len(p), nil
}

// decodeW2025 decodes a tracked .wxx fixture through the public pipeline.
func decodeW2025(t *testing.T, path string) *wxx.Map_t {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	m, err := xmlio.NewDecoder().Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return m
}

// TestClassicDowngradeStubError is the loss contract's error half: a W2025 map
// whose <extraTerrain> carries real content cannot be written as classic,
// because Map_t holds that content only as an opaque stub and the encoder cannot
// say what dropping it would cost.
//
// The two tracked 2.06 fixtures differ in exactly the way the contract turns on,
// which is why both are here: `layers` carries 183 bytes of <extraTerrain>
// children, `blank` carries "\n". If only the erroring fixture were tested, an
// encoder that refused EVERY W2025 downgrade would pass.
func TestClassicDowngradeStubError(t *testing.T) {
	for _, tc := range []struct {
		name      string
		fixture   string
		wantErr   bool
		wantInner string // the InnerXML the fixture must carry for this to be a real test
	}{
		{
			name:      "layers: populated <extraTerrain> is an unmodeled stub -> hard error",
			fixture:   sample2025_206Layers,
			wantErr:   true,
			wantInner: "terrainAndLocation",
		},
		{
			name:    "blank: empty <extraTerrain> container loses nothing -> no stub error",
			fixture: sample2025_206,
			wantErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := decodeW2025(t, tc.fixture)

			// Guard against a vacuous pass. If the fixture ever stops carrying an
			// <extraTerrain> in the shape this case is about, the assertion below
			// proves nothing -- an absent element and an empty one both fail to
			// error, for different reasons.
			if m.ExtraTerrain == nil {
				t.Fatalf("%s: ExtraTerrain = nil; this fixture no longer exercises the stub gate", tc.fixture)
			}
			if tc.wantErr {
				if !strings.Contains(m.ExtraTerrain.InnerXML, tc.wantInner) {
					t.Fatalf("%s: ExtraTerrain.InnerXML = %q, want it to contain %q; without real stub content the error below would not be under test",
						tc.fixture, m.ExtraTerrain.InnerXML, tc.wantInner)
				}
			} else {
				if strings.TrimSpace(m.ExtraTerrain.InnerXML) != "" {
					t.Fatalf("%s: ExtraTerrain.InnerXML = %q, want whitespace only; this case is about an EMPTY container",
						tc.fixture, m.ExtraTerrain.InnerXML)
				}
				if m.ExtraTerrain.InnerXML == "" {
					t.Fatalf("%s: ExtraTerrain.InnerXML is \"\"; the emptiness test is only meaningful while the fixture carries whitespace (it carried %q)",
						tc.fixture, "\n")
				}
			}

			var w countingWriter
			err := xmlio.NewEncoder(xmlio.WithTargetVersion(classicTarget)).Encode(&w, m)

			if !tc.wantErr {
				if err != nil {
					t.Fatalf("%s -> classic %s: unexpected error: %v", tc.fixture, classicTarget, err)
				}
				if w.n == 0 {
					t.Errorf("%s -> classic %s: wrote 0 bytes, want a file", tc.fixture, classicTarget)
				}
				return
			}

			if err == nil {
				t.Fatalf("%s -> classic %s: want an error (populated <extraTerrain> is an unmodeled stub), got nil", tc.fixture, classicTarget)
			}
			if !errors.Is(err, wxx.ErrUnmodeledStubLoss) {
				t.Errorf("%s -> classic %s: error = %v, want it to wrap ErrUnmodeledStubLoss", tc.fixture, classicTarget, err)
			}
			// The refusal must precede the write. A file handed over alongside an
			// error is still a file the caller can save.
			if w.calls != 0 || w.n != 0 {
				t.Errorf("%s -> classic %s: wrote %d bytes in %d call(s), want 0: the encode must refuse before any output reaches the writer",
					tc.fixture, classicTarget, w.n, w.calls)
			}
			// The error must describe the loss, not merely announce it: the caller
			// cannot ask the model what was in a stub.
			if got := err.Error(); !strings.Contains(got, "extraTerrain") || !strings.Contains(got, "terrainAndLocation") {
				t.Errorf("%s -> classic %s: error = %q, want it to name the element and echo the stub content", tc.fixture, classicTarget, got)
			}
		})
	}
}

// TestClassicDowngradeDiagnostics is the loss contract's reporting half: a
// downgrade that drops only MODELED features succeeds and inventories them.
//
// It runs on the blank fixture because the layers fixture cannot be downgraded at
// all (its stub errors), which is itself the contract working.
func TestClassicDowngradeDiagnostics(t *testing.T) {
	m := decodeW2025(t, sample2025_206)

	var d xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&d), xmlio.WithTargetVersion(classicTarget)).Encode(&buf, m); err != nil {
		t.Fatalf("encode %s -> classic %s: %v", sample2025_206, classicTarget, err)
	}

	// The modeled losses this fixture DEMONSTRATES, proven against the audit
	// harness by TestClassicDowngradeLossInventory. hScrollbarPos/vScrollbarPos
	// are absent on purpose: the fixture carries 0.0 for both (see
	// TestClassicDowngradeScrollbarLatent).
	want := []string{
		"map/blurTerrainBG",
		"map/configuration/shape-config/shapestyle/@lineCap",
		"map/configuration/shape-config/shapestyle/@lineJoin",
		"map/maplayer/@opacity",
	}
	assertDroppedPaths(t, sample2025_206+" -> classic", want, d.Dropped)

	// Every entry must actually describe the loss. A Path with an empty Detail or
	// Reason is a bare string blob wearing a struct.
	for _, e := range d.Dropped {
		if e.Field == "" || e.Detail == "" || e.Reason == "" {
			t.Errorf("Dropped entry %q: Field=%q Detail=%q Reason=%q, want all three populated", e.Path, e.Field, e.Detail, e.Reason)
		}
	}

	// Spot-check that Detail carries the map's real values rather than a
	// restatement of Path: the fixture's 8 layers are all opacity 1.
	for _, e := range d.Dropped {
		if e.Path != "map/maplayer/@opacity" {
			continue
		}
		if !strings.Contains(e.Detail, "8 of 8 map layer(s)") {
			t.Errorf("opacity Detail = %q, want it to count the 8 layers the fixture carries", e.Detail)
		}
	}
}

// TestClassicDowngradeScrollbarLatent covers the one inventory entry no tracked
// fixture can demonstrate: map/@hScrollbarPos and map/@vScrollbarPos.
//
// The audit harness DOES show both attributes dropped on a 2.06 -> classic
// encode, but both tracked fixtures carry "0.0", and Map_t models them as plain
// float64 -- so absent and zero are the same value and the encoder cannot report
// one as a loss without inventing it. The entry is real by format (the classic
// <map> element has no such attribute) and LATENT on the samples, in the sense
// internal/v0_77/COVERAGE.md means by "latent-by-code".
//
// The non-zero source is therefore synthesized, exactly as
// TestW2025LabelStyleDropShadowGate synthesizes its drop-shadow-free source.
func TestClassicDowngradeScrollbarLatent(t *testing.T) {
	m := decodeW2025(t, sample2025_206)

	// Guard against a vacuous pass in both directions. If the fixture ever ships
	// a non-zero scrollbar position, the "latent" half below is wrong and the
	// inventory test must change with it.
	if m.HScrollbarPos != 0 || m.VScrollbarPos != 0 {
		t.Fatalf("%s: HScrollbarPos=%v VScrollbarPos=%v, want 0/0; this test asserts the entry is LATENT on the tracked fixtures",
			sample2025_206, m.HScrollbarPos, m.VScrollbarPos)
	}

	// Latent half: zero values report nothing.
	var zero xmlio.EncoderDiagnostics
	var zbuf bytes.Buffer
	if err := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&zero), xmlio.WithTargetVersion(classicTarget)).Encode(&zbuf, m); err != nil {
		t.Fatalf("encode: %v", err)
	}
	for _, e := range zero.Dropped {
		if strings.Contains(e.Path, "ScrollbarPos") {
			t.Errorf("zero-valued scrollbars reported as lost: %s; absent and 0.0 are indistinguishable in Map_t, so this invents a loss", e)
		}
	}

	// Live half: a non-zero position IS reported.
	m.HScrollbarPos, m.VScrollbarPos = 0.25, 0.5
	var live xmlio.EncoderDiagnostics
	var lbuf bytes.Buffer
	if err := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&live), xmlio.WithTargetVersion(classicTarget)).Encode(&lbuf, m); err != nil {
		t.Fatalf("encode (synthesized scrollbars): %v", err)
	}
	got := map[string]string{}
	for _, e := range live.Dropped {
		got[e.Path] = e.Detail
	}
	for path, want := range map[string]string{
		"map/@hScrollbarPos": "0.25",
		"map/@vScrollbarPos": "0.5",
	} {
		detail, ok := got[path]
		if !ok {
			t.Errorf("%s not reported for a non-zero scrollbar position; the classic <map> element cannot state it", path)
			continue
		}
		if !strings.Contains(detail, want) {
			t.Errorf("%s Detail = %q, want it to carry the lost value %s", path, detail, want)
		}
	}
}

// TestNoLossOnSameReleaseTargets pins the property the whole contract rests on:
// encoding a map as the release it already states loses nothing and reports
// nothing. It is the default target, so this is the ordinary path -- a false
// positive here would cry loss on every plain re-encode.
//
// It also pins that the loss check does not perturb the bytes: diagnostics are
// an observation, and an observation that changed the output would break the
// verbatim guarantee ADR 0002 left standing (ADR 0004 Decision 1).
func TestNoLossOnSameReleaseTargets(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fixture string
	}{
		{"classic 1.73", "../testdata/blank-2017-1.73-1.0.wxx"},
		{"classic 1.74", "../testdata/blank-2017-1.74-1.0.wxx"},
		{"classic 1.77", "../testdata/blank-2017-1.77-1.0.wxx"},
		{"classic 1.77 columns", "../testdata/2017-1.77-1.0-columns-blank.wxx"},
		{"classic 1.77 import", "../testdata/2017-1.77-1.0-import.wxx"},
		{"classic 1.77 merge-01", "../testdata/2017-1.77-1.0-merge-01.wxx"},
		{"classic 1.77 merge-02", "../testdata/2017-1.77-1.0-merge-02.wxx"},
		{"w2025 2.06 blank", sample2025_206},
		// The layers fixture carries a populated <extraTerrain> that hard-errors
		// on a classic target. Targeted at its OWN release it must sail through:
		// the stub error is a property of the target's expressiveness, not of the
		// content being unusual.
		{"w2025 2.06 layers", sample2025_206Layers},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := decodeW2025(t, tc.fixture)

			var d xmlio.EncoderDiagnostics
			var withDiag bytes.Buffer
			if err := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&d)).Encode(&withDiag, m); err != nil {
				t.Fatalf("%s: encode as its own release: %v", tc.fixture, err)
			}
			if len(d.Dropped) != 0 {
				for _, e := range d.Dropped {
					t.Errorf("%s: reported a loss encoding as its own release: %s", tc.fixture, e)
				}
			}
			if withDiag.Len() == 0 {
				t.Fatalf("%s: wrote 0 bytes", tc.fixture)
			}

			// Asking for diagnostics must not move a byte.
			var noDiag bytes.Buffer
			if err := xmlio.NewEncoder().Encode(&noDiag, m); err != nil {
				t.Fatalf("%s: encode without diagnostics: %v", tc.fixture, err)
			}
			if !bytes.Equal(withDiag.Bytes(), noDiag.Bytes()) {
				t.Errorf("%s: output differs with and without diagnostics (%d vs %d bytes); loss detection must not alter output",
					tc.fixture, withDiag.Len(), noDiag.Len())
			}
		})
	}
}

// TestClassicDowngradeLossInventory is the EVIDENCE test: it re-derives the
// inventory with the round-trip audit harness instead of trusting it.
//
// It encodes the decoded W2025 blank fixture through the classic target, diffs
// the result against the W2025 original with xmlAggregate/computeLoss, strips the
// two classes of harness finding that are not downgrade losses (target identity,
// and the classic codec gaps that classicRoundTripExpect proves classic inflicts
// on itself), and requires the residual to match what the encoder reported --
// modulo the documented zero-valued latents.
//
// The subset assertion is the load-bearing one: the encoder may not report a loss
// the harness does not show. That is the ADR 0003 failure mode -- a claim
// asserted from memory that a fixture contradicts -- made unrepeatable.
func TestClassicDowngradeLossInventory(t *testing.T) {
	f, err := os.Open(sample2025_206)
	if err != nil {
		t.Fatalf("open %s: %v", sample2025_206, err)
	}
	defer f.Close()

	var dd xmlio.DecoderDiagnostics
	m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", sample2025_206, err)
	}

	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(xmlio.WithEncoderDiagnostics(&ed), xmlio.WithTargetVersion(classicTarget)).Encode(&buf, m); err != nil {
		t.Fatalf("encode %s -> classic %s: %v", sample2025_206, classicTarget, err)
	}

	inAgg, err := xmlAggregate(stripXMLDecl(dd.Converted))
	if err != nil {
		t.Fatalf("aggregate input: %v", err)
	}
	outAgg, err := xmlAggregate(stripXMLDecl(ed.Utf8Encoded))
	if err != nil {
		t.Fatalf("aggregate output: %v", err)
	}
	harness := computeLoss(inAgg, outAgg)
	if len(harness) == 0 {
		t.Fatalf("the harness observed no loss at all downgrading %s to classic; it cannot be the evidence for an inventory", sample2025_206)
	}
	for _, l := range harness {
		t.Logf("HARNESS %s", l)
	}

	// Not downgrade losses. Every path here is justified in downgrade.go's
	// classicDowngradeLoss doc comment, and each is independently evidenced:
	// the identity entries by Release_t.identify, the codec-gap entries by
	// classicRoundTripExpect (classic loses them to ITSELF).
	notDowngrade := map[string]string{
		"attr-altered\tmap\tversion":                                            "target identity: the file states the release the caller asked for",
		"attr-dropped\tmap\trelease":                                            "target identity: a classic file states no @release",
		"attr-dropped\tmap\tschema":                                             "target identity: a classic file states no @schema",
		"attr-altered\tmap/mapkey\tviewlevel":                                   "classic codec gap: encodeMapKey writes a constant block (classic->classic loses it too)",
		"element-dropped\tmap/configuration/text-config/labelstyle":             "classic codec gap: encodeLabelStyle is a no-op (classic->classic loses it too)",
		"element-dropped\tmap/informations/information":                         "classic codec gap: encodeInformations emits an empty wrapper (classic->classic loses it too)",
		"element-dropped\tmap/informations/information/information":             "classic codec gap: as above",
		"element-dropped\tmap/informations/information/information/information": "classic codec gap: as above",
		// The container is dropped, but this fixture's is empty ("\n"), so
		// nothing is lost. The populated case never reaches a diff: it errors.
		"element-dropped\tmap/extraTerrain": "empty container: no children, no text, nothing to lose",
	}

	// Zero-valued: the harness sees the ATTRIBUTE dropped, but the value carried
	// no information and Map_t cannot tell 0.0 from absent. See
	// TestClassicDowngradeScrollbarLatent, which synthesizes the non-zero case.
	zeroValued := map[string]bool{
		"attr-dropped\tmap\thScrollbarPos": true,
		"attr-dropped\tmap\tvScrollbarPos": true,
	}

	// Map each surviving harness entry to the inventory Path it evidences.
	harnessToPath := map[string]string{
		"attr-dropped\tmap/maplayer\topacity":                               "map/maplayer/@opacity",
		"attr-dropped\tmap/configuration/shape-config/shapestyle\tlineCap":  "map/configuration/shape-config/shapestyle/@lineCap",
		"attr-dropped\tmap/configuration/shape-config/shapestyle\tlineJoin": "map/configuration/shape-config/shapestyle/@lineJoin",
		"element-dropped\tmap/blurTerrainBG":                                "map/blurTerrainBG",
	}

	evidenced := map[string]bool{}
	var unexplained []string
	for _, l := range harness {
		if _, ok := notDowngrade[l]; ok {
			continue
		}
		if zeroValued[l] {
			continue
		}
		if p, ok := harnessToPath[l]; ok {
			evidenced[p] = true
			continue
		}
		unexplained = append(unexplained, l)
	}

	// A harness finding this test cannot account for is a loss nobody has
	// classified. Failing here is correct: it must be triaged into an inventory
	// entry, a codec gap, or identity -- deliberately.
	if len(unexplained) > 0 {
		t.Errorf("the harness shows loss this inventory does not account for:\n  + %s\nTriage each into downgrade.go's inventory, a documented codec gap, or target identity.",
			strings.Join(unexplained, "\n  + "))
	}

	// Guard against a vacuous pass: if the subtraction ever leaves nothing, the
	// comparison below is between two empty sets and proves nothing.
	if len(evidenced) == 0 {
		t.Fatalf("no harness finding survived the controls, so the inventory is not under test")
	}

	reported := map[string]bool{}
	for _, e := range ed.Dropped {
		reported[e.Path] = true
	}

	// THE load-bearing direction: never report a loss the harness did not show.
	for p := range reported {
		if !evidenced[p] {
			t.Errorf("encoder reports %q as a downgrade loss, but the harness does not show it on %s: an inventory entry must be demonstrated, not asserted", p, sample2025_206)
		}
	}
	// And the converse: a demonstrated loss that goes unreported is silent data
	// loss, which is the failure ADR 0004 calls the worse one.
	for p := range evidenced {
		if !reported[p] {
			t.Errorf("the harness shows %q dropped downgrading %s to classic, but the encoder reported no loss for it", p, sample2025_206)
		}
	}
}

// assertDroppedPaths compares reported loss paths to an expected set.
func assertDroppedPaths(t *testing.T, label string, want []string, got []xmlio.DroppedFeature_t) {
	t.Helper()
	var gotPaths []string
	for _, e := range got {
		gotPaths = append(gotPaths, e.Path)
	}
	sort.Strings(gotPaths)
	sorted := append([]string(nil), want...)
	sort.Strings(sorted)
	if strings.Join(gotPaths, "\n") != strings.Join(sorted, "\n") {
		t.Errorf("%s: dropped-feature paths =\n  %s\nwant\n  %s", label, strings.Join(gotPaths, "\n  "), strings.Join(sorted, "\n  "))
	}
}

// stripXMLDecl removes a leading <?xml ...?> declaration.
//
// The audit harness tokenizes with encoding/xml, which rejects version="1.1" --
// the declaration every W2025 file opens with -- so a W2025 document cannot be
// aggregated with its declaration attached. xmlAggregate ignores processing
// instructions anyway, so dropping it costs the diff nothing. Classic documents
// (version="1.0") pass through this unharmed, which is why roundtrip_2017_test.go
// never needed it.
func stripXMLDecl(data []byte) []byte {
	trimmed := bytes.TrimLeft(data, "\xef\xbb\xbf \t\r\n")
	if !bytes.HasPrefix(trimmed, []byte("<?xml")) {
		return data
	}
	i := bytes.Index(data, []byte("?>"))
	if i < 0 {
		return data
	}
	return data[i+2:]
}
