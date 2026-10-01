// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// This file holds the downgrade-loss tests (#32, ADR 0004 Decision 7). The
// classic downgrade inventory and the tests that derived it from the audit
// harness went with the classic codec (issue #103). What is left is the
// property the loss contract rests on for the one remaining format -- a target
// reports no loss on content it can express -- and the place the next
// inventory's tests go: W2025-to-W2025 loss (#92). downgrade_internal_test.go
// holds the other half, that a schema with no inventory is an error.

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

// TestNoLossOnSameReleaseTargets pins the property the whole contract rests on:
// encoding a map as the release it already states loses nothing and reports
// nothing. It is the default target, so this is the ordinary path -- a false
// positive here would cry loss on every plain re-encode.
//
// It also pins that the loss check does not perturb the bytes: diagnostics are
// an observation, and an observation that changed the output would break the
// verbatim guarantee ADR 0002 left standing (ADR 0004 Decision 1).
func TestNoLossOnSameReleaseTargets(t *testing.T) {
	cases := []struct {
		name    string
		fixture string
	}{
		{"w2025 2.06 blank", sample2025_206},
		// The layers fixture places terrain in <extraTerrain>, which a classic
		// target used to report as lost. Targeted at its OWN release it must
		// report nothing: the loss is a property of the target's expressiveness,
		// not of the content being unusual.
		{"w2025 2.06 layers beta", sample2025_206LayersBeta},
	}

	// Guard against a registered release drifting out of test: the claim is "no
	// supported target reports a loss on its own content", so every registered
	// application version must be the own release of some case.
	covered := map[string]bool{}
	for _, tc := range cases {
		covered[decodeW2025(t, tc.fixture).MetaData.Version.App.Raw] = true
	}
	for _, c := range codecsForTest() {
		for _, a := range c.AcceptedApps().Apps {
			if !covered[a.Version] {
				t.Errorf("version %q is registered but no case encodes a map as it: add a fixture it wrote", a.Version)
			}
		}
	}

	for _, tc := range cases {
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

// stripXMLDecl removes a leading <?xml ...?> declaration.
//
// encoding/xml rejects version="1.1" -- the declaration every W2025 file opens
// with -- so a test that tokenizes a W2025 document must drop it first.
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
