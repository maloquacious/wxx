// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeFeatures copies each <feature> (with its <location> and optional inline
// <label>) into the domain map. A labelless feature leaves Feature.Label nil so
// the encoder omits the <label> child.
func decodeFeatures(src Features, w *wxx.Map_t) error {
	var err error
	for _, mFeature := range src.Features {
		f := &wxx.Feature_t{}
		f.Type = mFeature.Type
		f.Rotate = mFeature.Rotate
		f.Uuid = mFeature.Uuid
		f.MapLayer = mFeature.MapLayer
		f.IsFlipHorizontal = mFeature.IsFlipHorizontal
		f.IsFlipVertical = mFeature.IsFlipVertical
		f.Scale = mFeature.Scale
		f.ScaleHt = mFeature.ScaleHt
		f.Tags = mFeature.Tags
		if f.Color, err = decodeRgba(mFeature.Color); err != nil {
			return fmt.Errorf("feature.Color: %w", err)
		}
		if f.RingColor, err = decodeRgba(mFeature.RingColor); err != nil {
			return fmt.Errorf("feature.RingColor: %w", err)
		}
		f.IsGMOnly = mFeature.IsGMOnly
		f.IsPlaceFreely = mFeature.IsPlaceFreely
		f.LabelPosition = mFeature.LabelPosition
		f.LabelDistance = mFeature.LabelDistance
		f.IsWorld = mFeature.IsWorld
		f.IsContinent = mFeature.IsContinent
		f.IsKingdom = mFeature.IsKingdom
		f.IsProvince = mFeature.IsProvince
		f.IsFillHexBottom = mFeature.IsFillHexBottom
		f.IsHideTerrainIcon = mFeature.IsHideTerrainIcon
		f.Location = &wxx.FeatureLocation_t{
			ViewLevel: mFeature.Location.ViewLevel,
			X:         mFeature.Location.X,
			Y:         mFeature.Location.Y,
		}

		// only build f.Label when the source feature actually had a <label> child;
		// a labelless feature must leave f.Label nil so the encoder omits <label>.
		if mFeature.Label != nil {
			f.Label = &wxx.Label_t{
				MapLayer:    mFeature.Label.MapLayer,
				Style:       mFeature.Label.Style,
				FontFace:    mFeature.Label.FontFace,
				OutlineSize: mFeature.Label.OutlineSize,
				Rotate:      mFeature.Label.Rotate,
				IsBold:      mFeature.Label.IsBold,
				IsItalic:    mFeature.Label.IsItalic,
				IsWorld:     mFeature.Label.IsWorld,
				IsContinent: mFeature.Label.IsContinent,
				IsKingdom:   mFeature.Label.IsKingdom,
				IsProvince:  mFeature.Label.IsProvince,
				IsGMOnly:    mFeature.Label.IsGMOnly,
				Tags:        mFeature.Label.Tags,
				InnerText:   mFeature.Label.InnerText,

				DropShadowColor:  mFeature.Label.DropShadowColor,
				DropShadowRadius: mFeature.Label.DropShadowRadius,
				DropShadowSpread: mFeature.Label.DropShadowSpread,
			}
			if f.Label.Color, err = decodeRgba(mFeature.Label.Color); err != nil {
				return fmt.Errorf("feature.label.color: %w", err)
			}
			if f.Label.OutlineColor, err = decodeRgba(mFeature.Label.OutlineColor); err != nil {
				return fmt.Errorf("feature.label.outlineColor: %w", err)
			}
			if f.Label.BackgroundColor, err = decodeRgba(mFeature.Label.BackgroundColor); err != nil {
				return fmt.Errorf("feature.label.backgroundColor: %w", err)
			}
			f.Label.Location = &wxx.LabelLocation_t{
				ViewLevel: mFeature.Label.Location.ViewLevel,
				X:         mFeature.Label.Location.X,
				Y:         mFeature.Label.Location.Y,
				Scale:     mFeature.Label.Location.Scale,
			}
		}
		w.Features = append(w.Features, f)
	}
	return nil
}

// add features
func encodeFeatures(features []*wxx.Feature_t, wb *bytes.Buffer) error {
	wb.WriteString("<features>\n")
	for _, feature := range features {
		if err := encodeFeature(feature, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</features>\n")
	return nil
}

func encodeFeature(feature *wxx.Feature_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<feature"))
	wb.WriteString(fmt.Sprintf(" type=%q", feature.Type))
	wb.WriteString(fmt.Sprintf(" rotate=%q", floats(feature.Rotate)))
	wb.WriteString(fmt.Sprintf(" uuid=%q", feature.Uuid))
	wb.WriteString(fmt.Sprintf(" mapLayer=%q", feature.MapLayer))
	wb.WriteString(fmt.Sprintf(" isFlipHorizontal=%q", bools(feature.IsFlipHorizontal)))
	wb.WriteString(fmt.Sprintf(" isFlipVertical=%q", bools(feature.IsFlipVertical)))
	wb.WriteString(fmt.Sprintf(" scale=%q", floats(feature.Scale)))
	wb.WriteString(fmt.Sprintf(" scaleHt=%q", floats(feature.ScaleHt)))
	wb.WriteString(fmt.Sprintf(" tags=%q", feature.Tags))
	wb.WriteString(fmt.Sprintf(" color=%q", rgbans(feature.Color))) // nullable
	wb.WriteString(fmt.Sprintf(" ringcolor=%q", rgbans(feature.RingColor)))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(feature.IsGMOnly)))
	wb.WriteString(fmt.Sprintf(" isPlaceFreely=%q", bools(feature.IsPlaceFreely)))
	wb.WriteString(fmt.Sprintf(" labelPosition=%q", feature.LabelPosition))
	wb.WriteString(fmt.Sprintf(" labelDistance=%q", ints(feature.LabelDistance)))
	wb.WriteString(fmt.Sprintf(" isWorld=%q", bools(feature.IsWorld)))
	wb.WriteString(fmt.Sprintf(" isContinent=%q", bools(feature.IsContinent)))
	wb.WriteString(fmt.Sprintf(" isKingdom=%q", bools(feature.IsKingdom)))
	wb.WriteString(fmt.Sprintf(" isProvince=%q", bools(feature.IsProvince)))
	wb.WriteString(fmt.Sprintf(" isFillHexBottom=%q", bools(feature.IsFillHexBottom)))
	wb.WriteString(fmt.Sprintf(" isHideTerrainIcon=%q", bools(feature.IsHideTerrainIcon)))
	wb.WriteString(">")
	if feature.Location != nil {
		if err := encodeFeatureLocation(feature.Location, wb); err != nil {
			return err
		}
	}
	if feature.Label != nil {
		if err := encodeFeatureLabel(feature.Label, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</feature>\n")
	return nil
}

func encodeFeatureLocation(location *wxx.FeatureLocation_t, wb *bytes.Buffer) error {
	wb.WriteString("<location")
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", location.ViewLevel))
	wb.WriteString(fmt.Sprintf(" x=%q", floats(location.X)))
	wb.WriteString(fmt.Sprintf(" y=%q", floats(location.Y)))
	wb.WriteString(" />")
	return nil
}

func encodeFeatureLabel(label *wxx.Label_t, wb *bytes.Buffer) error {
	return encodeLabel(label, wb)
}
