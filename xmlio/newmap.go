// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio

import (
	"errors"
	"fmt"

	"github.com/maloquacious/wxx"
)

// CurrentApp returns the current application version: the newest one the
// registry accepts ("2.08" today).
//
// It is what NewMap uses when no version is named, and a caller can name it
// too, so a tool can offer "-app current" and tell its user which version that
// means. It is derived from the registry rather than stated, so registering a
// newer version in a codec's apps.go moves it without changing any caller.
func CurrentApp() string {
	return currentApp
}

// currentApp is set by init (codecs.go) from the registry; see newestApp.
var currentApp string

// newestApp returns the newest of the registry's application versions.
//
// Like the registry's other guards it runs at load: the registry is a constant
// of the program, so a version that cannot be ordered, or two that order
// equal ("2.6" and "2.06"), are programming errors, and choosing between them
// silently would make "current" mean whichever one map iteration found first.
func newestApp(apps []string) (string, error) {
	var newest wxx.Dotted
	tied := false
	for _, app := range apps {
		v, err := wxx.ParseDotted(app)
		if err != nil {
			return "", errors.Join(err, fmt.Errorf("version %q", app))
		}
		if newest.Raw == "" {
			newest = v
			continue
		}
		c, err := v.Compare(newest)
		if err != nil {
			return "", err
		}
		switch {
		case c > 0:
			newest, tied = v, false
		case c == 0:
			tied = true
		}
	}
	if newest.Raw == "" {
		return "", errors.Join(wxx.ErrUnsupportedMapVersion, fmt.Errorf("no application version is registered"))
	}
	if tied {
		return "", errors.Join(wxx.ErrAmbiguousAppCodec, fmt.Errorf("version %q: another registered version orders equal to it", newest.Raw))
	}
	return newest.Raw, nil
}

// NewMapOption configures NewMap.
type NewMapOption func(*newMapOptions)

type newMapOptions struct {
	app            string
	hexOrientation string
}

// WithApp names the application version whose new-map defaults NewMap uses.
// It is verbatim map/@version, as for MarshalXML; an unregistered version is
// the error MarshalXML returns for it. Without it, NewMap uses CurrentApp().
func WithApp(app string) NewMapOption {
	return func(o *newMapOptions) { o.app = app }
}

// WithHexOrientation sets the map's orientation, "COLUMNS" (the default) or
// "ROWS".
func WithHexOrientation(hexOrientation string) NewMapOption {
	return func(o *newMapOptions) { o.hexOrientation = hexOrientation }
}

// NewMap returns a new map of columns x rows hexes, every one of them Blank
// (issue #136).
//
// The map carries what the application version writes for File > New
// World/Kingdom map with the suggested pixel sizes: flat projection, WORLD view
// level, the grid and numbering, the map key, the label and shape styles, the
// eight map layers in the app's order, and an empty <informations>. The app
// fills <informations> with lore generated from its random seed, and NewMap
// generates none; that is the one difference from a map the app makes.
//
// Tiles is indexed [col][row] in both orientations, with TilesWide = columns
// and TilesHigh = rows, and each tile's Coords, Column and Row set as the
// decoder sets them. The terrain table holds Blank alone, at index 0.
//
// columns and rows must each be at least 2, the smallest map wxx makes (see
// cmd/resize). The result passes Validate.
func NewMap(columns, rows int, opts ...NewMapOption) (*wxx.Map_t, error) {
	o := newMapOptions{app: CurrentApp(), hexOrientation: "COLUMNS"}
	for _, opt := range opts {
		opt(&o)
	}
	if columns < 2 || rows < 2 {
		return nil, errors.Join(wxx.ErrInvalidTileGrid, fmt.Errorf("new map: %d x %d: want at least 2 x 2", columns, rows))
	}
	switch o.hexOrientation {
	case "COLUMNS", "ROWS":
	default:
		return nil, errors.Join(wxx.ErrInvalidHexOrientation, fmt.Errorf("new map: hexOrientation %q: want \"COLUMNS\" or \"ROWS\"", o.hexOrientation))
	}
	c, err := codecFor(o.app)
	if err != nil {
		return nil, err
	}
	return c.NewMap(o.app, columns, rows, o.hexOrientation)
}
