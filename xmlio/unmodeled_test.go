// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// unmodeledFixture is the fixture the injection tests edit: a 2.08 save with
// labels, label styles, features, notes and three-deep lore.
const unmodeledFixture = "2025-2.08-13x11-941577-populated.wxx"

// TestDecodeRefusesUnmodeledContent: decode refuses an attribute, element or
// text that the schema does not model, wherever it appears, and names where
// it is (issue #161). Before #161 each of these decoded and was dropped on the
// next write without a word.
//
// The first three are the attributes Worldographer 2.09 writes (from the
// application's string constants; no 2.09 fixture exists yet). Until #161
// models them, a 2.09 map that states them is refused rather than losing them.
func TestDecodeRefusesUnmodeledContent(t *testing.T) {
	src, _ := readNotesFixture(t, unmodeledFixture)
	src = bytes.TrimLeft(stripXMLDecl(src), "\n")

	for _, tc := range []struct {
		name     string
		from, to string
		want     []string // substrings the error must contain
	}{
		{"2.09 label lineSpacing",
			`<label  mapLayer="Features"`, `<label  lineSpacing="1.0" mapLayer="Features"`,
			[]string{"/label[", "@lineSpacing"}},
		{"2.09 label textAlignment",
			`<label  mapLayer="Features"`, `<label  textAlignment="CENTER" mapLayer="Features"`,
			[]string{"/label[", "@textAlignment"}},
		{"2.09 labelstyle lineSpacing",
			`<labelstyle name="Nation"`, `<labelstyle name="Nation" lineSpacing="1.0"`,
			[]string{"map/configuration[1]/text-config[1]/labelstyle[1]", "@lineSpacing"}},
		{"attribute on map",
			`<map `, `<map futureAttr="x" `,
			[]string{"map: unmodeled attribute @futureAttr"}},
		{"namespaced attribute on map",
			`<map `, `<map xml:lang="en" `,
			[]string{"map: unmodeled attribute @", `lang="en"`}}, // the prefix is reported as its namespace URL
		{"attribute on feature",
			`isHideTerrainIcon="false">`, `isHideTerrainIcon="false" futureAttr="x">`,
			[]string{"map/features[1]/feature[1]", "@futureAttr"}},
		{"attribute on three-deep lore",
			`uuid="b89a1c33-bdee-4f34-bdba-946d8c191907"`, `uuid="b89a1c33-bdee-4f34-bdba-946d8c191907" futureAttr="x"`,
			[]string{"map/informations[1]/information[", "]/information[", "@futureAttr"}},
		{"attribute on a text-only element",
			`<notetext>`, `<notetext future="1">`,
			[]string{"/notetext[1]", "@future"}},
		{"top-level element",
			`<terrainmap>`, `<future/><terrainmap>`,
			[]string{"map/future[1]: unmodeled element <future>"}},
		{"element in tiles",
			`<tiles viewLevel="WORLD" tilesWide="13" tilesHigh="11">`, `<tiles viewLevel="WORLD" tilesWide="13" tilesHigh="11"><future/>`,
			[]string{"map/tiles[1]/future[1]: unmodeled element <future>"}},
		{"text in labels",
			`<labels>`, `<labels>stray`,
			[]string{"map/labels[1]: unmodeled text \"stray\""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if n := bytes.Count(src, []byte(tc.from)); n == 0 {
				t.Fatalf("fixture does not state %s", tc.from)
			}
			doc := bytes.Replace(src, []byte(tc.from), []byte(tc.to), 1)
			_, err := decodeRawXML(t, doc, "1.1")
			if !errors.Is(err, wxx.ErrUnmodeledContent) {
				t.Fatalf("decode: err = %v, want errors.Is(err, %v)", err, wxx.ErrUnmodeledContent)
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("decode: err = %q, want it to contain %q", err, w)
				}
			}
		})
	}

	// Controls: the unmodified document decodes, so each refusal above is the
	// edit's; and whitespace between elements is layout, not content.
	if _, err := decodeRawXML(t, src, "1.1"); err != nil {
		t.Fatalf("decode unmodified: %v", err)
	}
	spaced := bytes.Replace(src, []byte(`<labels>`), []byte("<labels>\n\n\t  "), 1)
	if _, err := decodeRawXML(t, spaced, "1.1"); err != nil {
		t.Fatalf("decode with added whitespace: %v", err)
	}
}

// TestDecodeRefusesUnmodeledExtraTerrain: <extraTerrain> refused unmodeled
// content before #161 with its own check (issue #34). That check is gone, so
// this pins that the document-wide one covers each of its levels.
func TestDecodeRefusesUnmodeledExtraTerrain(t *testing.T) {
	src, _ := readNotesFixture(t, "2025-2.08-13x11-941577-extraterrain-bgcolor.wxx")
	src = bytes.TrimLeft(stripXMLDecl(src), "\n")
	for _, tc := range []struct {
		name, from, to, want string
	}{
		{"element in extraTerrain", `<extraTerrain>`, `<extraTerrain><future/>`,
			"map/extraTerrain[1]/future[1]: unmodeled element <future>"},
		{"attribute on mapLayer", `<mapLayer name="Below All"`, `<mapLayer name="Below All" futureAttr="x"`,
			"map/extraTerrain[1]/mapLayer[1]: unmodeled attribute @futureAttr"},
		{"attribute on terrainAndLocation", `<terrainAndLocation name="Classic/Flat Desert Cold"`, `<terrainAndLocation name="Classic/Flat Desert Cold" futureAttr="x"`,
			"map/extraTerrain[1]/mapLayer[1]/terrainAndLocation[1]: unmodeled attribute @futureAttr"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !bytes.Contains(src, []byte(tc.from)) {
				t.Fatalf("fixture does not state %s", tc.from)
			}
			_, err := decodeRawXML(t, bytes.Replace(src, []byte(tc.from), []byte(tc.to), 1), "1.1")
			if !errors.Is(err, wxx.ErrUnmodeledContent) {
				t.Fatalf("decode: err = %v, want errors.Is(err, %v)", err, wxx.ErrUnmodeledContent)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("decode: err = %q, want it to contain %q", err, tc.want)
			}
		})
	}
	if _, err := decodeRawXML(t, src, "1.1"); err != nil {
		t.Fatalf("decode unmodified: %v", err)
	}
}

// TestEveryW2025FixtureHasOnlyModeledContent: every tracked W2025 fixture
// decodes under the #161 refusal, so nothing any 2.06, 2.07 or 2.08 save states
// is outside the schema.
func TestEveryW2025FixtureHasOnlyModeledContent(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "testdata", "2025-*.wxx"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 20 {
		t.Fatalf("found %d W2025 fixtures, want at least 20; is testdata/ present?", len(paths))
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if _, err := xmlio.NewDecoder().Decode(f); err != nil {
				t.Fatalf("decode: %v", err)
			}
		})
	}
}
