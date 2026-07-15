// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"testing"
)

// TestParseDotted checks that a well-formed dotted version parses into the
// expected components while Raw keeps the input bytes untouched.
func TestParseDotted(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		major int
		minor int
	}{
		{name: "classic 1.73", input: "1.73", major: 1, minor: 73},
		{name: "classic 1.77", input: "1.77", major: 1, minor: 77},
		{name: "zero-padded schema 1.06", input: "1.06", major: 1, minor: 6},
		{name: "zero-padded app 2.06", input: "2.06", major: 2, minor: 6},
		{name: "unpadded 2.6", input: "2.6", major: 2, minor: 6},
		{name: "zero major", input: "0.9", major: 0, minor: 9},
		{name: "all zeroes", input: "0.0", major: 0, minor: 0},
		{name: "zero-padded major", input: "02.06", major: 2, minor: 6},
		{name: "multi-digit major", input: "10.06", major: 10, minor: 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDotted(tc.input)
			if err != nil {
				t.Fatalf("ParseDotted(%q): unexpected error: %v", tc.input, err)
			}
			if got.Raw != tc.input {
				t.Errorf("ParseDotted(%q).Raw = %q, want %q (Raw must be verbatim)", tc.input, got.Raw, tc.input)
			}
			if got.Major != tc.major {
				t.Errorf("ParseDotted(%q).Major = %d, want %d", tc.input, got.Major, tc.major)
			}
			if got.Minor != tc.minor {
				t.Errorf("ParseDotted(%q).Minor = %d, want %d", tc.input, got.Minor, tc.minor)
			}
		})
	}
}

// TestDottedStringIsVerbatim is the centerpiece: the reason Dotted exists.
//
// These values are not semver. "2.06" through a semver round-trip comes back as
// "2.6" -- a different string, and therefore a different file on disk. String()
// must hand back the bytes we were given and must never re-render them from
// Major and Minor.
//
// The zero-padded cases are the only ones that can catch a re-render: "1.73"
// rendered from {1,73} is still "1.73", so it passes even against a broken
// String(). Keep 1.06 and 2.06 -- without them this test cannot fail.
func TestDottedStringIsVerbatim(t *testing.T) {
	for _, input := range []string{
		"1.06", // load-bearing: padded
		"1.73",
		"1.77",
		"2.06", // load-bearing: padded
	} {
		t.Run(input, func(t *testing.T) {
			d, err := ParseDotted(input)
			if err != nil {
				t.Fatalf("ParseDotted(%q): unexpected error: %v", input, err)
			}
			if got := d.String(); got != input {
				t.Errorf("ParseDotted(%q).String() = %q, want %q: the dotted version was re-rendered from its components instead of Raw; these values are not semver and the zero padding is part of the file", input, got, input)
			}
			// and again through a second parse, to pin that a Dotted that has
			// been round-tripped is still byte-identical to what was read.
			again, err := ParseDotted(d.String())
			if err != nil {
				t.Fatalf("ParseDotted(%q) (round trip): unexpected error: %v", d.String(), err)
			}
			if again.Raw != input {
				t.Errorf("round trip of %q produced Raw %q, want %q", input, again.Raw, input)
			}
		})
	}
}

// TestDottedComparePadding pins the distinction the type exists to preserve:
// "2.06" and "2.6" are the same ordinal but not the same string. They must
// compare equal, and they must still write back differently.
func TestDottedComparePadding(t *testing.T) {
	padded, err := ParseDotted("2.06")
	if err != nil {
		t.Fatalf("ParseDotted(\"2.06\"): unexpected error: %v", err)
	}
	unpadded, err := ParseDotted("2.6")
	if err != nil {
		t.Fatalf("ParseDotted(\"2.6\"): unexpected error: %v", err)
	}

	// same components ...
	if padded.Major != unpadded.Major || padded.Minor != unpadded.Minor {
		t.Errorf("components differ: %q = {%d,%d}, %q = {%d,%d}, want identical",
			padded.Raw, padded.Major, padded.Minor, unpadded.Raw, unpadded.Major, unpadded.Minor)
	}
	if got := padded.Compare(unpadded); got != 0 {
		t.Errorf("ParseDotted(%q).Compare(%q) = %d, want 0: Compare must use the components only", padded.Raw, unpadded.Raw, got)
	}
	if padded.Less(unpadded) || unpadded.Less(padded) {
		t.Errorf("%q and %q must not order either way", padded.Raw, unpadded.Raw)
	}

	// ... but they are not interchangeable on disk.
	if padded.Raw == unpadded.Raw {
		t.Errorf("Raw of %q and %q are equal; the padding was lost", "2.06", "2.6")
	}
	if padded.String() != "2.06" {
		t.Errorf("padded.String() = %q, want %q", padded.String(), "2.06")
	}
	if unpadded.String() != "2.6" {
		t.Errorf("unpadded.String() = %q, want %q", unpadded.String(), "2.6")
	}
}

