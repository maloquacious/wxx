// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
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

// TestClassicDowngradeExtraTerrain covers the terrain-layers loss (issue #34).
//
// Until #34 this was the loss contract's ERROR half: <extraTerrain> was an
// opaque stub, so a W2025 map carrying real content could not be written as
// classic at all. It is modeled now, so the downgrade succeeds and the loss is
// reported with what was dropped. The contract's rule -- a modeled loss is
// reported, not refused -- is what this pins.
//
// The two tracked 2.06 fixtures differ in exactly the way the report turns on:
// `layers` places one terrain on "Terrain Layer", `blank` carries an empty
// container. If only the fixture with a placement were tested, an encoder that
// reported <extraTerrain> for EVERY W2025 downgrade would pass.
func TestClassicDowngradeExtraTerrain(t *testing.T) {
	for _, tc := range []struct {
		name       string
		fixture    string
		wantDetail string // "" means no map/extraTerrain entry may be reported
	}{
		{
			name:       "layers: one placement on one layer is reported",
			fixture:    sample2025_206LayersBeta,
			wantDetail: `1 terrain placement(s) on 1 layer(s) are dropped ("Terrain Layer": 1)`,
		},
		{
			name:    "blank: an empty container loses nothing and is not reported",
			fixture: sample2025_206,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := decodeW2025(t, tc.fixture)

			// Guard against a vacuous pass: each case is only under test while
			// its fixture carries <extraTerrain> in the shape the case is about.
			if m.ExtraTerrain == nil {
				t.Fatalf("%s: ExtraTerrain = nil; this fixture no longer exercises the report", tc.fixture)
			}
			if populated := len(m.ExtraTerrain.MapLayers) != 0; populated != (tc.wantDetail != "") {
				t.Fatalf("%s: ExtraTerrain has %d layer(s); this case needs the fixture %s",
					tc.fixture, len(m.ExtraTerrain.MapLayers), map[bool]string{true: "to place terrain", false: "to carry an empty container"}[tc.wantDetail != ""])
			}

			var d xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(classicTarget, xmlio.WithEncoderDiagnostics(&d)).Encode(&buf, m); err != nil {
				t.Fatalf("%s -> classic %s: %v -- a modeled loss must be reported, not refused", tc.fixture, classicTarget, err)
			}
			if buf.Len() == 0 {
				t.Errorf("%s -> classic %s: wrote 0 bytes, want a file", tc.fixture, classicTarget)
			}

			var got *xmlio.DroppedFeature_t
			for i := range d.Dropped {
				if d.Dropped[i].Path == "map/extraTerrain" {
					got = &d.Dropped[i]
				}
			}
			if tc.wantDetail == "" {
				if got != nil {
					t.Errorf("%s -> classic: reported %+v, want no map/extraTerrain entry for an empty container", tc.fixture, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("%s -> classic: no map/extraTerrain entry reported; the placement is dropped silently", tc.fixture)
			}
			if !strings.Contains(got.Detail, tc.wantDetail) {
				t.Errorf("%s -> classic: Detail = %q, want it to contain %q", tc.fixture, got.Detail, tc.wantDetail)
			}
			if got.Field == "" || got.Reason == "" {
				t.Errorf("%s -> classic: Field=%q Reason=%q, want both populated", tc.fixture, got.Field, got.Reason)
			}
		})
	}
}

// TestClassicDowngradeDiagnostics is the loss contract's reporting half: a
// downgrade that drops only MODELED features succeeds and inventories them.
//
// It runs on the blank fixture; the layers fixture's one extra loss,
// map/extraTerrain, is TestClassicDowngradeExtraTerrain's.
func TestClassicDowngradeDiagnostics(t *testing.T) {
	m := decodeW2025(t, sample2025_206)

	var d xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(classicTarget, xmlio.WithEncoderDiagnostics(&d)).Encode(&buf, m); err != nil {
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
		"map/configuration/text-config/labelstyle/@dropShadow*",
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

	// The same demand of the @dropShadow* entry (issue #36). It must name the
	// styles and what each spells, not merely that a trio was dropped: a caller
	// recovering the shadows needs the values, and "the trio is dropped" is a
	// restatement of Path.
	for _, e := range d.Dropped {
		if e.Path != "map/configuration/text-config/labelstyle/@dropShadow*" {
			continue
		}
		for _, want := range []string{"label style(s)", `"Nation"`, "color=", "radius=", "spread="} {
			if !strings.Contains(e.Detail, want) {
				t.Errorf("dropShadow Detail = %q, want it to contain %q", e.Detail, want)
			}
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
	if err := xmlio.NewEncoder(classicTarget, xmlio.WithEncoderDiagnostics(&zero)).Encode(&zbuf, m); err != nil {
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
	if err := xmlio.NewEncoder(classicTarget, xmlio.WithEncoderDiagnostics(&live)).Encode(&lbuf, m); err != nil {
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
		// The layers fixture places terrain in <extraTerrain>, which a classic
		// target reports as lost. Targeted at its OWN release it must report
		// nothing: the loss is a property of the target's expressiveness, not of
		// the content being unusual.
		{"w2025 2.06 layers beta", sample2025_206LayersBeta},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := decodeW2025(t, tc.fixture)

			// "Its own release" is now something the caller says out loud: the
			// encoder has no default target (issue #45), so this reads the version
			// the fixture states and names it. That is a CLIENT reading provenance
			// and choosing a target, which is exactly what this test means by
			// "encode as its own release".
			own := m.MetaData.Version.App.Raw
			var d xmlio.EncoderDiagnostics
			var withDiag bytes.Buffer
			if err := xmlio.NewEncoder(own, xmlio.WithEncoderDiagnostics(&d)).Encode(&withDiag, m); err != nil {
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
			if err := xmlio.NewEncoder(own).Encode(&noDiag, m); err != nil {
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
	// The blank fixture evidences every inventory entry but one. Its
	// <extraTerrain> is an empty container, so the terrain-layers entry needs the
	// layers fixture, whose one placement the harness sees dropped (issue #34).
	// Each run must account for everything its own harness shows, so the blank
	// run treats the empty container as nothing lost, and the layers run maps
	// the dropped elements to the entry they evidence.
	t.Run(sample2025_206, func(t *testing.T) {
		checkClassicDowngradeInventory(t, sample2025_206, map[string]string{
			// The container is dropped, but this fixture's is empty, so nothing
			// is lost and nothing is reported.
			"element-dropped\tmap/extraTerrain": "empty container: no children, no text, nothing to lose",
		}, nil)
	})
	t.Run(sample2025_206LayersBeta, func(t *testing.T) {
		checkClassicDowngradeInventory(t, sample2025_206LayersBeta, nil, map[string]string{
			"element-dropped\tmap/extraTerrain":                             "map/extraTerrain",
			"element-dropped\tmap/extraTerrain/mapLayer":                    "map/extraTerrain",
			"element-dropped\tmap/extraTerrain/mapLayer/terrainAndLocation": "map/extraTerrain",
			// The fixture's feature labels carry the drop-shadow trio, which
			// the harness could not see until #34 let this fixture downgrade.
			"attr-dropped\tmap/features/feature/label\tdropShadowColor":  "map/features/feature/label/@dropShadow*",
			"attr-dropped\tmap/features/feature/label\tdropShadowRadius": "map/features/feature/label/@dropShadow*",
			"attr-dropped\tmap/features/feature/label\tdropShadowSpread": "map/features/feature/label/@dropShadow*",
		})
	})
}

// checkClassicDowngradeInventory runs the inventory evidence check on one
// fixture. extraNotDowngrade and extraHarnessToPath add the harness findings
// particular to that fixture to the shared tables below.
func checkClassicDowngradeInventory(t *testing.T, fixture string, extraNotDowngrade, extraHarnessToPath map[string]string) {
	f, err := os.Open(fixture)
	if err != nil {
		t.Fatalf("open %s: %v", fixture, err)
	}
	defer f.Close()

	var dd xmlio.DecoderDiagnostics
	m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&dd)).Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", fixture, err)
	}

	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(classicTarget, xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("encode %s -> classic %s: %v", fixture, classicTarget, err)
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
		t.Fatalf("the harness observed no loss at all downgrading %s to classic; it cannot be the evidence for an inventory", fixture)
	}
	for _, l := range harness {
		t.Logf("HARNESS %s", l)
	}

	// Not downgrade losses. Every path here is justified in downgrade.go's
	// classicDowngradeLoss doc comment, and each is independently evidenced:
	// the identity entries by the classic codec writing the identity of the app it
	// was given, the codec-gap entries by classicRoundTripExpect (classic loses
	// them to ITSELF).
	notDowngrade := map[string]string{
		"attr-altered\tmap\tversion":                                            "target identity: the file states the release the caller asked for",
		"attr-dropped\tmap\trelease":                                            "target identity: a classic file states no @release",
		"attr-dropped\tmap\tschema":                                             "target identity: a classic file states no @schema",
		"attr-altered\tmap/mapkey\tviewlevel":                                   "classic codec gap: encodeMapKey writes a constant block (classic->classic loses it too)",
		"element-dropped\tmap/informations/information":                         "classic codec gap: encodeInformations emits an empty wrapper (classic->classic loses it too)",
		"element-dropped\tmap/informations/information/information":             "classic codec gap: as above",
		"element-dropped\tmap/informations/information/information/information": "classic codec gap: as above",
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
		// Issue #36. These three lines did not exist before the classic
		// <labelstyle> encode gap was closed: the harness reported
		// `element-dropped ... labelstyle` and the attribute-level loss was
		// masked by it. One inventory entry evidences all three, because the
		// trio is one feature (see labelStyleDropShadowLoss).
		"attr-dropped\tmap/configuration/text-config/labelstyle\tdropShadowColor":  "map/configuration/text-config/labelstyle/@dropShadow*",
		"attr-dropped\tmap/configuration/text-config/labelstyle\tdropShadowRadius": "map/configuration/text-config/labelstyle/@dropShadow*",
		"attr-dropped\tmap/configuration/text-config/labelstyle\tdropShadowSpread": "map/configuration/text-config/labelstyle/@dropShadow*",
	}

	for k, v := range extraNotDowngrade {
		notDowngrade[k] = v
	}
	for k, v := range extraHarnessToPath {
		harnessToPath[k] = v
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
			t.Errorf("encoder reports %q as a downgrade loss, but the harness does not show it on %s: an inventory entry must be demonstrated, not asserted", p, fixture)
		}
	}
	// And the converse: a demonstrated loss that goes unreported is silent data
	// loss, which is the failure ADR 0004 calls the worse one.
	for p := range evidenced {
		if !reported[p] {
			t.Errorf("the harness shows %q dropped downgrading %s to classic, but the encoder reported no loss for it", p, fixture)
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
