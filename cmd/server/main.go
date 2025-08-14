// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a web server that displays information about Worldographer files.
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/maloquacious/wxx/xmlio"
)

var m *Map_t

type Map_t struct {
	MetaData struct {
		AppVersion  string
		DataVersion string
		Created     string
	}
	HexWidth        float64
	HexHeight       float64
	HexOrientation  string
	GridOrientation string
	Rows            int
	Columns         int
}

var htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Worldographer Map Information</title>
    <link rel="stylesheet" href="https://unpkg.com/missing.css@1.2.0">
</head>
<body>
    <h1>Worldographer Map Information</h1>
    
    <section>
        <h2>Metadata</h2>
        <dl>
            <dt>Application Version</dt>
            <dd>{{.MetaData.AppVersion}}</dd>
            
            <dt>Data Version</dt>
            <dd>{{.MetaData.DataVersion}}</dd>
            
            <dt>Created</dt>
            <dd>{{.MetaData.Created}}</dd>
        </dl>
    </section>
    
    <section>
        <h2>Map Dimensions</h2>
        <dl>
            <dt>Hex Width</dt>
            <dd>{{.HexWidth}}</dd>
            
            <dt>Hex Height</dt>
            <dd>{{.HexHeight}}</dd>
            
            <dt>Rows</dt>
            <dd>{{.Rows}}</dd>
            
            <dt>Columns</dt>
            <dd>{{.Columns}}</dd>
        </dl>
    </section>
    
    <section>
        <h2>Orientation</h2>
        <dl>
            <dt>Hex Orientation</dt>
            <dd>{{.HexOrientation}}</dd>
            
            <dt>Grid Orientation</dt>
            <dd>{{.GridOrientation}}</dd>
        </dl>
    </section>
</body>
</html>`

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <worldographer-file>\n", os.Args[0])
		os.Exit(1)
	}
	
	filename := os.Args[1]
	
	// Load the Worldographer file
	fp, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer fp.Close()

	var decoderDiagnostics xmlio.DecoderDiagnostics
	decoder := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&decoderDiagnostics))
	worldMap, err := decoder.Decode(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading Worldographer file: %v\n", err)
		os.Exit(1)
	}

	// Use the pre-computed rows and columns from the decoder
	rows := worldMap.RowsHigh
	columns := worldMap.ColumnsWide

	// Create a simplified map structure for the template
	m = &Map_t{
		MetaData: struct {
			AppVersion  string
			DataVersion string
			Created     string
		}{
			AppVersion:  worldMap.MetaData.AppVersion.String(),
			DataVersion: worldMap.MetaData.DataVersion.String(),
			Created:     worldMap.MetaData.Created,
		},
		HexWidth:        worldMap.HexWidth,
		HexHeight:       worldMap.HexHeight,
		HexOrientation:  worldMap.HexOrientation,
		GridOrientation: worldMap.GridOrientation.String(),
		Rows:            rows,
		Columns:         columns,
	}

	// Setup HTTP server
	http.HandleFunc("/", handleRoot)
	
	fmt.Println("Starting server on :8080")
	fmt.Println("Visit http://localhost:8080 to view the map information")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintf(os.Stderr, "error starting web server: %v\n", err)
		os.Exit(1)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("info").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, m); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}
