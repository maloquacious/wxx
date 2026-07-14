// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

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
		}
		if wLabel.Color, err = decodeRgba(mLabel.Color); err != nil {
			return fmt.Errorf("label.color: %w", err)
		}
		if wLabel.OutlineColor, err = decodeRgba(mLabel.OutlineColor); err != nil {
			return fmt.Errorf("label.outlineColor: %w", err)
		}
		if mLabel.BackgroundColor == "" {
			wLabel.BackgroundColor = nil
		} else if wLabel.BackgroundColor, err = decodeZeroableRgba(mLabel.BackgroundColor); err != nil {
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
	wb.WriteString(fmt.Sprintf("  mapLayer=%q", label.MapLayer))
	wb.WriteString(fmt.Sprintf(" style=%q", label.Style))       // can be null!
	wb.WriteString(fmt.Sprintf(" fontFace=%q", label.FontFace)) // can be null!
	wb.WriteString(fmt.Sprintf(" color=%q", rgbas(label.Color)))
	// todo: backgroundColor is sometimes not displayed when its value is "0.0,0.0,0.0,1.0".
	// I may need to ask on the Inkwell Discord about this; I can't figure out the pattern.
	// Until then, seems to be no harm in excluding it (other than noise in the diff).
	if attr := rgbas(label.BackgroundColor); attr != "0.0,0.0,0.0,1.0" { // do not include if null
		wb.WriteString(fmt.Sprintf(" backgroundColor=%q", attr))
	}
	wb.WriteString(fmt.Sprintf(" outlineColor=%q", rgbas(label.OutlineColor)))
	wb.WriteString(fmt.Sprintf(" outlineSize=%q", floats(label.OutlineSize)))
	wb.WriteString(fmt.Sprintf(" rotate=%q", floats(label.Rotate)))
	wb.WriteString(fmt.Sprintf(" isBold=%q", bools(label.IsBold)))
	wb.WriteString(fmt.Sprintf(" isItalic=%q", bools(label.IsItalic)))
	wb.WriteString(fmt.Sprintf(" isWorld=%q", bools(label.IsWorld)))
	wb.WriteString(fmt.Sprintf(" isContinent=%q", bools(label.IsContinent)))
	wb.WriteString(fmt.Sprintf(" isKingdom=%q", bools(label.IsKingdom)))
	wb.WriteString(fmt.Sprintf(" isProvince=%q", bools(label.IsProvince)))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(label.IsGMOnly)))
	wb.WriteString(fmt.Sprintf(" tags=%q", label.Tags))
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
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", location.ViewLevel))
	wb.WriteString(fmt.Sprintf(" x=%q", floats(location.X)))
	wb.WriteString(fmt.Sprintf(" y=%q", floats(location.Y)))
	wb.WriteString(fmt.Sprintf(" scale=%q", floats(location.Scale)))
	wb.WriteString(" />")
	return nil
}
