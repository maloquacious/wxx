// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"github.com/maloquacious/wxx"
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("input/blank-30x21.wxx")
	if err != nil {
		log.Fatal(err)
	}
	w, err := wxx.Unmarshal(data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("map: application version %q\n", w.Version)
	data, err = json.MarshalIndent(w, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("map: json data %d bytes\n", len(data))
	if err = os.WriteFile("output/blank-30x21.json", data, 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("map: created %q\n", "output/blank-30x21.json")
}
