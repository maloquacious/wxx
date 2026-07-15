// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeConfiguration copies the <configuration> tree into the domain map. The
// terrain-config / feature-config / texture-config sub-sections are copied as
// raw chardata (they are no-op(intentional) on encode); text-config/labelstyle
// and shape-config/shapestyle are fully modeled.
func decodeConfiguration(src Configuration_t, w *wxx.Map_t) error {
	var err error
	w.Configuration = &wxx.Configuration_t{}
	for _, mTerrainConfig := range src.TerrainConfig {
		wTerrainConfig := &wxx.TerrainConfig_t{
			InnerText: mTerrainConfig.InnerText,
		}
		// append the terrain configuration
		w.Configuration.TerrainConfig = append(w.Configuration.TerrainConfig, wTerrainConfig)
	}
	for _, mFeatureConfig := range src.FeatureConfig {
		wFeatureConfig := &wxx.FeatureConfig_t{
			InnerText: mFeatureConfig.InnerText,
		}
		w.Configuration.FeatureConfig = append(w.Configuration.FeatureConfig, wFeatureConfig)
	}
	for _, mTextureConfig := range src.TextureConfig {
		wTextureConfig := &wxx.TextureConfig_t{
			InnerText: mTextureConfig.InnerText,
		}
		w.Configuration.TextureConfig = append(w.Configuration.TextureConfig, wTextureConfig)
	}
	w.Configuration.TextConfig = &wxx.TextConfig_t{}
	for _, mTextConfig := range src.TextConfig {
		for _, mLabelStyle := range mTextConfig.LabelStyles {
			wLabelStyle := &wxx.LabelStyle_t{
				Name:             mLabelStyle.Name,
				FontFace:         mLabelStyle.FontFace,
				Scale:            mLabelStyle.Scale,
				IsBold:           mLabelStyle.IsBold,
				IsItalic:         mLabelStyle.IsItalic,
				OutlineSize:      mLabelStyle.OutlineSize,
				DropShadowColor:  mLabelStyle.DropShadowColor,
				DropShadowRadius: mLabelStyle.DropShadowRadius,
				DropShadowSpread: mLabelStyle.DropShadowSpread,
			}
			if wLabelStyle.Color, err = decodeRgba(mLabelStyle.Color); err != nil {
				return fmt.Errorf("labelStyle.color: %w", err)
			}
			if wLabelStyle.BackgroundColor, err = decodeRgba(mLabelStyle.BackgroundColor); err != nil {
				return fmt.Errorf("labelStyle.backgroundColor: %w", err)
			}
			if mLabelStyle.OutlineColor == "null" {
				wLabelStyle.OutlineColor = nil
			} else if wLabelStyle.OutlineColor, err = decodeZeroableRgba(mLabelStyle.OutlineColor); err != nil {
				return fmt.Errorf("labelStyle.outlineColor: %w", err)
			}
			w.Configuration.TextConfig.LabelStyles = append(w.Configuration.TextConfig.LabelStyles, wLabelStyle)
		}
	}
	w.Configuration.ShapeConfig = &wxx.ShapeConfig_t{}
	for _, mShapeConfig := range src.ShapeConfig {
		for _, mShapeStyle := range mShapeConfig.ShapeStyles {
			wShapeStyle := &wxx.ShapeStyle_t{
				Name:          mShapeStyle.Name,
				StrokeType:    mShapeStyle.StrokeType,
				IsFractal:     mShapeStyle.IsFractal,
				StrokeWidth:   mShapeStyle.StrokeWidth,
				Opacity:       mShapeStyle.Opacity,
				SnapVertices:  mShapeStyle.SnapVertices,
				Tags:          mShapeStyle.Tags,
				DropShadow:    mShapeStyle.DropShadow,
				InnerShadow:   mShapeStyle.InnerShadow,
				BoxBlur:       mShapeStyle.BoxBlur,
				DsSpread:      mShapeStyle.DsSpread,
				DsRadius:      mShapeStyle.DsRadius,
				DsOffsetX:     mShapeStyle.DsOffsetX,
				DsOffsetY:     mShapeStyle.DsOffsetY,
				InsChoke:      mShapeStyle.InsChoke,
				InsRadius:     mShapeStyle.InsRadius,
				InsOffsetX:    mShapeStyle.InsOffsetX,
				InsOffsetY:    mShapeStyle.InsOffsetY,
				BbWidth:       mShapeStyle.BbWidth,
				BbHeight:      mShapeStyle.BbHeight,
				BbIterations:  mShapeStyle.BbIterations,
				FillTexture:   mShapeStyle.FillTexture,
				StrokeTexture: mShapeStyle.StrokeTexture,
				LineCap:       mShapeStyle.LineCap,
				LineJoin:      mShapeStyle.LineJoin,
			}
			if wShapeStyle.StrokePaint, err = decodeRgba(mShapeStyle.StrokePaint); err != nil {
				return fmt.Errorf("shapeStyle.strokePaint: %w", err)
			}
			if wShapeStyle.FillPaint, err = decodeRgba(mShapeStyle.FillPaint); err != nil {
				return fmt.Errorf("shapeStyle.fillPaint: %w", err)
			}
			if wShapeStyle.DsColor, err = decodeRgba(mShapeStyle.Dscolor); err != nil {
				return fmt.Errorf("shapeStyle.dsColor: %w", err)
			}
			if wShapeStyle.InsColor, err = decodeRgba(mShapeStyle.InsColor); err != nil {
				return fmt.Errorf("shapeStyle.insColor: %w", err)
			}
			w.Configuration.ShapeConfig.ShapeStyles = append(w.Configuration.ShapeConfig.ShapeStyles, wShapeStyle)
		}
	}
	return nil
}

