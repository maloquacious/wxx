// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a tool to read a Worldographer file and print
// the orientation, height, and width.
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
		fmt.Printf("bounds:\t%s\n", arg)

		w, err := xmlio.ReadFile(arg)
		if err != nil {
			fmt.Printf("\t%v\n", err)
			continue
		}
		fmt.Printf("\tversion %q: orientation %q: height %d: width %d\n", w.MetaData.DataVersion, w.HexOrientation, w.Tiles.TilesHigh, w.Tiles.TilesWide)
	}

}

//for _, loc := range []string{"AA 0101", "JK 1813", "NL 0306", "ZZ 3021"} {
//	coords, err := adapters.ConvertTNCoords(loc)
//	if err != nil {
//		fmt.Printf("%s: %v\n", loc, err)
//		continue
//	}
//	axial := adapters.OffsetToAxial(coords)
//	fmt.Printf("%s: %+v: %+v\n", loc, coords, axial)
//	tnc, err := axial.ToTNCoords()
//	if err != nil {
//		fmt.Printf("%s: %v\n", loc, err)
//		continue
//	}
//	fmt.Printf("%s: %+v: %+v: round-trip %q\n", loc, coords, axial, tnc)
//}
//
//var maps []adapters.MapData
//for _, v := range []struct {
//	loc    string
//	height int
//	width  int
//}{
//	{"AA 0101", 3, 5},
//} {
//	coords, err := adapters.ConvertTNCoords(v.loc)
//	if err != nil {
//		fmt.Printf("%s: %v\n", v.loc, err)
//		continue
//	}
//	axial := adapters.OffsetToAxial(coords)
//	md := adapters.MapData{Origin: axial, Height: v.height, Width: v.width}
//	fmt.Printf("%s: %+v: %+v\n", v.loc, coords, md)
//	maps = append(maps, md)
//}
//merged := adapters.MergeMaps(maps...)
//fmt.Printf("merged %+v\n", merged)
//
//tl, br := merged.Bounds()
//fmt.Printf("bounds: tl %+v: br %+v\n", tl, br)
//tnul, err := tl.ToTNCoords()
//if err != nil {
//	fmt.Println(err)
//}
//tnbr, err := br.ToTNCoords()
//if err != nil {
//	fmt.Println(err)
//}
//fmt.Printf("tncoords: tl %q: br %q\n", tnul, tnbr)
