// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"cmp"
	"errors"
	"strconv"
	"strings"
)

// Dotted is an on-disk dotted version. It is NOT semver: "2.06" != "2.6".
// Raw is authoritative for output; the components exist to compare.
//
// Worldographer writes these values as zero-padded ordinals (map/@version
// "2.06", map/@schema "1.06"). A semver round-trip drops the padding and
// returns "2.6", which is a different string and therefore a different file.
// Dotted keeps the bytes it was given and parses the components only so that
// two versions can be ordered.
type Dotted struct {
	Raw   string // verbatim, exactly as read or to be written
	Major int
	Minor int
}

// ParseDotted parses an on-disk dotted version such as "2.06" or "1.73".
//
// Raw is set to s verbatim, including any zero padding; Major and Minor are
// parsed only to support comparison. The grammar is deliberately strict:
// exactly two components separated by a single ".", each one or more ASCII
// digits. No sign, no whitespace, no third component.
func ParseDotted(s string) (Dotted, error) {
	if s == "" {
		return Dotted{}, errors.Join(ErrInvalidDottedVersion, ErrMissingVersion)
	}
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return Dotted{}, errors.Join(ErrInvalidDottedVersion, ErrInvalidDottedComponentCount)
	}
	major, err := parseDottedComponent(parts[0])
	if err != nil {
		return Dotted{}, errors.Join(ErrInvalidDottedVersion, err)
	}
	minor, err := parseDottedComponent(parts[1])
	if err != nil {
		return Dotted{}, errors.Join(ErrInvalidDottedVersion, err)
	}
	return Dotted{Raw: s, Major: major, Minor: minor}, nil
}

// parseDottedComponent converts one dotted component to an int. It accepts only
// ASCII digits so that strconv's tolerance for signs does not leak in.
func parseDottedComponent(s string) (int, error) {
	if s == "" {
		return 0, ErrInvalidDottedComponent
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, ErrInvalidDottedComponent
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.Join(ErrInvalidDottedComponent, err)
	}
	return n, nil
}

// Compare orders d against other by Major, then Minor, returning -1, 0, or +1.
//
// Raw is never consulted. "2.06" and "2.6" are different strings but the same
// ordinal, so they compare equal. That is the point of the type: the components
// answer "which is newer" and Raw answers "what do we write", and the two must
// not be confused.
func (d Dotted) Compare(other Dotted) int {
	if c := cmp.Compare(d.Major, other.Major); c != 0 {
		return c
	}
	return cmp.Compare(d.Minor, other.Minor)
}

// Less reports whether d orders before other. See Compare.
func (d Dotted) Less(other Dotted) bool {
	return d.Compare(other) < 0
}

// String returns the verbatim string this Dotted was parsed from.
//
// It never re-renders from Major and Minor: that would turn "2.06" into "2.6"
// and corrupt any file written from it. Anything written to disk comes from Raw.
func (d Dotted) String() string {
	return d.Raw
}
