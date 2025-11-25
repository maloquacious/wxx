// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package main

type Input_t struct {
	// Terrain is a map of Worldographer terrains.
	// The key is the name of the terrain or a shortcut for it
	Terrain map[string]*Terrain_t `json:"terrain,omitempty"` // <terrainmap>
	Tiles   []*Tile_t             `json:"tiles,omitempty"`   // <tiles>
}

type Terrain_t struct {
	Slot int    `json:"slot,omitempty"`
	Name string `json:"name,omitempty"`
}

type Tile_t struct {
	Row     int    `json:"row"`
	Column  int    `json:"column"`
	Terrain string `json:"terrain,omitempty"`
}
