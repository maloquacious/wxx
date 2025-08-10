// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements a command line tool that shows information
// on WXX data files.
package main

import (
	"fmt"
	"os"

	"github.com/maloquacious/wxx/xmlio"
)

func main() {
	for n, arg := range os.Args {
		if n == 0 {
			continue
		}

		fmt.Printf("info:\t%s\n", arg)
		fp, err := os.Open(arg)
		if err != nil {
			fmt.Printf("\t%v\n", err)
			continue
		}
		defer fp.Close()

		var bif xmlio.Diagnostics
		joy := xmlio.NewDecoder(xmlio.WithDiagnostics(&bif))
		w, err := joy.Decode(fp)
		if err != nil {
			fmt.Printf("\t%v\n", err)
			continue
		}
		fmt.Printf("\t%8s schema version %q\n", bif.Schema, w.MetaData.DataVersion.String())
		fmt.Printf("\t%8d tiles high\n", w.Tiles.TilesHigh)
		fmt.Printf("\t%8d tiles wide\n", w.Tiles.TilesWide)
		fmt.Printf("\t%8d terrain tiles defined\n", len(w.TerrainMap.List))
	}
}
