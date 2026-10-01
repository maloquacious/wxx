// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeLabels copies each standalone <label> into the domain map.
func decodeLabels(src Labels_t, w *wxx.Map_t) error {
	var err error
	for _, mLabel := range src.Labels {
		wLabel := &wxx.Label_t{
			MapLayer:    mLabel.MapLayer,
			Style:       mLabel.Style,
			FontFace:    mLabel.FontFace,
			OutlineSize: mLabel.OutlineSize,
			Rotate:      mLabel.Rotate,
			IsBold:      mLabel.IsBold,
			IsItalic:    mLabel.IsItalic,
			IsWorld:     mLabel.IsWorld,
			IsContinent: mLabel.IsContinent,
			IsKingdom:   mLabel.IsKingdom,
			IsProvince:  mLabel.IsProvince,
			IsGMOnly:    mLabel.IsGMOnly,
			Tags:        mLabel.Tags,

			DropShadowColor:  mLabel.DropShadowColor,
			DropShadowRadius: mLabel.DropShadowRadius,
			DropShadowSpread: mLabel.DropShadowSpread,
		}
		if wLabel.Color, err = decodeRgba(mLabel.Color); err != nil {
			return fmt.Errorf("label.color: %w", err)
		}
		if wLabel.OutlineColor, err = decodeRgba(mLabel.OutlineColor); err != nil {
			return fmt.Errorf("label.outlineColor: %w", err)
		}
		if wLabel.BackgroundColor, err = decodeLabelBackgroundColor(mLabel.BackgroundColor); err != nil {
			return fmt.Errorf("label.backgroundColor: %w", err)
		}
		wLabel.Location = &wxx.LabelLocation_t{
			ViewLevel: mLabel.Location.ViewLevel,
			X:         mLabel.Location.X,
			Y:         mLabel.Location.Y,
			Scale:     mLabel.Location.Scale,
		}
		wLabel.InnerText = mLabel.InnerText
		w.Labels = append(w.Labels, wLabel)
	}
	return nil
}

// decodeLabelBackgroundColor decodes label/@backgroundColor for both label
// contexts (labels/label and feature/label). No saved fixture carries it; an
// app check (#115) showed Worldographer 2025 draws it on a labels/label but not
// on a feature/label, and that the app's save keeps it in both contexts and
// writes it only when set. wxx keeps what the app keeps, drawn or not. An
// absent attribute is nil; opaque black is a colour, not nil.
func decodeLabelBackgroundColor(s string) (*wxx.RGBA_t, error) {
	return decodeZeroableRgba(s)
}

func encodeLabels(labels []*wxx.Label_t, wb *bytes.Buffer) error {
	wb.WriteString("<labels>\n")
	for _, label := range labels {
		if err := encodeLabel(label, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</labels>\n")
	return nil
}

func encodeLabel(label *wxx.Label_t, wb *bytes.Buffer) error {
	wb.WriteString("<label")
	wb.WriteString(fmt.Sprintf("  mapLayer=%s", xmlAttr(label.MapLayer)))
	wb.WriteString(fmt.Sprintf(" style=%s", xmlAttr(label.Style)))       // can be null!
	wb.WriteString(fmt.Sprintf(" fontFace=%s", xmlAttr(label.FontFace))) // can be null!
	wb.WriteString(fmt.Sprintf(" color=%s", xmlAttr(rgbas(label.Color))))
	// The app writes backgroundColor after color when the label has one and omits
	// it when it does not (#115), so nil is the absent attribute. Opaque black is a
	// real background and is written like any other colour.
	if label.BackgroundColor != nil {
		wb.WriteString(fmt.Sprintf(" backgroundColor=%s", xmlAttr(rgbas(label.BackgroundColor))))
	}
	wb.WriteString(fmt.Sprintf(" outlineColor=%s", xmlAttr(rgbas(label.OutlineColor))))
	wb.WriteString(fmt.Sprintf(" outlineSize=%s", xmlAttr(floats(label.OutlineSize))))
	// The W2025 drop-shadow trio is present all-or-none in real data;
	// dropShadowColor is "null" or an RGBA string when present, never empty, so an
	// empty DropShadowColor reliably means "absent from the source". Gate the whole
	// group on that sentinel so a round-trip does not spuriously add the attributes
	// (ADR 0002: never emit what was not on input). Do not gate on the numeric
	// fields: 0 is a legal radius/spread value.
	//
	// The source writes the trio between outlineSize and rotate; emit it there so a
	// round trip matches the source's attribute order.
	if label.DropShadowColor != "" {
		color, err := rgbaAttr("map/labels/label/@dropShadowColor", label.DropShadowColor, true)
		if err != nil {
			return err
		}
		wb.WriteString(fmt.Sprintf(" dropShadowColor=%s", xmlAttr(color))) // nullable string ("null")
		wb.WriteString(fmt.Sprintf(" dropShadowRadius=%s", xmlAttr(floats(label.DropShadowRadius))))
		wb.WriteString(fmt.Sprintf(" dropShadowSpread=%s", xmlAttr(floats(label.DropShadowSpread))))
	}
	wb.WriteString(fmt.Sprintf(" rotate=%s", xmlAttr(floats(label.Rotate))))
	wb.WriteString(fmt.Sprintf(" isBold=%s", xmlAttr(bools(label.IsBold))))
	wb.WriteString(fmt.Sprintf(" isItalic=%s", xmlAttr(bools(label.IsItalic))))
	wb.WriteString(fmt.Sprintf(" isWorld=%s", xmlAttr(bools(label.IsWorld))))
	wb.WriteString(fmt.Sprintf(" isContinent=%s", xmlAttr(bools(label.IsContinent))))
	wb.WriteString(fmt.Sprintf(" isKingdom=%s", xmlAttr(bools(label.IsKingdom))))
	wb.WriteString(fmt.Sprintf(" isProvince=%s", xmlAttr(bools(label.IsProvince))))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%s", xmlAttr(bools(label.IsGMOnly))))
	wb.WriteString(fmt.Sprintf(" tags=%s", xmlAttr(label.Tags)))
	wb.WriteString(">")
	if err := encodeLabelLocation(label.Location, wb); err != nil {
		return err
	}
	if label.InnerText != "" {
		wb.WriteString(encodeInnerText(label.InnerText))
	}
	wb.WriteString("</label>\n")
	return nil
}

func encodeLabelLocation(location *wxx.LabelLocation_t, wb *bytes.Buffer) error {
	wb.WriteString("<location")
	wb.WriteString(fmt.Sprintf(" viewLevel=%s", xmlAttr(location.ViewLevel)))
	wb.WriteString(fmt.Sprintf(" x=%s", xmlAttr(floats(location.X))))
	wb.WriteString(fmt.Sprintf(" y=%s", xmlAttr(floats(location.Y))))
	wb.WriteString(fmt.Sprintf(" scale=%s", xmlAttr(floats(location.Scale))))
	wb.WriteString(" />")
	return nil
}
