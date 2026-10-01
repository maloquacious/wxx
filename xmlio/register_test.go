// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// w2025XMLHeader is the declaration every W2025 fixture opens with, quoting
// included, as the decoder reads it back from disk.
const w2025XMLHeader = "<?xml version='1.1' encoding='utf-16'?>\n"

// registeredFixtureSets is every registered W2025 version after the 2.06
// baseline, with every tracked save it wrote and the issue that registered it.
var registeredFixtureSets = []struct {
	app      string
	issue    string
	fixtures []string
}{
	{"2.07", "#92", fixtures207},
	{"2.08", "#73", fixtures208},
}

// TestRegisteredFixtureListsAreEveryTrackedFixture holds fixtures207 and
// fixtures208 to the testdata directory, so a save added later is not silently
// left out of the tests that loop over the lists.
func TestRegisteredFixtureListsAreEveryTrackedFixture(t *testing.T) {
	for _, set := range registeredFixtureSets {
		t.Run(set.app, func(t *testing.T) {
			onDisk, err := filepath.Glob(filepath.Join("..", "testdata", "2025-"+set.app+"-*.wxx"))
			if err != nil {
				t.Fatalf("glob %s fixtures: %v", set.app, err)
			}
			var listed []string
			for _, p := range set.fixtures {
				listed = append(listed, filepath.Base(p))
			}
			// Worldographer leaves a *-autosave.wxx beside the map it saves,
			// even after a clean exit; it is git-ignored and never a fixture.
			onDisk = slices.DeleteFunc(onDisk, func(p string) bool {
				return strings.HasSuffix(p, "-autosave.wxx")
			})
			for i := range onDisk {
				onDisk[i] = filepath.Base(onDisk[i])
			}
			slices.Sort(listed)
			slices.Sort(onDisk)
			if !slices.Equal(listed, onDisk) {
				t.Errorf("the %s list has %v, testdata has %v", set.app, listed, onDisk)
			}
			// Vacuity: an empty list matches an empty glob.
			if len(onDisk) == 0 {
				t.Errorf("testdata has no %s fixtures", set.app)
			}
		})
	}
}

// TestEncodeRegisteredFixturesAsThemselves is the proof for issues #92 and #73.
// Every 2.07 and 2.08 fixture encodes as the version it states through the full
// public pipeline, and the file it writes states exactly release="2025",
// that version, and schema="1.06", under the XML 1.1 declaration the source
// opens with.
//
// It also pins that each version is data and not a format: the XML written as
// "2.07" or "2.08" is the XML written as "2.06" with the one version attribute
// changed. No difference between what the builds write is known, so none may
// appear here (ruling on #92, kept for #73: report only proven losses).
func TestEncodeRegisteredFixturesAsThemselves(t *testing.T) {
	for _, set := range registeredFixtureSets {
		t.Run(set.app, func(t *testing.T) {
			checked := 0
			for _, path := range set.fixtures {
				t.Run(filepath.Base(path), func(t *testing.T) {
					encodeAsItself(t, path, set.app)
					checked++
				})
			}
			if checked != len(set.fixtures) || checked == 0 {
				t.Errorf("checked %d of %d %s fixture(s)", checked, len(set.fixtures), set.app)
			}
		})
	}
}

// encodeAsItself is one case of TestEncodeRegisteredFixturesAsThemselves.
func encodeAsItself(t *testing.T, path, app string) {
	t.Helper()
	want := map[string]string{"release": "2025", "version": app, "schema": "1.06"}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	var src xmlio.DecoderDiagnostics
	m, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&src)).Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	// Guard: the source states the identity asserted below, so the test is
	// about writing the version as itself.
	if got := string(src.XMLHeader); got != w2025XMLHeader {
		t.Fatalf("%s: source opens with %q, want %q", path, got, w2025XMLHeader)
	}
	if got := mapIdentity(t, path+" source", src.Converted); !maps.Equal(got, want) {
		t.Fatalf("%s: source states %v, want %v", path, got, want)
	}

	var ed xmlio.EncoderDiagnostics
	var buf bytes.Buffer
	if err := xmlio.NewEncoder(app, xmlio.WithEncoderDiagnostics(&ed)).Encode(&buf, m); err != nil {
		t.Fatalf("encode %s as %q: %v", path, app, err)
	}
	if len(ed.Dropped) != 0 {
		t.Errorf("%s: encoding as %q reported a loss: %v", path, app, ed.Dropped)
	}

	// Read the written file back to its text, declaration included.
	var out xmlio.DecoderDiagnostics
	back, err := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&out)).Decode(&buf)
	if err != nil {
		t.Fatalf("re-decode %s written as %q: %v", path, app, err)
	}
	if got := string(out.XMLHeader); got != w2025XMLHeader {
		t.Errorf("%s written as %q opens with %q, want %q", path, app, got, w2025XMLHeader)
	}
	if got := mapIdentity(t, path+" output", out.Converted); !maps.Equal(got, want) {
		t.Errorf("%s written as %q states %v, want exactly %v", path, app, got, want)
	}
	if v := back.MetaData.Version; v.App.Raw != app || v.Schema == nil || v.Schema.Raw != "1.06" {
		t.Errorf("%s written as %q re-decodes as %v", path, app, v)
	}

	// The version and 2.06 differ in the version attribute and nothing else.
	xApp, err := xmlio.MarshalXML(m, app)
	if err != nil {
		t.Fatalf("MarshalXML(%s, %q): %v", path, app, err)
	}
	x06, err := xmlio.MarshalXML(m, "2.06")
	if err != nil {
		t.Fatalf("MarshalXML(%s, %q): %v", path, "2.06", err)
	}
	if !bytes.Equal(xApp, ed.Utf8Encoded) {
		t.Errorf("%s: MarshalXML and Encode disagree on the XML for %q", path, app)
	}
	if bytes.Equal(x06, xApp) {
		t.Fatalf("%s: the XML written as 2.06 and as %s is identical, so the version attribute was not written from the target", path, app)
	}
	swapped := bytes.Replace(x06, []byte(` version="2.06" `), []byte(` version="`+app+`" `), 1)
	if !bytes.Equal(swapped, xApp) {
		i := 0
		for i < len(swapped) && i < len(xApp) && swapped[i] == xApp[i] {
			i++
		}
		t.Errorf("%s: the XML written as %s differs from the 2.06 XML beyond map/@version, at byte %d:\n2.06: %q\n%s: %q",
			path, app, i, window(swapped, i), app, window(xApp, i))
	}
}

// mapIdentity returns the identity attributes the document's <map> start tag
// states. An attribute the tag does not state is absent from the result.
func mapIdentity(t *testing.T, label string, doc []byte) map[string]string {
	t.Helper()
	tags := startTagAttrs(doc, "map")
	if len(tags) != 1 {
		t.Fatalf("%s: %d <map> start tags, want 1", label, len(tags))
	}
	got := map[string]string{}
	for _, a := range tags[0] {
		switch a[0] {
		case "release", "version", "schema":
			if _, dup := got[a[0]]; dup {
				t.Fatalf("%s: <map> states @%s twice", label, a[0])
			}
			got[a[0]] = a[1]
		}
	}
	return got
}
