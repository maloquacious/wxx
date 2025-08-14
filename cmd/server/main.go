// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package main implements a web server that displays information about Worldographer files.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/hexg"
	"github.com/maloquacious/wxx/xmlio"
)

var worldMap *wxx.Map_t

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
    
    <section>
        <h2>Hex Grid Preview (First 5x5)</h2>
        <img src="/hex-grid.svg" alt="Hex Grid Preview" style="max-width: 100%; height: auto;">
    </section>
</body>
</html>`

func main() {
	var host = flag.String("host", "localhost", "host to bind to")
	var port = flag.String("port", "8081", "port to listen on")
	var timeout = flag.Duration("timeout", 0, "automatically shutdown after this duration (e.g. 30s, 5m, 1h)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <worldographer-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filename := flag.Arg(0)

	// Load the Worldographer file
	fp, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer fp.Close()

	var decoderDiagnostics xmlio.DecoderDiagnostics
	decoder := xmlio.NewDecoder(xmlio.WithDecoderDiagnostics(&decoderDiagnostics))
	worldMap, err = decoder.Decode(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading Worldographer file: %v\n", err)
		os.Exit(1)
	}

	// Setup HTTP server
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/hex-grid.svg", handleHexGridSVG)
	http.HandleFunc("/shutdown", handleShutdown)
	http.HandleFunc("/ping", handlePing)

	addr := *host + ":" + *port
	fmt.Printf("Starting server on %s\n", addr)
	fmt.Printf("Visit http://%s to view the map information\n", addr)

	// Setup auto-shutdown if timeout is specified
	if *timeout > 0 {
		fmt.Printf("Server will automatically shutdown after %v\n", *timeout)
		go func() {
			time.Sleep(*timeout)
			fmt.Println("Timeout reached, shutting down server...")
			os.Exit(0)
		}()
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
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

	// Create a simplified map structure for the template
	m := &Map_t{
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
		Rows:            worldMap.RowsHigh,
		Columns:         worldMap.ColumnsWide,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, m); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func handleHexGridSVG(w http.ResponseWriter, r *http.Request) {
	svg := generateHexGridSVG(worldMap)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(svg))
}

func handleShutdown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Server shutting down...\n"))
	go func() {
		os.Exit(0)
	}()
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
}

func generateHexGridSVG(m *wxx.Map_t) string {
	if m.Tiles == nil || len(m.Tiles.TileRows) == 0 {
		return `<svg><text x="10" y="20">No tiles available</text></svg>`
	}

	// Limit to first 5 rows and columns
	maxRows := 5
	maxCols := 5
	if m.RowsHigh < maxRows {
		maxRows = m.RowsHigh
	}
	if m.ColumnsWide < maxCols {
		maxCols = m.ColumnsWide
	}

	// Calculate hex layout parameters
	hexWidth := m.HexWidth
	hexHeight := m.HexHeight

	// Calculate SVG dimensions with some padding
	padding := 50.0
	svgWidth := float64(maxCols)*hexWidth + padding*2
	svgHeight := float64(maxRows)*hexHeight + padding*2

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%.1f" height="%.1f" xmlns="http://www.w3.org/2000/svg">`, svgWidth, svgHeight))
	svg.WriteString(`<style>
	.hex { fill: none; stroke: #333; stroke-width: 1; }
	.label { text-anchor: middle; font-family: monospace; font-size: 10px; fill: #000; }
	</style>`)

	// Determine if we have flat-top or pointy-top hexes
	isFlat := m.GridOrientation == hexg.EvenQ || m.GridOrientation == hexg.OddQ

	// Draw hexes
	for row := 0; row < maxRows && row < len(m.Tiles.TileRows); row++ {
		tileRow := m.Tiles.TileRows[row]
		for col := 0; col < maxCols && col < len(tileRow); col++ {
			if tile := tileRow[col]; tile != nil {
				hexPoints := generateHexPath(row, col, hexWidth, hexHeight, isFlat, padding)
				svg.WriteString(fmt.Sprintf(`<polygon class="hex" points="%s"/>`, hexPoints))

				// Calculate center for label
				centerX, centerY := calculateHexCenter(row, col, hexWidth, hexHeight, isFlat, padding)

				// Generate labels: row,col on first line, cube coords on second line
				cubeCoords := tile.Coords.String()
				svg.WriteString(fmt.Sprintf(
					`<text class="label" x="%.1f" y="%.1f">%d,%d</text>`,
					centerX, centerY-5, row, col))
				svg.WriteString(fmt.Sprintf(
					`<text class="label" x="%.1f" y="%.1f">%s</text>`,
					centerX, centerY+5, cubeCoords))
			}
		}
	}

	svg.WriteString(`</svg>`)
	return svg.String()
}

func generateHexPath(row, col int, hexWidth, hexHeight float64, isFlat bool, padding float64) string {
	centerX, centerY := calculateHexCenter(row, col, hexWidth, hexHeight, isFlat, padding)

	var points []string
	for i := 0; i < 6; i++ {
		x, y := getHexCorner(centerX, centerY, hexWidth/2, hexHeight/2, i, isFlat)
		points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	return strings.Join(points, " ")
}

func calculateHexCenter(row, col int, hexWidth, hexHeight float64, isFlat bool, padding float64) (float64, float64) {
	// Calculate position based on hex grid layout
	var x, y float64

	if isFlat { // flat-topped
		x = hexWidth * 0.75 * float64(col)
		y = hexHeight * (float64(row) + 0.5*float64(col%2))
	} else { // pointy-topped
		x = hexWidth * (float64(col) + 0.5*float64(row%2))
		y = hexHeight * 0.75 * float64(row)
	}

	return x + padding, y + padding
}

func getHexCorner(centerX, centerY, radiusX, radiusY float64, corner int, isFlat bool) (float64, float64) {
	// Calculate corner positions for hexagon
	// For flat-top hexes, first corner is at top-right
	// For pointy-top hexes, first corner is at top

	var angleDeg float64
	if isFlat {
		angleDeg = 60.0*float64(corner) + 30.0 // Flat-top hex starts at 30 degrees
	} else {
		angleDeg = 60.0 * float64(corner) // Pointy-top hex starts at 0 degrees
	}

	angleRad := angleDeg * math.Pi / 180.0
	x := centerX + radiusX*math.Cos(angleRad)
	y := centerY + radiusY*math.Sin(angleRad)

	return x, y
}
