// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a tool to read a Worldographer file and print
// the orientation, height, and width.
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
		fmt.Printf("bounds:\t%s\n", arg)

		fp, err := os.Open(arg)
		if err != nil {
			fmt.Printf("\t%v\n", err)
			continue
		}
		defer fp.Close()

		w, err := xmlio.NewDecoder().Decode(fp)
		if err != nil {
			fmt.Printf("\t%v\n", err)
			continue
		}
		fmt.Printf("\t%s: orientation %q: height %d: width %d\n", w.MetaData.Version, w.HexOrientation, w.Tiles.TilesHigh, w.Tiles.TilesWide)
	}

}
