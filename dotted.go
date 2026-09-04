// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

import (
	"cmp"
	"errors"
	"fmt"
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
//
// A Dotted carries its bytes whether or not they were parseable, because a
// decoder must be able to keep a malformed on-disk version verbatim without
// failing (see the fallbacks in the codec packages). Such a value has NO
// components — not zero components, none — and the unexported parsed flag is
// what makes that distinction real rather than documented. Only the two
// constructors, ParseDotted and NewDotted, set it, so a composite literal from
// any other package (wxx.Dotted{Raw: s}, and the zero Dotted with it) is
// unparsed by construction and cannot claim an ordering it has no data for.
// Compare and Less report an error rather than answer from components that are
// absent; see Parsed.
//
// This is issue #38. Before it, a malformed version decoded to {Raw: "garbage",
// Major: 0, Minor: 0}, which compared EQUAL to "0.0" and LESS than every real
// version — a silent wrong answer waiting for the first caller to order two
// versions.
type Dotted struct {
	Raw   string // verbatim, exactly as read or to be written
	Major int
	Minor int

	// parsed reports whether Major and Minor were derived from Raw by
	// ParseDotted. It is unexported so that it cannot be set from outside this
	// package: the only way to obtain a comparable Dotted is to parse one, which
	// is what keeps components and Raw from disagreeing.
	parsed bool
}

// ParseDotted parses an on-disk dotted version such as "2.06" or "1.73".
//
// Raw is set to s verbatim, including any zero padding; Major and Minor are
// parsed only to support comparison. The grammar is deliberately strict:
// exactly two components separated by a single ".", each one or more ASCII
// digits. No sign, no whitespace, no third component.
//
// On success the result reports Parsed; on failure it is the zero Dotted, which
// does not. Parsing a string is the usual way to obtain a comparable Dotted --
// the values come off disk -- but see NewDotted for a caller that holds
// components rather than bytes.
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
	return Dotted{Raw: s, Major: major, Minor: minor, parsed: true}, nil
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

// NewDotted builds a comparable Dotted from components, for a caller that has
// no string to parse -- a threshold to compare against, a version computed
// rather than read.
//
// Raw is rendered as "<major>.<minor>", and that rendering is the whole of what
// this constructor can know: components carry no padding, so NewDotted(2, 6)
// gives Raw "2.6" and there is NO WAY to reach "2.06" through it. That is not an
// omission to be fixed with a padding argument. "2.06" and "2.6" are different
// files (ADR 0004 Decision 1), and a caller who wants the padded one HAS the
// string it wants -- ParseDotted("2.06") is the constructor for that, and it is
// the only one, because a Raw supplied alongside components is a second source
// for the same fact and the two can disagree. That disagreement is the class of
// bug issue #38 closed; this constructor does not reopen it from the other side.
//
// So: use this to ORDER versions, not to name one. A Dotted from here is a
// legitimate operand of Compare and Less, and its Raw is a legitimate string to
// display -- but passing that Raw as an encode target names the unpadded string
// literally, which for the supported releases is no application version at all
// and misses the registry cleanly.
//
// It rejects a negative component, wrapping the same sentinels ParseDotted uses
// for the same fact, so a caller can test one thing whichever constructor it
// used. There is no upper bound: no dotted component has a meaningful ceiling.
//
// NewDotted(major, minor) and ParseDotted(fmt.Sprintf("%d.%d", major, minor))
// produce identical values, flag included. This is the constructor the #38
// amendment to ADR 0004 left open, added in the same change on the maintainer's
// call.
func NewDotted(major, minor int) (Dotted, error) {
	if major < 0 {
		return Dotted{}, errors.Join(ErrInvalidDottedVersion, ErrInvalidDottedComponent, fmt.Errorf("major %d: negative", major))
	}
	if minor < 0 {
		return Dotted{}, errors.Join(ErrInvalidDottedVersion, ErrInvalidDottedComponent, fmt.Errorf("minor %d: negative", minor))
	}
	return Dotted{
		Raw:    strconv.Itoa(major) + "." + strconv.Itoa(minor),
		Major:  major,
		Minor:  minor,
		parsed: true,
	}, nil
}

// Parsed reports whether Major and Minor mean anything.
//
// It is false for the zero Dotted and for one a decoder built from a version
// string that does not fit the dotted grammar. Raw is trustworthy in every
// case — it is what goes back to disk — but only a Parsed Dotted can be
// ordered, and only ParseDotted and NewDotted produce one.
func (d Dotted) Parsed() bool {
	return d.parsed
}

// Compare orders d against other by Major, then Minor, returning -1, 0, or +1.
//
// Raw is never consulted. "2.06" and "2.6" are different strings but the same
// ordinal, so they compare equal. That is the point of the type: the components
// answer "which is newer" and Raw answers "what do we write", and the two must
// not be confused.
//
// It returns an error, wrapping ErrUnparsedDottedVersion, when either side is
// not Parsed. There is no ordering to report in that case and the returned int
// is not one: an unparsed Dotted has no components, so answering from them
// would be inventing a fact about a file (issue #38). Callers that hold a
// version straight from a decoder must handle this; callers holding one from
// ParseDotted or NewDotted cannot reach it.
func (d Dotted) Compare(other Dotted) (int, error) {
	if !d.parsed || !other.parsed {
		return 0, errors.Join(ErrUnparsedDottedVersion, fmt.Errorf("compare %q to %q: %s", d.Raw, other.Raw, unparsedSide(d, other)))
	}
	if c := cmp.Compare(d.Major, other.Major); c != 0 {
		return c, nil
	}
	return cmp.Compare(d.Minor, other.Minor), nil
}

// unparsedSide names which operand of a failed Compare lacked components, so
// the error says which version string is the unusable one.
func unparsedSide(d, other Dotted) string {
	switch {
	case !d.parsed && !other.parsed:
		return "neither version was parsed into components"
	case !d.parsed:
		return fmt.Sprintf("%q was not parsed into components", d.Raw)
	default:
		return fmt.Sprintf("%q was not parsed into components", other.Raw)
	}
}

// Less reports whether d orders before other. See Compare, including the error.
func (d Dotted) Less(other Dotted) (bool, error) {
	c, err := d.Compare(other)
	if err != nil {
		return false, err
	}
	return c < 0, nil
}

// String returns the verbatim string this Dotted was parsed from.
//
// It never re-renders from Major and Minor: that would turn "2.06" into "2.6"
// and corrupt any file written from it. Anything written to disk comes from Raw.
// It is defined for an unparsed Dotted too — Raw is exactly what such a value
// exists to carry.
func (d Dotted) String() string {
	return d.Raw
}
