// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeBlurTerrainBG copies the optional top-level <blurTerrainBG> element into
// the domain map. src is nil when the element is absent from the file, in which
// case the domain field is left nil (absent).
func decodeBlurTerrainBG(src *BlurTerrainBG_t, w *wxx.Map_t) {
	if src == nil {
		return
	}
	w.BlurTerrainBG = &wxx.BlurTerrainBG_t{
		Blur:        src.Blur,
		TopBleed:    src.TopBleed,
		BottomBleed: src.BottomBleed,
		Randomness:  src.Randomness,
		BlurStart:   src.BlurStart,
		BlurEnd:     src.BlurEnd,
	}
}

// encodeBlurTerrainBG emits <blurTerrainBG> only when present (non-nil), so an
// absent element does not reappear on a round-trip.
func encodeBlurTerrainBG(blurTerrainBG *wxx.BlurTerrainBG_t, wb *bytes.Buffer) error {
	if blurTerrainBG == nil {
		return nil
	}
	wb.WriteString("<blurTerrainBG")
	wb.WriteString(fmt.Sprintf(" blur=%q", bools(blurTerrainBG.Blur)))
	wb.WriteString(fmt.Sprintf(" topBleed=%q", floats(blurTerrainBG.TopBleed)))
	wb.WriteString(fmt.Sprintf(" bottomBleed=%q", floats(blurTerrainBG.BottomBleed)))
	wb.WriteString(fmt.Sprintf(" randomness=%q", floats(blurTerrainBG.Randomness)))
	wb.WriteString(fmt.Sprintf(" blurStart=%q", floats(blurTerrainBG.BlurStart)))
	wb.WriteString(fmt.Sprintf(" blurEnd=%q", floats(blurTerrainBG.BlurEnd)))
	wb.WriteString("/>\n")
	return nil
}
