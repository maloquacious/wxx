// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/maloquacious/wxx"
)

// decodeInformations copies the <informations> tree into the domain map,
// every level of nested <information> included.
func decodeInformations(src Informations_t, w *wxx.Map_t) {
	w.Informations = &wxx.Informations_t{}
	for _, info := range src.Informations {
		w.Informations.Informations = append(w.Informations.Informations, decodeInformation(info))
	}
	w.Informations.InnerText = src.InnerText
}

// decodeInformation copies one <information> and, recursively, every entry
// nested inside it. It used to copy two levels and stop, which dropped the
// third level -- every god in every pantheon -- on decode (issue #69).
func decodeInformation(info Information_t) *wxx.Information_t {
	wInfo := &wxx.Information_t{
		Uuid:         info.Uuid,
		Type:         info.Type,
		Title:        info.Title,
		Rulers:       info.Rulers,
		Government:   info.Government,
		Cultures:     info.Cultures,
		Language:     info.Language,
		ReligionType: info.ReligionType,
		Culture:      info.Culture,
		HolySymbol:   info.HolySymbol,
		Domains:      info.Domains,
		InnerText:    info.InnerText,
	}
	for _, detail := range info.Details {
		wInfo.Details = append(wInfo.Details, decodeInformation(detail))
	}
	return wInfo
}

// encodeInformations writes <informations> in the layout Worldographer 2.06,
// 2.07 and 2.08 use (issue #107): the start tag, a newline, then each entry
// followed by a newline. That layout is what the wrapper's chardata holds
// after decode -- one newline, plus one per entry -- so it is written back
// as plain newlines. Chardata in any other shape (a map wxx did not decode)
// is written as escaped text ahead of the entries, so a re-decode still gives
// back exactly informations.InnerText.
//
// Every body is checked before anything is written: a body holding "]]>"
// cannot be written as CDATA.
func encodeInformations(informations *wxx.Informations_t, wb *bytes.Buffer) error {
	for i, information := range informations.Informations {
		if err := checkInformationBodies(information, fmt.Sprintf("map/informations/information[%d]", i+1)); err != nil {
			return err
		}
	}
	wb.WriteString("<informations>")
	layout := strings.Repeat("\n", 1+len(informations.Informations))
	if informations.InnerText == layout {
		wb.WriteString("\n")
		for _, information := range informations.Informations {
			encodeInformation(information, wb)
			wb.WriteString("\n")
		}
	} else {
		wb.WriteString(encodeInnerText(informations.InnerText))
		for _, information := range informations.Informations {
			encodeInformation(information, wb)
		}
	}
	wb.WriteString("</informations>\n")
	return nil
}

// checkInformationBodies refuses an entry, at any depth, whose body holds the
// CDATA terminator "]]>". Worldographer never writes one: its lore editor
// stores a typed "]]>" as "]]&gt;" (see
// testdata/2025-2.06-13x11-941577-cdata-guard.wxx), so a body holding it came
// from a caller, and splitting the section would be a spelling the app has
// never been seen to read (issue #107).
func checkInformationBodies(information *wxx.Information_t, path string) error {
	if strings.Contains(information.InnerText, cdataTerminator) {
		return errors.Join(wxx.ErrCDATATerminator,
			fmt.Errorf("%s (title %q): body holds %q", path, information.Title, cdataTerminator))
	}
	for i, detail := range information.Details {
		if err := checkInformationBodies(detail, fmt.Sprintf("%s/information[%d]", path, i+1)); err != nil {
			return err
		}
	}
	return nil
}

// encodeInformation writes one <information> and, recursively, its nested
// entries, in the app's layout (issue #107): the start tag, the body as a
// CDATA section, a newline, each nested entry followed by a newline, then a
// newline and the end tag. The body was checked for "]]>" by
// checkInformationBodies.
//
// After decode, InnerText is the body followed by that layout's newlines --
// two, plus one per nested entry -- because a parser hands CDATA and the
// whitespace between elements back as one run of chardata. Those trailing
// newlines are split off and written as layout. InnerText that does not end
// in them (a map wxx did not decode) is written whole as the CDATA body with
// the entries back to back, so a re-decode still gives back exactly
// information.InnerText.
func encodeInformation(information *wxx.Information_t, wb *bytes.Buffer) {
	wb.WriteString("<information")
	wb.WriteString(fmt.Sprintf(" uuid=%s", xmlAttr(information.Uuid)))
	wb.WriteString(fmt.Sprintf(" type=%s", xmlAttr(information.Type)))
	wb.WriteString(fmt.Sprintf(" title=%s", xmlAttr(information.Title)))
	encodeLoreAttr(wb, "rulers", information.Rulers)
	encodeLoreAttr(wb, "government", information.Government)
	encodeLoreAttr(wb, "cultures", information.Cultures)
	encodeLoreAttr(wb, "language", information.Language)
	encodeLoreAttr(wb, "religionType", information.ReligionType)
	encodeLoreAttr(wb, "culture", information.Culture)
	encodeLoreAttr(wb, "holySymbol", information.HolySymbol)
	encodeLoreAttr(wb, "domains", information.Domains)
	// The app ends the start tag with " >" when it states any lore attribute,
	// and with ">" when it states none: 953 and 90 entries across the tracked
	// W2025 fixtures, no exception (issue #107).
	if information.Rulers != nil || information.Government != nil || information.Cultures != nil ||
		information.Language != nil || information.ReligionType != nil || information.Culture != nil ||
		information.HolySymbol != nil || information.Domains != nil {
		wb.WriteString(" ")
	}
	wb.WriteString(">")
	layout := strings.Repeat("\n", 2+len(information.Details))
	if body, ok := strings.CutSuffix(information.InnerText, layout); ok {
		writeCDATA(wb, body)
		wb.WriteString("\n")
		for _, detail := range information.Details {
			encodeInformation(detail, wb)
			wb.WriteString("\n")
		}
		wb.WriteString("\n")
	} else {
		writeCDATA(wb, information.InnerText)
		for _, detail := range information.Details {
			encodeInformation(detail, wb)
		}
	}
	wb.WriteString("</information>")
}

// encodeLoreAttr writes one of the eight optional lore attributes, and only if
// the source stated it (issue #66). nil means the source did not; a pointer to
// "" means it stated the attribute empty, and that is written back as "".
// Gating on the value instead would drop the domains="" Worldographer writes on
// every Religion entry.
func encodeLoreAttr(wb *bytes.Buffer, name string, value *string) {
	if value == nil {
		return
	}
	wb.WriteString(fmt.Sprintf(" %s=%s", name, xmlAttr(*value)))
}
