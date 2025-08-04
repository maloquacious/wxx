// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements a command line tool that shows information
// on WXX data files.
package main

import (
	"fmt"
	"github.com/maloquacious/wxx/xmlio"
	"os"
)

func main() {
	for n, arg := range os.Args {
		if n == 0 {
			continue
		}
		fmt.Printf("info:\t%s\n", arg)

		w, err := xmlio.ReadFile(arg)
		if err != nil {
			fmt.Printf("\t%v\n", err)
			continue
		}
		fmt.Printf("\t%8s data version\n", w.MetaData.DataVersion.String())
		fmt.Printf("\t%8d tiles high\n", w.Tiles.TilesHigh)
		fmt.Printf("\t%8d tiles wide\n", w.Tiles.TilesWide)
		fmt.Printf("\t%8d terrain tiles defined\n", len(w.TerrainMap.List))
	}
}