// TestDottedCompare checks ordering on Major then Minor.
func TestDottedCompare(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "equal", a: "2.06", b: "2.06", want: 0},
		{name: "equal across padding", a: "2.06", b: "2.6", want: 0},
		{name: "minor less", a: "1.73", b: "1.77", want: -1},
		{name: "minor greater", a: "1.77", b: "1.73", want: 1},
		{name: "major less", a: "1.06", b: "2.06", want: -1},
		{name: "major greater", a: "2.06", b: "1.06", want: 1},
		{name: "major wins over minor", a: "2.00", b: "1.99", want: 1},
		{name: "minor is an ordinal not a fraction", a: "1.9", b: "1.73", want: -1},
		{name: "padded minor is an ordinal", a: "1.06", b: "1.6", want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, err := ParseDotted(tc.a)
			if err != nil {
				t.Fatalf("ParseDotted(%q): unexpected error: %v", tc.a, err)
			}
			b, err := ParseDotted(tc.b)
			if err != nil {
				t.Fatalf("ParseDotted(%q): unexpected error: %v", tc.b, err)
			}
			if got := a.Compare(b); got != tc.want {
				t.Errorf("ParseDotted(%q).Compare(%q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
			if got, want := a.Less(b), tc.want < 0; got != want {
				t.Errorf("ParseDotted(%q).Less(%q) = %v, want %v", tc.a, tc.b, got, want)
			}
		})
	}
}

// TestParseDottedErrors checks that malformed input is rejected rather than
// coerced into something plausible.
func TestParseDottedErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  error // sentinel the returned error must wrap
	}{
		{name: "empty", input: "", want: ErrMissingVersion},
		{name: "no separator", input: "206", want: ErrInvalidDottedComponentCount},
		{name: "three components", input: "2.0.6", want: ErrInvalidDottedComponentCount},
		{name: "four components", input: "1.2.3.4", want: ErrInvalidDottedComponentCount},
		{name: "trailing dot", input: "2.", want: ErrInvalidDottedComponent},
		{name: "leading dot", input: ".06", want: ErrInvalidDottedComponent},
		{name: "only a dot", input: ".", want: ErrInvalidDottedComponent},
		{name: "non-numeric major", input: "v2.06", want: ErrInvalidDottedComponent},
		{name: "non-numeric minor", input: "2.06a", want: ErrInvalidDottedComponent},
		{name: "alphabetic", input: "two.six", want: ErrInvalidDottedComponent},
		{name: "negative major", input: "-2.06", want: ErrInvalidDottedComponent},
		{name: "negative minor", input: "2.-06", want: ErrInvalidDottedComponent},
		{name: "plus sign", input: "+2.06", want: ErrInvalidDottedComponent},
		{name: "leading space", input: " 2.06", want: ErrInvalidDottedComponent},
		{name: "trailing space", input: "2.06 ", want: ErrInvalidDottedComponent},
		{name: "internal space", input: "2. 06", want: ErrInvalidDottedComponent},
		{name: "semver-ish prerelease", input: "2.06-alpha", want: ErrInvalidDottedComponent},
		{name: "comma separator", input: "2,06", want: ErrInvalidDottedComponentCount},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDotted(tc.input)
			if err == nil {
				t.Fatalf("ParseDotted(%q) = %#v, want an error", tc.input, got)
			}
			if !errors.Is(err, ErrInvalidDottedVersion) {
				t.Errorf("ParseDotted(%q) error = %v, want it to wrap %v", tc.input, err, ErrInvalidDottedVersion)
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("ParseDotted(%q) error = %v, want it to wrap %v", tc.input, err, tc.want)
			}
			if got != (Dotted{}) {
				t.Errorf("ParseDotted(%q) = %#v, want the zero Dotted on error", tc.input, got)
			}
		})
	}
}
