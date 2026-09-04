// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"errors"
	"fmt"
	"strings"
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
	got, err := padded.Compare(unpadded)
	if err != nil {
		t.Fatalf("%q.Compare(%q): unexpected error: %v", padded.Raw, unpadded.Raw, err)
	}
	if got != 0 {
		t.Errorf("ParseDotted(%q).Compare(%q) = %d, want 0: Compare must use the components only", padded.Raw, unpadded.Raw, got)
	}
	forward, err := padded.Less(unpadded)
	if err != nil {
		t.Fatalf("%q.Less(%q): unexpected error: %v", padded.Raw, unpadded.Raw, err)
	}
	backward, err := unpadded.Less(padded)
	if err != nil {
		t.Fatalf("%q.Less(%q): unexpected error: %v", unpadded.Raw, padded.Raw, err)
	}
	if forward || backward {
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
			got, err := a.Compare(b)
			if err != nil {
				t.Fatalf("ParseDotted(%q).Compare(%q): unexpected error: %v", tc.a, tc.b, err)
			}
			if got != tc.want {
				t.Errorf("ParseDotted(%q).Compare(%q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
			less, err := a.Less(b)
			if err != nil {
				t.Fatalf("ParseDotted(%q).Less(%q): unexpected error: %v", tc.a, tc.b, err)
			}
			if want := tc.want < 0; less != want {
				t.Errorf("ParseDotted(%q).Less(%q) = %v, want %v", tc.a, tc.b, less, want)
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

// TestParsedReportsWhetherComponentsExist pins the flag that issue #38 turns on:
// only ParseDotted produces a Dotted whose components mean anything.
//
// The zero Dotted and a decoder-shaped literal are both unparsed, and they are
// unparsed for the same reason -- neither went through ParseDotted. That is the
// whole mechanism: there is no other way to set the flag, so there is no way to
// hold a Dotted that claims components it does not have.
func TestParsedReportsWhetherComponentsExist(t *testing.T) {
	parsed, err := ParseDotted("2.06")
	if err != nil {
		t.Fatalf("ParseDotted(\"2.06\"): unexpected error: %v", err)
	}
	if !parsed.Parsed() {
		t.Errorf("ParseDotted(\"2.06\").Parsed() = false, want true: a parsed version has components")
	}

	// The zero value, and what ParseDotted returns when it fails.
	if (Dotted{}).Parsed() {
		t.Errorf("Dotted{}.Parsed() = true, want false: the zero Dotted has no components")
	}
	failed, err := ParseDotted("garbage")
	if err == nil {
		t.Fatalf("ParseDotted(\"garbage\") = %#v, want an error", failed)
	}
	if failed.Parsed() {
		t.Errorf("the Dotted returned with an error reports Parsed() = true, want false")
	}

	// The decoder fallback's shape. This literal is what
	// v1_06.dottedOrRaw and v0_77.classicVersionIdentity build; written here
	// inside package wxx it could name the flag, and the point is that outside
	// the package it cannot.
	fallback := Dotted{Raw: "garbage"}
	if fallback.Parsed() {
		t.Errorf("Dotted{Raw: %q}.Parsed() = true, want false: the decoder fallback never parsed anything", fallback.Raw)
	}
	if fallback.String() != "garbage" {
		t.Errorf("fallback.String() = %q, want %q: Raw is trustworthy even when the components are not", fallback.String(), "garbage")
	}
}

// TestUnparsedDottedCannotBeOrdered is issue #38.
//
// The bug: a malformed on-disk version decoded to {Raw: "garbage", Major: 0,
// Minor: 0}, and Compare read those zeroes as an answer -- reporting the file
// EQUAL to "0.0" and LESS than every real version. Raw was right; the
// components were a lie, and Compare had no way to say so.
//
// So the assertion is not "Compare returns something sensible" but "Compare
// refuses". Each case below is one the old int-returning Compare answered
// confidently and wrongly.
func TestUnparsedDottedCannotBeOrdered(t *testing.T) {
	real206 := mustParseDotted(t, "2.06")
	zeroes := mustParseDotted(t, "0.0")
	malformed := Dotted{Raw: "garbage"} // the decoder fallback

	for _, tc := range []struct {
		name string
		a, b Dotted
	}{
		// The headline: "garbage" compared equal to "0.0".
		{name: "malformed against 0.0", a: malformed, b: zeroes},
		{name: "0.0 against malformed", a: zeroes, b: malformed},
		// And less than every real version, in either position.
		{name: "malformed against a real version", a: malformed, b: real206},
		{name: "a real version against malformed", a: real206, b: malformed},
		// The zero Dotted is unparsed for the same reason.
		{name: "zero Dotted against a real version", a: Dotted{}, b: real206},
		// Two unparsed values compared equal to each other, too.
		{name: "malformed against itself", a: malformed, b: malformed},
		{name: "two different malformed versions", a: malformed, b: Dotted{Raw: "also garbage"}},
		// A hand-built literal carrying components it never parsed is unparsed
		// as well: the components are unverified bytes, not a parse result.
		{name: "hand-built components", a: Dotted{Raw: "2.06", Major: 2, Minor: 6}, b: real206},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.a.Compare(tc.b)
			if err == nil {
				t.Fatalf("%#v.Compare(%#v) = %d, nil; want an error: at least one side has no components, so there is no ordering to report", tc.a, tc.b, got)
			}
			if !errors.Is(err, ErrUnparsedDottedVersion) {
				t.Errorf("Compare error = %v, want it to wrap %v", err, ErrUnparsedDottedVersion)
			}
			if got != 0 {
				t.Errorf("Compare returned %d alongside its error, want 0: the int is not an ordering", got)
			}

			less, err := tc.a.Less(tc.b)
			if err == nil {
				t.Fatalf("%#v.Less(%#v) = %v, nil; want an error", tc.a, tc.b, less)
			}
			if !errors.Is(err, ErrUnparsedDottedVersion) {
				t.Errorf("Less error = %v, want it to wrap %v", err, ErrUnparsedDottedVersion)
			}
			if less {
				t.Errorf("Less returned true alongside its error, want false: the bool is not an ordering")
			}
		})
	}
}

// TestUnparsedCompareErrorNamesTheBadVersion checks that the error says WHICH
// version string could not be ordered. A caller holding two versions needs to
// know which of them came out of a file it could not read.
func TestUnparsedCompareErrorNamesTheBadVersion(t *testing.T) {
	real206 := mustParseDotted(t, "2.06")
	malformed := Dotted{Raw: "not-a-version"}

	for _, tc := range []struct {
		name string
		a, b Dotted
	}{
		{name: "bad on the left", a: malformed, b: real206},
		{name: "bad on the right", a: real206, b: malformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.a.Compare(tc.b)
			if err == nil {
				t.Fatalf("Compare: want an error")
			}
			if !strings.Contains(err.Error(), malformed.Raw) {
				t.Errorf("Compare error = %v, want it to name the unparsed version %q", err, malformed.Raw)
			}
		})
	}
}

// mustParseDotted parses s or fails the test.
func mustParseDotted(t *testing.T, s string) Dotted {
	t.Helper()
	d, err := ParseDotted(s)
	if err != nil {
		t.Fatalf("ParseDotted(%q): unexpected error: %v", s, err)
	}
	return d
}

// TestNewDottedMatchesParseDotted is the constructor's central invariant: the
// two constructors are interchangeable when they describe the same version.
//
// Not "equivalent" loosely -- identical under ==, unexported flag included. If
// NewDotted ever forgot to set the flag, or set a component differently from the
// parser, this catches it without needing to name the flag.
func TestNewDottedMatchesParseDotted(t *testing.T) {
	for _, tc := range []struct {
		name         string
		major, minor int
	}{
		{name: "classic 1.73", major: 1, minor: 73},
		{name: "classic 1.77", major: 1, minor: 77},
		{name: "unpadded 2.6", major: 2, minor: 6},
		{name: "all zeroes", major: 0, minor: 0},
		{name: "zero major", major: 0, minor: 9},
		{name: "multi-digit major", major: 10, minor: 6},
		{name: "large components", major: 2026, minor: 12345},
	} {
		t.Run(tc.name, func(t *testing.T) {
			built, err := NewDotted(tc.major, tc.minor)
			if err != nil {
				t.Fatalf("NewDotted(%d, %d): unexpected error: %v", tc.major, tc.minor, err)
			}
			rendered := fmt.Sprintf("%d.%d", tc.major, tc.minor)
			if built.Raw != rendered {
				t.Errorf("NewDotted(%d, %d).Raw = %q, want %q", tc.major, tc.minor, built.Raw, rendered)
			}

			parsed, err := ParseDotted(rendered)
			if err != nil {
				t.Fatalf("ParseDotted(%q): unexpected error: %v", rendered, err)
			}
			if built != parsed {
				t.Errorf("NewDotted(%d, %d) = %#v, want it identical to ParseDotted(%q) = %#v",
					tc.major, tc.minor, built, rendered, parsed)
			}

			// It is comparable, which is the reason the constructor exists.
			if !built.Parsed() {
				t.Errorf("NewDotted(%d, %d).Parsed() = false, want true", tc.major, tc.minor)
			}
			c, err := built.Compare(parsed)
			if err != nil {
				t.Fatalf("NewDotted(%d, %d).Compare(ParseDotted(%q)): unexpected error: %v", tc.major, tc.minor, rendered, err)
			}
			if c != 0 {
				t.Errorf("NewDotted(%d, %d).Compare(ParseDotted(%q)) = %d, want 0", tc.major, tc.minor, rendered, c)
			}
		})
	}
}

// TestNewDottedRendersUnpadded pins the constructor's one real limitation, so it
// is a stated property rather than a surprise.
//
// Components carry no padding, so NewDotted cannot produce "2.06" -- the string
// the W2025 baseline actually states. It renders "2.6", which ORDERS the same
// and IS a different file. A caller that needs the padded bytes has them already
// and must use ParseDotted.
func TestNewDottedRendersUnpadded(t *testing.T) {
	const padded = "2.06" // what a real W2025 file states

	onDisk, err := ParseDotted(padded)
	if err != nil {
		t.Fatalf("ParseDotted(%q): unexpected error: %v", padded, err)
	}
	built, err := NewDotted(onDisk.Major, onDisk.Minor)
	if err != nil {
		t.Fatalf("NewDotted(%d, %d): unexpected error: %v", onDisk.Major, onDisk.Minor, err)
	}

	// Same ordinal ...
	c, err := built.Compare(onDisk)
	if err != nil {
		t.Fatalf("Compare: unexpected error: %v", err)
	}
	if c != 0 {
		t.Errorf("NewDotted(%d, %d).Compare(%q) = %d, want 0: the components are the same ordinal", onDisk.Major, onDisk.Minor, padded, c)
	}

	// ... different bytes, and that is the documented limit.
	if built.Raw == padded {
		t.Fatalf("NewDotted(%d, %d).Raw = %q; the constructor is documented as unable to render padding, and this test exists to keep that documented limit true", onDisk.Major, onDisk.Minor, padded)
	}
	if built.Raw != "2.6" {
		t.Errorf("NewDotted(%d, %d).Raw = %q, want %q", onDisk.Major, onDisk.Minor, built.Raw, "2.6")
	}
	if onDisk.Raw != padded {
		t.Errorf("ParseDotted(%q).Raw = %q, want the padded bytes back verbatim", padded, onDisk.Raw)
	}
}

// TestNewDottedRejectsNegativeComponents checks that a negative component is
// refused rather than rendered into a string no parser would accept.
//
// The error wraps the same sentinels ParseDotted uses, so a caller can test one
// thing whichever constructor produced the failure. Note what would happen
// without the guard: NewDotted(-2, 6) would render Raw "-2.6", which ParseDotted
// rejects -- a Dotted whose own Raw does not parse, reporting Parsed() == true.
func TestNewDottedRejectsNegativeComponents(t *testing.T) {
	for _, tc := range []struct {
		name         string
		major, minor int
	}{
		{name: "negative major", major: -2, minor: 6},
		{name: "negative minor", major: 2, minor: -6},
		{name: "both negative", major: -2, minor: -6},
		{name: "negative zero-ish minor", major: 1, minor: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewDotted(tc.major, tc.minor)
			if err == nil {
				t.Fatalf("NewDotted(%d, %d) = %#v, want an error", tc.major, tc.minor, got)
			}
			if !errors.Is(err, ErrInvalidDottedVersion) {
				t.Errorf("NewDotted(%d, %d) error = %v, want it to wrap %v", tc.major, tc.minor, err, ErrInvalidDottedVersion)
			}
			if !errors.Is(err, ErrInvalidDottedComponent) {
				t.Errorf("NewDotted(%d, %d) error = %v, want it to wrap %v", tc.major, tc.minor, err, ErrInvalidDottedComponent)
			}
			if got != (Dotted{}) {
				t.Errorf("NewDotted(%d, %d) = %#v, want the zero Dotted on error", tc.major, tc.minor, got)
			}
			// And the zero Dotted a failure returns is unparsed, so a caller that
			// ignores the error still cannot order it.
			if got.Parsed() {
				t.Errorf("the Dotted returned with an error reports Parsed() = true, want false")
			}
		})
	}
}

