// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// hostile is user text carrying every character fmt's %q got wrong: the three
// markup characters XML requires escaped in a double-quoted attribute, a
// character that is legal raw but conventionally escaped (>), an apostrophe
// Worldographer writes raw, the whitespace a parser would normalize away, and a
// non-ASCII letter.
const hostile = "Salt & \"Pepper\" <Isles> 'n' tab\there\nnew line café"

// TestAttributeEscaping asserts that user text containing XML's markup
// characters is written as a well-formed attribute and reads back unchanged,
// through both codecs (issue #71). Both write the é as caf&#233;, the decimal
// character reference Worldographer writes for every non-ASCII attribute
// character in classic and W2025 files alike (issue #96).
//
// The bug this pins: every attribute was written with fmt's %q, a Go string
// literal. A title of `Salt & "Pepper"` came out as title="Salt & \"Pepper\"",
// which is not XML: the & is bare and the attribute ends at the backslash.
// Nothing caught it because no fixture carries such text, so %q and correct
// escaping agreed on every byte the suite ever wrote. The source is synthesized
// for that reason.
//
// Decoding the output back through the public pipeline is the well-formedness
// check: the decoder runs encoding/xml, which refuses a bare & or a broken
// attribute. The exact spellings are asserted as well, so a fix that stayed
// well-formed while changing the text -- dropping the tab, say -- would fail.
func TestAttributeEscaping(t *testing.T) {
	for _, tc := range []struct {
		fixture string
		app     string
		// lore reports whether the codec writes <information>. The classic
		// encoder does not yet (v0_77/COVERAGE.md).
		lore bool
	}{
		{fixture: "2025-2.06-13x11-941577-layers-beta.wxx", app: "2.06", lore: true},
		{fixture: "2017-1.77-1.0-columns-blank.wxx", app: "1.77"},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			m, err := xmlio.ReadFile(filepath.Join("..", "testdata", tc.fixture))
			if err != nil {
				t.Fatalf("read %s: %v", tc.fixture, err)
			}

			// Each target is a field a user types, reached through a setter and a
			// getter so the same list drives both the write and the read-back.
			type target struct {
				name string
				set  func(*wxx.Map_t)
				get  func(*wxx.Map_t) string
			}
			targets := []target{
				{"labelstyle/@name",
					func(m *wxx.Map_t) { m.Configuration.TextConfig.LabelStyles[0].Name = hostile },
					func(m *wxx.Map_t) string { return m.Configuration.TextConfig.LabelStyles[0].Name }},
				{"maplayer/@name",
					func(m *wxx.Map_t) { m.MapLayers[0].Name = hostile },
					func(m *wxx.Map_t) string { return m.MapLayers[0].Name }},
				{"feature/@mapLayer",
					func(m *wxx.Map_t) { m.Features[0].MapLayer = hostile },
					func(m *wxx.Map_t) string { return m.Features[0].MapLayer }},
			}
			if tc.lore {
				targets = append(targets,
					target{"information/@title",
						func(m *wxx.Map_t) { m.Informations.Informations[0].Title = hostile },
						func(m *wxx.Map_t) string { return m.Informations.Informations[0].Title }},
					target{"information/@rulers",
						func(m *wxx.Map_t) { v := hostile; m.Informations.Informations[0].Rulers = &v },
						func(m *wxx.Map_t) string {
							if p := m.Informations.Informations[0].Rulers; p != nil {
								return *p
							}
							return "<nil>"
						}},
				)
			}
			for _, tg := range targets {
				tg.set(m)
			}

			var ed xmlio.EncoderDiagnostics
			var buf bytes.Buffer
			if err := xmlio.NewEncoder(tc.app, xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
				t.Fatalf("encode: %v", err)
			}

			want := `"Salt &amp; &quot;Pepper&quot; &lt;Isles&gt; 'n' tab&#9;here&#10;new line caf&#233;"`
			if got := strings.Count(string(ed.Utf8Encoded), want); got != len(targets) {
				t.Errorf("found the escaped value %d time(s), want %d (one per target): %s", got, len(targets), want)
			}
			if bytes.Contains(ed.Utf8Encoded, []byte(`\"Pepper\"`)) {
				t.Errorf(`output contains \"Pepper\": an attribute was written as a Go string literal`)
			}

			back, err := xmlio.NewDecoder().Decode(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("re-decode: %v -- the encoder wrote a file that is not well-formed XML", err)
			}
			for _, tg := range targets {
				if got := tg.get(back); got != hostile {
					t.Errorf("%s read back as %q, want %q", tg.name, got, hostile)
				}
			}
		})
	}
}