func encodeConfiguration(configuration *wxx.Configuration_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<configuration>\n"))
	if err := encodeTerrainConfig(configuration.TerrainConfig, wb); err != nil {
		return err
	}
	if err := encodeFeatureConfig(configuration.FeatureConfig, wb); err != nil {
		return err
	}
	if err := encodeTextureConfig(configuration.TextureConfig, wb); err != nil {
		return err
	}
	if err := encodeTextConfig(configuration.TextConfig, wb); err != nil {
		return err
	}
	if err := encodeShapeConfig(configuration.ShapeConfig, wb); err != nil {
		return err
	}
	wb.WriteString(fmt.Sprintf("  </configuration>\n"))
	return nil
}

// encodeTerrainConfig intentionally emits an empty <terrain-config> wrapper and
// ignores its argument. Real Worldographer 2025 samples leave <terrain-config>
// empty: the element carries only whitespace chardata and no child elements
// (verified in issue #4, guarded by TestW2025ConfigSectionsEmpty). Dropping the
// (empty) decoded content here is therefore lossless. If a future sample ever
// populates this section, the fix is to switch the corresponding schema.go field
// from `xml:",chardata"` to `xml:",innerxml"` and implement this encoder to emit
// the preserved inner XML.
func encodeTerrainConfig(terrainConfig []*wxx.TerrainConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <terrain-config>\n")
	wb.WriteString("  </terrain-config>\n")
	return nil
}

// encodeFeatureConfig intentionally emits an empty <feature-config> wrapper and
// ignores its argument. Real Worldographer 2025 samples leave <feature-config>
// empty: the element carries only whitespace chardata and no child elements
// (verified in issue #4, guarded by TestW2025ConfigSectionsEmpty). Dropping the
// (empty) decoded content here is therefore lossless. If a future sample ever
// populates this section, the fix is to switch the corresponding schema.go field
// from `xml:",chardata"` to `xml:",innerxml"` and implement this encoder to emit
// the preserved inner XML.
func encodeFeatureConfig(featureConfig []*wxx.FeatureConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <feature-config>\n")
	wb.WriteString("  </feature-config>\n")
	return nil
}

// encodeTextureConfig intentionally emits an empty <texture-config> wrapper and
// ignores its argument. Real Worldographer 2025 samples leave <texture-config>
// empty: the element carries only whitespace chardata and no child elements
// (verified in issue #4, guarded by TestW2025ConfigSectionsEmpty). Dropping the
// (empty) decoded content here is therefore lossless. If a future sample ever
// populates this section, the fix is to switch the corresponding schema.go field
// from `xml:",chardata"` to `xml:",innerxml"` and implement this encoder to emit
// the preserved inner XML.
func encodeTextureConfig(textureConfig []*wxx.TextureConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <texture-config>\n")
	wb.WriteString("  </texture-config>\n")
	return nil
}

func encodeTextConfig(textConfig *wxx.TextConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <text-config>\n")
	for _, labelStyle := range textConfig.LabelStyles {
		if err := encodeLabelStyle(labelStyle, wb); err != nil {
			return err
		}
	}
	wb.WriteString("  </text-config>\n")
	return nil
}

func encodeLabelStyle(labelStyle *wxx.LabelStyle_t, wb *bytes.Buffer) error {
	wb.WriteString("<labelstyle")
	wb.WriteString(fmt.Sprintf(" name=%q", labelStyle.Name))
	wb.WriteString(fmt.Sprintf(" fontFace=%q", labelStyle.FontFace))
	wb.WriteString(fmt.Sprintf(" scale=%q", floats(labelStyle.Scale)))
	wb.WriteString(fmt.Sprintf(" isBold=%q", bools(labelStyle.IsBold)))
	wb.WriteString(fmt.Sprintf(" isItalic=%q", bools(labelStyle.IsItalic)))
	wb.WriteString(fmt.Sprintf(" color=%q", rgbas(labelStyle.Color)))                     // decodeRgba
	wb.WriteString(fmt.Sprintf(" backgroundColor=%q", rgbas(labelStyle.BackgroundColor))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" outlineSize=%q", floats(labelStyle.OutlineSize)))
	wb.WriteString(fmt.Sprintf(" outlineColor=%q", rgbans(labelStyle.OutlineColor))) // "null" or decodeZeroableRgba
	// The W2025 drop-shadow trio is present all-or-none in real data;
	// dropShadowColor is "null" or an RGBA string when present, never empty, so an
	// empty DropShadowColor reliably means "absent from the source". Gate the whole
	// group on that sentinel so a round-trip does not spuriously add the attributes
	// (ADR 0002: never emit what was not on input). Do not gate on the numeric
	// fields: 0 is a legal radius/spread value.
	if labelStyle.DropShadowColor != "" {
		wb.WriteString(fmt.Sprintf(" dropShadowColor=%q", labelStyle.DropShadowColor)) // nullable string ("null")
		wb.WriteString(fmt.Sprintf(" dropShadowRadius=%q", floats(labelStyle.DropShadowRadius)))
		wb.WriteString(fmt.Sprintf(" dropShadowSpread=%q", floats(labelStyle.DropShadowSpread)))
	}
	wb.WriteString(" />\n")
	return nil
}

