// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestW2025MapAttrsMatchSource decodes every tracked W2025 fixture, encodes it
// as the version it states (see sameVersionTarget), and asserts the <map> start
// tag states every attribute the source does, with the same names in the same
// order and the same values (issue #111). That includes the identity: every
// fixture's version is registered (2.08 by issue #73), so none may differ.
//
// The bug this pins: floats reformatted an exponent spelling with %f, so the
// 2.07 and 2.08 rows maps' vScrollbarPos="4.263256414560601E-14" came back as
// "0.0". No test compared <map>'s own attributes, and the 2.06 rows map states
// "0.0", so nothing noticed.
func TestW2025MapAttrsMatchSource(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "testdata", "2025-*.wxx"))
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("glob W2025 fixtures: %d found, err %v", len(fixtures), err)
	}
	sawExponent := false
	for _, path := range fixtures {
		fixture := filepath.Base(path)
		t.Run(fixture, func(t *testing.T) {
			in, _, out := encodeFixture(t, fixture)
			src := startTagAttrs(in, "map")
			for _, a := range src {
				for _, kv := range a {
					if strings.Contains(kv[1], "E") {
						sawExponent = true
					}
				}
			}
			compareStartTags(t, fixture, "map", src, startTagAttrs(out, "map"))
		})
	}
	// Vacuity: without a value in exponent spelling, this test cannot see #111.
	if !sawExponent {
		t.Errorf("no fixture's <map> states a value in exponent spelling; the 2.07 and 2.08 rows maps should")
	}
}
