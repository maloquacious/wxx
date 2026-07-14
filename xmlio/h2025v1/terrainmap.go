// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// decodeTerrainMap parses the tab-delimited name/slot table in <terrainmap> into
// the domain TerrainMap_t (List plus Data lookup).
func decodeTerrainMap(src TerrainMap_t, w *wxx.Map_t) error {
	// in the source, the terrain key and values are stored as tab delimited columns.
	w.TerrainMap = &wxx.TerrainMap_t{Data: map[string]int{}}
	if fields := strings.Split(src.InnerText, "\t"); len(fields)%2 != 0 {
		return errors.Join(wxx.ErrInvalidTerrainMapFieldCount, fmt.Errorf("field count '%d' is not even", len(fields)))
	} else {
		for len(fields) != 0 {
			t := &wxx.Terrain_t{
				Label: fields[0],
			}
			var err error
			t.Index, err = strconv.Atoi(fields[1])
			if err != nil {
				return fmt.Errorf("field: %s: invalid index: %w", fields[0], err)
			}
			w.TerrainMap.List = append(w.TerrainMap.List, t)
			w.TerrainMap.Data[t.Label] = t.Index
			fields = fields[2:]
		}
	}
	return nil
}

func encodeTerrainMap(terrainMap *wxx.TerrainMap_t, wb *bytes.Buffer) error {
	wb.WriteString(fmt.Sprintf("<terrainmap>"))
	for k, v := range terrainMapToSlice(terrainMap.Data) {
		if k == 0 {
			wb.WriteString(fmt.Sprintf("%s\t%d", v, k))
		} else {
			wb.WriteString(fmt.Sprintf("\t%s\t%d", v, k))
		}
	}
	wb.WriteString(fmt.Sprintf("</terrainmap>\n"))
	return nil
}
