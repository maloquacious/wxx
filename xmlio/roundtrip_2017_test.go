// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// This file is the first automated h2017 (classic) codec test. It is an AUDIT
// harness, not a fidelity check: for every classic fixture it decodes the file,
// re-encodes it, and diffs the ORIGINAL on-disk UTF-8 XML against the
// re-encoded UTF-8 XML at the element/attribute-set level. The point is to
// inventory exactly what the frozen classic codec drops or alters on a round
// trip -- losses that a Map_t-level comparison is structurally blind to, since
// decode and encode ignore the same fields symmetrically.
//
// The per-fixture loss set is asserted against a documented expectation (see
// classicRoundTripExpect below), mirroring how the h2025 coverage-matrix test
// asserts its matrix: any drift (a newly dropped/altered field, or a
// previously dropped field that starts surviving) trips the test so a
// maintainer must update the inventory in xmlio/internal/v0_77/COVERAGE.md
// deliberately. Run with `-v` to dump the full per-fixture loss set; the
// harness doubles as the report generator for that document.

const classicInputDir = "../testdata/"

// rowsFixture decodes but cannot be re-encoded: classic ROWS encode is a
// documented hard-error, refused by encode.go's verifyOrientation before any
// output is built (issue #20; the refusal used to be an assert inside
// encodeTiles). TestClassicRowsRefusedUpFront asserts what the error says.
const rowsFixture = "2017-1.77-1.0-rows-blank.wxx"

// classicFixtures are the eight classic 2017 fixtures under testdata/.
var classicFixtures = []string{
	"blank-2017-1.73-1.0.wxx",
	"blank-2017-1.74-1.0.wxx",
	"blank-2017-1.77-1.0.wxx",
	"2017-1.77-1.0-columns-blank.wxx",
	"2017-1.77-1.0-import.wxx",
	"2017-1.77-1.0-merge-01.wxx",
	"2017-1.77-1.0-merge-02.wxx",
	rowsFixture,
}

// classicRoundTrip decodes a classic fixture (capturing the input UTF-8 XML in
// diagnostics), re-encodes it (capturing the output UTF-8 XML), and returns the
// element/attribute-set loss between input and output. If encode hard-errors
// (the ROWS case), it returns a nil loss set and the encode error.
func classicRoundTrip(t *testing.T, fixture string) (loss []string, encodeErr error) {
	t.Helper()
	path := classicInputDir + fixture
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var d xmlio.DecoderDiagnostics
	m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&d)).Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", fixture, err)
	}
	if len(d.Converted) == 0 {
		t.Fatalf("decode %s: diagnostics.Converted is empty", fixture)
	}

	// The round trip writes the version the file states, which since issue #45 the
	// caller names rather than the encoder assuming: a CLIENT may read provenance
	// and choose it as the target, and that is what a round trip means.
	var e xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(m.MetaData.Version.App.Raw, xmlio.WithEncoderDiagnostics(&e)).Encode(&buf, m); err != nil {
		return nil, err
	}
	if len(e.Utf8Encoded) == 0 {
		t.Fatalf("encode %s: diagnostics.Utf8Encoded is empty", fixture)
	}

	inAgg, err := xmlAggregate(d.Converted)
	if err != nil {
		t.Fatalf("aggregate input %s: %v", fixture, err)
	}
	outAgg, err := xmlAggregate(e.Utf8Encoded)
	if err != nil {
		t.Fatalf("aggregate output %s: %v", fixture, err)
	}
	return computeLoss(inAgg, outAgg), nil
}

// TestRoundTrip2017LossInventory is the executable inventory. For each classic
// fixture it asserts the on-disk round-trip loss set against the documented
// expectation. Run with -v to dump the full per-fixture loss set.
func TestRoundTrip2017LossInventory(t *testing.T) {
	for _, fixture := range classicFixtures {
		t.Run(fixture, func(t *testing.T) {
			loss, encErr := classicRoundTrip(t, fixture)

			if fixture == rowsFixture {
				if encErr == nil {
					t.Fatalf("%s: expected encode hard-error (classic ROWS), got nil", fixture)
				}
				t.Logf("%s: round-trip not possible -- encode hard-errors: %v", fixture, encErr)
				return
			}
			if encErr != nil {
				t.Fatalf("%s: unexpected encode error: %v", fixture, encErr)
			}

			for _, l := range loss {
				t.Logf("LOSS %s :: %s", fixture, l)
			}

			want := classicRoundTripExpect[fixture]
			assertLossSet(t, fixture, want, loss)
		})
	}
}