func encodeShapeConfig(shapeConfig *wxx.ShapeConfig_t, wb *bytes.Buffer) error {
	wb.WriteString("  <shape-config>\n")
	for _, shapeStyle := range shapeConfig.ShapeStyles {
		if err := encodeShapeStyle(shapeStyle, wb); err != nil {
			return err
		}
	}
	wb.WriteString("  </shape-config>\n")
	return nil
}

func encodeShapeStyle(shapeStyle *wxx.ShapeStyle_t, wb *bytes.Buffer) error {
	wb.WriteString("<shapestyle")
	wb.WriteString(fmt.Sprintf(" name=%q", shapeStyle.Name))
	wb.WriteString(fmt.Sprintf(" strokeType=%q", shapeStyle.StrokeType))
	wb.WriteString(fmt.Sprintf(" isFractal=%q", bools(shapeStyle.IsFractal)))
	wb.WriteString(fmt.Sprintf(" strokeWidth=%q", floats(shapeStyle.StrokeWidth)))
	wb.WriteString(fmt.Sprintf(" opacity=%q", floats(shapeStyle.Opacity)))
	wb.WriteString(fmt.Sprintf(" snapVertices=%q", bools(shapeStyle.SnapVertices)))
	wb.WriteString(fmt.Sprintf(" tags=%q", shapeStyle.Tags))
	wb.WriteString(fmt.Sprintf(" dropShadow=%q", bools(shapeStyle.DropShadow)))
	wb.WriteString(fmt.Sprintf(" innerShadow=%q", bools(shapeStyle.InnerShadow)))
	wb.WriteString(fmt.Sprintf(" boxBlur=%q", bools(shapeStyle.BoxBlur)))
	wb.WriteString(fmt.Sprintf(" dsSpread=%q", floats(shapeStyle.DsSpread)))
	wb.WriteString(fmt.Sprintf(" dsRadius=%q", floats(shapeStyle.DsRadius)))
	wb.WriteString(fmt.Sprintf(" dsOffsetX=%q", floats(shapeStyle.DsOffsetX)))
	wb.WriteString(fmt.Sprintf(" dsOffsetY=%q", floats(shapeStyle.DsOffsetY)))
	wb.WriteString(fmt.Sprintf(" insChoke=%q", floats(shapeStyle.InsChoke)))
	wb.WriteString(fmt.Sprintf(" insRadius=%q", floats(shapeStyle.InsRadius)))
	wb.WriteString(fmt.Sprintf(" insOffsetX=%q", floats(shapeStyle.InsOffsetX)))
	wb.WriteString(fmt.Sprintf(" insOffsetY=%q", floats(shapeStyle.InsOffsetY)))
	wb.WriteString(fmt.Sprintf(" bbWidth=%q", floats(shapeStyle.BbWidth)))
	wb.WriteString(fmt.Sprintf(" bbHeight=%q", floats(shapeStyle.BbHeight)))
	wb.WriteString(fmt.Sprintf(" bbIterations=%q", ints(shapeStyle.BbIterations)))
	wb.WriteString(fmt.Sprintf(" fillTexture=%q", shapeStyle.FillTexture))         // nullable
	wb.WriteString(fmt.Sprintf(" strokeTexture=%q", shapeStyle.StrokeTexture))     // nullable
	wb.WriteString(fmt.Sprintf("  strokePaint=%q", rgbas(shapeStyle.StrokePaint))) // not nullable
	wb.WriteString(fmt.Sprintf("  fillPaint=%q", rgbans(shapeStyle.FillPaint)))    // nullable
	wb.WriteString(fmt.Sprintf("  dscolor=%q", rgbans(shapeStyle.DsColor)))        // nullable
	wb.WriteString(fmt.Sprintf("  insColor=%q", rgbans(shapeStyle.InsColor)))      // nullable
	wb.WriteString(fmt.Sprintf(" lineCap=%q", shapeStyle.LineCap))
	wb.WriteString(fmt.Sprintf(" lineJoin=%q", shapeStyle.LineJoin))
	wb.WriteString(" />\n")
	return nil
}
