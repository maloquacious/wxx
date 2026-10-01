// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// This is an INTERNAL test (package xmlio, not xmlio_test): newestApp and the
// registry it reads are unexported.
package xmlio

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/maloquacious/wxx"
)

// TestNewestApp holds CurrentApp's choice to the rules init relies on: order by
// components, not by string, and refuse rather than pick when there is no single
// newest version.
func TestNewestApp(t *testing.T) {
	for _, tc := range []struct {
		name string
		apps []string
		want string
		err  error
	}{
		{"one", []string{"2.06"}, "2.06", nil},
		{"registry order", []string{"2.06", "2.07", "2.08"}, "2.08", nil},
		{"any order", []string{"2.08", "2.06", "2.07"}, "2.08", nil},
		{"by components", []string{"2.9", "2.10"}, "2.10", nil},
		{"major first", []string{"2.99", "3.00"}, "3.00", nil},
		{"tie", []string{"2.6", "2.06"}, "", wxx.ErrAmbiguousAppCodec},
		{"tie below newest", []string{"2.6", "2.06", "2.07"}, "2.07", nil},
		{"unparsable", []string{"2.06", "beta"}, "", wxx.ErrInvalidDottedVersion},
		{"none", nil, "", wxx.ErrUnsupportedMapVersion},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newestApp(tc.apps)
			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("newestApp(%q) error = %v, want %v", tc.apps, err, tc.err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("newestApp(%q) = %q, %v, want %q", tc.apps, got, err, tc.want)
			}
		})
	}
}

// TestEveryRegisteredAppHasABlankFixture: NewMap's defaults are checked against
// the blank fixture of each version (newmap_test.go), so a version registered
// without one would have defaults nothing checks.
func TestEveryRegisteredAppHasABlankFixture(t *testing.T) {
	for app := range byApp {
		path := fmt.Sprintf("../testdata/2025-%s-13x11-941577-blank.wxx", app)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("version %q is registered but has no blank fixture: %v", app, err)
		}
	}
}
