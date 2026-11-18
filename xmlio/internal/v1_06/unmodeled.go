// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/maloquacious/wxx"
)

// unmodeled_t is embedded in every schema struct. It catches the attributes and
// child elements the struct does not declare, so that decode can refuse them
// instead of dropping them (issue #161).
//
// xml.Unmarshal ignores what a struct does not declare. Before #161 a file
// stating an attribute wxx had never seen decoded without complaint, and the
// attribute was gone after the next write. Worldographer 2.09's label
// @lineSpacing and @textAlignment were the first known case, but nothing was
// specific to them. <extraTerrain> already refused unmodeled content (issue
// #34); this is that rule applied to every element.
//
// The catch-alls are filled during the single xml.Unmarshal pass, so refusing
// costs no second parse of the document.
type unmodeled_t struct {
	OtherAttrs []xml.Attr           `xml:",any,attr"`
	Other      []unmodeledElement_t `xml:",any"`
}

// unmodeledAttrs_t is embedded instead of unmodeled_t in a schema struct that
// keeps its element's content as raw inner XML (issue #164). Its children are
// preserved, not dropped, so only an unknown attribute on the element itself is
// refused.
type unmodeledAttrs_t struct {
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// unmodeledText_t is embedded in every schema struct that does not keep its
// element's text. Whitespace is layout and is allowed; anything else would be
// dropped, so it is refused.
type unmodeledText_t struct {
	OtherText string `xml:",chardata"`
}

// unmodeledElement_t records only the name of a child element a schema struct
// does not model, so decode can refuse it by name.
type unmodeledElement_t struct {
	XMLName xml.Name
}

var (
	unmodeledType      = reflect.TypeOf(unmodeled_t{})
	unmodeledAttrsType = reflect.TypeOf(unmodeledAttrs_t{})
	unmodeledTextType  = reflect.TypeOf(unmodeledText_t{})
)

// refuseUnmodeledContent walks the decoded schema and returns an error naming
// the first attribute, child element or text that no schema struct models.
//
// Paths are spelled like the rest of the codec's diagnostics, with each element's
// position among same-named siblings: map/labels[1]/label[3].
func refuseUnmodeledContent(m *XMLSchema) error {
	return refuseUnmodeledIn(reflect.ValueOf(m).Elem(), "map")
}

func refuseUnmodeledIn(v reflect.Value, path string) error {
	t := v.Type()
	// The element's own catch-alls first, then its children, so the error names
	// the outermost unmodeled content.
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		switch f.Type {
		case unmodeledAttrsType:
			if err := refuseUnmodeledAttrs(v.Field(i), path); err != nil {
				return err
			}
		case unmodeledType:
			// The embedded type is unexported, so its value cannot be taken with
			// Interface; its exported fields are read through reflect instead.
			u := v.Field(i)
			if err := refuseUnmodeledAttrs(u, path); err != nil {
				return err
			}
			if other := u.FieldByName("Other"); other.Len() != 0 {
				name := other.Index(0).FieldByName("XMLName").FieldByName("Local").String()
				return errors.Join(wxx.ErrUnmodeledContent, fmt.Errorf("%s/%s[1]: unmodeled element <%s>; wxx does not understand it and will not drop it silently (issue #161)", path, name, name))
			}
		case unmodeledTextType:
			if text := strings.TrimSpace(v.Field(i).FieldByName("OtherText").String()); text != "" {
				if len(text) > 40 {
					text = text[:40] + "..."
				}
				return errors.Join(wxx.ErrUnmodeledContent, fmt.Errorf("%s: unmodeled text %q; wxx does not understand it and will not drop it silently (issue #161)", path, text))
			}
		}
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, ok := childElementName(f)
		if !ok {
			continue
		}
		fv := v.Field(i)
		for fv.Kind() == reflect.Pointer {
			if fv.IsNil() {
				break
			}
			fv = fv.Elem()
		}
		switch fv.Kind() {
		case reflect.Struct:
			if err := refuseUnmodeledIn(fv, fmt.Sprintf("%s/%s[1]", path, name)); err != nil {
				return err
			}
		case reflect.Slice:
			for j := 0; j < fv.Len(); j++ {
				ev := fv.Index(j)
				for ev.Kind() == reflect.Pointer && !ev.IsNil() {
					ev = ev.Elem()
				}
				if ev.Kind() != reflect.Struct {
					continue
				}
				if err := refuseUnmodeledIn(ev, fmt.Sprintf("%s/%s[%d]", path, name, j+1)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// refuseUnmodeledAttrs returns an error naming the first attribute caught by the
// OtherAttrs field of u, an embedded unmodeled_t or unmodeledAttrs_t.
func refuseUnmodeledAttrs(u reflect.Value, path string) error {
	attrs := u.FieldByName("OtherAttrs")
	if attrs.Len() == 0 {
		return nil
	}
	a := attrs.Index(0)
	name := a.FieldByName("Name").FieldByName("Local").String()
	if space := a.FieldByName("Name").FieldByName("Space").String(); space != "" {
		name = space + ":" + name
	}
	return errors.Join(wxx.ErrUnmodeledContent, fmt.Errorf("%s: unmodeled attribute @%s=%q; wxx does not understand it and will not drop it silently (issue #161)", path, name, a.FieldByName("Value").String()))
}

// childElementName reports the element name of a field that decodes a child
// element, and false for attributes, text, catch-alls and XMLName.
func childElementName(f reflect.StructField) (string, bool) {
	if f.Anonymous || f.Name == "XMLName" || !f.IsExported() {
		return "", false
	}
	tag := f.Tag.Get("xml")
	if tag == "-" {
		return "", false
	}
	name, flags, _ := strings.Cut(tag, ",")
	if flags != "" && flags != "omitempty" {
		return "", false // attr, chardata, any, innerxml, comment
	}
	if name == "" {
		name = f.Name
	}
	return name, true
}
