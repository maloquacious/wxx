// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// Int_t is an attribute this schema states as an INTEGER: no decimal point, ever
// (issue #64).
//
// It is not a formatting preference. Worldographer reads these with Java's
// Integer.parseInt, which rejects anything a decimal point appears in, and the
// failure is a refusal to open the file at all:
//
//	java.lang.NumberFormatException: For input string: "-1.0"
//	    at java.base/java.lang.Integer.parseInt(Unknown Source)
//	    at com.inkwellideas.ographer.task.LoadMapTask.readMapKey(LoadMapTask.java:1925)
//
// Verified by experiment: two files identical but for the spelling of these
// attributes, one opened, the other produced the stack trace above. The
// application's own save of the first wrote every one of them back integrally,
// so the spelling is confirmed against what Worldographer itself emits rather
// than inferred from a sample.
//
// WHY THE TYPE, RATHER THAN A FORMATTER AT THE CALL SITE. The on-disk type of an
// attribute is a fact about the SCHEMA, so it belongs to the codec that owns the
// schema and it belongs in one place. Map_t deliberately does not carry it:
// Map_t is the superset of every supported format (ADR 0004 Decision 6) and
// stays generic, so these fields remain float64 there and are converted at this
// boundary, which is where every other schema-specific decision already lives.
// It also cannot live in Map_t.Validate: package wxx cannot see codec knowledge
// without an import cycle, which is the same reason issue #20 scoped out its
// version half.
//
// The declaration that matters is schema.go. It already typed 15 of this
// schema's 22 integer attributes as int and lied about 7 -- and the 7 lies were
// exactly the attributes that broke. Naming the type there is what makes the
// struct a truthful description of the document instead of a rough one.
type Int_t int

// String renders the value as the file must state it.
//
// Note for callers: fmt's %q does NOT call this -- on an integer %q produces a
// character literal -- so emit code must pass s.String() explicitly. The
// encoders here write attributes through fmt.Sprintf(" name=%q", ...), so this
// is a live trap rather than a hypothetical one.
func (v Int_t) String() string {
	return strconv.Itoa(int(v))
}

// UnmarshalXMLAttr parses an integer attribute, and refuses anything else.
//
// It is STRICT, and that is a decision rather than an oversight. This codec
// wrote "-1.0" into these attributes between 2026-07-13 (b1efddf) and issue
// #64's fix, and an earlier draft of the fix accepted that spelling so such a
// file would repair itself on the next decode-encode cycle. The maintainer has
// ruled the regression window irrelevant -- those files have no consumers -- so
// the repair path buys nothing, and accepting the spelling would mean quietly
// reading a file Worldographer itself refuses to open, then carrying the value
// to a write that has to refuse it anyway. Failing at the point the file is
// wrong is the better report.
//
// The consequence is deliberate: a file this codec wrote during that window will
// not decode. The error says so, because "invalid syntax" from strconv would
// leave the reader guessing which of a hundred attributes it meant.
func (v *Int_t) UnmarshalXMLAttr(attr xml.Attr) error {
	n, err := strconv.Atoi(strings.TrimSpace(attr.Value))
	if err != nil {
		return errors.Join(wxx.ErrInvalidIntegerAttribute, fmt.Errorf(
			"@%s = %q: this schema states it as an integer, and Worldographer reads it with Integer.parseInt (issue #64)",
			attr.Name.Local, attr.Value))
	}
	*v = Int_t(n)
	return nil
}

// toInt converts a Map_t float64 into the integer this schema states, and
// refuses a value the schema cannot express.
//
// The refusal is the point. Map_t models these as float64 because it is the
// superset of formats and does not know this schema's types, so nothing stops a
// caller assigning 50.5 to a field the file must state as an integer. Rounding
// it would put a value on disk the caller never asked for, and truncating it
// would do the same more quietly. The encoder refuses instead, before writing a
// byte, exactly as it refuses a classic ROWS map (#20) and an unmodeled stub
// downgrade (ADR 0004 Decision 7).
//
// path names the attribute in the on-disk vocabulary, so the caller is told
// which of their fields is the problem rather than which of ours.
//
// Decoded maps always satisfy this: every value reaching Map_t through a decoder
// came from an integer attribute in the first place. Only a hand-built or
// hand-edited map can fail here.
func toInt(path string, f float64) (Int_t, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, errors.Join(wxx.ErrInvalidIntegerAttribute, fmt.Errorf(
			"%s: %s is not a whole number, and this schema states the attribute as an integer; Worldographer will not open a file with a decimal point here (issue #64)",
			path, strconv.FormatFloat(f, 'g', -1, 64)))
	}
	if f > math.MaxInt32 || f < math.MinInt32 {
		return 0, errors.Join(wxx.ErrInvalidIntegerAttribute, fmt.Errorf(
			"%s: %s does not fit the 32-bit integer Worldographer parses (issue #64)",
			path, strconv.FormatFloat(f, 'g', -1, 64)))
	}
	return Int_t(f), nil
}
