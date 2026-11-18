// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// Issue #164: a map that uses a custom terrain tile must decode, and must keep
// the tile's definition through a copy.
//
// Worldographer writes the definition of each non-stock (or edited) terrain,
// feature and texture type the map uses inside <configuration>, as children of
// <terrain-config>, <feature-config> and <texture-config>. Issue #163 made decode
// refuse unmodeled children, which refused every such map; before #163 the
// definitions were dropped silently. No fixture has a custom type -- they are
// paid add-ons -- so the content below is SYNTHETIC. Its element and attribute
// names follow the writer's string constants in the 2.08 jar, but the tests do
// not depend on them: the codec keeps the inner XML verbatim and models none of
// it, so any well-formed content must survive byte for byte.
const (
	customTerrainConfig = "\n<terraintype type=\"Custom Swamp\" elevation=\"0\" similarTerrain=\"Swamp:::Marsh\" iconSize=\"1.0\" isDefault=\"false\" offsetXDecimal=\"0.0\" offsetYDecimal=\"0.0\" color=\"null\" >iVBORw0KGgo=&amp;</terraintype>\n<terraintype type=\"Ice &quot;Shelf&quot;\" elevation=\"-10\" similarTerrain=\"\" iconSize=\"1.0\" isDefault=\"false\" offsetXDecimal=\"0.5\" offsetYDecimal=\"0.5\" color=\"0.5,0.5,1.0,1.0\" >AAAA</terraintype>\n  "
	customFeatureConfig = "\n<featuretype type=\"Custom Tower\" iconSize=\"1.0\" iconHeight=\"1.0\" offsetX=\"0.0\" offsetY=\"0.0\" isDefault=\"false\" alternateIcon=\"null\" >AAAA</featuretype>\n  "
	customTextureConfig = "\n<texturetype type=\"Custom Stone\" width=\"100.0\" isDefault=\"false\" >AAAA</texturetype>\n  "
)

// withCustomTypes returns the 2.08 blank fixture encoded as 2.08, with the
// synthetic custom types spliced into the three config elements: the document a
// Worldographer user with custom tiles would have saved.
func withCustomTypes(t *testing.T) []byte {
	t.Helper()
	m, err := xmlio.ReadFile(filepath.Join("..", "testdata", "2025-2.08-13x11-941577-blank.wxx"))
	if err != nil {
		t.Fatalf("decode blank: %v", err)
	}
	doc, err := xmlio.MarshalXML(m, "2.08")
	if err != nil {
		t.Fatalf("encode blank: %v", err)
	}
	for _, r := range []struct{ name, content string }{
		{"terrain-config", customTerrainConfig},
		{"feature-config", customFeatureConfig},
		{"texture-config", customTextureConfig},
	} {
		empty := []byte("  <" + r.name + ">\n  </" + r.name + ">\n")
		if bytes.Count(doc, empty) != 1 {
			t.Fatalf("encoded blank has %d empty <%s> elements, want 1", bytes.Count(doc, empty), r.name)
		}
		doc = bytes.Replace(doc, empty, []byte("  <"+r.name+">"+r.content+"</"+r.name+">\n"), 1)
	}
	return doc
}

// decodeXML decodes marshaled XML (no header, no UTF-16, no gzip) through the
// public decoder; see decodeRawXML in chimera_test.go.
func decodeXML(doc []byte) (*wxx.Map_t, error) {
	buf := append([]byte("<?xml version='1.1' encoding='utf-8'?>\n"), doc...)
	return xmlio.NewDecoder(xmlio.WithSkipUncompress(), xmlio.WithUTF16BEInput(false)).Decode(bytes.NewReader(buf))
}

// TestCustomTypesSurviveRoundTrip is the bug: before the fix, decode refused
// this document with ErrUnmodeledContent at <terraintype>.
func TestCustomTypesSurviveRoundTrip(t *testing.T) {
	src := withCustomTypes(t)
	m, err := decodeXML(src)
	if err != nil {
		t.Fatalf("decode a map with custom types: %v", err)
	}
	if got := m.Configuration.TerrainConfig[0].InnerXML; got != customTerrainConfig {
		t.Errorf("TerrainConfig[0].InnerXML = %q, want %q", got, customTerrainConfig)
	}
	for _, app := range []string{"2.06", "2.07", "2.08"} {
		out, err := xmlio.MarshalXML(m, app)
		if err != nil {
			t.Fatalf("encode as %s: %v", app, err)
		}
		want := src
		if app != "2.08" {
			want = bytes.Replace(src, []byte(`version="2.08"`), []byte(`version="`+app+`"`), 1)
		}
		if !bytes.Equal(out, want) {
			t.Errorf("encode as %s: output differs from source (%d bytes vs %d)", app, len(out), len(want))
		}
	}
}

// TestCustomTypesSurviveFullPipeline runs the same map through gzip and UTF-16,
// as cmd/copy does.
func TestCustomTypesSurviveFullPipeline(t *testing.T) {
	m, err := decodeXML(withCustomTypes(t))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	var wxxFile bytes.Buffer
	if err := xmlio.NewEncoder("2.08").Encode(&wxxFile, m); err != nil {
		t.Fatalf("encode: %v", err)
	}
	m2, err := xmlio.NewDecoder().Decode(&wxxFile)
	if err != nil {
		t.Fatalf("decode encoded file: %v", err)
	}
	assertConfigSectionsPreserved(t, "full pipeline", m, m2)
}

// TestConfigWrapperAttributeStillRefused: the wrappers keep their content
// verbatim, but an unknown attribute on the wrapper itself is not part of that
// content and would be dropped, so it is still refused (issue #161).
func TestConfigWrapperAttributeStillRefused(t *testing.T) {
	doc := bytes.Replace(withCustomTypes(t), []byte("<terrain-config>"), []byte(`<terrain-config future="1">`), 1)
	_, err := decodeXML(doc)
	if !errors.Is(err, wxx.ErrUnmodeledContent) {
		t.Fatalf("decode = %v, want ErrUnmodeledContent", err)
	}
	if !strings.Contains(err.Error(), "map/configuration[1]/terrain-config[1]: unmodeled attribute @future") {
		t.Errorf("error does not name the attribute's path: %v", err)
	}
}

// TestMalformedConfigContentRefusedOnEncode: the codec cannot check what it does
// not model, but it will not write content that is not well-formed XML, which
// would leave a file Worldographer cannot open.
func TestMalformedConfigContentRefusedOnEncode(t *testing.T) {
	m, err := decodeXML(withCustomTypes(t))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, bad := range []string{
		"<terraintype type=\"x\">",          // unclosed
		"<terraintype type=x />",            // unquoted attribute
		"</terrain-config><injected/>",      // closes the wrapper
		"<terraintype>&nbsp;</terraintype>", // unknown entity
	} {
		m.Configuration.TerrainConfig[0].InnerXML = bad
		if _, err := xmlio.MarshalXML(m, "2.08"); !errors.Is(err, wxx.ErrInvalidXML) {
			t.Errorf("encode with %q = %v, want ErrInvalidXML", bad, err)
		}
	}
}