// TestNewDottedRawAlwaysReparses is the property that keeps a constructed Raw
// honest: whatever NewDotted renders must be something ParseDotted accepts, back
// to the identical value.
//
// A Dotted whose own Raw does not parse would be a new way for Raw and the
// components to disagree -- the exact defect class issue #38 closed, arriving
// from the constructor side instead of the decoder side.
func TestNewDottedRawAlwaysReparses(t *testing.T) {
	for _, major := range []int{0, 1, 2, 9, 10, 99, 100, 2026} {
		for _, minor := range []int{0, 1, 6, 9, 10, 73, 77, 99, 12345} {
			built, err := NewDotted(major, minor)
			if err != nil {
				t.Fatalf("NewDotted(%d, %d): unexpected error: %v", major, minor, err)
			}
			again, err := ParseDotted(built.Raw)
			if err != nil {
				t.Fatalf("ParseDotted(NewDotted(%d, %d).Raw = %q): %v: a constructed Raw must be a parseable dotted version", major, minor, built.Raw, err)
			}
			if again != built {
				t.Errorf("NewDotted(%d, %d) = %#v, but reparsing its own Raw gave %#v", major, minor, built, again)
			}
		}
	}
}

// TestNewDottedOrdersAsAThreshold exercises the use the constructor was added
// for: comparing a version read from a file against a bound stated in code.
func TestNewDottedOrdersAsAThreshold(t *testing.T) {
	// "at least 2.0", stated without a string to parse.
	minimum, err := NewDotted(2, 0)
	if err != nil {
		t.Fatalf("NewDotted(2, 0): unexpected error: %v", err)
	}

	for _, tc := range []struct {
		onDisk    string // as a file states it
		wantBelow bool
	}{
		{onDisk: "1.73", wantBelow: true},
		{onDisk: "1.77", wantBelow: true},
		{onDisk: "1.99", wantBelow: true},
		{onDisk: "2.00", wantBelow: false}, // padded, and equal to the bound
		{onDisk: "2.06", wantBelow: false},
		{onDisk: "10.0", wantBelow: false},
	} {
		t.Run(tc.onDisk, func(t *testing.T) {
			v := mustParseDotted(t, tc.onDisk)
			below, err := v.Less(minimum)
			if err != nil {
				t.Fatalf("%q.Less(%q): unexpected error: %v", tc.onDisk, minimum.Raw, err)
			}
			if below != tc.wantBelow {
				t.Errorf("%q.Less(%q) = %v, want %v", tc.onDisk, minimum.Raw, below, tc.wantBelow)
			}
		})
	}
}
