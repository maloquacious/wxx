// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeMapLayers copies each <maplayer> (name, isVisible, opacity) into the
// domain map.
func decodeMapLayers(src []MapLayer_t, w *wxx.Map_t) {
	for _, layer := range src {
		w.MapLayers = append(w.MapLayers, &wxx.MapLayer_t{Name: layer.Name, IsVisible: layer.IsVisible, Opacity: layer.Opacity})
	}
}

// order of layers is important; worldographer renders them from the bottom up.
func encodeMapLayers(mapLayers []*wxx.MapLayer_t, wb *bytes.Buffer) error {
	for _, mapLayer := range mapLayers {
		if err := encodeMapLayer(mapLayer, wb); err != nil {
			return err
		}
	}
	return nil
}

func encodeMapLayer(mapLayer *wxx.MapLayer_t, wb *bytes.Buffer) error {
	wb.WriteString("<maplayer")
	wb.WriteString(fmt.Sprintf(" name=%q", mapLayer.Name))
	wb.WriteString(fmt.Sprintf(" isVisible=%q", bools(mapLayer.IsVisible)))
	wb.WriteString(fmt.Sprintf(" opacity=%q", floats(mapLayer.Opacity)))
	wb.WriteString("/>\n")
	return nil
}
