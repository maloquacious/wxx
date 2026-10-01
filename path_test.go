// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import "testing"

// TestRGBAAttribute: a path's stroke color is spelled as Java's
// Double.toString spells each component, as the app writes it and the W2025
// codec's floats does.
func TestRGBAAttribute(t *testing.T) {
	for _, tc := range []struct {
		c    RGBA_t
		want string
	}{
		{RGBA_t{0, 0, 1, 1}, "0.0,0.0,1.0,1.0"},
		{RGBA_t{1, 0.8941176533699036, 0.7686274647712708, 1}, "1.0,0.8941176533699036,0.7686274647712708,1.0"},
		{RGBA_t{0.001, 0.0001, 0.5, 0.25}, "0.001,1.0E-4,0.5,0.25"},
		{RGBA_t{1.5e-5, 0, 0, 1}, "1.5E-5,0.0,0.0,1.0"},
	} {
		if got := rgbaAttribute(tc.c); got != tc.want {
			t.Errorf("rgbaAttribute(%+v) = %q, want %q", tc.c, got, tc.want)
		}
	}
}
