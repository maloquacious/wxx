// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

import (
	"bytes"

	"github.com/maloquacious/wxx"
)

// decodeExtraTerrain copies the optional top-level <extraTerrain> element into
// the domain map. src is nil when the element is absent; an empty container is
// preserved by carrying its (possibly whitespace) inner XML verbatim so that a
// present-but-empty element round-trips and any unseen children survive.
func decodeExtraTerrain(src *ExtraTerrain_t, w *wxx.Map_t) {
	if src == nil {
		return
	}
	w.ExtraTerrain = &wxx.ExtraTerrain_t{
		InnerXML: src.InnerXML,
	}
}

// encodeExtraTerrain emits <extraTerrain> only when present (non-nil). The
// preserved inner XML is written verbatim between the tags.
func encodeExtraTerrain(extraTerrain *wxx.ExtraTerrain_t, wb *bytes.Buffer) error {
	if extraTerrain == nil {
		return nil
	}
	wb.WriteString("<extraTerrain>")
	wb.WriteString(extraTerrain.InnerXML)
	wb.WriteString("</extraTerrain>\n")
	return nil
}