// TestRoundTrip2017RowsHardError is the focused ROWS subtest: the ROWS fixture
// must DECODE successfully but its re-encode must return a non-nil error
// (classic ROWS encode is intentionally unimplemented -- COVERAGE.md).
//
// It asserts only that the encode FAILS, which is deliberately weaker than
// TestClassicRowsRefusedUpFront: this one holds the round-trip inventory's
// premise (the ROWS fixture has no re-encode to diff), the other holds what the
// refusal tells the caller. Weakening either would not weaken the other.
func TestRoundTrip2017RowsHardError(t *testing.T) {
	path := classicInputDir + rowsFixture
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	m, err := xmlio.NewDecoder().Decode(f)
	if err != nil {
		t.Fatalf("decode %s: want success, got %v", rowsFixture, err)
	}
	if m.Tiles == nil {
		t.Fatalf("decode %s: nil Tiles", rowsFixture)
	}
	if got := m.HexOrientation; got != "ROWS" {
		t.Fatalf("decode %s: HexOrientation = %q, want ROWS", rowsFixture, got)
	}

	var buf bytes.Buffer
	if err := xmlio.NewEncoder(m.MetaData.Version.App.Raw).Encode(&buf, m); err == nil {
		t.Fatalf("encode %s: want non-nil error (classic ROWS is a hard-error), got nil", rowsFixture)
	}
}

// assertLossSet compares the observed loss set to the expected set, failing
// with an explicit list of unexpected additions and removals so a maintainer
// can update the inventory deliberately.
func assertLossSet(t *testing.T, fixture string, want, got []string) {
	t.Helper()
	wantSet := map[string]bool{}
	for _, w := range want {
		wantSet[w] = true
	}
	gotSet := map[string]bool{}
	for _, g := range got {
		gotSet[g] = true
	}

	var added, removed []string
	for g := range gotSet {
		if !wantSet[g] {
			added = append(added, g)
		}
	}
	for w := range wantSet {
		if !gotSet[w] {
			removed = append(removed, w)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)

	if len(added) > 0 || len(removed) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%s: round-trip loss set drifted from the documented inventory.\n", fixture)
		if len(added) > 0 {
			b.WriteString("  UNEXPECTED (present now, not in inventory):\n")
			for _, a := range added {
				fmt.Fprintf(&b, "    + %s\n", a)
			}
		}
		if len(removed) > 0 {
			b.WriteString("  MISSING (in inventory, no longer observed):\n")
			for _, r := range removed {
				fmt.Fprintf(&b, "    - %s\n", r)
			}
		}
		b.WriteString("  Update classicRoundTripExpect and xmlio/internal/v0_77/COVERAGE.md together.")
		t.Error(b.String())
	}
}

// classicRoundTripExpect is the documented per-fixture loss inventory, derived
// from the harness's first real run. It is the machine-checkable twin of the
// "Round-trip loss inventory (executable)" section in
// xmlio/internal/v0_77/COVERAGE.md; keep the two in sync.
//
// Shared across every classic fixture (blank, columns, import, merge):
//   - <informations>/<information> lore tree dropped (encode.go:438-442 emits
//     an empty <informations> wrapper). The nesting depth listed per fixture is
//     the deepest <information> chain that fixture actually carries.
//   - configuration <text-config>/<labelstyle> dropped (encode.go:483-507:
//     encodeLabelStyle is a commented-out no-op; the <text-config> wrapper is
//     still emitted, empty).
//
// Observed on a subset only:
//   - map/mapkey @viewlevel altered "null" -> "WORLD": encode.go:253-258 emits
//     a hardcoded constant <mapkey> block. The samples happen to match that
//     block on every other attribute, so viewlevel is the only OBSERVED
//     alteration (the full constant-block override is latent-by-code). The
//     1.74 and columns-blank fixtures already carry viewlevel="WORLD", so they
//     show no mapkey drift at all.
var classicRoundTripExpect = map[string][]string{
	"blank-2017-1.73-1.0.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"blank-2017-1.74-1.0.wxx": {
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
	},
	"blank-2017-1.77-1.0.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"2017-1.77-1.0-columns-blank.wxx": {
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
	},
	"2017-1.77-1.0-import.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"2017-1.77-1.0-merge-01.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	"2017-1.77-1.0-merge-02.wxx": {
		"attr-altered\tmap/mapkey\tviewlevel",
		"element-dropped\tmap/informations/information",
		"element-dropped\tmap/informations/information/information",
		"element-dropped\tmap/informations/information/information/information",
	},
	// rowsFixture has no loss set: it decodes but re-encode hard-errors, so the
	// round trip is impossible. Asserted separately (see the ROWS handling in
	// TestRoundTrip2017LossInventory and TestRoundTrip2017RowsHardError).
}
