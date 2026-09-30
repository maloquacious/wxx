// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v0_77

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// Int_t is an attribute this schema states as an INTEGER: no decimal point, ever
// (issues #64 and #67).
//
// It is the classic twin of v1_06.Int_t, whose comment carries the full
// rationale: Worldographer reads these attributes with Java's Integer.parseInt,
// so "-1.0" where the format says "-1" is a file the application refuses to
// open. #64 proved that for W2025's <mapkey>. Classic's <mapkey> has the same
// five attributes, and a survey of every classic fixture found each always
// spelled integrally (height="-1", backgroundopacity="50", titleScale="80",
// scaleScale="65", entryScale="55").
//
// Verified by experiment (issue #67): the 1.77 blank fixture with only
// height="-1" changed to height="-1.0" is refused by classic Worldographer 1.77
// itself, and by 2.08 when it opens the classic file:
//
//	java.lang.NumberFormatException: For input string: "-1.0"
//	    at java.base/java.lang.Integer.parseInt(Unknown Source)
//	    at com.inkwellideas.ographer.task.LoadMapTask.readMapKey(LoadMapTask.java:1432)   (1.77)
//	    at com.inkwellideas.ographer.task.LoadMapTask.readMapKey(LoadMapTask.java:1925)   (2.08)
//
// This schema used to declare those five float64, which was harmless only
// because the classic encoder writes <mapkey> as a hard-coded constant and never
// reads Map_t.MapKey. Whoever closes that gap will write the fields from Map_t,
// whose MapKey fields are float64. With the float64 declarations, the obvious
// implementation -- floats() on each field -- is exactly how #64 happened in the
// W2025 codec. The declaration here is the record of what the format is; the
// spelling audit in xmlio (TestClassicIntegerAttributeSpelling) is the guard
// that fails if the encoder writes one of these with a decimal point. When the
// gap is closed, convert each Map_t value with a refusal like v1_06's toInt
// rather than by rounding.
type Int_t int

// String renders the value as the file must state it.
func (v Int_t) String() string {
	return strconv.Itoa(int(v))
}

// UnmarshalXMLAttr parses an integer attribute and refuses anything else, as the
// W2025 codec does and as encoding/xml already does for this schema's plain int
// attributes. The error names the attribute and the value, which a bare
// strconv "invalid syntax" would not.
func (v *Int_t) UnmarshalXMLAttr(attr xml.Attr) error {
	n, err := strconv.Atoi(strings.TrimSpace(attr.Value))
	if err != nil {
		return errors.Join(wxx.ErrInvalidIntegerAttribute, fmt.Errorf(
			"@%s = %q: this schema states it as an integer, and Worldographer reads it with Integer.parseInt (issues #64, #67)",
			attr.Name.Local, attr.Value))
	}
	*v = Int_t(n)
	return nil
}
