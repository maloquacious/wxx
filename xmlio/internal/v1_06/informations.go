// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

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

func encodeInformations(informations *wxx.Informations_t, wb *bytes.Buffer) error {
	wb.WriteString("<informations>")
	// The wrapper's chardata (whitespace between <information> children) is
	// emitted here as escaped text. The <information> children below are emitted
	// back-to-back with no surrounding whitespace, so on re-decode the wrapper's
	// chardata is exactly informations.InnerText.
	wb.WriteString(encodeInnerText(informations.InnerText))
	for _, information := range informations.Informations {
		if err := encodeInformation(information, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</informations>\n")
	return nil
}

func encodeInformation(information *wxx.Information_t, wb *bytes.Buffer) error {
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
	wb.WriteString(">")
	// Emit this element's chardata first, then its nested <information> children
	// back-to-back with no surrounding whitespace, so on re-decode this element's
	// chardata is exactly information.InnerText. The children are written by this
	// same function, so the tree is written to whatever depth it has.
	wb.WriteString(encodeInnerText(information.InnerText))
	for _, detail := range information.Details {
		if err := encodeInformation(detail, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</information>")
	return nil
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
