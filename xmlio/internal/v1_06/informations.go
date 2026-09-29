// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeInformations copies the <informations> tree (including nested
// <information> detail children) into the domain map.
func decodeInformations(src Informations_t, w *wxx.Map_t) {
	w.Informations = &wxx.Informations_t{}
	for _, info := range src.Informations {
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
			wDetail := &wxx.InformationDetail_t{
				Uuid:         detail.Uuid,
				Type:         detail.Type,
				Title:        detail.Title,
				Rulers:       detail.Rulers,
				Government:   detail.Government,
				Cultures:     detail.Cultures,
				Language:     detail.Language,
				ReligionType: detail.ReligionType,
				Culture:      detail.Culture,
				HolySymbol:   detail.HolySymbol,
				Domains:      detail.Domains,
				InnerText:    detail.InnerText,
			}
			wInfo.Details = append(wInfo.Details, wDetail)
		}

		w.Informations.Informations = append(w.Informations.Informations, wInfo)
	}
	w.Informations.InnerText = src.InnerText
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
	wb.WriteString(fmt.Sprintf(" uuid=%q", information.Uuid))
	wb.WriteString(fmt.Sprintf(" type=%q", information.Type))
	wb.WriteString(fmt.Sprintf(" title=%q", information.Title))
	encodeLoreAttr(wb, "rulers", information.Rulers)
	encodeLoreAttr(wb, "government", information.Government)
	encodeLoreAttr(wb, "cultures", information.Cultures)
	encodeLoreAttr(wb, "language", information.Language)
	encodeLoreAttr(wb, "religionType", information.ReligionType)
	encodeLoreAttr(wb, "culture", information.Culture)
	encodeLoreAttr(wb, "holySymbol", information.HolySymbol)
	encodeLoreAttr(wb, "domains", information.Domains)
	wb.WriteString(">")
	// Emit this element's chardata first, then its <information> detail children
	// back-to-back with no surrounding whitespace, so on re-decode this element's
	// chardata is exactly information.InnerText.
	wb.WriteString(encodeInnerText(information.InnerText))
	for _, detail := range information.Details {
		if err := encodeInformationDetail(detail, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</information>")
	return nil
}

func encodeInformationDetail(detail *wxx.InformationDetail_t, wb *bytes.Buffer) error {
	wb.WriteString("<information")
	wb.WriteString(fmt.Sprintf(" uuid=%q", detail.Uuid))
	wb.WriteString(fmt.Sprintf(" type=%q", detail.Type))
	wb.WriteString(fmt.Sprintf(" title=%q", detail.Title))
	encodeLoreAttr(wb, "rulers", detail.Rulers)
	encodeLoreAttr(wb, "government", detail.Government)
	encodeLoreAttr(wb, "cultures", detail.Cultures)
	encodeLoreAttr(wb, "language", detail.Language)
	encodeLoreAttr(wb, "religionType", detail.ReligionType)
	encodeLoreAttr(wb, "culture", detail.Culture)
	encodeLoreAttr(wb, "holySymbol", detail.HolySymbol)
	encodeLoreAttr(wb, "domains", detail.Domains)
	wb.WriteString(">")
	wb.WriteString(encodeInnerText(detail.InnerText))
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
	wb.WriteString(fmt.Sprintf(" %s=%q", name, *value))
}
