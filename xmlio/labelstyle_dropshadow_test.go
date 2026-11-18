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

// Issue #166: <labelstyle> dropShadowRadius and dropShadowSpread are not integer
// attributes. Worldographer writes the constant `dropShadowColor="null"
// dropShadowRadius="0" dropShadowSpread="0"` for a style with no shadow, and
// Java doubles for a style with one. The values below are what 2.09 wrote for a
// City style edited in Configure Label Styles (a first save). Until #166 decode
// refused that spelling with ErrInvalidIntegerAttribute.
const (
	noShadow       = `dropShadowColor="null" dropShadowRadius="0" dropShadowSpread="0"`
	appShadow      = `dropShadowColor="0.6000000238418579,0.7019608020782471,1.0,1.0" dropShadowRadius="75.0" dropShadowSpread="1.0"`
	nullShadowDec  = `dropShadowColor="null" dropShadowRadius="0.0" dropShadowSpread="0.0"`
	fractionShadow = `dropShadowColor="0.6000000238418579,0.7019608020782471,1.0,1.0" dropShadowRadius="4.5" dropShadowSpread="0.25"`
)

// blankWithCityShadow returns the 2.08 blank encoded as 2.08, with the City
// style's drop-shadow attributes replaced by shadow.
func blankWithCityShadow(t *testing.T, shadow string) []byte {
	t.Helper()
	m, err := xmlio.ReadFile(filepath.Join("..", "testdata", "2025-2.08-13x11-941577-blank.wxx"))
	if err != nil {
		t.Fatalf("decode blank: %v", err)
	}
	doc, err := xmlio.MarshalXML(m, "2.08")
	if err != nil {
		t.Fatalf("encode blank: %v", err)
	}
	start := bytes.Index(doc, []byte(`<labelstyle name="City"`))
	if start < 0 {
		t.Fatalf("encoded blank has no City label style")
	}
	end := start + bytes.IndexByte(doc[start:], '\n')
	line := doc[start:end]
	if !bytes.Contains(line, []byte(noShadow)) {
		t.Fatalf("City style does not state %s: %s", noShadow, line)
	}
	edited := bytes.Replace(line, []byte(noShadow), []byte(shadow), 1)
	return append(append(append([]byte{}, doc[:start]...), edited...), doc[end:]...)
}

func cityStyle(t *testing.T, m *wxx.Map_t) *wxx.LabelStyle_t {
	t.Helper()
	for _, ls := range m.Configuration.TextConfig.LabelStyles {
		if ls.Name == "City" {
			return ls
		}
	}
	t.Fatalf("no City label style")
	return nil
}

// TestW2025LabelStyleDropShadowSpelling: each spelling decodes to its value, and
// encodes as the app writes it.
func TestW2025LabelStyleDropShadowSpelling(t *testing.T) {
	for _, tc := range []struct {
		name           string
		in, out        string
		radius, spread float64
	}{
		// The bug: the app's own spelling for a style with a shadow.
		{"shadow set, as the app writes it", appShadow, appShadow, 75, 1},
		// App check (#166): 2.08 opens this and writes it back as "0".
		{"no shadow, decimal spelling", nullShadowDec, noShadow, 0, 0},
		{"no shadow, as the app writes it", noShadow, noShadow, 0, 0},
		// No fixture; the app's spinners are doubles, so a fraction is possible
		// and must not be refused or rounded.
		{"shadow set, fractional", fractionShadow, fractionShadow, 4.5, 0.25},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := blankWithCityShadow(t, tc.in)
			m, err := decodeXML(src)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			city := cityStyle(t, m)
			if city.DropShadowRadius != tc.radius || city.DropShadowSpread != tc.spread {
				t.Errorf("City radius, spread = %v, %v; want %v, %v", city.DropShadowRadius, city.DropShadowSpread, tc.radius, tc.spread)
			}
			for _, app := range []string{"2.06", "2.07", "2.08"} {
				out, err := xmlio.MarshalXML(m, app)
				if err != nil {
					t.Fatalf("encode as %s: %v", app, err)
				}
				want := bytes.Replace(blankWithCityShadow(t, tc.out), []byte(`version="2.08"`), []byte(`version="`+app+`"`), 1)
				if !bytes.Equal(out, want) {
					t.Errorf("encode as %s: output differs from the source with the app's spelling", app)
				}
			}
		})
	}
}

// TestW2025LabelStyleNullShadowNonzeroKept: a nonzero radius under a "null"
// colour never comes from the app, but a caller can set one. It is written as a
// double, not dropped and not refused.
func TestW2025LabelStyleNullShadowNonzeroKept(t *testing.T) {
	m, err := decodeXML(blankWithCityShadow(t, noShadow))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	cityStyle(t, m).DropShadowRadius = 1.5
	out, err := xmlio.MarshalXML(m, "2.08")
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if want := `dropShadowColor="null" dropShadowRadius="1.5" dropShadowSpread="0"`; !strings.Contains(string(out), want) {
		t.Errorf("encoded City style does not state %s", want)
	}
}

// TestW2025LabelStyleDropShadowNotANumberRefused: accepting decimals does not
// mean accepting anything.
func TestW2025LabelStyleDropShadowNotANumberRefused(t *testing.T) {
	_, err := decodeXML(blankWithCityShadow(t, `dropShadowColor="null" dropShadowRadius="wide" dropShadowSpread="0"`))
	if err == nil {
		t.Fatalf("decode: want an error for dropShadowRadius=\"wide\", got nil")
	}
	if errors.Is(err, wxx.ErrInvalidIntegerAttribute) {
		t.Errorf("decode: %v; this attribute is not an integer attribute any more", err)
	}
}
